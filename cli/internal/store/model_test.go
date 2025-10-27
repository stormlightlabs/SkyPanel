package store

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestGenerateUUID(t *testing.T) {
	t.Run("generates valid UUID", func(t *testing.T) {
		id := GenerateUUID()
		if id == "" {
			t.Error("GenerateUUID returned empty string")
		}

		if _, err := uuid.Parse(id); err != nil {
			t.Errorf("GenerateUUID returned invalid UUID: %v", err)
		}
	})

	t.Run("generates unique UUIDs", func(t *testing.T) {
		id1 := GenerateUUID()
		id2 := GenerateUUID()

		if id1 == id2 {
			t.Error("GenerateUUID generated duplicate UUIDs")
		}
	})

	t.Run("generates multiple unique UUIDs", func(t *testing.T) {
		seen := make(map[string]bool)
		count := 1000

		for range count {
			id := GenerateUUID()
			if seen[id] {
				t.Errorf("Duplicate UUID generated: %s", id)
			}
			seen[id] = true
		}

		if len(seen) != count {
			t.Errorf("Expected %d unique UUIDs, got %d", count, len(seen))
		}
	})
}

func TestFeedModel_ImplementsModelInterface(t *testing.T) {
	var _ Model = (*FeedModel)(nil)
}

func TestFeedModel_Getters(t *testing.T) {
	now := time.Now()
	feed := &FeedModel{
		id:        "test-id",
		createdAt: now,
		updatedAt: now.Add(time.Hour),
		Name:      "Test Feed",
		Source:    "test-source",
		Params:    map[string]string{"key": "value"},
		IsLocal:   true,
	}

	if got := feed.ID(); got != "test-id" {
		t.Errorf("ID() = %v, want %v", got, "test-id")
	}

	if got := feed.CreatedAt(); !got.Equal(now) {
		t.Errorf("CreatedAt() = %v, want %v", got, now)
	}

	if got := feed.UpdatedAt(); !got.Equal(now.Add(time.Hour)) {
		t.Errorf("UpdatedAt() = %v, want %v", got, now.Add(time.Hour))
	}

	if feed.Name != "Test Feed" {
		t.Errorf("Name = %v, want %v", feed.Name, "Test Feed")
	}

	if feed.Source != "test-source" {
		t.Errorf("Source = %v, want %v", feed.Source, "test-source")
	}

	if feed.Params["key"] != "value" {
		t.Errorf("Params[key] = %v, want %v", feed.Params["key"], "value")
	}

	if !feed.IsLocal {
		t.Error("IsLocal = false, want true")
	}
}

func TestFeedModel_Setters(t *testing.T) {
	feed := &FeedModel{}
	now := time.Now()

	feed.SetID("new-id")
	if got := feed.ID(); got != "new-id" {
		t.Errorf("After SetID, ID() = %v, want %v", got, "new-id")
	}

	feed.SetCreatedAt(now)
	if got := feed.CreatedAt(); !got.Equal(now) {
		t.Errorf("After SetCreatedAt, CreatedAt() = %v, want %v", got, now)
	}

	later := now.Add(time.Hour)
	feed.SetUpdatedAt(later)
	if got := feed.UpdatedAt(); !got.Equal(later) {
		t.Errorf("After SetUpdatedAt, UpdatedAt() = %v, want %v", got, later)
	}
}

func TestFeedModel_TouchUpdatedAt(t *testing.T) {
	feed := &FeedModel{}
	initial := time.Now().Add(-time.Hour)
	feed.SetUpdatedAt(initial)

	time.Sleep(10 * time.Millisecond)

	feed.TouchUpdatedAt()
	updated := feed.UpdatedAt()

	if !updated.After(initial) {
		t.Errorf("TouchUpdatedAt did not update timestamp: initial=%v, updated=%v", initial, updated)
	}

	if time.Since(updated) > time.Second {
		t.Errorf("TouchUpdatedAt timestamp is not recent: %v", updated)
	}
}

func TestFeedModel_ParamsMutability(t *testing.T) {
	feed := &FeedModel{
		Params: make(map[string]string),
	}

	feed.Params["key1"] = "value1"
	feed.Params["key2"] = "value2"

	if len(feed.Params) != 2 {
		t.Errorf("Expected 2 params, got %d", len(feed.Params))
	}

	if feed.Params["key1"] != "value1" {
		t.Errorf("Params[key1] = %v, want value1", feed.Params["key1"])
	}
}

func TestPostModel_ImplementsModelInterface(t *testing.T) {
	var _ Model = (*PostModel)(nil)
}

func TestPostModel_Getters(t *testing.T) {
	now := time.Now()
	indexed := now.Add(-time.Hour)
	post := &PostModel{
		id:        "post-id",
		createdAt: now,
		updatedAt: now.Add(time.Minute),
		URI:       "at://did:plc:test/app.bsky.feed.post/abc123",
		AuthorDID: "did:plc:test",
		Text:      "Hello world",
		FeedID:    "feed-id",
		IndexedAt: indexed,
	}

	if got := post.ID(); got != "post-id" {
		t.Errorf("ID() = %v, want %v", got, "post-id")
	}

	if got := post.CreatedAt(); !got.Equal(now) {
		t.Errorf("CreatedAt() = %v, want %v", got, now)
	}

	if got := post.UpdatedAt(); !got.Equal(now.Add(time.Minute)) {
		t.Errorf("UpdatedAt() = %v, want %v", got, now.Add(time.Minute))
	}

	if post.URI != "at://did:plc:test/app.bsky.feed.post/abc123" {
		t.Errorf("URI = %v, want at://...", post.URI)
	}

	if post.AuthorDID != "did:plc:test" {
		t.Errorf("AuthorDID = %v, want did:plc:test", post.AuthorDID)
	}

	if post.Text != "Hello world" {
		t.Errorf("Text = %v, want Hello world", post.Text)
	}

	if post.FeedID != "feed-id" {
		t.Errorf("FeedID = %v, want feed-id", post.FeedID)
	}

	if !post.IndexedAt.Equal(indexed) {
		t.Errorf("IndexedAt = %v, want %v", post.IndexedAt, indexed)
	}
}

func TestPostModel_Setters(t *testing.T) {
	post := &PostModel{}
	now := time.Now()

	post.SetID("new-post-id")
	if got := post.ID(); got != "new-post-id" {
		t.Errorf("After SetID, ID() = %v, want %v", got, "new-post-id")
	}

	post.SetCreatedAt(now)
	if got := post.CreatedAt(); !got.Equal(now) {
		t.Errorf("After SetCreatedAt, CreatedAt() = %v, want %v", got, now)
	}

	later := now.Add(time.Hour)
	post.SetUpdatedAt(later)
	if got := post.UpdatedAt(); !got.Equal(later) {
		t.Errorf("After SetUpdatedAt, UpdatedAt() = %v, want %v", got, later)
	}
}

func TestPostModel_TouchUpdatedAt(t *testing.T) {
	post := &PostModel{}
	initial := time.Now().Add(-time.Hour)
	post.SetUpdatedAt(initial)

	time.Sleep(10 * time.Millisecond)

	post.TouchUpdatedAt()
	updated := post.UpdatedAt()

	if !updated.After(initial) {
		t.Errorf("TouchUpdatedAt did not update timestamp: initial=%v, updated=%v", initial, updated)
	}

	if time.Since(updated) > time.Second {
		t.Errorf("TouchUpdatedAt timestamp is not recent: %v", updated)
	}
}

func TestPostModel_EmptyText(t *testing.T) {
	post := &PostModel{
		Text: "",
	}

	if post.Text != "" {
		t.Errorf("Expected empty text, got %v", post.Text)
	}
}

func TestSessionModel_ImplementsModelInterface(t *testing.T) {
	var _ Model = (*SessionModel)(nil)
}

func TestSessionModel_Getters(t *testing.T) {
	now := time.Now()
	session := &SessionModel{
		id:         "session-id",
		createdAt:  now,
		updatedAt:  now.Add(time.Minute),
		Handle:     "@user.bsky.social",
		Token:      "encrypted-token",
		ServiceURL: "https://bsky.social",
		IsValid:    true,
	}

	if got := session.ID(); got != "session-id" {
		t.Errorf("ID() = %v, want %v", got, "session-id")
	}

	if got := session.CreatedAt(); !got.Equal(now) {
		t.Errorf("CreatedAt() = %v, want %v", got, now)
	}

	if got := session.UpdatedAt(); !got.Equal(now.Add(time.Minute)) {
		t.Errorf("UpdatedAt() = %v, want %v", got, now.Add(time.Minute))
	}

	if session.Handle != "@user.bsky.social" {
		t.Errorf("Handle = %v, want @user.bsky.social", session.Handle)
	}

	if session.Token != "encrypted-token" {
		t.Errorf("Token = %v, want encrypted-token", session.Token)
	}

	if session.ServiceURL != "https://bsky.social" {
		t.Errorf("ServiceURL = %v, want https://bsky.social", session.ServiceURL)
	}

	if !session.IsValid {
		t.Error("IsValid = false, want true")
	}
}

func TestSessionModel_Setters(t *testing.T) {
	session := &SessionModel{}
	now := time.Now()

	session.SetID("new-session-id")
	if got := session.ID(); got != "new-session-id" {
		t.Errorf("After SetID, ID() = %v, want %v", got, "new-session-id")
	}

	session.SetCreatedAt(now)
	if got := session.CreatedAt(); !got.Equal(now) {
		t.Errorf("After SetCreatedAt, CreatedAt() = %v, want %v", got, now)
	}

	later := now.Add(time.Hour)
	session.SetUpdatedAt(later)
	if got := session.UpdatedAt(); !got.Equal(later) {
		t.Errorf("After SetUpdatedAt, UpdatedAt() = %v, want %v", got, later)
	}
}

func TestSessionModel_TouchUpdatedAt(t *testing.T) {
	session := &SessionModel{}
	initial := time.Now().Add(-time.Hour)
	session.SetUpdatedAt(initial)

	time.Sleep(10 * time.Millisecond)

	session.TouchUpdatedAt()
	updated := session.UpdatedAt()

	if !updated.After(initial) {
		t.Errorf("TouchUpdatedAt did not update timestamp: initial=%v, updated=%v", initial, updated)
	}

	if time.Since(updated) > time.Second {
		t.Errorf("TouchUpdatedAt timestamp is not recent: %v", updated)
	}
}

func TestSessionModel_IsValidFlag(t *testing.T) {
	t.Run("valid session", func(t *testing.T) {
		session := &SessionModel{IsValid: true}
		if !session.IsValid {
			t.Error("Expected IsValid to be true")
		}
	})

	t.Run("invalid session", func(t *testing.T) {
		session := &SessionModel{IsValid: false}
		if session.IsValid {
			t.Error("Expected IsValid to be false")
		}
	})

	t.Run("toggle validity", func(t *testing.T) {
		session := &SessionModel{IsValid: true}
		session.IsValid = false
		if session.IsValid {
			t.Error("Expected IsValid to be false after toggle")
		}
	})
}

func TestModel_ZeroValues(t *testing.T) {
	t.Run("FeedModel zero values", func(t *testing.T) {
		feed := &FeedModel{}
		if feed.ID() != "" {
			t.Errorf("Expected empty ID, got %v", feed.ID())
		}
		if !feed.CreatedAt().IsZero() {
			t.Errorf("Expected zero CreatedAt, got %v", feed.CreatedAt())
		}
		if !feed.UpdatedAt().IsZero() {
			t.Errorf("Expected zero UpdatedAt, got %v", feed.UpdatedAt())
		}
	})

	t.Run("PostModel zero values", func(t *testing.T) {
		post := &PostModel{}
		if post.ID() != "" {
			t.Errorf("Expected empty ID, got %v", post.ID())
		}
		if !post.CreatedAt().IsZero() {
			t.Errorf("Expected zero CreatedAt, got %v", post.CreatedAt())
		}
		if !post.UpdatedAt().IsZero() {
			t.Errorf("Expected zero UpdatedAt, got %v", post.UpdatedAt())
		}
	})

	t.Run("SessionModel zero values", func(t *testing.T) {
		session := &SessionModel{}
		if session.ID() != "" {
			t.Errorf("Expected empty ID, got %v", session.ID())
		}
		if !session.CreatedAt().IsZero() {
			t.Errorf("Expected zero CreatedAt, got %v", session.CreatedAt())
		}
		if !session.UpdatedAt().IsZero() {
			t.Errorf("Expected zero UpdatedAt, got %v", session.UpdatedAt())
		}
	})
}

func TestModel_AsModelInterface(t *testing.T) {
	now := time.Now()

	testCases := []struct {
		name  string
		model Model
		id    string
	}{
		{
			name: "FeedModel as Model",
			model: &FeedModel{
				id:        "feed-1",
				createdAt: now,
				updatedAt: now,
			},
			id: "feed-1",
		},
		{
			name: "PostModel as Model",
			model: &PostModel{
				id:        "post-1",
				createdAt: now,
				updatedAt: now,
			},
			id: "post-1",
		},
		{
			name: "SessionModel as Model",
			model: &SessionModel{
				id:        "session-1",
				createdAt: now,
				updatedAt: now,
			},
			id: "session-1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.model.ID(); got != tc.id {
				t.Errorf("ID() = %v, want %v", got, tc.id)
			}
			if got := tc.model.CreatedAt(); !got.Equal(now) {
				t.Errorf("CreatedAt() = %v, want %v", got, now)
			}
			if got := tc.model.UpdatedAt(); !got.Equal(now) {
				t.Errorf("UpdatedAt() = %v, want %v", got, now)
			}
		})
	}
}
