package store

import (
	"context"
	"testing"
	"time"

	"github.com/stormlightlabs/skypanel/cli/internal/utils"
)

// TestFeedRepository_Init verifies repository initialization
func TestFeedRepository_Init(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	repo := &FeedRepository{db: db}

	err := repo.Init(context.Background())
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM feeds").Scan(&count)
	if err != nil {
		t.Errorf("feeds table not created: %v", err)
	}
}

// TestFeedRepository_Save creates a new feed
func TestFeedRepository_Save(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	repo := &FeedRepository{db: db}
	if err := repo.Init(context.Background()); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	feed := &FeedModel{
		Name:    "Test Feed",
		Source:  "timeline",
		Params:  map[string]string{"key": "value"},
		IsLocal: true,
	}

	err := repo.Save(context.Background(), feed)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	if feed.ID() == "" {
		t.Error("expected ID to be set after Save")
	}
	if feed.CreatedAt().IsZero() {
		t.Error("expected CreatedAt to be set after Save")
	}
	if feed.UpdatedAt().IsZero() {
		t.Error("expected UpdatedAt to be set after Save")
	}
}

// TestFeedRepository_Get retrieves a feed by ID
func TestFeedRepository_Get(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	repo := &FeedRepository{db: db}
	if err := repo.Init(context.Background()); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	feed := &FeedModel{
		Name:    "Test Feed",
		Source:  "following",
		Params:  map[string]string{"limit": "50"},
		IsLocal: false,
	}

	if err := repo.Save(context.Background(), feed); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	retrieved, err := repo.Get(context.Background(), feed.ID())
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	retrievedFeed, ok := retrieved.(*FeedModel)
	if !ok {
		t.Fatal("expected *FeedModel")
	}

	if retrievedFeed.Name != "Test Feed" {
		t.Errorf("expected Name 'Test Feed', got %s", retrievedFeed.Name)
	}
	if retrievedFeed.Source != "following" {
		t.Errorf("expected Source 'following', got %s", retrievedFeed.Source)
	}
	if retrievedFeed.IsLocal != false {
		t.Error("expected IsLocal false")
	}
	if retrievedFeed.Params["limit"] != "50" {
		t.Errorf("expected Params['limit'] '50', got %s", retrievedFeed.Params["limit"])
	}
}

// TestFeedRepository_Get_NotFound verifies error on missing feed
func TestFeedRepository_Get_NotFound(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	repo := &FeedRepository{db: db}
	if err := repo.Init(context.Background()); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	_, err := repo.Get(context.Background(), "nonexistent-id")
	if err == nil {
		t.Error("expected error for nonexistent feed")
	}
}

// TestFeedRepository_List retrieves all feeds
func TestFeedRepository_List(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	repo := &FeedRepository{db: db}
	if err := repo.Init(context.Background()); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	feed1 := &FeedModel{Name: "Feed 1", Source: "timeline", IsLocal: true}
	feed2 := &FeedModel{Name: "Feed 2", Source: "following", IsLocal: false}

	if err := repo.Save(context.Background(), feed1); err != nil {
		t.Fatalf("Save feed1 failed: %v", err)
	}
	time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	if err := repo.Save(context.Background(), feed2); err != nil {
		t.Fatalf("Save feed2 failed: %v", err)
	}

	feeds, err := repo.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(feeds) != 2 {
		t.Errorf("expected 2 feeds, got %d", len(feeds))
	}

	// List should order by created_at DESC, so feed2 should be first
	if f, ok := feeds[0].(*FeedModel); ok {
		if f.Name != "Feed 2" {
			t.Errorf("expected first feed to be 'Feed 2', got %s", f.Name)
		}
	}
}

// TestFeedRepository_Update modifies an existing feed
func TestFeedRepository_Update(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	repo := &FeedRepository{db: db}
	if err := repo.Init(context.Background()); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	feed := &FeedModel{
		Name:    "Original Name",
		Source:  "timeline",
		IsLocal: true,
	}

	if err := repo.Save(context.Background(), feed); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	originalID := feed.ID()
	originalCreatedAt := feed.CreatedAt()

	feed.Name = "Updated Name"
	feed.Source = "following"

	if err := repo.Save(context.Background(), feed); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if feed.ID() != originalID {
		t.Error("ID should not change on update")
	}
	if feed.CreatedAt() != originalCreatedAt {
		t.Error("CreatedAt should not change on update")
	}

	retrieved, err := repo.Get(context.Background(), feed.ID())
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	updatedFeed := retrieved.(*FeedModel)
	if updatedFeed.Name != "Updated Name" {
		t.Errorf("expected Name 'Updated Name', got %s", updatedFeed.Name)
	}
	if updatedFeed.Source != "following" {
		t.Errorf("expected Source 'following', got %s", updatedFeed.Source)
	}
}

// TestFeedRepository_Delete removes a feed
func TestFeedRepository_Delete(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	repo := &FeedRepository{db: db}
	if err := repo.Init(context.Background()); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	feed := &FeedModel{Name: "To Delete", Source: "timeline", IsLocal: true}

	if err := repo.Save(context.Background(), feed); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	err := repo.Delete(context.Background(), feed.ID())
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = repo.Get(context.Background(), feed.ID())
	if err == nil {
		t.Error("expected error when getting deleted feed")
	}
}

// TestFeedRepository_Delete_NotFound verifies error on deleting nonexistent feed
func TestFeedRepository_Delete_NotFound(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	repo := &FeedRepository{db: db}
	if err := repo.Init(context.Background()); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	err := repo.Delete(context.Background(), "nonexistent-id")
	if err == nil {
		t.Error("expected error when deleting nonexistent feed")
	}
}

// TestFeedRepository_InvalidModelType verifies type checking on Save
func TestFeedRepository_InvalidModelType(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	repo := &FeedRepository{db: db}
	if err := repo.Init(context.Background()); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	invalidModel := &PostModel{URI: "test", AuthorDID: "did:test", Text: "hello"}

	err := repo.Save(context.Background(), invalidModel)
	if err == nil {
		t.Error("expected error when saving invalid model type")
	}
}

// TestFeedRepository_ParamsJSONMarshaling verifies Params field marshaling
func TestFeedRepository_ParamsJSONMarshaling(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	repo := &FeedRepository{db: db}
	if err := repo.Init(context.Background()); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	feed := &FeedModel{
		Name:   "Complex Feed",
		Source: "custom",
		Params: map[string]string{
			"key1": "value1",
			"key2": "value2",
			"key3": "value3",
		},
		IsLocal: true,
	}

	if err := repo.Save(context.Background(), feed); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	retrieved, err := repo.Get(context.Background(), feed.ID())
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	retrievedFeed := retrieved.(*FeedModel)
	if len(retrievedFeed.Params) != 3 {
		t.Errorf("expected 3 params, got %d", len(retrievedFeed.Params))
	}
	if retrievedFeed.Params["key1"] != "value1" {
		t.Errorf("expected Params['key1'] 'value1', got %s", retrievedFeed.Params["key1"])
	}
	if retrievedFeed.Params["key2"] != "value2" {
		t.Errorf("expected Params['key2'] 'value2', got %s", retrievedFeed.Params["key2"])
	}
	if retrievedFeed.Params["key3"] != "value3" {
		t.Errorf("expected Params['key3'] 'value3', got %s", retrievedFeed.Params["key3"])
	}
}

// TestFeedRepository_EmptyParams verifies handling of empty Params
func TestFeedRepository_EmptyParams(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	repo := &FeedRepository{db: db}
	if err := repo.Init(context.Background()); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	feed := &FeedModel{
		Name:    "No Params Feed",
		Source:  "simple",
		Params:  map[string]string{},
		IsLocal: false,
	}

	if err := repo.Save(context.Background(), feed); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	retrieved, err := repo.Get(context.Background(), feed.ID())
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	retrievedFeed := retrieved.(*FeedModel)
	if retrievedFeed.Params == nil {
		t.Error("expected non-nil Params map")
	}
	if len(retrievedFeed.Params) != 0 {
		t.Errorf("expected 0 params, got %d", len(retrievedFeed.Params))
	}
}

// TestFeedRepository_Close verifies repository cleanup
func TestFeedRepository_Close(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	repo := &FeedRepository{db: db}
	if err := repo.Init(context.Background()); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	err := repo.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}
}
