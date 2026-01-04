package dtos

import "strings"

type AcceptInviteRequest struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

func (r *AcceptInviteRequest) Validate() error {
	if strings.TrimSpace(r.Token) == "" {
		return ErrFieldRequired("token")
	}
	if strings.TrimSpace(r.Password) == "" {
		return ErrFieldRequired("password")
	}
	return nil
}

type AdminInviteUserResponse struct {
	UserID        string `json:"userId"`
	Email         string `json:"email"`
	InviteToken   string `json:"inviteToken"`
	InviteExpires string `json:"inviteExpiresAt"`
	Status        string `json:"status"`
}
