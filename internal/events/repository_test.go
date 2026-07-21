package events_test

import (
	"context"
	"testing"
	"time"

	"github.com/aadielpr/unnamed/internal/events"
	"github.com/aadielpr/unnamed/internal/testsupport"
	"github.com/stretchr/testify/require"
)

func TestRepository_Create(t *testing.T) {
	ctx, repo := setupRepo(t)

	params := newCreateParams(t, "Sarah's Birthday")
	params.Description = ptr("A fun night out.")

	event, token, err := repo.Create(ctx, params)
	require.NoError(t, err)
	require.NotEmpty(t, event.ID)
	require.Equal(t, "sarah-s-birthday", event.Slug)
	require.Equal(t, params.Title, event.Title)
	require.Equal(t, *params.Description, *event.Description)
	require.False(t, event.IsClosed)
	require.NotEmpty(t, token)

}

func TestRepository_Create_WithUserSlug(t *testing.T) {
	ctx, repo := setupRepo(t)

	params := newCreateParams(t, "Event")
	params.Slug = ptr("custom-slug")

	event, _, err := repo.Create(ctx, params)
	require.NoError(t, err)
	require.Equal(t, "custom-slug", event.Slug)
}

func TestRepository_Create_InvalidUserSlug(t *testing.T) {
	ctx, repo := setupRepo(t)

	params := newCreateParams(t, "Event")
	params.Slug = ptr("Invalid_Slug")

	_, _, err := repo.Create(ctx, params)
	require.ErrorIs(t, err, events.ErrSlugInvalid)
}

func TestRepository_Create_SlugTaken(t *testing.T) {
	ctx, repo := setupRepo(t)

	base := newCreateParams(t, "First")
	base.Slug = ptr("taken-slug")
	_, _, err := repo.Create(ctx, base)
	require.NoError(t, err)

	_, _, err = repo.Create(ctx, base)
	require.ErrorIs(t, err, events.ErrSlugTaken)
}

func TestRepository_Create_AutoSlugCollision(t *testing.T) {
	ctx, repo := setupRepo(t)

	base := newCreateParams(t, "Collision Party")

	first, _, err := repo.Create(ctx, base)
	require.NoError(t, err)
	require.Equal(t, "collision-party", first.Slug)

	second, _, err := repo.Create(ctx, base)
	require.NoError(t, err)
	require.Equal(t, "collision-party-2", second.Slug)

	third, _, err := repo.Create(ctx, base)
	require.NoError(t, err)
	require.Equal(t, "collision-party-3", third.Slug)
}

func TestRepository_GetBySlug(t *testing.T) {
	ctx, repo := setupRepo(t)

	params := newCreateParams(t, "Get Me")
	params.Slug = ptr("get-me")
	created, _, err := repo.Create(ctx, params)
	require.NoError(t, err)

	got, err := repo.GetBySlug(ctx, "get-me")
	require.NoError(t, err)
	require.Equal(t, created.ID, got.ID)
	require.Equal(t, created.Slug, got.Slug)
	require.Equal(t, created.Title, got.Title)
}

func TestRepository_GetBySlug_NotFound(t *testing.T) {
	ctx, repo := setupRepo(t)

	_, err := repo.GetBySlug(ctx, "does-not-exist")
	require.ErrorIs(t, err, events.ErrNotFound)
}

func setupRepo(t *testing.T) (context.Context, *events.Repository) {
	t.Helper()
	ctx := context.Background()
	repo := newRepo(t)
	cleanupEvents(t)
	return ctx, repo
}

func newRepo(t *testing.T) *events.Repository {
	t.Helper()
	return events.NewRepository(testsupport.TestDB(t))
}

func cleanupEvents(t *testing.T) {
	t.Helper()
	d := testsupport.TestDB(t)
	_, err := d.Exec("DELETE FROM events")
	require.NoError(t, err)
}

func ptr(s string) *string {
	return &s
}

func newCreateParams(t *testing.T, title string) events.CreateParams {
	t.Helper()
	return events.CreateParams{
		Title:            title,
		StartsAt:         time.Now().Add(time.Hour),
		UploadClosesAt:   time.Now().Add(2 * time.Hour),
		GalleryExpiresAt: time.Now().Add(24 * time.Hour),
	}
}
