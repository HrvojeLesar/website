package zkillboard

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	eveonline "github.com/HrvojeLesar/website/eve_online"
)

type killmailFetcher struct{}

var KillmailFetcherApiEndpoint killmailFetcher

const killmailEndpointFormat = "https://r2z2.zkillboard.com/ephemeral/%d.json"

func (fetcher *killmailFetcher) Fetch(sequenceNumber SequenceNumber, context context.Context) (*Killmail, error) {
	request, err := eveonline.CustomRequest.New(context, http.MethodGet, fmt.Sprintf(killmailEndpointFormat, sequenceNumber.Sequence), nil)
	if err != nil {
		return nil, err
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Unexpected status code: %d", response.StatusCode)
	}

	var killmail Killmail
	if err := json.NewDecoder(response.Body).Decode(&killmail); err != nil {
		return nil, fmt.Errorf("Error decoding JSON: %v", err)
	}

	return &killmail, nil
}
