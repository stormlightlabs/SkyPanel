package export

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stormlightlabs/skypanel/cli/internal/store"
)

// createTestPosts generates sample posts for testing
func createTestPosts() []*store.PostModel {
	now := time.Now()
	posts := []*store.PostModel{
		{
			URI:       "at://did:plc:test1/app.bsky.feed.post/1",
			AuthorDID: "did:plc:author1",
			Text:      "First test post",
			FeedID:    "feed-1",
			IndexedAt: now.Add(-2 * time.Hour),
		},
		{
			URI:       "at://did:plc:test2/app.bsky.feed.post/2",
			AuthorDID: "did:plc:author2",
			Text:      "Second test post",
			FeedID:    "feed-1",
			IndexedAt: now.Add(-1 * time.Hour),
		},
		{
			URI:       "at://did:plc:test3/app.bsky.feed.post/3",
			AuthorDID: "did:plc:author3",
			Text:      "Third test post with special chars: \"quotes\", commas, and\nnewlines",
			FeedID:    "feed-2",
			IndexedAt: now,
		},
	}

	// Set IDs and timestamps
	for i, post := range posts {
		post.SetID(string(rune('a' + i)))
		post.SetCreatedAt(now.Add(time.Duration(-i) * time.Hour))
		post.SetUpdatedAt(now)
	}

	return posts
}

// TestToJSON_Success verifies JSON export with valid data
func TestToJSON_Success(t *testing.T) {
	posts := createTestPosts()

	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.json")

	err := ToJSON(filename, posts)
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Fatal("exported file does not exist")
	}

	// Read and parse JSON
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("failed to read exported file: %v", err)
	}

	var exportedPosts []ExportPost
	if err := json.Unmarshal(data, &exportedPosts); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if len(exportedPosts) != 3 {
		t.Errorf("expected 3 posts, got %d", len(exportedPosts))
	}

	// Verify first post content
	if exportedPosts[0].URI != "at://did:plc:test1/app.bsky.feed.post/1" {
		t.Errorf("unexpected URI: %s", exportedPosts[0].URI)
	}
	if exportedPosts[0].Text != "First test post" {
		t.Errorf("unexpected text: %s", exportedPosts[0].Text)
	}
}

// TestToJSON_EmptyPosts verifies JSON export with empty slice
func TestToJSON_EmptyPosts(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "empty.json")

	err := ToJSON(filename, []*store.PostModel{})
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("failed to read exported file: %v", err)
	}

	var exportedPosts []ExportPost
	if err := json.Unmarshal(data, &exportedPosts); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if len(exportedPosts) != 0 {
		t.Errorf("expected 0 posts, got %d", len(exportedPosts))
	}
}

// TestToJSON_InvalidPath verifies error handling for invalid file paths
func TestToJSON_InvalidPath(t *testing.T) {
	posts := createTestPosts()

	err := ToJSON("/invalid/path/that/does/not/exist/test.json", posts)
	if err == nil {
		t.Error("expected error for invalid path, got nil")
	}
}

// TestToCSV_Success verifies CSV export with valid data
func TestToCSV_Success(t *testing.T) {
	posts := createTestPosts()

	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.csv")

	err := ToCSV(filename, posts)
	if err != nil {
		t.Fatalf("ToCSV failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Fatal("exported file does not exist")
	}

	// Read and parse CSV
	file, err := os.Open(filename)
	if err != nil {
		t.Fatalf("failed to open exported file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("failed to parse CSV: %v", err)
	}

	// Verify header + 3 data rows
	if len(records) != 4 {
		t.Errorf("expected 4 rows (header + 3 data), got %d", len(records))
	}

	// Verify header
	expectedHeader := []string{"ID", "URI", "AuthorDID", "Text", "FeedID", "IndexedAt", "CreatedAt"}
	for i, col := range expectedHeader {
		if records[0][i] != col {
			t.Errorf("header column %d: expected %s, got %s", i, col, records[0][i])
		}
	}

	// Verify first data row
	if records[1][1] != "at://did:plc:test1/app.bsky.feed.post/1" {
		t.Errorf("unexpected URI in row 1: %s", records[1][1])
	}
	if records[1][3] != "First test post" {
		t.Errorf("unexpected text in row 1: %s", records[1][3])
	}
}

// TestToCSV_SpecialCharacters verifies CSV escaping of special characters
func TestToCSV_SpecialCharacters(t *testing.T) {
	posts := createTestPosts()

	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "special.csv")

	err := ToCSV(filename, posts)
	if err != nil {
		t.Fatalf("ToCSV failed: %v", err)
	}

	file, err := os.Open(filename)
	if err != nil {
		t.Fatalf("failed to open exported file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("failed to parse CSV: %v", err)
	}

	// Third post has special characters
	specialText := records[3][3]
	if !strings.Contains(specialText, "quotes") {
		t.Error("special characters not properly preserved in CSV")
	}
	if !strings.Contains(specialText, "newlines") {
		t.Error("newlines not properly preserved in CSV")
	}
}

// TestToCSV_EmptyPosts verifies CSV export with empty slice
func TestToCSV_EmptyPosts(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "empty.csv")

	err := ToCSV(filename, []*store.PostModel{})
	if err != nil {
		t.Fatalf("ToCSV failed: %v", err)
	}

	file, err := os.Open(filename)
	if err != nil {
		t.Fatalf("failed to open exported file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("failed to parse CSV: %v", err)
	}

	// Should only have header row
	if len(records) != 1 {
		t.Errorf("expected 1 row (header only), got %d", len(records))
	}
}

// TestToCSV_InvalidPath verifies error handling for invalid file paths
func TestToCSV_InvalidPath(t *testing.T) {
	posts := createTestPosts()

	err := ToCSV("/invalid/path/that/does/not/exist/test.csv", posts)
	if err == nil {
		t.Error("expected error for invalid path, got nil")
	}
}

// TestToTXT_Success verifies TXT export with valid data
func TestToTXT_Success(t *testing.T) {
	posts := createTestPosts()

	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.txt")

	err := ToTXT(filename, posts)
	if err != nil {
		t.Fatalf("ToTXT failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Fatal("exported file does not exist")
	}

	// Read file content
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("failed to read exported file: %v", err)
	}

	content := string(data)

	// Verify content contains expected elements
	if !strings.Contains(content, "Post #1") {
		t.Error("missing post number")
	}
	if !strings.Contains(content, "at://did:plc:test1/app.bsky.feed.post/1") {
		t.Error("missing URI")
	}
	if !strings.Contains(content, "First test post") {
		t.Error("missing post text")
	}
	if !strings.Contains(content, "did:plc:author1") {
		t.Error("missing author DID")
	}
	if !strings.Contains(content, strings.Repeat("-", 80)) {
		t.Error("missing separator")
	}
}

// TestToTXT_EmptyPosts verifies TXT export with empty slice
func TestToTXT_EmptyPosts(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "empty.txt")

	err := ToTXT(filename, []*store.PostModel{})
	if err != nil {
		t.Fatalf("ToTXT failed: %v", err)
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("failed to read exported file: %v", err)
	}

	if len(data) != 0 {
		t.Errorf("expected empty file, got %d bytes", len(data))
	}
}

// TestToTXT_InvalidPath verifies error handling for invalid file paths
func TestToTXT_InvalidPath(t *testing.T) {
	posts := createTestPosts()

	err := ToTXT("/invalid/path/that/does/not/exist/test.txt", posts)
	if err == nil {
		t.Error("expected error for invalid path, got nil")
	}
}

// TestToTXT_MultiplePostsFormatting verifies proper formatting of multiple posts
func TestToTXT_MultiplePostsFormatting(t *testing.T) {
	posts := createTestPosts()

	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "multiple.txt")

	err := ToTXT(filename, posts)
	if err != nil {
		t.Fatalf("ToTXT failed: %v", err)
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("failed to read exported file: %v", err)
	}

	content := string(data)

	// Verify all posts are present
	for i := 1; i <= 3; i++ {
		postNum := "Post #" + string(rune('0'+i))
		if !strings.Contains(content, postNum) {
			t.Errorf("missing %s", postNum)
		}
	}

	// Count separators (should have 3)
	separatorCount := strings.Count(content, strings.Repeat("-", 80))
	if separatorCount != 3 {
		t.Errorf("expected 3 separators, got %d", separatorCount)
	}
}

// TestConvertPosts verifies post model conversion
func TestConvertPosts(t *testing.T) {
	posts := createTestPosts()

	exportPosts := convertPosts(posts)

	if len(exportPosts) != len(posts) {
		t.Errorf("expected %d posts, got %d", len(posts), len(exportPosts))
	}

	for i := range posts {
		if exportPosts[i].ID != posts[i].ID() {
			t.Errorf("post %d: ID mismatch", i)
		}
		if exportPosts[i].URI != posts[i].URI {
			t.Errorf("post %d: URI mismatch", i)
		}
		if exportPosts[i].Text != posts[i].Text {
			t.Errorf("post %d: Text mismatch", i)
		}
		if exportPosts[i].AuthorDID != posts[i].AuthorDID {
			t.Errorf("post %d: AuthorDID mismatch", i)
		}
		if exportPosts[i].FeedID != posts[i].FeedID {
			t.Errorf("post %d: FeedID mismatch", i)
		}
	}
}

// TestExportPost_JSONTags verifies JSON struct tags
func TestExportPost_JSONTags(t *testing.T) {
	now := time.Now()
	post := ExportPost{
		ID:        "test-id",
		URI:       "at://test/uri",
		AuthorDID: "did:plc:test",
		Text:      "test text",
		FeedID:    "feed-id",
		IndexedAt: now,
		CreatedAt: now,
	}

	data, err := json.Marshal(post)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	content := string(data)

	// Verify JSON uses snake_case
	expectedFields := []string{
		"\"id\":",
		"\"uri\":",
		"\"author_did\":",
		"\"text\":",
		"\"feed_id\":",
		"\"indexed_at\":",
		"\"created_at\":",
	}

	for _, field := range expectedFields {
		if !strings.Contains(content, field) {
			t.Errorf("missing expected JSON field: %s", field)
		}
	}
}

// TestToJSON_SinglePost verifies export with single post
func TestToJSON_SinglePost(t *testing.T) {
	now := time.Now()
	post := &store.PostModel{
		URI:       "at://test/single",
		AuthorDID: "did:plc:single",
		Text:      "Single post",
		FeedID:    "feed-1",
		IndexedAt: now,
	}
	post.SetID("single-id")
	post.SetCreatedAt(now)

	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "single.json")

	err := ToJSON(filename, []*store.PostModel{post})
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	var exportedPosts []ExportPost
	if err := json.Unmarshal(data, &exportedPosts); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if len(exportedPosts) != 1 {
		t.Errorf("expected 1 post, got %d", len(exportedPosts))
	}

	if exportedPosts[0].Text != "Single post" {
		t.Errorf("unexpected text: %s", exportedPosts[0].Text)
	}
}
