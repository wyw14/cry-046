// Package scheduler is the application service that wires platform
// scheduled jobs (overdue reminder, recalculation trigger) into the
// local scheduler adapter.
package scheduler

import (
	"context"
	"time"

	"github.com/welfare/settlement-resolver/internal/application"
	"github.com/welfare/settlement-resolver/internal/domain"
	"github.com/welfare/settlement-resolver/internal/platform/scheduler"
)

// Service registers platform jobs on the given scheduler.
type Service struct {
	workspace *application.WorkspaceApp
	users     *application.UsersApp
	sched     *scheduler.Scheduler
	tick      time.Duration
	reminder  time.Duration
}

// New constructs a scheduler Service.
func New(workspace *application.WorkspaceApp, users *application.UsersApp, sched *scheduler.Scheduler, tick, reminder time.Duration) *Service {
	return &Service{workspace: workspace, users: users, sched: sched, tick: tick, reminder: reminder}
}

// RegisterAll registers all scheduled jobs.
func (s *Service) RegisterAll(ctx context.Context, tenantID string) error {
	if err := s.sched.Register("overdue-reminder", s.reminder, func(ctx context.Context, now time.Time) error {
		users, _, err := s.users.List(ctx, application.ListQuery{TenantID: tenantID, PageSize: 1000})
		if err != nil {
			return err
		}
		assignees := make([]string, 0, len(users))
		for _, u := range users {
			if u.Role == domain.RoleAssignee || u.Role == domain.RoleReviewer || u.Role == domain.RoleAdmin {
				assignees = append(assignees, u.ID)
			}
		}
		_, err = s.workspace.RemindOverdue(ctx, tenantID, assignees)
		return err
	}); err != nil {
		return err
	}
	return nil
}
