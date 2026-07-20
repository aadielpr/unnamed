package migrations_test

import (
	"database/sql"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
)

// TestMigrations_RoundTrip satisfies issue #7 AC #3: "A migration can be applied
// and rolled back round-trip (tool proven, not just present)." It runs goose
// against a real test Postgres — up (table appears), down (table gone), up again
// (re-applyable) — so migrations are exercised, not just present on disk.
func TestMigrations_RoundTrip(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set; migration round-trip needs a real test Postgres")
	}

	dir := migrationDir(t)

	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	// Start from a clean slate so the round-trip is observable regardless of
	// whatever state the shared test DB was left in.
	resetDB(t, db)

	// Up — the events table appears.
	require.NoError(t, goose.Run("up", db, dir))
	require.True(t, tableExists(t, db, "events"),
		"events table should exist after goose up")

	// Down — the events table is removed.
	require.NoError(t, goose.Run("down", db, dir))
	require.False(t, tableExists(t, db, "events"),
		"events table should be gone after goose down")

	// Up again — the migration is re-applyable (a true round-trip).
	require.NoError(t, goose.Run("up", db, dir))
	require.True(t, tableExists(t, db, "events"),
		"events table should exist after the second goose up")

	// Leave the DB clean for any test that runs after this one.
	resetDB(t, db)
}

func tableExists(t *testing.T, db *sql.DB, name string) bool {
	t.Helper()
	var n int64
	err := db.QueryRow(`
		SELECT count(*) FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = $1`, name).Scan(&n)
	require.NoError(t, err)
	return n == 1
}

func resetDB(t *testing.T, db *sql.DB) {
	t.Helper()
	_, err := db.Exec(`DROP TABLE IF EXISTS events CASCADE`)
	require.NoError(t, err)
	_, err = db.Exec(`DROP TABLE IF EXISTS goose_db_version`)
	require.NoError(t, err)
}

// migrationDir is the absolute path of the directory holding this test file,
// which is also the directory holding the .sql migrations.
func migrationDir(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)
	return filepath.Dir(file)
}
