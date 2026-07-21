package events

import (
	"errors"
	"time"
)

// Common repository errors. Their text values are the stable external error
// identifiers used by the HTTP layer in later slices.
var (
	ErrSlugInvalid = errors.New("slug_invalid")
	ErrSlugTaken   = errors.New("slug_taken")
	ErrNotFound    = errors.New("not_found")
)

// Event is the domain model for a gathering and its gallery.
type Event struct {
	ID               string
	Slug             string
	Title            string
	Description      *string
	StartsAt         time.Time
	UploadClosesAt   time.Time
	GalleryExpiresAt time.Time
	IsClosed         bool
	CreatedAt        time.Time
}

// CreateParams is the input for creating an event.
// Slug is optional; when nil the repository generates one from Title.
type CreateParams struct {
	Slug             *string
	Title            string
	Description      *string
	StartsAt         time.Time
	UploadClosesAt   time.Time
	GalleryExpiresAt time.Time
}
