package store

import "time"

// ProfileModel represents a cached actor profile with TTL support.
// Stores the full ActorProfile as JSON for flexible access to all profile fields.
type ProfileModel struct {
	id        string
	createdAt time.Time
	updatedAt time.Time
	Did       string
	Handle    string
	DataJSON  string    // Serialized ActorProfile for full profile data
	FetchedAt time.Time // Track cache freshness for TTL-based invalidation
}

func (m *ProfileModel) ID() string               { return m.id }
func (m *ProfileModel) CreatedAt() time.Time     { return m.createdAt }
func (m *ProfileModel) UpdatedAt() time.Time     { return m.updatedAt }
func (m *ProfileModel) SetID(id string)          { m.id = id }
func (m *ProfileModel) SetCreatedAt(t time.Time) { m.createdAt = t }
func (m *ProfileModel) SetUpdatedAt(t time.Time) { m.updatedAt = t }
func (m *ProfileModel) TouchUpdatedAt()          { m.updatedAt = time.Now() }

// IsFresh returns true if the profile cache is within the TTL window.
// Default TTL is 1 hour - profiles older than this should be refetched.
func (m *ProfileModel) IsFresh(ttl time.Duration) bool {
	if ttl == 0 {
		ttl = time.Hour
	}
	return time.Since(m.FetchedAt) < ttl
}
