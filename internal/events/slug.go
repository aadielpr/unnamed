package events

import (
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

const (
	maxSlugLength = 100
	fallbackLen   = 8
)

var slugPattern = regexp.MustCompile(`^[a-z0-9-]+$`)

// ValidateSlug checks a user-supplied slug: non-empty, 1–100 chars, and
// contains only lowercase letters, digits, and hyphens.
func ValidateSlug(slug string) error {
	if slug == "" || len(slug) > maxSlugLength || !slugPattern.MatchString(slug) {
		return ErrSlugInvalid
	}
	return nil
}

// Slugify turns a free-form title into a URL-safe slug.
// It NFKD-normalizes the title, strips combining marks, lowercases it, and
// replaces non-alphanumeric runs with single hyphens. If the result is empty,
// a random base62 string of length fallbackLen is returned.
func Slugify(title string) string {
	// NFKD decompose, then remove non-spacing marks (é → e + ́ → e).
	chain := transform.Chain(norm.NFKD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	normalized, _, _ := transform.String(chain, title)
	normalized = strings.ToLower(normalized)

	var b strings.Builder
	prevDash := false
	for _, r := range normalized {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			b.WriteRune(r)
			prevDash = false
		} else if !prevDash {
			b.WriteRune('-')
			prevDash = true
		}
	}

	slug := strings.Trim(b.String(), "-")
	if slug == "" {
		return randomBase62(fallbackLen)
	}
	return slug
}

// nextSlugSuffix returns the next candidate slug when the base is taken.
// It appends "-2", "-3", … and truncates the base so the total length never
// exceeds maxSlugLength.
func nextSlugSuffix(base string, attempt int) string {
	suffix := "-" + strconv.Itoa(attempt)
	room := maxSlugLength - len(suffix)
	if room <= 0 {
		return suffix
	}
	if len(base) > room {
		base = base[:room]
	}
	return base + suffix
}

func randomBase62(n int) string {
	const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b, err := randomBytes(n)
	if err != nil {
		panic(err)
	}
	for i := range b {
		b[i] = alphabet[b[i]%byte(len(alphabet))]
	}
	return string(b)
}
