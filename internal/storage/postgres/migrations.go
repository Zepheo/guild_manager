package postgres

import (
	"database/sql"
	"embed"
	"fmt"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

func RunMigrations(db *sql.DB) error {
	// Simple logic: read your .sql files and execute them
	// In a larger app, you'd use a library like 'golang-migrate'
	entries, _ := migrationFiles.ReadDir("migrations")
	for _, entry := range entries {
		content, _ := migrationFiles.ReadFile("migrations/" + entry.Name())
		_, err := db.Exec(string(content))
		if err != nil {
			return fmt.Errorf("failed migration %s: %w", entry.Name(), err)
		}
	}
	return nil
}
