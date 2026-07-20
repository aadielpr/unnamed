package server_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/aadielpr/unnamed/internal/config"
	"github.com/aadielpr/unnamed/internal/db"
	"github.com/aadielpr/unnamed/internal/server"
	"github.com/aadielpr/unnamed/internal/testsupport"
	"github.com/stretchr/testify/require"
)

func TestServer_ServesIndexForRoot(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "index.html"), []byte("<!doctype html><html></html>"), 0644))

	database := testDB(t)
	cfg := config.Config{Port: "8080", StaticDir: root}
	srv := server.New(cfg, database)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "<!doctype html>")
}

func TestServer_ServesSPAFallback(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "index.html"), []byte("<html><body>SPA</body></html>"), 0644))

	database := testDB(t)
	cfg := config.Config{Port: "8080", StaticDir: root}
	srv := server.New(cfg, database)

	req := httptest.NewRequest(http.MethodGet, "/e/some-event", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "SPA")
}

// An unmatched /api/* path is a 404, not the SPA fallback (decision #4:
// /api is the only API prefix and must not masquerade as the SPA).
func TestServer_APIUnknownIsNotFound(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "index.html"), []byte("<html>SPA</html>"), 0644))

	database := testDB(t)
	cfg := config.Config{Port: "8080", StaticDir: root}
	srv := server.New(cfg, database)

	req := httptest.NewRequest(http.MethodGet, "/api/no-such-route", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
}

func testDB(t *testing.T) *db.DB {
	t.Helper()
	return testsupport.TestDB(t)
}
