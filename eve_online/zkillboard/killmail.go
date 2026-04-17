package zkillboard

import (
	"context"
	"log/slog"
	"math"
	"sync"
	"time"

	eveonline "github.com/HrvojeLesar/website/eve_online"
	"github.com/HrvojeLesar/website/eve_online/esi"
)

const (
	Capsule    = 670
	CapsuleIsk = 12000
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
	SequenceUpdated *int   `json:"sequence_updated"`

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

type KillmailCollection []*Killmail

type KillmailsTotalValue struct {
	Wins      float64
	Losses    float64
	Total     float64
	Kills     int
	LostShips int
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

		results[1] = fetchResult{
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
	return killmail.Esi.Victim.CorporationID != FIFTY_FIFTY_FIFTY_CORPORATION_ID
}

func (killmail *Killmail) IsWorthlessCapsule() bool {
	if killmail.Esi.Victim.ShipTypeID == Capsule {
		slog.Info("Capsule detected", "capsule", killmail)
	}
	return killmail.Esi.Victim.ShipTypeID == Capsule && killmail.Zkb.TotalValue <= CapsuleIsk
}

func (killmail *Killmail) Isk() string {
	return eveonline.FormatIsk(killmail.Zkb.TotalValue)
}

func (killmail *Killmail) IsNpcFeed() bool {
	return killmail.Zkb.Npc
}

func (killmailCollection KillmailCollection) TotalIskValue() KillmailsTotalValue {
	totals := KillmailsTotalValue{
		Wins:      0,
		Losses:    0,
		Total:     0,
		Kills:     0,
		LostShips: 0,
	}

	for _, killmail := range killmailCollection {
		if killmail.IsFiftyFiftyFiftyKill() {
			totals.Wins += killmail.Zkb.TotalValue
			totals.Kills += 1
			totals.Total += killmail.Zkb.TotalValue
		} else {
			totals.Losses += killmail.Zkb.TotalValue
			totals.LostShips += 1
			totals.Total -= killmail.Zkb.TotalValue
		}
	}

	return totals
}

func (totals KillmailsTotalValue) IskTotal() string {
	return eveonline.FormatIsk(math.Abs(totals.Total))
}

func (totals KillmailsTotalValue) IskWins() string {
	return eveonline.FormatIsk(totals.Wins)
}

func (totals KillmailsTotalValue) IskLosses() string {
	return eveonline.FormatIsk(totals.Losses)
}

func (totals KillmailsTotalValue) IsIskPositive() bool {
	return totals.Total > 0
}
