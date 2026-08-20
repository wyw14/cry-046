package http

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// validate is the package-level validator instance. Tags are described
// in the DTO field tags.
var validate = validator.New()

// validatorErrs is the type returned by the validator on failure.
type validatorErrs = validator.ValidationErrors

// validatorFieldError is a single field error.
type validatorFieldError = validator.FieldError

func init() {
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		tag := fld.Tag.Get("json")
		if name := strings.TrimSpace(strings.SplitN(tag, ",", 2)[0]); name != "" && name != "-" {
			return name
		}
		return fld.Name
	})
}
