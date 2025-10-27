package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/stormlightlabs/skypanel/cli/internal/export"
	"github.com/stormlightlabs/skypanel/cli/internal/registry"
	"github.com/stormlightlabs/skypanel/cli/internal/setup"
	"github.com/stormlightlabs/skypanel/cli/internal/ui"
	"github.com/urfave/cli/v3"
)

func ExportCommand() *cli.Command {
	return &cli.Command{
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
		Action: ExportAction,
	}
}

func ExportAction(ctx context.Context, cmd *cli.Command) error {
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
