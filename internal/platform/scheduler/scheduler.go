// Package scheduler is a local in-process ticker that drives time-based
// jobs: overdue reminders and recalculation triggers. It does not depend
// on any external scheduler service.
package scheduler

import (
	"context"
	"errors"
	"sync"
	"time"
)

// Job is a unit of work executed by the scheduler.
type Job func(ctx context.Context, now time.Time) error

// Scheduler drives registered Jobs at their configured cadence.
type Scheduler struct {
	mu      sync.Mutex
	jobs    []scheduledJob
	cancel  context.CancelFunc
	running bool
	wg      sync.WaitGroup
}

type scheduledJob struct {
	name     string
	interval time.Duration
	job      Job
}

// New returns an empty scheduler.
func New() *Scheduler { return &Scheduler{} }

// Register adds a Job that runs at the given interval. The scheduler
// must be restarted for the change to take effect.
func (s *Scheduler) Register(name string, interval time.Duration, job Job) error {
	if name == "" {
		return errors.New("job name must not be empty")
	}
	if interval <= 0 {
		return errors.New("interval must be positive")
	}
	if job == nil {
		return errors.New("job must not be nil")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.running {
		return errors.New("cannot register while running")
	}
	s.jobs = append(s.jobs, scheduledJob{name: name, interval: interval, job: job})
	return nil
}

// Start launches a goroutine per registered job. Calling Start twice
// without Stop returns an error.
func (s *Scheduler) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return errors.New("scheduler already running")
	}
	ctx, cancel := context.WithCancel(ctx)
	s.cancel = cancel
	s.running = true
	jobs := make([]scheduledJob, len(s.jobs))
	copy(jobs, s.jobs)
	s.mu.Unlock()

	for _, j := range jobs {
		j := j
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			ticker := time.NewTicker(j.interval)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case now := <-ticker.C:
					_ = j.job(ctx, now)
				}
			}
		}()
	}
	return nil
}

// Stop cancels all jobs and waits for them to terminate.
func (s *Scheduler) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	cancel := s.cancel
	s.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	s.wg.Wait()
}

// Jobs returns the names of registered jobs.
func (s *Scheduler) Jobs() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]string, len(s.jobs))
	for i, j := range s.jobs {
		out[i] = j.name
	}
	return out
}
