package httpserver

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/andrewcopp/Calcutta/backend/internal/app/strategy_runs"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type createEntryRunRequest struct {
	CalcuttaID            string  `json:"calcuttaId"`
	SyntheticCalcuttaID   *string `json:"syntheticCalcuttaId"`
	Name                  *string `json:"name"`
	OptimizerKey          *string `json:"optimizerKey"`
	MarketShareArtifactID *string `json:"marketShareArtifactId"`
	MarketShareRunID      *string `json:"marketShareRunId"`
}

type createEntryRunResponse struct {
	EntryRunID          string  `json:"entry_run_id"`
	RunKey              string  `json:"run_key"`
	SyntheticCalcuttaID *string `json:"synthetic_calcutta_id,omitempty"`
	CalcuttaSnapshotID  *string `json:"calcutta_snapshot_id,omitempty"`
}

type entryRunListItem struct {
	ID                    string  `json:"id"`
	RunKey                *string `json:"run_key,omitempty"`
	Name                  *string `json:"name,omitempty"`
	CalcuttaID            *string `json:"calcutta_id,omitempty"`
	SimulatedTournamentID *string `json:"simulated_tournament_id,omitempty"`
	Purpose               string  `json:"purpose"`
	ReturnsModelKey       string  `json:"returns_model_key"`
	InvestmentModelKey    string  `json:"investment_model_key"`
	OptimizerKey          string  `json:"optimizer_key"`
	CreatedAt             string  `json:"created_at"`
}

type listEntryRunsResponse struct {
	Items []entryRunListItem `json:"items"`
}

type entryRunArtifactListItem struct {
	ID                         string          `json:"id"`
	RunID                      string          `json:"run_id"`
	RunKey                     *string         `json:"run_key,omitempty"`
	ArtifactKind               string          `json:"artifact_kind"`
	SchemaVersion              string          `json:"schema_version"`
	StorageURI                 *string         `json:"storage_uri,omitempty"`
	SummaryJSON                json.RawMessage `json:"summary_json"`
	InputMarketShareArtifactID *string         `json:"input_market_share_artifact_id,omitempty"`
	InputAdvancementArtifactID *string         `json:"input_advancement_artifact_id,omitempty"`
	CreatedAt                  string          `json:"created_at"`
	UpdatedAt                  string          `json:"updated_at"`
}

type listEntryRunArtifactsResponse struct {
	Items []entryRunArtifactListItem `json:"items"`
}

func (s *Server) registerEntryArtifactRoutes(r *mux.Router) {
	r.HandleFunc(
		"/api/entry-artifacts/{id}",
		s.requirePermission("analytics.strategy_generation_runs.read", s.getEntryArtifactHandler),
	).Methods("GET", "OPTIONS")
}

func (s *Server) registerEntryRunRoutes(r *mux.Router) {
	r.HandleFunc(
		"/api/entry-runs",
		s.requirePermission("analytics.strategy_generation_runs.write", s.createEntryRunHandler),
	).Methods("POST", "OPTIONS")
	r.HandleFunc(
		"/api/entry-runs",
		s.requirePermission("analytics.strategy_generation_runs.read", s.listEntryRunsHandler),
	).Methods("GET", "OPTIONS")
	r.HandleFunc(
		"/api/entry-runs/{id}",
		s.requirePermission("analytics.strategy_generation_runs.read", s.getEntryRunHandler),
	).Methods("GET", "OPTIONS")
	r.HandleFunc(
		"/api/entry-runs/{id}/artifacts",
		s.requirePermission("analytics.strategy_generation_runs.read", s.listEntryRunArtifactsHandler),
	).Methods("GET", "OPTIONS")
	r.HandleFunc(
		"/api/entry-runs/{id}/artifacts/{artifactKind}",
		s.requirePermission("analytics.strategy_generation_runs.read", s.getEntryRunArtifactHandler),
	).Methods("GET", "OPTIONS")
}

func (s *Server) createEntryRunHandler(w http.ResponseWriter, r *http.Request) {
	var req createEntryRunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}

	req.CalcuttaID = strings.TrimSpace(req.CalcuttaID)
	if req.CalcuttaID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "calcuttaId is required", "calcuttaId")
		return
	}
	if _, err := uuid.Parse(req.CalcuttaID); err != nil {
		writeError(w, r, http.StatusBadRequest, "validation_error", "calcuttaId must be a valid UUID", "calcuttaId")
		return
	}

	syntheticCalcuttaID := ""
	if req.SyntheticCalcuttaID != nil {
		v := strings.TrimSpace(*req.SyntheticCalcuttaID)
		if v != "" {
			if _, err := uuid.Parse(v); err != nil {
				writeError(w, r, http.StatusBadRequest, "validation_error", "syntheticCalcuttaId must be a valid UUID", "syntheticCalcuttaId")
				return
			}
			syntheticCalcuttaID = v
		}
	}

	name := "api_entry_run"
	if req.Name != nil {
		v := strings.TrimSpace(*req.Name)
		if v != "" {
			name = v
		}
	}

	optimizerKey := ""
	if req.OptimizerKey != nil {
		optimizerKey = strings.TrimSpace(*req.OptimizerKey)
	}
	if optimizerKey == "" {
		optimizerKey = "minlp_v1"
	}

	marketShareArtifactID := (*string)(nil)
	if req.MarketShareArtifactID != nil {
		v := strings.TrimSpace(*req.MarketShareArtifactID)
		if v != "" {
			if _, err := uuid.Parse(v); err != nil {
				writeError(w, r, http.StatusBadRequest, "validation_error", "marketShareArtifactId must be a valid UUID", "marketShareArtifactId")
				return
			}
			marketShareArtifactID = &v
		}
	}

	marketShareRunID := (*string)(nil)
	if req.MarketShareRunID != nil {
		v := strings.TrimSpace(*req.MarketShareRunID)
		if v != "" {
			if _, err := uuid.Parse(v); err != nil {
				writeError(w, r, http.StatusBadRequest, "validation_error", "marketShareRunId must be a valid UUID", "marketShareRunId")
				return
			}
			marketShareRunID = &v
		}
	}

	ctx := r.Context()
	if marketShareArtifactID == nil && marketShareRunID == nil {
		writeError(w, r, http.StatusBadRequest, "validation_error", "marketShareArtifactId is required", "marketShareArtifactId")
		return
	}
	runKeyUUID := uuid.NewString()
	runKeyText := runKeyUUID

	gitSHA := strings.TrimSpace(os.Getenv("GIT_SHA"))
	var gitSHAPtr *string
	if gitSHA != "" {
		gitSHAPtr = &gitSHA
	}
	createParams := strategy_runs.CreateRunParams{
		CalcuttaID: req.CalcuttaID,
		SyntheticCalcuttaID: func() *string {
			if syntheticCalcuttaID == "" {
				return nil
			}
			v := syntheticCalcuttaID
			return &v
		}(),
		Name:                  name,
		OptimizerKey:          optimizerKey,
		RunKey:                runKeyText,
		RunKeyUUID:            runKeyUUID,
		MarketShareArtifactID: marketShareArtifactID,
		MarketShareRunID:      marketShareRunID,
		GitSHA:                gitSHAPtr,
		Source:                "api_entry_run",
	}

	created, err := s.app.StrategyRuns.CreateEntryRun(ctx, createParams)
	if err != nil {
		if errors.Is(err, strategy_runs.ErrMarketShareArtifactNotFound) {
			writeError(w, r, http.StatusNotFound, "not_found", "Market share artifact not found", "marketShareArtifactId")
			return
		}
		if errors.Is(err, strategy_runs.ErrMarketShareRunMissingMetrics) {
			writeError(w, r, http.StatusBadRequest, "validation_error", "marketShareRunId must refer to a market share run with an available metrics artifact; pass marketShareArtifactId instead", "marketShareRunId")
			return
		}
		if errors.Is(err, strategy_runs.ErrSyntheticCalcuttaNotFound) {
			writeError(w, r, http.StatusNotFound, "not_found", "Synthetic calcutta not found", "syntheticCalcuttaId")
			return
		}
		if errors.Is(err, strategy_runs.ErrSyntheticCalcuttaMismatch) {
			writeError(w, r, http.StatusBadRequest, "validation_error", "syntheticCalcuttaId does not match calcuttaId", "syntheticCalcuttaId")
			return
		}
		writeErrorFromErr(w, r, err)
		return
	}

	resp := createEntryRunResponse{EntryRunID: created.RunID, RunKey: created.RunKey, SyntheticCalcuttaID: created.SyntheticCalcuttaID, CalcuttaSnapshotID: created.CalcuttaSnapshotID}
	writeJSON(w, http.StatusCreated, resp)
}

func (s *Server) listEntryRunsHandler(w http.ResponseWriter, r *http.Request) {
	calcuttaID := strings.TrimSpace(r.URL.Query().Get("calcutta_id"))
	if calcuttaID != "" {
		if _, err := uuid.Parse(calcuttaID); err != nil {
			writeError(w, r, http.StatusBadRequest, "validation_error", "calcutta_id must be a valid UUID", "calcutta_id")
			return
		}
	}

	limit := getLimit(r, 50)
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	offset := getOffset(r, 0)
	if offset < 0 {
		offset = 0
	}

	var calcuttaIDPtr *string
	if calcuttaID != "" {
		v := calcuttaID
		calcuttaIDPtr = &v
	}
	runs, err := s.app.StrategyRuns.ListRuns(r.Context(), calcuttaIDPtr, limit, offset)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	items := make([]entryRunListItem, 0, len(runs))
	for _, it := range runs {
		items = append(items, entryRunListItem{
			ID:                    it.ID,
			RunKey:                it.RunKey,
			Name:                  it.Name,
			CalcuttaID:            it.CalcuttaID,
			SimulatedTournamentID: it.SimulatedTournamentID,
			Purpose:               it.Purpose,
			ReturnsModelKey:       it.ReturnsModelKey,
			InvestmentModelKey:    it.InvestmentModelKey,
			OptimizerKey:          it.OptimizerKey,
			CreatedAt:             it.CreatedAt,
		})
	}
	writeJSON(w, http.StatusOK, listEntryRunsResponse{Items: items})
}

func (s *Server) getEntryRunHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := strings.TrimSpace(vars["id"])
	if id == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "id is required", "id")
		return
	}
	if _, err := uuid.Parse(id); err != nil {
		writeError(w, r, http.StatusBadRequest, "validation_error", "id must be a valid UUID", "id")
		return
	}

	it, err := s.app.StrategyRuns.GetRun(r.Context(), id)
	if err != nil {
		if errors.Is(err, strategy_runs.ErrEntryRunNotFound) || errors.Is(err, strategy_runs.ErrStrategyGenerationRunNotFound) {
			writeError(w, r, http.StatusNotFound, "not_found", "Entry run not found", "id")
			return
		}
		writeErrorFromErr(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, entryRunListItem{
		ID:                    it.ID,
		RunKey:                it.RunKey,
		Name:                  it.Name,
		CalcuttaID:            it.CalcuttaID,
		SimulatedTournamentID: it.SimulatedTournamentID,
		Purpose:               it.Purpose,
		ReturnsModelKey:       it.ReturnsModelKey,
		InvestmentModelKey:    it.InvestmentModelKey,
		OptimizerKey:          it.OptimizerKey,
		CreatedAt:             it.CreatedAt,
	})
}

func (s *Server) getEntryArtifactHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	artifactID := strings.TrimSpace(vars["id"])
	if artifactID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "id is required", "id")
		return
	}
	if _, err := uuid.Parse(artifactID); err != nil {
		writeError(w, r, http.StatusBadRequest, "validation_error", "id must be a valid UUID", "id")
		return
	}

	it, err := s.app.StrategyRuns.GetEntryArtifact(r.Context(), artifactID)
	if err != nil {
		if errors.Is(err, strategy_runs.ErrEntryArtifactNotFound) {
			writeError(w, r, http.StatusNotFound, "not_found", "Entry artifact not found", "id")
			return
		}
		writeErrorFromErr(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, entryRunArtifactListItem{
		ID:                         it.ID,
		RunID:                      it.RunID,
		RunKey:                     it.RunKey,
		ArtifactKind:               it.ArtifactKind,
		SchemaVersion:              it.SchemaVersion,
		StorageURI:                 it.StorageURI,
		SummaryJSON:                it.SummaryJSON,
		InputMarketShareArtifactID: it.InputMarketShareArtifactID,
		InputAdvancementArtifactID: it.InputAdvancementArtifactID,
		CreatedAt:                  it.CreatedAt,
		UpdatedAt:                  it.UpdatedAt,
	})
}

func (s *Server) listEntryRunArtifactsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	runID := strings.TrimSpace(vars["id"])
	if runID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "id is required", "id")
		return
	}
	if _, err := uuid.Parse(runID); err != nil {
		writeError(w, r, http.StatusBadRequest, "validation_error", "id must be a valid UUID", "id")
		return
	}

	arts, err := s.app.StrategyRuns.ListRunArtifacts(r.Context(), runID)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	items := make([]entryRunArtifactListItem, 0, len(arts))
	for _, it := range arts {
		items = append(items, entryRunArtifactListItem{
			ID:                         it.ID,
			RunID:                      it.RunID,
			RunKey:                     it.RunKey,
			ArtifactKind:               it.ArtifactKind,
			SchemaVersion:              it.SchemaVersion,
			StorageURI:                 it.StorageURI,
			SummaryJSON:                it.SummaryJSON,
			InputMarketShareArtifactID: it.InputMarketShareArtifactID,
			InputAdvancementArtifactID: it.InputAdvancementArtifactID,
			CreatedAt:                  it.CreatedAt,
			UpdatedAt:                  it.UpdatedAt,
		})
	}
	writeJSON(w, http.StatusOK, listEntryRunArtifactsResponse{Items: items})
}

func (s *Server) getEntryRunArtifactHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	runID := strings.TrimSpace(vars["id"])
	if runID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "id is required", "id")
		return
	}
	if _, err := uuid.Parse(runID); err != nil {
		writeError(w, r, http.StatusBadRequest, "validation_error", "id must be a valid UUID", "id")
		return
	}

	artifactKind := strings.TrimSpace(vars["artifactKind"])
	if artifactKind == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "artifactKind is required", "artifactKind")
		return
	}

	it, err := s.app.StrategyRuns.GetRunArtifact(r.Context(), runID, artifactKind)
	if err != nil {
		if errors.Is(err, strategy_runs.ErrEntryArtifactNotFound) {
			writeError(w, r, http.StatusNotFound, "not_found", "Entry artifact not found", "artifactKind")
			return
		}
		writeErrorFromErr(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, entryRunArtifactListItem{
		ID:                         it.ID,
		RunID:                      it.RunID,
		RunKey:                     it.RunKey,
		ArtifactKind:               it.ArtifactKind,
		SchemaVersion:              it.SchemaVersion,
		StorageURI:                 it.StorageURI,
		SummaryJSON:                it.SummaryJSON,
		InputMarketShareArtifactID: it.InputMarketShareArtifactID,
		InputAdvancementArtifactID: it.InputAdvancementArtifactID,
		CreatedAt:                  it.CreatedAt,
		UpdatedAt:                  it.UpdatedAt,
	})
}
