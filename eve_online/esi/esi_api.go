package esi

import (
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

	response, err := http.Get(fmt.Sprintf(esiCharacterEndpointFormat, id))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Unexpected status code: %d", response.StatusCode)
	}

	character := eveonline.Character{Id: id}
	if err := json.NewDecoder(response.Body).Decode(&character); err != nil {
		return nil, fmt.Errorf("Error decoding JSON: %v", err)
	}

	characterCache.Add(id, &character)

	return &character, nil
}
