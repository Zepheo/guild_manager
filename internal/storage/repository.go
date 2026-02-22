package storage

import (
	"context"

	"github.com/zepheo/guild_manager/internal/domain"
)

type RaidRepository interface {
	GetLastReserve(ctx context.Context, charName string) (string, error)
	GetBulkPlayerBonuses(ctx context.Context, charNames []string) (map[string]map[string]int, error)
	GetPlayerBonuses(ctx context.Context, charName string) (map[string]int, error)
	RecordRaid(ctx context.Context, raid *domain.RaidResult) error
}
