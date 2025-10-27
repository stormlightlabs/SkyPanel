package utils

import (
	"bytes"
	"database/sql"
	"io"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// CaptureOutput captures stdout during function execution
func CaptureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

// NewTestDB creates an in-memory SQLite database for testing.
// Returns the database connection and a cleanup function.
// The cleanup function should be called with defer to ensure proper cleanup.
func NewTestDB(t *testing.T) (*sql.DB, func()) {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory database: %v", err)
	}

	cleanup := func() {
		if err := db.Close(); err != nil {
			t.Errorf("failed to close test database: %v", err)
		}
	}

	return db, cleanup
}

// setupTestConfig creates a test config directory and returns cleanup function
func SetupTestConfig(t *testing.T) (string, func()) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")

	os.Setenv("HOME", tmpDir)

	configDir := filepath.Join(tmpDir, ".skycli")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create test config dir: %v", err)
	}

	cleanup := func() {
		os.Setenv("HOME", originalHome)
	}

	return configDir, cleanup
}
