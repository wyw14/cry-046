package domain

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// RuleEngine evaluates a RuleDefinition against a settlement entry
// and returns the resulting hit (or nil) when the rule fires.
type RuleEngine struct {
	now func() time.Time
}

// NewRuleEngine constructs a RuleEngine using the given clock.
func NewRuleEngine(now func() time.Time) *RuleEngine {
	if now == nil {
		now = time.Now
	}
	return &RuleEngine{now: now}
}

// HitResult is returned when a rule fires against an entry.
type HitResult struct {
	Rule       RuleDefinition
	Entry      SettlementEntry
	HitReason  string
	Snapshot   ExceptionSnapshot
	DeadlineAt time.Time
}

// Evaluate runs all rules in the version against the entry. The rules
// are evaluated in the deterministic order returned by SortedRules.
// Returns the list of hits (one per fired rule).
func (r *RuleEngine) Evaluate(rv RuleVersion, e SettlementEntry) []HitResult {
	hits := make([]HitResult, 0)
	for _, rule := range rv.SortedRules() {
		ok, reason, snap := r.evaluate(rule, e)
		if !ok {
			continue
		}
		deadline := r.computeDeadline(rule)
		hits = append(hits, HitResult{
			Rule:       rule,
			Entry:      e,
			HitReason:  reason,
			Snapshot:   snap,
			DeadlineAt: deadline,
		})
	}
	return hits
}

func (r *RuleEngine) computeDeadline(rule RuleDefinition) time.Time {
	if rule.DeadlineHours <= 0 {
		return time.Time{}
	}
	return r.now().Add(time.Duration(rule.DeadlineHours) * time.Hour)
}

func (r *RuleEngine) evaluate(rule RuleDefinition, e SettlementEntry) (bool, string, ExceptionSnapshot) {
	snap := ExceptionSnapshot{
		EntryAmountCents: e.Amount,
		EntryCurrency:    e.Currency,
		EntryOccurredAt:  e.OccurredAt,
		RuleExpression:   rule.Expression,
		RuleSeverity:     rule.Severity,
		InputFields:      copyMap(e.Metadata),
		SnapshotAt:       r.now(),
	}

	ok, reason := evalExpression(rule.Expression, e)
	if !ok {
		return false, "", snap
	}
	if reason == "" {
		reason = fmt.Sprintf("命中规则 %s", rule.Code)
	}
	return true, reason, snap
}

func copyMap(in map[string]string) map[string]string {
	if in == nil {
		return nil
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

// evalExpression is a tiny safe DSL. It supports a small set of
// predicates keyed on the entry's amount, currency and metadata.
// Unknown predicates evaluate to false.
//
// Supported forms (case-sensitive):
//   - "amount == 0"
//   - "amount > N"
//   - "amount < N"
//   - "currency == USD"
//   - "meta.<key> == <value>"
//   - "occurred_before 2026-01-01"
//   - "occurred_after  2026-01-01"
func evalExpression(expr string, e SettlementEntry) (bool, string) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return false, ""
	}
	parts := strings.SplitN(expr, " ", 3)
	if len(parts) < 2 {
		return false, ""
	}
	switch {
	case parts[0] == "amount":
		if len(parts) < 3 {
			return false, ""
		}
		return evalAmount(parts[1], parts[2], e)
	case parts[0] == "currency":
		if len(parts) < 3 {
			return false, ""
		}
		return evalCurrency(parts[1], parts[2], e)
	case strings.HasPrefix(parts[0], "meta."):
		key := strings.TrimPrefix(parts[0], "meta.")
		if key == "" || len(parts) < 3 {
			return false, ""
		}
		return evalMeta(key, parts[1], parts[2], e)
	case parts[0] == "occurred_before":
		if len(parts) < 2 {
			return false, ""
		}
		return evalOccurredBefore(parts[1:], e)
	case parts[0] == "occurred_after":
		if len(parts) < 2 {
			return false, ""
		}
		return evalOccurredAfter(parts[1:], e)
	}
	return false, ""
}

func evalAmount(op, raw string, e SettlementEntry) (bool, string) {
	n, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil {
		return false, ""
	}
	switch op {
	case "==":
		if e.Amount == n {
			return true, fmt.Sprintf("金额等于 %d 分", n)
		}
	case ">":
		if e.Amount > n {
			return true, fmt.Sprintf("金额大于 %d 分", n)
		}
	case "<":
		if e.Amount < n {
			return true, fmt.Sprintf("金额小于 %d 分", n)
		}
	case ">=":
		if e.Amount >= n {
			return true, fmt.Sprintf("金额大于等于 %d 分", n)
		}
	case "<=":
		if e.Amount <= n {
			return true, fmt.Sprintf("金额小于等于 %d 分", n)
		}
	case "!=":
		if e.Amount != n {
			return true, fmt.Sprintf("金额不等于 %d 分", n)
		}
	}
	return false, ""
}

func evalCurrency(op, raw string, e SettlementEntry) (bool, string) {
	if op != "==" {
		return false, ""
	}
	cur := strings.TrimSpace(raw)
	if e.Currency == cur {
		return true, fmt.Sprintf("币种为 %s", cur)
	}
	return false, ""
}

func evalMeta(key, op, val string, e SettlementEntry) (bool, string) {
	val = strings.TrimSpace(val)
	if e.Metadata == nil {
		return false, ""
	}
	got, ok := e.Metadata[key]
	if !ok {
		return false, ""
	}
	switch op {
	case "==":
		if got == val {
			return true, fmt.Sprintf("元数据 %s 等于 %s", key, val)
		}
	case "!=":
		if got != val {
			return true, fmt.Sprintf("元数据 %s 不等于 %s", key, val)
		}
	}
	return false, ""
}

func evalOccurredBefore(parts []string, e SettlementEntry) (bool, string) {
	if len(parts) == 0 {
		return false, ""
	}
	t, err := time.Parse("2006-01-02", strings.TrimSpace(parts[0]))
	if err != nil {
		return false, ""
	}
	if e.OccurredAt.Before(t) {
		return true, fmt.Sprintf("发生日期早于 %s", t.Format("2006-01-02"))
	}
	return false, ""
}

func evalOccurredAfter(parts []string, e SettlementEntry) (bool, string) {
	if len(parts) == 0 {
		return false, ""
	}
	t, err := time.Parse("2006-01-02", strings.TrimSpace(parts[0]))
	if err != nil {
		return false, ""
	}
	if e.OccurredAt.After(t) {
		return true, fmt.Sprintf("发生日期晚于 %s", t.Format("2006-01-02"))
	}
	return false, ""
}
