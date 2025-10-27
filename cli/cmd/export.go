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

	postIdentifier := cmd.Args().First()
	format := strings.ToLower(cmd.String("format"))

	if format != "json" && format != "txt" {
		return fmt.Errorf("invalid format for post: %s (must be json or txt)", format)
	}

	service, err := reg.GetService()
	if err != nil {
		return fmt.Errorf("failed to get service: %w", err)
	}

	if !service.Authenticated() {
		return fmt.Errorf("not authenticated: run 'skycli login' first")
	}

	postURI, err := parsePostURI(postIdentifier)
	if err != nil {
		return fmt.Errorf("failed to parse post identifier: %w", err)
	}

	logger.Debug("Fetching post for export", "uri", postURI)

	response, err := service.GetPosts(ctx, []string{postURI})
	if err != nil {
		return fmt.Errorf("failed to fetch post: %w", err)
	}

	if len(response.Posts) == 0 {
		return fmt.Errorf("post not found: %s", postURI)
	}

	post := &response.Posts[0]

	filename := fmt.Sprintf("post_%s_%s.%s", extractRkey(postURI), time.Now().Format("2006-01-02"), format)

	switch format {
	case "json":
		err = export.FeedViewPostToJSON(filename, post)
	case "txt":
		err = export.FeedViewPostToTXT(filename, post)
	}

	if err != nil {
		logger.Error("Failed to export", "error", err)
		return err
	}

	ui.Successln("Exported post to %s", filename)
	return nil
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

// parsePostURI converts a bsky.app URL or AT URI to an AT URI
func parsePostURI(identifier string) (string, error) {
	if strings.HasPrefix(identifier, "at://") {
		return identifier, nil
	}

	if strings.HasPrefix(identifier, "https://bsky.app/profile/") ||
		strings.HasPrefix(identifier, "http://bsky.app/profile/") {
		parts := strings.Split(identifier, "/")
		if len(parts) < 7 || parts[5] != "post" {
			return "", fmt.Errorf("invalid bsky.app URL format")
		}
		handle := parts[4]
		rkey := parts[6]
		return fmt.Sprintf("at://%s/app.bsky.feed.post/%s", handle, rkey), nil
	}

	return "", fmt.Errorf("identifier must be an AT URI (at://...) or bsky.app URL")
}

// extractRkey extracts the record key from an AT URI
func extractRkey(uri string) string {
	parts := strings.Split(uri, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return "unknown"
}
