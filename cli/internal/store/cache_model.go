package store

import "time"

// PostRateCacheModel represents a cached post rate computation for an actor.
// Stores expensive post rate calculations with TTL support (24 hours default).
type PostRateCacheModel struct {
	ActorDid     string
	PostsPerDay  float64
	LastPostDate time.Time
	SampleSize   int
	FetchedAt    time.Time
	ExpiresAt    time.Time
}

// IsFresh returns true if the cached post rate has not expired.
// Post rates expire after 24 hours by default.
func (m *PostRateCacheModel) IsFresh() bool {
	return time.Now().Before(m.ExpiresAt)
}

// ActivityCacheModel represents cached activity data (last post date) for an actor.
// Stores last post date lookups with TTL support (24 hours default).
type ActivityCacheModel struct {
	ActorDid     string
	LastPostDate time.Time // May be zero if actor has never posted
	FetchedAt    time.Time
	ExpiresAt    time.Time
}

// IsFresh returns true if the cached activity data has not expired.
// Activity data expires after 24 hours by default.
func (m *ActivityCacheModel) IsFresh() bool {
	return time.Now().Before(m.ExpiresAt)
}

// HasPosted returns true if the actor has posted at least once.
// A zero LastPostDate indicates the actor has never posted.
func (m *ActivityCacheModel) HasPosted() bool {
	return !m.LastPostDate.IsZero()
}
