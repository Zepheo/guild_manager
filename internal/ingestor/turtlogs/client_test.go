package turtlogs

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetRaidLoot(t *testing.T) {
	// 1. Create a mock server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the URL path is correct
		expectedPath := "/api/v1/logs/12345"
		if r.URL.Path != expectedPath {
			t.Errorf("expected path %s, got %s", expectedPath, r.URL.Path)
		}

		// Return a mock JSON response
		response := LogResponse{
			Loot: []LootEntry{
				{ItemName: "Ashkandi", PlayerName: "Zepheo"},
				{ItemName: "Revenant Hebdomad", PlayerName: "Grom"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer mockServer.Close()

	// 2. Initialize client pointing to mock server
	client := NewTurtlogsClient(mockServer.URL)

	// 3. Execute
	data, err := client.GetRaidLoot("12345")

	// 4. Assertions
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(data.Loot) != 2 {
		t.Errorf("expected 2 loot entries, got %d", len(data.Loot))
	}

	if data.Loot[0].ItemName != "Ashkandi" {
		t.Errorf("expected Ashkandi, got %s", data.Loot[0].ItemName)
	}
}

func TestGetRaidLoot_HttpError(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockServer.Close()

	client := NewTurtlogsClient(mockServer.URL)
	_, err := client.GetRaidLoot("invalid_id")

	if err == nil {
		t.Error("expected error for 404 status, got nil")
	}
}
