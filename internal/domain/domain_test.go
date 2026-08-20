package domain

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestProjectValidate(t *testing.T) {
	now := time.Now()
	cases := []struct {
		name    string
		project Project
		wantErr string
	}{
		{
			name:    "empty id",
			project: Project{},
			wantErr: "project id must not be empty",
		},
		{
			name: "empty tenant",
			project: Project{
				ID: "p1",
			},
			wantErr: "tenant id must not be empty",
		},
		{
			name: "empty code",
			project: Project{
				ID:       "p1",
				TenantID: "t1",
			},
			wantErr: "project code must not be empty",
		},
		{
			name: "empty name",
			project: Project{
				ID:       "p1",
				TenantID: "t1",
				Code:     "WS-01",
			},
			wantErr: "project name must not be empty",
		},
		{
			name: "empty sponsor",
			project: Project{
				ID:       "p1",
				TenantID: "t1",
				Code:     "WS-01",
				Name:     "Project 1",
			},
			wantErr: "sponsor must not be empty",
		},
		{
			name: "negative budget",
			project: Project{
				ID: "p1", TenantID: "t1", Code: "WS-01", Name: "Project 1", Sponsor: "S",
				AnnualBudget: -1,
			},
			wantErr: "annual budget must be non-negative",
		},
		{
			name: "invalid years",
			project: Project{
				ID: "p1", TenantID: "t1", Code: "WS-01", Name: "Project 1", Sponsor: "S",
				StartYear: 0, EndYear: 0,
			},
			wantErr: "year must be positive",
		},
		{
			name: "end before start",
			project: Project{
				ID: "p1", TenantID: "t1", Code: "WS-01", Name: "Project 1", Sponsor: "S",
				StartYear: 2027, EndYear: 2026,
			},
			wantErr: "end year must be >= start year",
		},
		{
			name: "valid",
			project: Project{
				ID: "p1", TenantID: "t1", Code: "WS-01", Name: "Project 1", Sponsor: "S",
				StartYear: 2026, EndYear: 2027, CreatedAt: now, UpdatedAt: now,
			},
			wantErr: "",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.project.Validate()
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error %q, got nil", tc.wantErr)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("expected error containing %q, got %q", tc.wantErr, err.Error())
			}
		})
	}
}

func TestProjectAnnualBudgetFloat(t *testing.T) {
	p := Project{AnnualBudget: 12345}
	got := p.AnnualBudgetFloat()
	if got != 123.45 {
		t.Fatalf("expected 123.45, got %f", got)
	}
}

func TestProjectContainsYear(t *testing.T) {
	p := Project{StartYear: 2026, EndYear: 2027}
	if !p.ContainsYear(2026) {
		t.Error("2026 should be contained")
	}
	if !p.ContainsYear(2027) {
		t.Error("2027 should be contained")
	}
	if p.ContainsYear(2025) {
		t.Error("2025 should not be contained")
	}
	if p.ContainsYear(2028) {
		t.Error("2028 should not be contained")
	}
}

func TestNewID(t *testing.T) {
	id1 := NewID()
	id2 := NewID()
	if id1 == "" {
		t.Fatal("id must not be empty")
	}
	if len(id1) != 32 {
		t.Fatalf("expected 32-char hex id, got %d chars", len(id1))
	}
	if id1 == id2 {
		t.Fatal("ids must be unique")
	}
}

func TestSanitiseContact(t *testing.T) {
	cases := []struct {
		in      string
		want    string
		wantErr bool
	}{
		{"", "", true},
		{"sponsor@local", "sponsor@local", false},
		{"user@example.com", "user@example.com", false},
		{"+8613800000001", "+8613800000001", false},
		{"123456", "123456", false},
		{"ab", "", true},
		{"not-an-email-or-phone", "", true},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got, err := SanitiseContact(tc.in)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error for %q, got %q", tc.in, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for %q: %v", tc.in, err)
			}
			if got != tc.want {
				t.Fatalf("expected %q, got %q", tc.want, got)
			}
		})
	}
}

func TestMaskContact(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"", ""},
		{"a@b.c", "*@b.c"},
		{"ab@b.c", "**@b.c"},
		{"abc@b.c", "a*c@b.c"},
		{"abcdef@b.c", "a****f@b.c"},
		{"123", "***"},
		{"1234", "****"},
		{"12345", "12*45"},
		{"1234567890", "12******90"},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got := MaskContact(tc.in)
			if got != tc.want {
				t.Fatalf("expected %q, got %q", tc.want, got)
			}
		})
	}
}

func TestMaskEmail(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"", ""},
		{"no-email", "no-email"},
		{"a@b.c", "*@b.c"},
		{"ab@b.c", "**@b.c"},
		{"abc@b.c", "a*c@b.c"},
		{"abcdef@b.c", "a****f@b.c"},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got := MaskEmail(tc.in)
			if got != tc.want {
				t.Fatalf("expected %q, got %q", tc.want, got)
			}
		})
	}
}

func TestSeverityWeight(t *testing.T) {
	cases := []struct {
		s    Severity
		want int
	}{
		{SeverityLow, 1},
		{SeverityMedium, 2},
		{SeverityHigh, 3},
		{SeverityCritical, 4},
		{Severity("unknown"), 0},
	}
	for _, tc := range cases {
		t.Run(string(tc.s), func(t *testing.T) {
			if got := tc.s.Weight(); got != tc.want {
				t.Fatalf("expected %d, got %d", tc.want, got)
			}
		})
	}
}

func TestIsValidSeverity(t *testing.T) {
	for _, s := range []Severity{SeverityLow, SeverityMedium, SeverityHigh, SeverityCritical} {
		if !IsValidSeverity(s) {
			t.Errorf("expected %s to be valid", s)
		}
	}
	if IsValidSeverity(Severity("bogus")) {
		t.Error("expected bogus to be invalid")
	}
}

func TestIsValidEntrySource(t *testing.T) {
	for _, s := range []EntrySource{EntrySourceImport, EntrySourceManual, EntrySourceResubmit} {
		if !IsValidEntrySource(s) {
			t.Errorf("expected %s to be valid", s)
		}
	}
	if IsValidEntrySource(EntrySource("bogus")) {
		t.Error("expected bogus to be invalid")
	}
}

func TestIsValidNoteKind(t *testing.T) {
	for _, k := range []NoteKind{
		NoteKindComment, NoteKindAssignment, NoteKindClaim,
		NoteKindResubmit, NoteKindReview, NoteKindEscalation, NoteKindRework,
	} {
		if !IsValidNoteKind(k) {
			t.Errorf("expected %s to be valid", k)
		}
	}
	if IsValidNoteKind(NoteKind("bogus")) {
		t.Error("expected bogus to be invalid")
	}
}

func TestIsValidRole(t *testing.T) {
	for _, r := range []Role{RoleOperator, RoleAssignee, RoleReviewer, RoleAdmin} {
		if !IsValidRole(r) {
			t.Errorf("expected %s to be valid", r)
		}
	}
	if IsValidRole(Role("bogus")) {
		t.Error("expected bogus to be invalid")
	}
}

func TestUserCan(t *testing.T) {
	cases := []struct {
		role Role
		want struct {
			canResolve bool
			canAssign  bool
			canArchive bool
		}
	}{
		{RoleAdmin, struct {
			canResolve bool
			canAssign  bool
			canArchive bool
		}{true, true, true}},
		{RoleReviewer, struct {
			canResolve bool
			canAssign  bool
			canArchive bool
		}{true, true, false}},
		{RoleOperator, struct {
			canResolve bool
			canAssign  bool
			canArchive bool
		}{false, true, false}},
		{RoleAssignee, struct {
			canResolve bool
			canAssign  bool
			canArchive bool
		}{false, false, false}},
	}
	for _, tc := range cases {
		t.Run(string(tc.role), func(t *testing.T) {
			u := User{Role: tc.role}
			if u.CanResolve() != tc.want.canResolve {
				t.Errorf("CanResolve: expected %v, got %v", tc.want.canResolve, u.CanResolve())
			}
			if u.CanAssign() != tc.want.canAssign {
				t.Errorf("CanAssign: expected %v, got %v", tc.want.canAssign, u.CanAssign())
			}
			if u.CanArchive() != tc.want.canArchive {
				t.Errorf("CanArchive: expected %v, got %v", tc.want.canArchive, u.CanArchive())
			}
		})
	}
}

func TestDomainError(t *testing.T) {
	err := NewErr(CodeInvalidArgument, "test error")
	if err.Code != CodeInvalidArgument {
		t.Errorf("expected code %s, got %s", CodeInvalidArgument, err.Code)
	}
	if err.Message != "test error" {
		t.Errorf("expected message 'test error', got %s", err.Message)
	}
	wrapped := WrapErr(CodeUnknown, "wrapped", err)
	if !errors.Is(wrapped.Unwrap(), err) {
		t.Error("expected wrapped error to unwrap to original")
	}
	fieldErr := err.WithField("name")
	if fieldErr.Field != "name" {
		t.Errorf("expected field 'name', got %s", fieldErr.Field)
	}
}

func TestIsCode(t *testing.T) {
	err := NewErr(CodeNotFound, "missing")
	if !IsCode(err, CodeNotFound) {
		t.Error("expected IsCode to return true for matching code")
	}
	if IsCode(err, CodeAlreadyExists) {
		t.Error("expected IsCode to return false for non-matching code")
	}
	if IsCode(nil, CodeNotFound) {
		t.Error("expected IsCode to return false for nil")
	}
	if !IsNotFound(err) {
		t.Error("expected IsNotFound to return true")
	}
}

func TestExceptionStatusTransitions(t *testing.T) {
	cases := []struct {
		from ExceptionStatus
		to   ExceptionStatus
		want bool
	}{
		{ExceptionStatusPending, ExceptionStatusProcessing, true},
		{ExceptionStatusPending, ExceptionStatusEscalated, true},
		{ExceptionStatusPending, ExceptionStatusClosed, true},
		{ExceptionStatusPending, ExceptionStatusResolved, false},
		{ExceptionStatusProcessing, ExceptionStatusReview, true},
		{ExceptionStatusProcessing, ExceptionStatusResolved, true},
		{ExceptionStatusReview, ExceptionStatusResolved, true},
		{ExceptionStatusResolved, ExceptionStatusClosed, true},
		{ExceptionStatusResolved, ExceptionStatusProcessing, true},
		{ExceptionStatusClosed, ExceptionStatusPending, false},
		{ExceptionStatusClosed, ExceptionStatusResolved, false},
		{ExceptionStatusEscalated, ExceptionStatusProcessing, true},
		{ExceptionStatusEscalated, ExceptionStatusClosed, true},
	}
	for _, tc := range cases {
		t.Run(string(tc.from)+"->"+string(tc.to), func(t *testing.T) {
			got := tc.from.CanTransition(tc.to)
			if got != tc.want {
				t.Errorf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

func TestExceptionStatusIsTerminal(t *testing.T) {
	if !ExceptionStatusClosed.IsTerminal() {
		t.Error("closed should be terminal")
	}
	if ExceptionStatusPending.IsTerminal() {
		t.Error("pending should not be terminal")
	}
}

func TestRuleVersionPublishArchive(t *testing.T) {
	now := time.Now()
	t.Run("publish draft", func(t *testing.T) {
		rv := RuleVersion{Status: RuleVersionStatusDraft}
		rv, err := rv.Publish(now)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if rv.Status != RuleVersionStatusPublished {
			t.Errorf("expected published, got %s", rv.Status)
		}
		if !rv.PublishedAt.Equal(now) {
			t.Error("expected PublishedAt to be set")
		}
	})
	t.Run("publish published fails", func(t *testing.T) {
		rv := RuleVersion{Status: RuleVersionStatusPublished, PublishedAt: now}
		_, err := rv.Publish(now)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
	t.Run("archive published", func(t *testing.T) {
		rv := RuleVersion{Status: RuleVersionStatusPublished, PublishedAt: now}
		rv, err := rv.Archive(now)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if rv.Status != RuleVersionStatusArchived {
			t.Errorf("expected archived, got %s", rv.Status)
		}
	})
	t.Run("archive draft fails", func(t *testing.T) {
		rv := RuleVersion{Status: RuleVersionStatusDraft}
		_, err := rv.Archive(now)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestRuleVersionValidate(t *testing.T) {
	t.Run("empty id", func(t *testing.T) {
		rv := RuleVersion{Code: "RV-1", ProjectID: "p1", Rules: []RuleDefinition{{Code: "R1"}}}
		err := rv.Validate()
		if err == nil || !strings.Contains(err.Error(), "rule version id must not be empty") {
			t.Fatalf("expected id error, got %v", err)
		}
	})
	t.Run("empty code", func(t *testing.T) {
		rv := RuleVersion{ID: "rv1", ProjectID: "p1", Rules: []RuleDefinition{{Code: "R1"}}}
		err := rv.Validate()
		if err == nil || !strings.Contains(err.Error(), "rule version code must not be empty") {
			t.Fatalf("expected code error, got %v", err)
		}
	})
	t.Run("no rules", func(t *testing.T) {
		rv := RuleVersion{ID: "rv1", Code: "RV-1", ProjectID: "p1"}
		err := rv.Validate()
		if err == nil || !strings.Contains(err.Error(), "at least one rule") {
			t.Fatalf("expected rules error, got %v", err)
		}
	})
	t.Run("duplicate rule codes", func(t *testing.T) {
		rv := RuleVersion{
			ID: "rv1", Code: "RV-1", ProjectID: "p1",
			Rules: []RuleDefinition{
				{Code: "R1"},
				{Code: "R1"},
			},
		}
		err := rv.Validate()
		if err == nil || !strings.Contains(err.Error(), "duplicated") {
			t.Fatalf("expected duplicate error, got %v", err)
		}
	})
	t.Run("valid", func(t *testing.T) {
		rv := RuleVersion{
			ID: "rv1", Code: "RV-1", ProjectID: "p1",
			Rules: []RuleDefinition{{Code: "R1"}, {Code: "R2"}},
		}
		if err := rv.Validate(); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})
}

func TestRuleEngineEvaluate(t *testing.T) {
	now := time.Now()
	engine := NewRuleEngine(func() time.Time { return now })
	rv := RuleVersion{
		ID: "rv1", Code: "RV-1", ProjectID: "p1",
		Rules: []RuleDefinition{
			{Code: "AMOUNT_ZERO", Severity: SeverityHigh, Expression: "amount == 0", DeadlineHours: 48},
			{Code: "AMOUNT_GT", Severity: SeverityMedium, Expression: "amount > 1000"},
			{Code: "CURRENCY_USD", Severity: SeverityLow, Expression: "currency == USD"},
			{Code: "META_MATCH", Severity: SeverityMedium, Expression: "meta.note == bad"},
		},
	}
	entry := SettlementEntry{
		ID: "e1", CycleID: "c1", BatchID: "b1", ProjectID: "p1",
		Source: EntrySourceImport, PayeePartyID: "py1", PayerPartyID: "pp1",
		Amount: 0, Currency: "CNY", OccurredAt: now,
		SourceFingerprint: "fp1", Metadata: map[string]string{"note": "bad"},
	}
	hits := engine.Evaluate(rv, entry)
	if len(hits) != 2 {
		t.Fatalf("expected 2 hits (AMOUNT_ZERO and META_MATCH), got %d", len(hits))
	}
	for _, h := range hits {
		if h.Snapshot.SnapshotAt.IsZero() {
			t.Error("snapshot time must be set")
		}
		if h.Snapshot.EntryAmountCents != 0 {
			t.Errorf("expected 0 amount in snapshot, got %d", h.Snapshot.EntryAmountCents)
		}
	}
}

func TestEntryDedupFingerprint(t *testing.T) {
	t1 := time.Now()
	fp1 := EntryDedupFingerprint("c1", "b1", "s1", "pp1", "py1", 100, t1)
	fp2 := EntryDedupFingerprint("c1", "b1", "s1", "pp1", "py1", 100, t1)
	fp3 := EntryDedupFingerprint("c1", "b1", "s1", "pp1", "py1", 200, t1)
	if fp1 != fp2 {
		t.Error("same tuple should produce same fingerprint")
	}
	if fp1 == fp3 {
		t.Error("different amount should produce different fingerprint")
	}
}

func TestAnnualAccumulator(t *testing.T) {
	acc := AnnualAccumulator{ProjectID: "p1", Year: 2026, BudgetCents: 10000}
	if acc.AvailableCents() != 10000 {
		t.Errorf("expected 10000 available, got %d", acc.AvailableCents())
	}
	acc = acc.ApplyAdjustment(Adjustment{ID: "a1", ProjectID: "p1", Year: 2026, DeltaCents: 3000, Reason: "test"})
	if acc.DisbursedCents != 3000 {
		t.Errorf("expected 3000 disbursed, got %d", acc.DisbursedCents)
	}
	if acc.AvailableCents() != 7000 {
		t.Errorf("expected 7000 available, got %d", acc.AvailableCents())
	}
	if acc.OverrunCents() != 0 {
		t.Errorf("expected 0 overrun, got %d", acc.OverrunCents())
	}
	acc = acc.ApplyAdjustment(Adjustment{ID: "a2", ProjectID: "p1", Year: 2026, DeltaCents: 8000, Reason: "overrun"})
	if acc.OverrunCents() != 1000 {
		t.Errorf("expected 1000 overrun, got %d", acc.OverrunCents())
	}
}

func TestSortSummariesByComputedAt(t *testing.T) {
	in := []Summary{
		{ID: "s1", ComputedAt: time.Now(), Version: 1},
		{ID: "s2", ComputedAt: time.Now().Add(-time.Hour), Version: 1},
		{ID: "s3", ComputedAt: time.Now().Add(time.Hour), Version: 1},
	}
	out := SortSummariesByComputedAt(in)
	if out[0].ID != "s3" {
		t.Errorf("expected s3 first, got %s", out[0].ID)
	}
	if out[2].ID != "s2" {
		t.Errorf("expected s2 last, got %s", out[2].ID)
	}
}

func TestRecalculationBatchLifecycle(t *testing.T) {
	now := time.Now()
	rb := RecalculationBatch{
		ID: "rb1", CycleID: "c1", RuleVersionID: "rv1",
		InputDigest: "abc", TriggerReason: "test",
		StartedAt: now, Status: RecalcStatusRunning,
	}
	if err := rb.Validate(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	rb = rb.MarkCompleted(now.Add(time.Minute), Summary{ID: "s1"})
	if rb.Status != RecalcStatusCompleted {
		t.Errorf("expected completed, got %s", rb.Status)
	}
	if !rb.FinishedAt.Equal(now.Add(time.Minute)) {
		t.Error("expected FinishedAt to be set")
	}
	rb = rb.MarkFailed(now.Add(2*time.Minute), "fail")
	if rb.Status != RecalcStatusFailed {
		t.Errorf("expected failed, got %s", rb.Status)
	}
}

func TestMustParseTime(t *testing.T) {
	t1 := MustParseTime("2006-01-02", "2026-01-01")
	if t1.Year() != 2026 {
		t.Errorf("expected 2026, got %d", t1.Year())
	}
}

func TestMustParseTimePanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for invalid time")
		}
	}()
	_ = MustParseTime("2006-01-02", "invalid")
}
