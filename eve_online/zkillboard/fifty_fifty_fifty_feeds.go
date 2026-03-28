package zkillboard

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	eveonline "github.com/HrvojeLesar/website/eve_online"
	"github.com/HrvojeLesar/website/eve_online/esi"
)

type fiftyFiftyFiftyFeeds struct{}

var FiftyFiftyFiftyFeeds fiftyFiftyFiftyFeeds

const zkillboardCorpEndpointFormat = "https://zkillboard.com/api/corporationID/%d/"

type fiftyFiftyFiftyFeedsCache struct {
	feeds               []*Killmail
	newKillmailsChannel chan *Killmail
	dirty               bool
	mutex               sync.Mutex
	itemLimit           int
}

type FiftyFiftyFiftyFeedsCache interface {
	SetNotDirty()
	IsDirty() bool
	Killmails() []*Killmail
}

func (f *fiftyFiftyFiftyFeeds) Fetch(ctx context.Context) []*Killmail {
	request, err := eveonline.CustomRequest.New(ctx, http.MethodGet, f.corporationEndpointUrl(FIFTY_FIFTY_FIFTY_CORPORATION_ID), nil)
	if err != nil {
		slog.Error("Failed to create FiftyFiftyFiftyFeeds request", "error", err)
		return nil
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		slog.Error("FiftyFiftyFiftyFeeds response failed", "error", err)
		return nil
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		slog.Error("Unexpected status code", "status code", response.StatusCode)
		return nil
	}

	gzipReader, err := gzip.NewReader(response.Body)
	if err != nil {
		slog.Error("GZip reader creation failed", "error", err)
		return nil
	}
	defer gzipReader.Close()

	var killmails []CorporationKillmail
	if err := json.NewDecoder(gzipReader).Decode(&killmails); err != nil {
		slog.Error("Failed to parse corporation killmails json", "error", err)
		return nil
	}

	return f.fetchKillmailInfo(killmails)
}

func (f *fiftyFiftyFiftyFeeds) NewCache(newKillmailsChannel chan *Killmail, cacheItemLimit int) fiftyFiftyFiftyFeedsCache {
	return fiftyFiftyFiftyFeedsCache{
		newKillmailsChannel: newKillmailsChannel,
		itemLimit:           cacheItemLimit,
		feeds:               make([]*Killmail, 0, cacheItemLimit),
	}
}

func (f *fiftyFiftyFiftyFeeds) corporationEndpointUrl(corporationId int) string {
	return fmt.Sprintf(zkillboardCorpEndpointFormat, corporationId)
}

func (f *fiftyFiftyFiftyFeeds) fetchKillmailInfo(corpKillmails []CorporationKillmail) []*Killmail {
	killmails := make([]*Killmail, 0, len(corpKillmails))

	for idx := range corpKillmails {
		corpKillmail := &corpKillmails[idx]

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		esiKillmail, err := esi.EsiKillmail.Fetch(int64(corpKillmail.KillmailID), corpKillmail.Zkb.Hash, ctx)
		cancel()
		if err != nil {
			slog.Error("Failed to fetch esi killmail", "error", err)
			continue
		}

		killmail := KillmailConverter.FromEsiAndCorporationKillmail(esiKillmail, corpKillmail)
		err = killmail.FetchCharacters()
		if err != nil {
			slog.Error("Failed to fetch killmail characters", "error", err)
			continue
		}

		if killmail.IsWorthlessCapsule() {
			slog.Info("Detected worthless capsule, skipping killmail")
			continue
		}

		if killmail.IsNpcFeed() {
			slog.Info("Detected npc feed, skipping killmail")
			continue
		}

		killmails = append([]*Killmail{&killmail}, killmails...)

		if len(killmails) >= eveonline.KILLMAILCOUNT {
			break
		}
	}

	return killmails
}

func (cache *fiftyFiftyFiftyFeedsCache) FetchKillmails() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	killmails := FiftyFiftyFiftyFeeds.Fetch(ctx)
	cancel()

	killmailsToAdd := min(len(killmails), cache.itemLimit)

	cache.mutex.Lock()
	for idx := range killmailsToAdd {
		cache.addKillmail(killmails[idx])
	}
	cache.mutex.Unlock()
}

func (cache *fiftyFiftyFiftyFeedsCache) StartListening() {
	go func() {
		for {
			killmail := <-cache.newKillmailsChannel
			if killmail == nil {
				continue
			}

			cache.mutex.Lock()
			cache.addKillmail(killmail)
			cache.mutex.Unlock()
		}
	}()
}

func (cache *fiftyFiftyFiftyFeedsCache) addKillmail(killmail *Killmail) {
	cache.feeds = append([]*Killmail{killmail}, cache.feeds...)
	if len(cache.feeds) > cache.itemLimit {
		cache.feeds = cache.feeds[:cache.itemLimit]
	}

	cache.dirty = true
}

func (cache *fiftyFiftyFiftyFeedsCache) IsDirty() bool {
	return cache.dirty
}

func (cache *fiftyFiftyFiftyFeedsCache) Killmails() []*Killmail {
	return cache.feeds
}

func (cache *fiftyFiftyFiftyFeedsCache) SetNotDirty() {
	cache.mutex.Lock()
	cache.dirty = false
	cache.mutex.Unlock()
}
