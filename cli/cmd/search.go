package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/stormlightlabs/skypanel/cli/internal/registry"
	"github.com/stormlightlabs/skypanel/cli/internal/setup"
	"github.com/stormlightlabs/skypanel/cli/internal/store"
	"github.com/stormlightlabs/skypanel/cli/internal/ui"
	"github.com/urfave/cli/v3"
)

// SearchUsersAction searches for users (actors) by query string
func SearchUsersAction(ctx context.Context, cmd *cli.Command) error {
	if err := setup.EnsurePersistenceReady(ctx); err != nil {
		return fmt.Errorf("persistence layer not ready: %w", err)
	}

	reg := registry.Get()

	if cmd.Args().Len() == 0 {
		return fmt.Errorf("search query required")
	}

	query := cmd.Args().First()
	limit := cmd.Int("limit")
	cursor := cmd.String("cursor")
	asJSON := cmd.Bool("json")

	service, err := reg.GetService()
	if err != nil {
		return fmt.Errorf("failed to get service: %w", err)
	}

	if !service.Authenticated() {
		return fmt.Errorf("not authenticated: run 'skycli login' first")
	}

	logger.Debug("Searching users", "query", query, "limit", limit, "cursor", cursor)

	result, err := service.SearchActors(ctx, query, limit, cursor)
	if err != nil {
		return fmt.Errorf("failed to search users: %w", err)
	}

	if asJSON {
		return ui.DisplayJSON(result)
	}

	if len(result.Actors) == 0 {
		ui.Infoln("No users found matching query: %s", query)
		return nil
	}

	ui.Titleln("Search Results: %s", query)
	fmt.Println()

	for i, actor := range result.Actors {
		ui.Subtitleln("[%d] @%s", i+1, actor.Handle)
		if actor.DisplayName != "" {
			ui.Infoln("  Name: %s", actor.DisplayName)
		}
		ui.Infoln("  DID: %s", actor.Did)
		if actor.Description != "" {
			desc := actor.Description
			if len(desc) > 100 {
				desc = desc[:100] + "..."
			}
			ui.Infoln("  Bio: %s", desc)
		}
		ui.Infoln("  Followers: %d | Following: %d | Posts: %d",
			actor.FollowersCount, actor.FollowsCount, actor.PostsCount)
		fmt.Println()
	}

	ui.Successln("Found %d user(s)", len(result.Actors))
	if result.Cursor != "" {
		ui.Infoln("Next cursor: %s", result.Cursor)
	}

	return nil
}

// SearchPostsAction searches for posts by query string
func SearchPostsAction(ctx context.Context, cmd *cli.Command) error {
	if err := setup.EnsurePersistenceReady(ctx); err != nil {
		return fmt.Errorf("persistence layer not ready: %w", err)
	}

	reg := registry.Get()

	if cmd.Args().Len() == 0 {
		return fmt.Errorf("search query required")
	}

	query := cmd.Args().First()
	limit := cmd.Int("limit")
	cursor := cmd.String("cursor")
	asJSON := cmd.Bool("json")

	service, err := reg.GetService()
	if err != nil {
		return fmt.Errorf("failed to get service: %w", err)
	}

	if !service.Authenticated() {
		return fmt.Errorf("not authenticated: run 'skycli login' first")
	}

	logger.Debug("Searching posts", "query", query, "limit", limit, "cursor", cursor)

	result, err := service.SearchPosts(ctx, query, limit, cursor)
	if err != nil {
		return fmt.Errorf("failed to search posts: %w", err)
	}

	if asJSON {
		return ui.DisplayJSON(result)
	}

	if len(result.Posts) == 0 {
		ui.Infoln("No posts found matching query: %s", query)
		return nil
	}

	ui.Titleln("Search Results: %s", query)
	ui.DisplayFeed(result.Posts, result.Cursor)

	return nil
}

// SearchFeedsAction searches for feeds in the local database by name or source
func SearchFeedsAction(ctx context.Context, cmd *cli.Command) error {
	if err := setup.EnsurePersistenceReady(ctx); err != nil {
		return fmt.Errorf("persistence layer not ready: %w", err)
	}

	reg := registry.Get()

	if cmd.Args().Len() == 0 {
		return fmt.Errorf("search query required")
	}

	query := cmd.Args().First()
	asJSON := cmd.Bool("json")

	feedRepo, err := reg.GetFeedRepo()
	if err != nil {
		return fmt.Errorf("failed to get feed repository: %w", err)
	}

	logger.Debug("Searching local feeds", "query", query)

	allFeeds, err := feedRepo.List(ctx)
	if err != nil {
		logger.Error("Failed to list feeds", "error", err)
		return err
	}

	var matchingFeeds []store.Model
	queryLower := strings.ToLower(query)

	for _, model := range allFeeds {
		if feed, ok := model.(*store.FeedModel); ok {
			nameLower := strings.ToLower(feed.Name)
			sourceLower := strings.ToLower(feed.Source)
			if strings.Contains(nameLower, queryLower) || strings.Contains(sourceLower, queryLower) {
				matchingFeeds = append(matchingFeeds, feed)
			}
		}
	}

	if len(matchingFeeds) == 0 {
		ui.Infoln("No feeds found matching query: %s", query)
		return nil
	}

	if asJSON {
		return ui.DisplayJSON(matchingFeeds)
	}

	ui.Titleln("Search Results: %s", query)
	fmt.Println()

	for i, model := range matchingFeeds {
		if feed, ok := model.(*store.FeedModel); ok {
			ui.Subtitleln("[%d] %s", i+1, feed.Name)
			ui.Infoln("  ID: %s", feed.ID())
			ui.Infoln("  Source: %s", feed.Source)
			ui.Infoln("  Local: %t", feed.IsLocal)
			fmt.Println()
		}
	}

	ui.Successln("Found %d feed(s)", len(matchingFeeds))
	return nil
}

// SearchCommand returns the search command with subcommands for users, posts, and feeds
func SearchCommand() *cli.Command {
	commonFlags := []cli.Flag{
		&cli.IntFlag{
			Name:    "limit",
			Aliases: []string{"l"},
			Usage:   "Maximum number of results to return",
			Value:   25,
		},
		&cli.StringFlag{
			Name:    "cursor",
			Aliases: []string{"c"},
			Usage:   "Pagination cursor for fetching additional results",
		},
		&cli.BoolFlag{
			Name:    "json",
			Aliases: []string{"j"},
			Usage:   "Output raw JSON response",
		},
	}

	feedFlags := []cli.Flag{
		&cli.BoolFlag{
			Name:    "json",
			Aliases: []string{"j"},
			Usage:   "Output raw JSON response",
		},
	}

	return &cli.Command{
		Name:  "search",
		Usage: "Search for users, posts, or feeds",
		Commands: []*cli.Command{
			{
				Name:      "users",
				Usage:     "Search for users by handle or name",
				ArgsUsage: "<query>",
				Flags:     commonFlags,
				Action:    SearchUsersAction,
			},
			{
				Name:      "posts",
				Usage:     "Search for posts by text content",
				ArgsUsage: "<query>",
				Flags:     commonFlags,
				Action:    SearchPostsAction,
			},
			{
				Name:      "feeds",
				Usage:     "Search local feeds by name or source (local search only)",
				ArgsUsage: "<query>",
				Flags:     feedFlags,
				Action:    SearchFeedsAction,
			},
		},
	}
}
