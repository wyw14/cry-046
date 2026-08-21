package application

import (
	"context"
	"github.com/welfare/settlement-resolver/internal/domain"
)

type AllocationResult struct {
	Plan     domain.AllocationPlan
	Warnings []string
}

func BuildAllocationResult(ctx context.Context, projectID string, budget int64, lines []domain.AllocationLine) (AllocationResult, error) {
	plan, err := domain.BuildAllocationPlan(projectID, budget, lines)
	if err != nil {
		return AllocationResult{}, err
	}
	return AllocationResult{Plan: plan, Warnings: []string{}}, nil
}
