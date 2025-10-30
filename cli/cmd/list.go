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

// ListPostsAction lists the authenticated user's own posts
func ListPostsAction(ctx context.Context, cmd *cli.Command) error {
	if err := setup.EnsurePersistenceReady(ctx); err != nil {
		return fmt.Errorf("persistence layer not ready: %w", err)
	}

	reg := registry.Get()

	service, err := reg.GetService()
	if err != nil {
		return fmt.Errorf("failed to get service: %w", err)
	}

	if !service.Authenticated() {
		return fmt.Errorf("not authenticated: run 'skycli login' first")
	}

	sessionRepo, err := reg.GetSessionRepo()
	if err != nil {
		return fmt.Errorf("failed to get session repository: %w", err)
	}

	did, err := sessionRepo.GetDid(ctx)
	if err != nil {
		return fmt.Errorf("failed to get user DID: %w", err)
	}

	limit := cmd.Int("limit")
	asJSON := cmd.Bool("json")

	logger.Debug("Fetching user's posts", "did", did, "limit", limit)

	response, err := service.GetAuthorFeed(ctx, did, limit, "")
	if err != nil {
		return fmt.Errorf("failed to fetch user posts: %w", err)
	}

	if asJSON {
		return ui.DisplayJSON(response)
	}

	ui.Titleln("Your Posts")
	ui.DisplayFeed(response.Feed, response.Cursor)
	return nil
}

// ListFeedsAction lists user's feeds (from local cache or refetch from API)
func ListFeedsAction(ctx context.Context, cmd *cli.Command) error {
	if err := setup.EnsurePersistenceReady(ctx); err != nil {
		return fmt.Errorf("persistence layer not ready: %w", err)
	}

	reg := registry.Get()

	refetch := cmd.Bool("refetch")
	asJSON := cmd.Bool("json")

	if refetch {
		service, err := reg.GetService()
		if err != nil {
			return fmt.Errorf("failed to get service: %w", err)
		}

		if !service.Authenticated() {
			return fmt.Errorf("not authenticated: run 'skycli login' first")
		}

		sessionRepo, err := reg.GetSessionRepo()
		if err != nil {
			return fmt.Errorf("failed to get session repository: %w", err)
		}

		handle, err := sessionRepo.GetHandle(ctx)
		if err != nil {
			return fmt.Errorf("failed to get user handle: %w", err)
		}

		logger.Debug("Refetching feeds from API", "handle", handle)

		// TODO: Implement GetUserFeeds in BlueskyService
		// For now, fall back to local feeds
		ui.Warningln("API refetch not yet implemented, showing local feeds")
	}

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

	if asJSON {
		return ui.DisplayJSON(feeds)
	}

	ui.Titleln("Your Feeds")
	fmt.Println()

	for i, model := range feeds {
		if feed, ok := model.(*store.FeedModel); ok {
			ui.Subtitleln("[%d] %s", i+1, feed.Name)
			ui.Infoln("  ID: %s", feed.ID())
			ui.Infoln("  Source: %s", feed.Source)
			ui.Infoln("  Local: %t", feed.IsLocal)
			ui.Infoln("  Created: %s", feed.CreatedAt().Format(time.RFC3339))
			fmt.Println()
		}
	}

	ui.Successln("Total: %d feed(s)", len(feeds))
	return nil
}

// ListCommand returns the list command with subcommands for posts and feeds
func ListCommand() *cli.Command {
	commonFlags := []cli.Flag{
		&cli.BoolFlag{
			Name:    "refetch",
			Aliases: []string{"r"},
			Usage:   "Refetch from API instead of using local cache",
		},
		&cli.BoolFlag{
			Name:    "json",
			Aliases: []string{"j"},
			Usage:   "Output raw JSON response",
		},
	}

	return &cli.Command{
		Name:  "list",
		Usage: "List user's posts or feeds",
		Commands: []*cli.Command{
			{
				Name:      "posts",
				Usage:     "List authenticated user's own posts",
				ArgsUsage: " ",
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:    "limit",
						Aliases: []string{"l"},
						Usage:   "Maximum number of posts to list",
						Value:   25,
					},
					&cli.BoolFlag{
						Name:    "json",
						Aliases: []string{"j"},
						Usage:   "Output raw JSON response",
					},
				},
				Action: ListPostsAction,
			},
			{
				Name:      "feeds",
				Usage:     "List user's feeds (local cache or refetch with -r)",
				ArgsUsage: " ",
				Flags:     commonFlags,
				Action:    ListFeedsAction,
			},
		},
		// Default action when no subcommand is provided
		Action: func(ctx context.Context, cmd *cli.Command) error {
			// Default to posts
			return ListPostsAction(ctx, cmd)
		},
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:    "limit",
				Aliases: []string{"l"},
				Usage:   "Maximum number of posts to list",
				Value:   25,
			},
			&cli.BoolFlag{
				Name:    "json",
				Aliases: []string{"j"},
				Usage:   "Output raw JSON response",
			},
		},
	}
}
