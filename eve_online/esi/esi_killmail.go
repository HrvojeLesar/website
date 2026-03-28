package esi

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	eveonline "github.com/HrvojeLesar/website/eve_online"
)

const esiKillmailEndpointFormat = "https://esi.evetech.net/latest/killmails/%d/%s/?datasource=tranquility"

type esiKillmailEndpoint struct{}

var EsiKillmail esiKillmailEndpoint

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

func (endpoint *esiKillmailEndpoint) Fetch(killmailid int64, hash string, ctx context.Context) (*ESIKillmail, error) {
	request, err := eveonline.CustomRequest.New(ctx, http.MethodGet, fmt.Sprintf(esiKillmailEndpointFormat, killmailid, hash), nil)
	if err != nil {
		slog.Error("Failed to create Esikillmail request", "error", err)
		return nil, err
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		slog.Error("Esikillmail response failed", "error", err)
		return nil, err
	}

	gzipReader, err := gzip.NewReader(response.Body)
	if err != nil {
		slog.Error("GZip reader creation failed", "error", err)
		return nil, err
	}
	defer gzipReader.Close()

	var killmail ESIKillmail
	if err := json.NewDecoder(gzipReader).Decode(&killmail); err != nil {
		slog.Error("Failed to parse esi killmail json", "error", err)
		return nil, err
	}

	return &killmail, nil
}
