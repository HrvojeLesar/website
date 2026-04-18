package esi

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/HrvojeLesar/website/internal/testingutils"
)

func TestFetchKillmail(t *testing.T) {
	t.Run("returns killmail from API", func(t *testing.T) {
		client := new(testingutils.MockHTTPClient)
		endpoint := NewESIKillmail(client)

		expected := ESIKillmail{
			KillmailID:    123,
			KillmailTime:  "2024-01-01T12:00:00Z",
			SolarSystemID: 30000142,
		}
		client.Response = &http.Response{
			StatusCode: http.StatusOK,
			Body:       client.MockBody(testingutils.JsonBytes(expected)),
		}

		result, err := endpoint.Fetch(123, "hash123", context.Background())

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.KillmailID != expected.KillmailID {
			t.Errorf("got killmail_id %d, want %d", result.KillmailID, expected.KillmailID)
		}
	})

	t.Run("returns error on HTTP failure", func(t *testing.T) {
		client := new(testingutils.MockHTTPClient)
		endpoint := NewESIKillmail(client)
		client.Err = errors.New("connection refused")

		_, err := endpoint.Fetch(123, "hash", context.Background())

		if err == nil {
			t.Fatal("want error, got nil")
		}
	})

	t.Run("returns error on non-200 status", func(t *testing.T) {
		client := new(testingutils.MockHTTPClient)
		endpoint := NewESIKillmail(client)
		client.Response = &http.Response{
			StatusCode: http.StatusForbidden,
			Body:       client.MockBody(nil),
		}

		_, err := endpoint.Fetch(999, "hash", context.Background())

		if err == nil {
			t.Fatal("want error, got nil")
		}
	})

	t.Run("returns error on invalid JSON", func(t *testing.T) {
		client := new(testingutils.MockHTTPClient)
		endpoint := NewESIKillmail(client)
		client.Response = &http.Response{
			StatusCode: http.StatusOK,
			Body:       client.MockBody([]byte("not valid")),
		}

		_, err := endpoint.Fetch(999, "hash", context.Background())

		if err == nil {
			t.Fatal("want error, got nil")
		}
	})
}
