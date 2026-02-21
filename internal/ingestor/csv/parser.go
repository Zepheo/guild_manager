package csv

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"

	"github.com/zepheo/guild_manager/internal/domain"
)

// Parser handles the specific format exported by raidres.top
type Parser struct {
	// We can add configuration here if the CSV format changes
}

func NewParser() *Parser {
	return &Parser{}
}

// ParseRaidResCSV converts the raidres.top CSV into a list of domain.ReserveEntry
func (p *Parser) ParseRaidResCSV(r io.Reader) ([]domain.ReserveEntry, error) {
	reader := csv.NewReader(r)

	// raidres.top usually uses commas, but some exports use semicolons.
	// We can add a check or keep it standard.
	reader.Comma = ','

	// Read the header row first
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	// Map columns to indices to stay flexible if they change the export order
	colMap := make(map[string]int)
	for i, name := range header {
		colMap[strings.ToLower(strings.TrimSpace(name))] = i
	}

	// Common column names in these tools: "player", "character", "item", "reserve"
	nameIdx, nameOk := colMap["player"] // or "character"
	if !nameOk {
		nameIdx, _ = colMap["character"]
	}
	itemIdx, itemOk := colMap["item"]

	if !itemOk {
		return nil, fmt.Errorf("required columns (player/item) not found in CSV")
	}

	var entries []domain.ReserveEntry

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		entry := domain.ReserveEntry{
			PlayerName: strings.TrimSpace(record[nameIdx]),
			ItemName:   strings.TrimSpace(record[itemIdx]),
		}

		// Only add if both fields are present
		if entry.PlayerName != "" && entry.ItemName != "" {
			entries = append(entries, entry)
		}
	}

	return entries, nil
}
