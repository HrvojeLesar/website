package zkillboard

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	eveonline "github.com/HrvojeLesar/website/eve_online"
)

type sequenceApiEndpoit struct{}

var SequenceApiEndpoint sequenceApiEndpoit

const sequenceUrl = "https://r2z2.zkillboard.com/ephemeral/sequence.json"

func (sequence *sequenceApiEndpoit) GetLatestSequence(ctx context.Context) (SequenceNumber, error) {
	var sequenceNumber SequenceNumber

	request, err := eveonline.CustomRequest.New(ctx, http.MethodGet, sequenceUrl, nil)
	if err != nil {
		return sequenceNumber, err
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return sequenceNumber, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return sequenceNumber, fmt.Errorf("Unexpected status code: %d", response.StatusCode)
	}

	if err := json.NewDecoder(response.Body).Decode(&sequenceNumber); err != nil {
		return sequenceNumber, fmt.Errorf("Error decoding JSON: %v", err)
	}

	return sequenceNumber, nil
}
