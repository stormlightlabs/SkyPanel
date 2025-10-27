package store

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Repository defines a generic persistence contract.
type Repository interface {
	Init(ctx context.Context) error
	Close() error
	Get(ctx context.Context, id string) (Model, error)
	List(ctx context.Context) ([]Model, error)
	Save(ctx context.Context, feed Model) error
	Delete(ctx context.Context, id string) error
}

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
	id        string
	createdAt time.Time
	updatedAt time.Time
	Name      string
	Source    string
	Params    map[string]string
	IsLocal   bool
}

func (m *FeedModel) ID() string               { return m.id }
func (m *FeedModel) CreatedAt() time.Time     { return m.createdAt }
func (m *FeedModel) UpdatedAt() time.Time     { return m.updatedAt }
func (m *FeedModel) SetID(id string)          { m.id = id }
func (m *FeedModel) SetCreatedAt(t time.Time) { m.createdAt = t }
func (m *FeedModel) SetUpdatedAt(t time.Time) { m.updatedAt = t }
func (m *FeedModel) TouchUpdatedAt()          { m.updatedAt = time.Now() }

// PostModel represents a single cached or fetched post.
type PostModel struct {
	id        string
	createdAt time.Time
	updatedAt time.Time
	URI       string
	AuthorDID string
	Text      string
	FeedID    string
	IndexedAt time.Time
}

func (m *PostModel) ID() string               { return m.id }
func (m *PostModel) CreatedAt() time.Time     { return m.createdAt }
func (m *PostModel) UpdatedAt() time.Time     { return m.updatedAt }
func (m *PostModel) SetID(id string)          { m.id = id }
func (m *PostModel) SetCreatedAt(t time.Time) { m.createdAt = t }
func (m *PostModel) SetUpdatedAt(t time.Time) { m.updatedAt = t }
func (m *PostModel) TouchUpdatedAt()          { m.updatedAt = time.Now() }

// SessionModel represents a user session and API context.
type SessionModel struct {
	id         string
	createdAt  time.Time
	updatedAt  time.Time
	Handle     string
	Token      string
	ServiceURL string
	IsValid    bool
}

func (m *SessionModel) ID() string               { return m.id }
func (m *SessionModel) CreatedAt() time.Time     { return m.createdAt }
func (m *SessionModel) UpdatedAt() time.Time     { return m.updatedAt }
func (m *SessionModel) SetID(id string)          { m.id = id }
func (m *SessionModel) SetCreatedAt(t time.Time) { m.createdAt = t }
func (m *SessionModel) SetUpdatedAt(t time.Time) { m.updatedAt = t }
func (m *SessionModel) TouchUpdatedAt()          { m.updatedAt = time.Now() }
