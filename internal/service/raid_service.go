package service

import (
	"context"
	"io"

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

func NewRaidService(r storage.RaidRepository, p *csv.Parser, c *turtlogs.TurtlogsClient) *RaidService {
	return &RaidService{
		repo:      r,
		csvParser: p,
		logClient: c,
	}
}

type SRPlus struct {
	ItemName string
	SR       int
}

func (s *RaidService) CalculateSRPlus(ctx context.Context, csvFile io.Reader, raid domain.Raid) (map[string][]SRPlus, error) {
	// 1. Parse the incoming CSV
	reserves, err := s.csvParser.ParseRaidResCSV(csvFile)
	if err != nil {
		return nil, err
	}

	// 2. Prepare the result map: PlayerName -> List of SRPlus objects
	nameMap := make(map[string]bool)
	for _, r := range reserves {
		nameMap[r.PlayerName] = true
	}

	var names []string
	for name := range nameMap {
		names = append(names, name)
	}

	// Bulk fetch bonuses
	bulkData, err := s.repo.GetBulkPlayerBonuses(ctx, names, raid)
	if err != nil {
		return nil, err
	}

	result := make(map[string][]SRPlus)
	for _, entry := range reserves {
		if playerBonuses, ok := bulkData[entry.PlayerName]; ok {
			if bonus, exists := playerBonuses[entry.ItemName]; exists && bonus > 0 {
				result[entry.PlayerName] = append(result[entry.PlayerName], SRPlus{
					ItemName: entry.ItemName,
					SR:       bonus,
				})
			}
		}
	}
	return result, nil
}

func (s *RaidService) ProcessRaid(ctx context.Context, csvFile io.Reader, logID string) error {
	reserves, err := s.csvParser.ParseRaidResCSV(csvFile)
	if err != nil {
		return err
	}

	logData, err := s.logClient.GetRaidData(logID)
	if err != nil {
		return err
	}

	result := &domain.RaidResult{
		RaidID:   logData.Meta.RaidID,
		RaidDate: logData.Meta.RaidDate.Time,
		Raid:     logData.Meta.Raid,
	}

	for _, res := range reserves {
		result.Reserves = append(result.Reserves, res)
	}

	for _, drop := range logData.LootDrops {
		result.Drops = append(result.Drops, domain.LootDrop{
			ItemName:   drop.ItemName,
			WinnerName: drop.WinnerName,
		})
	}

	for _, att := range logData.Players {
		result.Attendees = append(result.Attendees, att)
	}

	return s.repo.RecordRaid(ctx, result)
}
