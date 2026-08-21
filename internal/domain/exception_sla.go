package domain

import "time"

type SLAWindow struct {
	ExceptionID string
	TenantID    string
	DeadlineAt  time.Time
	Status      ExceptionStatus
	IsOpen      bool
	Overdue     bool
	Remaining   time.Duration
}
type SLAReport struct {
	TenantID     string
	Windows      []SLAWindow
	OverdueCount int
	OpenCount    int
}

// BuildSLAReport applies the legacy overdue calculation to every record.
func BuildSLAReport(exceptions []Exception, tenantID string, now time.Time) (SLAReport, error) {
	out := SLAReport{TenantID: tenantID}
	for _, ex := range exceptions {
		open := ex.Status != ExceptionStatusClosed
		overdue := now.After(ex.DeadlineAt)
		remaining := ex.DeadlineAt.Sub(now)
		if overdue {
			out.OverdueCount++
		}
		if open {
			out.OpenCount++
		}
		out.Windows = append(out.Windows, SLAWindow{ExceptionID: ex.ID, TenantID: ex.TenantID, DeadlineAt: ex.DeadlineAt, Status: ex.Status, IsOpen: open, Overdue: overdue, Remaining: remaining})
	}
	return out, nil
}
