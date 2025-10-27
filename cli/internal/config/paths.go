package config

import (
	"os"
	"path/filepath"
	"runtime"
)

const appName = "skycli"

// GetConfigDir returns the platform-specific configuration directory path.
// On Unix-like systems: ~/.skycli
// On Windows: %APPDATA%/skycli
func GetConfigDir() (string, error) {
	var baseDir string

	if runtime.GOOS == "windows" {
		baseDir = os.Getenv("APPDATA")
		if baseDir == "" {
			return "", &PathError{Op: "GetConfigDir", Err: "APPDATA environment variable not set"}
		}
	} else {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", &PathError{Op: "GetConfigDir", Err: err.Error()}
		}
		baseDir = homeDir
	}

	if runtime.GOOS == "windows" {
		return filepath.Join(baseDir, appName), nil
	}
	return filepath.Join(baseDir, "."+appName), nil
}

// GetConfigFile returns the full path to the configuration file.
// Returns: ~/.skycli/.config.json (Unix) or %APPDATA%/skycli/.config.json (Windows)
func GetConfigFile() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, ".config.json"), nil
}

// GetCacheDB returns the full path to the SQLite cache database.
// Returns: ~/.skycli/cache.db (Unix) or %APPDATA%/skycli/cache.db (Windows)
func GetCacheDB() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "cache.db"), nil
}

// EnsureConfigDir creates the configuration directory if it doesn't exist.
// Sets permissions to 0700 (owner read/write/execute only) for security.
func EnsureConfigDir() error {
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(configDir, 0700); err != nil {
		return &PathError{Op: "EnsureConfigDir", Err: err.Error()}
	}

	return nil
}

// PathError represents an error that occurred during path operations
type PathError struct {
	Op  string
	Err string
}

func (e *PathError) Error() string {
	return e.Op + ": " + e.Err
}
