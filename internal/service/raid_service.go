package service

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/zepheo/guild_manager/internal/domain"
	"github.com/zepheo/guild_manager/internal/ingestor/csv"
	"github.com/zepheo/guild_manager/internal/ingestor/turtlogs"
	"github.com/zepheo/guild_manager/internal/storage"
)

type RaidService struct {
	repo      storage.RaidRepository
	csvParser *csv.Parser
	logClient *turtlogs.TurtlogsClient
}

// NewRaidService is the constructor (Dependency Injection)
func NewRaidService(r storage.RaidRepository, p *csv.Parser, c *turtlogs.TurtlogsClient) *RaidService {
	return &RaidService{
		repo:      r,
		csvParser: p,
		logClient: c,
	}
}

// ProcessRaid brings everything together
func (s *RaidService) ProcessRaid(ctx context.Context, csvFile io.Reader, logID string) error {
	// 1. Use the specific method we defined in the CSV package
	reserves, err := s.csvParser.ParseRaidResCSV(csvFile)
	if err != nil {
		return fmt.Errorf("csv parsing failed: %w", err)
	}

	// 2. Fetch the JSON loot from Turtlogs
	logData, err := s.logClient.GetRaidLoot(logID)
	if err != nil {
		return fmt.Errorf("turtlogs fetch failed: %w", err)
	}

	// 3. Process each player found in the CSV
	for _, res := range reserves {
		// 1. Get current state (or default to 0/empty)
		prevItem, _ := s.repo.GetLastReserve(ctx, res.PlayerName)
		currentBonus, err := s.repo.GetPlayerBonus(ctx, res.PlayerName)
		if err != nil {
			// If the player doesn't exist, GetPlayerBonus returns 0, nil already.
			// But if there's a real DB error, we should log it.
			fmt.Printf("Warning: could not fetch bonus for %s: %v\n", res.PlayerName, err)
		}

		dropped, winner := s.analyzeLoot(res.ItemName, logData)

		// 2. Apply SR+ Logic
		newScore := domain.CalculateNewBonus(
			currentBonus,
			prevItem,
			res.ItemName,
			dropped,
			winner == res.PlayerName,
		)

		// 3. Persist
		reason := fmt.Sprintf("Log %s: %s (Dropped: %v, Won: %v)", logID, res.ItemName, dropped, winner == res.PlayerName)
		s.repo.UpdatePlayerBonus(ctx, res.PlayerName, newScore, reason)
	}

	return nil
}

func (s *RaidService) analyzeLoot(reservedItem string, logData *turtlogs.LogResponse) (bool, string) {
	for _, drop := range logData.Loot {
		// We use Case Insensitive comparison for safety
		if strings.EqualFold(strings.TrimSpace(drop.ItemName), strings.TrimSpace(reservedItem)) {
			return true, drop.PlayerName
		}
	}
	return false, ""
}
