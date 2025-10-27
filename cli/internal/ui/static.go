package ui

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/stormlightlabs/skypanel/cli/internal/store"
)

// DisplayProfileHeader shows a formatted profile summary
func DisplayProfileHeader(profile *store.ActorProfile) {
	Titleln("@%s", profile.Handle)
	if profile.DisplayName != "" {
		Subtitleln("%s", profile.DisplayName)
	}
	if profile.Description != "" {
		fmt.Printf("  %s\n", profile.Description)
	}
	Infoln("  Followers: %d | Following: %d | Posts: %d", profile.FollowersCount, profile.FollowsCount, profile.PostsCount)

	fmt.Println()
}

// DisplayFeed shows a formatted list of posts from a feed
func DisplayFeed(feed []store.FeedViewPost, cursor string) {
	if len(feed) == 0 {
		Infoln("No posts found.")
		return
	}

	for i, item := range feed {
		post := item.Post
		if post == nil {
			continue
		}

		Subtitleln("[%d] Post by @%s", i+1, post.Author.Handle)
		Infoln("  URI: %s", post.Uri)

		if recordMap, ok := post.Record.(map[string]any); ok {
			if text, ok := recordMap["text"].(string); ok {
				displayText := text
				if len(displayText) > 200 {
					displayText = displayText[:200] + "..."
				}
				fmt.Printf("  %s\n", displayText)
			}
		}

		Infoln("  ❤️  %d | 🔁 %d | 💬 %d", post.LikeCount, post.RepostCount, post.ReplyCount)

		if item.Reason != nil && item.Reason.By != nil {
			Infoln("  ↻ Reposted by @%s", item.Reason.By.Handle)
		}

		Infoln("  Indexed: %s", post.IndexedAt)
		fmt.Println()
	}

	Successln("Showing %d post(s)", len(feed))
	if cursor != "" {
		Infoln("Next cursor: %s", cursor)
	}
}

// DisplayJSON marshals and prints data as JSON
func DisplayJSON(data any) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}
