package httpserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gorilla/mux"
)

type apiErrorEnvelopeForRunIDValidationTest struct {
	Error struct {
		Code      string `json:"code"`
		Message   string `json:"message"`
		Field     string `json:"field,omitempty"`
		RequestID string `json:"requestId"`
	} `json:"error"`
}

type handlerErrorResult struct {
	Status  int
	Code    string
	Message string
	Field   string
}

func decodeHandlerErrorResult(t *testing.T, r *http.Response) handlerErrorResult {
	t.Helper()

	var env apiErrorEnvelopeForRunIDValidationTest
	_ = json.NewDecoder(r.Body).Decode(&env)

	return handlerErrorResult{
		Status:  r.StatusCode,
		Code:    env.Error.Code,
		Message: env.Error.Message,
		Field:   env.Error.Field,
	}
}

func TestThatPredictedReturnsHandlerReturns400WhenGameOutcomeRunIDIsMissing(t *testing.T) {
	// GIVEN a request missing required game_outcome_run_id
	s := &Server{}
	req := httptest.NewRequest(http.MethodGet, "/analytics/calcuttas/abc/predicted-returns", nil)
	req = req.WithContext(context.WithValue(req.Context(), requestIDKey, "req-1"))
	req = mux.SetURLVars(req, map[string]string{"id": "abc"})
	w := httptest.NewRecorder()

	// WHEN calling the handler
	s.handleGetCalcuttaPredictedReturns(w, req)

	// THEN we get a 400 validation_error on game_outcome_run_id
	got := decodeHandlerErrorResult(t, w.Result())
	want := handlerErrorResult{
		Status:  http.StatusBadRequest,
		Code:    "validation_error",
		Message: "game_outcome_run_id is required",
		Field:   "game_outcome_run_id",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("unexpected error result: got=%+v want=%+v", got, want)
	}
}

func TestThatPredictedInvestmentHandlerReturns400WhenMarketShareRunIDIsMissing(t *testing.T) {
	// GIVEN a request missing required market_share_run_id
	s := &Server{}
	req := httptest.NewRequest(http.MethodGet, "/analytics/calcuttas/abc/predicted-investment?game_outcome_run_id=go1", nil)
	req = req.WithContext(context.WithValue(req.Context(), requestIDKey, "req-1"))
	req = mux.SetURLVars(req, map[string]string{"id": "abc"})
	w := httptest.NewRecorder()

	// WHEN calling the handler
	s.handleGetCalcuttaPredictedInvestment(w, req)

	// THEN we get a 400 validation_error on market_share_run_id
	got := decodeHandlerErrorResult(t, w.Result())
	want := handlerErrorResult{
		Status:  http.StatusBadRequest,
		Code:    "validation_error",
		Message: "market_share_run_id is required",
		Field:   "market_share_run_id",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("unexpected error result: got=%+v want=%+v", got, want)
	}
}

func TestThatPredictedInvestmentHandlerReturns400WhenGameOutcomeRunIDIsMissing(t *testing.T) {
	// GIVEN a request missing required game_outcome_run_id
	s := &Server{}
	req := httptest.NewRequest(http.MethodGet, "/analytics/calcuttas/abc/predicted-investment?market_share_run_id=ms1", nil)
	req = req.WithContext(context.WithValue(req.Context(), requestIDKey, "req-1"))
	req = mux.SetURLVars(req, map[string]string{"id": "abc"})
	w := httptest.NewRecorder()

	// WHEN calling the handler
	s.handleGetCalcuttaPredictedInvestment(w, req)

	// THEN we get a 400 validation_error on game_outcome_run_id
	got := decodeHandlerErrorResult(t, w.Result())
	want := handlerErrorResult{
		Status:  http.StatusBadRequest,
		Code:    "validation_error",
		Message: "game_outcome_run_id is required",
		Field:   "game_outcome_run_id",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("unexpected error result: got=%+v want=%+v", got, want)
	}
}

func TestThatPredictedMarketShareHandlerReturns400WhenMarketShareRunIDIsMissing(t *testing.T) {
	// GIVEN a request missing required market_share_run_id
	s := &Server{}
	req := httptest.NewRequest(http.MethodGet, "/analytics/calcuttas/abc/predicted-market-share?game_outcome_run_id=go1", nil)
	req = req.WithContext(context.WithValue(req.Context(), requestIDKey, "req-1"))
	req = mux.SetURLVars(req, map[string]string{"id": "abc"})
	w := httptest.NewRecorder()

	// WHEN calling the handler
	s.handleGetCalcuttaPredictedMarketShare(w, req)

	// THEN we get a 400 validation_error on market_share_run_id
	got := decodeHandlerErrorResult(t, w.Result())
	want := handlerErrorResult{
		Status:  http.StatusBadRequest,
		Code:    "validation_error",
		Message: "market_share_run_id is required",
		Field:   "market_share_run_id",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("unexpected error result: got=%+v want=%+v", got, want)
	}
}

func TestThatPredictedMarketShareHandlerReturns400WhenGameOutcomeRunIDIsMissing(t *testing.T) {
	// GIVEN a request missing required game_outcome_run_id
	s := &Server{}
	req := httptest.NewRequest(http.MethodGet, "/analytics/calcuttas/abc/predicted-market-share?market_share_run_id=ms1", nil)
	req = req.WithContext(context.WithValue(req.Context(), requestIDKey, "req-1"))
	req = mux.SetURLVars(req, map[string]string{"id": "abc"})
	w := httptest.NewRecorder()

	// WHEN calling the handler
	s.handleGetCalcuttaPredictedMarketShare(w, req)

	// THEN we get a 400 validation_error on game_outcome_run_id
	got := decodeHandlerErrorResult(t, w.Result())
	want := handlerErrorResult{
		Status:  http.StatusBadRequest,
		Code:    "validation_error",
		Message: "game_outcome_run_id is required",
		Field:   "game_outcome_run_id",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("unexpected error result: got=%+v want=%+v", got, want)
	}
}

func TestThatPredictedAdvancementHandlerReturns400WhenGameOutcomeRunIDIsMissing(t *testing.T) {
	// GIVEN a request missing required game_outcome_run_id
	s := &Server{}
	req := httptest.NewRequest(http.MethodGet, "/analytics/tournaments/abc/predicted-advancement", nil)
	req = req.WithContext(context.WithValue(req.Context(), requestIDKey, "req-1"))
	req = mux.SetURLVars(req, map[string]string{"id": "abc"})
	w := httptest.NewRecorder()

	// WHEN calling the handler
	s.handleGetTournamentPredictedAdvancement(w, req)

	// THEN we get a 400 validation_error on game_outcome_run_id
	got := decodeHandlerErrorResult(t, w.Result())
	want := handlerErrorResult{
		Status:  http.StatusBadRequest,
		Code:    "validation_error",
		Message: "game_outcome_run_id is required",
		Field:   "game_outcome_run_id",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("unexpected error result: got=%+v want=%+v", got, want)
	}
}

func TestThatSimulatedEntryHandlerReturns400WhenEntryRunIDIsMissing(t *testing.T) {
	// GIVEN a request missing required entry_run_id
	s := &Server{}
	req := httptest.NewRequest(http.MethodGet, "/analytics/calcuttas/abc/simulated-entry", nil)
	req = req.WithContext(context.WithValue(req.Context(), requestIDKey, "req-1"))
	req = mux.SetURLVars(req, map[string]string{"id": "abc"})
	w := httptest.NewRecorder()

	// WHEN calling the handler
	s.handleGetCalcuttaSimulatedEntry(w, req)

	// THEN we get a 400 validation_error on entry_run_id
	got := decodeHandlerErrorResult(t, w.Result())
	want := handlerErrorResult{
		Status:  http.StatusBadRequest,
		Code:    "validation_error",
		Message: "entry_run_id is required",
		Field:   "entry_run_id",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("unexpected error result: got=%+v want=%+v", got, want)
	}
}
