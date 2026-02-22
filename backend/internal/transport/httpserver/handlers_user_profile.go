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

var validLabelKeys = map[string]bool{
	"site_admin":       true,
	"tournament_admin": true,
	"calcutta_admin":   true,
	"player":           true,
	"user_manager":     true,
}

var validLabelScopes = map[string]map[string]bool{
	"site_admin":       {"global": true},
	"user_manager":     {"global": true},
	"calcutta_admin":   {"global": true, "calcutta": true},
	"tournament_admin": {"global": true, "tournament": true},
	"player":           {"global": true, "calcutta": true},
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

	labels, err := s.authzRepo.ListUserGlobalLabels(r.Context(), userID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}
	if labels == nil {
		labels = []string{}
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
		Labels:      labels,
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

	grantRows, err := s.authzRepo.ListUserLabelsWithScope(r.Context(), userID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}

	labels := make([]dtos.LabelGrant, 0, len(grantRows))
	for _, g := range grantRows {
		labels = append(labels, dtos.LabelGrant{
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
		Labels:      labels,
		Permissions: row.Permissions,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	})
}

func (s *Server) adminGrantLabelHandler(w http.ResponseWriter, r *http.Request) {
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
		LabelKey  string `json:"labelKey"`
		ScopeType string `json:"scopeType"`
		ScopeID   string `json:"scopeId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperr.Write(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}
	req.LabelKey = strings.TrimSpace(req.LabelKey)
	if req.LabelKey == "" {
		httperr.WriteFromErr(w, r, dtos.ErrFieldRequired("labelKey"), authUserID)
		return
	}
	if !validLabelKeys[req.LabelKey] {
		httperr.WriteFromErr(w, r, dtos.ErrFieldInvalid("labelKey", "unknown label key"), authUserID)
		return
	}

	if req.ScopeType == "" {
		req.ScopeType = "global"
	}
	scopes := validLabelScopes[req.LabelKey]
	if !scopes[req.ScopeType] {
		httperr.WriteFromErr(w, r, dtos.ErrFieldInvalid("scopeType", "invalid scope for this label"), authUserID)
		return
	}

	if req.ScopeType == "global" {
		if err := s.authzRepo.GrantGlobalLabel(r.Context(), userID, req.LabelKey); err != nil {
			httperr.WriteFromErr(w, r, err, authUserID)
			return
		}
	} else {
		if _, err := uuid.Parse(req.ScopeID); err != nil {
			httperr.WriteFromErr(w, r, dtos.ErrFieldInvalid("scopeId", "invalid uuid"), authUserID)
			return
		}
		if err := s.authzRepo.GrantLabel(r.Context(), userID, req.LabelKey, req.ScopeType, req.ScopeID); err != nil {
			httperr.WriteFromErr(w, r, err, authUserID)
			return
		}
	}

	response.WriteJSON(w, http.StatusOK, map[string]string{"status": "granted"})
}

func (s *Server) adminRevokeLabelHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := strings.TrimSpace(vars["id"])
	labelKey := strings.TrimSpace(vars["labelKey"])

	if userID == "" {
		httperr.WriteFromErr(w, r, dtos.ErrFieldRequired("id"), authUserID)
		return
	}
	if _, err := uuid.Parse(userID); err != nil {
		httperr.WriteFromErr(w, r, dtos.ErrFieldInvalid("id", "invalid uuid"), authUserID)
		return
	}
	if labelKey == "" {
		httperr.WriteFromErr(w, r, dtos.ErrFieldRequired("labelKey"), authUserID)
		return
	}
	if !validLabelKeys[labelKey] {
		httperr.WriteFromErr(w, r, dtos.ErrFieldInvalid("labelKey", "unknown label key"), authUserID)
		return
	}

	scopeType := r.URL.Query().Get("scopeType")
	scopeID := r.URL.Query().Get("scopeId")
	if scopeType == "" {
		scopeType = "global"
	}

	if scopeType == "global" {
		if err := s.authzRepo.RevokeGlobalLabel(r.Context(), userID, labelKey); err != nil {
			httperr.WriteFromErr(w, r, err, authUserID)
			return
		}
	} else {
		if _, err := uuid.Parse(scopeID); err != nil {
			httperr.WriteFromErr(w, r, dtos.ErrFieldInvalid("scopeId", "invalid uuid"), authUserID)
			return
		}
		if err := s.authzRepo.RevokeGrant(r.Context(), userID, labelKey, scopeType, scopeID); err != nil {
			httperr.WriteFromErr(w, r, err, authUserID)
			return
		}
	}

	response.WriteJSON(w, http.StatusOK, map[string]string{"status": "revoked"})
}
