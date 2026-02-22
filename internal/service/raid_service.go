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
