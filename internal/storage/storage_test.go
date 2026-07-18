package storage_test

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/aadielpr/unnamed/internal/storage"
	"github.com/stretchr/testify/require"
)

func testStore(t *testing.T) storage.Store {
	t.Helper()

	cfg := storage.Config{
		Endpoint:        getEnv("TEST_STORAGE_ENDPOINT", "http://localhost:9000"),
		Region:          getEnv("TEST_STORAGE_REGION", "us-east-1"),
		Bucket:          getEnv("TEST_STORAGE_BUCKET", "eventlens-test"),
		AccessKeyID:     getEnv("TEST_STORAGE_ACCESS_KEY_ID", "minioadmin"),
		SecretAccessKey: getEnv("TEST_STORAGE_SECRET_ACCESS_KEY", "minioadmin"),
		UsePathStyle:    true,
		PublicURLBase:   getEnv("TEST_STORAGE_ENDPOINT", "http://localhost:9000"),
	}

	store, err := storage.NewS3Store(cfg)
	require.NoError(t, err)

	createBucket(t, cfg)
	return store
}

func TestS3Store_PutGetDelete(t *testing.T) {
	store := testStore(t)
	ctx := context.Background()
	key := "hello.txt"
	content := []byte("hello from the storage test")

	exists, err := store.Exists(ctx, key)
	require.NoError(t, err)
	require.False(t, exists)

	err = store.Put(ctx, key, bytes.NewReader(content), int64(len(content)), "text/plain")
	require.NoError(t, err)

	exists, err = store.Exists(ctx, key)
	require.NoError(t, err)
	require.True(t, exists)

	reader, err := store.Get(ctx, key)
	require.NoError(t, err)
	defer reader.Close()

	got, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.Equal(t, content, got)

	url := store.URL(key)
	require.True(t, strings.Contains(url, key), "public URL should contain the key")

	err = store.Delete(ctx, key)
	require.NoError(t, err)

	exists, err = store.Exists(ctx, key)
	require.NoError(t, err)
	require.False(t, exists)
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}
