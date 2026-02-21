package domain

// CalculateNewBonus determines the next SR+ score for a player.
func CalculateNewBonus(currentBonus int, lastItemReserved, currentItemReserved string, itemDropped bool, playerWon bool) int {
	// Rule: If they change their reserved item, bonus is wiped.
	if lastItemReserved != "" && lastItemReserved != currentItemReserved {
		return 0
	}

	// Rule: If they win the item, bonus is wiped regardless of how high it was.
	if playerWon {
		return 0
	}

	// Rule: If their reserved item dropped but they didn't win, they get +20.
	if itemDropped && !playerWon {
		return currentBonus + 20
	}

	// If the item didn't drop and they didn't change their reserve,
	// the bonus carries over as is.
	return currentBonus
}
