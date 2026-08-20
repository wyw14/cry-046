package domain

import (
	"sort"
	"strings"
	"time"
)

// ExceptionStatus is the lifecycle state of an exception.
type ExceptionStatus string

const (
	ExceptionStatusPending    ExceptionStatus = "pending"
	ExceptionStatusProcessing ExceptionStatus = "processing"
	ExceptionStatusReview     ExceptionStatus = "review"
	ExceptionStatusResolved   ExceptionStatus = "resolved"
	ExceptionStatusClosed     ExceptionStatus = "closed"
	ExceptionStatusEscalated  ExceptionStatus = "escalated"
)

// AllowedTransitions returns the set of statuses the exception may
// transition to from the current status. The state machine is
// deliberately strict: once an exception is closed it cannot be
// re-opened, and once resolved it can only be closed (or re-opened
// back to processing with a new note in the rare case of false
// resolution — recorded as a rework event).
func (s ExceptionStatus) AllowedTransitions() []ExceptionStatus {
	switch s {
	case ExceptionStatusPending:
		return []ExceptionStatus{ExceptionStatusProcessing, ExceptionStatusEscalated, ExceptionStatusClosed}
	case ExceptionStatusProcessing:
		return []ExceptionStatus{ExceptionStatusReview, ExceptionStatusPending, ExceptionStatusEscalated, ExceptionStatusResolved}
	case ExceptionStatusReview:
		return []ExceptionStatus{ExceptionStatusProcessing, ExceptionStatusResolved, ExceptionStatusPending}
	case ExceptionStatusResolved:
		return []ExceptionStatus{ExceptionStatusClosed, ExceptionStatusProcessing}
	case ExceptionStatusClosed:
		return []ExceptionStatus{}
	case ExceptionStatusEscalated:
		return []ExceptionStatus{ExceptionStatusProcessing, ExceptionStatusClosed}
	}
	return nil
}

// CanTransition reports whether a transition is allowed.
func (s ExceptionStatus) CanTransition(to ExceptionStatus) bool {
	if s == to {
		return true
	}
	for _, allowed := range s.AllowedTransitions() {
		if allowed == to {
			return true
		}
	}
	return false
}

// Exception is a hit of an exception rule against a settlement entry.
// An entry may have multiple exceptions, each tied to a different rule.
type Exception struct {
	ID            string
	TenantID      string
	CycleID       string
	EntryID       string
	RuleVersionID string
	RuleCode      string
	Severity      Severity
	Title         string
	Description   string
	HitReason     string
	Status        ExceptionStatus
	AssigneeID    string // empty means unassigned
	ReporterID    string // operator who detected it
	DeadlineAt    time.Time
	ResolvedAt    time.Time
	ClosedAt      time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Version       int // optimistic concurrency
	Notes         []ExceptionNote
	Attachments   []Attachment
	Snapshot      ExceptionSnapshot
}

// ExceptionNote is an append-only comment added to an exception.
type ExceptionNote struct {
	ID        string
	AuthorID  string
	Body      string
	CreatedAt time.Time
	Kind      NoteKind
}

// NoteKind is the kind of a note.
type NoteKind string

const (
	NoteKindComment    NoteKind = "comment"
	NoteKindAssignment NoteKind = "assignment"
	NoteKindClaim      NoteKind = "claim"
	NoteKindResubmit   NoteKind = "resubmit"
	NoteKindReview     NoteKind = "review"
	NoteKindEscalation NoteKind = "escalation"
	NoteKindRework     NoteKind = "rework"
)

// IsValidNoteKind reports whether k is a known note kind.
func IsValidNoteKind(k NoteKind) bool {
	switch k {
	case NoteKindComment, NoteKindAssignment, NoteKindClaim,
		NoteKindResubmit, NoteKindReview, NoteKindEscalation, NoteKindRework:
		return true
	}
	return false
}

// Attachment is an evidence file attached to an exception.
type Attachment struct {
	ID           string
	OriginalName string
	ContentType  string
	Size         int64
	SHA256       string
	StoredPath   string
	UploadedBy   string
	CreatedAt    time.Time
}

// ExceptionSnapshot is the immutable input snapshot captured at the
// time the exception was created. It is kept forever so that any later
// recalculation can be reproduced exactly.
type ExceptionSnapshot struct {
	EntryAmountCents int64
	EntryCurrency    string
	EntryOccurredAt  time.Time
	RuleExpression   string
	RuleSeverity     Severity
	InputFields      map[string]string // copy of business fields used by the rule
	SnapshotAt       time.Time
}

// Validate checks the invariants.
func (e Exception) Validate() error {
	if e.ID == "" {
		return NewErr(CodeInvalidArgument, "exception id must not be empty").WithField("id")
	}
	if e.TenantID == "" {
		return NewErr(CodeInvalidArgument, "tenant id must not be empty").WithField("tenant_id")
	}
	if e.CycleID == "" {
		return NewErr(CodeInvalidArgument, "cycle id must not be empty").WithField("cycle_id")
	}
	if e.EntryID == "" {
		return NewErr(CodeInvalidArgument, "entry id must not be empty").WithField("entry_id")
	}
	if e.RuleVersionID == "" {
		return NewErr(CodeInvalidArgument, "rule version id must not be empty").WithField("rule_version_id")
	}
	if strings.TrimSpace(e.RuleCode) == "" {
		return NewErr(CodeInvalidArgument, "rule code must not be empty").WithField("rule_code")
	}
	if !IsValidSeverity(e.Severity) {
		return NewErr(CodeInvalidArgument, "invalid severity").WithField("severity")
	}
	if strings.TrimSpace(e.Title) == "" {
		return NewErr(CodeInvalidArgument, "title must not be empty").WithField("title")
	}
	if e.Snapshot.SnapshotAt.IsZero() {
		return NewErr(CodeFailedPrecondition, "exception snapshot must be captured at creation").WithField("snapshot")
	}
	return nil
}

// IsTerminal reports whether the status is terminal.
func (s ExceptionStatus) IsTerminal() bool {
	return s == ExceptionStatusClosed
}

// AssertCanTransition returns a domain error if the transition is not allowed.
func (s ExceptionStatus) AssertCanTransition(to ExceptionStatus) error {
	if !s.CanTransition(to) {
		return NewErrf(CodeFailedPrecondition, "cannot transition from %s to %s", s, to)
	}
	return nil
}

// SortedNotes returns notes sorted by creation time, then by id.
func (e Exception) SortedNotes() []ExceptionNote {
	out := make([]ExceptionNote, len(e.Notes))
	copy(out, e.Notes)
	sort.Slice(out, func(i, j int) bool {
		if out[i].CreatedAt.Equal(out[j].CreatedAt) {
			return out[i].ID < out[j].ID
		}
		return out[i].CreatedAt.Before(out[j].CreatedAt)
	})
	return out
}
