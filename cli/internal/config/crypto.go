package config

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
	"os"
)

// EncryptToken encrypts a plaintext token using AES-256-GCM with a key derived from SHA256.
// The encryption key is derived from either SKYCLI_SECRET env var or machine-specific identifier.
// Returns base64-encoded ciphertext with prepended nonce.
func EncryptToken(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	key, err := getDerivedKey()
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", &CryptoError{Op: "NewCipher", Err: err}
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", &CryptoError{Op: "NewGCM", Err: err}
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", &CryptoError{Op: "GenerateNonce", Err: err}
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptToken decrypts a base64-encoded token encrypted with EncryptToken.
// Returns the original plaintext token or an error if decryption fails.
func DecryptToken(encrypted string) (string, error) {
	if encrypted == "" {
		return "", nil
	}

	key, err := getDerivedKey()
	if err != nil {
		return "", err
	}

	ciphertext, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", &CryptoError{Op: "DecodeBase64", Err: err}
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", &CryptoError{Op: "NewCipher", Err: err}
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", &CryptoError{Op: "NewGCM", Err: err}
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", &CryptoError{Op: "DecryptToken", Err: errors.New("ciphertext too short")}
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", &CryptoError{Op: "Decrypt", Err: err}
	}

	return string(plaintext), nil
}

// getDerivedKey derives a 32-byte AES key from either SKYCLI_SECRET env var
// or a combination of hostname and username. Uses SHA256 for key derivation.
func getDerivedKey() ([]byte, error) {
	secret := os.Getenv("SKYCLI_SECRET")
	if secret == "" {
		hostname, err := os.Hostname()
		if err != nil {
			return nil, &CryptoError{Op: "GetHostname", Err: err}
		}
		username := os.Getenv("USER")
		if username == "" {
			username = os.Getenv("USERNAME") // Windows
		}
		secret = hostname + ":" + username
	}

	hash := sha256.Sum256([]byte(secret))
	return hash[:], nil
}

// CryptoError represents an error that occurred during cryptographic operations
type CryptoError struct {
	Op  string
	Err error
}

func (e *CryptoError) Error() string {
	return "crypto." + e.Op + ": " + e.Err.Error()
}

func (e *CryptoError) Unwrap() error {
	return e.Err
}
