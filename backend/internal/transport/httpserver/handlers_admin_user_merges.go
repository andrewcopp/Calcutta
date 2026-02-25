package httpserver

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/dtos"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/httperr"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/response"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func (s *Server) registerAdminUserMergeRoutes(r *mux.Router) {
	r.HandleFunc("/api/admin/users/stubs", s.requirePermission("admin.users.read", s.adminListStubUsersHandler)).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/admin/users/merge", s.requirePermission("admin.users.write", s.adminMergeUsersHandler)).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/admin/users/{id}/merge-candidates", s.requirePermission("admin.users.read", s.adminFindMergeCandidatesHandler)).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/admin/users/{id}/merges", s.requirePermission("admin.users.read", s.adminListMergeHistoryHandler)).Methods("GET", "OPTIONS")
}

func (s *Server) adminListStubUsersHandler(w http.ResponseWriter, r *http.Request) {
	if s.app.UserManagement == nil {
		httperr.Write(w, r, http.StatusInternalServerError, "internal_error", "user management not available", "")
		return
	}

	users, err := s.app.UserManagement.ListStubUsers(r.Context())
	if err != nil {
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}

	items := make([]dtos.StubUserResponse, 0, len(users))
	for _, u := range users {
		items = append(items, dtos.NewStubUserResponse(u))
	}

	response.WriteJSON(w, http.StatusOK, dtos.StubUsersListResponse{Items: items})
}

func (s *Server) adminFindMergeCandidatesHandler(w http.ResponseWriter, r *http.Request) {
	if s.app.UserManagement == nil {
		httperr.Write(w, r, http.StatusInternalServerError, "internal_error", "user management not available", "")
		return
	}

	userID := strings.TrimSpace(mux.Vars(r)["id"])
	if userID == "" {
		httperr.WriteFromErr(w, r, dtos.ErrFieldRequired("id"), authUserID)
		return
	}
	if _, err := uuid.Parse(userID); err != nil {
		httperr.WriteFromErr(w, r, dtos.ErrFieldInvalid("id", "invalid uuid"), authUserID)
		return
	}

	users, err := s.app.UserManagement.FindMergeCandidates(r.Context(), userID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}

	items := make([]dtos.MergeCandidateResponse, 0, len(users))
	for _, u := range users {
		items = append(items, dtos.NewMergeCandidateResponse(u))
	}

	response.WriteJSON(w, http.StatusOK, dtos.MergeCandidatesListResponse{Items: items})
}

func (s *Server) adminMergeUsersHandler(w http.ResponseWriter, r *http.Request) {
	if s.app.UserManagement == nil {
		httperr.Write(w, r, http.StatusInternalServerError, "internal_error", "user management not available", "")
		return
	}

	var req dtos.MergeUsersRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperr.Write(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}

	if strings.TrimSpace(req.SourceUserID) == "" {
		httperr.WriteFromErr(w, r, dtos.ErrFieldRequired("sourceUserId"), authUserID)
		return
	}
	if _, err := uuid.Parse(req.SourceUserID); err != nil {
		httperr.WriteFromErr(w, r, dtos.ErrFieldInvalid("sourceUserId", "invalid uuid"), authUserID)
		return
	}
	if strings.TrimSpace(req.TargetUserID) == "" {
		httperr.WriteFromErr(w, r, dtos.ErrFieldRequired("targetUserId"), authUserID)
		return
	}
	if _, err := uuid.Parse(req.TargetUserID); err != nil {
		httperr.WriteFromErr(w, r, dtos.ErrFieldInvalid("targetUserId", "invalid uuid"), authUserID)
		return
	}
	if req.SourceUserID == req.TargetUserID {
		httperr.WriteFromErr(w, r, dtos.ErrFieldInvalid("targetUserId", "source and target must be different users"), authUserID)
		return
	}

	mergedBy := authUserID(r.Context())
	merge, err := s.app.UserManagement.MergeUsers(r.Context(), req.SourceUserID, req.TargetUserID, mergedBy)
	if err != nil {
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}

	response.WriteJSON(w, http.StatusOK, dtos.NewUserMergeResponse(merge))
}

func (s *Server) adminListMergeHistoryHandler(w http.ResponseWriter, r *http.Request) {
	if s.app.UserManagement == nil {
		httperr.Write(w, r, http.StatusInternalServerError, "internal_error", "user management not available", "")
		return
	}

	userID := strings.TrimSpace(mux.Vars(r)["id"])
	if userID == "" {
		httperr.WriteFromErr(w, r, dtos.ErrFieldRequired("id"), authUserID)
		return
	}
	if _, err := uuid.Parse(userID); err != nil {
		httperr.WriteFromErr(w, r, dtos.ErrFieldInvalid("id", "invalid uuid"), authUserID)
		return
	}

	merges, err := s.app.UserManagement.ListMergeHistory(r.Context(), userID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}

	items := make([]dtos.UserMergeResponse, 0, len(merges))
	for _, m := range merges {
		items = append(items, dtos.NewUserMergeResponse(m))
	}

	response.WriteJSON(w, http.StatusOK, dtos.MergeHistoryResponse{Items: items})
}
