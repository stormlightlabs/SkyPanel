package config

import (
	"encoding/json"
	"errors"
	"os"
)

// Config represents the application configuration stored in ~/.skycli/.config.json
// Tokens are encrypted at rest using AES-256-GCM
type Config struct {
	Session *SessionConfig `json:"session,omitempty"`
}

// SessionConfig holds the current session information with encrypted tokens
type SessionConfig struct {
	Handle           string `json:"handle"`
	Did              string `json:"did"`
	ServiceURL       string `json:"serviceUrl"`
	EncryptedAccess  string `json:"encryptedAccessToken"`
	EncryptedRefresh string `json:"encryptedRefreshToken"`
	Email            string `json:"email,omitempty"`
}

// Load reads and decrypts the configuration from ~/.skycli/.config.json
// Returns a default empty config if the file doesn't exist
func Load() (*Config, error) {
	configPath, err := GetConfigFile()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &Config{}, nil
		}
		return nil, &ConfigError{Op: "ReadFile", Err: err}
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, &ConfigError{Op: "Unmarshal", Err: err}
	}

	return &cfg, nil
}

// Save encrypts tokens and persists the configuration to ~/.skycli/.config.json
// Creates the config directory if it doesn't exist
func (c *Config) Save() error {
	if err := EnsureConfigDir(); err != nil {
		return err
	}

	configPath, err := GetConfigFile()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return &ConfigError{Op: "Marshal", Err: err}
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return &ConfigError{Op: "WriteFile", Err: err}
	}

	return nil
}

// GetAccessToken decrypts and returns the access token
func (s *SessionConfig) GetAccessToken() (string, error) {
	if s == nil || s.EncryptedAccess == "" {
		return "", nil
	}
	return DecryptToken(s.EncryptedAccess)
}

// GetRefreshToken decrypts and returns the refresh token
func (s *SessionConfig) GetRefreshToken() (string, error) {
	if s == nil || s.EncryptedRefresh == "" {
		return "", nil
	}
	return DecryptToken(s.EncryptedRefresh)
}

// SetAccessToken encrypts and stores the access token
func (s *SessionConfig) SetAccessToken(token string) error {
	encrypted, err := EncryptToken(token)
	if err != nil {
		return err
	}
	s.EncryptedAccess = encrypted
	return nil
}

// SetRefreshToken encrypts and stores the refresh token
func (s *SessionConfig) SetRefreshToken(token string) error {
	encrypted, err := EncryptToken(token)
	if err != nil {
		return err
	}
	s.EncryptedRefresh = encrypted
	return nil
}

// ConfigError represents an error that occurred during config operations
type ConfigError struct {
	Op  string
	Err error
}

func (e *ConfigError) Error() string {
	return "config." + e.Op + ": " + e.Err.Error()
}

func (e *ConfigError) Unwrap() error {
	return e.Err
}
