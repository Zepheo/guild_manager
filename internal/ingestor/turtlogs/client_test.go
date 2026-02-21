package turtlogs

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestMapClassId tests the utility function for WoW class mapping
func TestMapClassId(t *testing.T) {
	tests := []struct {
		id       int
		expected string
	}{
		{1, "Warrior"},
		{11, "Druid"},
		{99, "Unknown"},
	}

	for _, tt := range tests {
		if res := mapClassId(tt.id); res != tt.expected {
			t.Errorf("mapClassId(%d) = %s; want %s", tt.id, res, tt.expected)
		}
	}
}

// TestGetRaidPlayers tests the participant fetching and class mapping integration
func TestGetRaidPlayers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mockParticipants := []Participants{
			{CharacterID: 101, Name: "Thrall", HeroClassID: 7}, // Shaman
			{CharacterID: 102, Name: "Uther", HeroClassID: 2},  // Paladin
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockParticipants)
	}))
	defer server.Close()

	client := NewTurtlogsClient(server.URL)
	players, err := client.GetRaidPlayers("12345")

	if err != nil {
		t.Fatalf("Failed to get players: %v", err)
	}

	if players[101].Class != "Shaman" || players[102].CharacterName != "Uther" {
		t.Errorf("Players data mismatch: %+v", players)
	}
}

// TestGetRaidLoot tests the extraction from the raw interface slice
func TestGetRaidLoot(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The API returns a list of lists with mixed types
		// Entry[3] is WinnerID (float64 in JSON), Entry[5] is ItemName (string)
		mockLoot := [][]any{
			{0, 0, 0, 500, 0, "Ashkandi", 0, 0},
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockLoot)
	}))
	defer server.Close()

	client := NewTurtlogsClient(server.URL)
	loot, err := client.GetRaidLoot("12345")

	if err != nil {
		t.Fatalf("Failed to get loot: %v", err)
	}

	// Note: You will need to fix the bug in GetRaidLoot for this to pass (see below)
	if len(loot) == 0 || loot[0].ItemName != "Ashkandi" || loot[0].WinnerID != 500 {
		t.Errorf("Loot data mismatch: %+v", loot)
	}
}

// TestGetLogId tests the Regex logic for different URL formats
func TestGetLogId(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The API returns a list of lists with mixed types
		mockTiny := TinyResolverResponse{ID: 90027, URLPayload: "{\"payload\":{\"instance_meta_id\":90027}}"}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockTiny)
	}))
	defer server.Close()
	client := NewTurtlogsClient(server.URL)

	tests := []struct {
		input    string
		expected string
	}{
		{"12345", "12345"},                           // Raw ID
		{server.URL + "/viewer/67890/base", "67890"}, // Full URL
		{server.URL + "/tiny_url/67890", "90027"},    // Tiny URL
		{"invalid-string", ""},                       // Error case
	}

	for _, tt := range tests {
		res, err := client.getLogId(tt.input)
		if tt.expected == "" && err == nil {
			t.Errorf("Expected error for input %s, but got none", tt.input)
		}
		if res != tt.expected {
			t.Errorf("getLogId(%s) = %s; want %s", tt.input, res, tt.expected)
		}
	}
}
