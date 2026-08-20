# Bug reproduction

A resolved exception is no longer actionable and must never appear overdue, even when its deadline is in the past. Reproduce the state transition boundary without changing overdue behavior for pending or processing items.

The deterministic regression test is stored in internal/domain/bug_11_test.go.
