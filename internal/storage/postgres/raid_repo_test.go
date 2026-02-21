package postgres

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zepheo/guild_manager/internal/domain"
)

func TestGetLastReserve(t *testing.T) {
	// 1. Setup mock DB
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := &PostgresRaidRepo{Db: db}
	charName := "Zepheo"

	// 2. Define expectations for the query
	rows := sqlmock.NewRows([]string{"name"}).AddRow("Ashkandi")
	mock.ExpectQuery("SELECT i.name FROM reserves r").
		WithArgs(charName).
		WillReturnRows(rows)

	// 3. Execute the function
	item, err := repo.GetLastReserve(context.Background(), charName)

	// 4. Assertions
	if err != nil {
		t.Errorf("error was not expected: %s", err)
	}
	if item != "Ashkandi" {
		t.Errorf("expected Ashkandi, got %s", item)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestProcessRaidUpdate_Transaction(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error opening mock: %s", err)
	}
	defer db.Close()

	repo := &PostgresRaidRepo{Db: db}
	ctx := context.Background()

	// Expect a transaction to start
	mock.ExpectBegin()

	// Expect the UPDATE statement
	mock.ExpectExec("UPDATE characters SET current_bonus").
		WithArgs(20, "Zepheo").
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Expect the INSERT into history
	mock.ExpectExec("INSERT INTO bonus_history").
		WithArgs(20, "Raid Drop", "Zepheo").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Expect the transaction to commit
	mock.ExpectCommit()

	err = repo.ProcessRaidUpdate(ctx, "Zepheo", 20, "Raid Drop")

	if err != nil {
		t.Errorf("error was not expected: %s", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("transaction expectations failed: %s", err)
	}
}

func TestGetPlayerBonus(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error opening mock: %s", err)
	}
	defer db.Close()

	repo := &PostgresRaidRepo{Db: db}
	charName := "Zepheo"

	// Test Case: Player Exists
	mock.ExpectQuery("SELECT current_bonus FROM characters").
		WithArgs(charName).
		WillReturnRows(sqlmock.NewRows([]string{"current_bonus"}).AddRow(40))

	bonus, err := repo.GetPlayerBonus(context.Background(), charName)
	if err != nil || bonus != 40 {
		t.Errorf("Expected bonus 40, got %d (err: %v)", bonus, err)
	}

	// Test Case: New Player (No Rows)
	mock.ExpectQuery("SELECT current_bonus FROM characters").
		WithArgs("NewGuy").
		WillReturnError(sql.ErrNoRows)

	bonus, err = repo.GetPlayerBonus(context.Background(), "NewGuy")
	if err != nil || bonus != 0 {
		t.Errorf("Expected default bonus 0 for new player, got %d", bonus)
	}
}

func TestUpdatePlayerBonus(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error opening mock: %s", err)
	}
	defer db.Close()

	repo := &PostgresRaidRepo{Db: db}

	mock.ExpectExec("UPDATE characters SET current_bonus").
		WithArgs(60, "Zepheo").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.UpdatePlayerBonus(context.Background(), "Zepheo", 60, "Manual adjustment")
	if err != nil {
		t.Errorf("UpdatePlayerBonus failed: %v", err)
	}
}

func TestRecordRaid(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error opening mock: %s", err)
	}
	defer db.Close()

	repo := &PostgresRaidRepo{Db: db}
	raidTime := time.Now()
	raid := &domain.RaidResult{
		RaidID:   "log_123",
		RaidDate: raidTime,
		Reserves: []domain.ReserveEntry{
			{PlayerName: "Zepheo", ItemName: "Ashkandi"},
		},
	}

	// 1. Expect Transaction Begin
	mock.ExpectBegin()

	// 2. Expect Raid Metadata Insert
	mock.ExpectExec("INSERT INTO raids").
		WithArgs(raid.RaidID, raid.RaidDate).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// 3. Expect Reserves Loop Insert
	mock.ExpectExec("INSERT INTO reserves").
		WithArgs(raid.RaidID, "Zepheo", "Ashkandi").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// 4. Expect Transaction Commit
	mock.ExpectCommit()

	err = repo.RecordRaid(context.Background(), raid)
	if err != nil {
		t.Errorf("RecordRaid failed: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}
