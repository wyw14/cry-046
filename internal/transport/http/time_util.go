package http

import (
	"fmt"
	"time"
)

// parseRFC3339 parses an RFC3339 timestamp and returns a time.Time.
func parseRFC3339(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, fmt.Errorf("timestamp must not be empty")
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid timestamp %q: %w", s, err)
	}
	return t, nil
}

// formatRFC3339 formats a time.Time as RFC3339.
func formatRFC3339(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}
