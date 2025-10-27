package store

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	defaultServiceURL = "https://bsky.social"
	defaultTimeout    = 30 * time.Second
)

// BlueskyService implements the Service interface for AT Protocol / Bluesky API
type BlueskyService struct {
	baseURL       string
	client        *http.Client
	accessToken   string
	refreshToken  string
	tokenExpiry   time.Time
	authenticated bool
}

// NewBlueskyService creates a new Bluesky service client
func NewBlueskyService(serviceURL string) *BlueskyService {
	if serviceURL == "" {
		serviceURL = defaultServiceURL
	}

	return &BlueskyService{
		baseURL: serviceURL,
		client: &http.Client{
			Timeout: defaultTimeout,
		},
		authenticated: false,
	}
}

// Name returns the service identifier
func (s *BlueskyService) Name() ServiceIdentifier {
	return "Bluesky"
}

// BaseURL returns the root endpoint of the remote API
func (s *BlueskyService) BaseURL() string {
	return s.baseURL
}

// Authenticated reports whether the client is currently authorized
func (s *BlueskyService) Authenticated() bool {
	return s.authenticated && s.accessToken != ""
}

// Authenticate establishes credentials with the service using handle and app password
// Expects credentials to be a map with "identifier" and "password" keys
func (s *BlueskyService) Authenticate(ctx context.Context, credentials any) error {
	creds, ok := credentials.(map[string]string)
	if !ok {
		return errors.New("credentials must be map[string]string with identifier and password")
	}

	identifier, ok := creds["identifier"]
	if !ok {
		return errors.New("identifier required in credentials")
	}

	password, ok := creds["password"]
	if !ok {
		return errors.New("password required in credentials")
	}

	body := map[string]string{
		"identifier": identifier,
		"password":   password,
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST",
		s.baseURL+"/xrpc/com.atproto.server.createSession",
		bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyText, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("authentication failed: %s - %s", resp.Status, string(bodyText))
	}

	var session CreateSessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&session); err != nil {
		return err
	}

	s.accessToken = session.AccessJwt
	s.refreshToken = session.RefreshJwt
	s.authenticated = true

	if expiry, err := parseJWTExpiry(s.accessToken); err == nil {
		s.tokenExpiry = expiry
	}

	return nil
}

// Request performs a generic API request with automatic token refresh
func (s *BlueskyService) Request(ctx context.Context, method, path string, body io.Reader, headers map[string]string) (*http.Response, error) {
	if !s.Authenticated() {
		return nil, errors.New("service not authenticated")
	}

	if s.shouldRefreshToken() {
		if err := s.refreshAccessToken(ctx); err != nil {
			return nil, fmt.Errorf("token refresh failed: %w", err)
		}
	}

	url := s.baseURL + path
	if !strings.HasPrefix(path, "/") {
		url = s.baseURL + "/" + path
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+s.accessToken)
	req.Header.Set("Content-Type", "application/json")

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()

		if err := s.refreshAccessToken(ctx); err != nil {
			return nil, fmt.Errorf("auth refresh failed after 401: %w", err)
		}

		req.Header.Set("Authorization", "Bearer "+s.accessToken)
		return s.client.Do(req)
	}

	return resp, nil
}

// HealthCheck verifies connectivity to the service
func (s *BlueskyService) HealthCheck(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", s.baseURL+"/xrpc/_health", nil)
	if err != nil {
		return err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed: %s", resp.Status)
	}

	return nil
}

// Close releases resources (no-op for HTTP client)
func (s *BlueskyService) Close() error {
	s.authenticated = false
	s.accessToken = ""
	s.refreshToken = ""
	return nil
}

// GetTimeline fetches the authenticated user's home timeline
func (s *BlueskyService) GetTimeline(ctx context.Context, limit int, cursor string) (*GetTimelineResponse, error) {
	url := fmt.Sprintf("/xrpc/app.bsky.feed.getTimeline?limit=%d", limit)
	if cursor != "" {
		url += "&cursor=" + cursor
	}

	resp, err := s.Request(ctx, "GET", url, nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyText, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("getTimeline failed: %s - %s", resp.Status, string(bodyText))
	}

	var timeline GetTimelineResponse
	if err := json.NewDecoder(resp.Body).Decode(&timeline); err != nil {
		return nil, err
	}

	return &timeline, nil
}

// GetAuthorFeed fetches posts by a specific author
func (s *BlueskyService) GetAuthorFeed(ctx context.Context, actor string, limit int, cursor string) (*GetAuthorFeedResponse, error) {
	url := fmt.Sprintf("/xrpc/app.bsky.feed.getAuthorFeed?actor=%s&limit=%d", actor, limit)
	if cursor != "" {
		url += "&cursor=" + cursor
	}

	resp, err := s.Request(ctx, "GET", url, nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyText, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("getAuthorFeed failed: %s - %s", resp.Status, string(bodyText))
	}

	var feed GetAuthorFeedResponse
	if err := json.NewDecoder(resp.Body).Decode(&feed); err != nil {
		return nil, err
	}

	return &feed, nil
}

// GetFollows fetches the list of accounts that an actor follows
func (s *BlueskyService) GetFollows(ctx context.Context, actor string, limit int, cursor string) (*GetFollowsResponse, error) {
	url := fmt.Sprintf("/xrpc/app.bsky.graph.getFollows?actor=%s&limit=%d", actor, limit)
	if cursor != "" {
		url += "&cursor=" + cursor
	}

	resp, err := s.Request(ctx, "GET", url, nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyText, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("getFollows failed: %s - %s", resp.Status, string(bodyText))
	}

	var follows GetFollowsResponse
	if err := json.NewDecoder(resp.Body).Decode(&follows); err != nil {
		return nil, err
	}

	return &follows, nil
}

// SetTokens allows external code to set tokens (e.g., from SessionRepository)
func (s *BlueskyService) SetTokens(accessToken, refreshToken string) {
	s.accessToken = accessToken
	s.refreshToken = refreshToken
	s.authenticated = true

	if expiry, err := parseJWTExpiry(accessToken); err == nil {
		s.tokenExpiry = expiry
	}
}

// GetAccessToken returns the current access token
func (s *BlueskyService) GetAccessToken() string {
	return s.accessToken
}

// GetRefreshToken returns the current refresh token
func (s *BlueskyService) GetRefreshToken() string {
	return s.refreshToken
}

// shouldRefreshToken checks if token is 90% through its lifetime
func (s *BlueskyService) shouldRefreshToken() bool {
	if s.tokenExpiry.IsZero() {
		return false
	}

	now := time.Now()
	lifetime := s.tokenExpiry.Sub(now)
	threshold := lifetime / 10 // 10% remaining

	return now.Add(threshold).After(s.tokenExpiry)
}

// refreshAccessToken uses the refresh token to get a new access token
func (s *BlueskyService) refreshAccessToken(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "POST",
		s.baseURL+"/xrpc/com.atproto.server.refreshSession", nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+s.refreshToken)

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyText, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("token refresh failed: %s - %s", resp.Status, string(bodyText))
	}

	var session CreateSessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&session); err != nil {
		return err
	}

	s.accessToken = session.AccessJwt
	s.refreshToken = session.RefreshJwt

	if expiry, err := parseJWTExpiry(s.accessToken); err == nil {
		s.tokenExpiry = expiry
	}

	return nil
}

// parseJWTExpiry extracts the expiry time from a JWT token
func parseJWTExpiry(token string) (time.Time, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return time.Time{}, errors.New("invalid JWT format")
	}

	// Decode payload (second part)
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return time.Time{}, err
	}

	var claims struct {
		Exp int64 `json:"exp"`
	}

	if err := json.Unmarshal(payload, &claims); err != nil {
		return time.Time{}, err
	}

	if claims.Exp == 0 {
		return time.Time{}, errors.New("no exp claim in JWT")
	}

	return time.Unix(claims.Exp, 0), nil
}
