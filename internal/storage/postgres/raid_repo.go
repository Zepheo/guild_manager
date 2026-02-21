package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

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

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Verify connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error connecting to the database: %w", err)
	}

	return db, nil
}

func (r *PostgresRaidRepo) GetLastReserve(ctx context.Context, charName string) (string, error) {
	query := `
		SELECT i.name
		FROM reserves r
		JOIN characters c ON r.character_id = c.id
		JOIN items i ON r.item_id = i.id
		JOIN raids rd ON r.raid_id = rd.id
		WHERE c.name = $1
		ORDER BY rd.raid_date DESC
		LIMIT 1`

	var itemName string
	err := r.Db.QueryRowContext(ctx, query, charName).Scan(&itemName)
	if err == sql.ErrNoRows {
		return "", nil // New player, no previous reserve
	}
	return itemName, err
}

func (r *PostgresRaidRepo) GetPlayerBonus(ctx context.Context, charName string) (int, error) {
	var bonus int
	err := r.Db.QueryRowContext(ctx, "SELECT current_bonus FROM characters WHERE name = $1", charName).Scan(&bonus)
	if err == sql.ErrNoRows {
		return 0, nil // New players start at 0
	}
	return bonus, err
}

// Ensure the UpdatePlayerBonus method also exists to satisfy the interface
func (r *PostgresRaidRepo) UpdatePlayerBonus(ctx context.Context, charName string, newBonus int, reason string) error {
	_, err := r.Db.ExecContext(ctx, "UPDATE characters SET current_bonus = $1 WHERE name = $2", newBonus, charName)
	return err
}

func (r *PostgresRaidRepo) ProcessRaidUpdate(ctx context.Context, charName string, newBonus int, reason string) error {
	tx, err := r.Db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Update the character's current state
	_, err = tx.ExecContext(ctx,
		"UPDATE characters SET current_bonus = $1 WHERE name = $2",
		newBonus, charName)
	if err != nil {
		return err
	}

	// 2. Insert into history for manual audit/overrides
	_, err = tx.ExecContext(ctx,
		`INSERT INTO bonus_history (character_id, change_amount, reason)
		 SELECT id, $1, $2 FROM characters WHERE name = $3`,
		newBonus, reason, charName)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *PostgresRaidRepo) RecordRaid(ctx context.Context, raidResult *domain.RaidResult) error {
	tx, err := r.Db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Record the Raid metadata
	raidMetaQuery := `
		INSERT INTO raids (id, raid_date)
		VALUES ($1, $2)
		ON CONFLICT (id) DO NOTHING
	`
	_, err = tx.ExecContext(ctx, raidMetaQuery, raidResult.RaidID, raidResult.RaidDate)
	if err != nil {
		return fmt.Errorf("failed to record raid metadata: %w", err)
	}

	// 2. Record each specific reserve entry
	for _, res := range raidResult.Reserves {
		reservesQuery := `
			INSERT INTO reserves (character_id, raid_id, item_id)
			SELECT c.id, $1, i.id
			FROM characters c, items i
			WHERE c.name = $2 AND i.name = $3
		`
		_, err = tx.ExecContext(ctx, reservesQuery, raidResult.RaidID, res.PlayerName, res.ItemName)
		if err != nil {
			return fmt.Errorf("failed to record reserve for %s: %w", res.PlayerName, err)
		}
	}

	return tx.Commit()
}
