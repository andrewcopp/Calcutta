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
