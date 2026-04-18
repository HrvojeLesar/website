package zkillboard

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"
)

type zkillboard struct{}

const FIFTY_FIFTY_FIFTY_CORPORATION_ID = 98684728

const SKIP_AFTER_FETCH_FAIL_COUNT = 10

var Zkillboard zkillboard

type zkillboardR2Z2 struct {
	KillMailsChan chan *Killmail
	filterFunc    func(*Killmail) bool
}

func (zkillboard *zkillboard) NewZkillboardR2Z2(filterFunc func(*Killmail) bool) zkillboardR2Z2 {
	return zkillboardR2Z2{
		KillMailsChan: make(chan *Killmail, 10),
		filterFunc:    filterFunc,
	}
}

func (zkillboard *zkillboard) DefaultKillmailFilterFunc(killmail *Killmail) bool {
	if killmail.IsNpcFeed() || killmail.IsWorthlessCapsule() {
		return false
	}

	if killmail.Esi.Victim.CorporationID == FIFTY_FIFTY_FIFTY_CORPORATION_ID {
		return true
	}

	for idx := range killmail.Esi.Attackers {
		if killmail.Esi.Attackers[idx].CorporationID == FIFTY_FIFTY_FIFTY_CORPORATION_ID {
			return true
		}
	}

	return false
}

func (r2z2 *zkillboardR2Z2) Start() {
	sequenceNumber, err := r2z2.getLatestSequence()
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to get latest sequence number. Live killmail fetching will not work! Error: %v", err))
		return
	}

	go r2z2.startLivekillmailFetching(*sequenceNumber)
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
		slog.Error(fmt.Sprintf("No api error, sequence %v %v", zkillSequenceNumber, zkillSequenceNumber.Sequence))
		return &zkillSequenceNumber, nil
	}

	if apiError != nil && fileError == nil {
		slog.Error(fmt.Sprintf("No file error, sequence %v", savedSequenceNumber))
		return &savedSequenceNumber, nil
	}

	if zkillSequenceNumber.Sequence > savedSequenceNumber.Sequence {
		slog.Error(fmt.Sprintf("No errors, saved sequence: %v", savedSequenceNumber))
		return &savedSequenceNumber, nil
	} else {
		slog.Error(fmt.Sprintf("No errors, zkill sequence: %v", savedSequenceNumber))
		return &zkillSequenceNumber, nil
	}
}

func (r2z2 *zkillboardR2Z2) startLivekillmailFetching(sequenceNumber SequenceNumber) {
	fetchFailCounter := 0

	for {
		context, cancelContext := context.WithTimeout(context.Background(), 5*time.Second)
		killmail, err := KillmailFetcherApiEndpoint.Fetch(sequenceNumber, context)
		cancelContext()

		if !errors.Is(err, SequenceNumberNotFoundError) {
			sequenceNumber.Sequence += 1
			SequenceFile.Save(sequenceNumber)
		}

		if fetchFailCounter >= SKIP_AFTER_FETCH_FAIL_COUNT {
			sequenceNumber.Sequence += 1
			fetchFailCounter = 0
		}

		if err != nil {
			slog.Error(fmt.Sprintf("Failed to fetch killmail on sequence number: %d. Error: %v", sequenceNumber.Sequence, err))
			slog.Debug("Sleeping for 6 seconds after error")
			time.Sleep(6 * time.Second)

			fetchFailCounter += 1
			continue
		}

		if killmail == nil {
			panic("Killmail was expected to be not nil, got nil")
		}

		if r2z2.filterFunc == nil || r2z2.filterFunc(killmail) {
			err = killmail.FetchCharacters()
			if err != nil {
				slog.Error("Failed to fetch characters", "error", err)
				continue
			}
			r2z2.KillMailsChan <- killmail
		}

		// Keeps poll rate at 10/s
		time.Sleep(100 * time.Millisecond)
	}
}
