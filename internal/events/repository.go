package events

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/aadielpr/unnamed/internal/db"
)

// Repository provides persistence operations for events.
type Repository struct {
	db *db.DB
}

// Columns returned by repository queries, in the order scanEvent expects.
const eventColumns = "id, slug, title, description, starts_at, upload_closes_at, gallery_expires_at, is_closed, created_at"

// NewRepository returns a Repository backed by database.
func NewRepository(database *db.DB) *Repository {
	return &Repository{db: database}
}

// Create inserts a new event. It validates/generates the slug, creates an admin
// token, and returns the stored event together with the plain admin token.
// The caller must persist the plain token; it cannot be recovered later.
func (r *Repository) Create(ctx context.Context, params CreateParams) (Event, string, error) {
	var slug string
	var supplied bool
	if params.Slug != nil {
		slug = *params.Slug
		supplied = true
		if err := ValidateSlug(slug); err != nil {
			return Event{}, "", err
		}
	} else {
		slug = Slugify(params.Title)
	}

	taken, err := r.slugExists(ctx, slug)
	if err != nil {
		return Event{}, "", err
	}

	if taken {
		if supplied {
			return Event{}, "", ErrSlugTaken
		}
		slug, err = r.resolveCollision(ctx, slug)
		if err != nil {
			return Event{}, "", err
		}
	}

	plainToken, hash, err := GenerateAdminToken()
	if err != nil {
		return Event{}, "", err
	}

	row := r.db.QueryRowContext(ctx, `
		INSERT INTO events (slug, title, description, starts_at, upload_closes_at, gallery_expires_at, admin_token_hash)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING `+eventColumns+`
	`, slug, params.Title, params.Description, params.StartsAt, params.UploadClosesAt, params.GalleryExpiresAt, hash)

	event, err := scanEvent(row)
	if err != nil {
		return Event{}, "", fmt.Errorf("insert event: %w", err)
	}

	return event, plainToken, nil
}

// GetBySlug returns the event with the given slug, or ErrNotFound.
func (r *Repository) GetBySlug(ctx context.Context, slug string) (Event, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT `+eventColumns+`
		FROM events
		WHERE slug = $1
	`, slug)

	event, err := scanEvent(row)
	if errors.Is(err, sql.ErrNoRows) {
		return Event{}, ErrNotFound
	}
	if err != nil {
		return Event{}, fmt.Errorf("get event by slug: %w", err)
	}
	return event, nil
}

func (r *Repository) slugExists(ctx context.Context, slug string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM events WHERE slug = $1)`, slug).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check slug existence: %w", err)
	}
	return exists, nil
}

func (r *Repository) resolveCollision(ctx context.Context, base string) (string, error) {
	for attempt := 2; attempt <= 1000; attempt++ {
		candidate := nextSlugSuffix(base, attempt)
		exists, err := r.slugExists(ctx, candidate)
		if err != nil {
			return "", err
		}
		if !exists {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("could not resolve slug collision for %q", base)
}

func scanEvent(row *sql.Row) (Event, error) {
	var e Event
	err := row.Scan(
		&e.ID,
		&e.Slug,
		&e.Title,
		&e.Description,
		&e.StartsAt,
		&e.UploadClosesAt,
		&e.GalleryExpiresAt,
		&e.IsClosed,
		&e.CreatedAt,
	)
	return e, err
}
