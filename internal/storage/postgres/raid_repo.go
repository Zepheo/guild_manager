package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"github.com/zepheo/guild_manager/internal/domain"
)

type PostgresRaidRepo struct {
	Db *sql.DB
}

func Connect(connStr string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error connecting to the database: %w", err)
	}

	return db, nil
}

func (r *PostgresRaidRepo) GetLastReserve(ctx context.Context, charName string, raid domain.Raid) (string, error) {
	query := `
		SELECT i.name
		FROM reserves r
		JOIN characters c ON r.character_id = c.id
		JOIN items i ON r.item_id = i.id
		JOIN raids rd ON r.raid_id = rd.id
		WHERE c.name = $1
		AND rd.raid = $2
		ORDER BY rd.raid_date DESC
		LIMIT 1`

	var itemName string
	err := r.Db.QueryRowContext(ctx, query, charName, raid.String()).Scan(&itemName)
	if err == sql.ErrNoRows {
		return "", nil // New player, no previous reserve
	}
	return itemName, err
}

func (r *PostgresRaidRepo) GetBulkPlayerBonuses(ctx context.Context, charNames []string, raid domain.Raid) (map[string]map[string]int, error) {
	query := `
	WITH RELEVANT_HISTORY AS (
		SELECT
			c.name as char_name,
			i.name as item_name,
			rd.raid_date,
			EXISTS(SELECT 1 FROM reserves res2 WHERE res2.raid_id = rd.id AND res2.character_id = c.id AND res2.item_id = i.id) as is_reserved,
			EXISTS(SELECT 1 FROM loot_history lh WHERE lh.raid_id = rd.id AND lh.item_id = i.id) as item_dropped,
			EXISTS(SELECT 1 FROM loot_history lh WHERE lh.raid_id = rd.id AND lh.winner_id = c.id AND lh.item_id = i.id) as player_won
		FROM raids rd
		CROSS JOIN (SELECT id, name FROM characters WHERE name = ANY($1)) c
		CROSS JOIN (SELECT id, name FROM items) i
		WHERE rd.raid = $2
	),
	STREAKS AS (
		SELECT
			char_name,
			item_name,
			is_reserved,
			item_dropped,
			player_won,
			-- Create a grouping factor that increments every time a "reset" happens
			SUM(CASE WHEN NOT is_reserved OR player_won THEN 1 ELSE 0 END)
				OVER (PARTITION BY char_name, item_name ORDER BY raid_date ASC) as streak_id
		FROM RELEVANT_HISTORY
	)
	SELECT char_name, item_name, SUM(20) as total_bonus
	FROM STREAKS
	WHERE is_reserved = TRUE AND item_dropped = TRUE AND player_won = FALSE
	-- We only care about the CURRENT streak (the rows after the last reset)
	AND streak_id = (SELECT MAX(streak_id) FROM STREAKS s2 WHERE s2.char_name = STREAKS.char_name AND s2.item_name = STREAKS.item_name)
	GROUP BY char_name, item_name
	HAVING SUM(20) > 0;`

	rows, err := r.Db.QueryContext(ctx, query, pq.Array(charNames), raid.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make(map[string]map[string]int)
	for rows.Next() {
		var pName, iName string
		var bonus int
		if err := rows.Scan(&pName, &iName, &bonus); err != nil {
			return nil, err
		}
		if results[pName] == nil {
			results[pName] = make(map[string]int)
		}
		results[pName][iName] = bonus
	}
	return results, nil
}

func (r *PostgresRaidRepo) GetPlayerBonuses(ctx context.Context, charName string, raid domain.Raid) (map[string]int, error) {
	currentReserves, err := r.getCurrentReserves(ctx, charName, raid)
	if err != nil {
		return nil, err
	}

	bonuses := make(map[string]int)
	for _, itemName := range currentReserves {
		itemBonus, err := r.calculateBonusForItem(ctx, charName, itemName, raid)
		if err != nil {
			fmt.Printf("error calculating bonus: %v\n", err)
			bonuses[itemName] = 0
			continue
		}
		bonuses[itemName] = itemBonus
	}

	return bonuses, nil
}

func (r *PostgresRaidRepo) RecordRaid(ctx context.Context, raid *domain.RaidResult) error {
	tx, err := r.Db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx,
		"INSERT INTO raids (id, raid_date, raid) VALUES ($1, $2, $3) ON CONFLICT (id) DO NOTHING",
		raid.RaidID, raid.RaidDate, raid.Raid.String())
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return tx.Commit()
	}

	for _, char := range raid.Attendees {
		_, err = tx.ExecContext(ctx,
			"INSERT INTO characters (id, name, class) VALUES ($1, $2, $3) ON CONFLICT (id) DO NOTHING",
			char.ID, char.CharacterName, char.Class)
		if err != nil {
			return fmt.Errorf("error on character insert: %v", err)
		}
		_, err = tx.ExecContext(ctx,
			"INSERT INTO attendance (raid_id, character_id) VALUES ($1, $2) ON CONFLICT DO NOTHING",
			raid.RaidID, char.ID)
		if err != nil {
			return fmt.Errorf("error on attendance insert: %v", err)
		}
	}

	for _, res := range raid.Reserves {
		tx.ExecContext(ctx, "INSERT INTO items (name) VALUES ($1) ON CONFLICT (name) DO NOTHING", res.ItemName)

		_, err = tx.ExecContext(ctx, `
			INSERT INTO reserves (character_id, raid_id, item_id)
			SELECT c.id, $1, i.id FROM characters c, items i
			WHERE c.name = $2 AND i.name = $3`,
			raid.RaidID, res.PlayerName, res.ItemName)
		if err != nil {
			return err
		}
	}

	for _, drop := range raid.Drops {
		_, err := tx.ExecContext(ctx, `INSERT INTO items (name) VALUES ($1) OFF CONFLICT (name) DO NOTHING`, drop.ItemName)
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `
			INSERT INTO loot_history (raid_id, item_id, winner_id, is_reserve_win)
			SELECT $1, i.id, c.id, EXISTS (
				SELECT 1 from reserves r
				WHERE r.raid_id = $1
				AND r.character_id = c.id
				AND r.item_id = i.id
			)
			FROM items i, characters c
			WHERE i.name = $2 AND c.name = $3`,
			raid.RaidID, drop.ItemName, drop.WinnerName)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *PostgresRaidRepo) calculateBonusForItem(ctx context.Context, charName string, itemName string, raid domain.Raid) (int, error) {
	query := `
		SELECT
			rd.id,
			EXISTS(SELECT 1 FROM reserves res2 JOIN items i2 ON res2.item_id = i2.id
			       WHERE res2.raid_id = rd.id AND res2.character_id = c.id AND i2.name = $2) as is_reserved,
			EXISTS(SELECT 1 FROM loot_history lh JOIN items i3 ON lh.item_id = i3.id
			       WHERE lh.raid_id = rd.id AND i3.name = $2) as item_dropped,
			EXISTS(SELECT 1 FROM loot_history lh JOIN items i4 ON lh.item_id = i4.id
			       WHERE lh.raid_id = rd.id AND lh.winner_id = c.id AND i4.name = $2) as player_won
		FROM raids rd
		CROSS JOIN characters c
		WHERE c.name = $1
		AND rd.raid = $3
		ORDER BY rd.raid_date ASC`

	rows, err := r.Db.QueryContext(ctx, query, charName, itemName, raid.String())
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	bonus := 0
	for rows.Next() {
		var raidID string
		var isReserved, itemDropped, playerWon bool
		rows.Scan(&raidID, &isReserved, &itemDropped, &playerWon)

		if !isReserved || playerWon {
			bonus = 0
			continue
		}

		if itemDropped && !playerWon {
			bonus += 20
		}
	}
	return bonus, nil
}

func (r *PostgresRaidRepo) getCurrentReserves(ctx context.Context, charName string, raid domain.Raid) ([]string, error) {
	query := `
		WITH latest_raid AS (
			SELECT r.raid_id
			FROM reserves r
			JOIN characters c ON r.character_id = c.id
			JOIN raids rd ON r.raid_id = rd.id
			WHERE c.name = $1
			AND rd.raid = $2
			ORDER BY rd.raid_date DESC
			LIMIT 1
		)
		SELECT i.name
		FROM reserves res
		JOIN latest_raid lr ON res.raid_id = lr.raid_id
		JOIN characters c ON res.character_id = c.id
		JOIN items i ON res.item_id = i.id
		WHERE c.name = $1`

	rows, err := r.Db.QueryContext(ctx, query, charName, raid.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		items = append(items, name)
	}
	return items, nil
}
