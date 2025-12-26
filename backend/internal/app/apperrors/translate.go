package apperrors

import (
	"errors"

	"github.com/andrewcopp/Calcutta/backend/pkg/services"
)

func Translate(err error) error {
	if err == nil {
		return nil
	}

	var notFoundErr *services.NotFoundError
	if errors.As(err, &notFoundErr) {
		return &NotFoundError{Resource: notFoundErr.Resource, ID: notFoundErr.ID}
	}

	var alreadyExistsErr *services.AlreadyExistsError
	if errors.As(err, &alreadyExistsErr) {
		return &AlreadyExistsError{Resource: alreadyExistsErr.Resource, Field: alreadyExistsErr.Field, Value: alreadyExistsErr.Value}
	}

	return err
}
