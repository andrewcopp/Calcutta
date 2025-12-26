package httpserver

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	dbadapters "github.com/andrewcopp/Calcutta/backend/internal/adapters/db"
	"github.com/andrewcopp/Calcutta/backend/internal/auth"
	"github.com/gorilla/mux"
)

type adminAPIKeyCreateRequest struct {
	Label *string `json:"label"`
}

type adminAPIKeyCreateResponse struct {
	ID        string  `json:"id"`
	Key       string  `json:"key"`
	Label     *string `json:"label,omitempty"`
	CreatedAt string  `json:"created_at"`
}

type adminAPIKeyListItem struct {
	ID         string     `json:"id"`
	Label      *string    `json:"label,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
}

type adminAPIKeyListResponse struct {
	Items []adminAPIKeyListItem `json:"items"`
}

func (s *Server) registerAdminAPIKeyRoutes(r *mux.Router) {
	r.HandleFunc("/api/admin/api-keys", s.requirePermission("admin.api_keys.write", s.adminAPIKeysCreateHandler)).Methods("POST")
	r.HandleFunc("/api/admin/api-keys", s.requirePermission("admin.api_keys.write", s.adminAPIKeysListHandler)).Methods("GET")
	r.HandleFunc("/api/admin/api-keys/{id}", s.requirePermission("admin.api_keys.write", s.adminAPIKeysRevokeHandler)).Methods("DELETE")
}

func (s *Server) adminAPIKeysCreateHandler(w http.ResponseWriter, r *http.Request) {
	if s.apiKeysRepo == nil {
		writeError(w, r, http.StatusInternalServerError, "internal_error", "api keys repo not available", "")
		return
	}

	userID := authUserID(r.Context())
	if userID == "" {
		writeError(w, r, http.StatusUnauthorized, "unauthorized", "Authentication required", "")
		return
	}

	var req adminAPIKeyCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}

	if req.Label != nil {
		lbl := strings.TrimSpace(*req.Label)
		if lbl == "" {
			req.Label = nil
		} else {
			req.Label = &lbl
		}
	}

	raw, err := auth.NewAPIKey()
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	keyHash := dbadapters.HashAPIKey(raw)

	apiKey, err := s.apiKeysRepo.Create(r.Context(), userID, keyHash, req.Label, time.Now().UTC())
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	writeJSON(w, http.StatusCreated, adminAPIKeyCreateResponse{ID: apiKey.ID, Key: raw, Label: apiKey.Label, CreatedAt: apiKey.CreatedAt.Format(time.RFC3339)})
}

func (s *Server) adminAPIKeysListHandler(w http.ResponseWriter, r *http.Request) {
	if s.apiKeysRepo == nil {
		writeError(w, r, http.StatusInternalServerError, "internal_error", "api keys repo not available", "")
		return
	}

	userID := authUserID(r.Context())
	if userID == "" {
		writeError(w, r, http.StatusUnauthorized, "unauthorized", "Authentication required", "")
		return
	}

	keys, err := s.apiKeysRepo.ListByUser(r.Context(), userID)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	items := make([]adminAPIKeyListItem, 0, len(keys))
	for _, k := range keys {
		items = append(items, adminAPIKeyListItem{ID: k.ID, Label: k.Label, CreatedAt: k.CreatedAt, RevokedAt: k.RevokedAt, LastUsedAt: k.LastUsedAt})
	}

	writeJSON(w, http.StatusOK, adminAPIKeyListResponse{Items: items})
}

func (s *Server) adminAPIKeysRevokeHandler(w http.ResponseWriter, r *http.Request) {
	if s.apiKeysRepo == nil {
		writeError(w, r, http.StatusInternalServerError, "internal_error", "api keys repo not available", "")
		return
	}

	userID := authUserID(r.Context())
	if userID == "" {
		writeError(w, r, http.StatusUnauthorized, "unauthorized", "Authentication required", "")
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "API key ID is required", "id")
		return
	}

	if err := s.apiKeysRepo.Revoke(r.Context(), id, userID, time.Now().UTC()); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
