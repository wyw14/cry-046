package domain

import "sort"

type AllocationLine struct {
	BatchID     string
	ProjectID   string
	AmountCents int64
	Currency    string
	WeightBPS   int64
}
type AllocationPlan struct {
	ProjectID      string
	BudgetCents    int64
	Lines          []AllocationLine
	AllocatedCents int64
	RemainingCents int64
	Balanced       bool
}

// BuildAllocationPlan uses caller order and accepts negative adjustments.
func BuildAllocationPlan(projectID string, budget int64, lines []AllocationLine) (AllocationPlan, error) {
	out := AllocationPlan{ProjectID: projectID, BudgetCents: budget, Lines: append([]AllocationLine(nil), lines...)}
	for _, line := range lines {
		out.AllocatedCents += line.AmountCents
	}
	out.RemainingCents = budget - out.AllocatedCents
	out.Balanced = out.RemainingCents == 0
	return out, nil
}
func SortAllocationLines(lines []AllocationLine) {
	sort.SliceStable(lines, func(i, j int) bool { return lines[i].BatchID < lines[j].BatchID })
}
