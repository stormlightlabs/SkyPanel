package store

import (
	"context"
	"testing"
	"time"

	"github.com/stormlightlabs/skypanel/cli/internal/utils"
)

// TestPostRepository_Init verifies repository initialization
func TestPostRepository_Init(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	repo := &PostRepository{db: db}

	err := repo.Init(context.Background())
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM posts").Scan(&count)
	if err != nil {
		t.Errorf("posts table not created: %v", err)
	}
}

// TestPostRepository_Save creates a new post
func TestPostRepository_Save(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	repo := &PostRepository{db: db}
	if err := repo.Init(context.Background()); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	post := &PostModel{
		URI:       "at://did:plc:test/app.bsky.feed.post/123",
		AuthorDID: "did:plc:test",
		Text:      "Hello, world!",
		FeedID:    "feed-123",
		IndexedAt: time.Now(),
	}

	err := repo.Save(context.Background(), post)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	if post.ID() == "" {
		t.Error("expected ID to be set after Save")
	}
	if post.CreatedAt().IsZero() {
		t.Error("expected CreatedAt to be set after Save")
	}
	if post.UpdatedAt().IsZero() {
		t.Error("expected UpdatedAt to be set after Save")
	}
}

// TestPostRepository_Get retrieves a post by ID
func TestPostRepository_Get(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	repo := &PostRepository{db: db}
	if err := repo.Init(context.Background()); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	indexedTime := time.Now()
	post := &PostModel{
		URI:       "at://did:plc:test/app.bsky.feed.post/456",
		AuthorDID: "did:plc:author",
		Text:      "Test post content",
		FeedID:    "feed-456",
		IndexedAt: indexedTime,
	}

	if err := repo.Save(context.Background(), post); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	retrieved, err := repo.Get(context.Background(), post.ID())
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	retrievedPost, ok := retrieved.(*PostModel)
	if !ok {
		t.Fatal("expected *PostModel")
	}

	if retrievedPost.URI != "at://did:plc:test/app.bsky.feed.post/456" {
		t.Errorf("expected URI 'at://did:plc:test/app.bsky.feed.post/456', got %s", retrievedPost.URI)
	}
	if retrievedPost.AuthorDID != "did:plc:author" {
		t.Errorf("expected AuthorDID 'did:plc:author', got %s", retrievedPost.AuthorDID)
	}
	if retrievedPost.Text != "Test post content" {
		t.Errorf("expected Text 'Test post content', got %s", retrievedPost.Text)
	}
	if retrievedPost.FeedID != "feed-456" {
		t.Errorf("expected FeedID 'feed-456', got %s", retrievedPost.FeedID)
	}
}

// TestPostRepository_Get_NotFound verifies error on missing post
func TestPostRepository_Get_NotFound(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	repo := &PostRepository{db: db}
	if err := repo.Init(context.Background()); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	_, err := repo.Get(context.Background(), "nonexistent-id")
	if err == nil {
		t.Error("expected error for nonexistent post")
	}
}

// TestPostRepository_List retrieves all posts ordered by indexed_at DESC
func TestPostRepository_List(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	repo := &PostRepository{db: db}
	if err := repo.Init(context.Background()); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	now := time.Now()
	post1 := &PostModel{
		URI:       "at://test/post1",
		AuthorDID: "did:plc:1",
		Text:      "Post 1",
		FeedID:    "feed-1",
		IndexedAt: now.Add(-2 * time.Hour),
	}
	post2 := &PostModel{
		URI:       "at://test/post2",
		AuthorDID: "did:plc:2",
		Text:      "Post 2",
		FeedID:    "feed-1",
		IndexedAt: now,
	}

	if err := repo.Save(context.Background(), post1); err != nil {
		t.Fatalf("Save post1 failed: %v", err)
	}
	if err := repo.Save(context.Background(), post2); err != nil {
		t.Fatalf("Save post2 failed: %v", err)
	}

	posts, err := repo.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(posts) != 2 {
		t.Errorf("expected 2 posts, got %d", len(posts))
	}

	// List should order by indexed_at DESC, so post2 should be first
	if p, ok := posts[0].(*PostModel); ok {
		if p.Text != "Post 2" {
			t.Errorf("expected first post to be 'Post 2', got %s", p.Text)
		}
	}
}

// TestPostRepository_Update modifies an existing post
func TestPostRepository_Update(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	repo := &PostRepository{db: db}
	if err := repo.Init(context.Background()); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	post := &PostModel{
		URI:       "at://test/updatepost",
		AuthorDID: "did:plc:author",
		Text:      "Original text",
		FeedID:    "feed-1",
		IndexedAt: time.Now(),
	}

	if err := repo.Save(context.Background(), post); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	originalCreatedAt := post.CreatedAt()

	post.Text = "Updated text"

	if err := repo.Save(context.Background(), post); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if post.CreatedAt() != originalCreatedAt {
		t.Error("CreatedAt should not change on update")
	}

	retrieved, err := repo.Get(context.Background(), post.ID())
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	updatedPost := retrieved.(*PostModel)
	if updatedPost.Text != "Updated text" {
		t.Errorf("expected Text 'Updated text', got %s", updatedPost.Text)
	}
}

// TestPostRepository_Delete removes a post
func TestPostRepository_Delete(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	repo := &PostRepository{db: db}
	if err := repo.Init(context.Background()); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	post := &PostModel{
		URI:       "at://test/deletepost",
		AuthorDID: "did:plc:author",
		Text:      "To be deleted",
		FeedID:    "feed-1",
		IndexedAt: time.Now(),
	}

	if err := repo.Save(context.Background(), post); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	err := repo.Delete(context.Background(), post.ID())
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = repo.Get(context.Background(), post.ID())
	if err == nil {
		t.Error("expected error when getting deleted post")
	}
}

// TestPostRepository_Delete_NotFound verifies error on deleting nonexistent post
func TestPostRepository_Delete_NotFound(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	repo := &PostRepository{db: db}
	if err := repo.Init(context.Background()); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	err := repo.Delete(context.Background(), "nonexistent-id")
	if err == nil {
		t.Error("expected error when deleting nonexistent post")
	}
}

// TestPostRepository_InvalidModelType verifies type checking on Save
func TestPostRepository_InvalidModelType(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	repo := &PostRepository{db: db}
	if err := repo.Init(context.Background()); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	invalidModel := &FeedModel{Name: "test", Source: "timeline", IsLocal: true}

	err := repo.Save(context.Background(), invalidModel)
	if err == nil {
		t.Error("expected error when saving invalid model type")
	}
}

// TestPostRepository_BatchSave inserts multiple posts efficiently
func TestPostRepository_BatchSave(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	repo := &PostRepository{db: db}
	if err := repo.Init(context.Background()); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	now := time.Now()
	posts := []*PostModel{
		{URI: "at://test/batch1", AuthorDID: "did:plc:1", Text: "Batch 1", FeedID: "feed-1", IndexedAt: now},
		{URI: "at://test/batch2", AuthorDID: "did:plc:2", Text: "Batch 2", FeedID: "feed-1", IndexedAt: now},
		{URI: "at://test/batch3", AuthorDID: "did:plc:3", Text: "Batch 3", FeedID: "feed-2", IndexedAt: now},
	}

	err := repo.BatchSave(context.Background(), posts)
	if err != nil {
		t.Fatalf("BatchSave failed: %v", err)
	}

	for _, post := range posts {
		if post.ID() == "" {
			t.Error("expected ID to be set after BatchSave")
		}
	}

	allPosts, err := repo.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(allPosts) != 3 {
		t.Errorf("expected 3 posts, got %d", len(allPosts))
	}
}

// TestPostRepository_BatchSave_Empty verifies handling of empty slice
func TestPostRepository_BatchSave_Empty(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	repo := &PostRepository{db: db}
	if err := repo.Init(context.Background()); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	err := repo.BatchSave(context.Background(), []*PostModel{})
	if err != nil {
		t.Errorf("BatchSave with empty slice should not error: %v", err)
	}
}

// TestPostRepository_URIConflict verifies ON CONFLICT behavior for duplicate URIs
func TestPostRepository_URIConflict(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	repo := &PostRepository{db: db}
	if err := repo.Init(context.Background()); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	post1 := &PostModel{
		URI:       "at://test/conflict",
		AuthorDID: "did:plc:author",
		Text:      "Original text",
		FeedID:    "feed-1",
		IndexedAt: time.Now(),
	}

	if err := repo.Save(context.Background(), post1); err != nil {
		t.Fatalf("Save post1 failed: %v", err)
	}

	post2 := &PostModel{
		URI:       "at://test/conflict",
		AuthorDID: "did:plc:author",
		Text:      "Updated text",
		FeedID:    "feed-2",
		IndexedAt: time.Now(),
	}

	if err := repo.Save(context.Background(), post2); err != nil {
		t.Fatalf("Save post2 failed: %v", err)
	}

	retrieved, err := repo.Get(context.Background(), post1.ID())
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	retrievedPost := retrieved.(*PostModel)
	if retrievedPost.Text != "Updated text" {
		t.Errorf("expected Text 'Updated text', got %s", retrievedPost.Text)
	}
	if retrievedPost.FeedID != "feed-2" {
		t.Errorf("expected FeedID 'feed-2', got %s", retrievedPost.FeedID)
	}
}

// TestPostRepository_QueryByFeedID retrieves posts for a specific feed with pagination
func TestPostRepository_QueryByFeedID(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	repo := &PostRepository{db: db}
	if err := repo.Init(context.Background()); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	now := time.Now()
	posts := []*PostModel{
		{URI: "at://test/f1p1", AuthorDID: "did:plc:1", Text: "Feed 1 Post 1", FeedID: "feed-1", IndexedAt: now.Add(-3 * time.Hour)},
		{URI: "at://test/f1p2", AuthorDID: "did:plc:2", Text: "Feed 1 Post 2", FeedID: "feed-1", IndexedAt: now.Add(-2 * time.Hour)},
		{URI: "at://test/f1p3", AuthorDID: "did:plc:3", Text: "Feed 1 Post 3", FeedID: "feed-1", IndexedAt: now.Add(-1 * time.Hour)},
		{URI: "at://test/f2p1", AuthorDID: "did:plc:4", Text: "Feed 2 Post 1", FeedID: "feed-2", IndexedAt: now},
	}

	if err := repo.BatchSave(context.Background(), posts); err != nil {
		t.Fatalf("BatchSave failed: %v", err)
	}

	feed1Posts, err := repo.QueryByFeedID(context.Background(), "feed-1", 10, 0)
	if err != nil {
		t.Fatalf("QueryByFeedID failed: %v", err)
	}

	if len(feed1Posts) != 3 {
		t.Errorf("expected 3 posts for feed-1, got %d", len(feed1Posts))
	}

	// Posts should be ordered by indexed_at DESC
	if feed1Posts[0].Text != "Feed 1 Post 3" {
		t.Errorf("expected first post to be 'Feed 1 Post 3', got %s", feed1Posts[0].Text)
	}
}

// TestPostRepository_QueryByFeedID_Pagination verifies pagination behavior
func TestPostRepository_QueryByFeedID_Pagination(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	repo := &PostRepository{db: db}
	if err := repo.Init(context.Background()); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	now := time.Now()
	posts := []*PostModel{
		{URI: "at://test/p1", AuthorDID: "did:plc:1", Text: "Post 1", FeedID: "feed-1", IndexedAt: now.Add(-4 * time.Hour)},
		{URI: "at://test/p2", AuthorDID: "did:plc:2", Text: "Post 2", FeedID: "feed-1", IndexedAt: now.Add(-3 * time.Hour)},
		{URI: "at://test/p3", AuthorDID: "did:plc:3", Text: "Post 3", FeedID: "feed-1", IndexedAt: now.Add(-2 * time.Hour)},
		{URI: "at://test/p4", AuthorDID: "did:plc:4", Text: "Post 4", FeedID: "feed-1", IndexedAt: now.Add(-1 * time.Hour)},
		{URI: "at://test/p5", AuthorDID: "did:plc:5", Text: "Post 5", FeedID: "feed-1", IndexedAt: now},
	}

	if err := repo.BatchSave(context.Background(), posts); err != nil {
		t.Fatalf("BatchSave failed: %v", err)
	}

	page1, err := repo.QueryByFeedID(context.Background(), "feed-1", 2, 0)
	if err != nil {
		t.Fatalf("QueryByFeedID page 1 failed: %v", err)
	}

	if len(page1) != 2 {
		t.Errorf("expected 2 posts in page 1, got %d", len(page1))
	}
	if page1[0].Text != "Post 5" {
		t.Errorf("expected first post 'Post 5', got %s", page1[0].Text)
	}

	page2, err := repo.QueryByFeedID(context.Background(), "feed-1", 2, 2)
	if err != nil {
		t.Fatalf("QueryByFeedID page 2 failed: %v", err)
	}

	if len(page2) != 2 {
		t.Errorf("expected 2 posts in page 2, got %d", len(page2))
	}
	if page2[0].Text != "Post 3" {
		t.Errorf("expected first post in page 2 'Post 3', got %s", page2[0].Text)
	}
}

// TestPostRepository_CountByFeedID counts posts for a feed
func TestPostRepository_CountByFeedID(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	repo := &PostRepository{db: db}
	if err := repo.Init(context.Background()); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	now := time.Now()
	posts := []*PostModel{
		{URI: "at://test/c1", AuthorDID: "did:plc:1", Text: "Count 1", FeedID: "feed-1", IndexedAt: now},
		{URI: "at://test/c2", AuthorDID: "did:plc:2", Text: "Count 2", FeedID: "feed-1", IndexedAt: now},
		{URI: "at://test/c3", AuthorDID: "did:plc:3", Text: "Count 3", FeedID: "feed-2", IndexedAt: now},
	}

	if err := repo.BatchSave(context.Background(), posts); err != nil {
		t.Fatalf("BatchSave failed: %v", err)
	}

	count1, err := repo.CountByFeedID(context.Background(), "feed-1")
	if err != nil {
		t.Fatalf("CountByFeedID feed-1 failed: %v", err)
	}

	if count1 != 2 {
		t.Errorf("expected count 2 for feed-1, got %d", count1)
	}

	count2, err := repo.CountByFeedID(context.Background(), "feed-2")
	if err != nil {
		t.Fatalf("CountByFeedID feed-2 failed: %v", err)
	}

	if count2 != 1 {
		t.Errorf("expected count 1 for feed-2, got %d", count2)
	}

	count3, err := repo.CountByFeedID(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("CountByFeedID nonexistent failed: %v", err)
	}

	if count3 != 0 {
		t.Errorf("expected count 0 for nonexistent feed, got %d", count3)
	}
}

// TestPostRepository_Close verifies repository cleanup
func TestPostRepository_Close(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	repo := &PostRepository{db: db}
	if err := repo.Init(context.Background()); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	err := repo.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}
}
