package main

import (
	"context"
	"fmt"

	"github.com/stormlightlabs/skypanel/cli/internal/registry"
	"github.com/stormlightlabs/skypanel/cli/internal/setup"
	"github.com/stormlightlabs/skypanel/cli/internal/store"
	"github.com/urfave/cli/v3"
)

// ListFollowingAction fetches and displays accounts the user follows
func ListFollowingAction(ctx context.Context, cmd *cli.Command) error {
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

	cacheRepo, err := reg.GetCacheRepo()
	if err != nil {
		return fmt.Errorf("failed to get cache repository: %w", err)
	}

	actor := cmd.String("user")
	if actor == "" {
		actor = service.GetDid()
	}
	inactiveDays := cmd.Int("inactive")
	mutual := cmd.Bool("mutual")
	quietPosters := cmd.Bool("quiet")
	quietThreshold := cmd.Float("threshold")
	outputFormat := cmd.String("output")
	refresh := cmd.Bool("refresh")

	logger.Debugf("Fetching following for actor %v", actor)

	var allFollowing []store.ActorProfile
	cursor := ""
	page := 0
	for {
		page++
		response, err := service.GetFollows(ctx, actor, 100, cursor)
		if err != nil {
			return fmt.Errorf("failed to fetch following: %w", err)
		}

		allFollowing = append(allFollowing, response.Follows...)

		if response.Cursor != "" {
			logger.Infof("Fetched page %d (%d following so far)...", page, len(allFollowing))
		}

		if response.Cursor == "" {
			break
		}
		cursor = response.Cursor
	}

	logger.Infof("Fetched %d total following", len(allFollowing))

	if mutual {
		var mutualFollows []store.ActorProfile
		for _, follow := range allFollowing {
			if follow.Viewer != nil && follow.Viewer.FollowedBy != "" {
				mutualFollows = append(mutualFollows, follow)
			}
		}
		allFollowing = mutualFollows
	}

	followerInfos, actors := enrichFollowerProfiles(ctx, service, allFollowing, logger)

	if inactiveDays > 0 {
		followerInfos = filterInactive(ctx, service, cacheRepo, followerInfos, actors, inactiveDays, refresh, logger)
	}

	if quietPosters {
		followerInfos = filterQuiet(ctx, service, cacheRepo, followerInfos, actors, quietThreshold, refresh, logger)
	}

	switch outputFormat {
	case "json":
		return outputFollowersJSON(followerInfos)
	case "csv":
		return outputFollowersCSV(followerInfos, inactiveDays > 0 || quietPosters)
	default:
		displayFollowersTable(followerInfos, inactiveDays > 0 || quietPosters)
	}

	return nil
}

// FollowingCommand returns the following command
func FollowingCommand() *cli.Command {
	return &cli.Command{
		Name:  "following",
		Usage: "Manage and analyze accounts you follow",
		Commands: []*cli.Command{
			{
				Name:      "list",
				Usage:     "List accounts you follow",
				UsageText: "Fetch all accounts you follow with optional filters for inactive accounts and mutual follows.",
				ArgsUsage: " ",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "user",
						Aliases: []string{"u"},
						Usage:   "User handle or DID (defaults to authenticated user)",
					},
					&cli.IntFlag{
						Name:  "inactive",
						Usage: "Show only accounts with no posts in N days",
						Value: 0,
					},
					&cli.BoolFlag{
						Name:  "mutual",
						Usage: "Show only mutual follows",
					},
					&cli.BoolFlag{
						Name:  "quiet",
						Usage: "Show only quiet posters (low posting frequency)",
					},
					&cli.FloatFlag{
						Name:  "threshold",
						Usage: "Posts per day threshold for quiet posters (used with --quiet)",
						Value: 1.0,
					},
					&cli.StringFlag{
						Name:    "output",
						Aliases: []string{"o"},
						Usage:   "Output format: table, json, csv",
						Value:   "table",
					},
					&cli.BoolFlag{
						Name:  "refresh",
						Usage: "Force refresh cached data (bypasses 24-hour cache)",
					},
				},
				Action: ListFollowingAction,
			},
		},
	}
}
