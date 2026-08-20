package domain

import (
	"strings"
	"time"
)

// Project is the welfare project under which funding batches are organised.
type Project struct {
	ID           string
	TenantID     string
	Code         string // short, unique code, e.g. "WS-2026-01"
	Name         string
	Sponsor      string // the funding sponsor organisation
	AnnualBudget int64  // annual budget in cents (CNY). Stored as integer cents to avoid float drift.
	StartYear    int
	EndYear      int
	Metadata     map[string]string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Validate checks invariants.
func (p Project) Validate() error {
	if p.ID == "" {
		return NewErr(CodeInvalidArgument, "project id must not be empty").WithField("id")
	}
	if p.TenantID == "" {
		return NewErr(CodeInvalidArgument, "tenant id must not be empty").WithField("tenant_id")
	}
	if strings.TrimSpace(p.Code) == "" {
		return NewErr(CodeInvalidArgument, "project code must not be empty").WithField("code")
	}
	if strings.TrimSpace(p.Name) == "" {
		return NewErr(CodeInvalidArgument, "project name must not be empty").WithField("name")
	}
	if strings.TrimSpace(p.Sponsor) == "" {
		return NewErr(CodeInvalidArgument, "sponsor must not be empty").WithField("sponsor")
	}
	if p.AnnualBudget < 0 {
		return NewErr(CodeOutOfRange, "annual budget must be non-negative").WithField("annual_budget")
	}
	if p.StartYear <= 0 || p.EndYear <= 0 {
		return NewErr(CodeInvalidArgument, "year must be positive").WithField("year")
	}
	if p.EndYear < p.StartYear {
		return NewErr(CodeOutOfRange, "end year must be >= start year").WithField("end_year")
	}
	return nil
}

// AnnualBudgetFloat returns the annual budget as a fractional CNY value.
func (p Project) AnnualBudgetFloat() float64 { return float64(p.AnnualBudget) / 100.0 }

// ContainsYear reports whether the project covers the given year.
func (p Project) ContainsYear(year int) bool { return year >= p.StartYear && year <= p.EndYear }
