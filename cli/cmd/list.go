package main

import (
	"context"
	"fmt"
	"time"

	"github.com/stormlightlabs/skypanel/cli/internal/registry"
	"github.com/stormlightlabs/skypanel/cli/internal/setup"
	"github.com/stormlightlabs/skypanel/cli/internal/store"
	"github.com/stormlightlabs/skypanel/cli/internal/ui"
	"github.com/urfave/cli/v3"
)

func ListAction(ctx context.Context, cmd *cli.Command) error {
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

func ListCommand() *cli.Command {
	return &cli.Command{
		Name:   "list",
		Usage:  "List all feeds",
		Action: ListAction,
	}
}
