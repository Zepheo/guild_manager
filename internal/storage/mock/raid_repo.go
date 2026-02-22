package mock

import (
	"context"

	"github.com/zepheo/guild_manager/internal/domain"
)

type MockRaidRepo struct {
	// Simple in-memory storage for testing
	Bonuses      map[string]int
	LastReserves map[string]string
	RaidHistory  []*domain.RaidResult
}

func NewMockRaidRepo() *MockRaidRepo {
	return &MockRaidRepo{
		Bonuses:      make(map[string]int),
		LastReserves: make(map[string]string),
	}
}

func (m *MockRaidRepo) GetLastReserve(ctx context.Context, charName string) (string, error) {
	return m.LastReserves[charName], nil
}

func (m *MockRaidRepo) GetPlayerBonus(ctx context.Context, charName string) (int, error) {
	return m.Bonuses[charName], nil
}

func (m *MockRaidRepo) RecordRaid(ctx context.Context, raid *domain.RaidResult) error {
	m.RaidHistory = append(m.RaidHistory, raid)
	return nil
}
