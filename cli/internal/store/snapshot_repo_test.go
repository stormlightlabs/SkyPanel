package store

import (
	"context"
	"testing"
	"time"

	"github.com/stormlightlabs/skypanel/cli/internal/utils"
)

func TestSnapshotRepository_Init(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	repo := &SnapshotRepository{db: db}

	err := repo.Init(context.Background())
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM follower_snapshots").Scan(&count)
	if err != nil {
		t.Errorf("follower_snapshots table not created: %v", err)
	}

	err = db.QueryRow("SELECT COUNT(*) FROM follower_snapshot_entries").Scan(&count)
	if err != nil {
		t.Errorf("follower_snapshot_entries table not created: %v", err)
	}
}

func TestSnapshotRepository_SaveAndGet(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	repo := &SnapshotRepository{db: db}
	if err := repo.Init(context.Background()); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	snapshot := &SnapshotModel{
		UserDid:      "did:plc:testuser",
		SnapshotType: "followers",
		TotalCount:   150,
	}

	err := repo.Save(context.Background(), snapshot)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	if snapshot.ID() == "" {
		t.Error("expected ID to be set after Save")
	}
	if snapshot.CreatedAt().IsZero() {
		t.Error("expected CreatedAt to be set after Save")
	}
	if snapshot.ExpiresAt.IsZero() {
		t.Error("expected ExpiresAt to be set after Save")
	}

	retrieved, err := repo.Get(context.Background(), snapshot.ID())
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	retrievedSnapshot, ok := retrieved.(*SnapshotModel)
	if !ok {
		t.Fatal("expected *SnapshotModel")
	}

	if retrievedSnapshot.UserDid != "did:plc:testuser" {
		t.Errorf("expected UserDid 'did:plc:testuser', got %s", retrievedSnapshot.UserDid)
	}
	if retrievedSnapshot.SnapshotType != "followers" {
		t.Errorf("expected SnapshotType 'followers', got %s", retrievedSnapshot.SnapshotType)
	}
	if retrievedSnapshot.TotalCount != 150 {
		t.Errorf("expected TotalCount 150, got %d", retrievedSnapshot.TotalCount)
	}
}

func TestSnapshotRepository_List(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	repo := &SnapshotRepository{db: db}
	if err := repo.Init(context.Background()); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	snapshot1 := &SnapshotModel{
		UserDid:      "did:plc:user1",
		SnapshotType: "followers",
		TotalCount:   100,
	}
	snapshot2 := &SnapshotModel{
		UserDid:      "did:plc:user1",
		SnapshotType: "following",
		TotalCount:   50,
	}

	if err := repo.Save(context.Background(), snapshot1); err != nil {
		t.Fatalf("Save snapshot1 failed: %v", err)
	}
	time.Sleep(10 * time.Millisecond)
	if err := repo.Save(context.Background(), snapshot2); err != nil {
		t.Fatalf("Save snapshot2 failed: %v", err)
	}

	snapshots, err := repo.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(snapshots) != 2 {
		t.Errorf("expected 2 snapshots, got %d", len(snapshots))
	}

	if s, ok := snapshots[0].(*SnapshotModel); ok {
		if s.SnapshotType != "following" {
			t.Errorf("expected first snapshot to be 'following', got %s", s.SnapshotType)
		}
	}
}

func TestSnapshotRepository_FindByUserAndType(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	repo := &SnapshotRepository{db: db}
	if err := repo.Init(context.Background()); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	snapshot := &SnapshotModel{
		UserDid:      "did:plc:alice",
		SnapshotType: "followers",
		TotalCount:   200,
	}

	err := repo.Save(context.Background(), snapshot)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	retrieved, err := repo.FindByUserAndType(context.Background(), "did:plc:alice", "followers")
	if err != nil {
		t.Fatalf("FindByUserAndType failed: %v", err)
	}

	if retrieved == nil {
		t.Fatal("expected snapshot, got nil")
	}

	if retrieved.UserDid != "did:plc:alice" {
		t.Errorf("expected UserDid 'did:plc:alice', got %s", retrieved.UserDid)
	}
	if retrieved.SnapshotType != "followers" {
		t.Errorf("expected SnapshotType 'followers', got %s", retrieved.SnapshotType)
	}
}

func TestSnapshotRepository_FindByUserAndType_NotFound(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	repo := &SnapshotRepository{db: db}
	if err := repo.Init(context.Background()); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	retrieved, err := repo.FindByUserAndType(context.Background(), "did:plc:nonexistent", "followers")
	if err != nil {
		t.Fatalf("FindByUserAndType failed: %v", err)
	}

	if retrieved != nil {
		t.Error("expected nil for nonexistent snapshot")
	}
}

func TestSnapshotRepository_FindByUserAndType_Expired(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	repo := &SnapshotRepository{db: db}
	if err := repo.Init(context.Background()); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	snapshot := &SnapshotModel{
		UserDid:      "did:plc:bob",
		SnapshotType: "followers",
		TotalCount:   100,
	}
	snapshot.SetID(GenerateUUID())
	snapshot.SetCreatedAt(time.Now().Add(-25 * time.Hour))
	snapshot.ExpiresAt = time.Now().Add(-1 * time.Hour)

	err := repo.Save(context.Background(), snapshot)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	retrieved, err := repo.FindByUserAndType(context.Background(), "did:plc:bob", "followers")
	if err != nil {
		t.Fatalf("FindByUserAndType failed: %v", err)
	}

	if retrieved != nil {
		t.Error("expected nil for expired snapshot")
	}
}

func TestSnapshotRepository_FindByUserTypeAndDate(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	repo := &SnapshotRepository{db: db}
	if err := repo.Init(context.Background()); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	oldSnapshot := &SnapshotModel{
		UserDid:      "did:plc:charlie",
		SnapshotType: "followers",
		TotalCount:   80,
	}
	oldSnapshot.SetID(GenerateUUID())
	oldSnapshot.SetCreatedAt(time.Now().Add(-48 * time.Hour))
	oldSnapshot.ExpiresAt = time.Now().Add(24 * time.Hour)

	recentSnapshot := &SnapshotModel{
		UserDid:      "did:plc:charlie",
		SnapshotType: "followers",
		TotalCount:   100,
	}
	recentSnapshot.SetID(GenerateUUID())
	recentSnapshot.SetCreatedAt(time.Now().Add(-12 * time.Hour))
	recentSnapshot.ExpiresAt = time.Now().Add(24 * time.Hour)

	if err := repo.Save(context.Background(), oldSnapshot); err != nil {
		t.Fatalf("Save oldSnapshot failed: %v", err)
	}
	if err := repo.Save(context.Background(), recentSnapshot); err != nil {
		t.Fatalf("Save recentSnapshot failed: %v", err)
	}

	targetDate := time.Now().Add(-24 * time.Hour)

	retrieved, err := repo.FindByUserTypeAndDate(context.Background(), "did:plc:charlie", "followers", targetDate)
	if err != nil {
		t.Fatalf("FindByUserTypeAndDate failed: %v", err)
	}

	if retrieved == nil {
		t.Fatal("expected snapshot, got nil")
	}

	if retrieved.TotalCount != 80 {
		t.Errorf("expected TotalCount 80 (old snapshot), got %d", retrieved.TotalCount)
	}
}

func TestSnapshotRepository_SaveAndGetEntry(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	repo := &SnapshotRepository{db: db}
	if err := repo.Init(context.Background()); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	snapshot := &SnapshotModel{
		UserDid:      "did:plc:testuser",
		SnapshotType: "followers",
		TotalCount:   1,
	}

	err := repo.Save(context.Background(), snapshot)
	if err != nil {
		t.Fatalf("Save snapshot failed: %v", err)
	}

	entry := &SnapshotEntry{
		SnapshotID: snapshot.ID(),
		ActorDid:   "did:plc:follower1",
		IndexedAt:  "2024-01-15T10:00:00Z",
	}

	err = repo.SaveEntry(context.Background(), entry)
	if err != nil {
		t.Fatalf("SaveEntry failed: %v", err)
	}

	entries, err := repo.GetEntries(context.Background(), snapshot.ID())
	if err != nil {
		t.Fatalf("GetEntries failed: %v", err)
	}

	if len(entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(entries))
	}

	if entries[0].ActorDid != "did:plc:follower1" {
		t.Errorf("expected ActorDid 'did:plc:follower1', got %s", entries[0].ActorDid)
	}
	if entries[0].IndexedAt != "2024-01-15T10:00:00Z" {
		t.Errorf("expected IndexedAt '2024-01-15T10:00:00Z', got %s", entries[0].IndexedAt)
	}
}

func TestSnapshotRepository_SaveEntries_Batch(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	repo := &SnapshotRepository{db: db}
	if err := repo.Init(context.Background()); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	snapshot := &SnapshotModel{
		UserDid:      "did:plc:testuser",
		SnapshotType: "followers",
		TotalCount:   3,
	}

	err := repo.Save(context.Background(), snapshot)
	if err != nil {
		t.Fatalf("Save snapshot failed: %v", err)
	}

	entries := []*SnapshotEntry{
		{
			SnapshotID: snapshot.ID(),
			ActorDid:   "did:plc:follower1",
			IndexedAt:  "2024-01-15T10:00:00Z",
		},
		{
			SnapshotID: snapshot.ID(),
			ActorDid:   "did:plc:follower2",
			IndexedAt:  "2024-01-15T11:00:00Z",
		},
		{
			SnapshotID: snapshot.ID(),
			ActorDid:   "did:plc:follower3",
			IndexedAt:  "2024-01-15T12:00:00Z",
		},
	}

	err = repo.SaveEntries(context.Background(), entries)
	if err != nil {
		t.Fatalf("SaveEntries failed: %v", err)
	}

	retrieved, err := repo.GetEntries(context.Background(), snapshot.ID())
	if err != nil {
		t.Fatalf("GetEntries failed: %v", err)
	}

	if len(retrieved) != 3 {
		t.Errorf("expected 3 entries, got %d", len(retrieved))
	}
}

func TestSnapshotRepository_GetActorDids(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	repo := &SnapshotRepository{db: db}
	if err := repo.Init(context.Background()); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	snapshot := &SnapshotModel{
		UserDid:      "did:plc:testuser",
		SnapshotType: "followers",
		TotalCount:   2,
	}

	err := repo.Save(context.Background(), snapshot)
	if err != nil {
		t.Fatalf("Save snapshot failed: %v", err)
	}

	entries := []*SnapshotEntry{
		{SnapshotID: snapshot.ID(), ActorDid: "did:plc:actor1", IndexedAt: "2024-01-15T10:00:00Z"},
		{SnapshotID: snapshot.ID(), ActorDid: "did:plc:actor2", IndexedAt: "2024-01-15T11:00:00Z"},
	}

	err = repo.SaveEntries(context.Background(), entries)
	if err != nil {
		t.Fatalf("SaveEntries failed: %v", err)
	}

	dids, err := repo.GetActorDids(context.Background(), snapshot.ID())
	if err != nil {
		t.Fatalf("GetActorDids failed: %v", err)
	}

	if len(dids) != 2 {
		t.Errorf("expected 2 DIDs, got %d", len(dids))
	}

	didMap := make(map[string]bool)
	for _, did := range dids {
		didMap[did] = true
	}

	if !didMap["did:plc:actor1"] {
		t.Error("expected 'did:plc:actor1' in results")
	}
	if !didMap["did:plc:actor2"] {
		t.Error("expected 'did:plc:actor2' in results")
	}
}

func TestSnapshotRepository_Delete_CascadesEntries(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		t.Fatalf("failed to enable foreign keys: %v", err)
	}

	repo := &SnapshotRepository{db: db}
	if err := repo.Init(context.Background()); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	snapshot := &SnapshotModel{
		UserDid:      "did:plc:testuser",
		SnapshotType: "followers",
		TotalCount:   1,
	}

	err := repo.Save(context.Background(), snapshot)
	if err != nil {
		t.Fatalf("Save snapshot failed: %v", err)
	}

	entry := &SnapshotEntry{
		SnapshotID: snapshot.ID(),
		ActorDid:   "did:plc:follower1",
		IndexedAt:  "2024-01-15T10:00:00Z",
	}

	err = repo.SaveEntry(context.Background(), entry)
	if err != nil {
		t.Fatalf("SaveEntry failed: %v", err)
	}

	err = repo.Delete(context.Background(), snapshot.ID())
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	entries, err := repo.GetEntries(context.Background(), snapshot.ID())
	if err != nil {
		t.Fatalf("GetEntries failed: %v", err)
	}

	if len(entries) != 0 {
		t.Errorf("expected 0 entries after cascade delete, got %d", len(entries))
	}
}

func TestSnapshotRepository_DeleteExpiredSnapshots(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	repo := &SnapshotRepository{db: db}
	if err := repo.Init(context.Background()); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	freshSnapshot := &SnapshotModel{
		UserDid:      "did:plc:user1",
		SnapshotType: "followers",
		TotalCount:   100,
	}
	freshSnapshot.SetID(GenerateUUID())
	freshSnapshot.SetCreatedAt(time.Now())
	freshSnapshot.ExpiresAt = time.Now().Add(24 * time.Hour)

	expiredSnapshot := &SnapshotModel{
		UserDid:      "did:plc:user1",
		SnapshotType: "following",
		TotalCount:   50,
	}
	expiredSnapshot.SetID(GenerateUUID())
	expiredSnapshot.SetCreatedAt(time.Now().Add(-25 * time.Hour))
	expiredSnapshot.ExpiresAt = time.Now().Add(-1 * time.Hour)

	if err := repo.Save(context.Background(), freshSnapshot); err != nil {
		t.Fatalf("Save freshSnapshot failed: %v", err)
	}
	if err := repo.Save(context.Background(), expiredSnapshot); err != nil {
		t.Fatalf("Save expiredSnapshot failed: %v", err)
	}

	deleted, err := repo.DeleteExpiredSnapshots(context.Background())
	if err != nil {
		t.Fatalf("DeleteExpiredSnapshots failed: %v", err)
	}

	if deleted != 1 {
		t.Errorf("expected 1 deleted snapshot, got %d", deleted)
	}

	snapshots, err := repo.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(snapshots) != 1 {
		t.Errorf("expected 1 snapshot remaining, got %d", len(snapshots))
	}

	if s, ok := snapshots[0].(*SnapshotModel); ok {
		if s.SnapshotType != "followers" {
			t.Errorf("expected remaining snapshot to be 'followers', got %s", s.SnapshotType)
		}
	}
}

func TestSnapshotRepository_IsFresh(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	repo := &SnapshotRepository{db: db}
	if err := repo.Init(context.Background()); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	snapshot := &SnapshotModel{
		UserDid:      "did:plc:testuser",
		SnapshotType: "followers",
		TotalCount:   100,
	}

	err := repo.Save(context.Background(), snapshot)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	if !snapshot.IsFresh() {
		t.Error("newly saved snapshot should be fresh")
	}

	expiredSnapshot := &SnapshotModel{
		UserDid:      "did:plc:testuser2",
		SnapshotType: "followers",
		TotalCount:   50,
	}
	expiredSnapshot.SetID(GenerateUUID())
	expiredSnapshot.SetCreatedAt(time.Now().Add(-25 * time.Hour))
	expiredSnapshot.ExpiresAt = time.Now().Add(-1 * time.Hour)

	if expiredSnapshot.IsFresh() {
		t.Error("expired snapshot should not be fresh")
	}
}

func TestSnapshotRepository_Close(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	repo := &SnapshotRepository{db: db}
	if err := repo.Init(context.Background()); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	err := repo.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}
}
