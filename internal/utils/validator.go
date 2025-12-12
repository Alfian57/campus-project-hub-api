package utils

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()

	// Use JSON field names in validation errors
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
}

// Validate validates a struct based on its tags
func Validate(s interface{}) error {
	return validate.Struct(s)
}

// ValidationError represents a single validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// FormatValidationErrors converts validator errors to a readable format
func FormatValidationErrors(err error) []ValidationError {
	var errors []ValidationError

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			errors = append(errors, ValidationError{
				Field:   e.Field(),
				Message: getErrorMessage(e),
			})
		}
	}

	return errors
}

func getErrorMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "Field ini wajib diisi"
	case "email":
		return "Format email tidak valid"
	case "min":
		return "Minimal " + e.Param() + " karakter"
	case "max":
		return "Maksimal " + e.Param() + " karakter"
	case "oneof":
		return "Harus salah satu dari: " + e.Param()
	case "url":
		return "Format URL tidak valid"
	case "uuid":
		return "Format UUID tidak valid"
	default:
		return "Field tidak valid"
	}
}
