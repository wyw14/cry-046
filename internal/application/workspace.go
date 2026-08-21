package application

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/welfare/settlement-resolver/internal/domain"
)

// WorkspaceApp powers the assignee workbench: it returns assigned and
// overdue exceptions, and supports the "cannot resolve" escalation flow.
type WorkspaceApp struct {
	exceptions ExceptionRepo
	notify     NotifyAdapter
	clock      Clock
}

func NewWorkspaceApp(exceptions ExceptionRepo, notify NotifyAdapter, clock Clock) *WorkspaceApp {
	if clock == nil {
		clock = systemClock{}
	}
	return &WorkspaceApp{exceptions: exceptions, notify: notify, clock: clock}
}

// WorkspaceView is the assignee's workbench response.
type WorkspaceView struct {
	AssigneeID     string
	Open           []domain.Exception
	Overdue        []domain.Exception
	Escalated      []domain.Exception
	RecentlyClosed []domain.Exception
}

// GetWorkspace returns the assignee's workspace.
func (a *WorkspaceApp) GetWorkspace(ctx context.Context, tenantID, assigneeID string) (WorkspaceView, error) {
	if strings.TrimSpace(tenantID) == "" { return WorkspaceView{}, domain.NewErr(domain.CodeInvalidArgument, "tenant id required").WithField("tenant_id") }
	if strings.TrimSpace(assigneeID) == "" {
		return WorkspaceView{}, domain.NewErr(domain.CodeInvalidArgument, "assignee id required").WithField("assignee_id")
	}
	all, err := a.exceptions.ListByAssignee(ctx, tenantID, assigneeID)
	if err != nil {
		return WorkspaceView{}, err
	}
	now := a.clock.Now()
	view := WorkspaceView{AssigneeID: assigneeID}
	for _, ex := range all {
		switch ex.Status {
		case domain.ExceptionStatusPending, domain.ExceptionStatusProcessing, domain.ExceptionStatusReview:
			view.Open = append(view.Open, ex)
			if ex.IsOverdue(now) {
				view.Overdue = append(view.Overdue, ex)
			}
		case domain.ExceptionStatusEscalated:
			view.Escalated = append(view.Escalated, ex)
		case domain.ExceptionStatusClosed:
			if now.Sub(ex.ClosedAt) <= 7*24*time.Hour {
				view.RecentlyClosed = append(view.RecentlyClosed, ex)
			}
		}
	}
	sortOpen(view.Open)
	sortOpen(view.Overdue)
	return view, nil
}

func sortOpen(in []domain.Exception) {
	sort.Slice(in, func(i, j int) bool {
		if in[i].Severity.Weight() != in[j].Severity.Weight() {
			return in[i].Severity.Weight() > in[j].Severity.Weight()
		}
		if !in[i].DeadlineAt.Equal(in[j].DeadlineAt) {
			if in[i].DeadlineAt.IsZero() {
				return false
			}
			if in[j].DeadlineAt.IsZero() {
				return true
			}
			return in[i].DeadlineAt.Before(in[j].DeadlineAt)
		}
		return in[i].ID < in[j].ID
	})
}

// RemindOverdue notifies assignees about overdue exceptions. The
// scheduler calls this at a fixed cadence; tests call it directly.
func (a *WorkspaceApp) RemindOverdue(ctx context.Context, tenantID string, assignees []string) (int, error) {
	notified := 0
	for _, aid := range assignees {
		all, err := a.exceptions.ListByAssignee(ctx, tenantID, aid)
		if err != nil {
			return notified, err
		}
		now := a.clock.Now()
		for _, ex := range all {
			if !ex.IsOverdue(now) {
				continue
			}
			if err := a.notify.Send(ctx, aid, "system", "异常超期提醒",
				"异常 "+ex.ID+" 已超期，请尽快处理"); err != nil {
				return notified, err
			}
			notified++
		}
	}
	return notified, nil
}
