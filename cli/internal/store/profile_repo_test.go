package store

import (
	"context"
	"testing"
	"time"

	"github.com/stormlightlabs/skypanel/cli/internal/utils"
)

func TestProfileRepository_Init(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	repo := &ProfileRepository{db: db}

	err := repo.Init(context.Background())
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM profiles").Scan(&count)
	if err != nil {
		t.Errorf("profiles table not created: %v", err)
	}
}

func TestProfileRepository_Save(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	repo := &ProfileRepository{db: db}
	if err := repo.Init(context.Background()); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	profile := &ProfileModel{
		Did:       "did:plc:test123",
		Handle:    "alice.bsky.social",
		DataJSON:  `{"did":"did:plc:test123","handle":"alice.bsky.social"}`,
		FetchedAt: time.Now(),
	}

	err := repo.Save(context.Background(), profile)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	if profile.ID() == "" {
		t.Error("expected ID to be set after Save")
	}
	if profile.CreatedAt().IsZero() {
		t.Error("expected CreatedAt to be set after Save")
	}
	if profile.UpdatedAt().IsZero() {
		t.Error("expected UpdatedAt to be set after Save")
	}
}

func TestProfileRepository_Get(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	repo := &ProfileRepository{db: db}
	if err := repo.Init(context.Background()); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	profile := &ProfileModel{
		Did:       "did:plc:bob456",
		Handle:    "bob.bsky.social",
		DataJSON:  `{"did":"did:plc:bob456","handle":"bob.bsky.social","followersCount":100}`,
		FetchedAt: time.Now(),
	}

	if err := repo.Save(context.Background(), profile); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	retrieved, err := repo.Get(context.Background(), profile.ID())
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	retrievedProfile, ok := retrieved.(*ProfileModel)
	if !ok {
		t.Fatal("expected *ProfileModel")
	}

	if retrievedProfile.Did != "did:plc:bob456" {
		t.Errorf("expected Did 'did:plc:bob456', got %s", retrievedProfile.Did)
	}
	if retrievedProfile.Handle != "bob.bsky.social" {
		t.Errorf("expected Handle 'bob.bsky.social', got %s", retrievedProfile.Handle)
	}
	if retrievedProfile.DataJSON == "" {
		t.Error("expected non-empty DataJSON")
	}
}

func TestProfileRepository_GetByDid(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	repo := &ProfileRepository{db: db}
	if err := repo.Init(context.Background()); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	profile := &ProfileModel{
		Did:       "did:plc:charlie789",
		Handle:    "charlie.bsky.social",
		DataJSON:  `{"did":"did:plc:charlie789"}`,
		FetchedAt: time.Now(),
	}

	if err := repo.Save(context.Background(), profile); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	retrieved, err := repo.GetByDid(context.Background(), "did:plc:charlie789")
	if err != nil {
		t.Fatalf("GetByDid failed: %v", err)
	}

	if retrieved == nil {
		t.Fatal("expected profile, got nil")
	}

	if retrieved.Did != "did:plc:charlie789" {
		t.Errorf("expected Did 'did:plc:charlie789', got %s", retrieved.Did)
	}
	if retrieved.Handle != "charlie.bsky.social" {
		t.Errorf("expected Handle 'charlie.bsky.social', got %s", retrieved.Handle)
	}
}

func TestProfileRepository_GetByDid_NotFound(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	repo := &ProfileRepository{db: db}
	if err := repo.Init(context.Background()); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	retrieved, err := repo.GetByDid(context.Background(), "did:plc:nonexistent")
	if err != nil {
		t.Fatalf("GetByDid failed: %v", err)
	}

	if retrieved != nil {
		t.Error("expected nil for nonexistent profile")
	}
}

func TestProfileRepository_Upsert_ByDid(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	repo := &ProfileRepository{db: db}
	if err := repo.Init(context.Background()); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	profile := &ProfileModel{
		Did:       "did:plc:diana999",
		Handle:    "diana.bsky.social",
		DataJSON:  `{"followersCount":50}`,
		FetchedAt: time.Now().Add(-1 * time.Hour),
	}

	if err := repo.Save(context.Background(), profile); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	firstID := profile.ID()

	updatedProfile := &ProfileModel{
		Did:       "did:plc:diana999",
		Handle:    "diana.updated.bsky.social",
		DataJSON:  `{"followersCount":100}`,
		FetchedAt: time.Now(),
	}

	if err := repo.Save(context.Background(), updatedProfile); err != nil {
		t.Fatalf("Update save failed: %v", err)
	}

	profiles, err := repo.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(profiles) != 1 {
		t.Errorf("expected 1 profile after upsert, got %d", len(profiles))
	}

	retrieved, err := repo.GetByDid(context.Background(), "did:plc:diana999")
	if err != nil {
		t.Fatalf("GetByDid failed: %v", err)
	}

	if retrieved.Handle != "diana.updated.bsky.social" {
		t.Errorf("expected Handle 'diana.updated.bsky.social', got %s", retrieved.Handle)
	}

	if retrieved.DataJSON != `{"followersCount":100}` {
		t.Errorf("expected updated DataJSON, got %s", retrieved.DataJSON)
	}

	if retrieved.ID() != firstID {
		t.Errorf("expected ID to remain %s after upsert, got %s", firstID, retrieved.ID())
	}
}

func TestProfileRepository_List(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	repo := &ProfileRepository{db: db}
	if err := repo.Init(context.Background()); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	profile1 := &ProfileModel{
		Did:       "did:plc:user1",
		Handle:    "user1.bsky.social",
		DataJSON:  `{}`,
		FetchedAt: time.Now().Add(-1 * time.Hour),
	}
	profile2 := &ProfileModel{
		Did:       "did:plc:user2",
		Handle:    "user2.bsky.social",
		DataJSON:  `{}`,
		FetchedAt: time.Now(),
	}

	if err := repo.Save(context.Background(), profile1); err != nil {
		t.Fatalf("Save profile1 failed: %v", err)
	}
	time.Sleep(10 * time.Millisecond)
	if err := repo.Save(context.Background(), profile2); err != nil {
		t.Fatalf("Save profile2 failed: %v", err)
	}

	profiles, err := repo.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(profiles) != 2 {
		t.Errorf("expected 2 profiles, got %d", len(profiles))
	}

	if p, ok := profiles[0].(*ProfileModel); ok {
		if p.Handle != "user2.bsky.social" {
			t.Errorf("expected first profile to be 'user2.bsky.social', got %s", p.Handle)
		}
	}
}

func TestProfileRepository_Delete(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	repo := &ProfileRepository{db: db}
	if err := repo.Init(context.Background()); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	profile := &ProfileModel{
		Did:       "did:plc:todelete",
		Handle:    "todelete.bsky.social",
		DataJSON:  `{}`,
		FetchedAt: time.Now(),
	}

	if err := repo.Save(context.Background(), profile); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	err := repo.Delete(context.Background(), profile.ID())
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = repo.Get(context.Background(), profile.ID())
	if err == nil {
		t.Error("expected error when getting deleted profile")
	}
}

func TestProfileRepository_DeleteByDid(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	repo := &ProfileRepository{db: db}
	if err := repo.Init(context.Background()); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	profile := &ProfileModel{
		Did:       "did:plc:deletebydid",
		Handle:    "deletebydid.bsky.social",
		DataJSON:  `{}`,
		FetchedAt: time.Now(),
	}

	if err := repo.Save(context.Background(), profile); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	err := repo.DeleteByDid(context.Background(), "did:plc:deletebydid")
	if err != nil {
		t.Fatalf("DeleteByDid failed: %v", err)
	}

	retrieved, err := repo.GetByDid(context.Background(), "did:plc:deletebydid")
	if err != nil {
		t.Fatalf("GetByDid failed: %v", err)
	}
	if retrieved != nil {
		t.Error("expected nil after DeleteByDid")
	}
}

func TestProfileRepository_DeleteByDid_NotFound(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	repo := &ProfileRepository{db: db}
	if err := repo.Init(context.Background()); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	err := repo.DeleteByDid(context.Background(), "did:plc:nonexistent")
	if err == nil {
		t.Error("expected error when deleting nonexistent profile by DID")
	}
}

func TestProfileRepository_InvalidModelType(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	repo := &ProfileRepository{db: db}
	if err := repo.Init(context.Background()); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	invalidModel := &FeedModel{Name: "test", Source: "test"}

	err := repo.Save(context.Background(), invalidModel)
	if err == nil {
		t.Error("expected error when saving invalid model type")
	}
}

func TestProfileRepository_CacheFreshnessScenario(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	repo := &ProfileRepository{db: db}
	if err := repo.Init(context.Background()); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	profile := &ProfileModel{
		Did:       "did:plc:freshtest",
		Handle:    "freshtest.bsky.social",
		DataJSON:  `{"followers":100}`,
		FetchedAt: time.Now().Add(-30 * time.Minute),
	}

	if err := repo.Save(context.Background(), profile); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	retrieved, err := repo.GetByDid(context.Background(), "did:plc:freshtest")
	if err != nil {
		t.Fatalf("GetByDid failed: %v", err)
	}

	if !retrieved.IsFresh(time.Hour) {
		t.Error("profile should be fresh within 1 hour")
	}

	if retrieved.IsFresh(15 * time.Minute) {
		t.Error("profile should be stale with 15 minute TTL")
	}
}

func TestProfileRepository_Close(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	repo := &ProfileRepository{db: db}
	if err := repo.Init(context.Background()); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	err := repo.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}
}
