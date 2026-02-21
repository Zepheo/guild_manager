package postgres

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/zepheo/guild_manager/internal/domain"
)

var testRepo *PostgresRaidRepo

func TestMain(m *testing.M) {
	ctx := context.Background()

	// 1. Define the Postgres container
	dbName := "sr_plus_test"
	dbUser := "user"
	dbPass := "pass"

	container, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:17-alpine"),
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPass),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	if err != nil {
		panic(err)
	}

	// 2. Get the dynamically assigned connection string
	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		panic(err)
	}
	db, err := Connect(connStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	if err := RunMigrations(db); err != nil {
		panic(err)
	}

	testRepo = &PostgresRaidRepo{Db: db}

	// Run migrations for the test DB [cite: 1]
	RunMigrations(db)

	code := m.Run()

	db.Close()
	os.Exit(code)
}

func TestIntegration_RecordRaidAndFetch(t *testing.T) {
	ctx := context.Background()

	// 1. Seed necessary data (Characters and Items)
	_, _ = testRepo.Db.Exec("INSERT INTO characters (name) VALUES ('Zepheo') ON CONFLICT DO NOTHING")
	_, _ = testRepo.Db.Exec("INSERT INTO items (name) VALUES ('Ashkandi') ON CONFLICT DO NOTHING")

	// 2. Test RecordRaid
	raid := &domain.RaidResult{
		RaidID:   "integration_test_log",
		RaidDate: time.Now(),
		Reserves: []domain.ReserveEntry{
			{PlayerName: "Zepheo", ItemName: "Ashkandi"},
		},
	}

	err := testRepo.RecordRaid(ctx, raid)
	if err != nil {
		t.Fatalf("Failed to record raid: %v", err)
	}

	// 3. Verify GetLastReserve returns the correct item
	item, err := testRepo.GetLastReserve(ctx, "Zepheo")
	if err != nil {
		t.Errorf("Error getting last reserve: %v", err)
	}
	if item != "Ashkandi" {
		t.Errorf("Expected Ashkandi, got %s", item)
	}
}
