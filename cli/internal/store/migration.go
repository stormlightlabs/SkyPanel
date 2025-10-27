package store

import (
	"database/sql"
	"embed"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// RunMigrations executes all pending up migrations in order.
// Creates a schema_migrations table to track applied migrations.
func RunMigrations(db *sql.DB) error {
	if err := createMigrationsTable(db); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	applied, err := getAppliedMigrations(db)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	migrations, err := loadMigrations("up")
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	for _, m := range migrations {
		if applied[m.Version] {
			continue
		}

		if err := executeMigration(db, m); err != nil {
			return fmt.Errorf("failed to execute migration %d: %w", m.Version, err)
		}

		if err := recordMigration(db, m.Version); err != nil {
			return fmt.Errorf("failed to record migration %d: %w", m.Version, err)
		}
	}

	return nil
}

// Rollback executes down migrations back to the specified version.
// If version is 0, rolls back all migrations.
func Rollback(db *sql.DB, targetVersion int) error {
	applied, err := getAppliedMigrations(db)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	migrations, err := loadMigrations("down")
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	for i := len(migrations) - 1; i >= 0; i-- {
		m := migrations[i]
		if !applied[m.Version] || m.Version <= targetVersion {
			continue
		}

		if err := executeMigration(db, m); err != nil {
			return fmt.Errorf("failed to rollback migration %d: %w", m.Version, err)
		}

		if err := removeMigration(db, m.Version); err != nil {
			return fmt.Errorf("failed to remove migration record %d: %w", m.Version, err)
		}
	}

	return nil
}

type migration struct {
	Version int
	Name    string
	SQL     string
}

// loadMigrations reads all migration files of the specified direction (up/down)
func loadMigrations(direction string) ([]migration, error) {
	entries, err := migrationFiles.ReadDir("migrations")
	if err != nil {
		return nil, err
	}

	var migrations []migration
	suffix := "." + direction + ".sql"

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), suffix) {
			continue
		}

		content, err := migrationFiles.ReadFile("migrations/" + entry.Name())
		if err != nil {
			return nil, err
		}

		parts := strings.Split(entry.Name(), "_")
		if len(parts) < 2 {
			continue
		}

		version, err := strconv.Atoi(parts[0])
		if err != nil {
			continue
		}

		migrations = append(migrations, migration{
			Version: version,
			Name:    strings.TrimSuffix(entry.Name(), suffix),
			SQL:     string(content),
		})
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

// createMigrationsTable creates the schema_migrations tracking table
func createMigrationsTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`
	_, err := db.Exec(query)
	return err
}

// getAppliedMigrations returns a map of applied migration versions
func getAppliedMigrations(db *sql.DB) (map[int]bool, error) {
	rows, err := db.Query("SELECT version FROM schema_migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[int]bool)
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = true
	}

	return applied, rows.Err()
}

// executeMigration runs a single migration SQL
func executeMigration(db *sql.DB, m migration) error {
	_, err := db.Exec(m.SQL)
	return err
}

// recordMigration adds a migration to the schema_migrations table
func recordMigration(db *sql.DB, version int) error {
	_, err := db.Exec("INSERT INTO schema_migrations (version) VALUES (?)", version)
	return err
}

// removeMigration removes a migration from the schema_migrations table
func removeMigration(db *sql.DB, version int) error {
	_, err := db.Exec("DELETE FROM schema_migrations WHERE version = ?", version)
	return err
}
