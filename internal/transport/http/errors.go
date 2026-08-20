// Package http defines the Gin-based HTTP transport for the platform.
// It contains handlers, request/response DTOs, middleware and the
// router wiring. The transport depends on the application layer via
// constructor injection; it never reaches into the domain layer's
// private helpers.
package http

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/welfare/settlement-resolver/internal/domain"
)

// ErrorEnvelope is the JSON shape returned for every error.
type ErrorEnvelope struct {
	Code      string       `json:"code"`
	Message   string       `json:"message"`
	Fields    []FieldError `json:"fields,omitempty"`
	RequestID string       `json:"request_id"`
}

// FieldError is a single field-level validation error.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// writeError writes a domain.Error to the response. If err is not a
// *domain.Error, it is wrapped as CodeUnknown.
func writeError(c *gin.Context, err error) {
	if err == nil {
		c.Status(http.StatusNoContent)
		return
	}
	de := domain.AsDomainError(err)
	status := statusFor(de.Code)
	env := ErrorEnvelope{
		Code:      de.Code,
		Message:   de.Message,
		RequestID: c.GetString("request_id"),
	}
	if de.Field != "" {
		env.Fields = []FieldError{{Field: de.Field, Message: de.Message}}
	}
	c.AbortWithStatusJSON(status, env)
}

// statusFor maps a stable error code to an HTTP status.
func statusFor(code string) int {
	switch code {
	case domain.CodeInvalidArgument, domain.CodeOutOfRange:
		return http.StatusBadRequest
	case domain.CodeUnauthenticated:
		return http.StatusUnauthorized
	case domain.CodePermissionDenied:
		return http.StatusForbidden
	case domain.CodeNotFound:
		return http.StatusNotFound
	case domain.CodeAlreadyExists, domain.CodeConflict:
		return http.StatusConflict
	case domain.CodeFailedPrecondition:
		return http.StatusPreconditionFailed
	case domain.CodeAborted:
		return http.StatusConflict
	}
	return http.StatusInternalServerError
}

// writeOK writes a 200 with the given JSON body.
func writeOK(c *gin.Context, body any) {
	if body == nil {
		c.Status(http.StatusNoContent)
		return
	}
	c.JSON(http.StatusOK, body)
}

// writeCreated writes a 201 with the given JSON body.
func writeCreated(c *gin.Context, body any) {
	c.JSON(http.StatusCreated, body)
}

// accepted writes a 202 with the given JSON body.
func accepted(c *gin.Context, body any) {
	c.JSON(http.StatusAccepted, body)
}

// bindAndValidate binds the request body to dst, performs go-playground
// validator validation, and translates errors into the ErrorEnvelope.
func bindAndValidate(c *gin.Context, dst any) bool {
	if err := c.ShouldBindJSON(dst); err != nil {
		writeError(c, domain.NewErr(domain.CodeInvalidArgument, err.Error()).WithField("body"))
		return false
	}
	if err := validate.Struct(dst); err != nil {
		fields := make([]FieldError, 0)
		if verrs, ok := err.(validatorErrs); ok {
			for _, fe := range verrs {
				fields = append(fields, FieldError{
					Field:   lowerFirst(fe.Field()),
					Message: msgForTag(fe),
				})
			}
		}
		env := ErrorEnvelope{
			Code:      domain.CodeInvalidArgument,
			Message:   "validation failed",
			Fields:    fields,
			RequestID: c.GetString("request_id"),
		}
		c.AbortWithStatusJSON(http.StatusBadRequest, env)
		return false
	}
	return true
}

func lowerFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToLower(s[:1]) + s[1:]
}

func msgForTag(fe validatorFieldError) string {
	switch fe.Tag() {
	case "required":
		return "字段必填"
	case "min":
		return "长度过短"
	case "max":
		return "长度过长"
	case "oneof":
		return "取值不在允许范围"
	case "len":
		return "长度不符"
	case "gte":
		return "数值过小"
	case "lte":
		return "数值过大"
	case "email":
		return "邮箱格式无效"
	}
	return "校验失败"
}

// ErrNoActor is returned when the request context has no actor.
var ErrNoActor = errors.New("no actor in context")
