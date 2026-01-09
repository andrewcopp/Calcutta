package httpserver

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/andrewcopp/Calcutta/backend/internal/app/suite_evaluations"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/dtos"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func (s *Server) listCohortSimulationsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cohortID := strings.TrimSpace(vars["cohortId"])
	if cohortID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "cohortId is required", "cohortId")
		return
	}
	if _, err := uuid.Parse(cohortID); err != nil {
		writeError(w, r, http.StatusBadRequest, "validation_error", "cohortId must be a valid UUID", "cohortId")
		return
	}

	q := r.URL.Query()
	q.Set("cohort_id", cohortID)
	r.URL.RawQuery = q.Encode()
	s.listSuiteCalcuttaEvaluationsHandler(w, r)
}

func (s *Server) getCohortSimulationHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cohortID := strings.TrimSpace(vars["cohortId"])
	id := strings.TrimSpace(vars["id"])
	if cohortID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "cohortId is required", "cohortId")
		return
	}
	if _, err := uuid.Parse(cohortID); err != nil {
		writeError(w, r, http.StatusBadRequest, "validation_error", "cohortId must be a valid UUID", "cohortId")
		return
	}
	if id == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "id is required", "id")
		return
	}

	it, err := s.loadSuiteCalcuttaEvaluationByID(r.Context(), id)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	if strings.TrimSpace(it.CohortID) != cohortID {
		writeError(w, r, http.StatusNotFound, "not_found", "Simulation not found", "id")
		return
	}
	writeJSON(w, http.StatusOK, it)
}

func (s *Server) getCohortSimulationResultHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cohortID := strings.TrimSpace(vars["cohortId"])
	id := strings.TrimSpace(vars["id"])
	if cohortID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "cohortId is required", "cohortId")
		return
	}
	if _, err := uuid.Parse(cohortID); err != nil {
		writeError(w, r, http.StatusBadRequest, "validation_error", "cohortId must be a valid UUID", "cohortId")
		return
	}
	if id == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "id is required", "id")
		return
	}

	it, err := s.loadSuiteCalcuttaEvaluationByID(r.Context(), id)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	if strings.TrimSpace(it.CohortID) != cohortID {
		writeError(w, r, http.StatusNotFound, "not_found", "Simulation not found", "id")
		return
	}

	s.getSuiteCalcuttaEvaluationResultHandler(w, r)
}

func (s *Server) getCohortSimulationSnapshotEntryHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cohortID := strings.TrimSpace(vars["cohortId"])
	id := strings.TrimSpace(vars["id"])
	if cohortID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "cohortId is required", "cohortId")
		return
	}
	if _, err := uuid.Parse(cohortID); err != nil {
		writeError(w, r, http.StatusBadRequest, "validation_error", "cohortId must be a valid UUID", "cohortId")
		return
	}
	if id == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "id is required", "id")
		return
	}

	it, err := s.loadSuiteCalcuttaEvaluationByID(r.Context(), id)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	if strings.TrimSpace(it.CohortID) != cohortID {
		writeError(w, r, http.StatusNotFound, "not_found", "Simulation not found", "id")
		return
	}

	s.getSuiteCalcuttaEvaluationSnapshotEntryHandler(w, r)
}

func (s *Server) createCohortSimulationHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cohortID := strings.TrimSpace(vars["cohortId"])
	if cohortID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "cohortId is required", "cohortId")
		return
	}
	if _, err := uuid.Parse(cohortID); err != nil {
		writeError(w, r, http.StatusBadRequest, "validation_error", "cohortId must be a valid UUID", "cohortId")
		return
	}

	var req dtos.CreateSimulationRunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}
	if req.CohortID == nil {
		v := cohortID
		req.CohortID = &v
	}

	b, _ := json.Marshal(req)
	cloned := r.Clone(r.Context())
	cloned.Body = io.NopCloser(bytes.NewReader(b))
	s.createSuiteCalcuttaEvaluationHandler(w, cloned)
}

func (s *Server) getSuiteCalcuttaEvaluationSnapshotEntryHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	evalID := vars["id"]
	snapshotEntryID := vars["snapshotEntryId"]
	if evalID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "id is required", "id")
		return
	}
	if snapshotEntryID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "snapshotEntryId is required", "snapshotEntryId")
		return
	}

	ctx := r.Context()
	entry, err := s.app.SuiteEvaluations.GetSnapshotEntry(ctx, evalID, snapshotEntryID)
	if err != nil {
		if errors.Is(err, suite_evaluations.ErrEvaluationHasNoCalcuttaEvaluationRunID) {
			writeError(w, r, http.StatusConflict, "invalid_state", "Evaluation has no calcutta_evaluation_run_id", "calcutta_evaluation_run_id")
			return
		}
		if errors.Is(err, suite_evaluations.ErrSnapshotEntryNotFoundForEvaluation) {
			writeError(w, r, http.StatusNotFound, "not_found", "Snapshot entry not found for this evaluation", "snapshotEntryId")
			return
		}
		writeErrorFromErr(w, r, err)
		return
	}

	teams := make([]suiteCalcuttaSnapshotEntryTeam, 0, len(entry.Teams))
	for _, t := range entry.Teams {
		teams = append(teams, suiteCalcuttaSnapshotEntryTeam{TeamID: t.TeamID, School: t.School, Seed: t.Seed, Region: t.Region, BidPoints: t.BidPoints})
	}

	writeJSON(w, http.StatusOK, suiteCalcuttaSnapshotEntryResponse{SnapshotEntryID: snapshotEntryID, DisplayName: entry.DisplayName, IsSynthetic: entry.IsSynthetic, Teams: teams})
}

func (s *Server) createSuiteCalcuttaEvaluationHandler(w http.ResponseWriter, r *http.Request) {
	var req dtos.CreateSimulationRunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}
	if err := req.Validate(); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	ctx := r.Context()
	if req.SimulationRunBatchID == nil {
		if req.CohortID == nil || strings.TrimSpace(*req.CohortID) == "" {
			writeError(w, r, http.StatusBadRequest, "validation_error", "cohortId is required", "cohortId")
			return
		}
	}
	cohortID := ""
	if req.CohortID != nil {
		cohortID = strings.TrimSpace(*req.CohortID)
	}
	optimizerKey := ""
	if req.OptimizerKey != nil {
		optimizerKey = strings.TrimSpace(*req.OptimizerKey)
	}
	var nSimsPtr *int
	if req.NSims > 0 {
		v := req.NSims
		nSimsPtr = &v
	}
	var seedPtr *int
	if req.Seed != 0 {
		v := req.Seed
		seedPtr = &v
	}

	created, err := s.app.SuiteEvaluations.CreateEvaluation(ctx, suite_evaluations.CreateSimulationParams{
		SimulationRunBatchID: req.SimulationRunBatchID,
		CohortID:             cohortID,
		CalcuttaID:           req.CalcuttaID,
		SimulatedCalcuttaID:  req.SimulatedCalcuttaID,
		GameOutcomeRunID:     req.GameOutcomeRunID,
		MarketShareRunID:     req.MarketShareRunID,
		OptimizerKey:         optimizerKey,
		NSims:                nSimsPtr,
		Seed:                 seedPtr,
		StartingStateKey:     req.StartingStateKey,
		ExcludedEntryName:    req.ExcludedEntryName,
	})
	if err != nil {
		if errors.Is(err, suite_evaluations.ErrMissingGameOutcomeRunForBatch) {
			writeError(w, r, http.StatusConflict, "missing_run", "Missing game-outcome run for simulation batch", "gameOutcomeRunId")
			return
		}
		if errors.Is(err, suite_evaluations.ErrMissingMarketShareRunForBatch) {
			writeError(w, r, http.StatusConflict, "missing_run", "Missing market-share run for simulation batch", "marketShareRunId")
			return
		}
		writeErrorFromErr(w, r, err)
		return
	}

	writeJSON(w, http.StatusCreated, createSuiteCalcuttaEvaluationResponse{ID: created.ID, Status: created.Status})
}

func (s *Server) listSuiteCalcuttaEvaluationsHandler(w http.ResponseWriter, r *http.Request) {
	calcuttaID := r.URL.Query().Get("calcutta_id")
	cohortID := r.URL.Query().Get("cohort_id")
	simulationBatchID := r.URL.Query().Get("simulation_batch_id")
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

	items, err := s.loadSuiteCalcuttaEvaluations(r.Context(), calcuttaID, cohortID, simulationBatchID, limit, offset)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, suiteCalcuttaEvaluationListResponse{Items: items})
}

func (s *Server) getSuiteCalcuttaEvaluationHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "id is required", "id")
		return
	}

	it, err := s.loadSuiteCalcuttaEvaluationByID(r.Context(), id)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, it)
}

func (s *Server) getSuiteCalcuttaEvaluationResultHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "id is required", "id")
		return
	}

	ctx := r.Context()

	eval, err := s.loadSuiteCalcuttaEvaluationByID(ctx, id)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	portfolio := make([]suiteCalcuttaEvaluationPortfolioBid, 0)
	var our *suiteCalcuttaEvaluationOurStrategyPerformance
	if eval.StrategyGenerationRunID != nil && strings.TrimSpace(*eval.StrategyGenerationRunID) != "" {
		portfolioRows, err := s.app.SuiteEvaluations.ListPortfolioBids(ctx, *eval.StrategyGenerationRunID)
		if err != nil {
			writeErrorFromErr(w, r, err)
			return
		}
		portfolio = make([]suiteCalcuttaEvaluationPortfolioBid, 0, len(portfolioRows))
		for _, b := range portfolioRows {
			portfolio = append(portfolio, suiteCalcuttaEvaluationPortfolioBid{TeamID: b.TeamID, SchoolName: b.SchoolName, Seed: b.Seed, Region: b.Region, BidPoints: b.BidPoints, ExpectedROI: b.ExpectedROI})
		}

		if eval.CalcuttaEvaluationRunID != nil && *eval.CalcuttaEvaluationRunID != "" {
			it, err := s.app.SuiteEvaluations.GetOurStrategyPerformance(ctx, *eval.CalcuttaEvaluationRunID, eval.ID)
			if err != nil {
				writeErrorFromErr(w, r, err)
				return
			}
			if it != nil {
				tmp := suiteCalcuttaEvaluationOurStrategyPerformance{
					Rank:                   it.Rank,
					EntryName:              it.EntryName,
					MeanNormalizedPayout:   it.MeanNormalizedPayout,
					MedianNormalizedPayout: it.MedianNormalizedPayout,
					PTop1:                  it.PTop1,
					PInMoney:               it.PInMoney,
					TotalSimulations:       it.TotalSimulations,
				}
				our = &tmp
			}
		}
	}

	entries := make([]suiteCalcuttaEvaluationEntryPerformance, 0)
	finishByName := map[string]*suiteCalcuttaEvalFinish{}
	if eval.CalcuttaID != "" && eval.StrategyGenerationRunID != nil && strings.TrimSpace(*eval.StrategyGenerationRunID) != "" {
		if m, ok, err := s.computeHypotheticalFinishByEntryNameForStrategyRun(ctx, eval.CalcuttaID, *eval.StrategyGenerationRunID); err != nil {
			writeErrorFromErr(w, r, err)
			return
		} else if ok {
			finishByName = m
		}
	}

	if eval.CalcuttaEvaluationRunID != nil && strings.TrimSpace(*eval.CalcuttaEvaluationRunID) != "" {
		rows, err := s.app.SuiteEvaluations.ListEntryPerformance(ctx, *eval.CalcuttaEvaluationRunID)
		if err != nil {
			writeErrorFromErr(w, r, err)
			return
		}
		for _, r0 := range rows {
			it := suiteCalcuttaEvaluationEntryPerformance{Rank: r0.Rank, EntryName: r0.EntryName, SnapshotEntryID: r0.SnapshotEntryID, MeanNormalizedPayout: r0.MeanNormalizedPayout, PTop1: r0.PTop1, PInMoney: r0.PInMoney}
			if f := finishByName[it.EntryName]; f != nil {
				it.FinishPosition = &f.FinishPosition
				it.IsTied = &f.IsTied
				it.InTheMoney = &f.InTheMoney
				it.PayoutCents = &f.PayoutCents
				it.TotalPoints = &f.TotalPoints
			}
			entries = append(entries, it)
		}
	}

	writeJSON(w, http.StatusOK, suiteCalcuttaEvaluationResultResponse{
		Evaluation:  *eval,
		Portfolio:   portfolio,
		OurStrategy: our,
		Entries:     entries,
	})
}
