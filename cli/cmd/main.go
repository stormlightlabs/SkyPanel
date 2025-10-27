package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stormlightlabs/skypanel/cli/internal/config"
	"github.com/stormlightlabs/skypanel/cli/internal/export"
	"github.com/stormlightlabs/skypanel/cli/internal/imports"
	"github.com/stormlightlabs/skypanel/cli/internal/registry"
	"github.com/stormlightlabs/skypanel/cli/internal/setup"
	"github.com/stormlightlabs/skypanel/cli/internal/store"
	"github.com/stormlightlabs/skypanel/cli/internal/ui"
	"github.com/urfave/cli/v3"
)

func main() {
	ui.InitLogger(log.InfoLevel)
	logger := ui.GetLogger()

	ctx := context.Background()
	reg := registry.Get()

	if err := reg.Init(ctx); err != nil {
		logger.Fatal("Failed to initialize registry", "error", err)
	}
	defer reg.Close()

	app := &cli.Command{
		Name:    "skycli",
		Usage:   "A companion CLI tool for your Bluesky feed ecosystem",
		Version: "0.1.0",
		Commands: []*cli.Command{
			{
				Name:  "setup",
				Usage: "Initialize the persistence layer (database and config)",
				Description: `Initialize the skycli persistence layer by creating:
   - Config directory at ~/.skycli
   - SQLite database at ~/.skycli/cache.db
   - Running all database migrations

   This command is idempotent and safe to run multiple times.
   It will show the current state and only make necessary changes.`,
				Action: setupAction,
			},
			{
				Name:  "login",
				Usage: "Authenticate with Bluesky",
				Description: `Authenticate with Bluesky using one of two methods:

   1. Direct credentials via flags:
      skycli login --handle @user.bsky.social --password your-app-password

   2. Credentials from an env file:
      skycli login --file /path/to/.env

   The env file should contain:
      BLUESKY_HANDLE=your.handle.bsky.social
      BLUESKY_PASSWORD=your-app-password

   File paths can be relative or absolute.`,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "file",
						Aliases: []string{"f"},
						Usage:   "Path to env file containing BLUESKY_HANDLE and BLUESKY_PASSWORD",
					},
					&cli.StringFlag{
						Name:    "handle",
						Aliases: []string{"u"},
						Usage:   "Your Bluesky handle (e.g., @user.bsky.social)",
					},
					&cli.StringFlag{
						Name:    "password",
						Aliases: []string{"p"},
						Usage:   "Your app password",
					},
				},
				Action: loginAction,
			},
			{
				Name:   "status",
				Usage:  "Show current session status",
				Action: statusAction,
			},
			{
				Name:   "list",
				Usage:  "List all feeds",
				Action: listAction,
			},
			{
				Name:      "view",
				Usage:     "View posts from a feed",
				ArgsUsage: "<feed-id-or-uri>",
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:    "size",
						Aliases: []string{"s"},
						Usage:   "Number of posts to display",
						Value:   25,
					},
				},
				Action: viewAction,
			},
			{
				Name:      "export",
				Usage:     "Export posts from a feed to file",
				ArgsUsage: "<feed-id>",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "format",
						Aliases: []string{"f"},
						Usage:   "Export format: json, csv, or txt",
						Value:   "json",
					},
					&cli.IntFlag{
						Name:    "size",
						Aliases: []string{"s"},
						Usage:   "Number of posts to export",
						Value:   25,
					},
				},
				Action: exportAction,
			},
		},
	}

	if err := app.Run(ctx, os.Args); err != nil {
		logger.Fatalf("Command failed with error: %v", err)
	}
}

func loginAction(ctx context.Context, cmd *cli.Command) error {
	if err := setup.EnsurePersistenceReady(ctx); err != nil {
		return fmt.Errorf("persistence layer not ready: %w", err)
	}

	logger := ui.GetLogger()
	reg := registry.Get()

	var handle, password string
	filePath := cmd.String("file")

	if filePath != "" {
		env, err := imports.ParseEnvFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to parse env file: %w", err)
		}

		handle = env["BLUESKY_HANDLE"]
		password = env["BLUESKY_PASSWORD"]

		if handle == "" {
			return fmt.Errorf("BLUESKY_HANDLE not found in env file")
		}
		if password == "" {
			return fmt.Errorf("BLUESKY_PASSWORD not found in env file")
		}
	} else {
		handle = cmd.String("handle")
		password = cmd.String("password")

		if handle == "" || password == "" {
			return fmt.Errorf("either --file or both --handle and --password are required")
		}
	}

	logger.Info("Authenticating with Bluesky", "handle", handle)

	service, err := reg.GetService()
	if err != nil {
		return fmt.Errorf("failed to get service: %w", err)
	}

	credentials := map[string]string{
		"identifier": handle,
		"password":   password,
	}

	if err := service.Authenticate(ctx, credentials); err != nil {
		logger.Error("Authentication failed", "error", err)
		return err
	}

	sessionRepo, err := reg.GetSessionRepo()
	if err != nil {
		return fmt.Errorf("failed to get session repository: %w", err)
	}

	if err := sessionRepo.UpdateTokens(ctx, service.GetAccessToken(), service.GetRefreshToken()); err != nil {
		logger.Warn("Failed to save session tokens", "error", err)
	}

	ui.Successln("Successfully authenticated as %s", handle)
	return nil
}

func statusAction(ctx context.Context, cmd *cli.Command) error {
	if err := setup.EnsurePersistenceReady(ctx); err != nil {
		return fmt.Errorf("persistence layer not ready: %w", err)
	}

	logger := ui.GetLogger()
	reg := registry.Get()

	sessionRepo, err := reg.GetSessionRepo()
	if err != nil {
		return fmt.Errorf("failed to get session repository: %w", err)
	}

	if !sessionRepo.HasValidSession(ctx) {
		ui.Infoln("Not authenticated. Run 'skycli login' to authenticate.")
		return nil
	}

	session, err := sessionRepo.List(ctx)
	if err != nil {
		logger.Error("Failed to get session", "error", err)
		return err
	}

	if len(session) > 0 {
		if s, ok := session[0].(*store.SessionModel); ok {
			ui.Titleln("Session Status")
			ui.Infoln("Handle: %s", s.Handle)
			ui.Infoln("Service: %s", s.ServiceURL)
			ui.Successln("Authenticated")
		}
	}

	return nil
}

func listAction(ctx context.Context, cmd *cli.Command) error {
	if err := setup.EnsurePersistenceReady(ctx); err != nil {
		return fmt.Errorf("persistence layer not ready: %w", err)
	}

	logger := ui.GetLogger()
	reg := registry.Get()

	feedRepo, err := reg.GetFeedRepo()
	if err != nil {
		return fmt.Errorf("failed to get feed repository: %w", err)
	}

	feeds, err := feedRepo.List(ctx)
	if err != nil {
		logger.Error("Failed to list feeds", "error", err)
		return err
	}

	if len(feeds) == 0 {
		ui.Infoln("No feeds found.")
		return nil
	}

	ui.Titleln("Feeds")
	fmt.Println()

	for _, model := range feeds {
		if feed, ok := model.(*store.FeedModel); ok {
			ui.Subtitleln("ID: %s", feed.ID())
			ui.Infoln("  Name: %s", feed.Name)
			ui.Infoln("  Source: %s", feed.Source)
			ui.Infoln("  Local: %t", feed.IsLocal)
			ui.Infoln("  Created: %s", feed.CreatedAt().Format(time.RFC3339))
			fmt.Println()
		}
	}

	ui.Successln("Total: %d feed(s)", len(feeds))
	return nil
}

func viewAction(ctx context.Context, cmd *cli.Command) error {
	if err := setup.EnsurePersistenceReady(ctx); err != nil {
		return fmt.Errorf("persistence layer not ready: %w", err)
	}

	logger := ui.GetLogger()
	reg := registry.Get()

	if cmd.Args().Len() == 0 {
		return fmt.Errorf("feed ID or URI required")
	}

	feedIdentifier := cmd.Args().First()
	size := cmd.Int("size")

	feedRepo, err := reg.GetFeedRepo()
	if err != nil {
		return fmt.Errorf("failed to get feed repository: %w", err)
	}

	postRepo, err := reg.GetPostRepo()
	if err != nil {
		return fmt.Errorf("failed to get post repository: %w", err)
	}

	var feedID string

	if _, err := uuid.Parse(feedIdentifier); err == nil {
		feedID = feedIdentifier
	} else {
		feeds, err := feedRepo.List(ctx)
		if err != nil {
			logger.Error("Failed to list feeds", "error", err)
			return err
		}

		found := false
		for _, model := range feeds {
			if feed, ok := model.(*store.FeedModel); ok {
				if feed.Source == feedIdentifier {
					feedID = feed.ID()
					found = true
					break
				}
			}
		}

		if !found {
			return fmt.Errorf("feed not found with identifier: %s", feedIdentifier)
		}
	}

	posts, err := postRepo.QueryByFeedID(ctx, feedID, size, 0)
	if err != nil {
		logger.Error("Failed to query posts", "error", err)
		return err
	}

	if len(posts) == 0 {
		ui.Infoln("No posts found for this feed.")
		return nil
	}

	totalCount, err := postRepo.CountByFeedID(ctx, feedID)
	if err != nil {
		logger.Warn("Failed to get total count", "error", err)
	}

	ui.Titleln("Posts for Feed: %s", feedID)
	fmt.Println()

	for i, post := range posts {
		ui.Subtitleln("[%d] %s", i+1, post.URI)
		ui.Infoln("  Author: %s", post.AuthorDID)
		text := post.Text
		if len(text) > 100 {
			text = text[:100] + "..."
		}
		ui.Infoln("  Text: %s", text)
		ui.Infoln("  Indexed: %s", post.IndexedAt.Format(time.RFC3339))
		fmt.Println()
	}

	ui.Successln("Showing %d of %d post(s)", len(posts), totalCount)
	return nil
}

func exportAction(ctx context.Context, cmd *cli.Command) error {
	if err := setup.EnsurePersistenceReady(ctx); err != nil {
		return fmt.Errorf("persistence layer not ready: %w", err)
	}

	logger := ui.GetLogger()
	reg := registry.Get()

	if cmd.Args().Len() == 0 {
		return fmt.Errorf("feed ID required")
	}

	feedID := cmd.Args().First()
	format := strings.ToLower(cmd.String("format"))
	size := cmd.Int("size")

	if format != "json" && format != "csv" && format != "txt" {
		return fmt.Errorf("invalid format: %s (must be json, csv, or txt)", format)
	}

	feedRepo, err := reg.GetFeedRepo()
	if err != nil {
		return fmt.Errorf("failed to get feed repository: %w", err)
	}

	postRepo, err := reg.GetPostRepo()
	if err != nil {
		return fmt.Errorf("failed to get post repository: %w", err)
	}

	_, err = feedRepo.Get(ctx, feedID)
	if err != nil {
		return fmt.Errorf("feed not found: %w", err)
	}

	posts, err := postRepo.QueryByFeedID(ctx, feedID, size, 0)
	if err != nil {
		logger.Error("Failed to query posts", "error", err)
		return err
	}

	if len(posts) == 0 {
		ui.Warningln("No posts found for this feed.")
		return nil
	}

	filename := fmt.Sprintf("feed_%s_%s.%s", feedID, time.Now().Format("2006-01-02"), format)

	switch format {
	case "json":
		err = export.ToJSON(filename, posts)
	case "csv":
		err = export.ToCSV(filename, posts)
	case "txt":
		err = export.ToTXT(filename, posts)
	}

	if err != nil {
		logger.Error("Failed to export", "error", err)
		return err
	}

	ui.Successln("Exported %d post(s) to %s", len(posts), filename)
	return nil
}

func setupAction(ctx context.Context, cmd *cli.Command) error {
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

	// Check if config directory exists
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

	// Check if database exists
	dbExists := true
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		dbExists = false
		ui.Infoln("Database does not exist, will be created")
	} else {
		ui.Successln("Database file exists")
	}

	fmt.Println()

	// Open database (creates if doesn't exist)
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		logger.Error("Failed to open database", "error", err)
		return err
	}
	defer db.Close()

	// Get migration status before running migrations
	statusBefore, err := store.GetMigrationStatus(db)
	if err != nil && !dbExists {
		// If database is new, this is expected
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

	// Run migrations
	ui.Infoln("Running migrations...")
	if err := store.RunMigrations(db); err != nil {
		logger.Error("Failed to run migrations", "error", err)
		return err
	}

	// Get migration status after running migrations
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
