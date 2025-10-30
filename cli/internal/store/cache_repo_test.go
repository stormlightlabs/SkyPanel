package store

import (
	"context"
	"testing"
	"time"

	"github.com/stormlightlabs/skypanel/cli/internal/utils"
)

func TestCacheRepository_Init(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	repo := &CacheRepository{db: db}

	err := repo.Init(context.Background())
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM cached_post_rates").Scan(&count)
	if err != nil {
		t.Errorf("cached_post_rates table not created: %v", err)
	}

	err = db.QueryRow("SELECT COUNT(*) FROM cached_activity").Scan(&count)
	if err != nil {
		t.Errorf("cached_activity table not created: %v", err)
	}
}

func TestCacheRepository_SaveAndGetPostRate(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	repo := &CacheRepository{db: db}
	if err := repo.Init(context.Background()); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	cache := &PostRateCacheModel{
		ActorDid:     "did:plc:test123",
		PostsPerDay:  2.5,
		LastPostDate: time.Now().Add(-2 * time.Hour),
		SampleSize:   30,
	}

	err := repo.SavePostRate(context.Background(), cache)
	if err != nil {
		t.Fatalf("SavePostRate failed: %v", err)
	}

	if cache.FetchedAt.IsZero() {
		t.Error("expected FetchedAt to be set after Save")
	}
	if cache.ExpiresAt.IsZero() {
		t.Error("expected ExpiresAt to be set after Save")
	}

	retrieved, err := repo.GetPostRate(context.Background(), "did:plc:test123")
	if err != nil {
		t.Fatalf("GetPostRate failed: %v", err)
	}

	if retrieved == nil {
		t.Fatal("expected cache entry, got nil")
	}

	if retrieved.ActorDid != "did:plc:test123" {
		t.Errorf("expected ActorDid 'did:plc:test123', got %s", retrieved.ActorDid)
	}
	if retrieved.PostsPerDay != 2.5 {
		t.Errorf("expected PostsPerDay 2.5, got %f", retrieved.PostsPerDay)
	}
	if retrieved.SampleSize != 30 {
		t.Errorf("expected SampleSize 30, got %d", retrieved.SampleSize)
	}
}

func TestCacheRepository_GetPostRate_NotFound(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	repo := &CacheRepository{db: db}
	if err := repo.Init(context.Background()); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	retrieved, err := repo.GetPostRate(context.Background(), "did:plc:nonexistent")
	if err != nil {
		t.Fatalf("GetPostRate failed: %v", err)
	}

	if retrieved != nil {
		t.Error("expected nil for nonexistent cache entry")
	}
}

func TestCacheRepository_GetPostRate_Expired(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	repo := &CacheRepository{db: db}
	if err := repo.Init(context.Background()); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	cache := &PostRateCacheModel{
		ActorDid:    "did:plc:expired",
		PostsPerDay: 1.0,
		SampleSize:  10,
		FetchedAt:   time.Now().Add(-25 * time.Hour),
		ExpiresAt:   time.Now().Add(-1 * time.Hour),
	}

	err := repo.SavePostRate(context.Background(), cache)
	if err != nil {
		t.Fatalf("SavePostRate failed: %v", err)
	}

	retrieved, err := repo.GetPostRate(context.Background(), "did:plc:expired")
	if err != nil {
		t.Fatalf("GetPostRate failed: %v", err)
	}

	if retrieved != nil {
		t.Error("expected nil for expired cache entry")
	}
}

func TestCacheRepository_SavePostRates_Batch(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	repo := &CacheRepository{db: db}
	if err := repo.Init(context.Background()); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	caches := []*PostRateCacheModel{
		{
			ActorDid:     "did:plc:user1",
			PostsPerDay:  1.5,
			LastPostDate: time.Now().Add(-1 * time.Hour),
			SampleSize:   20,
		},
		{
			ActorDid:     "did:plc:user2",
			PostsPerDay:  3.0,
			LastPostDate: time.Now().Add(-2 * time.Hour),
			SampleSize:   30,
		},
		{
			ActorDid:    "did:plc:user3",
			PostsPerDay: 0.0,
			SampleSize:  0,
		},
	}

	err := repo.SavePostRates(context.Background(), caches)
	if err != nil {
		t.Fatalf("SavePostRates failed: %v", err)
	}

	retrieved, err := repo.GetPostRates(context.Background(), []string{"did:plc:user1", "did:plc:user2", "did:plc:user3"})
	if err != nil {
		t.Fatalf("GetPostRates failed: %v", err)
	}

	if len(retrieved) != 3 {
		t.Errorf("expected 3 cache entries, got %d", len(retrieved))
	}

	if cache, ok := retrieved["did:plc:user1"]; ok {
		if cache.PostsPerDay != 1.5 {
			t.Errorf("expected PostsPerDay 1.5 for user1, got %f", cache.PostsPerDay)
		}
	} else {
		t.Error("expected cache entry for user1")
	}

	if cache, ok := retrieved["did:plc:user3"]; ok {
		if cache.LastPostDate.IsZero() == false {
			t.Error("expected zero LastPostDate for user3 (never posted)")
		}
	}
}

func TestCacheRepository_GetPostRates_PartialMatch(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	repo := &CacheRepository{db: db}
	if err := repo.Init(context.Background()); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	cache := &PostRateCacheModel{
		ActorDid:    "did:plc:cached",
		PostsPerDay: 2.0,
		SampleSize:  25,
	}

	err := repo.SavePostRate(context.Background(), cache)
	if err != nil {
		t.Fatalf("SavePostRate failed: %v", err)
	}

	retrieved, err := repo.GetPostRates(context.Background(), []string{"did:plc:cached", "did:plc:notcached"})
	if err != nil {
		t.Fatalf("GetPostRates failed: %v", err)
	}

	if len(retrieved) != 1 {
		t.Errorf("expected 1 cache entry, got %d", len(retrieved))
	}

	if _, ok := retrieved["did:plc:cached"]; !ok {
		t.Error("expected cache entry for 'did:plc:cached'")
	}

	if _, ok := retrieved["did:plc:notcached"]; ok {
		t.Error("did not expect cache entry for 'did:plc:notcached'")
	}
}

func TestCacheRepository_SaveAndGetActivity(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	repo := &CacheRepository{db: db}
	if err := repo.Init(context.Background()); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	cache := &ActivityCacheModel{
		ActorDid:     "did:plc:active",
		LastPostDate: time.Now().Add(-3 * time.Hour),
	}

	err := repo.SaveActivity(context.Background(), cache)
	if err != nil {
		t.Fatalf("SaveActivity failed: %v", err)
	}

	if cache.FetchedAt.IsZero() {
		t.Error("expected FetchedAt to be set after Save")
	}
	if cache.ExpiresAt.IsZero() {
		t.Error("expected ExpiresAt to be set after Save")
	}

	retrieved, err := repo.GetActivity(context.Background(), "did:plc:active")
	if err != nil {
		t.Fatalf("GetActivity failed: %v", err)
	}

	if retrieved == nil {
		t.Fatal("expected cache entry, got nil")
	}

	if retrieved.ActorDid != "did:plc:active" {
		t.Errorf("expected ActorDid 'did:plc:active', got %s", retrieved.ActorDid)
	}
	if retrieved.LastPostDate.IsZero() {
		t.Error("expected non-zero LastPostDate")
	}
	if !retrieved.HasPosted() {
		t.Error("expected HasPosted to be true")
	}
}

func TestCacheRepository_SaveActivity_NeverPosted(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	repo := &CacheRepository{db: db}
	if err := repo.Init(context.Background()); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	cache := &ActivityCacheModel{
		ActorDid:     "did:plc:neverposted",
		LastPostDate: time.Time{},
	}

	err := repo.SaveActivity(context.Background(), cache)
	if err != nil {
		t.Fatalf("SaveActivity failed: %v", err)
	}

	retrieved, err := repo.GetActivity(context.Background(), "did:plc:neverposted")
	if err != nil {
		t.Fatalf("GetActivity failed: %v", err)
	}

	if retrieved == nil {
		t.Fatal("expected cache entry, got nil")
	}

	if !retrieved.LastPostDate.IsZero() {
		t.Error("expected zero LastPostDate for actor who never posted")
	}
	if retrieved.HasPosted() {
		t.Error("expected HasPosted to be false")
	}
}

func TestCacheRepository_SaveActivities_Batch(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	repo := &CacheRepository{db: db}
	if err := repo.Init(context.Background()); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	caches := []*ActivityCacheModel{
		{
			ActorDid:     "did:plc:actor1",
			LastPostDate: time.Now().Add(-1 * time.Hour),
		},
		{
			ActorDid:     "did:plc:actor2",
			LastPostDate: time.Now().Add(-5 * time.Hour),
		},
		{
			ActorDid:     "did:plc:actor3",
			LastPostDate: time.Time{},
		},
	}

	err := repo.SaveActivities(context.Background(), caches)
	if err != nil {
		t.Fatalf("SaveActivities failed: %v", err)
	}

	retrieved, err := repo.GetActivities(context.Background(), []string{"did:plc:actor1", "did:plc:actor2", "did:plc:actor3"})
	if err != nil {
		t.Fatalf("GetActivities failed: %v", err)
	}

	if len(retrieved) != 3 {
		t.Errorf("expected 3 cache entries, got %d", len(retrieved))
	}

	if cache, ok := retrieved["did:plc:actor1"]; ok {
		if cache.LastPostDate.IsZero() {
			t.Error("expected non-zero LastPostDate for actor1")
		}
	} else {
		t.Error("expected cache entry for actor1")
	}

	if cache, ok := retrieved["did:plc:actor3"]; ok {
		if !cache.LastPostDate.IsZero() {
			t.Error("expected zero LastPostDate for actor3")
		}
	}
}

func TestCacheRepository_Upsert_PostRate(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	repo := &CacheRepository{db: db}
	if err := repo.Init(context.Background()); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	cache := &PostRateCacheModel{
		ActorDid:    "did:plc:upserttest",
		PostsPerDay: 1.0,
		SampleSize:  10,
	}

	err := repo.SavePostRate(context.Background(), cache)
	if err != nil {
		t.Fatalf("SavePostRate failed: %v", err)
	}

	updatedCache := &PostRateCacheModel{
		ActorDid:    "did:plc:upserttest",
		PostsPerDay: 3.0,
		SampleSize:  30,
	}

	err = repo.SavePostRate(context.Background(), updatedCache)
	if err != nil {
		t.Fatalf("Update SavePostRate failed: %v", err)
	}

	retrieved, err := repo.GetPostRate(context.Background(), "did:plc:upserttest")
	if err != nil {
		t.Fatalf("GetPostRate failed: %v", err)
	}

	if retrieved.PostsPerDay != 3.0 {
		t.Errorf("expected PostsPerDay 3.0 after upsert, got %f", retrieved.PostsPerDay)
	}
	if retrieved.SampleSize != 30 {
		t.Errorf("expected SampleSize 30 after upsert, got %d", retrieved.SampleSize)
	}
}

func TestCacheRepository_DeletePostRate(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	repo := &CacheRepository{db: db}
	if err := repo.Init(context.Background()); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	cache := &PostRateCacheModel{
		ActorDid:    "did:plc:todelete",
		PostsPerDay: 2.0,
		SampleSize:  20,
	}

	err := repo.SavePostRate(context.Background(), cache)
	if err != nil {
		t.Fatalf("SavePostRate failed: %v", err)
	}

	err = repo.DeletePostRate(context.Background(), "did:plc:todelete")
	if err != nil {
		t.Fatalf("DeletePostRate failed: %v", err)
	}

	retrieved, err := repo.GetPostRate(context.Background(), "did:plc:todelete")
	if err != nil {
		t.Fatalf("GetPostRate failed: %v", err)
	}
	if retrieved != nil {
		t.Error("expected nil after DeletePostRate")
	}
}

func TestCacheRepository_DeleteExpiredPostRates(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	repo := &CacheRepository{db: db}
	if err := repo.Init(context.Background()); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	freshCache := &PostRateCacheModel{
		ActorDid:    "did:plc:fresh",
		PostsPerDay: 2.0,
		SampleSize:  20,
		FetchedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(24 * time.Hour),
	}

	expiredCache := &PostRateCacheModel{
		ActorDid:    "did:plc:expired",
		PostsPerDay: 1.0,
		SampleSize:  10,
		FetchedAt:   time.Now().Add(-25 * time.Hour),
		ExpiresAt:   time.Now().Add(-1 * time.Hour),
	}

	err := repo.SavePostRate(context.Background(), freshCache)
	if err != nil {
		t.Fatalf("SavePostRate failed: %v", err)
	}

	err = repo.SavePostRate(context.Background(), expiredCache)
	if err != nil {
		t.Fatalf("SavePostRate failed: %v", err)
	}

	deleted, err := repo.DeleteExpiredPostRates(context.Background())
	if err != nil {
		t.Fatalf("DeleteExpiredPostRates failed: %v", err)
	}

	if deleted != 1 {
		t.Errorf("expected 1 deleted entry, got %d", deleted)
	}

	retrieved, err := repo.GetPostRate(context.Background(), "did:plc:fresh")
	if err != nil {
		t.Fatalf("GetPostRate failed: %v", err)
	}
	if retrieved == nil {
		t.Error("expected fresh cache to still exist")
	}

	retrieved, err = repo.GetPostRate(context.Background(), "did:plc:expired")
	if err != nil {
		t.Fatalf("GetPostRate failed: %v", err)
	}
	if retrieved != nil {
		t.Error("expected expired cache to be deleted")
	}
}

func TestCacheRepository_Close(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	repo := &CacheRepository{db: db}
	if err := repo.Init(context.Background()); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	err := repo.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}
}
