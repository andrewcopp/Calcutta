package httpserver

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/andrewcopp/Calcutta/backend/internal/app/apperrors"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/dtos"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/httperr"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/response"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

var validRoleKeys = map[string]bool{
	"site_admin":       true,
	"tournament_admin": true,
	"pool_admin":       true,
	"player":           true,
	"user_manager":     true,
}

var validRoleScopes = map[string]map[string]bool{
	"site_admin":       {"global": true},
	"user_manager":     {"global": true},
	"pool_admin":       {"global": true, "pool": true},
	"tournament_admin": {"global": true, "tournament": true},
	"player":           {"global": true, "pool": true},
}

func (s *Server) meProfileHandler(w http.ResponseWriter, r *http.Request) {
	userID := authUserID(r.Context())
	if userID == "" {
		httperr.Write(w, r, http.StatusUnauthorized, "unauthorized", "Authentication required", "")
		return
	}

	user, err := s.userRepo.GetByID(r.Context(), userID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}
	if user == nil {
		httperr.WriteFromErr(w, r, &apperrors.NotFoundError{Resource: "user", ID: userID}, authUserID)
		return
	}

	roles, err := s.authzRepo.ListUserGlobalRoles(r.Context(), userID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}
	if roles == nil {
		roles = []string{}
	}

	permissions, err := s.authzRepo.ListUserGlobalPermissions(r.Context(), userID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}
	if permissions == nil {
		permissions = []string{}
	}

	response.WriteJSON(w, http.StatusOK, dtos.UserProfileResponse{
		ID:          user.ID,
		Email:       user.Email,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Status:      user.Status,
		Roles:       roles,
		Permissions: permissions,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	})
}

func (s *Server) adminUserDetailHandler(w http.ResponseWriter, r *http.Request) {
	userID := strings.TrimSpace(mux.Vars(r)["id"])
	if userID == "" {
		httperr.WriteFromErr(w, r, dtos.ErrFieldRequired("id"), authUserID)
		return
	}
	if _, err := uuid.Parse(userID); err != nil {
		httperr.WriteFromErr(w, r, dtos.ErrFieldInvalid("id", "invalid uuid"), authUserID)
		return
	}

	row, err := s.userRepo.GetAdminUserByID(r.Context(), userID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}
	if row == nil {
		httperr.WriteFromErr(w, r, &apperrors.NotFoundError{Resource: "user", ID: userID}, authUserID)
		return
	}

	grantRows, err := s.authzRepo.ListUserRolesWithScope(r.Context(), userID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}

	roles := make([]dtos.RoleGrant, 0, len(grantRows))
	for _, g := range grantRows {
		roles = append(roles, dtos.RoleGrant{
			Key:       g.Key,
			ScopeType: g.ScopeType,
			ScopeID:   g.ScopeID,
			ScopeName: g.ScopeName,
		})
	}

	response.WriteJSON(w, http.StatusOK, dtos.AdminUserDetailResponse{
		ID:          row.ID,
		Email:       row.Email,
		FirstName:   row.FirstName,
		LastName:    row.LastName,
		Status:      row.Status,
		Roles:       roles,
		Permissions: row.Permissions,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	})
}

func (s *Server) adminGrantRoleHandler(w http.ResponseWriter, r *http.Request) {
	userID := strings.TrimSpace(mux.Vars(r)["id"])
	if userID == "" {
		httperr.WriteFromErr(w, r, dtos.ErrFieldRequired("id"), authUserID)
		return
	}
	if _, err := uuid.Parse(userID); err != nil {
		httperr.WriteFromErr(w, r, dtos.ErrFieldInvalid("id", "invalid uuid"), authUserID)
		return
	}

	var req struct {
		RoleKey   string `json:"roleKey"`
		ScopeType string `json:"scopeType"`
		ScopeID   string `json:"scopeId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperr.Write(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}
	req.RoleKey = strings.TrimSpace(req.RoleKey)
	if req.RoleKey == "" {
		httperr.WriteFromErr(w, r, dtos.ErrFieldRequired("roleKey"), authUserID)
		return
	}
	if !validRoleKeys[req.RoleKey] {
		httperr.WriteFromErr(w, r, dtos.ErrFieldInvalid("roleKey", "unknown role key"), authUserID)
		return
	}

	if req.ScopeType == "" {
		req.ScopeType = "global"
	}
	scopes := validRoleScopes[req.RoleKey]
	if !scopes[req.ScopeType] {
		httperr.WriteFromErr(w, r, dtos.ErrFieldInvalid("scopeType", "invalid scope for this role"), authUserID)
		return
	}

	if req.ScopeType == "global" {
		if err := s.authzRepo.GrantGlobalRole(r.Context(), userID, req.RoleKey); err != nil {
			httperr.WriteFromErr(w, r, err, authUserID)
			return
		}
	} else {
		if _, err := uuid.Parse(req.ScopeID); err != nil {
			httperr.WriteFromErr(w, r, dtos.ErrFieldInvalid("scopeId", "invalid uuid"), authUserID)
			return
		}
		if err := s.authzRepo.GrantRole(r.Context(), userID, req.RoleKey, req.ScopeType, req.ScopeID); err != nil {
			httperr.WriteFromErr(w, r, err, authUserID)
			return
		}
	}

	response.WriteJSON(w, http.StatusOK, map[string]string{"status": "granted"})
}

func (s *Server) adminRevokeRoleHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := strings.TrimSpace(vars["id"])
	roleKey := strings.TrimSpace(vars["roleKey"])

	if userID == "" {
		httperr.WriteFromErr(w, r, dtos.ErrFieldRequired("id"), authUserID)
		return
	}
	if _, err := uuid.Parse(userID); err != nil {
		httperr.WriteFromErr(w, r, dtos.ErrFieldInvalid("id", "invalid uuid"), authUserID)
		return
	}
	if roleKey == "" {
		httperr.WriteFromErr(w, r, dtos.ErrFieldRequired("roleKey"), authUserID)
		return
	}
	if !validRoleKeys[roleKey] {
		httperr.WriteFromErr(w, r, dtos.ErrFieldInvalid("roleKey", "unknown role key"), authUserID)
		return
	}

	scopeType := r.URL.Query().Get("scopeType")
	scopeID := r.URL.Query().Get("scopeId")
	if scopeType == "" {
		scopeType = "global"
	}

	if scopeType == "global" {
		if err := s.authzRepo.RevokeGlobalRole(r.Context(), userID, roleKey); err != nil {
			httperr.WriteFromErr(w, r, err, authUserID)
			return
		}
	} else {
		if _, err := uuid.Parse(scopeID); err != nil {
			httperr.WriteFromErr(w, r, dtos.ErrFieldInvalid("scopeId", "invalid uuid"), authUserID)
			return
		}
		if err := s.authzRepo.RevokeGrant(r.Context(), userID, roleKey, scopeType, scopeID); err != nil {
			httperr.WriteFromErr(w, r, err, authUserID)
			return
		}
	}

	response.WriteJSON(w, http.StatusOK, map[string]string{"status": "revoked"})
}
