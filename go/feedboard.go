package main

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/nkall/compactnumber"
)

const (
	ZkillboardCorpEndpoint = "https://zkillboard.com/api/corporationID/"
	UserAgentKey           = "User-Agent"
	UserAgentValue         = "https://hrveklesarov.com/ Maintainer: Hrvoje (hrvoje.lesar1@hotmail.com)"
	AcceptEncodingKey      = "Accept-Encoding"
	AcceptEncodingValue    = "gzip"
	Capsule                = 670
)

var formatter = compactnumber.NewFormatter("en-UK", compactnumber.Short)

func esiKillmailEndpoint(killmailid int64, hash string) string {
	return fmt.Sprintf("https://esi.evetech.net/latest/killmails/%d/%s/?datasource=tranquility", killmailid, hash)
}

func esiCharacterEndpoint(characterId int64) string {
	return fmt.Sprintf("https://esi.evetech.net/latest/characters/%d/?datasource=tranquility", characterId)
}

type Zkillboard struct {
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

func (zk *Zkillboard) requestUrl() string {
	return fmt.Sprintf("%s%d/", ZkillboardCorpEndpoint, zk.CorporationId)
}

func (zk *Zkillboard) fetchKillmails() error {
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
	return pk.EsiKillmail.Victim.ShipTypeId == Capsule && pk.ZkillboardKillmail.Zkb.FittedValue <= 10001
}

func (pk *PairedKillmail) getCharacters() (*EsiCharacter, *EsiCharacter, error) {
	if pk.ZkillboardKillmail.Zkb.Npc {
		return nil, nil, errors.New("Npc kill")
	}

	victimCharId := pk.EsiKillmail.Victim.CharacterId
	finalBlowCharacterId, err := pk.EsiKillmail.findFinalBlowCharacter()
	if err != nil {
		return nil, nil, err
	}

	victim, err := fetchCharacter(victimCharId)
	if err != nil {
		return nil, nil, err
	}
	finalBlowCharacter, err := fetchCharacter(finalBlowCharacterId)
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
	iskValue, err := formatter.Format(int(fk.Killmail.ZkillboardKillmail.Zkb.TotalValue))
	if err != nil {
		log.Println(err)
		return ""
	}
	return iskValue
}

type Esi struct {
	KillmailLimit int
	Killmails     []FeedboardKillmail
}

func newEsi(killmailLimit int) *Esi {
	return &Esi{
		KillmailLimit: killmailLimit,
	}
}

func (e *Esi) fetchKillmails(zk Zkillboard) error {
	killmails := make([]FeedboardKillmail, 0, e.KillmailLimit)
	for idx := range zk.Killmails {
		if len(killmails) == e.KillmailLimit {
			break
		}
		zkillboardKillmail := &zk.Killmails[idx]
		requestUrl := esiKillmailEndpoint(zkillboardKillmail.KillmailId, zkillboardKillmail.Zkb.Hash)
		request, err := http.NewRequest(http.MethodGet, requestUrl, nil)
		if err != nil {
			return nil
		}

		response, err := http.DefaultClient.Do(request)
		if err != nil {
			return nil
		}

		responseBodyBytes, err := io.ReadAll(response.Body)
		if err != nil {
			return nil
		}

		var esiKillmail EsiKillmail
		err = json.Unmarshal(responseBodyBytes, &esiKillmail)
		if err != nil {
			return nil
		}

		pair := PairedKillmail{
			EsiKillmail:        &esiKillmail,
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

	return nil
}

func fetchCharacter(charId int64) (*EsiCharacter, error) {
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

	return &character, nil
}

func (e *Esi) fetchFiftyFiftyFiftyFeeds() error {
	zkillboard := Zkillboard{CorporationId: 98684728}
	err := zkillboard.fetchKillmails()
	if err != nil {
		return err
	}

	err = e.fetchKillmails(zkillboard)
	if err != nil {
		return err
	}

	return nil
}
