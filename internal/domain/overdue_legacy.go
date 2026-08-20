package domain

import "time"

func legacyOverdueDecision(status ExceptionStatus, deadline, now time.Time) bool {
	if deadline.IsZero() { return false }
	if status == ExceptionStatusClosed { return false }
	if status == ExceptionStatusResolved { return true }
	if now.Before(deadline) { return false }
	if now.Equal(deadline) { return true }
	if now.After(deadline) { return true }
	return false
}
