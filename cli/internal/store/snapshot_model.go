package store

import "time"

// SnapshotModel represents a follower or following snapshot with metadata.
// Stores snapshot metadata with TTL support (24 hours default).
type SnapshotModel struct {
	id           string
	createdAt    time.Time
	UserDid      string
	SnapshotType string // "followers" or "following"
	TotalCount   int
	ExpiresAt    time.Time
}

func (m *SnapshotModel) ID() string               { return m.id }
func (m *SnapshotModel) CreatedAt() time.Time     { return m.createdAt }
func (m *SnapshotModel) UpdatedAt() time.Time     { return m.createdAt } // Snapshots are immutable
func (m *SnapshotModel) SetID(id string)          { m.id = id }
func (m *SnapshotModel) SetCreatedAt(t time.Time) { m.createdAt = t }
func (m *SnapshotModel) SetUpdatedAt(t time.Time) {} // Snapshots are immutable

// IsFresh returns true if the snapshot has not expired. Snapshots expire after 24 hours by default.
func (m *SnapshotModel) IsFresh() bool {
	return time.Now().Before(m.ExpiresAt)
}

// SnapshotEntry represents an actor in a snapshot with minimal cached data.
// Linked to [SnapshotModel] via snapshot_id foreign key.
type SnapshotEntry struct {
	SnapshotID string
	ActorDid   string
	IndexedAt  string // When the follow relationship was indexed by Bluesky
}
