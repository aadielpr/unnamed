package events_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/aadielpr/unnamed/internal/events"
	"github.com/stretchr/testify/require"
)

func TestValidateSlug_Accepted(t *testing.T) {
	cases := []string{"a", "abc", "sarah-bday", "event-2024", "a-b-c-123"}
	for _, c := range cases {
		err := events.ValidateSlug(c)
		require.NoError(t, err, "slug %q should be valid", c)
	}
}

func TestValidateSlug_Rejected(t *testing.T) {
	cases := []string{
		"",
		"ABC",
		"hello_world",
		"hello world",
		"hello.world",
		strings.Repeat("a", 101),
	}
	for _, c := range cases {
		err := events.ValidateSlug(c)
		require.ErrorIs(t, err, events.ErrSlugInvalid, "slug %q should be invalid", c)
	}
}

func TestValidateSlug_LeadingTrailingHyphensAllowed(t *testing.T) {
	// The spec regex is ^[a-z0-9-]+$; leading/trailing hyphens are valid.
	require.NoError(t, events.ValidateSlug("-leading"))
	require.NoError(t, events.ValidateSlug("trailing-"))
}

func TestSlugify(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"Sarah's Birthday", "sarah-s-birthday"},
		{"  Multiple   Spaces  ", "multiple-spaces"},
		{"Café Noir", "cafe-noir"},
		{"日本語イベント", ""},
		{"!@#$%", ""},
	}

	urlSafePattern := regexp.MustCompile(`^[a-zA-Z0-9-]+$`)
	for _, c := range cases {
		got := events.Slugify(c.input)
		if c.expected != "" {
			require.Equal(t, c.expected, got)
		}
		require.LessOrEqual(t, len(got), 100)
		require.Regexp(t, urlSafePattern, got)
	}
}

func TestSlugify_EmptyTitleReturnsBase62(t *testing.T) {
	slug := events.Slugify("!@#$")
	require.NotEmpty(t, slug)
	require.Len(t, slug, 8)
	require.Regexp(t, regexp.MustCompile(`^[a-zA-Z0-9]+$`), slug)
}
