package apperrors

import "fmt"

type NotFoundError struct {
	Resource string
	ID       string
}

func (e *NotFoundError) Error() string {
	if e.ID == "" {
		return fmt.Sprintf("%s not found", e.Resource)
	}
	return fmt.Sprintf("%s not found: %s", e.Resource, e.ID)
}

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

func FieldRequired(field string) error {
	return &ValidationError{Field: field, Message: "field is required"}
}

func FieldInvalid(field, reason string) error {
	return &ValidationError{Field: field, Message: reason}
}

type AlreadyExistsError struct {
	Resource string
	Field    string
	Value    string
}

func (e *AlreadyExistsError) Error() string {
	if e.Field == "" {
		return fmt.Sprintf("%s already exists", e.Resource)
	}
	if e.Value == "" {
		return fmt.Sprintf("%s already exists: %s", e.Resource, e.Field)
	}
	return fmt.Sprintf("%s already exists: %s=%s", e.Resource, e.Field, e.Value)
}

type UnauthorizedError struct {
	Message string
}

func (e *UnauthorizedError) Error() string {
	if e.Message == "" {
		return "unauthorized"
	}
	return e.Message
}

type InvalidArgumentError struct {
	Field   string
	Message string
}

func (e *InvalidArgumentError) Error() string {
	if e.Message == "" {
		return "invalid argument"
	}
	return e.Message
}

type ForbiddenError struct {
	Message string
}

func (e *ForbiddenError) Error() string {
	if e.Message == "" {
		return "forbidden"
	}
	return e.Message
}

type ConflictError struct {
	Message string
}

func (e *ConflictError) Error() string {
	if e.Message == "" {
		return "conflict"
	}
	return e.Message
}
