package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/welfare/settlement-resolver/internal/domain"
)

// EvaluateApp runs the rule engine over a cycle's entries and creates
// exceptions for each hit. Idempotency is enforced via the rule_version_id
// plus entry_id plus rule_code: re-evaluation with the same rule version
// produces no new exceptions.
type EvaluateApp struct {
	rules      RuleVersionRepo
	entries    EntryRepo
	exceptions ExceptionRepo
	audit      AuditRepo
	clock      Clock
	engine     *domain.RuleEngine
}

func NewEvaluateApp(rules RuleVersionRepo, entries EntryRepo, exceptions ExceptionRepo, audit AuditRepo, clock Clock) *EvaluateApp {
	if clock == nil {
		clock = systemClock{}
	}
	return &EvaluateApp{
		rules:      rules,
		entries:    entries,
		exceptions: exceptions,
		audit:      audit,
		clock:      clock,
		engine:     domain.NewRuleEngine(clock.Now),
	}
}

// EvaluateCycleInput is the request to evaluate a cycle.
type EvaluateCycleInput struct {
	TenantID      string
	CycleID       string
	RuleVersionID string
	ActorID       string
}

// EvaluateResult reports the number of exceptions created and the
// number of entries scanned.
type EvaluateResult struct {
	ScannedEntries    int
	CreatedExceptions int
	HitEntries        int
}

// EvaluateCycle evaluates the given rule version against every entry of
// the cycle. It is idempotent: existing (entry_id, rule_code) pairs are
// not re-created.
func (a *EvaluateApp) EvaluateCycle(ctx context.Context, in EvaluateCycleInput) (EvaluateResult, error) {
	if strings.TrimSpace(in.TenantID) == "" {
		return EvaluateResult{}, domain.NewErr(domain.CodeInvalidArgument, "tenant id required").WithField("tenant_id")
	}
	if in.CycleID == "" || in.RuleVersionID == "" {
		return EvaluateResult{}, domain.NewErr(domain.CodeInvalidArgument, "cycle_id and rule_version_id required")
	}
	rv, err := a.rules.Get(ctx, in.TenantID, in.RuleVersionID)
	if err != nil {
		return EvaluateResult{}, err
	}
	if rv.Status != domain.RuleVersionStatusPublished {
		return EvaluateResult{}, domain.NewErrf(domain.CodeFailedPrecondition, "rule version %s is not published", rv.Code)
	}
	entries, err := a.entries.ListByCycle(ctx, in.TenantID, in.CycleID)
	if err != nil {
		return EvaluateResult{}, err
	}
	existing, err := a.exceptions.ListByCycle(ctx, in.TenantID, in.CycleID)
	if err != nil {
		return EvaluateResult{}, err
	}
	seen := make(map[string]struct{}, len(existing))
	for _, ex := range existing {
		seen[ex.EntryID+"|"+ex.RuleCode] = struct{}{}
	}
	created := 0
	hitEntries := 0
	for _, e := range entries {
		hits := a.engine.Evaluate(rv, e)
		if len(hits) == 0 {
			continue
		}
		hitEntries++
		for _, h := range hits {
			key := e.ID + "|" + h.Rule.Code
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			ex := domain.Exception{
				ID:            domain.NewID(),
				TenantID:      in.TenantID,
				CycleID:       in.CycleID,
				EntryID:       e.ID,
				RuleVersionID: rv.ID,
				RuleCode:      h.Rule.Code,
				Severity:      h.Rule.Severity,
				Title:         fmt.Sprintf("%s: %s", h.Rule.Code, e.ID),
				Description:   h.Rule.Description,
				HitReason:     h.HitReason,
				Status:        domain.ExceptionStatusPending,
				ReporterID:    in.ActorID,
				DeadlineAt:    h.DeadlineAt,
				CreatedAt:     a.clock.Now(),
				UpdatedAt:     a.clock.Now(),
				Version:       1,
				Snapshot:      h.Snapshot,
			}
			if err := ex.Validate(); err != nil {
				return EvaluateResult{}, err
			}
			if _, err := a.exceptions.Create(ctx, ex); err != nil {
				return EvaluateResult{}, err
			}
			created++
		}
	}
	_, _ = a.audit.Append(ctx, domain.AuditEntry{
		ID:         domain.NewID(),
		TenantID:   in.TenantID,
		ActorID:    in.ActorID,
		Action:     domain.AuditActionRecalculate,
		EntityType: "settlement_cycle",
		EntityID:   in.CycleID,
		Detail: map[string]string{
			"scanned":      itoa(len(entries)),
			"created":      itoa(created),
			"hit_entries":  itoa(hitEntries),
			"rule_version": rv.Code,
		},
		CreatedAt: a.clock.Now(),
	})
	return EvaluateResult{
		ScannedEntries:    len(entries),
		CreatedExceptions: created,
		HitEntries:        hitEntries,
	}, nil
}

// EnsureTime is a convenience helper used by the scheduler to compute
// deadlines in the past.
func (a *EvaluateApp) EnsureTime(t time.Time) time.Time {
	if t.IsZero() {
		return a.clock.Now()
	}
	return t
}
