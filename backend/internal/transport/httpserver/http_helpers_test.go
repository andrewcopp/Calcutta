package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestThatWriteErrorFromErrReturnsInternalErrorCodeForUnknownError(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(context.WithValue(r.Context(), requestIDKey, "req-1"))
	w := httptest.NewRecorder()

	writeErrorFromErr(w, r, errors.New("boom"))

	var env apiErrorEnvelopeForTest
	if err := json.NewDecoder(w.Result().Body).Decode(&env); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	got := env.Error.Code
	want := "internal_error"
	if got != want {
		t.Errorf("expected code %q, got %q", want, got)
	}
}

func TestThatWriteErrorFromErrIncludesRequestIDInResponse(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(context.WithValue(r.Context(), requestIDKey, "req-1"))
	w := httptest.NewRecorder()

	writeErrorFromErr(w, r, errors.New("boom"))

	var env apiErrorEnvelopeForTest
	if err := json.NewDecoder(w.Result().Body).Decode(&env); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	got := env.Error.RequestID
	want := "req-1"
	if got != want {
		t.Errorf("expected requestId %q, got %q", want, got)
	}
}

func TestThatWriteErrorFromErrLogsStructuredJSONForUnknownError(t *testing.T) {
	var buf strings.Builder
	old := log.Writer()
	oldFlags := log.Flags()
	log.SetOutput(&buf)
	log.SetFlags(0)
	t.Cleanup(func() {
		log.SetOutput(old)
		log.SetFlags(oldFlags)
	})

	r := httptest.NewRequest(http.MethodGet, "/boom", nil)
	r = r.WithContext(context.WithValue(r.Context(), requestIDKey, "req-1"))
	w := httptest.NewRecorder()

	writeErrorFromErr(w, r, errors.New("boom"))

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) < 1 {
		t.Fatalf("expected at least 1 log line")
	}
	last := lines[len(lines)-1]

	var payload map[string]any
	if err := json.Unmarshal([]byte(last), &payload); err != nil {
		t.Fatalf("expected JSON log payload, got %q", last)
	}

	if got, want := payload["event"], "http_error"; got != want {
		t.Errorf("expected event %v, got %v", want, got)
	}
	if got, want := payload["request_id"], "req-1"; got != want {
		t.Errorf("expected request_id %v, got %v", want, got)
	}
	if got, want := payload["method"], http.MethodGet; got != want {
		t.Errorf("expected method %v, got %v", want, got)
	}
	if got, want := payload["path"], "/boom"; got != want {
		t.Errorf("expected path %v, got %v", want, got)
	}
	if got, want := payload["status"], float64(http.StatusInternalServerError); got != want {
		t.Errorf("expected status %v, got %v", want, got)
	}
	if got, want := payload["error"], "boom"; got != want {
		t.Errorf("expected error %v, got %v", want, got)
	}
}

func TestThatWriteErrorFromErrReturnsUnauthorizedCodeForUnauthorizedError(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(context.WithValue(r.Context(), requestIDKey, "req-1"))
	w := httptest.NewRecorder()

	writeErrorFromErr(w, r, &apperrors.UnauthorizedError{})

	var env apiErrorEnvelopeForTest
	if err := json.NewDecoder(w.Result().Body).Decode(&env); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	got := env.Error.Code
	want := "unauthorized"
	if got != want {
		t.Errorf("expected code %q, got %q", want, got)
	}
}

func TestThatWriteErrorFromErrReturnsNotFoundCodeForNotFoundError(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(context.WithValue(r.Context(), requestIDKey, "req-1"))
	w := httptest.NewRecorder()

	writeErrorFromErr(w, r, &apperrors.NotFoundError{Resource: "thing", ID: "id"})

	var env apiErrorEnvelopeForTest
	if err := json.NewDecoder(w.Result().Body).Decode(&env); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	got := env.Error.Code
	want := "not_found"
	if got != want {
		t.Errorf("expected code %q, got %q", want, got)
	}
}

func TestThatWriteErrorFromErrReturns409ForAlreadyExistsError(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(context.WithValue(r.Context(), requestIDKey, "req-1"))
	w := httptest.NewRecorder()

	writeErrorFromErr(w, r, &apperrors.AlreadyExistsError{Resource: "thing", Field: "name", Value: "x"})

	got := w.Result().StatusCode
	want := http.StatusConflict
	if got != want {
		t.Errorf("expected status %d, got %d", want, got)
	}
}

func TestThatWriteErrorFromErrReturnsConflictCodeForAlreadyExistsError(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(context.WithValue(r.Context(), requestIDKey, "req-1"))
	w := httptest.NewRecorder()

	writeErrorFromErr(w, r, &apperrors.AlreadyExistsError{Resource: "thing", Field: "name", Value: "x"})

	var env apiErrorEnvelopeForTest
	if err := json.NewDecoder(w.Result().Body).Decode(&env); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	got := env.Error.Code
	want := "conflict"
	if got != want {
		t.Errorf("expected code %q, got %q", want, got)
	}
}

func TestThatWriteErrorFromErrReturnsConflictFieldForAlreadyExistsError(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(context.WithValue(r.Context(), requestIDKey, "req-1"))
	w := httptest.NewRecorder()

	writeErrorFromErr(w, r, &apperrors.AlreadyExistsError{Resource: "thing", Field: "name", Value: "x"})

	var env apiErrorEnvelopeForTest
	if err := json.NewDecoder(w.Result().Body).Decode(&env); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	got := env.Error.Field
	want := "name"
	if got != want {
		t.Errorf("expected field %q, got %q", want, got)
	}
}

func TestThatWriteErrorFromErrReturnsValidationErrorCodeForValidationError(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(context.WithValue(r.Context(), requestIDKey, "req-1"))
	w := httptest.NewRecorder()

	writeErrorFromErr(w, r, &dtos.ValidationError{Field: "name", Message: "bad"})

	var env apiErrorEnvelopeForTest
	if err := json.NewDecoder(w.Result().Body).Decode(&env); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	got := env.Error.Code
	want := "validation_error"
	if got != want {
		t.Errorf("expected code %q, got %q", want, got)
	}
}

func TestThatWriteErrorFromErrReturnsValidationErrorMessageForValidationError(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(context.WithValue(r.Context(), requestIDKey, "req-1"))
	w := httptest.NewRecorder()

	writeErrorFromErr(w, r, &dtos.ValidationError{Field: "name", Message: "bad"})

	var env apiErrorEnvelopeForTest
	if err := json.NewDecoder(w.Result().Body).Decode(&env); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	got := env.Error.Message
	want := "bad"
	if got != want {
		t.Errorf("expected message %q, got %q", want, got)
	}
}

func TestThatWriteErrorFromErrReturnsValidationErrorFieldForValidationError(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(context.WithValue(r.Context(), requestIDKey, "req-1"))
	w := httptest.NewRecorder()

	writeErrorFromErr(w, r, &dtos.ValidationError{Field: "name", Message: "bad"})

	var env apiErrorEnvelopeForTest
	if err := json.NewDecoder(w.Result().Body).Decode(&env); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	got := env.Error.Field
	want := "name"
	if got != want {
		t.Errorf("expected field %q, got %q", want, got)
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

func TestThatWriteErrorFromErrReturns400ForInvalidArgumentError(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(context.WithValue(r.Context(), requestIDKey, "req-1"))
	w := httptest.NewRecorder()

	writeErrorFromErr(w, r, &apperrors.InvalidArgumentError{Field: "seed", Message: "invalid"})

	got := w.Result().StatusCode
	want := http.StatusBadRequest
	if got != want {
		t.Errorf("expected status %d, got %d", want, got)
	}
}

func TestThatWriteErrorFromErrReturnsInvalidArgumentFieldForInvalidArgumentError(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(context.WithValue(r.Context(), requestIDKey, "req-1"))
	w := httptest.NewRecorder()

	writeErrorFromErr(w, r, &apperrors.InvalidArgumentError{Field: "seed", Message: "invalid"})

	var env apiErrorEnvelopeForTest
	if err := json.NewDecoder(w.Result().Body).Decode(&env); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	got := env.Error.Field
	want := "seed"
	if got != want {
		t.Errorf("expected field %q, got %q", want, got)
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
