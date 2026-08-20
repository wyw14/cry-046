// Package domain defines the core business entities and value objects
// for the Welfare Settlement Exception Resolution Platform. The domain
// layer has no infrastructure imports: it is plain Go and can be tested
// without a database or HTTP server.
package domain

import (
	"errors"
	"fmt"
	"time"
)

// Stable error codes used by the HTTP layer to produce stable API errors.
const (
	CodeUnknown            = "UNKNOWN"
	CodeInvalidArgument    = "INVALID_ARGUMENT"
	CodeNotFound           = "NOT_FOUND"
	CodeAlreadyExists      = "ALREADY_EXISTS"
	CodePermissionDenied   = "PERMISSION_DENIED"
	CodeFailedPrecondition = "FAILED_PRECONDITION"
	CodeAborted            = "ABORTED"
	CodeOutOfRange         = "OUT_OF_RANGE"
	CodeUnauthenticated    = "UNAUTHENTICATED"
	CodeConflict           = "CONFLICT"
)

// Error is the canonical domain error. The Code field is stable and
// surfaced to the HTTP layer; the Message is human-readable and is
// safe to log/return to the operator; fields is an optional list of
// (field, value) pairs used to produce field-level validation errors.
type Error struct {
	Code    string
	Message string
	Field   string
	Wrapped error
}

func (e *Error) Error() string {
	if e.Wrapped != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Wrapped)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *Error) Unwrap() error { return e.Wrapped }

// NewErr constructs a domain Error.
func NewErr(code, msg string) *Error { return &Error{Code: code, Message: msg} }

// NewErrf is NewErr with fmt.Sprintf.
func NewErrf(code, format string, args ...any) *Error {
	return &Error{Code: code, Message: fmt.Sprintf(format, args...)}
}

// WrapErr wraps an existing error with the given code/message.
func WrapErr(code, msg string, err error) *Error {
	return &Error{Code: code, Message: msg, Wrapped: err}
}

// WithField returns a copy of the error with the field name set.
func (e *Error) WithField(field string) *Error {
	cp := *e
	cp.Field = field
	return &cp
}

// AsDomainError extracts a *Error from any error. If err is nil or
// already a *Error, it is returned as-is. Otherwise it is wrapped
// as an unknown error.
func AsDomainError(err error) *Error {
	if err == nil {
		return nil
	}
	var de *Error
	if errors.As(err, &de) {
		return de
	}
	return &Error{Code: CodeUnknown, Message: err.Error(), Wrapped: err}
}

// IsCode returns true if err carries the given stable code.
func IsCode(err error, code string) bool {
	var de *Error
	if errors.As(err, &de) {
		return de.Code == code
	}
	return false
}

// IsNotFound is a convenience check.
func IsNotFound(err error) bool { return IsCode(err, CodeNotFound) }

// IsAlreadyExists is a convenience check.
func IsAlreadyExists(err error) bool { return IsCode(err, CodeAlreadyExists) }

// IsPermissionDenied is a convenience check.
func IsPermissionDenied(err error) bool { return IsCode(err, CodePermissionDenied) }

// IsFailedPrecondition is a convenience check.
func IsFailedPrecondition(err error) bool { return IsCode(err, CodeFailedPrecondition) }

// IsOutOfRange is a convenience check.
func IsOutOfRange(err error) bool { return IsCode(err, CodeOutOfRange) }

// IsAborted is a convenience check (used for optimistic-concurrency conflicts).
func IsAborted(err error) bool { return IsCode(err, CodeAborted) }

// MustParseTime parses a time string or panics. Used only in tests.
func MustParseTime(layout, s string) time.Time {
	t, err := time.Parse(layout, s)
	if err != nil {
		panic(err)
	}
	return t
}
