package dtos

import "strings"

func ValidatePassword(password string) error {
	if strings.TrimSpace(password) == "" {
		return ErrFieldRequired("password")
	}
	if len(password) < 12 {
		return ErrFieldInvalid("password", "must be at least 12 characters")
	}
	if len(password) > 128 {
		return ErrFieldInvalid("password", "must be at most 128 characters")
	}
	return nil
}
