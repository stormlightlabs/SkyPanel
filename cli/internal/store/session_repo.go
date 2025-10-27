package store

import (
	"context"
	"errors"
	"time"

	"github.com/stormlightlabs/skypanel/cli/internal/config"
)

// SessionRepository implements Repository for SessionModel.
// Persists session data to ~/.skycli/.config.json with encrypted tokens.
type SessionRepository struct {
	config *config.Config
}

// NewSessionRepository creates a new session repository instance
func NewSessionRepository() (*SessionRepository, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	return &SessionRepository{config: cfg}, nil
}

// Init ensures the config directory exists and loads existing configuration
func (r *SessionRepository) Init(ctx context.Context) error {
	return config.EnsureConfigDir()
}

// Close is a no-op for file-based storage
func (r *SessionRepository) Close() error {
	return nil
}

// Get retrieves the current session by ID (only one session supported)
func (r *SessionRepository) Get(ctx context.Context, id string) (Model, error) {
	if r.config.Session == nil {
		return nil, errors.New("no active session")
	}

	accessToken, err := r.config.Session.GetAccessToken()
	if err != nil {
		return nil, err
	}

	refreshToken, err := r.config.Session.GetRefreshToken()
	if err != nil {
		return nil, err
	}

	session := &SessionModel{
		Handle:     r.config.Session.Handle,
		Token:      accessToken + "|" + refreshToken,
		ServiceURL: r.config.Session.ServiceURL,
		IsValid:    true,
	}
	session.SetID(r.config.Session.Did)
	session.SetCreatedAt(time.Now()) // TODO: store creation time
	session.SetUpdatedAt(time.Now())

	return session, nil
}

// List returns all sessions (only one supported)
func (r *SessionRepository) List(ctx context.Context) ([]Model, error) {
	if r.config.Session == nil {
		return []Model{}, nil
	}

	session, err := r.Get(ctx, r.config.Session.Did)
	if err != nil {
		return nil, err
	}

	return []Model{session}, nil
}

// Save persists a session with encrypted tokens to ~/.skycli/.config.json
func (r *SessionRepository) Save(ctx context.Context, model Model) error {
	session, ok := model.(*SessionModel)
	if !ok {
		return errors.New("invalid model type: expected *SessionModel")
	}

	var accessToken, refreshToken string
	parts := splitToken(session.Token)
	if len(parts) == 2 {
		accessToken = parts[0]
		refreshToken = parts[1]
	} else {
		accessToken = session.Token
	}

	sessionConfig := &config.SessionConfig{
		Handle:     session.Handle,
		Did:        session.ID(),
		ServiceURL: session.ServiceURL,
		Email:      "", // TODO: add Email field to SessionModel
	}

	if err := sessionConfig.SetAccessToken(accessToken); err != nil {
		return err
	}

	if err := sessionConfig.SetRefreshToken(refreshToken); err != nil {
		return err
	}

	r.config.Session = sessionConfig
	return r.config.Save()
}

// Delete removes the current session
func (r *SessionRepository) Delete(ctx context.Context, id string) error {
	r.config.Session = nil
	return r.config.Save()
}

// GetAccessToken returns the decrypted access token for the current session
func (r *SessionRepository) GetAccessToken(ctx context.Context) (string, error) {
	if r.config.Session == nil {
		return "", errors.New("no active session")
	}
	return r.config.Session.GetAccessToken()
}

// GetRefreshToken returns the decrypted refresh token for the current session
func (r *SessionRepository) GetRefreshToken(ctx context.Context) (string, error) {
	if r.config.Session == nil {
		return "", errors.New("no active session")
	}
	return r.config.Session.GetRefreshToken()
}

// UpdateTokens updates both access and refresh tokens for the current session
func (r *SessionRepository) UpdateTokens(ctx context.Context, accessToken, refreshToken string) error {
	if r.config.Session == nil {
		return errors.New("no active session")
	}

	if err := r.config.Session.SetAccessToken(accessToken); err != nil {
		return err
	}

	if err := r.config.Session.SetRefreshToken(refreshToken); err != nil {
		return err
	}

	return r.config.Save()
}

// HasValidSession checks if there is an active session
func (r *SessionRepository) HasValidSession(ctx context.Context) bool {
	return r.config.Session != nil && r.config.Session.EncryptedAccess != ""
}

// GetDid returns the DID for the current session
func (r *SessionRepository) GetDid(ctx context.Context) (string, error) {
	if r.config.Session == nil {
		return "", errors.New("no active session")
	}
	return r.config.Session.Did, nil
}

// GetHandle returns the handle for the current session
func (r *SessionRepository) GetHandle(ctx context.Context) (string, error) {
	if r.config.Session == nil {
		return "", errors.New("no active session")
	}
	return r.config.Session.Handle, nil
}

// splitToken splits a combined token string (accessToken|refreshToken)
func splitToken(token string) []string {
	result := []string{}
	current := ""
	for _, ch := range token {
		if ch == '|' {
			result = append(result, current)
			current = ""
		} else {
			current += string(ch)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}
