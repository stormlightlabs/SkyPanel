package imports

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseEnvFile(t *testing.T) {
	t.Run("parses basic key-value pairs", func(t *testing.T) {
		tmpDir := t.TempDir()
		envPath := filepath.Join(tmpDir, ".env")

		content := `BLUESKY_HANDLE=test.bsky.social
BLUESKY_PASSWORD=secret123`

		if err := os.WriteFile(envPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		env, err := ParseEnvFile(envPath)
		if err != nil {
			t.Fatalf("ParseEnvFile failed: %v", err)
		}

		if env["BLUESKY_HANDLE"] != "test.bsky.social" {
			t.Errorf("expected BLUESKY_HANDLE=test.bsky.social, got %s", env["BLUESKY_HANDLE"])
		}

		if env["BLUESKY_PASSWORD"] != "secret123" {
			t.Errorf("expected BLUESKY_PASSWORD=secret123, got %s", env["BLUESKY_PASSWORD"])
		}
	})

	t.Run("ignores comments and empty lines", func(t *testing.T) {
		tmpDir := t.TempDir()
		envPath := filepath.Join(tmpDir, ".env")

		content := `# This is a comment
BLUESKY_HANDLE=test.bsky.social

# Another comment
BLUESKY_PASSWORD=secret123

`

		if err := os.WriteFile(envPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		env, err := ParseEnvFile(envPath)
		if err != nil {
			t.Fatalf("ParseEnvFile failed: %v", err)
		}

		if len(env) != 2 {
			t.Errorf("expected 2 entries, got %d", len(env))
		}

		if env["BLUESKY_HANDLE"] != "test.bsky.social" {
			t.Errorf("expected BLUESKY_HANDLE=test.bsky.social, got %s", env["BLUESKY_HANDLE"])
		}

		if env["BLUESKY_PASSWORD"] != "secret123" {
			t.Errorf("expected BLUESKY_PASSWORD=secret123, got %s", env["BLUESKY_PASSWORD"])
		}
	})

	t.Run("handles whitespace around keys and values", func(t *testing.T) {
		tmpDir := t.TempDir()
		envPath := filepath.Join(tmpDir, ".env")

		content := `  BLUESKY_HANDLE  =  test.bsky.social
BLUESKY_PASSWORD=  secret123`

		if err := os.WriteFile(envPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		env, err := ParseEnvFile(envPath)
		if err != nil {
			t.Fatalf("ParseEnvFile failed: %v", err)
		}

		if env["BLUESKY_HANDLE"] != "test.bsky.social" {
			t.Errorf("expected BLUESKY_HANDLE=test.bsky.social, got %s", env["BLUESKY_HANDLE"])
		}

		if env["BLUESKY_PASSWORD"] != "secret123" {
			t.Errorf("expected BLUESKY_PASSWORD=secret123, got %s", env["BLUESKY_PASSWORD"])
		}
	})

	t.Run("handles values containing equals signs", func(t *testing.T) {
		tmpDir := t.TempDir()
		envPath := filepath.Join(tmpDir, ".env")

		content := `API_URL=https://api.example.com?key=value&other=thing
TOKEN=abc=def=ghi`

		if err := os.WriteFile(envPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		env, err := ParseEnvFile(envPath)
		if err != nil {
			t.Fatalf("ParseEnvFile failed: %v", err)
		}

		if env["API_URL"] != "https://api.example.com?key=value&other=thing" {
			t.Errorf("expected API_URL=https://api.example.com?key=value&other=thing, got %s", env["API_URL"])
		}

		if env["TOKEN"] != "abc=def=ghi" {
			t.Errorf("expected TOKEN=abc=def=ghi, got %s", env["TOKEN"])
		}
	})

	t.Run("ignores malformed lines", func(t *testing.T) {
		tmpDir := t.TempDir()
		envPath := filepath.Join(tmpDir, ".env")

		content := `BLUESKY_HANDLE=test.bsky.social
INVALID_LINE_NO_EQUALS
BLUESKY_PASSWORD=secret123
ANOTHER_INVALID
`

		if err := os.WriteFile(envPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		env, err := ParseEnvFile(envPath)
		if err != nil {
			t.Fatalf("ParseEnvFile failed: %v", err)
		}

		if len(env) != 2 {
			t.Errorf("expected 2 entries, got %d", len(env))
		}

		if env["BLUESKY_HANDLE"] != "test.bsky.social" {
			t.Errorf("expected BLUESKY_HANDLE=test.bsky.social, got %s", env["BLUESKY_HANDLE"])
		}

		if env["BLUESKY_PASSWORD"] != "secret123" {
			t.Errorf("expected BLUESKY_PASSWORD=secret123, got %s", env["BLUESKY_PASSWORD"])
		}
	})

	t.Run("returns error for non-existent file", func(t *testing.T) {
		tmpDir := t.TempDir()
		envPath := filepath.Join(tmpDir, "nonexistent.env")

		_, err := ParseEnvFile(envPath)
		if err == nil {
			t.Error("expected error for non-existent file, got nil")
		}
	})

	t.Run("handles relative paths", func(t *testing.T) {
		tmpDir := t.TempDir()
		originalWd, err := os.Getwd()
		if err != nil {
			t.Fatalf("failed to get working directory: %v", err)
		}
		defer os.Chdir(originalWd)

		if err := os.Chdir(tmpDir); err != nil {
			t.Fatalf("failed to change directory: %v", err)
		}

		content := `BLUESKY_HANDLE=test.bsky.social
BLUESKY_PASSWORD=secret123`

		if err := os.WriteFile(".env", []byte(content), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		env, err := ParseEnvFile(".env")
		if err != nil {
			t.Fatalf("ParseEnvFile failed: %v", err)
		}

		if env["BLUESKY_HANDLE"] != "test.bsky.social" {
			t.Errorf("expected BLUESKY_HANDLE=test.bsky.social, got %s", env["BLUESKY_HANDLE"])
		}
	})

	t.Run("handles absolute paths", func(t *testing.T) {
		tmpDir := t.TempDir()
		envPath := filepath.Join(tmpDir, ".env")

		content := `BLUESKY_HANDLE=test.bsky.social
BLUESKY_PASSWORD=secret123`

		if err := os.WriteFile(envPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		absPath, err := filepath.Abs(envPath)
		if err != nil {
			t.Fatalf("failed to get absolute path: %v", err)
		}

		env, err := ParseEnvFile(absPath)
		if err != nil {
			t.Fatalf("ParseEnvFile failed: %v", err)
		}

		if env["BLUESKY_HANDLE"] != "test.bsky.social" {
			t.Errorf("expected BLUESKY_HANDLE=test.bsky.social, got %s", env["BLUESKY_HANDLE"])
		}
	})

	t.Run("handles empty file", func(t *testing.T) {
		tmpDir := t.TempDir()
		envPath := filepath.Join(tmpDir, ".env")

		if err := os.WriteFile(envPath, []byte(""), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		env, err := ParseEnvFile(envPath)
		if err != nil {
			t.Fatalf("ParseEnvFile failed: %v", err)
		}

		if len(env) != 0 {
			t.Errorf("expected empty map, got %d entries", len(env))
		}
	})
}
