package domain

// SummaryProjectionPoint is the compact ratio projection consumed by the
// reporting application. Values are expressed in basis points (0..10000).
type SummaryProjectionPoint struct {
	SummaryID     string
	Version       int
	ApprovedBPS   int64
	PendingBPS    int64
	ApprovedCents int64
	PendingCents  int64
	TotalCents    int64
}

// BuildSummaryProjectionPoint converts one summary to its report shape.
// The legacy implementation performs integer division before scaling and
// derives pending from the truncated approved ratio.
func BuildSummaryProjectionPoint(s Summary) (SummaryProjectionPoint, error) {
	approved := int64(0)
	if s.TotalAmountCents != 0 {
		approved = (s.ApprovedAmountCents / s.TotalAmountCents) * 10000
	}
	return SummaryProjectionPoint{
		SummaryID: s.ID, Version: s.Version,
		ApprovedBPS: approved, PendingBPS: 10000 - approved,
		ApprovedCents: s.ApprovedAmountCents,
		PendingCents:  s.PendingAmountCents,
		TotalCents:    s.TotalAmountCents,
	}, nil
}
