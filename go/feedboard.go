package main

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
)

const (
	ZkillboardCorpEndpoint = "https://zkillboard.com/api/corporationID/"
	UserAgentKey           = "User-Agent"
	UserAgentValue         = "https://hrveklesarov.com/ Maintainer: Hrvoje (hrvoje.lesar1@hotmail.com)"
	AcceptEncodingKey      = "Accept-Encoding"
	AcceptEncodingValue    = "gzip"
	Capsule                = 670
)

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
	Name          string `json:"name"`
	CorporationId int64  `json:"corporation_id"`
}

type PairedKillmail struct {
	EsiKillmail        EsiKillmail
	ZkillboardKillmail ZkillboardKillmail
}

type PairedKillmailCharacterId struct {
	FinalBlowCharId int64
	VictimCharId    int64
	PairdKillmail   PairedKillmail
}

type PairedCharacters struct {
	FinalBlowChar   EsiCharacter
	VictimChar      EsiCharacter
	FinalBlowCharId int64
	VictimCharId    int64
	PairdKillmail   PairedKillmail
}

type Esi struct {
	KillmailLimit            int
	Killmails                []PairedKillmail
	PairedKillmailCharacters []PairedKillmailCharacterId
	PairedCharacters         []PairedCharacters
}

func newEsi(killmailLimit int) *Esi {
	return &Esi{
		KillmailLimit: killmailLimit,
	}
}

func (e *Esi) fetchKillmails(zk Zkillboard) error {
	killmails := make([]PairedKillmail, 0, e.KillmailLimit)
	for idx := range zk.Killmails {
		if len(killmails) == e.KillmailLimit {
			break
		}
		zkillboardKillmail := &zk.Killmails[idx]
		requestUrl := esiKillmailEndpoint(zkillboardKillmail.KillmailId, zkillboardKillmail.Zkb.Hash)
		request, err := http.NewRequest(http.MethodGet, requestUrl, nil)
		if err != nil {
			return err
		}

		response, err := http.DefaultClient.Do(request)
		if err != nil {
			return err
		}

		responseBodyBytes, err := io.ReadAll(response.Body)
		if err != nil {
			return err
		}

		var esiKillmail EsiKillmail
		err = json.Unmarshal(responseBodyBytes, &esiKillmail)
		if err != nil {
			return err
		}

		pair := PairedKillmail{
			EsiKillmail:        esiKillmail,
			ZkillboardKillmail: *zkillboardKillmail,
		}

		if pair.EsiKillmail.Victim.ShipTypeId == Capsule && pair.ZkillboardKillmail.Zkb.FittedValue <= 10001 {
			continue
		}
		e.Killmails = append(e.Killmails, pair)
	}

	err := e.pairCharactersToKillmail()
	if err != nil {
		return err
	}
	err = e.getCharacters()
	if err != nil {
		return err
	}
	return nil
}

func (e *Esi) pairCharactersToKillmail() error {
	e.PairedKillmailCharacters = make([]PairedKillmailCharacterId, 0, len(e.Killmails))
	for idx := range e.Killmails {
		pair := &e.Killmails[idx]
		if pair.ZkillboardKillmail.Zkb.Npc {
			continue
		}
		victimCharId := pair.EsiKillmail.Victim.CharacterId
		finalBlowCharacterId, err := pair.EsiKillmail.findFinalBlowCharacter()
		if err != nil {
			log.Println(err)
			continue
		}

		e.PairedKillmailCharacters = append(e.PairedKillmailCharacters, PairedKillmailCharacterId{
			VictimCharId:    victimCharId,
			FinalBlowCharId: finalBlowCharacterId,
			PairdKillmail:   *pair,
		})
	}

	return nil
}

func (e *Esi) fetchCharacter(charId int64, character *EsiCharacter) error {
	requestUrl := esiCharacterEndpoint(charId)

	request, err := http.NewRequest(http.MethodGet, requestUrl, nil)
	if err != nil {
		return err
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}

	responseBodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(responseBodyBytes, character)
	if err != nil {
		return err
	}

	return nil
}

func (e *Esi) getCharacters() error {
	e.PairedCharacters = make([]PairedCharacters, 0, len(e.PairedKillmailCharacters))
	for idx := range e.PairedKillmailCharacters {
		killmailPair := &e.PairedKillmailCharacters[idx]
		pairedCharacters := PairedCharacters{
			FinalBlowCharId: killmailPair.FinalBlowCharId,
			VictimCharId:    killmailPair.VictimCharId,
			PairdKillmail:   killmailPair.PairdKillmail,
		}

		err := e.fetchCharacter(killmailPair.FinalBlowCharId, &pairedCharacters.FinalBlowChar)
		if err != nil {
			return err
		}

		e.fetchCharacter(killmailPair.VictimCharId, &pairedCharacters.VictimChar)
		if err != nil {
			return err
		}

		e.PairedCharacters = append(e.PairedCharacters, pairedCharacters)
	}

	return nil
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

	e.clear()
	return nil
}

func (e *Esi) clear() {
	e.PairedKillmailCharacters = nil
	e.PairedCharacters = nil
}
