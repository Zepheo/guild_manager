package domain

import "time"

// Player represents a guild member
type Player struct {
	ID            int    `json:"id"`
	CharacterName string `json:"character_name"`
	Class         string `json:"class"`
}

// Item represents a WoW item
type Item struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// RaidResult represents the combined data from CSV and Turtlogs
type RaidResult struct {
	RaidID    int
	RaidDate  time.Time
	Attendees []Player
	Reserves  []ReserveEntry
	Drops     []LootDrop
}

type ReserveEntry struct {
	PlayerName string
	ItemName   string
}

type LootDrop struct {
	ItemName   string
	WinnerName string
}
