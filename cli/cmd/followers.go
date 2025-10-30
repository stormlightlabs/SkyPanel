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
func filterInactive(ctx context.Context, service *store.BlueskyService, followerInfos []followerInfo, actors []string, inactiveDays int, logger *log.Logger) []followerInfo {
	logger.Infof("Checking activity status (threshold: %d days)...", inactiveDays)

	lastPostDates := service.BatchGetLastPostDates(ctx, actors, 10)

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
func filterQuiet(ctx context.Context, service *store.BlueskyService, followerInfos []followerInfo, actors []string, threshold float64, logger *log.Logger) []followerInfo {
	logger.Infof("Computing post rates (threshold: %.2f posts/day, this may take a while)...", threshold)

	postRates := service.BatchGetPostRates(ctx, actors, 30, 30, 10, func(current, total int) {
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
		followerInfos = filterInactive(ctx, service, followerInfos, actors, inactiveDays, logger)
	}

	if quietPosters {
		followerInfos = filterQuiet(ctx, service, followerInfos, actors, quietThreshold, logger)
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

	actor := cmd.String("user")
	if actor == "" {
		actor = service.GetDid()
	}
	inactiveDays := cmd.Int("inactive")
	mutual := cmd.Bool("mutual")
	quietPosters := cmd.Bool("quiet")
	quietThreshold := cmd.Float("threshold")
	outputFormat := cmd.String("output")

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
		followerInfos = filterInactive(ctx, service, followerInfos, actors, inactiveDays, logger)
	}

	if quietPosters {
		followerInfos = filterQuiet(ctx, service, followerInfos, actors, quietThreshold, logger)
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

	sinceStr := cmd.String("since")
	untilStr := cmd.String("until")

	if sinceStr == "" || untilStr == "" {
		return fmt.Errorf("both --since and --until are required")
	}

	// TODO: Implement snapshot storage and comparison
	// This requires a way to store historical follower lists
	// Options: SQLite table, JSON files with timestamps, etc.

	ui.Infoln("Diff functionality requires snapshot storage (not yet implemented)")
	ui.Infoln("Consider using 'followers export' to create manual snapshots")

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

	actor := cmd.String("user")
	if actor == "" {
		actor = service.GetDid()
	}
	inactiveDays := cmd.Int("inactive")
	quietPosters := cmd.Bool("quiet")
	quietThreshold := cmd.Float("threshold")
	outputFormat := cmd.String("output")

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
		followerInfos = filterInactive(ctx, service, followerInfos, actors, inactiveDays, logger)
	}

	if quietPosters {
		followerInfos = filterQuiet(ctx, service, followerInfos, actors, quietThreshold, logger)
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
			lastPostInfo := "never"
			if info.DaysSincePost >= 0 {
				lastPostInfo = fmt.Sprintf("%d days ago", info.DaysSincePost)
			}
			row = append(row, lastPostInfo)
		} else if info.IsQuiet {
			row = append(row, fmt.Sprintf("%.2f", info.PostsPerDay))
		} else if showInactive {
			lastPostInfo := "never"
			if info.DaysSincePost >= 0 {
				lastPostInfo = fmt.Sprintf("%d days ago", info.DaysSincePost)
			}
			row = append(row, lastPostInfo)
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
				UsageText: "Compare follower lists to identify new followers and unfollows. Requires snapshot storage (not yet implemented).",
				ArgsUsage: " ",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "since",
						Usage:    "Start date (YYYY-MM-DD)",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "until",
						Usage:    "End date (YYYY-MM-DD)",
						Required: true,
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
				},
				Action: FollowersExportAction,
			},
		},
	}
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
				},
				Action: ListFollowingAction,
			},
		},
	}
}
