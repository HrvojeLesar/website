package esi

import (
	"context"
	"errors"
	"net/http"
	"testing"

	eveonline "github.com/HrvojeLesar/website/eve_online"
	"github.com/HrvojeLesar/website/internal/testingutils"
)

func TestFetchCharacter(t *testing.T) {
	t.Run("returns character from API", func(t *testing.T) {
		client := new(testingutils.MockHTTPClient)
		endpoint := NewESIEndpoint(client)

		expected := eveonline.Character{Id: 123, Name: "Test Character", CorporationId: 456}
		client.Response = &http.Response{
			StatusCode: http.StatusOK,
			Body:       client.MockBody(testingutils.JsonBytes(expected)),
		}

		result, err := endpoint.FetchCharacter(123, context.Background())

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Name != expected.Name {
			t.Errorf("got name %q, want %q", result.Name, expected.Name)
		}
	})

	t.Run("returns error on HTTP failure", func(t *testing.T) {
		client := new(testingutils.MockHTTPClient)
		endpoint := NewESIEndpoint(client)
		client.Err = errors.New("connection refused")

		_, err := endpoint.FetchCharacter(123, context.Background())

		if err == nil {
			t.Fatal("want error, got nil")
		}
	})

	t.Run("returns error on non-200 status", func(t *testing.T) {
		client := new(testingutils.MockHTTPClient)
		endpoint := NewESIEndpoint(client)
		client.Response = &http.Response{
			StatusCode: http.StatusNotFound,
			Body:       client.MockBody(nil),
		}

		_, err := endpoint.FetchCharacter(999, context.Background())

		if err == nil {
			t.Fatal("want error, got nil")
		}
	})

	t.Run("returns error on invalid JSON", func(t *testing.T) {
		client := new(testingutils.MockHTTPClient)
		endpoint := NewESIEndpoint(client)
		client.Response = &http.Response{
			StatusCode: http.StatusOK,
			Body:       client.MockBody([]byte("invalid json")),
		}

		_, err := endpoint.FetchCharacter(999, context.Background())

		if err == nil {
			t.Fatal("want error, got nil")
		}
	})

	t.Run("returns cached character", func(t *testing.T) {
		client := new(testingutils.MockHTTPClient)
		endpoint := NewESIEndpoint(client)
		cached := &eveonline.Character{Id: 999, Name: "Cached", CorporationId: 111}
		endpoint.characterCache.Add(999, cached)

		result, err := endpoint.FetchCharacter(999, context.Background())

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Name != "Cached" {
			t.Errorf("got name %q, want %q", result.Name, "Cached")
		}
		if client.Called {
			t.Error("client should not be called for cached character")
		}
	})
}
