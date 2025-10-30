package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/stormlightlabs/skypanel/cli/internal/registry"
	"github.com/stormlightlabs/skypanel/cli/internal/setup"
	"github.com/stormlightlabs/skypanel/cli/internal/store"
	"github.com/stormlightlabs/skypanel/cli/internal/ui"
	"github.com/urfave/cli/v3"
)

// FetchTimelineAction fetches and displays the authenticated user's home timeline
func FetchTimelineAction(ctx context.Context, cmd *cli.Command) error {
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

	limit := cmd.Int("limit")
	cursor := cmd.String("cursor")
	asJSON := cmd.Bool("json")

	logger.Debug("Fetching timeline", "limit", limit, "cursor", cursor)

	response, err := service.GetTimeline(ctx, limit, cursor)
	if err != nil {
		return fmt.Errorf("failed to fetch timeline: %w", err)
	}

	if asJSON {
		return ui.DisplayJSON(response)
	}

	ui.DisplayFeed(response.Feed, response.Cursor)
	return nil
}

// FetchFeedAction fetches and displays posts from a specific feed
func FetchFeedAction(ctx context.Context, cmd *cli.Command) error {
	if err := setup.EnsurePersistenceReady(ctx); err != nil {
		return fmt.Errorf("persistence layer not ready: %w", err)
	}

	reg := registry.Get()

	if cmd.Args().Len() == 0 {
		return fmt.Errorf("feed URI or local feed ID required")
	}

	feedIdentifier := cmd.Args().First()
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

	feedRepo, err := reg.GetFeedRepo()
	if err != nil {
		return fmt.Errorf("failed to get feed repository: %w", err)
	}

	var feedURI string

	if _, err := uuid.Parse(feedIdentifier); err == nil {
		feed, err := feedRepo.Get(ctx, feedIdentifier)
		if err != nil {
			return fmt.Errorf("failed to get local feed: %w", err)
		}
		if feedModel, ok := feed.(*store.FeedModel); ok {
			feedURI = feedModel.Source
			logger.Debug("Resolved local feed ID to URI", "id", feedIdentifier, "uri", feedURI)
		}
	} else {
		feedURI = feedIdentifier
	}

	logger.Debug("Fetching feed", "uri", feedURI, "limit", limit, "cursor", cursor)

	response, err := service.GetAuthorFeed(ctx, feedURI, limit, cursor)
	if err != nil {
		return fmt.Errorf("failed to fetch feed: %w", err)
	}

	if asJSON {
		return ui.DisplayJSON(response)
	}

	ui.Titleln("Feed: %s", feedURI)
	ui.DisplayFeed(response.Feed, response.Cursor)
	return nil
}

// FetchAuthorAction fetches and displays posts from a specific author with profile caching
func FetchAuthorAction(ctx context.Context, cmd *cli.Command) error {
	if err := setup.EnsurePersistenceReady(ctx); err != nil {
		return fmt.Errorf("persistence layer not ready: %w", err)
	}

	reg := registry.Get()

	if cmd.Args().Len() == 0 {
		return fmt.Errorf("actor handle or DID required")
	}

	actor := cmd.Args().First()
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

	profileRepo, err := reg.GetProfileRepo()
	if err != nil {
		return fmt.Errorf("failed to get profile repository: %w", err)
	}

	cachedProfile, err := profileRepo.GetByDid(ctx, actor)
	if err != nil {
		logger.Warn("Failed to check profile cache", "error", err)
	}

	var profile *store.ActorProfile
	if cachedProfile != nil && cachedProfile.IsFresh(time.Hour) {
		logger.Debug("Using cached profile", "did", actor)
		if err := json.Unmarshal([]byte(cachedProfile.DataJSON), &profile); err != nil {
			logger.Warn("Failed to unmarshal cached profile", "error", err)
			cachedProfile = nil
		}
	}

	if cachedProfile == nil || !cachedProfile.IsFresh(time.Hour) {
		logger.Debug("Fetching profile from API", "actor", actor)
		profile, err = service.GetProfile(ctx, actor)
		if err != nil {
			return fmt.Errorf("failed to fetch profile: %w", err)
		}

		profileJSON, err := json.Marshal(profile)
		if err != nil {
			logger.Warn("Failed to marshal profile for caching", "error", err)
		} else {
			profileModel := &store.ProfileModel{
				Did:       profile.Did,
				Handle:    profile.Handle,
				DataJSON:  string(profileJSON),
				FetchedAt: time.Now(),
			}
			if err := profileRepo.Save(ctx, profileModel); err != nil {
				logger.Warn("Failed to cache profile", "error", err)
			} else {
				logger.Debug("Cached profile", "did", profile.Did)
			}
		}
	}

	logger.Debug("Fetching author feed", "actor", actor, "limit", limit, "cursor", cursor)

	response, err := service.GetAuthorFeed(ctx, actor, limit, cursor)
	if err != nil {
		return fmt.Errorf("failed to fetch author feed: %w", err)
	}

	if asJSON {
		return ui.DisplayJSON(response)
	}

	if profile != nil {
		ui.DisplayProfileHeader(profile)
	}

	ui.DisplayFeed(response.Feed, response.Cursor)
	return nil
}

// FetchCommand returns the fetch command with subcommands for timeline, feed, and author
func FetchCommand() *cli.Command {
	commonFlags := []cli.Flag{
		&cli.IntFlag{
			Name:    "limit",
			Aliases: []string{"l"},
			Usage:   "Maximum number of posts to fetch",
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

	return &cli.Command{
		Name:  "fetch",
		Usage: "Fetch posts from timeline, feeds, or authors",
		Commands: []*cli.Command{
			{
				Name:      "timeline",
				Usage:     "Fetch authenticated user's home timeline",
				ArgsUsage: " ",
				Flags:     commonFlags,
				Action:    FetchTimelineAction,
			},
			{
				Name:      "feed",
				Usage:     "Fetch posts from a specific feed by URI or local feed ID",
				ArgsUsage: "<feed-uri-or-id>",
				Flags:     commonFlags,
				Action:    FetchFeedAction,
			},
			{
				Name:      "author",
				Usage:     "Fetch posts from a specific author (with profile caching)",
				ArgsUsage: "<actor-handle-or-did>",
				Flags:     commonFlags,
				Action:    FetchAuthorAction,
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return FetchTimelineAction(ctx, cmd)
		},
		Flags: commonFlags,
	}
}
