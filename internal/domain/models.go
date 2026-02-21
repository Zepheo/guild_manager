package domain

import "time"

// Player represents a guild member
type Player struct {
	ID           int    `json:"id"`
	CharacterName string `json:"character_name"`
	Class        string `json:"class"`
	CurrentBonus int    `json:"current_bonus"`
}

// Item represents a WoW item
type Item struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	ItemID    int    `json:"item_id"` // e.g., Wowhead ID
}

// RaidResult represents the combined data from CSV and Turtlogs
type RaidResult struct {
	RaidID       string
	RaidDate     time.Time
	Reserves     []ReserveEntry
	Drops        []LootDrop
}

type ReserveEntry struct {
	PlayerName string
	ItemName   string
}

type LootDrop struct {
	ItemName   string
	WinnerName string
}
