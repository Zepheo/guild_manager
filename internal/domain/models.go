package domain

import (
	"encoding/json"
	"time"
)

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

type Raid int

const (
	MC Raid = iota
	ES
	BWL
)

var raidName = map[Raid]string{
	MC:  "mc",
	ES:  "es",
	BWL: "bwl",
}

func (r Raid) String() string {
	return raidName[r]
}

func (r *Raid) UnmarshalJSON(data []byte) error {
	var mapId int
	if err := json.Unmarshal(data, &mapId); err != nil {
		return err
	}
	switch mapId {
	case 469:
		*r = BWL
	case 807:
		*r = ES
	default:
		*r = MC
	}
	return nil
}

// RaidResult represents the combined data from CSV and Turtlogs
type RaidResult struct {
	RaidID    int
	RaidDate  time.Time
	Raid      Raid
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
