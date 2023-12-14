package main

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"
)

const (
	ZkillboardCorpEndpoint = "https://zkillboard.com/api/corporationID/"
	ZkillboardKillEndpoint = "https://zkillboard.com/api/killID/"
	UserAgentKey           = "User-Agent"
	UserAgentValue         = "https://hrveklesarov.com/ Maintainer: Hrvoje (hrvoje.lesar1@hotmail.com)"
	AcceptEncodingKey      = "Accept-Encoding"
	AcceptEncodingValue    = "gzip"
	Capsule                = 670
)

var cache = expirable.NewLRU[int64, EsiCharacter](50, nil, time.Hour*24*30)

func esiKillmailEndpoint(killmailid int64, hash string) string {
	return fmt.Sprintf("https://esi.evetech.net/latest/killmails/%d/%s/?datasource=tranquility", killmailid, hash)
}

func esiCharacterEndpoint(characterId int64) string {
	return fmt.Sprintf("https://esi.evetech.net/latest/characters/%d/?datasource=tranquility", characterId)
}

type CorporationZkillboard struct {
	CorporationId int64
	Killmails     []ZkillboardKillmail
}

type ZkillboardKillmail struct {
	KillmailId int64 `json:"killmail_id"`
	Zkb        struct {
		LocationId     int64   `json:"locationID"`
		Hash           string  `json:"hash"`
		FittedValue    float64 `json:"fittedValue"`
		DroppedValue   float64 `json:"droppedValue"`
		DestroyedValue float64 `json:"destroyedValue"`
		TotalValue     float64 `json:"totalValue"`
		Points         int     `json:"points"`
		Npc            bool    `json:"npc"`
		Solo           bool    `json:"solo"`
		Awox           bool    `json:"awox"`
	} `json:"zkb"`
}

func (zk *ZkillboardKillmail) fetchKillmail() error {
	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s%d/", ZkillboardKillEndpoint, zk.KillmailId), nil)
	request.Header.Set(UserAgentKey, UserAgentValue)
	request.Header.Set(AcceptEncodingKey, AcceptEncodingValue)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}

	gzipReader, err := gzip.NewReader(response.Body)
	if err != nil {
		return err
	}
	defer gzipReader.Close()

	responseBodyBytes, err := io.ReadAll(gzipReader)
	if err != nil {
		return err
	}

	// Zkillboard always returns an array even when fetching a single killmail
	// TODO: https://zkillboard.com/api/killID/113897332/ amarr control tower
	killmails := make([]ZkillboardKillmail, 0, 1)
	err = json.Unmarshal(responseBodyBytes, &killmails)
	if err != nil {
		return err
	}

	zk.Zkb = killmails[0].Zkb
	return nil
}

func (zk *CorporationZkillboard) requestUrl() string {
	return fmt.Sprintf("%s%d/", ZkillboardCorpEndpoint, zk.CorporationId)
}

func (zk *CorporationZkillboard) fetchKillmails() error {
	zk.Killmails = make([]ZkillboardKillmail, 0)
	request, err := http.NewRequest(http.MethodGet, zk.requestUrl(), nil)
	if err != nil {
		return err
	}
	request.Header.Set(UserAgentKey, UserAgentValue)
	request.Header.Set(AcceptEncodingKey, AcceptEncodingValue)

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}

	gzipReader, err := gzip.NewReader(response.Body)
	if err != nil {
		return err
	}
	defer gzipReader.Close()

	responseBodyBytes, err := io.ReadAll(gzipReader)
	if err != nil {
		return err
	}

	err = json.Unmarshal(responseBodyBytes, &zk.Killmails)
	if err != nil {
		return err
	}

	return nil
}

type EsiKillmail struct {
	Attacksers []struct {
		CharacterId int64 `json:"character_id"`
		FinalBlow   bool  `json:"final_blow"`
	} `json:"attackers"`
	KillmailId int64 `json:"killmail_id"`
	Victim     struct {
		CharacterId int64 `json:"character_id"`
		ShipTypeId  int64 `json:"ship_type_id"`
	} `json:"victim"`
}

func (ek *EsiKillmail) findFinalBlowCharacter() (int64, error) {
	for idx := range ek.Attacksers {
		attacker := &ek.Attacksers[idx]
		if attacker.FinalBlow {
			return attacker.CharacterId, nil
		}
	}
	return 0, errors.New("Failed to find final blow character")
}

type EsiCharacter struct {
	Id            int64
	Name          string `json:"name"`
	CorporationId int64  `json:"corporation_id"`
}

type PairedKillmail struct {
	EsiKillmail        *EsiKillmail
	ZkillboardKillmail *ZkillboardKillmail
}

func (pk *PairedKillmail) isWorthlessCapsule() bool {
	return pk.EsiKillmail.Victim.ShipTypeId == Capsule && pk.ZkillboardKillmail.Zkb.TotalValue <= 10001
}

func (pk *PairedKillmail) getCharacters() (victim *EsiCharacter, finalBlowCharacter *EsiCharacter, err error) {
	if pk.ZkillboardKillmail.Zkb.Npc {
		return nil, nil, errors.New("Npc kill")
	}

	victimCharId := pk.EsiKillmail.Victim.CharacterId
	finalBlowCharacterId, err := pk.EsiKillmail.findFinalBlowCharacter()
	if err != nil {
		return nil, nil, err
	}

	victim, err = fetchCharacter(victimCharId)
	if err != nil {
		return nil, nil, err
	}
	finalBlowCharacter, err = fetchCharacter(finalBlowCharacterId)
	if err != nil {
		return nil, nil, err
	}

	return victim, finalBlowCharacter, nil
}

type FeedboardKillmail struct {
	CorporationId int64
	Killmail      PairedKillmail
	Victim        EsiCharacter
	FinalBlow     EsiCharacter
}

func (fk *FeedboardKillmail) KillmailId() int64 {
	return fk.Killmail.ZkillboardKillmail.KillmailId
}

func (fk *FeedboardKillmail) ShipTypeId() int64 {
	return fk.Killmail.EsiKillmail.Victim.ShipTypeId
}

func (fk *FeedboardKillmail) IsKill() bool {
	return fk.Victim.CorporationId != fk.CorporationId
}

func (fk *FeedboardKillmail) Isk() string {
	return Format(fk.Killmail.ZkillboardKillmail.Zkb.TotalValue)
}

type Esi struct {
	CorporationId       int64
	Mutext              sync.RWMutex
	KillmailLimit       int
	Killmails           []FeedboardKillmail
	WebsocketServerChan chan<- []FeedboardKillmail
	CharacterCache      *expirable.LRU[int64, EsiCharacter]
}

func newEsi(killmailLimit int, websocketServerChan chan<- []FeedboardKillmail) *Esi {
	return &Esi{
		KillmailLimit:       killmailLimit,
		WebsocketServerChan: websocketServerChan,
	}
}

func (e *Esi) fetchKillmails(zk CorporationZkillboard) {
	killmails := make([]FeedboardKillmail, 0, e.KillmailLimit)
	for idx := range zk.Killmails {
		if len(killmails) == e.KillmailLimit {
			break
		}
		zkillboardKillmail := &zk.Killmails[idx]
		esiKillmail, err := fetchKillmail(zkillboardKillmail.KillmailId, zkillboardKillmail.Zkb.Hash)

		if err != nil {
			log.Println(err)
			continue
		}

		pair := PairedKillmail{
			EsiKillmail:        esiKillmail,
			ZkillboardKillmail: zkillboardKillmail,
		}

		if pair.isWorthlessCapsule() {
			continue
		}

		victim, finalBlow, err := pair.getCharacters()
		if err != nil {
			log.Println(err)
			continue
		}

		killmails = append(killmails, FeedboardKillmail{
			CorporationId: zk.CorporationId,
			Killmail:      pair,
			Victim:        *victim,
			FinalBlow:     *finalBlow,
		})
	}

	e.Killmails = killmails
}

func fetchKillmail(killmailid int64, hash string) (*EsiKillmail, error) {
	requestUrl := esiKillmailEndpoint(killmailid, hash)
	request, err := http.NewRequest(http.MethodGet, requestUrl, nil)
	if err != nil {
		return nil, err
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}

	responseBodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var esiKillmail EsiKillmail
	err = json.Unmarshal(responseBodyBytes, &esiKillmail)
	if err != nil {
		return nil, err
	}

	return &esiKillmail, nil
}

func fetchCharacter(charId int64) (*EsiCharacter, error) {
	cachedChar, ok := cache.Get(charId)
	if ok {
		return &cachedChar, nil
	}

	character := EsiCharacter{Id: charId}
	requestUrl := esiCharacterEndpoint(charId)

	request, err := http.NewRequest(http.MethodGet, requestUrl, nil)
	if err != nil {
		return nil, err
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}

	responseBodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(responseBodyBytes, &character)
	if err != nil {
		return nil, err
	}

	cache.Add(charId, character)
	return &character, nil
}

func (e *Esi) fetchFiftyFiftyFiftyFeeds() error {
	zkillboard := CorporationZkillboard{CorporationId: 98684728}
	err := zkillboard.fetchKillmails()
	if err != nil {
		return err
	}

	e.CorporationId = zkillboard.CorporationId

	e.fetchKillmails(zkillboard)

	return nil
}

func (e *Esi) handleWebsocketKillmail(km *ZkillWebsocketSimpleKillmail) {
	esiKillmail, err := fetchKillmail(km.KillId, *km.Hash)
	if err != nil {
		log.Println(err)
		return
	}

	zkillboardKillmail := ZkillboardKillmail{KillmailId: esiKillmail.KillmailId}
	err = zkillboardKillmail.fetchKillmail()
	if err != nil {
		log.Println(err)
		return
	}

	pair := PairedKillmail{
		EsiKillmail:        esiKillmail,
		ZkillboardKillmail: &zkillboardKillmail,
	}

	if pair.isWorthlessCapsule() {
		return
	}

	victim, finalBlow, err := pair.getCharacters()
	if err != nil {
		log.Println(err)
		return
	}

	killmail := FeedboardKillmail{
		CorporationId: e.CorporationId,
		Killmail:      pair,
		Victim:        *victim,
		FinalBlow:     *finalBlow,
	}

	e.appendKillmailToStart(killmail)
	e.sendKillmailsToWebsocket()
}

func (e *Esi) setKillmails(killmails []FeedboardKillmail) {
	e.Mutext.Lock()

	e.Killmails = killmails

	e.Mutext.Unlock()
}

func (e *Esi) appendKillmailToStart(killmail FeedboardKillmail) {
	e.Mutext.Lock()

	e.Killmails = append([]FeedboardKillmail{killmail}, e.Killmails[:e.KillmailLimit-1]...)

	e.Mutext.Unlock()
}

func (e *Esi) sendKillmailsToWebsocket() {
	e.Mutext.RLock()
	e.WebsocketServerChan <- e.Killmails
	e.Mutext.RUnlock()
}
