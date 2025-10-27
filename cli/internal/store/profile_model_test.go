package store

import (
	"testing"
	"time"
)

func TestProfileModel_ImplementsModelInterface(t *testing.T) {
	var _ Model = (*ProfileModel)(nil)
}

func TestProfileModel_Getters(t *testing.T) {
	now := time.Now()
	fetched := now.Add(-30 * time.Minute)
	profile := &ProfileModel{
		id:        "profile-id",
		createdAt: now,
		updatedAt: now.Add(time.Minute),
		Did:       "did:plc:test123",
		Handle:    "alice.bsky.social",
		DataJSON:  `{"did":"did:plc:test123","handle":"alice.bsky.social"}`,
		FetchedAt: fetched,
	}

	if got := profile.ID(); got != "profile-id" {
		t.Errorf("ID() = %v, want %v", got, "profile-id")
	}

	if got := profile.CreatedAt(); !got.Equal(now) {
		t.Errorf("CreatedAt() = %v, want %v", got, now)
	}

	if got := profile.UpdatedAt(); !got.Equal(now.Add(time.Minute)) {
		t.Errorf("UpdatedAt() = %v, want %v", got, now.Add(time.Minute))
	}

	if profile.Did != "did:plc:test123" {
		t.Errorf("Did = %v, want did:plc:test123", profile.Did)
	}

	if profile.Handle != "alice.bsky.social" {
		t.Errorf("Handle = %v, want alice.bsky.social", profile.Handle)
	}

	if !profile.FetchedAt.Equal(fetched) {
		t.Errorf("FetchedAt = %v, want %v", profile.FetchedAt, fetched)
	}

	if profile.DataJSON == "" {
		t.Error("DataJSON should not be empty")
	}
}

func TestProfileModel_Setters(t *testing.T) {
	profile := &ProfileModel{}
	now := time.Now()

	profile.SetID("new-profile-id")
	if got := profile.ID(); got != "new-profile-id" {
		t.Errorf("After SetID, ID() = %v, want %v", got, "new-profile-id")
	}

	profile.SetCreatedAt(now)
	if got := profile.CreatedAt(); !got.Equal(now) {
		t.Errorf("After SetCreatedAt, CreatedAt() = %v, want %v", got, now)
	}

	later := now.Add(time.Hour)
	profile.SetUpdatedAt(later)
	if got := profile.UpdatedAt(); !got.Equal(later) {
		t.Errorf("After SetUpdatedAt, UpdatedAt() = %v, want %v", got, later)
	}
}

func TestProfileModel_TouchUpdatedAt(t *testing.T) {
	profile := &ProfileModel{}
	initial := time.Now().Add(-time.Hour)
	profile.SetUpdatedAt(initial)

	time.Sleep(10 * time.Millisecond)

	profile.TouchUpdatedAt()
	updated := profile.UpdatedAt()

	if !updated.After(initial) {
		t.Errorf("TouchUpdatedAt did not update timestamp: initial=%v, updated=%v", initial, updated)
	}

	if time.Since(updated) > time.Second {
		t.Errorf("TouchUpdatedAt timestamp is not recent: %v", updated)
	}
}

func TestProfileModel_IsFresh_DefaultTTL(t *testing.T) {
	t.Run("fresh profile within 1 hour", func(t *testing.T) {
		profile := &ProfileModel{
			FetchedAt: time.Now().Add(-30 * time.Minute),
		}

		if !profile.IsFresh(0) {
			t.Error("profile should be fresh within 1 hour (default TTL)")
		}
	})

	t.Run("stale profile after 1 hour", func(t *testing.T) {
		profile := &ProfileModel{
			FetchedAt: time.Now().Add(-90 * time.Minute),
		}

		if profile.IsFresh(0) {
			t.Error("profile should be stale after 1 hour (default TTL)")
		}
	})

	t.Run("profile at exactly 1 hour boundary", func(t *testing.T) {
		profile := &ProfileModel{
			FetchedAt: time.Now().Add(-1 * time.Hour),
		}

		if profile.IsFresh(0) {
			t.Error("profile should be stale at exactly 1 hour")
		}
	})
}

func TestProfileModel_IsFresh_CustomTTL(t *testing.T) {
	t.Run("fresh with custom 5 minute TTL", func(t *testing.T) {
		profile := &ProfileModel{
			FetchedAt: time.Now().Add(-3 * time.Minute),
		}

		if !profile.IsFresh(5 * time.Minute) {
			t.Error("profile should be fresh within 5 minute TTL")
		}
	})

	t.Run("stale with custom 5 minute TTL", func(t *testing.T) {
		profile := &ProfileModel{
			FetchedAt: time.Now().Add(-10 * time.Minute),
		}

		if profile.IsFresh(5 * time.Minute) {
			t.Error("profile should be stale after 5 minute TTL")
		}
	})

	t.Run("fresh with custom 24 hour TTL", func(t *testing.T) {
		profile := &ProfileModel{
			FetchedAt: time.Now().Add(-12 * time.Hour),
		}

		if !profile.IsFresh(24 * time.Hour) {
			t.Error("profile should be fresh within 24 hour TTL")
		}
	})
}

func TestProfileModel_IsFresh_ZeroFetchedAt(t *testing.T) {
	profile := &ProfileModel{FetchedAt: time.Time{}}
	if profile.IsFresh(time.Hour) {
		t.Error("profile with zero FetchedAt should be stale")
	}
}

func TestProfileModel_ZeroValues(t *testing.T) {
	profile := &ProfileModel{}

	if profile.ID() != "" {
		t.Errorf("Expected empty ID, got %v", profile.ID())
	}
	if !profile.CreatedAt().IsZero() {
		t.Errorf("Expected zero CreatedAt, got %v", profile.CreatedAt())
	}
	if !profile.UpdatedAt().IsZero() {
		t.Errorf("Expected zero UpdatedAt, got %v", profile.UpdatedAt())
	}
	if !profile.FetchedAt.IsZero() {
		t.Errorf("Expected zero FetchedAt, got %v", profile.FetchedAt)
	}
	if profile.Did != "" {
		t.Errorf("Expected empty Did, got %v", profile.Did)
	}
	if profile.Handle != "" {
		t.Errorf("Expected empty Handle, got %v", profile.Handle)
	}
	if profile.DataJSON != "" {
		t.Errorf("Expected empty DataJSON, got %v", profile.DataJSON)
	}
}

func TestProfileModel_DataJSONStorage(t *testing.T) {
	jsonData := `{
		"did": "did:plc:abc123",
		"handle": "bob.bsky.social",
		"displayName": "Bob Smith",
		"followersCount": 100,
		"followsCount": 50
	}`

	profile := &ProfileModel{
		Did:      "did:plc:abc123",
		Handle:   "bob.bsky.social",
		DataJSON: jsonData,
	}

	if profile.DataJSON != jsonData {
		t.Error("DataJSON was not stored correctly")
	}
}
