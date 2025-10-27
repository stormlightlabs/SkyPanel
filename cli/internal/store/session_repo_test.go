package store

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stormlightlabs/skypanel/cli/internal/utils"
)

// TestNewSessionRepository_Success verifies repository creation
func TestNewSessionRepository_Success(t *testing.T) {
	_, cleanup := utils.SetupTestConfig(t)
	defer cleanup()

	repo, err := NewSessionRepository()
	if err != nil {
		t.Fatalf("NewSessionRepository failed: %v", err)
	}

	if repo == nil {
		t.Fatal("expected non-nil repository")
	}

	if repo.config == nil {
		t.Fatal("expected non-nil config")
	}
}

// TestInit_Success verifies initialization creates config directory
func TestInit_Success(t *testing.T) {
	configDir, cleanup := utils.SetupTestConfig(t)
	defer cleanup()

	repo, err := NewSessionRepository()
	if err != nil {
		t.Fatalf("NewSessionRepository failed: %v", err)
	}

	ctx := context.Background()
	err = repo.Init(ctx)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		t.Error("config directory was not created")
	}
}

// TestClose_Success verifies Close is a no-op
func TestClose_Success(t *testing.T) {
	_, cleanup := utils.SetupTestConfig(t)
	defer cleanup()

	repo, err := NewSessionRepository()
	if err != nil {
		t.Fatalf("NewSessionRepository failed: %v", err)
	}

	err = repo.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

// TestGet_NoSession verifies Get returns error when no session exists
func TestGet_NoSession(t *testing.T) {
	_, cleanup := utils.SetupTestConfig(t)
	defer cleanup()

	repo, err := NewSessionRepository()
	if err != nil {
		t.Fatalf("NewSessionRepository failed: %v", err)
	}

	ctx := context.Background()
	_, err = repo.Get(ctx, "any-id")
	if err == nil {
		t.Error("expected error for non-existent session, got nil")
	}

	if !strings.Contains(err.Error(), "no active session") {
		t.Errorf("expected 'no active session' error, got: %v", err)
	}
}

// TestSaveAndGet_Success verifies saving and retrieving a session
func TestSaveAndGet_Success(t *testing.T) {
	_, cleanup := utils.SetupTestConfig(t)
	defer cleanup()

	repo, err := NewSessionRepository()
	if err != nil {
		t.Fatalf("NewSessionRepository failed: %v", err)
	}

	ctx := context.Background()

	session := &SessionModel{
		Handle:     "test.bsky.social",
		Token:      "access_token|refresh_token",
		ServiceURL: "https://bsky.social",
		IsValid:    true,
	}
	session.SetID("did:plc:test123")

	err = repo.Save(ctx, session)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	retrieved, err := repo.Get(ctx, session.ID())
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	retrievedSession, ok := retrieved.(*SessionModel)
	if !ok {
		t.Fatal("retrieved model is not a SessionModel")
	}

	if retrievedSession.Handle != session.Handle {
		t.Errorf("expected handle %s, got %s", session.Handle, retrievedSession.Handle)
	}
	if retrievedSession.ServiceURL != session.ServiceURL {
		t.Errorf("expected service URL %s, got %s", session.ServiceURL, retrievedSession.ServiceURL)
	}
	if retrievedSession.ID() != session.ID() {
		t.Errorf("expected ID %s, got %s", session.ID(), retrievedSession.ID())
	}
}

// TestSave_InvalidModel verifies Save rejects non-SessionModel types
func TestSave_InvalidModel(t *testing.T) {
	_, cleanup := utils.SetupTestConfig(t)
	defer cleanup()

	repo, err := NewSessionRepository()
	if err != nil {
		t.Fatalf("NewSessionRepository failed: %v", err)
	}

	ctx := context.Background()
	post := &PostModel{}
	err = repo.Save(ctx, post)
	if err == nil {
		t.Error("expected error for invalid model type, got nil")
	}

	if !strings.Contains(err.Error(), "invalid model type") {
		t.Errorf("expected 'invalid model type' error, got: %v", err)
	}
}

// TestList_NoSession verifies List returns empty slice when no session exists
func TestList_NoSession(t *testing.T) {
	_, cleanup := utils.SetupTestConfig(t)
	defer cleanup()

	repo, err := NewSessionRepository()
	if err != nil {
		t.Fatalf("NewSessionRepository failed: %v", err)
	}

	ctx := context.Background()
	models, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(models) != 0 {
		t.Errorf("expected empty list, got %d models", len(models))
	}
}

// TestList_WithSession verifies List returns the saved session
func TestList_WithSession(t *testing.T) {
	_, cleanup := utils.SetupTestConfig(t)
	defer cleanup()

	repo, err := NewSessionRepository()
	if err != nil {
		t.Fatalf("NewSessionRepository failed: %v", err)
	}

	ctx := context.Background()

	session := &SessionModel{
		Handle:     "test.bsky.social",
		Token:      "access_token|refresh_token",
		ServiceURL: "https://bsky.social",
		IsValid:    true,
	}
	session.SetID("did:plc:test123")

	err = repo.Save(ctx, session)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	models, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(models) != 1 {
		t.Fatalf("expected 1 session, got %d", len(models))
	}

	retrievedSession, ok := models[0].(*SessionModel)
	if !ok {
		t.Fatal("listed model is not a SessionModel")
	}

	if retrievedSession.Handle != session.Handle {
		t.Errorf("expected handle %s, got %s", session.Handle, retrievedSession.Handle)
	}
}

// TestDelete_Success verifies session deletion
func TestDelete_Success(t *testing.T) {
	_, cleanup := utils.SetupTestConfig(t)
	defer cleanup()

	repo, err := NewSessionRepository()
	if err != nil {
		t.Fatalf("NewSessionRepository failed: %v", err)
	}

	ctx := context.Background()

	session := &SessionModel{
		Handle:     "test.bsky.social",
		Token:      "access_token|refresh_token",
		ServiceURL: "https://bsky.social",
		IsValid:    true,
	}
	session.SetID("did:plc:test123")

	err = repo.Save(ctx, session)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	err = repo.Delete(ctx, session.ID())
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = repo.Get(ctx, session.ID())
	if err == nil {
		t.Error("expected error after deletion, got nil")
	}
}

// TestGetAccessToken_NoSession verifies GetAccessToken errors when no session exists
func TestGetAccessToken_NoSession(t *testing.T) {
	_, cleanup := utils.SetupTestConfig(t)
	defer cleanup()

	repo, err := NewSessionRepository()
	if err != nil {
		t.Fatalf("NewSessionRepository failed: %v", err)
	}

	ctx := context.Background()
	_, err = repo.GetAccessToken(ctx)
	if err == nil {
		t.Error("expected error for no session, got nil")
	}

	if !strings.Contains(err.Error(), "no active session") {
		t.Errorf("expected 'no active session' error, got: %v", err)
	}
}

// TestGetAccessToken_Success verifies retrieving access token
func TestGetAccessToken_Success(t *testing.T) {
	_, cleanup := utils.SetupTestConfig(t)
	defer cleanup()

	repo, err := NewSessionRepository()
	if err != nil {
		t.Fatalf("NewSessionRepository failed: %v", err)
	}

	ctx := context.Background()

	session := &SessionModel{
		Handle:     "test.bsky.social",
		Token:      "test_access_token|test_refresh_token",
		ServiceURL: "https://bsky.social",
		IsValid:    true,
	}
	session.SetID("did:plc:test123")

	err = repo.Save(ctx, session)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	accessToken, err := repo.GetAccessToken(ctx)
	if err != nil {
		t.Fatalf("GetAccessToken failed: %v", err)
	}

	if accessToken == "" {
		t.Error("expected non-empty access token")
	}
}

// TestGetRefreshToken_NoSession verifies GetRefreshToken errors when no session exists
func TestGetRefreshToken_NoSession(t *testing.T) {
	_, cleanup := utils.SetupTestConfig(t)
	defer cleanup()

	repo, err := NewSessionRepository()
	if err != nil {
		t.Fatalf("NewSessionRepository failed: %v", err)
	}

	ctx := context.Background()
	_, err = repo.GetRefreshToken(ctx)
	if err == nil {
		t.Error("expected error for no session, got nil")
	}

	if !strings.Contains(err.Error(), "no active session") {
		t.Errorf("expected 'no active session' error, got: %v", err)
	}
}

// TestGetRefreshToken_Success verifies retrieving refresh token
func TestGetRefreshToken_Success(t *testing.T) {
	_, cleanup := utils.SetupTestConfig(t)
	defer cleanup()

	repo, err := NewSessionRepository()
	if err != nil {
		t.Fatalf("NewSessionRepository failed: %v", err)
	}

	ctx := context.Background()

	session := &SessionModel{
		Handle:     "test.bsky.social",
		Token:      "test_access_token|test_refresh_token",
		ServiceURL: "https://bsky.social",
		IsValid:    true,
	}
	session.SetID("did:plc:test123")

	err = repo.Save(ctx, session)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	refreshToken, err := repo.GetRefreshToken(ctx)
	if err != nil {
		t.Fatalf("GetRefreshToken failed: %v", err)
	}

	if refreshToken == "" {
		t.Error("expected non-empty refresh token")
	}
}

// TestUpdateTokens_NoSession verifies UpdateTokens errors when no session exists
func TestUpdateTokens_NoSession(t *testing.T) {
	_, cleanup := utils.SetupTestConfig(t)
	defer cleanup()

	repo, err := NewSessionRepository()
	if err != nil {
		t.Fatalf("NewSessionRepository failed: %v", err)
	}

	ctx := context.Background()
	err = repo.UpdateTokens(ctx, "new_access", "new_refresh")
	if err == nil {
		t.Error("expected error for no session, got nil")
	}

	if !strings.Contains(err.Error(), "no active session") {
		t.Errorf("expected 'no active session' error, got: %v", err)
	}
}

// TestUpdateTokens_Success verifies token updates
func TestUpdateTokens_Success(t *testing.T) {
	_, cleanup := utils.SetupTestConfig(t)
	defer cleanup()

	repo, err := NewSessionRepository()
	if err != nil {
		t.Fatalf("NewSessionRepository failed: %v", err)
	}

	ctx := context.Background()

	session := &SessionModel{
		Handle:     "test.bsky.social",
		Token:      "old_access|old_refresh",
		ServiceURL: "https://bsky.social",
		IsValid:    true,
	}
	session.SetID("did:plc:test123")

	err = repo.Save(ctx, session)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	err = repo.UpdateTokens(ctx, "new_access_token", "new_refresh_token")
	if err != nil {
		t.Fatalf("UpdateTokens failed: %v", err)
	}

	accessToken, err := repo.GetAccessToken(ctx)
	if err != nil {
		t.Fatalf("GetAccessToken failed: %v", err)
	}

	refreshToken, err := repo.GetRefreshToken(ctx)
	if err != nil {
		t.Fatalf("GetRefreshToken failed: %v", err)
	}

	if accessToken == "" {
		t.Error("access token should not be empty after update")
	}

	if refreshToken == "" {
		t.Error("refresh token should not be empty after update")
	}
}

// TestHasValidSession_NoSession verifies HasValidSession returns false when no session
func TestHasValidSession_NoSession(t *testing.T) {
	_, cleanup := utils.SetupTestConfig(t)
	defer cleanup()

	repo, err := NewSessionRepository()
	if err != nil {
		t.Fatalf("NewSessionRepository failed: %v", err)
	}

	ctx := context.Background()
	hasSession := repo.HasValidSession(ctx)
	if hasSession {
		t.Error("expected false for no session, got true")
	}
}

// TestHasValidSession_WithSession verifies HasValidSession returns true when session exists
func TestHasValidSession_WithSession(t *testing.T) {
	_, cleanup := utils.SetupTestConfig(t)
	defer cleanup()

	repo, err := NewSessionRepository()
	if err != nil {
		t.Fatalf("NewSessionRepository failed: %v", err)
	}

	ctx := context.Background()

	session := &SessionModel{
		Handle:     "test.bsky.social",
		Token:      "access_token|refresh_token",
		ServiceURL: "https://bsky.social",
		IsValid:    true,
	}
	session.SetID("did:plc:test123")

	err = repo.Save(ctx, session)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	hasSession := repo.HasValidSession(ctx)
	if !hasSession {
		t.Error("expected true for existing session, got false")
	}
}

// TestGetDid_NoSession verifies GetDid errors when no session exists
func TestGetDid_NoSession(t *testing.T) {
	_, cleanup := utils.SetupTestConfig(t)
	defer cleanup()

	repo, err := NewSessionRepository()
	if err != nil {
		t.Fatalf("NewSessionRepository failed: %v", err)
	}

	ctx := context.Background()
	_, err = repo.GetDid(ctx)
	if err == nil {
		t.Error("expected error for no session, got nil")
	}

	if !strings.Contains(err.Error(), "no active session") {
		t.Errorf("expected 'no active session' error, got: %v", err)
	}
}

// TestGetDid_Success verifies retrieving DID
func TestGetDid_Success(t *testing.T) {
	_, cleanup := utils.SetupTestConfig(t)
	defer cleanup()

	repo, err := NewSessionRepository()
	if err != nil {
		t.Fatalf("NewSessionRepository failed: %v", err)
	}

	ctx := context.Background()

	expectedDid := "did:plc:test123"
	session := &SessionModel{
		Handle:     "test.bsky.social",
		Token:      "access_token|refresh_token",
		ServiceURL: "https://bsky.social",
		IsValid:    true,
	}
	session.SetID(expectedDid)

	err = repo.Save(ctx, session)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	did, err := repo.GetDid(ctx)
	if err != nil {
		t.Fatalf("GetDid failed: %v", err)
	}

	if did != expectedDid {
		t.Errorf("expected DID %s, got %s", expectedDid, did)
	}
}

// TestGetHandle_NoSession verifies GetHandle errors when no session exists
func TestGetHandle_NoSession(t *testing.T) {
	_, cleanup := utils.SetupTestConfig(t)
	defer cleanup()

	repo, err := NewSessionRepository()
	if err != nil {
		t.Fatalf("NewSessionRepository failed: %v", err)
	}

	ctx := context.Background()
	_, err = repo.GetHandle(ctx)
	if err == nil {
		t.Error("expected error for no session, got nil")
	}

	if !strings.Contains(err.Error(), "no active session") {
		t.Errorf("expected 'no active session' error, got: %v", err)
	}
}

// TestGetHandle_Success verifies retrieving handle
func TestGetHandle_Success(t *testing.T) {
	_, cleanup := utils.SetupTestConfig(t)
	defer cleanup()

	repo, err := NewSessionRepository()
	if err != nil {
		t.Fatalf("NewSessionRepository failed: %v", err)
	}

	ctx := context.Background()

	expectedHandle := "test.bsky.social"
	session := &SessionModel{
		Handle:     expectedHandle,
		Token:      "access_token|refresh_token",
		ServiceURL: "https://bsky.social",
		IsValid:    true,
	}
	session.SetID("did:plc:test123")

	err = repo.Save(ctx, session)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	handle, err := repo.GetHandle(ctx)
	if err != nil {
		t.Fatalf("GetHandle failed: %v", err)
	}

	if handle != expectedHandle {
		t.Errorf("expected handle %s, got %s", expectedHandle, handle)
	}
}

// TestSplitToken_SinglePart verifies splitToken with single token
func TestSplitToken_SinglePart(t *testing.T) {
	token := "single_token"
	parts := splitToken(token)

	if len(parts) != 1 {
		t.Errorf("expected 1 part, got %d", len(parts))
	}

	if parts[0] != "single_token" {
		t.Errorf("expected 'single_token', got %s", parts[0])
	}
}

// TestSplitToken_TwoParts verifies splitToken with access and refresh tokens
func TestSplitToken_TwoParts(t *testing.T) {
	token := "access_token|refresh_token"
	parts := splitToken(token)

	if len(parts) != 2 {
		t.Fatalf("expected 2 parts, got %d", len(parts))
	}

	if parts[0] != "access_token" {
		t.Errorf("expected 'access_token', got %s", parts[0])
	}

	if parts[1] != "refresh_token" {
		t.Errorf("expected 'refresh_token', got %s", parts[1])
	}
}

// TestSplitToken_EmptyString verifies splitToken with empty string
func TestSplitToken_EmptyString(t *testing.T) {
	token := ""
	parts := splitToken(token)

	if len(parts) != 0 {
		t.Errorf("expected 0 parts, got %d", len(parts))
	}
}

// TestSplitToken_MultipleSeparators verifies splitToken with multiple separators
func TestSplitToken_MultipleSeparators(t *testing.T) {
	token := "part1|part2|part3"
	parts := splitToken(token)

	if len(parts) != 3 {
		t.Fatalf("expected 3 parts, got %d", len(parts))
	}

	if parts[0] != "part1" {
		t.Errorf("expected 'part1', got %s", parts[0])
	}
	if parts[1] != "part2" {
		t.Errorf("expected 'part2', got %s", parts[1])
	}
	if parts[2] != "part3" {
		t.Errorf("expected 'part3', got %s", parts[2])
	}
}

// TestSave_WithSingleToken verifies saving session with single token (no separator)
func TestSave_WithSingleToken(t *testing.T) {
	_, cleanup := utils.SetupTestConfig(t)
	defer cleanup()

	repo, err := NewSessionRepository()
	if err != nil {
		t.Fatalf("NewSessionRepository failed: %v", err)
	}

	ctx := context.Background()

	session := &SessionModel{
		Handle:     "test.bsky.social",
		Token:      "single_access_token",
		ServiceURL: "https://bsky.social",
		IsValid:    true,
	}
	session.SetID("did:plc:test123")

	err = repo.Save(ctx, session)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	retrieved, err := repo.Get(ctx, session.ID())
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if retrieved == nil {
		t.Fatal("expected non-nil session")
	}
}
