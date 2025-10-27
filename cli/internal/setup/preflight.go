package setup

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stormlightlabs/skypanel/cli/internal/config"
	"github.com/stormlightlabs/skypanel/cli/internal/store"
)

// EnsurePersistenceReady validates the persistence layer from package [store] is ready for use.
// On first run, automatically creates directories, database, and runs migrations; on subsequent runs, performs fast validation checks only.
func EnsurePersistenceReady(ctx context.Context) error {
	configDir, err := config.GetConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get config directory: %w", err)
	}

	dbPath, err := config.GetCacheDB()
	if err != nil {
		return fmt.Errorf("failed to get database path: %w", err)
	}

	dirInfo, err := os.Stat(configDir)
	if os.IsNotExist(err) {
		if err := initializeFirstRun(ctx, configDir, dbPath); err != nil {
			return fmt.Errorf("first-run initialization failed: %w", err)
		}
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to check config directory: %w", err)
	}
	if !dirInfo.IsDir() {
		return fmt.Errorf("config path exists but is not a directory: %s", configDir)
	}

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		if err := initializeDatabase(ctx, dbPath); err != nil {
			return fmt.Errorf("database initialization failed: %w", err)
		}
		return nil
	}

	if err := validateMigrations(dbPath); err != nil {
		return fmt.Errorf("migration validation failed: %w", err)
	}

	return nil
}

// initializeFirstRun creates config directory and initializes database for first-time use
func initializeFirstRun(ctx context.Context, configDir, dbPath string) error {
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := initializeDatabase(ctx, dbPath); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	return nil
}

// initializeDatabase creates database file and runs all migrations
func initializeDatabase(_ context.Context, dbPath string) error {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}
	defer db.Close()

	if err := store.RunMigrations(db); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// validateMigrations performs a fast check that all migrations are applied
func validateMigrations(dbPath string) error {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	status, err := store.GetMigrationStatus(db)
	if err != nil {
		return fmt.Errorf("failed to check migration status: %w", err)
	}

	if !status.IsUpToDate {
		return fmt.Errorf(
			"database has %d pending migrations (current: v%d, latest: v%d). Run 'skycli setup' to update",
			status.PendingCount,
			status.CurrentVersion,
			status.LatestVersion,
		)
	}

	return nil
}
