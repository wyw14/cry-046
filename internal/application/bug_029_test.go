package application

import (
	"context"
	"testing"
)

func TestWorkspace_RequiresTenantIdentity(t *testing.T) {
	app := NewWorkspaceApp(&exceptionRepoFake{}, nilNotify{}, newFakeClock())
	_, err := app.GetWorkspace(context.Background(), "", "u1")
	if err == nil {
		t.Fatal("empty tenant must be rejected")
	}
}
