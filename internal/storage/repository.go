package storage

import (
	"context"

	"github.com/zepheo/guild_manager/internal/domain"
)

type RaidRepository interface {
	GetLastReserve(ctx context.Context, charName string) (string, error)
	GetPlayerBonus(ctx context.Context, charName string) (int, error)
	ProcessRaidUpdate(ctx context.Context, charName string, newBonus int, reason string) error
	UpdatePlayerBonus(ctx context.Context, charName string, newBonus int, reason string) error
	RecordRaid(ctx context.Context, raid *domain.RaidResult) error
}
