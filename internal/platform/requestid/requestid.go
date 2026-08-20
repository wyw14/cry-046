// Package requestid provides a stable request identifier that is
// propagated through context. Handlers stamp incoming requests with
// a fresh UUID and downstream services can read it from context.
package requestid

import (
	"context"

	"github.com/google/uuid"
)

type ctxKey struct{}

// New generates a fresh request identifier.
func New() string { return uuid.NewString() }

// With stamps the context with the given request id.
func With(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, ctxKey{}, id)
}

// From returns the request id stored in context, or the empty string.
func From(ctx context.Context) string {
	if v, ok := ctx.Value(ctxKey{}).(string); ok {
		return v
	}
	return ""
}
