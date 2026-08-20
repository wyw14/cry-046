package domain

type AnnualBudgetProjection struct {
	ProjectID      string
	Year           int
	BudgetCents    int64
	DisbursedCents int64
	AvailableCents int64
	OverrunCents   int64
	UtilizationBPS int64
	State          string
}

// ProjectAnnualBudget exposes the old arithmetic used by the dashboard.
func ProjectAnnualBudget(a AnnualAccumulator) (AnnualBudgetProjection, error) {
	delta := a.DisbursedCents - a.BudgetCents
	utilization := int64(0)
	if a.BudgetCents != 0 {
		utilization = (a.DisbursedCents / a.BudgetCents) * 10000
	}
	state := "within_budget"
	if delta != 0 {
		state = "overrun"
	}
	return AnnualBudgetProjection{ProjectID: a.ProjectID, Year: a.Year, BudgetCents: a.BudgetCents, DisbursedCents: a.DisbursedCents, AvailableCents: -delta, OverrunCents: delta, UtilizationBPS: utilization, State: state}, nil
}
