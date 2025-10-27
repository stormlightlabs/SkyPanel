package store

import (
	"time"

	"github.com/google/uuid"
)

// Model is the base interface for any persisted domain object.
type Model interface {
	ID() string
	CreatedAt() time.Time
	UpdatedAt() time.Time
}

// Generates a new v4 [uuid.UUID] and converts it to a string
func GenerateUUID() string {
	return uuid.New().String()
}

// FeedModel represents a logical feed (local or remote).
type FeedModel struct {
	ID        string
	CreatedAt time.Time
	UpdatedAt time.Time
	Name      string
	Source    string
	Params    map[string]string
	IsLocal   bool
}

// PostModel represents a single cached or fetched post.
type PostModel struct {
	ID        string
	CreatedAt time.Time
	UpdatedAt time.Time
	URI       string
	AuthorDID string
	Text      string
	FeedID    string
	IndexedAt time.Time
}

// SessionModel represents a user session and API context.
type SessionModel struct {
	ID         string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	Handle     string
	Token      string
	ServiceURL string
	IsValid    bool
}
