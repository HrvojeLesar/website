package zkillboard

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	eveonline "github.com/HrvojeLesar/website/eve_online"
)

type killmailFetcher struct{}

var KillmailFetcherApiEndpoint killmailFetcher

type sequenceNumberNotFoundError struct{}

func (s sequenceNumberNotFoundError) Error() string {
	return "Requested sequence number does not exist"
}

var SequenceNumberNotFoundError sequenceNumberNotFoundError

const killmailEndpointFormat = "https://r2z2.zkillboard.com/ephemeral/%d.json"

func (fetcher *killmailFetcher) Fetch(sequenceNumber SequenceNumber, context context.Context) (*Killmail, error) {
	slog.Info("Fetching", "sequence number", sequenceNumber.Sequence)

	request, err := eveonline.CustomRequest.New(context, http.MethodGet, fmt.Sprintf(killmailEndpointFormat, sequenceNumber.Sequence), nil)
	if err != nil {
		return nil, err
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusNotFound {
		return nil, SequenceNumberNotFoundError
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Unexpected status code: %d", response.StatusCode)
	}

	gzipReader, err := gzip.NewReader(response.Body)
	if err != nil {
		return nil, err
	}
	defer gzipReader.Close()

	var killmail Killmail
	if err := json.NewDecoder(gzipReader).Decode(&killmail); err != nil {
		return nil, fmt.Errorf("Error decoding JSON: %v", err)
	}

	return &killmail, nil
}
