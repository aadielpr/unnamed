// Package testsupport holds helpers shared by integration tests.
package testsupport

import (
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"

	"github.com/aadielpr/unnamed/internal/db"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
)

var migrateOnce sync.Once

// TestDB returns a *db.DB connected to TEST_DATABASE_URL, skipping the test
// when that var isn't set — there is no portable default because a real test
// Postgres is required (PRD testing decision), not a per-developer socket.
//
// The schema is applied once per process from the real goose migrations, so
// tests run against the actual SQL rather than a hand-written stand-in.
func TestDB(t *testing.T) *db.DB {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set; DB-backed tests need a real test Postgres")
	}

	d, err := db.New(dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = d.Close() })

	migrateOnce.Do(func() {
		require.NoError(t, goose.Run("up", d.DB, migrationsDir()))
	})

	return d
}

// migrationsDir returns the absolute path to the repo's migrations directory.
// This file lives at internal/testsupport/db.go, so the repo root is ../..
func migrationsDir() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		panic("runtime.Caller failed in testsupport")
	}
	return filepath.Join(filepath.Dir(file), "..", "..", "migrations")
}
