package dtos

import "strings"

func ValidatePassword(password string) error {
	trimmed := strings.TrimSpace(password)
	if trimmed == "" {
		return ErrFieldRequired("password")
	}
	if len(trimmed) < 8 {
		return ErrFieldInvalid("password", "must be at least 8 characters")
	}
	return nil
}
