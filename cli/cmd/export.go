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

// ExportFeedAction exports posts from a feed to file
func ExportFeedAction(ctx context.Context, cmd *cli.Command) error {
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

// ExportProfileAction exports an actor profile to file
func ExportProfileAction(ctx context.Context, cmd *cli.Command) error {
	if err := setup.EnsurePersistenceReady(ctx); err != nil {
		return fmt.Errorf("persistence layer not ready: %w", err)
	}

	logger := ui.GetLogger()
	reg := registry.Get()

	if cmd.Args().Len() == 0 {
		return fmt.Errorf("actor handle or DID required")
	}

	actor := cmd.Args().First()
	format := strings.ToLower(cmd.String("format"))

	if format != "json" && format != "txt" {
		return fmt.Errorf("invalid format for profile: %s (must be json or txt)", format)
	}

	service, err := reg.GetService()
	if err != nil {
		return fmt.Errorf("failed to get service: %w", err)
	}

	if !service.Authenticated() {
		return fmt.Errorf("not authenticated: run 'skycli login' first")
	}

	logger.Debug("Fetching profile for export", "actor", actor)

	profile, err := service.GetProfile(ctx, actor)
	if err != nil {
		return fmt.Errorf("failed to fetch profile: %w", err)
	}

	filename := fmt.Sprintf("profile_%s_%s.%s", profile.Handle, time.Now().Format("2006-01-02"), format)

	switch format {
	case "json":
		err = export.ProfileToJSON(filename, profile)
	case "txt":
		err = export.ProfileToTXT(filename, profile)
	}

	if err != nil {
		logger.Error("Failed to export", "error", err)
		return err
	}

	ui.Successln("Exported profile @%s to %s", profile.Handle, filename)
	return nil
}

// ExportPostAction exports a single post to file
func ExportPostAction(ctx context.Context, cmd *cli.Command) error {
	if err := setup.EnsurePersistenceReady(ctx); err != nil {
		return fmt.Errorf("persistence layer not ready: %w", err)
	}

	logger := ui.GetLogger()
	reg := registry.Get()

	if cmd.Args().Len() == 0 {
		return fmt.Errorf("post URI or URL required")
	}

	postURI := cmd.Args().First()
	format := strings.ToLower(cmd.String("format"))

	if format != "json" && format != "txt" {
		return fmt.Errorf("invalid format for post: %s (must be json or txt)", format)
	}

	// TODO: Convert URL to URI if needed
	if strings.HasPrefix(postURI, "https://bsky.app/profile/") {
		// Extract URI from bsky.app URL
		// Example: https://bsky.app/profile/user.bsky.social/post/abc123
		// -> at://did:plc:.../app.bsky.feed.post/abc123
		ui.Warningln("URL to URI conversion not yet implemented, please provide AT URI directly")
		return fmt.Errorf("URL conversion not implemented, use AT URI format (at://...)")
	}

	service, err := reg.GetService()
	if err != nil {
		return fmt.Errorf("failed to get service: %w", err)
	}

	if !service.Authenticated() {
		return fmt.Errorf("not authenticated: run 'skycli login' first")
	}

	logger.Debug("Fetching post for export", "uri", postURI)

	// TODO: Implement GetPost in BlueskyService
	// For now, we'll fetch the author feed and find the post
	ui.Warningln("Direct post fetch not yet implemented")
	return fmt.Errorf("direct post export not yet implemented - use feed export")
}

// ExportCommand returns the export command with subcommands for feed, profile, and post
func ExportCommand() *cli.Command {
	return &cli.Command{
		Name:  "export",
		Usage: "Export feeds, profiles, or posts to file",
		Commands: []*cli.Command{
			{
				Name:      "feed",
				Usage:     "Export posts from a feed",
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
				Action: ExportFeedAction,
			},
			{
				Name:      "profile",
				Usage:     "Export an actor profile",
				ArgsUsage: "<actor-handle-or-did>",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "format",
						Aliases: []string{"f"},
						Usage:   "Export format: json or txt",
						Value:   "json",
					},
				},
				Action: ExportProfileAction,
			},
			{
				Name:      "post",
				Usage:     "Export a single post",
				ArgsUsage: "<post-uri-or-url>",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "format",
						Aliases: []string{"f"},
						Usage:   "Export format: json or txt",
						Value:   "json",
					},
				},
				Action: ExportPostAction,
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if cmd.Args().Len() == 0 {
				return fmt.Errorf("please use: export feed|profile|post <identifier>")
			}
			return ExportFeedAction(ctx, cmd)
		},
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
	}
}
