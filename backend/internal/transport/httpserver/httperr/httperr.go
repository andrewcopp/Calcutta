package httperr

import (
	"context"
	"errors"
	"net/http"

	"github.com/andrewcopp/Calcutta/backend/internal/app/apperrors"
	"github.com/andrewcopp/Calcutta/backend/internal/app/lab"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/dtos"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/requestctx"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/response"
)

type apiError struct {
	Code      string   `json:"code"`
	Message   string   `json:"message"`
	Field     string   `json:"field,omitempty"`
	RequestID string   `json:"requestId"`
	Errors    []string `json:"errors,omitempty"`
}

type apiErrorEnvelope struct {
	Error apiError `json:"error"`
}

func Write(w http.ResponseWriter, r *http.Request, status int, code string, message string, field string) {
	requestID := requestctx.GetRequestID(r.Context())
	response.WriteJSON(w, status, apiErrorEnvelope{
		Error: apiError{
			Code:      code,
			Message:   message,
			Field:     field,
			RequestID: requestID,
		},
	})
}

func WriteMultiError(w http.ResponseWriter, r *http.Request, status int, code string, message string, errors []string) {
	requestID := requestctx.GetRequestID(r.Context())
	response.WriteJSON(w, status, apiErrorEnvelope{
		Error: apiError{
			Code:      code,
			Message:   message,
			RequestID: requestID,
			Errors:    errors,
		},
	})
}

func WriteFromErr(w http.ResponseWriter, r *http.Request, err error, authUserID func(context.Context) string) {
	if err == nil {
		userID := ""
		if authUserID != nil {
			userID = authUserID(r.Context())
		}
		requestctx.Logger(r.Context()).ErrorContext(
			r.Context(),
			"http_error",
			"event", "http_error",
			"method", r.Method,
			"path", r.URL.Path,
			"status", http.StatusInternalServerError,
			"user_id", userID,
			"error", "nil error",
		)
		Write(w, r, http.StatusInternalServerError, "internal_error", "internal server error", "")
		return
	}

	var validationErr *dtos.ValidationError
	if errors.As(err, &validationErr) {
		Write(w, r, http.StatusBadRequest, "validation_error", validationErr.Message, validationErr.Field)
		return
	}

	var invalidArgErr *apperrors.InvalidArgumentError
	if errors.As(err, &invalidArgErr) {
		Write(w, r, http.StatusBadRequest, "invalid_argument", invalidArgErr.Error(), invalidArgErr.Field)
		return
	}

	var notFoundErr *apperrors.NotFoundError
	if errors.As(err, &notFoundErr) {
		Write(w, r, http.StatusNotFound, "not_found", notFoundErr.Error(), "")
		return
	}

	var alreadyExistsErr *apperrors.AlreadyExistsError
	if errors.As(err, &alreadyExistsErr) {
		Write(w, r, http.StatusConflict, "conflict", alreadyExistsErr.Error(), alreadyExistsErr.Field)
		return
	}

	var unauthorizedErr *apperrors.UnauthorizedError
	if errors.As(err, &unauthorizedErr) {
		Write(w, r, http.StatusUnauthorized, "unauthorized", unauthorizedErr.Error(), "")
		return
	}

	var pipelineAlreadyRunningErr *lab.PipelineAlreadyRunningError
	if errors.As(err, &pipelineAlreadyRunningErr) {
		Write(w, r, http.StatusConflict, "pipeline_already_running", pipelineAlreadyRunningErr.Error(), "")
		return
	}

	var noCalcuttasErr *lab.NoCalcuttasAvailableError
	if errors.As(err, &noCalcuttasErr) {
		Write(w, r, http.StatusBadRequest, "no_calcuttas_available", noCalcuttasErr.Error(), "")
		return
	}

	var pipelineNotCancellableErr *lab.PipelineNotCancellableError
	if errors.As(err, &pipelineNotCancellableErr) {
		Write(w, r, http.StatusConflict, "pipeline_not_cancellable", pipelineNotCancellableErr.Error(), "")
		return
	}

	var pipelineNotAvailableErr *lab.PipelineNotAvailableError
	if errors.As(err, &pipelineNotAvailableErr) {
		Write(w, r, http.StatusServiceUnavailable, "pipeline_not_available", pipelineNotAvailableErr.Error(), "")
		return
	}

	userID := ""
	if authUserID != nil {
		userID = authUserID(r.Context())
	}
	requestctx.Logger(r.Context()).ErrorContext(
		r.Context(),
		"http_error",
		"event", "http_error",
		"method", r.Method,
		"path", r.URL.Path,
		"status", http.StatusInternalServerError,
		"user_id", userID,
		"error", err.Error(),
	)
	Write(w, r, http.StatusInternalServerError, "internal_error", "internal server error", "")
}
