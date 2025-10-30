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
	"sync"
	"time"
)

const (
	defaultServiceURL = "https://bsky.social"
	defaultTimeout    = 30 * time.Second
)

type jwtClaims struct {
	Exp int64 `json:"exp"`
}

// BlueskyService implements the [Service] interface for AT Protocol / Bluesky API
type BlueskyService struct {
	baseURL       string
	client        *http.Client
	accessToken   string
	refreshToken  string
	tokenExpiry   time.Time
	authenticated bool
	did           string
	handle        string
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
	s.did = session.Did
	s.handle = session.Handle
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
	s.did = ""
	s.handle = ""
	s.tokenExpiry = time.Time{}
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

// GetFollows fetches the list of accounts that an actor follows.
// Limit must be between 1-100 (API enforced); defaults to 50 if not specified.
func (s *BlueskyService) GetFollows(ctx context.Context, actor string, limit int, cursor string) (*GetFollowsResponse, error) {
	if actor == "" {
		return nil, fmt.Errorf("actor is required")
	}

	if limit < 1 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	params := fmt.Sprintf("actor=%s&limit=%d", actor, limit)
	if cursor != "" {
		params += "&cursor=" + cursor
	}

	url := "/xrpc/app.bsky.graph.getFollows?" + params

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

// GetFollowers fetches the list of accounts that follow an actor.
// Limit must be between 1-100 (API enforced); defaults to 50 if not specified.
func (s *BlueskyService) GetFollowers(ctx context.Context, actor string, limit int, cursor string) (*GetFollowersResponse, error) {
	if actor == "" {
		return nil, fmt.Errorf("actor is required")
	}

	if limit < 1 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	params := fmt.Sprintf("actor=%s&limit=%d", actor, limit)
	if cursor != "" {
		params += "&cursor=" + cursor
	}

	url := "/xrpc/app.bsky.graph.getFollowers?" + params

	resp, err := s.Request(ctx, "GET", url, nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyText, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("getFollowers failed: %s - %s", resp.Status, string(bodyText))
	}

	var followers GetFollowersResponse
	if err := json.NewDecoder(resp.Body).Decode(&followers); err != nil {
		return nil, err
	}

	return &followers, nil
}

// GetProfile fetches detailed profile information for an actor.
// Actor can be a DID or handle (e.g., "did:plc:..." or "alice.bsky.social").
func (s *BlueskyService) GetProfile(ctx context.Context, actor string) (*ActorProfile, error) {
	url := fmt.Sprintf("/xrpc/app.bsky.actor.getProfile?actor=%s", actor)

	resp, err := s.Request(ctx, "GET", url, nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyText, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("getProfile failed: %s - %s", resp.Status, string(bodyText))
	}

	var profile ActorProfile
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return nil, err
	}

	return &profile, nil
}

// SearchActors searches for actors (users) matching the query string.
// Returns actor profiles with pagination support.
func (s *BlueskyService) SearchActors(ctx context.Context, query string, limit int, cursor string) (*SearchActorsResponse, error) {
	urlPath := fmt.Sprintf("/xrpc/app.bsky.actor.searchActors?q=%s&limit=%d", strings.ReplaceAll(query, " ", "+"), limit)
	if cursor != "" {
		urlPath += "&cursor=" + cursor
	}

	resp, err := s.Request(ctx, "GET", urlPath, nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyText, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("searchActors failed: %s - %s", resp.Status, string(bodyText))
	}

	var result SearchActorsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// SearchPosts searches for posts matching the query string returning feed view posts with pagination support.
func (s *BlueskyService) SearchPosts(ctx context.Context, query string, limit int, cursor string) (*SearchPostsResponse, error) {
	urlPath := fmt.Sprintf("/xrpc/app.bsky.feed.searchPosts?q=%s&limit=%d", strings.ReplaceAll(query, " ", "+"), limit)
	if cursor != "" {
		urlPath += "&cursor=" + cursor
	}

	resp, err := s.Request(ctx, "GET", urlPath, nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyText, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("searchPosts failed: %s - %s", resp.Status, string(bodyText))
	}

	var result SearchPostsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetPosts fetches specific posts by their AT URIs.
// Accepts a slice of URIs and returns the corresponding posts.
func (s *BlueskyService) GetPosts(ctx context.Context, uris []string) (*GetPostsResponse, error) {
	if len(uris) == 0 {
		return &GetPostsResponse{Posts: []FeedViewPost{}}, nil
	}

	url := "/xrpc/app.bsky.feed.getPosts?"
	for i, uri := range uris {
		if i > 0 {
			url += "&"
		}
		url += "uris=" + uri
	}

	resp, err := s.Request(ctx, "GET", url, nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyText, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("getPosts failed: %s - %s", resp.Status, string(bodyText))
	}

	var result GetPostsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
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

// GetLastPostDate fetches the most recent post date for an actor.
// Returns zero time if the actor has no posts or if an error occurs.
func (s *BlueskyService) GetLastPostDate(ctx context.Context, actor string) (time.Time, error) {
	feed, err := s.GetAuthorFeed(ctx, actor, 1, "")
	if err != nil {
		return time.Time{}, err
	}

	if len(feed.Feed) == 0 {
		return time.Time{}, nil
	}

	indexedAt := feed.Feed[0].Post.IndexedAt
	lastPost, err := time.Parse(time.RFC3339, indexedAt)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse indexedAt: %w", err)
	}

	return lastPost, nil
}

// BatchGetLastPostDates fetches last post dates for multiple actors concurrently.
// Uses a semaphore to limit concurrent requests to maxConcurrent.
// Returns a map of actor DID/handle to their last post date.
func (s *BlueskyService) BatchGetLastPostDates(ctx context.Context, actors []string, maxConcurrent int) map[string]time.Time {
	results := make(map[string]time.Time)
	resultsMu := &sync.Mutex{}
	sem := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup

	for _, actor := range actors {
		wg.Add(1)
		go func(a string) {
			defer wg.Done()

			sem <- struct{}{}
			defer func() { <-sem }()

			lastPost, err := s.GetLastPostDate(ctx, a)
			if err != nil {
				return
			}

			resultsMu.Lock()
			results[a] = lastPost
			resultsMu.Unlock()
		}(actor)
	}

	wg.Wait()
	return results
}

// BatchGetProfiles fetches full profiles for multiple actors concurrently.
// Uses a semaphore to limit concurrent requests to maxConcurrent.
// Returns a map of actor DID/handle to their full ActorProfile.
func (s *BlueskyService) BatchGetProfiles(ctx context.Context, actors []string, maxConcurrent int) map[string]*ActorProfile {
	results := make(map[string]*ActorProfile)
	resultsMu := &sync.Mutex{}
	sem := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup

	for _, actor := range actors {
		wg.Add(1)
		go func(a string) {
			defer wg.Done()

			sem <- struct{}{}
			defer func() { <-sem }()

			profile, err := s.GetProfile(ctx, a)
			if err != nil {
				return
			}

			resultsMu.Lock()
			results[a] = profile
			resultsMu.Unlock()
		}(actor)
	}

	wg.Wait()
	return results
}

// PostRate holds posting frequency metrics for an actor
type PostRate struct {
	PostsPerDay  float64
	LastPostDate time.Time
	SampleSize   int
}

// BatchGetPostRates calculates posting rates for multiple actors concurrently.
// Samples recent posts from each actor and calculates posts per day over the lookback period.
// Uses a semaphore to limit concurrent requests to maxConcurrent.
// Returns a map of actor DID/handle to their PostRate metrics.
func (s *BlueskyService) BatchGetPostRates(ctx context.Context, actors []string, sampleSize int, lookbackDays int, maxConcurrent int, progressFn func(current, total int)) map[string]*PostRate {
	results := make(map[string]*PostRate)
	resultsMu := &sync.Mutex{}
	sem := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup

	completed := 0
	completedMu := &sync.Mutex{}
	total := len(actors)

	for _, actor := range actors {
		wg.Add(1)
		go func(a string) {
			defer wg.Done()

			sem <- struct{}{}
			defer func() { <-sem }()

			feed, err := s.GetAuthorFeed(ctx, a, sampleSize, "")
			if err != nil {
				return
			}

			if len(feed.Feed) == 0 {
				resultsMu.Lock()
				results[a] = &PostRate{
					PostsPerDay:  0,
					LastPostDate: time.Time{},
					SampleSize:   0,
				}
				resultsMu.Unlock()

				completedMu.Lock()
				completed++
				if progressFn != nil {
					progressFn(completed, total)
				}
				completedMu.Unlock()
				return
			}

			// Get the last post date
			lastPost, err := time.Parse(time.RFC3339, feed.Feed[0].Post.IndexedAt)
			if err != nil {
				return
			}

			// Filter posts within lookback window
			cutoffTime := time.Now().AddDate(0, 0, -lookbackDays)
			recentPosts := 0
			for _, post := range feed.Feed {
				indexedAt, err := time.Parse(time.RFC3339, post.Post.IndexedAt)
				if err != nil {
					continue
				}
				if indexedAt.After(cutoffTime) {
					recentPosts++
				}
			}

			postsPerDay := float64(recentPosts) / float64(lookbackDays)

			resultsMu.Lock()
			results[a] = &PostRate{
				PostsPerDay:  postsPerDay,
				LastPostDate: lastPost,
				SampleSize:   len(feed.Feed),
			}
			resultsMu.Unlock()

			completedMu.Lock()
			completed++
			if progressFn != nil {
				progressFn(completed, total)
			}
			completedMu.Unlock()
		}(actor)
	}

	wg.Wait()
	return results
}

// GetAccessToken returns the current access token
func (s *BlueskyService) GetAccessToken() string {
	return s.accessToken
}

// GetRefreshToken returns the current refresh token
func (s *BlueskyService) GetRefreshToken() string {
	return s.refreshToken
}

// GetDid returns the authenticated user's DID
func (s *BlueskyService) GetDid() string {
	return s.did
}

// GetHandle returns the authenticated user's handle
func (s *BlueskyService) GetHandle() string {
	return s.handle
}

// SetDid sets the authenticated user's DID
func (s *BlueskyService) SetDid(did string) {
	s.did = did
}

// SetHandle sets the authenticated user's handle
func (s *BlueskyService) SetHandle(handle string) {
	s.handle = handle
}

// shouldRefreshToken checks if token is 90% through its lifetime
func (s *BlueskyService) shouldRefreshToken() bool {
	if s.tokenExpiry.IsZero() {
		return false
	}

	now := time.Now()
	lifetime := s.tokenExpiry.Sub(now)
	threshold := lifetime / 10

	return now.Add(threshold).After(s.tokenExpiry)
}

// refreshAccessToken uses the refresh token to get a new access token
func (s *BlueskyService) refreshAccessToken(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "POST", s.baseURL+"/xrpc/com.atproto.server.refreshSession", nil)
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

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return time.Time{}, err
	}

	var claims jwtClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return time.Time{}, err
	}

	if claims.Exp == 0 {
		return time.Time{}, errors.New("no exp claim in JWT")
	}

	return time.Unix(claims.Exp, 0), nil
}
