package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/stormlightlabs/skypanel/cli/internal/config"
	"github.com/stormlightlabs/skypanel/cli/internal/store"
	"github.com/stormlightlabs/skypanel/cli/internal/ui"
	"github.com/urfave/cli/v3"
)

func SetupAction(ctx context.Context, cmd *cli.Command) error {
	logger := ui.GetLogger()

	ui.Titleln("Setup: Initializing persistence layer")
	fmt.Println()

	configDir, err := config.GetConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get config directory: %w", err)
	}

	dbPath, err := config.GetCacheDB()
	if err != nil {
		return fmt.Errorf("failed to get database path: %w", err)
	}

	ui.Infoln("Config directory: %s", configDir)
	ui.Infoln("Database path: %s", dbPath)
	fmt.Println()

	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		ui.Infoln("Creating config directory...")
		if err := os.MkdirAll(configDir, 0700); err != nil {
			logger.Error("Failed to create config directory", "error", err)
			return err
		}
		ui.Successln("Config directory created")
	} else {
		ui.Successln("Config directory exists")
	}

	dbExists := true
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		dbExists = false
		ui.Infoln("Database does not exist, will be created")
	} else {
		ui.Successln("Database file exists")
	}

	fmt.Println()

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		logger.Error("Failed to open database", "error", err)
		return err
	}
	defer db.Close()

	statusBefore, err := store.GetMigrationStatus(db)
	if err != nil && !dbExists {
		logger.Debug("Migration status check returned error (expected for new database)", "error", err)
		statusBefore = &store.MigrationStatus{CurrentVersion: 0, LatestVersion: 0, PendingCount: 0}
	} else if err != nil {
		logger.Error("Failed to check migration status", "error", err)
		return err
	}

	if statusBefore.IsUpToDate && dbExists {
		ui.Successln("Database is up to date (v%d)", statusBefore.CurrentVersion)
		return nil
	}

	ui.Infoln("Running migrations...")
	if err := store.RunMigrations(db); err != nil {
		logger.Error("Failed to run migrations", "error", err)
		return err
	}

	statusAfter, err := store.GetMigrationStatus(db)
	if err != nil {
		logger.Error("Failed to verify migration status", "error", err)
		return err
	}

	fmt.Println()
	ui.Successln("Setup complete!")
	ui.Infoln("Database version: v%d", statusAfter.CurrentVersion)
	ui.Infoln("Migrations applied: %d", statusAfter.CurrentVersion-statusBefore.CurrentVersion)

	return nil
}

func SetupCommand() *cli.Command {
	return &cli.Command{
		Name:  "setup",
		Usage: "Initialize the persistence layer (database and config)",
		Description: `Initialize the skycli persistence layer by creating:
   - Config directory at ~/.skycli
   - SQLite database at ~/.skycli/cache.db
   - Running all database migrations

   This command is idempotent and safe to run multiple times.
   It will show the current state and only make necessary changes.`,
		Action: SetupAction,
	}
}
