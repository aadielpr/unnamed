package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aadielpr/unnamed/internal/config"
	"github.com/aadielpr/unnamed/internal/db"
	"github.com/aadielpr/unnamed/internal/server"
	"github.com/aadielpr/unnamed/internal/testsupport"
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
	return testsupport.TestDB(t)
}
