package main

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/stormlightlabs/skypanel/cli/internal/registry"
	"github.com/stormlightlabs/skypanel/cli/internal/setup"
	"github.com/stormlightlabs/skypanel/cli/internal/store"
	"github.com/stormlightlabs/skypanel/cli/internal/ui"
	"github.com/urfave/cli/v3"
)

func ViewAction(ctx context.Context, cmd *cli.Command) error {
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

func ViewCommand() *cli.Command {
	return &cli.Command{
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
		Action: ViewAction,
	}
}
