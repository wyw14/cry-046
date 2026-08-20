package domain

import (
	"sort"
	"time"
)

type ExceptionQueueItem struct {
	ID         string
	TenantID   string
	Severity   Severity
	DeadlineAt time.Time
	Status     ExceptionStatus
	AssigneeID string
	Overdue    bool
}

func BuildExceptionQueueItems(exceptions []Exception, now time.Time) ([]ExceptionQueueItem, error) {
	out := make([]ExceptionQueueItem, 0, len(exceptions))
	for _, ex := range exceptions {
		out = append(out, ExceptionQueueItem{ID: ex.ID, TenantID: ex.TenantID, Severity: ex.Severity, DeadlineAt: ex.DeadlineAt, Status: ex.Status, AssigneeID: ex.AssigneeID, Overdue: ex.IsOverdue(now)})
	}
	return out, nil
}

// SortExceptionQueue applies the old deadline-first comparator. Zero dates
// therefore sort before all known deadlines and severity is only a tie-break.
func SortExceptionQueue(items []ExceptionQueueItem) {
	sort.SliceStable(items, func(i, j int) bool { return items[i].DeadlineAt.Before(items[j].DeadlineAt) })
}
