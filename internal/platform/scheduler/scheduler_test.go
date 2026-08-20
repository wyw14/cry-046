package scheduler

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestSchedulerRegisterAndRun(t *testing.T) {
	s := New()
	var n atomic.Int64
	if err := s.Register("tick", 10*time.Millisecond, func(ctx context.Context, now time.Time) error {
		n.Add(1)
		return nil
	}); err != nil {
		t.Fatalf("register: %v", err)
	}
	if len(s.Jobs()) != 1 {
		t.Errorf("expected 1 job, got %d", len(s.Jobs()))
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := s.Start(ctx); err != nil {
		t.Fatalf("start: %v", err)
	}
	time.Sleep(55 * time.Millisecond)
	s.Stop()
	if n.Load() < 3 {
		t.Errorf("expected at least 3 ticks, got %d", n.Load())
	}
}

func TestSchedulerRegisterValidation(t *testing.T) {
	s := New()
	if err := s.Register("", time.Second, nil); err == nil {
		t.Error("expected error for empty name")
	}
	if err := s.Register("test", 0, nil); err == nil {
		t.Error("expected error for non-positive interval")
	}
	if err := s.Register("test", time.Second, nil); err == nil {
		t.Error("expected error for nil job")
	}
}

func TestSchedulerStartTwice(t *testing.T) {
	s := New()
	_ = s.Register("test", time.Second, func(ctx context.Context, now time.Time) error { return nil })
	ctx := context.Background()
	if err := s.Start(ctx); err != nil {
		t.Fatalf("first start: %v", err)
	}
	if err := s.Start(ctx); err == nil {
		t.Error("expected error for second start")
	}
	s.Stop()
}

func TestSchedulerStopWithoutStart(t *testing.T) {
	s := New()
	s.Stop() // should not panic
}
