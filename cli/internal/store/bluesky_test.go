package store

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestNewBlueskyService verifies service initialization with default and custom URLs.
func TestNewBlueskyService(t *testing.T) {
	t.Run("default URL", func(t *testing.T) {
		svc := NewBlueskyService("")
		if svc.BaseURL() != defaultServiceURL {
			t.Errorf("expected default URL %s, got %s", defaultServiceURL, svc.BaseURL())
		}
		if svc.Authenticated() {
			t.Error("new service should not be authenticated")
		}
	})

	t.Run("custom URL", func(t *testing.T) {
		customURL := "https://custom.bsky.social"
		svc := NewBlueskyService(customURL)
		if svc.BaseURL() != customURL {
			t.Errorf("expected custom URL %s, got %s", customURL, svc.BaseURL())
		}
	})
}

// TestBlueskyService_Name verifies the service identifier.
func TestBlueskyService_Name(t *testing.T) {
	svc := NewBlueskyService("")
	if svc.Name() != "Bluesky" {
		t.Errorf("expected name 'Bluesky', got %s", svc.Name())
	}
}

// TestBlueskyService_Authenticate verifies successful authentication flow.
func TestBlueskyService_Authenticate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/xrpc/com.atproto.server.createSession" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}

		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if body["identifier"] != "test.bsky.social" {
			t.Errorf("unexpected identifier: %s", body["identifier"])
		}
		if body["password"] != "test-password" {
			t.Errorf("unexpected password: %s", body["password"])
		}

		response := CreateSessionResponse{
			Did:        "did:plc:test123",
			Handle:     "test.bsky.social",
			AccessJwt:  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MzAwMDAwMDB9.test",
			RefreshJwt: "refresh-token",
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	svc := NewBlueskyService(server.URL)
	creds := map[string]string{
		"identifier": "test.bsky.social",
		"password":   "test-password",
	}

	err := svc.Authenticate(context.Background(), creds)
	if err != nil {
		t.Fatalf("authentication failed: %v", err)
	}

	if !svc.Authenticated() {
		t.Error("service should be authenticated")
	}
	if svc.GetAccessToken() == "" {
		t.Error("access token should be set")
	}
	if svc.GetRefreshToken() == "" {
		t.Error("refresh token should be set")
	}
}

// TestBlueskyService_Authenticate_InvalidCredentials verifies error handling for invalid credentials.
func TestBlueskyService_Authenticate_InvalidCredentials(t *testing.T) {
	tests := []struct {
		name        string
		credentials any
		wantErr     string
	}{
		{
			name:        "wrong type",
			credentials: "not-a-map",
			wantErr:     "credentials must be map[string]string",
		},
		{
			name:        "missing identifier",
			credentials: map[string]string{"password": "test"},
			wantErr:     "identifier required",
		},
		{
			name:        "missing password",
			credentials: map[string]string{"identifier": "test"},
			wantErr:     "password required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewBlueskyService("")
			err := svc.Authenticate(context.Background(), tt.credentials)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("expected error containing %q, got %q", tt.wantErr, err.Error())
			}
		})
	}
}

// TestBlueskyService_Authenticate_ServerError verifies error handling for server errors.
func TestBlueskyService_Authenticate_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"InvalidCredentials","message":"Invalid handle or password"}`))
	}))
	defer server.Close()

	svc := NewBlueskyService(server.URL)
	creds := map[string]string{
		"identifier": "test.bsky.social",
		"password":   "wrong-password",
	}

	err := svc.Authenticate(context.Background(), creds)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "401") {
		t.Errorf("expected 401 error, got: %v", err)
	}
}

// TestBlueskyService_Request verifies generic request handling.
func TestBlueskyService_Request(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			t.Errorf("missing or invalid authorization header: %s", auth)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success":true}`))
	}))
	defer server.Close()

	svc := NewBlueskyService(server.URL)
	svc.SetTokens("test-access-token", "test-refresh-token")

	resp, err := svc.Request(context.Background(), "GET", "/test", nil, nil)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

// TestBlueskyService_Request_Unauthenticated verifies error when not authenticated.
func TestBlueskyService_Request_Unauthenticated(t *testing.T) {
	svc := NewBlueskyService("")
	_, err := svc.Request(context.Background(), "GET", "/test", nil, nil)
	if err == nil {
		t.Fatal("expected error for unauthenticated request")
	}
	if !strings.Contains(err.Error(), "not authenticated") {
		t.Errorf("expected authentication error, got: %v", err)
	}
}

// TestBlueskyService_Request_TokenRefresh verifies automatic token refresh on 401.
func TestBlueskyService_Request_TokenRefresh(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/xrpc/com.atproto.server.refreshSession":
			response := CreateSessionResponse{
				AccessJwt:  "new-access-token",
				RefreshJwt: "new-refresh-token",
			}
			json.NewEncoder(w).Encode(response)
		case "/test":
			callCount++
			if callCount == 1 {
				w.WriteHeader(http.StatusUnauthorized)
			} else {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"success":true}`))
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	svc := NewBlueskyService(server.URL)
	svc.SetTokens("expired-token", "refresh-token")

	resp, err := svc.Request(context.Background(), "GET", "/test", nil, nil)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 after refresh, got %d", resp.StatusCode)
	}
	if callCount != 2 {
		t.Errorf("expected 2 calls (failed + retry), got %d", callCount)
	}
	if svc.GetAccessToken() != "new-access-token" {
		t.Error("access token should be updated after refresh")
	}
}

// TestBlueskyService_HealthCheck verifies connectivity checks.
func TestBlueskyService_HealthCheck(t *testing.T) {
	t.Run("healthy", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/xrpc/_health" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		svc := NewBlueskyService(server.URL)
		if err := svc.HealthCheck(context.Background()); err != nil {
			t.Errorf("health check failed: %v", err)
		}
	})

	t.Run("unhealthy", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusServiceUnavailable)
		}))
		defer server.Close()

		svc := NewBlueskyService(server.URL)
		err := svc.HealthCheck(context.Background())
		if err == nil {
			t.Error("expected health check to fail")
		}
	})
}

// TestBlueskyService_GetTimeline verifies timeline fetching.
func TestBlueskyService_GetTimeline(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "app.bsky.feed.getTimeline") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		limit := r.URL.Query().Get("limit")
		if limit != "50" {
			t.Errorf("expected limit=50, got %s", limit)
		}

		response := GetTimelineResponse{
			Cursor: "next-cursor",
			Feed: []FeedViewPost{
				{
					Post: &PostView{
						Uri: "at://test/post1",
						Author: &ActorProfile{
							Did:    "did:plc:test",
							Handle: "test.bsky.social",
						},
					},
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	svc := NewBlueskyService(server.URL)
	svc.SetTokens("test-token", "refresh-token")

	timeline, err := svc.GetTimeline(context.Background(), 50, "")
	if err != nil {
		t.Fatalf("GetTimeline failed: %v", err)
	}

	if timeline.Cursor != "next-cursor" {
		t.Errorf("expected cursor 'next-cursor', got %s", timeline.Cursor)
	}
	if len(timeline.Feed) != 1 {
		t.Errorf("expected 1 post, got %d", len(timeline.Feed))
	}
}

// TestBlueskyService_GetAuthorFeed verifies author feed fetching.
func TestBlueskyService_GetAuthorFeed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "app.bsky.feed.getAuthorFeed") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		actor := r.URL.Query().Get("actor")
		if actor != "test.bsky.social" {
			t.Errorf("expected actor=test.bsky.social, got %s", actor)
		}

		response := GetAuthorFeedResponse{
			Feed: []FeedViewPost{
				{
					Post: &PostView{
						Uri: "at://test/post1",
					},
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	svc := NewBlueskyService(server.URL)
	svc.SetTokens("test-token", "refresh-token")

	feed, err := svc.GetAuthorFeed(context.Background(), "test.bsky.social", 50, "")
	if err != nil {
		t.Fatalf("GetAuthorFeed failed: %v", err)
	}

	if len(feed.Feed) != 1 {
		t.Errorf("expected 1 post, got %d", len(feed.Feed))
	}
}

// TestBlueskyService_GetFollows verifies follows fetching.
func TestBlueskyService_GetFollows(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "app.bsky.graph.getFollows") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		cursor := r.URL.Query().Get("cursor")
		if cursor != "test-cursor" {
			t.Errorf("expected cursor=test-cursor, got %s", cursor)
		}

		response := GetFollowsResponse{
			Subject: "did:plc:test",
			Cursor:  "next-cursor",
			Follows: []ActorProfile{
				{
					Did:    "did:plc:follow1",
					Handle: "follow1.bsky.social",
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	svc := NewBlueskyService(server.URL)
	svc.SetTokens("test-token", "refresh-token")

	follows, err := svc.GetFollows(context.Background(), "test.bsky.social", 50, "test-cursor")
	if err != nil {
		t.Fatalf("GetFollows failed: %v", err)
	}

	if follows.Cursor != "next-cursor" {
		t.Errorf("expected cursor 'next-cursor', got %s", follows.Cursor)
	}
	if len(follows.Follows) != 1 {
		t.Errorf("expected 1 follow, got %d", len(follows.Follows))
	}
}

// TestBlueskyService_Close verifies cleanup.
func TestBlueskyService_Close(t *testing.T) {
	svc := NewBlueskyService("")
	svc.SetTokens("test-token", "refresh-token")

	if err := svc.Close(); err != nil {
		t.Errorf("Close failed: %v", err)
	}

	if svc.Authenticated() {
		t.Error("service should not be authenticated after close")
	}
	if svc.GetAccessToken() != "" {
		t.Error("access token should be cleared after close")
	}
}

// TestBlueskyService_SetTokens verifies token setting.
func TestBlueskyService_SetTokens(t *testing.T) {
	svc := NewBlueskyService("")
	svc.SetTokens("access-token", "refresh-token")

	if !svc.Authenticated() {
		t.Error("service should be authenticated after setting tokens")
	}
	if svc.GetAccessToken() != "access-token" {
		t.Errorf("expected access token 'access-token', got %s", svc.GetAccessToken())
	}
	if svc.GetRefreshToken() != "refresh-token" {
		t.Errorf("expected refresh token 'refresh-token', got %s", svc.GetRefreshToken())
	}
}

// TestParseJWTExpiry verifies JWT expiry parsing.
func TestParseJWTExpiry(t *testing.T) {
	t.Run("valid JWT", func(t *testing.T) {
		token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MzAwMDAwMDB9.dummysignature"
		expiry, err := parseJWTExpiry(token)
		if err != nil {
			t.Fatalf("parseJWTExpiry failed: %v", err)
		}

		expected := time.Unix(1730000000, 0)
		if !expiry.Equal(expected) {
			t.Errorf("expected expiry %v, got %v", expected, expiry)
		}
	})

	t.Run("invalid format", func(t *testing.T) {
		_, err := parseJWTExpiry("not.a.jwt")
		if err == nil {
			t.Error("expected error for invalid JWT")
		}
	})

	t.Run("missing exp claim", func(t *testing.T) {
		token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dummysignature"
		_, err := parseJWTExpiry(token)
		if err == nil {
			t.Error("expected error for JWT without exp claim")
		}
	})
}

// TestBlueskyService_ShouldRefreshToken verifies token refresh logic.
func TestBlueskyService_ShouldRefreshToken(t *testing.T) {
	t.Run("no expiry set", func(t *testing.T) {
		svc := NewBlueskyService("")
		if svc.shouldRefreshToken() {
			t.Error("should not refresh when expiry is zero")
		}
	})

	t.Run("token expired", func(t *testing.T) {
		svc := NewBlueskyService("")
		svc.tokenExpiry = time.Now().Add(-1 * time.Second)
		if !svc.shouldRefreshToken() {
			t.Error("should refresh token when expired")
		}
	})

	t.Run("token very close to expiry", func(t *testing.T) {
		svc := NewBlueskyService("")
		svc.tokenExpiry = time.Now().Add(100 * time.Millisecond)
		if svc.shouldRefreshToken() {
			t.Error("should not refresh with 100ms remaining")
		}
	})

	t.Run("token not near expiry", func(t *testing.T) {
		svc := NewBlueskyService("")
		svc.tokenExpiry = time.Now().Add(1 * time.Hour)
		if svc.shouldRefreshToken() {
			t.Error("should not refresh token when far from expiry")
		}
	})
}

// TestBlueskyService_RefreshAccessToken verifies token refresh flow.
func TestBlueskyService_RefreshAccessToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/xrpc/com.atproto.server.refreshSession" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		auth := r.Header.Get("Authorization")
		if !strings.Contains(auth, "old-refresh-token") {
			t.Errorf("expected refresh token in auth header, got: %s", auth)
		}

		response := CreateSessionResponse{
			AccessJwt:  "new-access-token",
			RefreshJwt: "new-refresh-token",
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	svc := NewBlueskyService(server.URL)
	svc.accessToken = "old-access-token"
	svc.refreshToken = "old-refresh-token"
	svc.authenticated = true

	err := svc.refreshAccessToken(context.Background())
	if err != nil {
		t.Fatalf("refreshAccessToken failed: %v", err)
	}

	if svc.GetAccessToken() != "new-access-token" {
		t.Errorf("expected new access token, got %s", svc.GetAccessToken())
	}
	if svc.GetRefreshToken() != "new-refresh-token" {
		t.Errorf("expected new refresh token, got %s", svc.GetRefreshToken())
	}
}

// TestBlueskyService_APIErrors verifies error handling for various API failures.
func TestBlueskyService_APIErrors(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		method     func(*BlueskyService) error
	}{
		{
			name:       "GetTimeline 500",
			statusCode: http.StatusInternalServerError,
			method: func(svc *BlueskyService) error {
				_, err := svc.GetTimeline(context.Background(), 50, "")
				return err
			},
		},
		{
			name:       "GetAuthorFeed 404",
			statusCode: http.StatusNotFound,
			method: func(svc *BlueskyService) error {
				_, err := svc.GetAuthorFeed(context.Background(), "nonexistent", 50, "")
				return err
			},
		},
		{
			name:       "GetFollows 403",
			statusCode: http.StatusForbidden,
			method: func(svc *BlueskyService) error {
				_, err := svc.GetFollows(context.Background(), "blocked", 50, "")
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(`{"error":"TestError","message":"Test error message"}`))
			}))
			defer server.Close()

			svc := NewBlueskyService(server.URL)
			svc.SetTokens("test-token", "refresh-token")

			err := tt.method(svc)
			if err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}
