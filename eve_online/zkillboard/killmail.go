package zkillboard

import (
	"context"
	"log/slog"
	"sync"
	"time"

	eveonline "github.com/HrvojeLesar/website/eve_online"
	"github.com/HrvojeLesar/website/eve_online/esi"
)

const (
	Capsule    = 670
	CapsuleIsk = 10001
)

type Zkb struct {
	LocationID     int     `json:"locationID"`
	Hash           string  `json:"hash"`
	FittedValue    float64 `json:"fittedValue"`
	DroppedValue   float64 `json:"droppedValue"`
	DestroyedValue float64 `json:"destroyedValue"`
	TotalValue     float64 `json:"totalValue"`
	Npc            bool    `json:"npc"`
	Solo           bool    `json:"solo"`
	Awox           bool    `json:"awox"`
	AttackerCount  int     `json:"attackerCount"`
	Href           string  `json:"href"`
}

type Killmail struct {
	KillmailID      int    `json:"killmail_id"`
	Hash            string `json:"hash"`
	SequenceUpdated *bool  `json:"sequence_updated"`

	Esi esi.ESIKillmail `json:"esi"`

	Zkb       Zkb `json:"zkb"`
	FinalBlow *eveonline.Character
	Victim    *eveonline.Character

	fetchMutex sync.Mutex
}

type CorporationKillmail struct {
	KillmailID int `json:"killmail_id"`
	Zkb        Zkb `json:"zkb"`
}

func (killmail *Killmail) FetchCharacters() error {
	killmail.fetchMutex.Lock()
	defer killmail.fetchMutex.Unlock()

	if killmail.FinalBlow != nil && killmail.Victim != nil {
		return nil
	}

	type fetchResult struct {
		character *eveonline.Character
		err       error
	}

	results := make([]fetchResult, 2)
	var waitGroup sync.WaitGroup
	waitGroup.Add(2)

	go func() {
		finalBlowFetchContext, finalBlowFetchCancel := context.WithTimeout(context.Background(), 5*time.Second)
		finalBlowCharacter, err := esi.EsiEndpoint.FetchCharacter(killmail.findFinalBlowCharacterId(), finalBlowFetchContext)
		finalBlowFetchCancel()

		results[0] = fetchResult{
			character: finalBlowCharacter,
			err:       err,
		}

		waitGroup.Done()
	}()

	go func() {
		victimFetchContext, victimFetchCancel := context.WithTimeout(context.Background(), 5*time.Second)
		victimCharacter, err := esi.EsiEndpoint.FetchCharacter(killmail.Esi.Victim.CharacterID, victimFetchContext)
		victimFetchCancel()

		results[0] = fetchResult{
			character: victimCharacter,
			err:       err,
		}

		waitGroup.Done()
	}()

	waitGroup.Wait()

	err := results[0].err
	if err != nil {
		slog.Error("Failed to fetch final blow character", "error", err)
		return err
	}

	err = results[1].err
	if err != nil {
		slog.Error("Failed to fetch victim character", "error", err)
		return err
	}

	killmail.FinalBlow = results[0].character
	killmail.Victim = results[1].character

	return nil
}

func (killmail *Killmail) findFinalBlowCharacterId() int {
	for idx := range killmail.Esi.Attackers {
		attacker := &killmail.Esi.Attackers[idx]
		if attacker.FinalBlow {
			return attacker.CharacterID
		}
	}

	return killmail.Esi.Victim.CharacterID
}

func (killmail *Killmail) IsFiftyFiftyFiftyKill() bool {
	return killmail.Victim.CorporationId != FIFTY_FIFTY_FIFTY_CORPORATION_ID
}

func (killmail *Killmail) isWorthlessCapsule() bool {
	return killmail.Esi.Victim.ShipTypeID == Capsule && killmail.Zkb.TotalValue <= CapsuleIsk
}
