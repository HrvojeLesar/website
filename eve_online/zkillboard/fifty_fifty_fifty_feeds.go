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

func (f *fiftyFiftyFiftyFeeds) corporationEndpointUrl(corporationId int) string {
	return fmt.Sprintf(zkillboardCorpEndpointFormat, corporationId)
}

func (f *fiftyFiftyFiftyFeeds) fetchKillmailInfo(corpKillmails []CorporationKillmail) []*Killmail {
	type result struct {
		esiKillmail  *esi.ESIKillmail
		corpKillmail *CorporationKillmail
		err          error
	}

	killmailsChannel := make(chan result, 10)

	for idx := range corpKillmails {
		corpKillmail := &corpKillmails[idx]

		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			esiKillmail, err := esi.EsiKillmail.Fetch(int64(corpKillmail.KillmailID), corpKillmail.Zkb.Hash, ctx)
			cancel()

			killmailsChannel <- result{esiKillmail: esiKillmail, corpKillmail: corpKillmail, err: err}
		}()
	}

	killmails := make([]*Killmail, 0, len(corpKillmails))
	var killmailsMutex sync.Mutex

	for range corpKillmails {
		esiKillmailResult := <-killmailsChannel
		if esiKillmailResult.err != nil {
			slog.Error("Failed fetching esi killmail", "error", esiKillmailResult.err)
			continue
		}

		killmail := KillmailConverter.FromEsiAndCorporationKillmail(esiKillmailResult.esiKillmail, esiKillmailResult.corpKillmail)
		go func() {
			err := killmail.FetchCharacters()
			if err != nil {
				slog.Error("Failed to fetch killmail characters", "error", err)
				return
			}

			if killmail.IsWorthlessCapsule() {
				slog.Info("Detected worthless capsule, skipping killmail")
				return
			}

			killmailsMutex.Lock()
			defer killmailsMutex.Unlock()
			killmails = append(killmails, &killmail)
		}()
	}

	return killmails
}
