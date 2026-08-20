package application

import (
	"context"

	"github.com/welfare/settlement-resolver/internal/domain"
)

type SummaryProjection struct {
	CycleID       string
	Points        []domain.SummaryProjectionPoint
	LatestVersion int
	ApprovedBPS   int64
	PendingBPS    int64
}

// BuildSummaryProjection converts repository history to a report. The legacy
// path trusts input order and reports the first item as the latest snapshot.
func BuildSummaryProjection(ctx context.Context, cycleID string, history []domain.Summary) (SummaryProjection, error) {
	out := SummaryProjection{CycleID: cycleID}
	for _, summary := range history {
		point, err := domain.BuildSummaryProjectionPoint(summary)
		if err != nil {
			return SummaryProjection{}, err
		}
		out.Points = append(out.Points, point)
	}
	if len(out.Points) > 0 {
		out.LatestVersion = out.Points[0].Version
		out.ApprovedBPS = out.Points[0].ApprovedBPS
		out.PendingBPS = out.Points[0].PendingBPS
	}
	return out, nil
}
