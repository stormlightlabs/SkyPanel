package store

import (
	"context"
	"io"
	"net/http"
)

// Enumeration of [Service] implementations
type ServiceID = int32

// String representation of [ServiceID] (e.g., "Bluesky", "GitHub", "WeatherAPI")
type ServiceIdentifier = string

// Service defines the general contract for any API service client.
// It abstracts away protocol details and provides lifecycle, request, and authentication semantics.
type Service interface {
	// Name returns the service identifier
	Name() ServiceIdentifier
	// BaseURL returns the root endpoint of the remote API.
	BaseURL() string
	// Authenticated reports whether the client is currently authorized.
	Authenticated() bool
	// Authenticate establishes credentials with the service (token, key, etc.).
	Authenticate(ctx context.Context, credentials any) error
	// Request performs a generic API request and returns the raw response.
	// Implementations may wrap or replace http.Client as needed.
	Request(ctx context.Context, method, path string, body io.Reader, headers map[string]string) (*http.Response, error)
	// HealthCheck verifies connectivity and minimal readiness of the remote API.
	HealthCheck(ctx context.Context) error
	// Close releases underlying network or session resources.
	Close() error
}

// CreateSessionResponse models response from com.atproto.server.createSession.
// Returns authentication tokens and user metadata after successful login.
type CreateSessionResponse struct {
	Did             string  `json:"did"`
	DidDoc          *DidDoc `json:"didDoc,omitempty"`
	Handle          string  `json:"handle"`
	Email           string  `json:"email,omitempty"`
	EmailConfirmed  bool    `json:"emailConfirmed"`
	EmailAuthFactor bool    `json:"emailAuthFactor"`
	AccessJwt       string  `json:"accessJwt"`
	RefreshJwt      string  `json:"refreshJwt"`
	Active          bool    `json:"active"`
	Status          string  `json:"status,omitempty"`
}

// DidDoc represents a DID document as per W3C DID specification.
// Contains verification methods and service endpoints for the DID.
type DidDoc struct {
	Context            []string             `json:"@context"`
	ID                 string               `json:"id"`
	AlsoKnownAs        []string             `json:"alsoKnownAs,omitempty"`
	VerificationMethod []VerificationMethod `json:"verificationMethod,omitempty"`
	Service            []DidService         `json:"service,omitempty"`
}

// VerificationMethod represents a cryptographic public key for DID verification
type VerificationMethod struct {
	ID                 string `json:"id"`
	Type               string `json:"type"`
	Controller         string `json:"controller"`
	PublicKeyMultibase string `json:"publicKeyMultibase,omitempty"`
}

// DidService represents a service endpoint in the DID document (e.g., PDS endpoint).
// Renamed from Service to avoid collision with the Service interface.
type DidService struct {
	ID              string `json:"id"`
	Type            string `json:"type"`
	ServiceEndpoint string `json:"serviceEndpoint"`
}

// CreateRecordResponse models response from com.atproto.repo.createRecord.
// Returns the URI and CID of the newly created record along with commit metadata.
type CreateRecordResponse struct {
	Uri    string `json:"uri"`
	Cid    string `json:"cid"`
	Commit struct {
		Cid string `json:"cid"`
		Rev string `json:"rev"`
	} `json:"commit"`
	ValidationStatus string `json:"validationStatus,omitempty"`
}

// GetTimelineResponse models response from app.bsky.feed.getTimeline.
// Returns the authenticated user's home timeline feed with pagination support.
type GetTimelineResponse struct {
	Cursor string         `json:"cursor,omitempty"`
	Feed   []FeedViewPost `json:"feed"`
}

// GetAuthorFeedResponse models response from app.bsky.feed.getAuthorFeed.
// Returns posts by a specific author with optional reply context.
type GetAuthorFeedResponse struct {
	Cursor string         `json:"cursor,omitempty"`
	Feed   []FeedViewPost `json:"feed"`
}

// GetFollowsResponse models response from app.bsky.graph.getFollows.
// Returns list of accounts that a given actor follows.
type GetFollowsResponse struct {
	Subject string         `json:"subject,omitempty"`
	Cursor  string         `json:"cursor,omitempty"`
	Follows []ActorProfile `json:"follows"`
}

// FeedViewPost represents a single item in a feed, containing the post and optional context.
// Includes repost reasoning and reply threading context when applicable.
type FeedViewPost struct {
	Post   *PostView   `json:"post"`
	Reason *ReasonView `json:"reason,omitempty"`
	Reply  *ReplyRefs  `json:"reply,omitempty"`
}

// PostView represents a post with full metadata including engagement metrics.
// Contains author info, content, embeds, and viewer-specific state.
type PostView struct {
	Uri           string        `json:"uri"`
	Cid           string        `json:"cid"`
	Author        *ActorProfile `json:"author"`
	Record        any           `json:"record"`
	Embed         any           `json:"embed,omitempty"`
	ReplyCount    int           `json:"replyCount"`
	RepostCount   int           `json:"repostCount"`
	LikeCount     int           `json:"likeCount"`
	QuoteCount    int           `json:"quoteCount"`
	BookmarkCount int           `json:"bookmarkCount,omitempty"`
	IndexedAt     string        `json:"indexedAt"`
	Viewer        *ViewerState  `json:"viewer,omitempty"`
	Labels        []Label       `json:"labels,omitempty"`
}

// ReasonView indicates why a post appears in the feed (e.g., repost by followed user)
type ReasonView struct {
	Type      string        `json:"$type"`
	By        *ActorProfile `json:"by,omitempty"`
	IndexedAt string        `json:"indexedAt,omitempty"`
}

// ReplyRefs contains references to the root and parent posts in a thread
type ReplyRefs struct {
	Root   *PostRef `json:"root"`
	Parent *PostRef `json:"parent"`
}

// PostRef is a minimal reference to a post (URI and CID only)
type PostRef struct {
	Uri string `json:"uri"`
	Cid string `json:"cid"`
}

// ActorProfile represents a user's profile with social graph metadata.
// Includes display information, verification status, and viewer's relationship to the actor.
type ActorProfile struct {
	Did            string        `json:"did"`
	Handle         string        `json:"handle"`
	DisplayName    string        `json:"displayName,omitempty"`
	Description    string        `json:"description,omitempty"`
	Avatar         string        `json:"avatar,omitempty"`
	Banner         string        `json:"banner,omitempty"`
	FollowersCount int           `json:"followersCount,omitempty"`
	FollowsCount   int           `json:"followsCount,omitempty"`
	PostsCount     int           `json:"postsCount,omitempty"`
	Associated     *Associated   `json:"associated,omitempty"`
	Viewer         *ViewerState  `json:"viewer,omitempty"`
	Labels         []Label       `json:"labels,omitempty"`
	CreatedAt      string        `json:"createdAt,omitempty"`
	IndexedAt      string        `json:"indexedAt,omitempty"`
	Verification   *Verification `json:"verification,omitempty"`
	Status         *ActorStatus  `json:"status,omitempty"`
}

// Associated contains chat and subscription preferences for an actor
type Associated struct {
	Chat                 *ChatSettings         `json:"chat,omitempty"`
	ActivitySubscription *SubscriptionSettings `json:"activitySubscription,omitempty"`
}

// ChatSettings defines who can send chat messages to this actor
type ChatSettings struct {
	AllowIncoming string `json:"allowIncoming"` // "all", "following", "none"
}

// SubscriptionSettings defines who can subscribe to activity updates
type SubscriptionSettings struct {
	AllowSubscriptions string `json:"allowSubscriptions"` // "all", "followers", "none"
}

// ViewerState represents the current user's relationship with an actor or post.
// Tracks muting, blocking, following, and engagement state.
type ViewerState struct {
	Muted             bool   `json:"muted"`
	BlockedBy         bool   `json:"blockedBy"`
	Blocking          string `json:"blocking,omitempty"`
	Following         string `json:"following,omitempty"`
	FollowedBy        string `json:"followedBy,omitempty"`
	Bookmarked        bool   `json:"bookmarked,omitempty"`
	ThreadMuted       bool   `json:"threadMuted,omitempty"`
	EmbeddingDisabled bool   `json:"embeddingDisabled,omitempty"`
	Pinned            bool   `json:"pinned,omitempty"`
	Like              string `json:"like,omitempty"`
	Repost            string `json:"repost,omitempty"`
}

// Label represents a content label applied to a post or actor (e.g., for moderation)
type Label struct {
	Src string `json:"src"`
	Uri string `json:"uri"`
	Cid string `json:"cid,omitempty"`
	Val string `json:"val"`
	Cts string `json:"cts"`           // created timestamp
	Exp string `json:"exp,omitempty"` // expiration
	Sig []byte `json:"sig,omitempty"`
}

// Verification contains verification status and trusted verifier information
type Verification struct {
	Verifications         []VerificationRecord `json:"verifications,omitempty"`
	VerifiedStatus        string               `json:"verifiedStatus,omitempty"`
	TrustedVerifierStatus string               `json:"trustedVerifierStatus,omitempty"`
}

// VerificationRecord represents a single verification from a trusted verifier
type VerificationRecord struct {
	Issuer    string `json:"issuer"`
	Uri       string `json:"uri"`
	IsValid   bool   `json:"isValid"`
	CreatedAt string `json:"createdAt"`
}

// ActorStatus represents live status (e.g., streaming status) for an actor
type ActorStatus struct {
	Record    any    `json:"record,omitempty"`
	Status    string `json:"status,omitempty"`
	Embed     any    `json:"embed,omitempty"`
	ExpiresAt string `json:"expiresAt,omitempty"`
	IsActive  bool   `json:"isActive"`
}
