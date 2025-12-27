package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/andrewcopp/Calcutta/backend/internal/app/apperrors"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/dtos"
)

type apiErrorEnvelopeForTest struct {
	Error struct {
		Code      string `json:"code"`
		Message   string `json:"message"`
		Field     string `json:"field,omitempty"`
		RequestID string `json:"requestId"`
	} `json:"error"`
}

func TestThatWriteErrorFromErrReturns500ForNilError(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(context.WithValue(r.Context(), requestIDKey, "req-1"))
	w := httptest.NewRecorder()

	writeErrorFromErr(w, r, nil)

	got := w.Result().StatusCode
	want := http.StatusInternalServerError
	if got != want {
		t.Errorf("expected status %d, got %d", want, got)
	}
}

func TestThatWriteErrorFromErrReturns400ForValidationError(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(context.WithValue(r.Context(), requestIDKey, "req-1"))
	w := httptest.NewRecorder()

	writeErrorFromErr(w, r, &dtos.ValidationError{Field: "name", Message: "bad"})

	got := w.Result().StatusCode
	want := http.StatusBadRequest
	if got != want {
		t.Errorf("expected status %d, got %d", want, got)
	}
}

func TestThatWriteErrorFromErrReturnsInvalidArgumentCodeForInvalidArgumentError(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(context.WithValue(r.Context(), requestIDKey, "req-1"))
	w := httptest.NewRecorder()

	writeErrorFromErr(w, r, &apperrors.InvalidArgumentError{Field: "seed", Message: "invalid"})

	var env apiErrorEnvelopeForTest
	if err := json.NewDecoder(w.Result().Body).Decode(&env); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	got := env.Error.Code
	want := "invalid_argument"
	if got != want {
		t.Errorf("expected code %q, got %q", want, got)
	}
}

func TestThatWriteErrorFromErrReturns404ForNotFoundError(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(context.WithValue(r.Context(), requestIDKey, "req-1"))
	w := httptest.NewRecorder()

	writeErrorFromErr(w, r, &apperrors.NotFoundError{Resource: "thing", ID: "id"})

	got := w.Result().StatusCode
	want := http.StatusNotFound
	if got != want {
		t.Errorf("expected status %d, got %d", want, got)
	}
}

func TestThatWriteErrorFromErrReturns401ForUnauthorizedError(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(context.WithValue(r.Context(), requestIDKey, "req-1"))
	w := httptest.NewRecorder()

	writeErrorFromErr(w, r, &apperrors.UnauthorizedError{})

	got := w.Result().StatusCode
	want := http.StatusUnauthorized
	if got != want {
		t.Errorf("expected status %d, got %d", want, got)
	}
}

func TestThatWriteErrorFromErrReturns500ForUnknownError(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(context.WithValue(r.Context(), requestIDKey, "req-1"))
	w := httptest.NewRecorder()

	writeErrorFromErr(w, r, errors.New("boom"))

	got := w.Result().StatusCode
	want := http.StatusInternalServerError
	if got != want {
		t.Errorf("expected status %d, got %d", want, got)
	}
}
