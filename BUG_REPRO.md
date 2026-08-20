# Bug reproduction

A project identifier is tenant-scoped. Reproduce a cross-tenant read that must return not found while same-tenant reads continue to work.

The deterministic regression test is stored in internal/repository/memory/bug_8_test.go.
