package dtos

import "fmt"

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

func ErrFieldRequired(field string) error {
	return &ValidationError{
		Field:   field,
		Message: "field is required",
	}
}

func ErrFieldInvalid(field, reason string) error {
	return &ValidationError{
		Field:   field,
		Message: reason,
	}
}
