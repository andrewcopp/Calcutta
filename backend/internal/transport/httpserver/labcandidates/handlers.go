package labcandidates

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/andrewcopp/Calcutta/backend/internal/app"
	applabcandidates "github.com/andrewcopp/Calcutta/backend/internal/app/lab_candidates"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/httperr"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/response"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type Handler struct {
	app        *app.App
	authUserID func(context.Context) string
}

func NewHandler(a *app.App) *Handler {
	return &Handler{app: a}
}

func NewHandlerWithAuthUserID(a *app.App, authUserID func(context.Context) string) *Handler {
	return &Handler{app: a, authUserID: authUserID}
}

type listLabCandidatesResponse struct {
	Items []labCandidateDetailResponse `json:"items"`
}

type labCandidateDetailTeam struct {
	TeamID    string `json:"team_id"`
	BidPoints int    `json:"bid_points"`
}

type labCandidateDetailResponse struct {
	CandidateID             string                   `json:"candidate_id"`
	DisplayName             string                   `json:"display_name"`
	SourceKind              string                   `json:"source_kind"`
	SourceEntryArtifactID   *string                  `json:"source_entry_artifact_id,omitempty"`
	CalcuttaID              string                   `json:"calcutta_id"`
	TournamentID            string                   `json:"tournament_id"`
	StrategyGenerationRunID string                   `json:"strategy_generation_run_id"`
	MarketShareRunID        string                   `json:"market_share_run_id"`
	MarketShareArtifactID   string                   `json:"market_share_artifact_id"`
	AdvancementRunID        string                   `json:"advancement_run_id"`
	OptimizerKey            string                   `json:"optimizer_key"`
	StartingStateKey        string                   `json:"starting_state_key"`
	ExcludedEntryName       *string                  `json:"excluded_entry_name,omitempty"`
	GitSHA                  *string                  `json:"git_sha,omitempty"`
	Teams                   []labCandidateDetailTeam `json:"teams"`
}

type createLabCandidateRequest struct {
	CalcuttaID            string  `json:"calcuttaId"`
	AdvancementRunID      string  `json:"advancementRunId"`
	MarketShareArtifactID string  `json:"marketShareArtifactId"`
	OptimizerKey          string  `json:"optimizerKey"`
	StartingStateKey      string  `json:"startingStateKey"`
	ExcludedEntryName     *string `json:"excludedEntryName"`
	DisplayName           *string `json:"displayName"`
}

type createLabCandidatesBulkRequest struct {
	Items []createLabCandidateRequest `json:"items"`
}

type createLabCandidateResponse struct {
	CandidateID             string `json:"candidateId"`
	StrategyGenerationRunID string `json:"strategyGenerationRunId"`
}

type createLabCandidatesBulkResponse struct {
	Items []createLabCandidateResponse `json:"items"`
}

func readBodyLimit(r *http.Request, maxBytes int64) ([]byte, error) {
	if r == nil || r.Body == nil {
		return []byte{}, nil
	}
	if maxBytes <= 0 {
		return io.ReadAll(r.Body)
	}

	limited := io.LimitReader(r.Body, maxBytes+1)
	b, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if int64(len(b)) > maxBytes {
		return nil, &apperrBodyTooLarge{}
	}
	return b, nil
}

type apperrBodyTooLarge struct{}

func (e *apperrBodyTooLarge) Error() string { return "request body too large" }

func getQueryInt(r *http.Request, name string, defaultValue int) int {
	v := strings.TrimSpace(r.URL.Query().Get(name))
	if v == "" {
		return defaultValue
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return defaultValue
	}
	return n
}

func (h *Handler) HandleListLabCandidates(w http.ResponseWriter, r *http.Request) {
	if h.app == nil || h.app.LabCandidates == nil {
		httperr.Write(w, r, http.StatusInternalServerError, "internal_error", "internal server error", "")
		return
	}

	calcuttaID := strings.TrimSpace(r.URL.Query().Get("calcutta_id"))
	tournamentID := strings.TrimSpace(r.URL.Query().Get("tournament_id"))
	strategyGenerationRunID := strings.TrimSpace(r.URL.Query().Get("strategy_generation_run_id"))
	marketShareArtifactID := strings.TrimSpace(r.URL.Query().Get("market_share_artifact_id"))
	advancementRunID := strings.TrimSpace(r.URL.Query().Get("advancement_run_id"))
	optimizerKey := strings.TrimSpace(r.URL.Query().Get("optimizer_key"))
	startingStateKey := strings.TrimSpace(r.URL.Query().Get("starting_state_key"))
	excludedEntryName := strings.TrimSpace(r.URL.Query().Get("excluded_entry_name"))
	sourceKind := strings.TrimSpace(r.URL.Query().Get("source_kind"))

	filter := applabcandidates.ListCandidatesFilter{}
	if calcuttaID != "" {
		filter.CalcuttaID = &calcuttaID
	}
	if tournamentID != "" {
		filter.TournamentID = &tournamentID
	}
	if strategyGenerationRunID != "" {
		filter.StrategyGenerationRunID = &strategyGenerationRunID
	}
	if marketShareArtifactID != "" {
		filter.MarketShareArtifactID = &marketShareArtifactID
	}
	if advancementRunID != "" {
		filter.AdvancementRunID = &advancementRunID
	}
	if optimizerKey != "" {
		filter.OptimizerKey = &optimizerKey
	}
	if startingStateKey != "" {
		filter.StartingStateKey = &startingStateKey
	}
	if excludedEntryName != "" {
		filter.ExcludedEntryName = &excludedEntryName
	}
	if sourceKind != "" {
		filter.SourceKind = &sourceKind
	}

	page := applabcandidates.ListCandidatesPagination{Limit: getQueryInt(r, "limit", 50), Offset: getQueryInt(r, "offset", 0)}

	items, err := h.app.LabCandidates.ListCandidates(r.Context(), filter, page)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	out := make([]labCandidateDetailResponse, 0, len(items))
	for i := range items {
		it := items[i]
		out = append(out, labCandidateDetailResponse{
			CandidateID:             it.CandidateID,
			DisplayName:             it.DisplayName,
			SourceKind:              it.SourceKind,
			SourceEntryArtifactID:   it.SourceEntryArtifactID,
			CalcuttaID:              it.CalcuttaID,
			TournamentID:            it.TournamentID,
			StrategyGenerationRunID: it.StrategyGenerationRunID,
			MarketShareRunID:        it.MarketShareRunID,
			MarketShareArtifactID:   it.MarketShareArtifactID,
			AdvancementRunID:        it.AdvancementRunID,
			OptimizerKey:            it.OptimizerKey,
			StartingStateKey:        it.StartingStateKey,
			ExcludedEntryName:       it.ExcludedEntryName,
			GitSHA:                  it.GitSHA,
			Teams:                   nil,
		})
	}

	response.WriteJSON(w, http.StatusOK, listLabCandidatesResponse{Items: out})
}

func (h *Handler) HandleGetLabCandidateDetail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	candidateID := strings.TrimSpace(vars["candidateId"])
	if candidateID == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "candidateId is required", "candidateId")
		return
	}
	if _, err := uuid.Parse(candidateID); err != nil {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "candidateId must be a valid UUID", "candidateId")
		return
	}
	if h.app == nil || h.app.LabCandidates == nil {
		httperr.Write(w, r, http.StatusInternalServerError, "internal_error", "internal server error", "")
		return
	}

	detail, err := h.app.LabCandidates.GetCandidateDetail(r.Context(), candidateID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	resp := labCandidateDetailResponse{
		CandidateID:             detail.CandidateID,
		DisplayName:             detail.DisplayName,
		SourceKind:              detail.SourceKind,
		SourceEntryArtifactID:   detail.SourceEntryArtifactID,
		CalcuttaID:              detail.CalcuttaID,
		TournamentID:            detail.TournamentID,
		StrategyGenerationRunID: detail.StrategyGenerationRunID,
		MarketShareRunID:        detail.MarketShareRunID,
		MarketShareArtifactID:   detail.MarketShareArtifactID,
		AdvancementRunID:        detail.AdvancementRunID,
		OptimizerKey:            detail.OptimizerKey,
		StartingStateKey:        detail.StartingStateKey,
		ExcludedEntryName:       detail.ExcludedEntryName,
		GitSHA:                  detail.GitSHA,
		Teams:                   make([]labCandidateDetailTeam, 0),
	}
	for _, t := range detail.Teams {
		resp.Teams = append(resp.Teams, labCandidateDetailTeam{TeamID: t.TeamID, BidPoints: t.BidPoints})
	}

	response.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) HandleDeleteLabCandidate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	candidateID := strings.TrimSpace(vars["candidateId"])
	if candidateID == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "candidateId is required", "candidateId")
		return
	}
	if _, err := uuid.Parse(candidateID); err != nil {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "candidateId must be a valid UUID", "candidateId")
		return
	}
	if h.app == nil || h.app.LabCandidates == nil {
		httperr.Write(w, r, http.StatusInternalServerError, "internal_error", "internal server error", "")
		return
	}

	if err := h.app.LabCandidates.DeleteCandidate(r.Context(), candidateID); err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) HandleCreateLabCandidates(w http.ResponseWriter, r *http.Request) {
	if h.app == nil || h.app.LabCandidates == nil {
		httperr.Write(w, r, http.StatusInternalServerError, "internal_error", "internal server error", "")
		return
	}

	body, err := readBodyLimit(r, 2<<20)
	if err != nil {
		if _, ok := err.(*apperrBodyTooLarge); ok {
			httperr.Write(w, r, http.StatusBadRequest, "validation_error", "request body too large", "")
			return
		}
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	bulk := createLabCandidatesBulkRequest{}
	if err := json.Unmarshal(body, &bulk); err == nil && len(bulk.Items) > 0 {
		resp, err := h.createLabCandidatesBulk(r.Context(), bulk.Items)
		if err != nil {
			httperr.WriteFromErr(w, r, err, h.authUserID)
			return
		}
		response.WriteJSON(w, http.StatusCreated, resp)
		return
	}

	single := createLabCandidateRequest{}
	if err := json.Unmarshal(body, &single); err != nil {
		httperr.Write(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}

	resp, err := h.createLabCandidatesBulk(r.Context(), []createLabCandidateRequest{single})
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}
	if len(resp.Items) != 1 {
		httperr.Write(w, r, http.StatusInternalServerError, "internal_error", "Unexpected create result", "")
		return
	}
	response.WriteJSON(w, http.StatusCreated, resp.Items[0])
}

func (h *Handler) createLabCandidatesBulk(ctx context.Context, items []createLabCandidateRequest) (*createLabCandidatesBulkResponse, error) {
	normalized := make([]applabcandidates.CreateCandidateRequest, 0, len(items))
	for i := range items {
		req := items[i]
		req.CalcuttaID = strings.TrimSpace(req.CalcuttaID)
		req.AdvancementRunID = strings.TrimSpace(req.AdvancementRunID)
		req.MarketShareArtifactID = strings.TrimSpace(req.MarketShareArtifactID)
		req.OptimizerKey = strings.TrimSpace(req.OptimizerKey)
		req.StartingStateKey = strings.TrimSpace(req.StartingStateKey)
		if req.ExcludedEntryName != nil {
			v := strings.TrimSpace(*req.ExcludedEntryName)
			req.ExcludedEntryName = &v
			if v == "" {
				req.ExcludedEntryName = nil
			}
		}
		if req.DisplayName != nil {
			v := strings.TrimSpace(*req.DisplayName)
			req.DisplayName = &v
			if v == "" {
				req.DisplayName = nil
			}
		}

		normalized = append(normalized, applabcandidates.CreateCandidateRequest{
			CalcuttaID:            req.CalcuttaID,
			AdvancementRunID:      req.AdvancementRunID,
			MarketShareArtifactID: req.MarketShareArtifactID,
			OptimizerKey:          req.OptimizerKey,
			StartingStateKey:      req.StartingStateKey,
			ExcludedEntryName:     req.ExcludedEntryName,
			DisplayName:           req.DisplayName,
		})
	}

	created, err := h.app.LabCandidates.CreateCandidatesBulk(ctx, normalized)
	if err != nil {
		return nil, err
	}

	out := make([]createLabCandidateResponse, 0, len(created))
	for i := range created {
		out = append(out, createLabCandidateResponse{CandidateID: created[i].CandidateID, StrategyGenerationRunID: created[i].StrategyGenerationRunID})
	}
	return &createLabCandidatesBulkResponse{Items: out}, nil
}
