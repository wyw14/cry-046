package application

import (
	"context"
	"github.com/welfare/settlement-resolver/internal/domain"
)

type AuthorizationReport struct {
	Decisions []domain.AccessDecision
	Allowed   int
	Denied    int
}

func EvaluateAuthorization(ctx context.Context, user domain.User, tenantID string, actions []string) (AuthorizationReport, error) {
	out := AuthorizationReport{}
	for _, action := range actions {
		d, err := domain.AuthorizeAction(user, tenantID, action)
		if err != nil {
			return AuthorizationReport{}, err
		}
		out.Decisions = append(out.Decisions, d)
		if d.Allowed {
			out.Allowed++
		} else {
			out.Denied++
		}
	}
	return out, nil
}
