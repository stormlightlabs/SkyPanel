package store

import (
	"testing"
)

// TestBlueskyServiceImplementsServiceInterface verifies that BlueskyService
// implements the Service interface correctly.
func TestBlueskyServiceImplementsServiceInterface(t *testing.T) {
	var _ Service = (*BlueskyService)(nil)
}

// TestServiceIdentifierTypes verifies the type aliases are correctly defined.
func TestServiceIdentifierTypes(t *testing.T) {
	var id ServiceID = 1
	if id != 1 {
		t.Errorf("ServiceID type alias broken")
	}

	var identifier ServiceIdentifier = "Bluesky"
	if identifier != "Bluesky" {
		t.Errorf("ServiceIdentifier type alias broken")
	}
}

// TestCreateSessionResponse verifies the response struct can be instantiated.
func TestCreateSessionResponse(t *testing.T) {
	response := CreateSessionResponse{
		Did:    "did:plc:test123",
		Handle: "test.bsky.social",
		Active: true,
		// AccessJwt:  "access-token",
		// RefreshJwt: "refresh-token",
	}

	if response.Did != "did:plc:test123" {
		t.Errorf("expected Did 'did:plc:test123', got %s", response.Did)
	}
	if response.Handle != "test.bsky.social" {
		t.Errorf("expected Handle 'test.bsky.social', got %s", response.Handle)
	}
	if !response.Active {
		t.Error("expected Active to be true")
	}
}

// TestActorProfile verifies the ActorProfile struct.
func TestActorProfile(t *testing.T) {
	profile := ActorProfile{
		Did:            "did:plc:test",
		FollowersCount: 100,
		// Handle:         "test.bsky.social",
		// DisplayName:    "Test User",
		// FollowsCount:   50,
		// PostsCount:     25,
	}

	if profile.Did != "did:plc:test" {
		t.Errorf("expected Did 'did:plc:test', got %s", profile.Did)
	}
	if profile.FollowersCount != 100 {
		t.Errorf("expected FollowersCount 100, got %d", profile.FollowersCount)
	}
}

// TestPostView verifies the PostView struct.
func TestPostView(t *testing.T) {
	post := PostView{
		Uri:       "at://did:plc:test/app.bsky.feed.post/123",
		LikeCount: 25,
		// Cid:         "bafyrei123",
		// ReplyCount:  5,
		// RepostCount: 10,
		// QuoteCount:  2,
	}

	if post.Uri != "at://did:plc:test/app.bsky.feed.post/123" {
		t.Errorf("unexpected Uri: %s", post.Uri)
	}
	if post.LikeCount != 25 {
		t.Errorf("expected LikeCount 25, got %d", post.LikeCount)
	}
}

// TestGetTimelineResponse verifies timeline response structure.
func TestGetTimelineResponse(t *testing.T) {
	response := GetTimelineResponse{
		Cursor: "next-cursor-123",
		Feed: []FeedViewPost{
			{
				Post: &PostView{
					Uri: "at://test/post1",
				},
			},
		},
	}

	if response.Cursor != "next-cursor-123" {
		t.Errorf("unexpected Cursor: %s", response.Cursor)
	}
	if len(response.Feed) != 1 {
		t.Errorf("expected 1 post, got %d", len(response.Feed))
	}
}

// TestViewerState verifies viewer state structure.
func TestViewerState(t *testing.T) {
	state := ViewerState{
		Muted:      false,
		Following:  "at://did:plc:me/app.bsky.graph.follow/abc",
		Bookmarked: true,
		// BlockedBy:  false,
		// Like:       "at://did:plc:me/app.bsky.feed.like/xyz",
	}

	if state.Muted {
		t.Error("expected Muted to be false")
	}
	if !state.Bookmarked {
		t.Error("expected Bookmarked to be true")
	}
	if state.Following == "" {
		t.Error("expected Following to be set")
	}
}

// TestLabel verifies label structure.
func TestLabel(t *testing.T) {
	label := Label{
		Src: "did:plc:moderator",
		Val: "nsfw",
		// Uri: "at://test/post1",
		// Cts: "2024-01-01T00:00:00Z",
	}

	if label.Val != "nsfw" {
		t.Errorf("expected Val 'nsfw', got %s", label.Val)
	}
	if label.Src != "did:plc:moderator" {
		t.Errorf("expected Src 'did:plc:moderator', got %s", label.Src)
	}
}
