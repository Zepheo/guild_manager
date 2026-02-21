package csv

import (
	"strings"
	"testing"
)

func TestParseRaidResCSV(t *testing.T) {
	csvData := `Item,Boss,Attendee,Class,Specialization,SR+
"Lawbringer Spaulders","Baron Geddon",Iskus,Paladin,Protection,0
"Molten Corehound",Ragnaros,Iskus,Paladin,Protection,0
"Molten Emberstone","Basalthar & Smoldaris",Garagho,Hunter,Marksmanship,0
"Quick Strike Ring",Shared,Garagho,Hunter,Marksmanship,0`

	p := NewParser()
	reserves, err := p.ParseRaidResCSV(strings.NewReader(csvData))

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(reserves) != 4 {
		t.Errorf("Expected 4 reserves, got %d", len(reserves))
	}

	if reserves[0].PlayerName != "Iskus" || reserves[0].ItemName != "Lawbringer Spaulders" {
		t.Errorf("Unexpected first entry: %+v", reserves[0])
	}
}
