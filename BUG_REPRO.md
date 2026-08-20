# Bug reproduction

Importing the same source line for two different payer parties must not collapse into one settlement entry. Preserve idempotency for an identical payer/payee tuple and keep the fingerprint stable across re-imports.

The deterministic regression test is stored in internal/domain/bug_7_test.go.
