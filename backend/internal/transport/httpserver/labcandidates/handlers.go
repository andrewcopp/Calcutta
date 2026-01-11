package labcandidates

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/andrewcopp/Calcutta/backend/internal/app"
	applabcandidates "github.com/andrewcopp/Calcutta/backend/internal/app/lab_candidates"
	appmodelcatalogs "github.com/andrewcopp/Calcutta/backend/internal/app/model_catalogs"
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
	CalcuttaName            string                   `json:"calcutta_name"`
	TournamentID            string                   `json:"tournament_id"`
	StrategyGenerationRunID string                   `json:"strategy_generation_run_id"`
	MarketShareRunID        string                   `json:"market_share_run_id"`
	MarketShareArtifactID   string                   `json:"market_share_artifact_id"`
	AdvancementRunID        string                   `json:"advancement_run_id"`
	OptimizerKey            string                   `json:"optimizer_key"`
	StartingStateKey        string                   `json:"starting_state_key"`
	SeedPreview             string                   `json:"seed_preview"`
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

type labCandidateComboItem struct {
	GameOutcomesAlgorithmID string                           `json:"game_outcomes_algorithm_id"`
	GameOutcomesName        string                           `json:"game_outcomes_name"`
	MarketShareAlgorithmID  string                           `json:"market_share_algorithm_id"`
	MarketShareName         string                           `json:"market_share_name"`
	OptimizerKey            string                           `json:"optimizer_key"`
	Optimizer               appmodelcatalogs.ModelDescriptor `json:"optimizer"`
	DisplayName             string                           `json:"display_name"`
	ExistingCandidates      int                              `json:"existing_candidates"`
	TotalCalcuttas          int                              `json:"total_calcuttas"`
}

type listLabCandidateCombosResponse struct {
	Items []labCandidateComboItem `json:"items"`
	Count int                     `json:"count"`
}

type generateLabCandidatesRequest struct {
	GameOutcomesAlgorithmID string  `json:"gameOutcomesAlgorithmId"`
	MarketShareAlgorithmID  string  `json:"marketShareAlgorithmId"`
	OptimizerKey            string  `json:"optimizerKey"`
	StartingStateKey        string  `json:"startingStateKey"`
	ExcludedEntryName       *string `json:"excludedEntryName"`
	DisplayName             string  `json:"displayName"`
}

type generateLabCandidatesResponse struct {
	TotalCalcuttas         int `json:"total_calcuttas"`
	EligibleCalcuttas      int `json:"eligible_calcuttas"`
	CreatedCandidates      int `json:"created_candidates"`
	SkippedExisting        int `json:"skipped_existing"`
	SkippedMissingUpstream int `json:"skipped_missing_upstream"`
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
	gameOutcomesAlgorithmID := strings.TrimSpace(r.URL.Query().Get("game_outcomes_algorithm_id"))
	if gameOutcomesAlgorithmID == "" {
		gameOutcomesAlgorithmID = strings.TrimSpace(r.URL.Query().Get("gameOutcomesAlgorithmId"))
	}
	marketShareAlgorithmID := strings.TrimSpace(r.URL.Query().Get("market_share_algorithm_id"))
	if marketShareAlgorithmID == "" {
		marketShareAlgorithmID = strings.TrimSpace(r.URL.Query().Get("marketShareAlgorithmId"))
	}
	optimizerKey := strings.TrimSpace(r.URL.Query().Get("optimizer_key"))
	if optimizerKey == "" {
		optimizerKey = strings.TrimSpace(r.URL.Query().Get("optimizerKey"))
	}
	startingStateKey := strings.TrimSpace(r.URL.Query().Get("starting_state_key"))
	if startingStateKey == "" {
		startingStateKey = strings.TrimSpace(r.URL.Query().Get("startingStateKey"))
	}
	excludedEntryName := strings.TrimSpace(r.URL.Query().Get("excluded_entry_name"))
	if excludedEntryName == "" {
		excludedEntryName = strings.TrimSpace(r.URL.Query().Get("excludedEntryName"))
	}
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
	if gameOutcomesAlgorithmID != "" {
		filter.GameOutcomesAlgorithmID = &gameOutcomesAlgorithmID
	}
	if marketShareAlgorithmID != "" {
		filter.MarketShareAlgorithmID = &marketShareAlgorithmID
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
			CalcuttaName:            it.CalcuttaName,
			TournamentID:            it.TournamentID,
			StrategyGenerationRunID: it.StrategyGenerationRunID,
			MarketShareRunID:        it.MarketShareRunID,
			MarketShareArtifactID:   it.MarketShareArtifactID,
			AdvancementRunID:        it.AdvancementRunID,
			OptimizerKey:            it.OptimizerKey,
			StartingStateKey:        it.StartingStateKey,
			SeedPreview:             it.SeedPreview,
			ExcludedEntryName:       it.ExcludedEntryName,
			GitSHA:                  it.GitSHA,
			Teams:                   nil,
		})
	}

	response.WriteJSON(w, http.StatusOK, listLabCandidatesResponse{Items: out})
}

func (h *Handler) HandleListCandidateCombos(w http.ResponseWriter, r *http.Request) {
	if h.app == nil || h.app.Analytics == nil || h.app.ModelCatalogs == nil || h.app.LabCandidates == nil || h.app.Calcutta == nil {
		httperr.Write(w, r, http.StatusInternalServerError, "internal_error", "internal server error", "")
		return
	}

	ctx := r.Context()
	startingStateKey := strings.TrimSpace(r.URL.Query().Get("starting_state_key"))
	excludedEntryName := strings.TrimSpace(r.URL.Query().Get("excluded_entry_name"))
	if startingStateKey == "" {
		startingStateKey = "post_first_four"
	}
	var excludedPtr *string
	if excludedEntryName != "" {
		v := excludedEntryName
		excludedPtr = &v
	}

	calcuttas, err := h.app.Calcutta.GetAllCalcuttas(ctx)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}
	totalCalcuttas := len(calcuttas)

	coverageSummary, err := h.app.LabCandidates.ListCandidateComboCoverage(ctx, startingStateKey, excludedPtr)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}
	coverageMap := make(map[string]int, len(coverageSummary.Items))
	for i := range coverageSummary.Items {
		it := coverageSummary.Items[i]
		k := it.GameOutcomesAlgorithmID + "|" + it.MarketShareAlgorithmID + "|" + it.OptimizerKey
		coverageMap[k] = it.ExistingCandidates
	}
	goKind := "game_outcomes"
	msKind := "market_share"

	goAlgos, err := h.app.Analytics.ListAlgorithms(ctx, &goKind)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}
	msAlgos, err := h.app.Analytics.ListAlgorithms(ctx, &msKind)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	optimizers := h.app.ModelCatalogs.ListEntryOptimizers()

	items := make([]labCandidateComboItem, 0)
	for _, goa := range goAlgos {
		for _, msa := range msAlgos {
			for _, opt := range optimizers {
				key := goa.ID + "|" + msa.ID + "|" + opt.ID
				existing := coverageMap[key]
				items = append(items, labCandidateComboItem{
					GameOutcomesAlgorithmID: goa.ID,
					GameOutcomesName:        goa.Name,
					MarketShareAlgorithmID:  msa.ID,
					MarketShareName:         msa.Name,
					OptimizerKey:            opt.ID,
					Optimizer:               opt,
					DisplayName:             strings.TrimSpace(goa.Name) + " - " + strings.TrimSpace(msa.Name) + " - " + strings.TrimSpace(opt.DisplayName),
					ExistingCandidates:      existing,
					TotalCalcuttas:          totalCalcuttas,
				})
			}
		}
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].GameOutcomesName != items[j].GameOutcomesName {
			return items[i].GameOutcomesName < items[j].GameOutcomesName
		}
		if items[i].MarketShareName != items[j].MarketShareName {
			return items[i].MarketShareName < items[j].MarketShareName
		}
		return items[i].OptimizerKey < items[j].OptimizerKey
	})

	response.WriteJSON(w, http.StatusOK, listLabCandidateCombosResponse{Items: items, Count: len(items)})
}

func (h *Handler) HandleGenerateCandidates(w http.ResponseWriter, r *http.Request) {
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

	var req generateLabCandidatesRequest
	if err := json.Unmarshal(body, &req); err != nil {
		httperr.Write(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}

	req.GameOutcomesAlgorithmID = strings.TrimSpace(req.GameOutcomesAlgorithmID)
	req.MarketShareAlgorithmID = strings.TrimSpace(req.MarketShareAlgorithmID)
	req.OptimizerKey = strings.TrimSpace(req.OptimizerKey)
	req.StartingStateKey = strings.TrimSpace(req.StartingStateKey)
	req.DisplayName = strings.TrimSpace(req.DisplayName)
	if req.StartingStateKey == "" {
		req.StartingStateKey = "post_first_four"
	}
	if req.ExcludedEntryName != nil {
		v := strings.TrimSpace(*req.ExcludedEntryName)
		req.ExcludedEntryName = &v
		if v == "" {
			req.ExcludedEntryName = nil
		}
	}

	if req.GameOutcomesAlgorithmID == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "gameOutcomesAlgorithmId is required", "gameOutcomesAlgorithmId")
		return
	}
	if _, err := uuid.Parse(req.GameOutcomesAlgorithmID); err != nil {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "gameOutcomesAlgorithmId must be a valid UUID", "gameOutcomesAlgorithmId")
		return
	}
	if req.MarketShareAlgorithmID == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "marketShareAlgorithmId is required", "marketShareAlgorithmId")
		return
	}
	if _, err := uuid.Parse(req.MarketShareAlgorithmID); err != nil {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "marketShareAlgorithmId must be a valid UUID", "marketShareAlgorithmId")
		return
	}

	res, err := h.app.LabCandidates.GenerateCandidatesFromAlgorithms(r.Context(), applabcandidates.GenerateCandidatesFromAlgorithmsRequest{
		GameOutcomesAlgorithmID: req.GameOutcomesAlgorithmID,
		MarketShareAlgorithmID:  req.MarketShareAlgorithmID,
		OptimizerKey:            req.OptimizerKey,
		StartingStateKey:        req.StartingStateKey,
		ExcludedEntryName:       req.ExcludedEntryName,
		DisplayName:             req.DisplayName,
	})
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	response.WriteJSON(w, http.StatusOK, generateLabCandidatesResponse{
		TotalCalcuttas:         res.TotalCalcuttas,
		EligibleCalcuttas:      res.EligibleCalcuttas,
		CreatedCandidates:      res.CreatedCandidates,
		SkippedExisting:        res.SkippedExisting,
		SkippedMissingUpstream: res.SkippedMissingUpstream,
	})
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
		CalcuttaName:            detail.CalcuttaName,
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
