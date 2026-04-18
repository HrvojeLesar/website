package esi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	httpclient "github.com/HrvojeLesar/website/internal/http_client"
)

type esiKillmailEndpoint struct {
	httpClient httpclient.Client
}

func NewESIKillmail(client httpclient.Client) *esiKillmailEndpoint {
	return &esiKillmailEndpoint{httpClient: client}
}

var EsiKillmail = NewESIKillmail(&http.Client{})

type ESIKillmail struct {
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
}

func (e *esiKillmailEndpoint) Fetch(killmailID int64, hash string, ctx context.Context) (*ESIKillmail, error) {
	url := fmt.Sprintf("https://esi.evetech.net/latest/killmails/%d/%s/?datasource=tranquility", killmailID, hash)
	resp, err := e.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var killmail ESIKillmail
	if err := json.NewDecoder(resp.Body).Decode(&killmail); err != nil {
		return nil, fmt.Errorf("decoding JSON: %w", err)
	}

	return &killmail, nil
}
