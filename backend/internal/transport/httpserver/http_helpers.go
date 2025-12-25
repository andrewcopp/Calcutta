package httpserver

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/dtos"
	"github.com/andrewcopp/Calcutta/backend/pkg/services"
)

type apiError struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Field     string `json:"field,omitempty"`
	RequestID string `json:"requestId"`
}

type apiErrorEnvelope struct {
	Error apiError `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, r *http.Request, status int, code string, message string, field string) {
	requestID := getRequestID(r.Context())
	writeJSON(w, status, apiErrorEnvelope{
		Error: apiError{
			Code:      code,
			Message:   message,
			Field:     field,
			RequestID: requestID,
		},
	})
}

func writeErrorFromErr(w http.ResponseWriter, r *http.Request, err error) {
	if err == nil {
		writeError(w, r, http.StatusInternalServerError, "internal_error", "internal server error", "")
		return
	}

	var validationErr *dtos.ValidationError
	if errors.As(err, &validationErr) {
		writeError(w, r, http.StatusBadRequest, "validation_error", validationErr.Message, validationErr.Field)
		return
	}

	var notFoundErr *services.NotFoundError
	if errors.As(err, &notFoundErr) {
		writeError(w, r, http.StatusNotFound, "not_found", notFoundErr.Error(), "")
		return
	}

	var alreadyExistsErr *services.AlreadyExistsError
	if errors.As(err, &alreadyExistsErr) {
		writeError(w, r, http.StatusConflict, "conflict", alreadyExistsErr.Error(), alreadyExistsErr.Field)
		return
	}

	log.Printf("[%s] internal error: %v", getRequestID(r.Context()), err)
	writeError(w, r, http.StatusInternalServerError, "internal_error", "internal server error", "")
}
