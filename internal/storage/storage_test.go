package storage

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestS3Store_PutGetDelete(t *testing.T) {
	store := testStore(t)
	ctx := context.Background()
	key := "hello.txt"
	content := []byte("hello from the storage test")

	err := store.Put(ctx, key, bytes.NewReader(content), int64(len(content)), "text/plain")
	require.NoError(t, err)

	reader, err := store.Get(ctx, key)
	require.NoError(t, err)
	defer reader.Close()

	got, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.Equal(t, content, got)

	require.NoError(t, store.Delete(ctx, key))

	// Delete took effect: a Get afterwards surfaces an error.
	_, err = store.Get(ctx, key)
	require.Error(t, err)
}

func testStore(t *testing.T) Store {
	t.Helper()

	cfg := Config{
		Endpoint:        getEnv("TEST_STORAGE_ENDPOINT", "http://localhost:9000"),
		Region:          getEnv("TEST_STORAGE_REGION", "us-east-1"),
		Bucket:          getEnv("TEST_STORAGE_BUCKET", "eventlens-test"),
		AccessKeyID:     getEnv("TEST_STORAGE_ACCESS_KEY_ID", "minioadmin"),
		SecretAccessKey: getEnv("TEST_STORAGE_SECRET_ACCESS_KEY", "minioadmin"),
		UsePathStyle:    true,
	}

	createBucket(t, cfg)

	store, err := NewS3Store(cfg)
	require.NoError(t, err)
	return store
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}
