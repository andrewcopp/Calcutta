package httpserver

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/dtos"
	"github.com/gorilla/mux"
)

func (s *Server) updateCalcuttaHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	calcuttaID := vars["id"]
	if calcuttaID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Calcutta ID is required", "id")
		return
	}

	userID := authUserID(r.Context())
	if userID == "" {
		writeError(w, r, http.StatusUnauthorized, "unauthorized", "Authentication required", "")
		return
	}

	calcutta, err := s.app.Calcutta.GetCalcuttaByID(r.Context(), calcuttaID)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	if calcutta.OwnerID != userID {
		ok, err := s.authzRepo.HasPermission(r.Context(), userID, "global", "", "calcutta.config.write")
		if err != nil {
			writeErrorFromErr(w, r, err)
			return
		}
		if !ok {
			writeError(w, r, http.StatusForbidden, "forbidden", "Insufficient permissions", "")
			return
		}
	}

	var req dtos.UpdateCalcuttaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}
	if err := req.Validate(); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			writeError(w, r, http.StatusBadRequest, "validation_error", "Name is required", "name")
			return
		}
		calcutta.Name = name
	}
	if req.MinTeams != nil {
		calcutta.MinTeams = *req.MinTeams
	}
	if req.MaxTeams != nil {
		calcutta.MaxTeams = *req.MaxTeams
	}
	if req.MaxBid != nil {
		calcutta.MaxBid = *req.MaxBid
	}

	if err := s.app.Calcutta.UpdateCalcutta(r.Context(), calcutta); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, dtos.NewCalcuttaResponse(calcutta))
}
