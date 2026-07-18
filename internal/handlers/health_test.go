package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/aadielpr/unnamed/internal/config"
	"github.com/aadielpr/unnamed/internal/db"
	"github.com/aadielpr/unnamed/internal/server"
	"github.com/stretchr/testify/require"
)

func TestHealth_DBReachable(t *testing.T) {
	database := testDB(t)
	cfg := config.Config{Port: "8080", StaticDir: "../../web/dist"}
	srv := server.New(cfg, database)

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var body map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, "ok", body["status"])
	require.Equal(t, "reachable", body["db"])
}

func TestHealth_DBUnreachable(t *testing.T) {
	// Use an invalid DSN so Ping fails.
	database, err := db.New("postgres://invalid:invalid@localhost:1/none?sslmode=disable")
	require.NoError(t, err)
	defer database.Close()

	cfg := config.Config{Port: "8080", StaticDir: "../../web/dist"}
	srv := server.New(cfg, database)

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	require.Equal(t, http.StatusServiceUnavailable, rec.Code)

	var body map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, "degraded", body["status"])
	require.Equal(t, "unreachable", body["db"])
}

func testDB(t *testing.T) *db.DB {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://gbmnx@/eventlens_test?sslmode=disable&host=/tmp"
	}

	database, err := db.New(dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = database.Close() })

	// Ensure the events table exists for the health ping to be meaningful.
	_, err = database.Exec(`CREATE TABLE IF NOT EXISTS events (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		slug TEXT UNIQUE NOT NULL,
		title TEXT NOT NULL,
		description TEXT,
		starts_at TIMESTAMPTZ NOT NULL,
		upload_closes_at TIMESTAMPTZ NOT NULL,
		gallery_expires_at TIMESTAMPTZ NOT NULL,
		is_closed BOOLEAN NOT NULL DEFAULT FALSE,
		admin_token TEXT NOT NULL,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	)`)
	require.NoError(t, err)

	return database
}
