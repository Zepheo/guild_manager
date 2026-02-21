package domain

import "testing"

func TestCalculateNewBonus(t *testing.T) {
	tests := []struct {
		name                string
		currentBonus        int
		lastItemReserved    string
		currentItemReserved string
		itemDropped         bool
		playerWon           bool
		want                int
	}{
		{"Reset on Item Change", 40, "Hand of Ragnaros", "Ashkandi", false, false, 0},
		{"Reset on Win", 60, "Ashkandi", "Ashkandi", true, true, 0},
		{"Bonus +20 on Loss", 20, "Ashkandi", "Ashkandi", true, false, 40},
		{"No Change if Not Dropped", 40, "Ashkandi", "Ashkandi", false, false, 40},
		{"New Player Start", 0, "", "Ashkandi", false, false, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateNewBonus(tt.currentBonus, tt.lastItemReserved, tt.currentItemReserved, tt.itemDropped, tt.playerWon)
			if got != tt.want {
				t.Errorf("CalculateNewBonus() = %v, want %v", got, tt.want)
			}
		})
	}
}
