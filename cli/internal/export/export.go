package export

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/stormlightlabs/skypanel/cli/internal/store"
)

// ExportPost represents a post structure for export operations
type ExportPost struct {
	ID        string    `json:"id"`
	URI       string    `json:"uri"`
	AuthorDID string    `json:"author_did"`
	Text      string    `json:"text"`
	FeedID    string    `json:"feed_id"`
	IndexedAt time.Time `json:"indexed_at"`
	CreatedAt time.Time `json:"created_at"`
}

// ToJSON exports posts to JSON format with pretty printing
func ToJSON(filename string, posts []*store.PostModel) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	exportPosts := convertPosts(posts)
	if err := encoder.Encode(exportPosts); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}

// ToCSV exports posts to CSV format with headers
func ToCSV(filename string, posts []*store.PostModel) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	if err := writer.Write([]string{"ID", "URI", "AuthorDID", "Text", "FeedID", "IndexedAt", "CreatedAt"}); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write rows
	for _, post := range posts {
		record := []string{
			post.ID(),
			post.URI,
			post.AuthorDID,
			post.Text,
			post.FeedID,
			post.IndexedAt.Format(time.RFC3339),
			post.CreatedAt().Format(time.RFC3339),
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write CSV record: %w", err)
		}
	}

	return nil
}

// ToTXT exports posts to plain text format with readable formatting
func ToTXT(filename string, posts []*store.PostModel) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	for i, post := range posts {
		fmt.Fprintf(file, "Post #%d\n", i+1)
		fmt.Fprintf(file, "ID: %s\n", post.ID())
		fmt.Fprintf(file, "URI: %s\n", post.URI)
		fmt.Fprintf(file, "Author DID: %s\n", post.AuthorDID)
		fmt.Fprintf(file, "Feed ID: %s\n", post.FeedID)
		fmt.Fprintf(file, "Indexed At: %s\n", post.IndexedAt.Format(time.RFC3339))
		fmt.Fprintf(file, "Created At: %s\n", post.CreatedAt().Format(time.RFC3339))
		fmt.Fprintf(file, "\nText:\n%s\n", post.Text)
		fmt.Fprintf(file, "\n%s\n\n", strings.Repeat("-", 80))
	}

	return nil
}

// convertPosts transforms PostModel slice to ExportPost slice
func convertPosts(posts []*store.PostModel) []ExportPost {
	exportPosts := make([]ExportPost, len(posts))
	for i, post := range posts {
		exportPosts[i] = ExportPost{
			ID:        post.ID(),
			URI:       post.URI,
			AuthorDID: post.AuthorDID,
			Text:      post.Text,
			FeedID:    post.FeedID,
			IndexedAt: post.IndexedAt,
			CreatedAt: post.CreatedAt(),
		}
	}
	return exportPosts
}
