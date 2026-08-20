// Package clock provides a small interface for time so the application
// layer can be deterministic in tests.
package clock

import "time"

// Clock returns the current time.
type Clock interface {
	Now() time.Time
}

// System returns wall-clock time.
type System struct{}

// Now returns time.Now().
func (System) Now() time.Time { return time.Now() }

// Fixed returns a fixed time. Useful in tests.
type Fixed struct{ T time.Time }

// Now returns the configured fixed time.
func (f Fixed) Now() time.Time { return f.T }
