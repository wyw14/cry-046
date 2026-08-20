package clock

import (
	"testing"
	"time"
)

func TestSystemNow(t *testing.T) {
	s := System{}
	before := time.Now()
	got := s.Now()
	after := time.Now()
	if got.Before(before) || got.After(after) {
		t.Errorf("System.Now() = %v, expected between %v and %v", got, before, after)
	}
}

func TestFixedNow(t *testing.T) {
	want := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	f := Fixed{T: want}
	if got := f.Now(); !got.Equal(want) {
		t.Errorf("Fixed.Now() = %v, want %v", got, want)
	}
}
