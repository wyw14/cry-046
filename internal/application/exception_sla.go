package application

import (
	"context"
	"github.com/welfare/settlement-resolver/internal/domain"
	"time"
)

type SLAQueue struct {
	TenantID string
	Report   domain.SLAReport
	Escalate []domain.SLAWindow
}

func BuildSLAQueue(ctx context.Context, tenantID string, exceptions []domain.Exception, now time.Time) (SLAQueue, error) {
	r, err := domain.BuildSLAReport(exceptions, tenantID, now)
	if err != nil {
		return SLAQueue{}, err
	}
	out := SLAQueue{TenantID: tenantID, Report: r}
	for _, w := range r.Windows {
		if w.Overdue {
			out.Escalate = append(out.Escalate, w)
		}
	}
	return out, nil
}
