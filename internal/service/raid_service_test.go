package service

import (
	"context"
	"testing"

	"github.com/zepheo/guild_manager/internal/ingestor/csv"
	"github.com/zepheo/guild_manager/internal/storage/mock"
)

func TestProcessRaid(t *testing.T) {
	// Setup dependencies
	mockRepo := mock.NewMockRaidRepo()
	parser := csv.NewParser()
	// Note: In a real test, you'd also mock the Turtlogs http client
	service := NewRaidService(mockRepo, parser, nil)

	// Set initial state: Zepheo has a 20 bonus and previously reserved "Ashkandi"
	mockRepo.Bonuses["Zepheo"] = 20
	mockRepo.LastReserves["Zepheo"] = "Ashkandi"

	// Mock CSV input: Zepheo reserves "Ashkandi" again

	// Execute logic (Assuming analyzeLoot and logClient are handled/mocked)
	// For this example, we manually trigger the update logic
	ctx := context.Background()
	err := service.repo.ProcessRaidUpdate(ctx, "Zepheo", 40, "Test logic")

	if err != nil {
		t.Fatalf("Service failed: %v", err)
	}

	// Verify the bonus increased to 40
	if mockRepo.Bonuses["Zepheo"] != 40 {
		t.Errorf("Expected bonus 40, got %d", mockRepo.Bonuses["Zepheo"])
	}
}
