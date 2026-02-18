package validation

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// Pagination represents pagination parameters.
type Pagination struct {
	Limit  int
	Offset int
}

// ValidatePagination validates and normalizes pagination parameters.
// Returns an error if values are invalid, otherwise normalizes them within bounds.
func ValidatePagination(limit, offset int, maxLimit int) (int, int, error) {
	if maxLimit <= 0 {
		maxLimit = 200
	}

	if limit < 0 {
		return 0, 0, errors.New("limit cannot be negative")
	}
	if offset < 0 {
		return 0, 0, errors.New("offset cannot be negative")
	}

	// Apply defaults and bounds
	if limit == 0 {
		limit = 50
	}
	if limit > maxLimit {
		limit = maxLimit
	}

	return limit, offset, nil
}

// ValidateUUID validates that a string is a valid UUID.
func ValidateUUID(id string, fieldName string) error {
	if id == "" {
		return fmt.Errorf("%s is required", fieldName)
	}
	if _, err := uuid.Parse(id); err != nil {
		return fmt.Errorf("%s is not a valid UUID: %s", fieldName, id)
	}
	return nil
}

// ValidatePositiveInt validates that an integer is positive (> 0).
func ValidatePositiveInt(val int, fieldName string) error {
	if val <= 0 {
		return fmt.Errorf("%s must be positive, got %d", fieldName, val)
	}
	return nil
}

// ValidateNonNegativeInt validates that an integer is non-negative (>= 0).
func ValidateNonNegativeInt(val int, fieldName string) error {
	if val < 0 {
		return fmt.Errorf("%s cannot be negative, got %d", fieldName, val)
	}
	return nil
}

// ValidateNonEmptyString validates that a string is not empty.
func ValidateNonEmptyString(val string, fieldName string) error {
	if val == "" {
		return fmt.Errorf("%s is required", fieldName)
	}
	return nil
}

// ValidateStringOneOf validates that a string is one of the allowed values.
func ValidateStringOneOf(val string, fieldName string, allowed []string) error {
	for _, a := range allowed {
		if val == a {
			return nil
		}
	}
	return fmt.Errorf("%s must be one of %v, got %q", fieldName, allowed, val)
}
