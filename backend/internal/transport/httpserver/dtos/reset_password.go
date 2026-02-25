package dtos

import "strings"

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

func (r *ForgotPasswordRequest) Validate() error {
	if strings.TrimSpace(r.Email) == "" {
		return ErrFieldRequired("email")
	}
	return nil
}

type ResetPasswordRequest struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

func (r *ResetPasswordRequest) Validate() error {
	if strings.TrimSpace(r.Token) == "" {
		return ErrFieldRequired("token")
	}
	if err := ValidatePassword(r.Password); err != nil {
		return err
	}
	return nil
}
