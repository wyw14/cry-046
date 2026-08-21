package application

import (
	"context"
	"github.com/welfare/settlement-resolver/internal/domain"
	"testing"
)

func TestRework_ClearsPreviousAssignee(t *testing.T) {
	ex := makeException()
	ex.Status = domain.ExceptionStatusResolved
	ex.AssigneeID = "old"
	ex.ResolvedAt = newFakeClock().Now()
	app := NewExceptionsApp(&exceptionRepoFake{items: []domain.Exception{ex}}, &auditRepoFake{}, newFakeClock())
	out, err := app.Rework(context.Background(), ReworkInput{TenantID: ex.TenantID, ExceptionID: ex.ID, AuthorID: "reviewer", Note: "reopen"})
	if err != nil {
		t.Fatal(err)
	}
	if out.AssigneeID != "" {
		t.Fatalf("rework retained stale assignee %q", out.AssigneeID)
	}
}
