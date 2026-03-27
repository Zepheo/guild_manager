package storage

import (
	"context"

	"github.com/zepheo/guild_manager/internal/domain"
)

type RaidRepository interface {
	GetLastReserve(ctx context.Context, charName string, raid domain.Raid) (string, error)
	GetBulkPlayerBonuses(ctx context.Context, charNames []string, raid domain.Raid) (map[string]map[string]int, error)
	GetPlayerBonuses(ctx context.Context, charName string, raid domain.Raid) (map[string]int, error)
	RecordRaid(ctx context.Context, raid *domain.RaidResult) error
}
