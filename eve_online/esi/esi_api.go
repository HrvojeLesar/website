package esi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	eveonline "github.com/HrvojeLesar/website/eve_online"
	httpclient "github.com/HrvojeLesar/website/internal/http_client"
	"github.com/hashicorp/golang-lru/v2/expirable"
)

type esiEndpoint struct {
	httpClient     httpclient.Client
	characterCache *expirable.LRU[int, *eveonline.Character]
}

func NewESIEndpoint(client httpclient.Client) *esiEndpoint {
	cache := expirable.NewLRU[int, *eveonline.Character](50, nil, time.Hour*24*30)
	return &esiEndpoint{
		httpClient:     client,
		characterCache: cache,
	}
}

func GetEsiCharacterUrl(id int) string {
	return fmt.Sprintf("https://esi.evetech.net/latest/characters/%d/?datasource=tranquility", id)
}

var EsiEndpoint = NewESIEndpoint(&http.Client{})

func (e *esiEndpoint) FetchCharacter(id int, ctx context.Context) (*eveonline.Character, error) {
	if cached, ok := e.characterCache.Get(id); ok {
		return cached, nil
	}

	url := GetEsiCharacterUrl(id)
	resp, err := e.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var character eveonline.Character
	character.Id = id
	if err := json.NewDecoder(resp.Body).Decode(&character); err != nil {
		return nil, fmt.Errorf("decoding JSON: %w", err)
	}

	e.characterCache.Add(id, &character)
	return &character, nil
}
