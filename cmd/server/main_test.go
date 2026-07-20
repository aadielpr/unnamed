package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestLoadEnv_MissingFile: a missing .env must not be an error. The .env load
// is best-effort (ADR-0002): the app must start when no .env exists, as long as
// the environment already points at reachable services (e.g. a prod container).
func TestLoadEnv_MissingFile(t *testing.T) {
	dir := t.TempDir()
	restore := chdir(t, dir)
	defer restore()

	// Ensure no .env exists in the temp dir.
	_, err := os.Stat(".env")
	require.True(t, os.IsNotExist(err))

	require.NoError(t, loadEnv())
}

// TestLoadEnv_BrokenFile: a present-but-unreadable .env must surface an error,
// not be silently swallowed.
func TestLoadEnv_BrokenFile(t *testing.T) {
	dir := t.TempDir()
	restore := chdir(t, dir)
	defer restore()

	// A directory named ".env" opens but cannot be read back, so godotenv
	// returns a non-"not found" error.
	require.NoError(t, os.Mkdir(".env", 0o755))

	err := loadEnv()
	require.Error(t, err)
}

// chdir changes to dir and returns a restore func. Tests must not rely on cwd.
func chdir(t *testing.T, dir string) (restore func()) {
	t.Helper()
	wd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	return func() { _ = os.Chdir(wd) }
}
