package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/welfare/settlement-resolver/internal/platform/scheduler"
)

func TestScheduler_RegisterAndRun(t *testing.T) {
	s := scheduler.New()
	if err := s.Register("test-job", 10*time.Millisecond, func(ctx context.Context, now time.Time) error {
		return nil
	}); err != nil {
		t.Fatalf("register: %v", err)
	}
	if err := s.Start(context.Background()); err != nil {
		t.Fatalf("start: %v", err)
	}
	defer s.Stop()
	time.Sleep(50 * time.Millisecond)
}

func TestScheduler_DuplicateStart(t *testing.T) {
	s := scheduler.New()
	_ = s.Register("j", 10*time.Millisecond, func(ctx context.Context, now time.Time) error { return nil })
	if err := s.Start(context.Background()); err != nil {
		t.Fatalf("first start: %v", err)
	}
	defer s.Stop()
	if err := s.Start(context.Background()); err == nil {
		t.Fatal("expected error on duplicate start")
	}
}

func TestScheduler_RegisterValidation(t *testing.T) {
	s := scheduler.New()
	cases := []struct {
		name     string
		interval time.Duration
		job      scheduler.Job
	}{
		{"", 10 * time.Millisecond, func(ctx context.Context, now time.Time) error { return nil }},
		{"j", 0, func(ctx context.Context, now time.Time) error { return nil }},
		{"j", 10 * time.Millisecond, nil},
	}
	for i, tc := range cases {
		if err := s.Register(tc.name, tc.interval, tc.job); err == nil {
			t.Errorf("case %d: expected error", i)
		}
	}
}

func TestScheduler_RegisterWhileRunning(t *testing.T) {
	s := scheduler.New()
	_ = s.Register("j", 10*time.Millisecond, func(ctx context.Context, now time.Time) error { return nil })
	_ = s.Start(context.Background())
	defer s.Stop()
	if err := s.Register("j2", 10*time.Millisecond, func(ctx context.Context, now time.Time) error { return nil }); err == nil {
		t.Fatal("expected error registering while running")
	}
}

func TestScheduler_StopWithoutStart(t *testing.T) {
	s := scheduler.New()
	s.Stop() // must not panic
}

func TestScheduler_JobsListed(t *testing.T) {
	s := scheduler.New()
	_ = s.Register("a", 10*time.Millisecond, func(ctx context.Context, now time.Time) error { return nil })
	_ = s.Register("b", 10*time.Millisecond, func(ctx context.Context, now time.Time) error { return nil })
	got := s.Jobs()
	if len(got) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(got))
	}
}
