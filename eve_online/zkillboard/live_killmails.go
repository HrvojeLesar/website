package zkillboard

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

type zkillboard struct{}

var Zkillboard zkillboard

type Killmail struct {
	KillmailID      int    `json:"killmail_id"`
	Hash            string `json:"hash"`
	UploadedAt      int64  `json:"uploaded_at"`
	SequenceID      int    `json:"sequence_id"`
	SequenceUpdated *bool  `json:"sequence_updated"`

	Esi struct {
		KillmailID    int    `json:"killmail_id"`
		KillmailTime  string `json:"killmail_time"`
		SolarSystemID int    `json:"solar_system_id"`

		Attackers []struct {
			AllianceID     int     `json:"alliance_id"`
			CharacterID    int     `json:"character_id"`
			CorporationID  int     `json:"corporation_id"`
			DamageDone     int     `json:"damage_done"`
			FinalBlow      bool    `json:"final_blow"`
			SecurityStatus float64 `json:"security_status"`
			ShipTypeID     int     `json:"ship_type_id"`
			WeaponTypeID   int     `json:"weapon_type_id"`
		} `json:"attackers"`

		Victim struct {
			CharacterID   int `json:"character_id"`
			CorporationID int `json:"corporation_id"`
			DamageTaken   int `json:"damage_taken"`
			FactionID     int `json:"faction_id"`
			ShipTypeID    int `json:"ship_type_id"`
		} `json:"victim"`
	} `json:"esi"`

	ZKB struct {
		LocationID     int     `json:"locationID"`
		Hash           string  `json:"hash"`
		FittedValue    float64 `json:"fittedValue"`
		DroppedValue   float64 `json:"droppedValue"`
		DestroyedValue float64 `json:"destroyedValue"`
		TotalValue     float64 `json:"totalValue"`
		NPC            bool    `json:"npc"`
		Solo           bool    `json:"solo"`
		Awox           bool    `json:"awox"`
		AttackerCount  int     `json:"attackerCount"`
		Href           string  `json:"href"`
	} `json:"zkb"`
}

type zkillboardR2Z2 struct {
	KillMailsChan chan *Killmail
}

func (zkillboard *zkillboard) NewZkillboardR2Z2() zkillboardR2Z2 {
	return zkillboardR2Z2{
		KillMailsChan: make(chan *Killmail, 10),
	}
}

func (r2z2 *zkillboardR2Z2) Start() {
	sequenceNumber, err := r2z2.getLatestSequence()
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to get latest sequence number. Live killmail fetching will not work! Error: %v", err))
		return
	}

	go r2z2.startFetchingSequences(*sequenceNumber)
}

func (r2z2 *zkillboardR2Z2) getLatestSequence() (*SequenceNumber, error) {
	context, cancelContext := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelContext()

	zkillSequenceNumber, apiError := SequenceApiEndpoint.GetLatestSequence(context)
	savedSequenceNumber, fileError := SequenceFile.LastSaved()

	if apiError != nil && fileError != nil {
		return nil, fmt.Errorf("Failed to get sequence number from api and file. Api error: %v. File error: %v", apiError, fileError)
	}

	if apiError == nil && fileError != nil {
		return &zkillSequenceNumber, nil
	}

	if apiError != nil && fileError == nil {
		return &savedSequenceNumber, nil
	}

	if zkillSequenceNumber.Sequence > savedSequenceNumber.Sequence {
		return &savedSequenceNumber, nil
	} else {
		return &zkillSequenceNumber, nil
	}
}

func (r2z2 *zkillboardR2Z2) startFetchingSequences(sequenceNumber SequenceNumber) {
	for {
		context, cancelContext := context.WithTimeout(context.Background(), 5*time.Second)
		killmail, err := KillmailFetcherApiEndpoint.Fetch(sequenceNumber, context)
		cancelContext()

		if err != nil {
			slog.Error(fmt.Sprintf("Failed to fetch killmail on sequence number: %d. Error: %v", sequenceNumber.Sequence, err))
			slog.Debug("Sleeping for 6 seconds after error")
			time.Sleep(6 * time.Second)
			continue
		}

		if killmail == nil {
			panic("Killmail was expected to be not nil, got nil")
		}

		sequenceNumber.Sequence += 1

		r2z2.KillMailsChan <- killmail

		// Keeps poll rate at 10/s
		time.Sleep(100 * time.Millisecond)
	}
}
