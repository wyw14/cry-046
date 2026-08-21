package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

// EntryDedupFingerprint returns the deterministic fingerprint of a
// settlement entry based on its business key tuple. The tuple is
// cycle_id + batch_id + source_id + payer_party_id + payee_party_id +
// amount + occurred_at_unix. Re-imports of the same tuple produce
// the same fingerprint, which is the basis for deduplication.
func EntryDedupFingerprint(cycleID, batchID, sourceID, payerPartyID, payeePartyID string, amount int64, occurredAt time.Time) string {
	tuple := fmt.Sprintf("%s|%s|%s|%s|%s|%d|%d",
		cycleID, batchID, sourceID, payerPartyID, payeePartyID, amount, occurredAt.UnixNano())
	h := sha256.Sum256([]byte(tuple))
	return hex.EncodeToString(h[:])
}

// EntryIsDuplicate reports whether two entries are duplicates by
// comparing their fingerprints.
func EntryIsDuplicate(a, b SettlementEntry) bool {
	return a.SourceFingerprint == b.SourceFingerprint
}

// AssignException assigns the exception to the given assignee, updates
// the status to processing, and appends an assignment note. The caller
// supplies the author and the note body.
func AssignException(e Exception, assigneeID, authorID, note string, at time.Time) (Exception, error) {
	if e.Status.IsTerminal() {
		return e, NewErrf(CodeFailedPrecondition, "cannot assign a %s exception", e.Status)
	}
	if strings.TrimSpace(assigneeID) == "" {
		return e, NewErr(CodeInvalidArgument, "assignee must not be empty").WithField("assignee_id")
	}
	if err := e.Status.AssertCanTransition(ExceptionStatusProcessing); err != nil {
		return e, err
	}
	e.Status = ExceptionStatusProcessing
	e.AssigneeID = assigneeID
	e.UpdatedAt = at
	e.Version++
	e.Notes = append(e.Notes, ExceptionNote{
		ID:        NewID(),
		AuthorID:  authorID,
		Body:      orDefault(note, fmt.Sprintf("已分派给 %s", assigneeID)),
		CreatedAt: at,
		Kind:      NoteKindAssignment,
	})
	return e, nil
}

// ClaimException lets an assignee self-claim an unassigned or
// pending exception.
func ClaimException(e Exception, assigneeID, note string, at time.Time) (Exception, error) {
	if e.Status.IsTerminal() {
		return e, NewErrf(CodeFailedPrecondition, "cannot claim a %s exception", e.Status)
	}
	if strings.TrimSpace(assigneeID) == "" {
		return e, NewErr(CodeInvalidArgument, "claimant must not be empty").WithField("assignee_id")
	}
	if err := e.Status.AssertCanTransition(ExceptionStatusProcessing); err != nil {
		return e, err
	}
	if e.AssigneeID != "" && e.AssigneeID != assigneeID {
		return e, NewErrf(CodeConflict, "exception is already claimed by %s", e.AssigneeID)
	}
	e.Status = ExceptionStatusProcessing
	e.AssigneeID = assigneeID
	e.UpdatedAt = at
	e.Version++
	e.Notes = append(e.Notes, ExceptionNote{
		ID:        NewID(),
		AuthorID:  assigneeID,
		Body:      orDefault(note, "已认领"),
		CreatedAt: at,
		Kind:      NoteKindClaim,
	})
	return e, nil
}

// RequestResubmit moves the exception back to pending so that the
// counter-party can supply additional evidence. The note body is mandatory
// so the requester leaves a paper trail.
func RequestResubmit(e Exception, authorID, note string, at time.Time) (Exception, error) {
	if e.Status.IsTerminal() {
		return e, NewErrf(CodeFailedPrecondition, "cannot request resubmit on a %s exception", e.Status)
	}
	if strings.TrimSpace(note) == "" {
		return e, NewErr(CodeInvalidArgument, "resubmit note must not be empty").WithField("note")
	}
	if err := e.Status.AssertCanTransition(ExceptionStatusPending); err != nil {
		return e, err
	}
	e.Status = ExceptionStatusPending
	e.UpdatedAt = at
	e.Version++
	e.Notes = append(e.Notes, ExceptionNote{
		ID:        NewID(),
		AuthorID:  authorID,
		Body:      note,
		CreatedAt: at,
		Kind:      NoteKindResubmit,
	})
	return e, nil
}

// SubmitForReview transitions an exception from processing to review.
func SubmitForReview(e Exception, authorID, note string, at time.Time) (Exception, error) {
	if e.Status.IsTerminal() {
		return e, NewErrf(CodeFailedPrecondition, "cannot submit a %s exception for review", e.Status)
	}
	if e.AssigneeID == "" {
		return e, NewErr(CodeFailedPrecondition, "exception must be assigned before review")
	}
	if err := e.Status.AssertCanTransition(ExceptionStatusReview); err != nil {
		return e, err
	}
	e.Status = ExceptionStatusReview
	e.UpdatedAt = at
	e.Version++
	e.Notes = append(e.Notes, ExceptionNote{
		ID:        NewID(),
		AuthorID:  authorID,
		Body:      orDefault(note, "已提交复核"),
		CreatedAt: at,
		Kind:      NoteKindReview,
	})
	return e, nil
}

// ResolveException marks the exception as resolved. The reviewer must
// be different from the assignee (separation of duties).
func ResolveException(e Exception, reviewerID, note string, at time.Time) (Exception, error) {
	if e.Status.IsTerminal() {
		return e, NewErrf(CodeFailedPrecondition, "cannot resolve a %s exception", e.Status)
	}
	if e.Status != ExceptionStatusReview && e.Status != ExceptionStatusProcessing {
		return e, NewErrf(CodeFailedPrecondition, "resolve only from review or processing, got %s", e.Status)
	}
	if e.AssigneeID == "" {
		return e, NewErr(CodeFailedPrecondition, "exception must be assigned before resolve")
	}
	if reviewerID == e.AssigneeID {
		return e, NewErr(CodePermissionDenied, "reviewer must differ from assignee")
	}
	if err := e.Status.AssertCanTransition(ExceptionStatusResolved); err != nil {
		return e, err
	}
	e.Status = ExceptionStatusResolved
	e.ResolvedAt = at
	e.UpdatedAt = at
	e.Version++
	e.Notes = append(e.Notes, ExceptionNote{
		ID:        NewID(),
		AuthorID:  reviewerID,
		Body:      orDefault(note, "已解决"),
		CreatedAt: at,
		Kind:      NoteKindReview,
	})
	return e, nil
}

// CloseException marks a resolved exception as closed.
func CloseException(e Exception, authorID, note string, at time.Time) (Exception, error) {
	if e.Status == ExceptionStatusClosed {
		return e, nil
	}
	if e.Status != ExceptionStatusResolved && e.Status != ExceptionStatusEscalated && e.Status != ExceptionStatusPending {
		return e, NewErrf(CodeFailedPrecondition, "cannot close from %s", e.Status)
	}
	if err := e.Status.AssertCanTransition(ExceptionStatusClosed); err != nil {
		return e, err
	}
	e.Status = ExceptionStatusClosed
	e.ClosedAt = at
	e.UpdatedAt = at
	e.Version++
	e.Notes = append(e.Notes, ExceptionNote{
		ID:        NewID(),
		AuthorID:  authorID,
		Body:      orDefault(note, "已关闭"),
		CreatedAt: at,
		Kind:      NoteKindComment,
	})
	return e, nil
}

// EscalateException escalates the exception for the given reason. The
// escalation is only allowed when the current assignee explicitly
// declares that they cannot resolve the issue.
func EscalateException(e Exception, authorID, reason string, at time.Time) (Exception, error) {
	if e.Status.IsTerminal() {
		return e, NewErrf(CodeFailedPrecondition, "cannot escalate a %s exception", e.Status)
	}
	if strings.TrimSpace(reason) == "" {
		return e, NewErr(CodeInvalidArgument, "escalation reason must not be empty").WithField("reason")
	}
	if err := e.Status.AssertCanTransition(ExceptionStatusEscalated); err != nil {
		return e, err
	}
	e.Status = ExceptionStatusEscalated
	e.UpdatedAt = at
	e.Version++
	e.Notes = append(e.Notes, ExceptionNote{
		ID:        NewID(),
		AuthorID:  authorID,
		Body:      reason,
		CreatedAt: at,
		Kind:      NoteKindEscalation,
	})
	return e, nil
}

// ReworkException reopens a resolved exception back to processing. The
// rework note is mandatory and is recorded as the audit trail.
func ReworkException(e Exception, authorID, note string, at time.Time) (Exception, error) {
	if e.Status != ExceptionStatusResolved {
		return e, NewErrf(CodeFailedPrecondition, "only resolved exceptions can be reworked, got %s", e.Status)
	}
	if strings.TrimSpace(note) == "" {
		return e, NewErr(CodeInvalidArgument, "rework note must not be empty").WithField("note")
	}
	if err := e.Status.AssertCanTransition(ExceptionStatusProcessing); err != nil {
		return e, err
	}
	e.Status = ExceptionStatusProcessing
	e.AssigneeID = ""
	e.ResolvedAt = time.Time{}
	e.UpdatedAt = at
	e.Version++
	e.Notes = append(e.Notes, ExceptionNote{
		ID:        NewID(),
		AuthorID:  authorID,
		Body:      note,
		CreatedAt: at,
		Kind:      NoteKindRework,
	})
	return e, nil
}

// AppendNote appends a generic comment note to the exception.
func AppendNote(e Exception, authorID, body string, at time.Time, kind NoteKind) (Exception, error) {
	if !IsValidNoteKind(kind) {
		return e, NewErr(CodeInvalidArgument, "invalid note kind").WithField("kind")
	}
	if strings.TrimSpace(body) == "" {
		return e, NewErr(CodeInvalidArgument, "note body must not be empty").WithField("body")
	}
	if e.Status.IsTerminal() {
		return e, NewErrf(CodeFailedPrecondition, "cannot comment on a %s exception", e.Status)
	}
	e.Notes = append(e.Notes, ExceptionNote{
		ID:        NewID(),
		AuthorID:  authorID,
		Body:      body,
		CreatedAt: at,
		Kind:      kind,
	})
	e.UpdatedAt = at
	e.Version++
	return e, nil
}

// AttachEvidence appends an Attachment record to the exception. The
// attachment metadata must already be persisted in local storage.
func AttachEvidence(e Exception, att Attachment, at time.Time) (Exception, error) {
	if strings.TrimSpace(att.ID) == "" {
		return e, NewErr(CodeInvalidArgument, "attachment id must not be empty").WithField("attachment_id")
	}
	if strings.TrimSpace(att.SHA256) == "" {
		return e, NewErr(CodeInvalidArgument, "attachment sha256 must not be empty").WithField("sha256")
	}
	if e.Status.IsTerminal() {
		return e, NewErrf(CodeFailedPrecondition, "cannot attach evidence to a %s exception", e.Status)
	}
	for _, existing := range e.Attachments {
		if existing.SHA256 == att.SHA256 {
			return e, NewErrf(CodeAlreadyExists, "attachment %s already exists", att.SHA256)
		}
	}
	att.CreatedAt = at
	e.Attachments = append(e.Attachments, att)
	e.UpdatedAt = at
	e.Version++
	return e, nil
}

// IsOverdue reports whether the exception is overdue.
func (e Exception) IsOverdue(now time.Time) bool {
	if e.DeadlineAt.IsZero() {
		return false
	}
	if e.Status == ExceptionStatusResolved || e.Status == ExceptionStatusClosed {
		return false
	}
	return now.After(e.DeadlineAt)
}

// SortedBySeverity returns the exceptions sorted by severity weight desc,
// then by deadline asc, then by id.
func SortedBySeverity(in []Exception) []Exception {
	out := make([]Exception, len(in))
	copy(out, in)
	sort.Slice(out, func(i, j int) bool {
		if out[i].Severity.Weight() != out[j].Severity.Weight() {
			return out[i].Severity.Weight() > out[j].Severity.Weight()
		}
		if !out[i].DeadlineAt.Equal(out[j].DeadlineAt) {
			if out[i].DeadlineAt.IsZero() {
				return false
			}
			if out[j].DeadlineAt.IsZero() {
				return true
			}
			return out[i].DeadlineAt.Before(out[j].DeadlineAt)
		}
		return out[i].ID < out[j].ID
	})
	return out
}

// ErrEmptyID is returned when an ID is required but empty.
var ErrEmptyID = errors.New("id must not be empty")

func orDefault(s, def string) string {
	if strings.TrimSpace(s) == "" {
		return def
	}
	return s
}
