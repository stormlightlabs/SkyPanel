package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	lgtable "github.com/charmbracelet/lipgloss/table"
	"github.com/charmbracelet/log"
	"github.com/stormlightlabs/skypanel/cli/internal/registry"
	"github.com/stormlightlabs/skypanel/cli/internal/setup"
	"github.com/stormlightlabs/skypanel/cli/internal/store"
	"github.com/stormlightlabs/skypanel/cli/internal/ui"
	"github.com/urfave/cli/v3"
)

// followerInfo holds enriched follower data for display and export
type followerInfo struct {
	Profile       *store.ActorProfile
	LastPostDate  time.Time
	DaysSincePost int
	IsInactive    bool
	PostsPerDay   float64
	IsQuiet       bool
}

type diffOutput struct {
	NewFollowers []string `json:"newFollowers"`
	Unfollows    []string `json:"unfollows"`
	Summary      struct {
		BaselineCount   int `json:"baselineCount"`
		ComparisonCount int `json:"comparisonCount"`
		NetChange       int `json:"netChange"`
		NewCount        int `json:"newCount"`
		UnfollowCount   int `json:"unfollowCount"`
	} `json:"summary"`
}

// FollowersCommand returns the followers command with all subcommands
func FollowersCommand() *cli.Command {
	return &cli.Command{
		Name:  "followers",
		Usage: "Manage and analyze followers",
		Commands: []*cli.Command{
			{
				Name:      "list",
				Usage:     "List followers for a user",
				UsageText: "Fetch all followers with optional filters for inactivity, date range, and output format.",
				ArgsUsage: " ",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "user",
						Aliases: []string{"u"},
						Usage:   "User handle or DID (defaults to authenticated user)",
					},
					&cli.IntFlag{
						Name:    "limit",
						Aliases: []string{"l"},
						Usage:   "Maximum number of followers to fetch (0 = all)",
						Value:   0,
					},
					&cli.StringFlag{
						Name:  "since",
						Usage: "Filter followers created after date (YYYY-MM-DD)",
					},
					&cli.IntFlag{
						Name:  "inactive",
						Usage: "Show only followers with no posts in N days",
						Value: 0,
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
				Action: ListFollowersAction,
			},
			{
				Name:      "stats",
				Usage:     "Show aggregate follower statistics",
				UsageText: "Calculate aggregate statistics including active/inactive counts, growth metrics, and optional ASCII chart.",
				ArgsUsage: " ",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "user",
						Aliases: []string{"u"},
						Usage:   "User handle or DID (defaults to authenticated user)",
					},
					&cli.StringFlag{
						Name:  "since",
						Usage: "Calculate growth since date (YYYY-MM-DD)",
					},
					&cli.IntFlag{
						Name:  "inactive",
						Usage: "Threshold for inactive status (days)",
						Value: 60,
					},
					&cli.BoolFlag{
						Name:  "chart",
						Usage: "Display ASCII bar chart",
					},
				},
				Action: FollowersStatsAction,
			},
			{
				Name:      "diff",
				Usage:     "Compare follower lists between two dates",
				UsageText: "Compare follower lists to identify new followers and unfollows. Without --until, compares snapshot to current live data.",
				ArgsUsage: " ",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "user",
						Aliases: []string{"u"},
						Usage:   "User handle or DID (defaults to authenticated user)",
					},
					&cli.StringFlag{
						Name:     "since",
						Usage:    "Start date (YYYY-MM-DD) or snapshot ID",
						Required: true,
					},
					&cli.StringFlag{
						Name:  "until",
						Usage: "End date (YYYY-MM-DD) or snapshot ID (omit to compare with live data)",
					},
					&cli.StringFlag{
						Name:    "output",
						Aliases: []string{"o"},
						Usage:   "Output format: table, json, csv",
						Value:   "table",
					},
				},
				Action: FollowersDiffAction,
			},
			{
				Name:      "export",
				Usage:     "Export followers to CSV or JSON",
				UsageText: "Export follower list to CSV or JSON for external analysis, archival, or backup purposes.",
				ArgsUsage: " ",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "user",
						Aliases: []string{"u"},
						Usage:   "User handle or DID (defaults to authenticated user)",
					},
					&cli.IntFlag{
						Name:  "inactive",
						Usage: "Export only followers with no posts in N days",
						Value: 0,
					},
					&cli.BoolFlag{
						Name:  "quiet",
						Usage: "Export only quiet posters (low posting frequency)",
					},
					&cli.FloatFlag{
						Name:  "threshold",
						Usage: "Posts per day threshold for quiet posters (used with --quiet)",
						Value: 1.0,
					},
					&cli.StringFlag{
						Name:     "output",
						Aliases:  []string{"o"},
						Usage:    "Output format: json, csv",
						Value:    "csv",
						Required: true,
					},
					&cli.BoolFlag{
						Name:  "refresh",
						Usage: "Force refresh cached data (bypasses 24-hour cache)",
					},
				},
				Action: FollowersExportAction,
			},
		},
	}
}

// ListFollowersAction fetches and displays followers for a user with optional filtering
func ListFollowersAction(ctx context.Context, cmd *cli.Command) error {
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
	limit := cmd.Int("limit")
	sinceStr := cmd.String("since")
	inactiveDays := cmd.Int("inactive")
	quietPosters := cmd.Bool("quiet")
	quietThreshold := cmd.Float("threshold")
	outputFormat := cmd.String("output")
	refresh := cmd.Bool("refresh")

	if limit == 0 {
		logger.Debugf("Fetching all followers for %v", actor)
	} else {
		logger.Debugf("Fetching %v followers for %v", actor, limit)
	}

	var allFollowers []store.ActorProfile
	cursor := ""
	page := 0
	for {
		page++
		response, err := service.GetFollowers(ctx, actor, 100, cursor)
		if err != nil {
			return fmt.Errorf("failed to fetch followers: %w", err)
		}

		allFollowers = append(allFollowers, response.Followers...)

		if response.Cursor != "" {
			logger.Infof("Fetched page %d (%d followers so far)...", page, len(allFollowers))
		}

		if response.Cursor == "" || (limit > 0 && len(allFollowers) >= limit) {
			break
		}
		cursor = response.Cursor
	}

	logger.Infof("Fetched %d total followers", len(allFollowers))

	if limit > 0 && len(allFollowers) > limit {
		allFollowers = allFollowers[:limit]
	}

	if sinceStr != "" {
		since, err := time.Parse("2006-01-02", sinceStr)
		if err != nil {
			return fmt.Errorf("invalid date format (use YYYY-MM-DD): %w", err)
		}

		var filtered []store.ActorProfile
		for _, follower := range allFollowers {
			if follower.IndexedAt == "" {
				continue
			}
			indexedAt, err := time.Parse(time.RFC3339, follower.IndexedAt)
			if err != nil {
				logger.Warn("Failed to parse indexedAt", "error", err)
				continue
			}
			if indexedAt.After(since) {
				filtered = append(filtered, follower)
			}
		}
		allFollowers = filtered
	}

	followerInfos, actors := enrichFollowerProfiles(ctx, service, allFollowers, logger)

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

// FollowersStatsAction displays aggregate statistics about followers
func FollowersStatsAction(ctx context.Context, cmd *cli.Command) error {
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

	actor := cmd.String("user")
	if actor == "" {
		actor = service.GetDid()
	}
	sinceStr := cmd.String("since")
	inactiveDays := cmd.Int("inactive")
	showChart := cmd.Bool("chart")

	logger.Debugf("Fetching followers stats for actor %v", actor)

	var allFollowers []store.ActorProfile
	cursor := ""
	page := 0
	for {
		page++
		response, err := service.GetFollowers(ctx, actor, 100, cursor)
		if err != nil {
			return fmt.Errorf("failed to fetch followers: %w", err)
		}

		allFollowers = append(allFollowers, response.Followers...)

		if response.Cursor != "" {
			logger.Infof("Fetched page %d (%d followers so far)...", page, len(allFollowers))
		}

		if response.Cursor == "" {
			break
		}
		cursor = response.Cursor
	}

	logger.Infof("Fetched %d total followers", len(allFollowers))

	totalFollowers := len(allFollowers)

	// Fetch full profiles for stats (required for accurate counts)
	logger.Infof("Fetching detailed profiles for %d followers...", len(allFollowers))
	actors := make([]string, len(allFollowers))
	for i, follower := range allFollowers {
		actors[i] = follower.Did
	}

	fullProfiles := service.BatchGetProfiles(ctx, actors, 10)
	logger.Infof("Fetched %d detailed profiles", len(fullProfiles))

	var growth int
	var sinceDate time.Time
	if sinceStr != "" {
		since, err := time.Parse("2006-01-02", sinceStr)
		if err != nil {
			return fmt.Errorf("invalid date format (use YYYY-MM-DD): %w", err)
		}
		sinceDate = since

		for _, follower := range allFollowers {
			if follower.IndexedAt == "" {
				continue
			}
			indexedAt, err := time.Parse(time.RFC3339, follower.IndexedAt)
			if err != nil {
				continue
			}
			if indexedAt.After(since) {
				growth++
			}
		}
	}

	var activeCount, inactiveCount int
	if inactiveDays > 0 {
		logger.Infof("Checking activity status (threshold: %d days)...", inactiveDays)

		actors := make([]string, len(allFollowers))
		for i, follower := range allFollowers {
			actors[i] = follower.Did
		}

		lastPostDates := service.BatchGetLastPostDates(ctx, actors, 10)

		for _, actor := range actors {
			lastPost, ok := lastPostDates[actor]
			if !ok || lastPost.IsZero() {
				inactiveCount++
			} else {
				daysSince := int(time.Since(lastPost).Hours() / 24)
				if daysSince > inactiveDays {
					inactiveCount++
				} else {
					activeCount++
				}
			}
		}
	}

	ui.Titleln("Follower Statistics")
	fmt.Printf("Total followers: %d\n", totalFollowers)

	if inactiveDays > 0 {
		fmt.Printf("Active: %d\n", activeCount)
		fmt.Printf("Inactive: %d (no post > %d days)\n", inactiveCount, inactiveDays)
	}

	if sinceStr != "" {
		fmt.Printf("\nGrowth since %s: +%d\n", sinceDate.Format("2006-01-02"), growth)
	}

	if showChart && inactiveDays > 0 {
		displayActivityChart(activeCount, inactiveCount)
	}

	return nil
}

// FollowersDiffAction compares follower lists between two dates
func FollowersDiffAction(ctx context.Context, cmd *cli.Command) error {
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

	snapshotRepo, err := reg.GetSnapshotRepo()
	if err != nil {
		return fmt.Errorf("failed to get snapshot repository: %w", err)
	}

	actor := cmd.String("user")
	if actor == "" {
		actor = service.GetDid()
	}
	sinceStr := cmd.String("since")
	untilStr := cmd.String("until")
	outputFormat := cmd.String("output")

	// Parse since parameter (date or snapshot ID)
	sinceDate, err := time.Parse("2006-01-02", sinceStr)
	var baselineSnapshot *store.SnapshotModel
	if err != nil {
		// Not a date, try as snapshot ID
		model, err := snapshotRepo.Get(ctx, sinceStr)
		if err != nil {
			return fmt.Errorf("invalid --since parameter (not a date or snapshot ID): %w", err)
		}
		if model == nil {
			return fmt.Errorf("snapshot not found: %s", sinceStr)
		}
		baselineSnapshot = model.(*store.SnapshotModel)
	} else {
		// Find snapshot by date
		baselineSnapshot, err = snapshotRepo.FindByUserTypeAndDate(ctx, actor, "followers", sinceDate)
		if err != nil {
			return fmt.Errorf("failed to find snapshot: %w", err)
		}
		if baselineSnapshot == nil {
			return fmt.Errorf("no snapshot found for %s on or before %s", actor, sinceStr)
		}
	}

	logger.Infof("Using baseline snapshot from %s (%d followers)", baselineSnapshot.CreatedAt().Format("2006-01-02 15:04"), baselineSnapshot.TotalCount)

	// Get baseline follower DIDs
	baselineDids, err := snapshotRepo.GetActorDids(ctx, baselineSnapshot.ID())
	if err != nil {
		return fmt.Errorf("failed to get baseline followers: %w", err)
	}

	var comparisonDids []string
	var comparisonLabel string

	if untilStr != "" {
		// Snapshot-to-snapshot comparison
		untilDate, err := time.Parse("2006-01-02", untilStr)
		var comparisonSnapshot *store.SnapshotModel
		if err != nil {
			// Not a date, try as snapshot ID
			model, err := snapshotRepo.Get(ctx, untilStr)
			if err != nil {
				return fmt.Errorf("invalid --until parameter (not a date or snapshot ID): %w", err)
			}
			if model == nil {
				return fmt.Errorf("snapshot not found: %s", untilStr)
			}
			comparisonSnapshot = model.(*store.SnapshotModel)
		} else {
			// Find snapshot by date
			comparisonSnapshot, err = snapshotRepo.FindByUserTypeAndDate(ctx, actor, "followers", untilDate)
			if err != nil {
				return fmt.Errorf("failed to find snapshot: %w", err)
			}
			if comparisonSnapshot == nil {
				return fmt.Errorf("no snapshot found for %s on or before %s", actor, untilStr)
			}
		}

		logger.Infof("Comparing with snapshot from %s (%d followers)", comparisonSnapshot.CreatedAt().Format("2006-01-02 15:04"), comparisonSnapshot.TotalCount)
		comparisonLabel = comparisonSnapshot.CreatedAt().Format("2006-01-02 15:04")

		comparisonDids, err = snapshotRepo.GetActorDids(ctx, comparisonSnapshot.ID())
		if err != nil {
			return fmt.Errorf("failed to get comparison followers: %w", err)
		}
	} else {
		// Snapshot-to-live comparison
		logger.Infof("Fetching current followers for comparison...")
		comparisonLabel = "now"

		var allFollowers []store.ActorProfile
		cursor := ""
		page := 0
		for {
			page++
			response, err := service.GetFollowers(ctx, actor, 100, cursor)
			if err != nil {
				return fmt.Errorf("failed to fetch followers: %w", err)
			}

			allFollowers = append(allFollowers, response.Followers...)

			if response.Cursor != "" {
				logger.Infof("Fetched page %d (%d followers so far)...", page, len(allFollowers))
			}

			if response.Cursor == "" {
				break
			}
			cursor = response.Cursor
		}

		logger.Infof("Fetched %d current followers", len(allFollowers))

		for _, follower := range allFollowers {
			comparisonDids = append(comparisonDids, follower.Did)
		}
	}

	// Calculate diff
	baselineSet := make(map[string]bool)
	for _, did := range baselineDids {
		baselineSet[did] = true
	}

	comparisonSet := make(map[string]bool)
	for _, did := range comparisonDids {
		comparisonSet[did] = true
	}

	// New followers: in comparison but not in baseline
	var newFollowers []string
	for _, did := range comparisonDids {
		if !baselineSet[did] {
			newFollowers = append(newFollowers, did)
		}
	}

	// Unfollows: in baseline but not in comparison
	var unfollows []string
	for _, did := range baselineDids {
		if !comparisonSet[did] {
			unfollows = append(unfollows, did)
		}
	}

	// Output results
	switch outputFormat {
	case "json":
		return outputDiffJSON(newFollowers, unfollows)
	case "csv":
		return outputDiffCSV(newFollowers, unfollows)
	default:
		displayDiffTable(baselineSnapshot.CreatedAt().Format("2006-01-02 15:04"), comparisonLabel, len(baselineDids), len(comparisonDids), newFollowers, unfollows)
	}

	return nil
}

// FollowersExportAction exports followers to CSV or JSON
func FollowersExportAction(ctx context.Context, cmd *cli.Command) error {
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
	quietPosters := cmd.Bool("quiet")
	quietThreshold := cmd.Float("threshold")
	outputFormat := cmd.String("output")
	refresh := cmd.Bool("refresh")

	logger.Debugf("Exporting followers for actor %v with fmt %v", actor, outputFormat)

	var allFollowers []store.ActorProfile
	cursor := ""
	page := 0
	for {
		page++
		response, err := service.GetFollowers(ctx, actor, 100, cursor)
		if err != nil {
			return fmt.Errorf("failed to fetch followers: %w", err)
		}

		allFollowers = append(allFollowers, response.Followers...)

		if response.Cursor != "" {
			logger.Infof("Fetched page %d (%d followers so far)...", page, len(allFollowers))
		}

		if response.Cursor == "" {
			break
		}
		cursor = response.Cursor
	}

	logger.Infof("Fetched %d total followers", len(allFollowers))

	followerInfos, actors := enrichFollowerProfiles(ctx, service, allFollowers, logger)

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
		return fmt.Errorf("output format must be 'json' or 'csv'")
	}
}

// enrichFollowerProfiles fetches full profiles and merges them with lightweight profiles
func enrichFollowerProfiles(ctx context.Context, service *store.BlueskyService, profiles []store.ActorProfile, logger *log.Logger) ([]followerInfo, []string) {
	logger.Infof("Fetching detailed profiles for %d accounts...", len(profiles))
	actors := make([]string, len(profiles))
	for i, profile := range profiles {
		actors[i] = profile.Did
	}

	fullProfiles := service.BatchGetProfiles(ctx, actors, 10)
	logger.Infof("Fetched %d detailed profiles", len(fullProfiles))

	followerInfos := make([]followerInfo, len(profiles))
	for i, profile := range profiles {
		if fullProfile, ok := fullProfiles[profile.Did]; ok {
			followerInfos[i] = followerInfo{Profile: fullProfile}
		} else {
			followerInfos[i] = followerInfo{Profile: &profile}
		}
	}

	return followerInfos, actors
}

// filterInactive filters follower infos to only include accounts inactive for N days
func filterInactive(ctx context.Context, service *store.BlueskyService, cacheRepo *store.CacheRepository, followerInfos []followerInfo, actors []string, inactiveDays int, refresh bool, logger *log.Logger) []followerInfo {
	logger.Infof("Checking activity status (threshold: %d days)...", inactiveDays)

	lastPostDates := service.BatchGetLastPostDatesCached(ctx, cacheRepo, actors, 10, refresh)

	var filtered []followerInfo
	for i, info := range followerInfos {
		lastPost, ok := lastPostDates[actors[i]]
		info.LastPostDate = lastPost

		if !ok || lastPost.IsZero() {
			info.IsInactive = true
			info.DaysSincePost = -1
		} else {
			daysSince := int(time.Since(lastPost).Hours() / 24)
			info.DaysSincePost = daysSince
			info.IsInactive = daysSince > inactiveDays
		}

		if info.IsInactive {
			filtered = append(filtered, info)
		}
		followerInfos[i] = info
	}

	return filtered
}

// filterQuiet filters follower infos to only include quiet posters
func filterQuiet(ctx context.Context, service *store.BlueskyService, cacheRepo *store.CacheRepository, followerInfos []followerInfo, actors []string, threshold float64, refresh bool, logger *log.Logger) []followerInfo {
	logger.Infof("Computing post rates (threshold: %.2f posts/day)...", threshold)
	if refresh {
		logger.Infof("Refreshing cache (this may take a while)...")
	}

	postRates := service.BatchGetPostRatesCached(ctx, cacheRepo, actors, 30, 30, 10, refresh, func(current, total int) {
		if current%10 == 0 || current == total {
			logger.Infof("Progress: %d/%d accounts analyzed", current, total)
		}
	})

	var filtered []followerInfo
	for i, info := range followerInfos {
		if rate, ok := postRates[actors[i]]; ok {
			info.PostsPerDay = rate.PostsPerDay
			info.LastPostDate = rate.LastPostDate
			info.IsQuiet = rate.PostsPerDay <= threshold
		}

		if info.IsQuiet {
			filtered = append(filtered, info)
		}
		followerInfos[i] = info
	}

	logger.Infof("Found %d quiet posters (posting <= %.2f times/day)", len(filtered), threshold)
	return filtered
}

func displayDiffTable(baselineLabel, comparisonLabel string, baselineCount, comparisonCount int, newFollowers, unfollows []string) {
	ui.Titleln("Follower Diff: %s → %s", baselineLabel, comparisonLabel)
	fmt.Println()

	fmt.Printf("Baseline:   %d followers\n", baselineCount)
	fmt.Printf("Comparison: %d followers\n", comparisonCount)
	fmt.Printf("Net change: %+d\n", comparisonCount-baselineCount)
	fmt.Println()

	if len(newFollowers) > 0 {
		ui.Titleln("New Followers (%d)", len(newFollowers))
		for _, did := range newFollowers {
			fmt.Printf("  + %s\n", did)
		}
		fmt.Println()
	}

	if len(unfollows) > 0 {
		ui.Titleln("Unfollows (%d)", len(unfollows))
		for _, did := range unfollows {
			fmt.Printf("  - %s\n", did)
		}
		fmt.Println()
	}

	if len(newFollowers) == 0 && len(unfollows) == 0 {
		ui.Infoln("No changes detected")
	}
}

func outputDiffJSON(newFollowers, unfollows []string) error {
	output := diffOutput{
		NewFollowers: newFollowers,
		Unfollows:    unfollows,
	}
	if output.NewFollowers == nil {
		output.NewFollowers = []string{}
	}
	if output.Unfollows == nil {
		output.Unfollows = []string{}
	}
	output.Summary.NewCount = len(newFollowers)
	output.Summary.UnfollowCount = len(unfollows)

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

func outputDiffCSV(newFollowers, unfollows []string) error {
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	if err := writer.Write([]string{"type", "did"}); err != nil {
		return err
	}

	for _, did := range newFollowers {
		if err := writer.Write([]string{"new_follower", did}); err != nil {
			return err
		}
	}

	for _, did := range unfollows {
		if err := writer.Write([]string{"unfollow", did}); err != nil {
			return err
		}
	}

	return nil
}

// formatTimeSince formats a time duration into a human-readable string.
//
// Returns
//   - "< 1 hour ago" for durations under 1 hour
//   - "X hours ago" for under 24 hours
//   - "X days ago" for longer durations.
func formatTimeSince(since time.Time) string {
	if since.IsZero() {
		return "never"
	}

	duration := time.Since(since)
	hours := duration.Hours()

	if hours < 1 {
		return "< 1 hour ago"
	} else if hours < 24 {
		return fmt.Sprintf("%d hours ago", int(hours))
	} else {
		days := int(hours / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}
}

func displayFollowersTable(followers []followerInfo, showInactive bool) {
	if len(followers) == 0 {
		ui.Infoln("No followers found")
		return
	}

	ui.Titleln("Followers (%d)", len(followers))
	fmt.Println()

	headers := []string{"Handle", "Display Name", "Followers", "Posts"}

	if showInactive && len(followers) > 0 && followers[0].IsQuiet {
		headers = append(headers, "Posts/Day", "Last Post")
	} else if len(followers) > 0 && followers[0].IsQuiet {
		headers = append(headers, "Posts/Day")
	} else if showInactive {
		headers = append(headers, "Last Post")
	}
	headers = append(headers, "Profile URL")

	data := make([][]string, len(followers))
	for i, info := range followers {
		displayName := info.Profile.DisplayName
		if displayName == "" {
			displayName = info.Profile.Handle
		}

		profileURL := fmt.Sprintf("https://bsky.app/profile/%s", info.Profile.Handle)

		row := []string{
			"@" + info.Profile.Handle,
			displayName,
			fmt.Sprintf("%d", info.Profile.FollowersCount),
			fmt.Sprintf("%d", info.Profile.PostsCount),
		}

		if showInactive && info.IsQuiet {
			row = append(row, fmt.Sprintf("%.2f", info.PostsPerDay))
			row = append(row, formatTimeSince(info.LastPostDate))
		} else if info.IsQuiet {
			row = append(row, fmt.Sprintf("%.2f", info.PostsPerDay))
		} else if showInactive {
			row = append(row, formatTimeSince(info.LastPostDate))
		}

		row = append(row, profileURL)
		data[i] = row
	}

	lastColIdx := len(headers) - 1

	re := lipgloss.NewRenderer(os.Stdout)
	t := lgtable.New().Border(lipgloss.NormalBorder()).BorderStyle(ui.TableBorderStyle).Headers(headers...).Rows(data...)
	t = t.StyleFunc(func(row, col int) lipgloss.Style {
		if row == lgtable.HeaderRow {
			return ui.TableHeaderStyle
		}

		if col == 0 {
			even := row%2 == 0
			if even {
				return ui.TableRowEvenStyle.Foreground(lipgloss.Color("#f6c177"))
			}
			return ui.TableRowOddStyle.Foreground(lipgloss.Color("#f6c177"))
		}

		if col == lastColIdx {
			even := row%2 == 0
			baseStyle := ui.TableRowEvenStyle
			if !even {
				baseStyle = ui.TableRowOddStyle
			}
			return baseStyle.Foreground(lipgloss.Color("#e0def4"))
		}

		if row%2 == 0 {
			return ui.TableRowEvenStyle
		}
		return ui.TableRowOddStyle
	})

	fmt.Println(re.NewStyle().Render(t.String()))
	fmt.Println()
}

func outputFollowersJSON(followers []followerInfo) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(followers)
}

func outputFollowersCSV(followers []followerInfo, includeInactive bool) error {
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	hasQuiet := len(followers) > 0 && followers[0].IsQuiet

	header := []string{"handle", "displayName", "did", "followersCount", "postsCount", "profileURL"}
	if hasQuiet {
		header = append(header, "postsPerDay")
	}
	if includeInactive {
		header = append(header, "daysSincePost", "lastPostDate")
	}
	if err := writer.Write(header); err != nil {
		return err
	}

	for _, info := range followers {
		// profileURL := fmt.Sprintf("https://bsky.app/profile/%s", info.Profile.Handle)
		row := []string{
			info.Profile.Handle,
			info.Profile.DisplayName,
			info.Profile.Did,
			fmt.Sprintf("%d", info.Profile.FollowersCount),
			fmt.Sprintf("%d", info.Profile.PostsCount),
			// profileURL,
		}

		if hasQuiet {
			row = append(row, fmt.Sprintf("%.2f", info.PostsPerDay))
		}

		if includeInactive {
			daysSince := "N/A"
			if info.DaysSincePost >= 0 {
				daysSince = fmt.Sprintf("%d", info.DaysSincePost)
			}
			lastPost := ""
			if !info.LastPostDate.IsZero() {
				lastPost = info.LastPostDate.Format(time.RFC3339)
			}
			row = append(row, daysSince, lastPost)
		}

		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}

func displayActivityChart(active, inactive int) {
	total := active + inactive
	if total == 0 {
		return
	}

	fmt.Println()

	activePercent := float64(active) / float64(total)
	inactivePercent := float64(inactive) / float64(total)

	chartWidth := 30
	activeBars := int(activePercent * float64(chartWidth))
	inactiveBars := int(inactivePercent * float64(chartWidth))

	if activeBars+inactiveBars < chartWidth {
		activeBars = chartWidth - inactiveBars
	}

	activeBar := strings.Repeat("█", activeBars)
	inactiveBar := strings.Repeat("▒", inactiveBars)

	fmt.Printf("%s%s\n", activeBar, inactiveBar)
	fmt.Printf("█ Active   ▒ Inactive\n")
}
