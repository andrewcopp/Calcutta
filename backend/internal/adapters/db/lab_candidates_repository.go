package db

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/andrewcopp/Calcutta/backend/internal/app/apperrors"
	applabcandidates "github.com/andrewcopp/Calcutta/backend/internal/app/lab_candidates"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LabCandidatesRepository struct {
	pool *pgxpool.Pool
}

func NewLabCandidatesRepository(pool *pgxpool.Pool) *LabCandidatesRepository {
	return &LabCandidatesRepository{pool: pool}
}

func nullOptionalString(s *string) any {
	if s == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*s)
	if trimmed == "" {
		return nil
	}
	return trimmed
}

func nullOptionalUUIDString(s *string) any {
	v := nullOptionalString(s)
	if v == nil {
		return nil
	}
	return v
}

func (r *LabCandidatesRepository) GetCandidateDetail(ctx context.Context, candidateID string) (*applabcandidates.CandidateDetail, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{AccessMode: pgx.ReadOnly})
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

	var resp applabcandidates.CandidateDetail
	if err := tx.QueryRow(ctx, `
		SELECT
			c.id::text,
			c.display_name,
			c.source_kind,
			c.source_entry_artifact_id::text,
			c.calcutta_id::text,
			COALESCE(cc.name, ''::text) AS calcutta_name,
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
		LEFT JOIN core.calcuttas cc ON cc.id = c.calcutta_id AND cc.deleted_at IS NULL
		WHERE c.id = $1::uuid
			AND c.deleted_at IS NULL
		LIMIT 1
	`, candidateID).Scan(
		&resp.CandidateID,
		&resp.DisplayName,
		&resp.SourceKind,
		&resp.SourceEntryArtifactID,
		&resp.CalcuttaID,
		&resp.CalcuttaName,
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
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &apperrors.NotFoundError{Resource: "candidate", ID: candidateID}
		}
		return nil, err
	}

	resp.Teams = make([]applabcandidates.CandidateDetailTeam, 0)
	rows, err := tx.Query(ctx, `
		SELECT team_id::text, bid_points::int
		FROM derived.candidate_bids
		WHERE candidate_id = $1::uuid
			AND deleted_at IS NULL
		ORDER BY bid_points DESC
	`, candidateID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var t applabcandidates.CandidateDetailTeam
		if err := rows.Scan(&t.TeamID, &t.BidPoints); err != nil {
			return nil, err
		}
		resp.Teams = append(resp.Teams, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	committed = true

	return &resp, nil
}

func (r *LabCandidatesRepository) DeleteCandidate(ctx context.Context, candidateID string) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	committed := false
	defer func() {
		if committed {
			return
		}
		_ = tx.Rollback(ctx)
	}()

	// Soft-delete bids first.
	_, err = tx.Exec(ctx, `
		UPDATE derived.candidate_bids
		SET deleted_at = NOW(),
			updated_at = NOW()
		WHERE candidate_id = $1::uuid
			AND deleted_at IS NULL
	`, candidateID)
	if err != nil {
		return err
	}

	// Soft-delete candidate.
	ct, err := tx.Exec(ctx, `
		UPDATE derived.candidates
		SET deleted_at = NOW(),
			updated_at = NOW()
		WHERE id = $1::uuid
			AND deleted_at IS NULL
	`, candidateID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return &apperrors.NotFoundError{Resource: "candidate", ID: candidateID}
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	committed = true
	return nil
}

func (r *LabCandidatesRepository) ListCandidates(ctx context.Context, filter applabcandidates.ListCandidatesFilter, page applabcandidates.ListCandidatesPagination) ([]applabcandidates.CandidateDetail, error) {
	rows, err := r.pool.Query(ctx, `
		WITH seed_list AS (
			SELECT
				cb.candidate_id,
				array_agg(t.seed::text ORDER BY t.seed ASC) FILTER (WHERE cb.bid_points > 0) AS seeds
			FROM derived.candidate_bids cb
			JOIN core.teams t ON t.id = cb.team_id AND t.deleted_at IS NULL
			WHERE cb.deleted_at IS NULL
			GROUP BY cb.candidate_id
		)
		SELECT
			c.id::text,
			c.display_name,
			c.source_kind,
			c.source_entry_artifact_id::text,
			c.calcutta_id::text,
			COALESCE(cc.name, ''::text) AS calcutta_name,
			c.tournament_id::text,
			c.strategy_generation_run_id::text,
			c.market_share_run_id::text,
			c.market_share_artifact_id::text,
			c.advancement_run_id::text,
			c.optimizer_key,
			c.starting_state_key,
			array_to_string(
				COALESCE(sl.seeds, ARRAY[]::text[])
				|| COALESCE(
					array_fill(
						'-'::text,
						ARRAY[GREATEST(cc.max_teams - COALESCE(array_length(sl.seeds, 1), 0), 0)]
					),
					ARRAY[]::text[]
				),
				','
			) AS seed_preview,
			c.excluded_entry_name,
			c.git_sha
		FROM derived.candidates c
		LEFT JOIN core.calcuttas cc ON cc.id = c.calcutta_id AND cc.deleted_at IS NULL
		LEFT JOIN seed_list sl ON sl.candidate_id = c.id
		LEFT JOIN derived.game_outcome_runs gor ON gor.id = c.advancement_run_id AND gor.deleted_at IS NULL
		LEFT JOIN derived.market_share_runs msr ON msr.id = c.market_share_run_id AND msr.deleted_at IS NULL
		WHERE c.deleted_at IS NULL
			AND ($1::uuid IS NULL OR c.calcutta_id = $1::uuid)
			AND ($2::uuid IS NULL OR c.tournament_id = $2::uuid)
			AND ($3::uuid IS NULL OR c.strategy_generation_run_id = $3::uuid)
			AND ($4::uuid IS NULL OR c.market_share_artifact_id = $4::uuid)
			AND ($5::uuid IS NULL OR c.advancement_run_id = $5::uuid)
			AND ($6::text IS NULL OR c.optimizer_key = $6::text)
			AND ($7::text IS NULL OR c.starting_state_key = $7::text)
			AND ($8::text IS NULL OR c.excluded_entry_name = $8::text)
			AND ($9::text IS NULL OR c.source_kind = $9::text)
			AND ($10::uuid IS NULL OR gor.algorithm_id = $10::uuid)
			AND ($11::uuid IS NULL OR msr.algorithm_id = $11::uuid)
		ORDER BY c.created_at DESC
		LIMIT $12::int
		OFFSET $13::int
	`,
		nullOptionalUUIDString(filter.CalcuttaID),
		nullOptionalUUIDString(filter.TournamentID),
		nullOptionalUUIDString(filter.StrategyGenerationRunID),
		nullOptionalUUIDString(filter.MarketShareArtifactID),
		nullOptionalUUIDString(filter.AdvancementRunID),
		nullOptionalString(filter.OptimizerKey),
		nullOptionalString(filter.StartingStateKey),
		nullOptionalString(filter.ExcludedEntryName),
		nullOptionalString(filter.SourceKind),
		nullOptionalUUIDString(filter.GameOutcomesAlgorithmID),
		nullOptionalUUIDString(filter.MarketShareAlgorithmID),
		page.Limit,
		page.Offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]applabcandidates.CandidateDetail, 0)
	for rows.Next() {
		var it applabcandidates.CandidateDetail
		if err := rows.Scan(
			&it.CandidateID,
			&it.DisplayName,
			&it.SourceKind,
			&it.SourceEntryArtifactID,
			&it.CalcuttaID,
			&it.CalcuttaName,
			&it.TournamentID,
			&it.StrategyGenerationRunID,
			&it.MarketShareRunID,
			&it.MarketShareArtifactID,
			&it.AdvancementRunID,
			&it.OptimizerKey,
			&it.StartingStateKey,
			&it.SeedPreview,
			&it.ExcludedEntryName,
			&it.GitSHA,
		); err != nil {
			return nil, err
		}
		it.Teams = nil
		out = append(out, it)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

func (r *LabCandidatesRepository) ListCandidateComboCoverage(ctx context.Context, startingStateKey string, excludedEntryName *string) (*applabcandidates.CandidateComboCoverageSummary, error) {
	excluded := ""
	if excludedEntryName != nil {
		excluded = strings.TrimSpace(*excludedEntryName)
	}

	var totalCalcuttas int
	if err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*)::int
		FROM core.calcuttas
		WHERE deleted_at IS NULL
	`).Scan(&totalCalcuttas); err != nil {
		return nil, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT
			gor.algorithm_id::text AS game_outcomes_algorithm_id,
			msr.algorithm_id::text AS market_share_algorithm_id,
			c.optimizer_key,
			COUNT(DISTINCT c.calcutta_id)::int AS existing_candidates
		FROM derived.candidates c
		JOIN derived.game_outcome_runs gor
			ON gor.id = c.advancement_run_id
			AND gor.deleted_at IS NULL
		JOIN derived.market_share_runs msr
			ON msr.id = c.market_share_run_id
			AND msr.deleted_at IS NULL
		WHERE c.deleted_at IS NULL
			AND c.starting_state_key = $1::text
			AND ($2::text = '' OR COALESCE(c.excluded_entry_name, ''::text) = $2::text)
		GROUP BY 1, 2, 3
	`, startingStateKey, excluded)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]applabcandidates.CandidateComboCoverage, 0)
	for rows.Next() {
		var it applabcandidates.CandidateComboCoverage
		if err := rows.Scan(&it.GameOutcomesAlgorithmID, &it.MarketShareAlgorithmID, &it.OptimizerKey, &it.ExistingCandidates); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &applabcandidates.CandidateComboCoverageSummary{TotalCalcuttas: totalCalcuttas, Items: out}, nil
}

func (r *LabCandidatesRepository) GenerateCandidatesFromAlgorithms(ctx context.Context, req applabcandidates.GenerateCandidatesFromAlgorithmsRequest) (*applabcandidates.GenerateCandidatesFromAlgorithmsResult, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
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

	excluded := ""
	if req.ExcludedEntryName != nil {
		excluded = strings.TrimSpace(*req.ExcludedEntryName)
	}

	var totalCalcuttas int
	if err := tx.QueryRow(ctx, `
		SELECT COUNT(*)::int
		FROM core.calcuttas
		WHERE deleted_at IS NULL
	`).Scan(&totalCalcuttas); err != nil {
		return nil, err
	}

	type eligibleRow struct {
		calcuttaID            string
		tournamentID          string
		advancementRunID      string
		marketShareRunID      string
		marketShareArtifactID string
		marketShareExcluded   string
	}

	rows, err := tx.Query(ctx, `
		SELECT
			c.id::text AS calcutta_id,
			c.tournament_id::text AS tournament_id,
			COALESCE((
				SELECT gor.id::text
				FROM derived.game_outcome_runs gor
				WHERE gor.tournament_id = c.tournament_id
					AND gor.algorithm_id = $1::uuid
					AND gor.deleted_at IS NULL
				ORDER BY gor.created_at DESC
				LIMIT 1
			), ''::text) AS advancement_run_id,
			COALESCE((
				SELECT msr.id::text
				FROM derived.market_share_runs msr
				WHERE msr.calcutta_id = c.id
					AND msr.algorithm_id = $2::uuid
					AND msr.deleted_at IS NULL
					AND ($3::text = '' OR COALESCE(msr.params_json->>'excluded_entry_name','') = $3::text)
				ORDER BY msr.created_at DESC
				LIMIT 1
			), ''::text) AS market_share_run_id,
			COALESCE((
				SELECT COALESCE(msr3.params_json->>'excluded_entry_name', ''::text)
				FROM derived.market_share_runs msr3
				WHERE msr3.calcutta_id = c.id
					AND msr3.algorithm_id = $2::uuid
					AND msr3.deleted_at IS NULL
					AND ($3::text = '' OR COALESCE(msr3.params_json->>'excluded_entry_name','') = $3::text)
				ORDER BY msr3.created_at DESC
				LIMIT 1
			), ''::text) AS market_share_excluded,
			COALESCE((
				SELECT a.id::text
				FROM derived.run_artifacts a
				WHERE a.run_kind = 'market_share'
					AND a.artifact_kind = 'metrics'
					AND a.deleted_at IS NULL
					AND a.run_id = (
						SELECT msr2.id
						FROM derived.market_share_runs msr2
						WHERE msr2.calcutta_id = c.id
							AND msr2.algorithm_id = $2::uuid
							AND msr2.deleted_at IS NULL
							AND ($3::text = '' OR COALESCE(msr2.params_json->>'excluded_entry_name','') = $3::text)
						ORDER BY msr2.created_at DESC
						LIMIT 1
					)
				ORDER BY a.created_at DESC
				LIMIT 1
			), ''::text) AS market_share_artifact_id
		FROM core.calcuttas c
		WHERE c.deleted_at IS NULL
		ORDER BY c.created_at ASC
	`, req.GameOutcomesAlgorithmID, req.MarketShareAlgorithmID, excluded)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	eligible := make([]eligibleRow, 0)
	for rows.Next() {
		var rrow eligibleRow
		if err := rows.Scan(&rrow.calcuttaID, &rrow.tournamentID, &rrow.advancementRunID, &rrow.marketShareRunID, &rrow.marketShareExcluded, &rrow.marketShareArtifactID); err != nil {
			return nil, err
		}
		if strings.TrimSpace(rrow.advancementRunID) == "" || strings.TrimSpace(rrow.marketShareRunID) == "" || strings.TrimSpace(rrow.marketShareArtifactID) == "" {
			continue
		}
		eligible = append(eligible, rrow)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	created := 0
	skippedExisting := 0
	for _, e := range eligible {
		candidateExcluded := excluded
		if candidateExcluded == "" {
			candidateExcluded = strings.TrimSpace(e.marketShareExcluded)
		}
		var candidateExcludedPtr *string
		if strings.TrimSpace(candidateExcluded) != "" {
			v := candidateExcluded
			candidateExcludedPtr = &v
		}

		var existingID string
		err := tx.QueryRow(ctx, `
			SELECT id::text
			FROM derived.candidates
			WHERE deleted_at IS NULL
				AND calcutta_id = $1::uuid
				AND optimizer_key = $2::text
				AND market_share_artifact_id = $3::uuid
				AND advancement_run_id = $4::uuid
				AND starting_state_key = $5::text
				AND COALESCE(excluded_entry_name, ''::text) = $6::text
			LIMIT 1
		`, e.calcuttaID, req.OptimizerKey, e.marketShareArtifactID, e.advancementRunID, req.StartingStateKey, candidateExcluded).Scan(&existingID)
		if err == nil {
			skippedExisting++
			continue
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}

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
		`, req.DisplayName, e.calcuttaID, e.tournamentID, e.marketShareRunID, e.marketShareArtifactID, e.advancementRunID, req.OptimizerKey, req.StartingStateKey, candidateExcludedPtr, nil).Scan(&candidateID); err != nil {
			return nil, err
		}

		runKeyUUID := uuid.NewString()
		runKeyText := runKeyUUID
		name := fmt.Sprintf("lab_candidate_%s", req.OptimizerKey)
		paramsJSON := fmt.Sprintf(`{"candidate_id":"%s","market_share_artifact_id":"%s","advancement_run_id":"%s","source":"lab_candidates_generate"}`,
			candidateID, e.marketShareArtifactID, e.advancementRunID,
		)

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
		`, runKeyText, runKeyUUID, name, e.calcuttaID, req.OptimizerKey, e.marketShareRunID, e.advancementRunID, candidateExcludedPtr, req.StartingStateKey, paramsJSON).Scan(&runID); err != nil {
			return nil, err
		}

		// Ensure the strategy generation worker sees this run by enqueuing it.
		// (If DB triggers already enqueue run_jobs, this is a no-op due to the unique index.)
		if _, err := tx.Exec(ctx, `
			INSERT INTO derived.run_jobs (
				run_kind,
				run_id,
				run_key,
				status,
				params_json,
				created_at,
				updated_at
			)
			VALUES (
				'strategy_generation',
				$1::uuid,
				$2::uuid,
				'queued',
				$3::jsonb,
				NOW(),
				NOW()
			)
			ON CONFLICT (run_kind, run_id)
			DO NOTHING
		`, runID, runKeyUUID, paramsJSON); err != nil {
			return nil, err
		}

		if _, err := tx.Exec(ctx, `
			UPDATE derived.candidates
			SET strategy_generation_run_id = $2::uuid,
				updated_at = NOW()
			WHERE id = $1::uuid
				AND deleted_at IS NULL
		`, candidateID, runID); err != nil {
			return nil, err
		}

		created++
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	committed = true

	eligibleCalcuttas := len(eligible)
	skippedMissingUpstream := totalCalcuttas - eligibleCalcuttas
	if skippedMissingUpstream < 0 {
		skippedMissingUpstream = 0
	}

	return &applabcandidates.GenerateCandidatesFromAlgorithmsResult{
		TotalCalcuttas:         totalCalcuttas,
		EligibleCalcuttas:      eligibleCalcuttas,
		CreatedCandidates:      created,
		SkippedExisting:        skippedExisting,
		SkippedMissingUpstream: skippedMissingUpstream,
	}, nil
}

func (r *LabCandidatesRepository) CreateCandidatesBulk(ctx context.Context, items []applabcandidates.CreateCandidateRequest) ([]applabcandidates.CreateCandidateResult, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
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

	out := make([]applabcandidates.CreateCandidateResult, 0, len(items))
	for i := range items {
		req := items[i]

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
			return nil, apperrors.FieldInvalid("calcuttaId", "not found")
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
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, apperrors.FieldInvalid("marketShareArtifactId", "not found for calcutta")
			}
			return nil, err
		}
		marketShareRunID = strings.TrimSpace(marketShareRunID)
		if marketShareRunID == "" {
			return nil, apperrors.FieldInvalid("marketShareArtifactId", "not found for calcutta")
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
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, apperrors.FieldInvalid("advancementRunId", "not found for tournament")
			}
			return nil, err
		}
		_ = verifiedAdvancementRunID

		displayName := "Lab Candidate"
		if req.DisplayName != nil {
			displayName = strings.TrimSpace(*req.DisplayName)
			if displayName == "" {
				displayName = "Lab Candidate"
			}
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
		paramsJSON := fmt.Sprintf(`{"candidate_id":"%s","market_share_artifact_id":"%s","advancement_run_id":"%s","source":"lab_candidates_create"}`,
			candidateID, req.MarketShareArtifactID, req.AdvancementRunID,
		)

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
		`, runKeyText, runKeyUUID, name, req.CalcuttaID, req.OptimizerKey, marketShareRunID, req.AdvancementRunID, req.ExcludedEntryName, req.StartingStateKey, paramsJSON).Scan(&runID); err != nil {
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

		out = append(out, applabcandidates.CreateCandidateResult{CandidateID: candidateID, StrategyGenerationRunID: runID})
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	committed = true

	return out, nil
}
