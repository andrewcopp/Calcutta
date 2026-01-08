package httpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/dtos"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
)

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
		return nil, dtos.ErrFieldInvalid("", "request body too large")
	}
	return b, nil
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

func (s *Server) registerLabCandidatesRoutes(r *mux.Router) {
	r.HandleFunc(
		"/api/lab/candidates",
		s.requirePermission("analytics.suites.read", s.listLabCandidatesHandler),
	).Methods("GET", "OPTIONS")
	r.HandleFunc(
		"/api/lab/candidates",
		s.requirePermission("analytics.suites.write", s.createLabCandidatesHandler),
	).Methods("POST", "OPTIONS")

	r.HandleFunc(
		"/api/lab/candidates/{candidateId}",
		s.requirePermission("analytics.suites.read", s.getLabCandidateDetailHandler),
	).Methods("GET", "OPTIONS")
}

func (s *Server) listLabCandidatesHandler(w http.ResponseWriter, r *http.Request) {
	calcuttaID := strings.TrimSpace(r.URL.Query().Get("calcutta_id"))
	if calcuttaID != "" {
		if _, err := uuid.Parse(calcuttaID); err != nil {
			writeError(w, r, http.StatusBadRequest, "validation_error", "calcutta_id must be a valid UUID", "calcutta_id")
			return
		}
	}

	tournamentID := strings.TrimSpace(r.URL.Query().Get("tournament_id"))
	if tournamentID != "" {
		if _, err := uuid.Parse(tournamentID); err != nil {
			writeError(w, r, http.StatusBadRequest, "validation_error", "tournament_id must be a valid UUID", "tournament_id")
			return
		}
	}

	strategyGenerationRunID := strings.TrimSpace(r.URL.Query().Get("strategy_generation_run_id"))
	if strategyGenerationRunID != "" {
		if _, err := uuid.Parse(strategyGenerationRunID); err != nil {
			writeError(w, r, http.StatusBadRequest, "validation_error", "strategy_generation_run_id must be a valid UUID", "strategy_generation_run_id")
			return
		}
	}

	marketShareArtifactID := strings.TrimSpace(r.URL.Query().Get("market_share_artifact_id"))
	if marketShareArtifactID != "" {
		if _, err := uuid.Parse(marketShareArtifactID); err != nil {
			writeError(w, r, http.StatusBadRequest, "validation_error", "market_share_artifact_id must be a valid UUID", "market_share_artifact_id")
			return
		}
	}

	advancementRunID := strings.TrimSpace(r.URL.Query().Get("advancement_run_id"))
	if advancementRunID != "" {
		if _, err := uuid.Parse(advancementRunID); err != nil {
			writeError(w, r, http.StatusBadRequest, "validation_error", "advancement_run_id must be a valid UUID", "advancement_run_id")
			return
		}
	}

	optimizerKey := strings.TrimSpace(r.URL.Query().Get("optimizer_key"))
	startingStateKey := strings.TrimSpace(r.URL.Query().Get("starting_state_key"))
	excludedEntryName := strings.TrimSpace(r.URL.Query().Get("excluded_entry_name"))
	sourceKind := strings.TrimSpace(r.URL.Query().Get("source_kind"))

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

	rows, err := s.pool.Query(r.Context(), `
		SELECT
			id::text,
			display_name,
			source_kind,
			source_entry_artifact_id::text,
			calcutta_id::text,
			tournament_id::text,
			strategy_generation_run_id::text,
			market_share_run_id::text,
			market_share_artifact_id::text,
			advancement_run_id::text,
			optimizer_key,
			starting_state_key,
			excluded_entry_name,
			git_sha
		FROM derived.candidates
		WHERE deleted_at IS NULL
			AND ($1::uuid IS NULL OR calcutta_id = $1::uuid)
			AND ($2::uuid IS NULL OR tournament_id = $2::uuid)
			AND ($3::uuid IS NULL OR strategy_generation_run_id = $3::uuid)
			AND ($4::uuid IS NULL OR market_share_artifact_id = $4::uuid)
			AND ($5::uuid IS NULL OR advancement_run_id = $5::uuid)
			AND ($6::text IS NULL OR optimizer_key = $6::text)
			AND ($7::text IS NULL OR starting_state_key = $7::text)
			AND ($8::text IS NULL OR excluded_entry_name = $8::text)
			AND ($9::text IS NULL OR source_kind = $9::text)
		ORDER BY created_at DESC
		LIMIT $10::int
		OFFSET $11::int
	`,
		nullUUIDParam(calcuttaID),
		nullUUIDParam(tournamentID),
		nullUUIDParam(strategyGenerationRunID),
		nullUUIDParam(marketShareArtifactID),
		nullUUIDParam(advancementRunID),
		nullUUIDParam(optimizerKey),
		nullUUIDParam(startingStateKey),
		nullUUIDParam(excludedEntryName),
		nullUUIDParam(sourceKind),
		limit,
		offset,
	)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	defer rows.Close()

	items := make([]labCandidateDetailResponse, 0)
	for rows.Next() {
		var it labCandidateDetailResponse
		if err := rows.Scan(
			&it.CandidateID,
			&it.DisplayName,
			&it.SourceKind,
			&it.SourceEntryArtifactID,
			&it.CalcuttaID,
			&it.TournamentID,
			&it.StrategyGenerationRunID,
			&it.MarketShareRunID,
			&it.MarketShareArtifactID,
			&it.AdvancementRunID,
			&it.OptimizerKey,
			&it.StartingStateKey,
			&it.ExcludedEntryName,
			&it.GitSHA,
		); err != nil {
			writeErrorFromErr(w, r, err)
			return
		}
		it.Teams = nil
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, listLabCandidatesResponse{Items: items})
}

func (s *Server) getLabCandidateDetailHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	candidateID := strings.TrimSpace(vars["candidateId"])
	if candidateID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "candidateId is required", "candidateId")
		return
	}
	if _, err := uuid.Parse(candidateID); err != nil {
		writeError(w, r, http.StatusBadRequest, "validation_error", "candidateId must be a valid UUID", "candidateId")
		return
	}

	ctx := r.Context()

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{AccessMode: pgx.ReadOnly})
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var resp labCandidateDetailResponse
	if err := tx.QueryRow(ctx, `
		SELECT
			c.id::text,
			c.display_name,
			c.source_kind,
			c.source_entry_artifact_id::text,
			c.calcutta_id::text,
			c.tournament_id::text,
			c.strategy_generation_run_id::text,
			c.market_share_run_id::text,
			c.market_share_artifact_id::text,
			c.advancement_run_id::text,
			c.optimizer_key,
			c.starting_state_key,
			c.excluded_entry_name,
			c.git_sha
		FROM derived.candidates c
		WHERE c.id = $1::uuid
			AND c.deleted_at IS NULL
		LIMIT 1
	`, candidateID).Scan(
		&resp.CandidateID,
		&resp.DisplayName,
		&resp.SourceKind,
		&resp.SourceEntryArtifactID,
		&resp.CalcuttaID,
		&resp.TournamentID,
		&resp.StrategyGenerationRunID,
		&resp.MarketShareRunID,
		&resp.MarketShareArtifactID,
		&resp.AdvancementRunID,
		&resp.OptimizerKey,
		&resp.StartingStateKey,
		&resp.ExcludedEntryName,
		&resp.GitSHA,
	); err != nil {
		if err == pgx.ErrNoRows {
			writeError(w, r, http.StatusNotFound, "not_found", "Candidate not found", "candidateId")
			return
		}
		writeErrorFromErr(w, r, err)
		return
	}

	resp.Teams = make([]labCandidateDetailTeam, 0)
	if strings.TrimSpace(resp.StrategyGenerationRunID) != "" {
		bidRows, err := tx.Query(ctx, `
			SELECT team_id::text, bid_points::int
			FROM derived.recommended_entry_bids
			WHERE strategy_generation_run_id = $1::uuid
				AND deleted_at IS NULL
			ORDER BY bid_points DESC
		`, resp.StrategyGenerationRunID)
		if err != nil {
			writeErrorFromErr(w, r, err)
			return
		}
		defer bidRows.Close()

		for bidRows.Next() {
			var t labCandidateDetailTeam
			if err := bidRows.Scan(&t.TeamID, &t.BidPoints); err != nil {
				writeErrorFromErr(w, r, err)
				return
			}
			resp.Teams = append(resp.Teams, t)
		}
		if err := bidRows.Err(); err != nil {
			writeErrorFromErr(w, r, err)
			return
		}
	}

	if err := tx.Commit(ctx); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) createLabCandidatesHandler(w http.ResponseWriter, r *http.Request) {
	// Accept either a single createLabCandidateRequest or a bulk wrapper {items:[...]}.
	body, err := readBodyLimit(r, 2<<20)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	bulk := createLabCandidatesBulkRequest{}
	if err := json.Unmarshal(body, &bulk); err == nil && len(bulk.Items) > 0 {
		resp, err := s.createLabCandidatesBulk(r.Context(), bulk.Items)
		if err != nil {
			writeErrorFromErr(w, r, err)
			return
		}
		writeJSON(w, http.StatusCreated, resp)
		return
	}

	single := createLabCandidateRequest{}
	if err := json.Unmarshal(body, &single); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}

	resp, err := s.createLabCandidatesBulk(r.Context(), []createLabCandidateRequest{single})
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	if len(resp.Items) != 1 {
		writeError(w, r, http.StatusInternalServerError, "internal_error", "Unexpected create result", "")
		return
	}
	writeJSON(w, http.StatusCreated, resp.Items[0])
}

func (s *Server) createLabCandidatesBulk(ctx context.Context, items []createLabCandidateRequest) (*createLabCandidatesBulkResponse, error) {
	if len(items) == 0 {
		return nil, dtos.ErrFieldRequired("items")
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	committed := false
	defer func() {
		if committed {
			return
		}
		_ = tx.Rollback(ctx)
	}()

	out := make([]createLabCandidateResponse, 0, len(items))
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

		if req.CalcuttaID == "" {
			return nil, dtos.ErrFieldRequired("calcuttaId")
		}
		if _, err := uuid.Parse(req.CalcuttaID); err != nil {
			return nil, dtos.ErrFieldInvalid("calcuttaId", "must be a valid UUID")
		}
		if req.AdvancementRunID == "" {
			return nil, dtos.ErrFieldRequired("advancementRunId")
		}
		if _, err := uuid.Parse(req.AdvancementRunID); err != nil {
			return nil, dtos.ErrFieldInvalid("advancementRunId", "must be a valid UUID")
		}
		if req.MarketShareArtifactID == "" {
			return nil, dtos.ErrFieldRequired("marketShareArtifactId")
		}
		if _, err := uuid.Parse(req.MarketShareArtifactID); err != nil {
			return nil, dtos.ErrFieldInvalid("marketShareArtifactId", "must be a valid UUID")
		}
		if req.OptimizerKey == "" {
			return nil, dtos.ErrFieldRequired("optimizerKey")
		}
		if req.StartingStateKey == "" {
			return nil, dtos.ErrFieldRequired("startingStateKey")
		}

		// Resolve tournament_id from calcutta.
		tournamentID := ""
		if err := tx.QueryRow(ctx, `
			SELECT tournament_id::text
			FROM core.calcuttas
			WHERE id = $1::uuid
				AND deleted_at IS NULL
			LIMIT 1
		`, req.CalcuttaID).Scan(&tournamentID); err != nil {
			return nil, err
		}
		tournamentID = strings.TrimSpace(tournamentID)
		if tournamentID == "" {
			return nil, dtos.ErrFieldInvalid("calcuttaId", "not found")
		}

		// Resolve market_share_run_id from the artifact and validate it belongs to this calcutta.
		marketShareRunID := ""
		if err := tx.QueryRow(ctx, `
			SELECT r.id::text
			FROM derived.run_artifacts a
			JOIN derived.market_share_runs r
				ON r.id = a.run_id
				AND r.deleted_at IS NULL
			WHERE a.id = $1::uuid
				AND a.run_kind = 'market_share'
				AND a.artifact_kind = 'metrics'
				AND a.deleted_at IS NULL
				AND r.calcutta_id = $2::uuid
			LIMIT 1
		`, req.MarketShareArtifactID, req.CalcuttaID).Scan(&marketShareRunID); err != nil {
			if err == pgx.ErrNoRows {
				return nil, dtos.ErrFieldInvalid("marketShareArtifactId", "not found for calcutta")
			}
			return nil, err
		}
		marketShareRunID = strings.TrimSpace(marketShareRunID)
		if marketShareRunID == "" {
			return nil, dtos.ErrFieldInvalid("marketShareArtifactId", "not found for calcutta")
		}

		// Validate advancement run belongs to this tournament.
		var verifiedAdvancementRunID string
		if err := tx.QueryRow(ctx, `
			SELECT id::text
			FROM derived.game_outcome_runs
			WHERE id = $1::uuid
				AND tournament_id = $2::uuid
				AND deleted_at IS NULL
			LIMIT 1
		`, req.AdvancementRunID, tournamentID).Scan(&verifiedAdvancementRunID); err != nil {
			if err == pgx.ErrNoRows {
				return nil, dtos.ErrFieldInvalid("advancementRunId", "not found for tournament")
			}
			return nil, err
		}

		displayName := "Lab Candidate"
		if req.DisplayName != nil {
			displayName = *req.DisplayName
		} else {
			displayName = fmt.Sprintf("lab_candidate_%s", req.OptimizerKey)
		}

		// Create candidate identity first.
		candidateID := ""
		if err := tx.QueryRow(ctx, `
			INSERT INTO derived.candidates (
				source_kind,
				source_entry_artifact_id,
				display_name,
				metadata_json,
				calcutta_id,
				tournament_id,
				strategy_generation_run_id,
				market_share_run_id,
				market_share_artifact_id,
				advancement_run_id,
				optimizer_key,
				starting_state_key,
				excluded_entry_name,
				git_sha
			)
			VALUES (
				'entry_artifact',
				NULL,
				$1,
				'{}'::jsonb,
				$2::uuid,
				$3::uuid,
				NULL,
				$4::uuid,
				$5::uuid,
				$6::uuid,
				$7,
				$8,
				$9,
				$10
			)
			RETURNING id::text
		`, displayName, req.CalcuttaID, tournamentID, marketShareRunID, req.MarketShareArtifactID, req.AdvancementRunID, req.OptimizerKey, req.StartingStateKey, req.ExcludedEntryName, nil).Scan(&candidateID); err != nil {
			return nil, err
		}

		runKeyUUID := uuid.NewString()
		runKeyText := runKeyUUID
		name := fmt.Sprintf("lab_candidate_%s", req.OptimizerKey)
		params := map[string]any{
			"candidate_id":             candidateID,
			"market_share_artifact_id": req.MarketShareArtifactID,
			"advancement_run_id":       req.AdvancementRunID,
			"source":                   "lab_candidates_create",
		}
		paramsJSON, _ := json.Marshal(params)

		var runID string
		if err := tx.QueryRow(ctx, `
			INSERT INTO derived.strategy_generation_runs (
				run_key,
				run_key_uuid,
				name,
				simulated_tournament_id,
				calcutta_id,
				purpose,
				returns_model_key,
				investment_model_key,
				optimizer_key,
				market_share_run_id,
				game_outcome_run_id,
				excluded_entry_name,
				starting_state_key,
				params_json,
				git_sha,
				created_at,
				updated_at
			)
			VALUES ($1, $2::uuid, $3, NULL, $4::uuid, 'lab_candidates_generation', 'pgo_dp', 'predicted_market_share', $5, $6::uuid, $7::uuid, $8::text, $9::text, $10::jsonb, NULL, NOW(), NOW())
			RETURNING id::text
		`, runKeyText, runKeyUUID, name, req.CalcuttaID, req.OptimizerKey, marketShareRunID, req.AdvancementRunID, req.ExcludedEntryName, req.StartingStateKey, string(paramsJSON)).Scan(&runID); err != nil {
			return nil, err
		}

		// Link candidate -> strategy generation run.
		if _, err := tx.Exec(ctx, `
			UPDATE derived.candidates
			SET strategy_generation_run_id = $2::uuid,
				updated_at = NOW()
			WHERE id = $1::uuid
				AND deleted_at IS NULL
		`, candidateID, runID); err != nil {
			return nil, err
		}

		out = append(out, createLabCandidateResponse{CandidateID: candidateID, StrategyGenerationRunID: runID})
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	committed = true

	return &createLabCandidatesBulkResponse{Items: out}, nil
}
