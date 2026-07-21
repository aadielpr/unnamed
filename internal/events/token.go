package events

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

const tokenBytes = 32

// GenerateAdminToken creates a new admin token, returns the plain bearer value
// (shown to the organizer once), and its SHA-256 digest (stored in the DB).
func GenerateAdminToken() (string, string, error) {
	b, err := randomBytes(tokenBytes)
	if err != nil {
		return "", "", fmt.Errorf("generate admin token: %w", err)
	}

	plain := base64.RawURLEncoding.EncodeToString(b)
	digest := sha256.Sum256([]byte(plain))
	hash := base64.RawURLEncoding.EncodeToString(digest[:])
	return plain, hash, nil
}

