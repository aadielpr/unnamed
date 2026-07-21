package events_test

import (
	"crypto/sha256"
	"encoding/base64"
	"strings"
	"testing"

	"github.com/aadielpr/unnamed/internal/events"
	"github.com/stretchr/testify/require"
)

func TestGenerateAdminToken(t *testing.T) {
	plain, hash, err := events.GenerateAdminToken()
	require.NoError(t, err)
	require.NotEmpty(t, plain)
	require.NotEmpty(t, hash)

	// Base64url raw alphabet contains no padding and no +/ chars.
	require.False(t, strings.ContainsAny(plain, "+/="))
	require.False(t, strings.ContainsAny(hash, "+/="))

	// Decode the plain token to verify it is 32 random bytes.
	decoded, err := base64.RawURLEncoding.DecodeString(plain)
	require.NoError(t, err)
	require.Len(t, decoded, 32)

	// The stored hash is the SHA-256 digest of the plain token.
	expected := sha256.Sum256([]byte(plain))
	expectedHash := base64.RawURLEncoding.EncodeToString(expected[:])
	require.Equal(t, expectedHash, hash)
}

func TestGenerateAdminToken_Unique(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		plain, _, err := events.GenerateAdminToken()
		require.NoError(t, err)
		require.False(t, seen[plain], "token should be unique")
		seen[plain] = true
	}
}

