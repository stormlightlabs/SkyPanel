package main

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/stormlightlabs/skypanel/cli/internal/registry"
	"github.com/stormlightlabs/skypanel/cli/internal/setup"
	"github.com/stormlightlabs/skypanel/cli/internal/store"
	"github.com/stormlightlabs/skypanel/cli/internal/ui"
	"github.com/urfave/cli/v3"
)

// ViewFeedAction views posts from a feed (fetches from API)
func ViewFeedAction(ctx context.Context, cmd *cli.Command) error {
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

	logger.Debug("Fetching feed from API", "uri", feedURI, "limit", limit, "cursor", cursor)

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

// ViewPostAction views a single post by URI or URL
func ViewPostAction(ctx context.Context, cmd *cli.Command) error {
	if err := setup.EnsurePersistenceReady(ctx); err != nil {
		return fmt.Errorf("persistence layer not ready: %w", err)
	}

	reg := registry.Get()

	if cmd.Args().Len() == 0 {
		return fmt.Errorf("post URI or URL required")
	}

	postIdentifier := cmd.Args().First()
	asJSON := cmd.Bool("json")

	service, err := reg.GetService()
	if err != nil {
		return fmt.Errorf("failed to get service: %w", err)
	}

	if !service.Authenticated() {
		return fmt.Errorf("not authenticated: run 'skycli login' first")
	}

	postURI, err := parsePostIdentifier(postIdentifier)
	if err != nil {
		return fmt.Errorf("failed to parse post identifier: %w", err)
	}

	logger.Debug("Fetching post", "uri", postURI)

	response, err := service.GetPosts(ctx, []string{postURI})
	if err != nil {
		return fmt.Errorf("failed to fetch post: %w", err)
	}

	if len(response.Posts) == 0 {
		return fmt.Errorf("post not found: %s", postURI)
	}

	if asJSON {
		return ui.DisplayJSON(response.Posts[0])
	}

	ui.Titleln("Post View")
	ui.DisplayFeed([]store.FeedViewPost{response.Posts[0]}, "")

	return nil
}

// ViewProfileAction views an actor's profile with stats
func ViewProfileAction(ctx context.Context, cmd *cli.Command) error {
	if err := setup.EnsurePersistenceReady(ctx); err != nil {
		return fmt.Errorf("persistence layer not ready: %w", err)
	}

	reg := registry.Get()

	if cmd.Args().Len() == 0 {
		return fmt.Errorf("actor handle or DID required")
	}

	actor := cmd.Args().First()
	showPosts := cmd.Bool("with-posts")
	asJSON := cmd.Bool("json")

	service, err := reg.GetService()
	if err != nil {
		return fmt.Errorf("failed to get service: %w", err)
	}

	if !service.Authenticated() {
		return fmt.Errorf("not authenticated: run 'skycli login' first")
	}

	logger.Debug("Fetching profile", "actor", actor)

	profile, err := service.GetProfile(ctx, actor)
	if err != nil {
		return fmt.Errorf("failed to fetch profile: %w", err)
	}

	if asJSON {
		return ui.DisplayJSON(profile)
	}

	ui.DisplayProfileHeader(profile)

	if showPosts {
		logger.Debug("Fetching recent posts", "actor", actor)
		feed, err := service.GetAuthorFeed(ctx, actor, 10, "")
		if err != nil {
			ui.Warningln("Failed to fetch recent posts: %v", err)
		} else {
			fmt.Println()
			ui.Subtitleln("Recent Posts")
			ui.DisplayFeed(feed.Feed, "")
		}
	}

	return nil
}

// ViewCommand returns the view command with subcommands for feed, post, and profile
func ViewCommand() *cli.Command {
	return &cli.Command{
		Name:  "view",
		Usage: "View feeds, posts, or profiles",
		Commands: []*cli.Command{
			{
				Name:      "feed",
				Usage:     "View posts from a feed by URI or local feed ID",
				ArgsUsage: "<feed-uri-or-id>",
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:    "limit",
						Aliases: []string{"l"},
						Usage:   "Maximum number of posts to display",
						Value:   25,
					},
					&cli.StringFlag{
						Name:    "cursor",
						Aliases: []string{"c"},
						Usage:   "Pagination cursor",
					},
					&cli.BoolFlag{
						Name:    "json",
						Aliases: []string{"j"},
						Usage:   "Output raw JSON response",
					},
				},
				Action: ViewFeedAction,
			},
			{
				Name:      "post",
				Usage:     "View a single post by URI or bsky.app URL",
				ArgsUsage: "<post-uri-or-url>",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "json",
						Aliases: []string{"j"},
						Usage:   "Output raw JSON response",
					},
				},
				Action: ViewPostAction,
			},
			{
				Name:      "profile",
				Usage:     "View an actor's profile",
				ArgsUsage: "<actor-handle-or-did>",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "with-posts",
						Aliases: []string{"p"},
						Usage:   "Also display recent posts from this profile",
					},
					&cli.BoolFlag{
						Name:    "json",
						Aliases: []string{"j"},
						Usage:   "Output raw JSON response",
					},
				},
				Action: ViewProfileAction,
			},
		},
	}
}

// parsePostIdentifier converts a bsky.app URL or AT URI to an AT URI
// Examples:
// - https://bsky.app/profile/alice.bsky.social/post/abc123
// - at://did:plc:xyz/app.bsky.feed.post/abc123
func parsePostIdentifier(identifier string) (string, error) {
	if strings.HasPrefix(identifier, "at://") {
		return identifier, nil
	}

	if strings.HasPrefix(identifier, "https://bsky.app/profile/") ||
		strings.HasPrefix(identifier, "http://bsky.app/profile/") {
		re := regexp.MustCompile(`^https?://bsky\.app/profile/([^/]+)/post/([^/]+)`)
		matches := re.FindStringSubmatch(identifier)
		if len(matches) != 3 {
			return "", fmt.Errorf("invalid bsky.app URL format")
		}

		handle := matches[1]
		rkey := matches[2]

		return fmt.Sprintf("at://%s/app.bsky.feed.post/%s", handle, rkey), nil
	}

	return "", fmt.Errorf("identifier must be an AT URI (at://...) or bsky.app URL")
}
