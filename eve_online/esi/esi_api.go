package esi

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	eveonline "github.com/HrvojeLesar/website/eve_online"
	"github.com/hashicorp/golang-lru/v2/expirable"
)

type esiEndpoint struct{}

var EsiEndpoint esiEndpoint

const esiCharacterEndpointFormat = "https://esi.evetech.net/latest/characters/%d/?datasource=tranquility"

var characterCache = expirable.NewLRU[int, *eveonline.Character](50, nil, time.Hour*24*30)

func (esiEndpoint *esiEndpoint) FetchCharacter(id int, ctx context.Context) (*eveonline.Character, error) {
	cachedChar, ok := characterCache.Get(id)
	if ok {
		return cachedChar, nil
	}

	request, err := eveonline.CustomRequest.New(ctx, http.MethodGet, fmt.Sprintf(esiCharacterEndpointFormat, id), nil)
	if err != nil {
		return nil, err
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}

	gzipReader, err := gzip.NewReader(response.Body)
	if err != nil {
		return nil, err
	}
	defer gzipReader.Close()

	character := eveonline.Character{Id: id}
	if err := json.NewDecoder(gzipReader).Decode(&character); err != nil {
		return nil, fmt.Errorf("Error decoding JSON: %v", err)
	}

	characterCache.Add(id, &character)

	return &character, nil
}
