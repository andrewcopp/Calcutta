package db

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/app/synthetic_scenarios"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SyntheticScenariosRepository struct {
	pool *pgxpool.Pool
}

func NewSyntheticScenariosRepository(pool *pgxpool.Pool) *SyntheticScenariosRepository {
	return &SyntheticScenariosRepository{pool: pool}
}

func (r *SyntheticScenariosRepository) ListSyntheticCalcuttas(ctx context.Context, cohortID, calcuttaID *string, limit, offset int) ([]synthetic_scenarios.SyntheticCalcuttaListItem, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT
			sc.id::text,
			sc.cohort_id::text,
			sc.calcutta_id::text,
			sc.calcutta_snapshot_id::text,
			sc.highlighted_snapshot_entry_id::text,
			sc.focus_strategy_generation_run_id::text,
			sc.focus_entry_name,
			ls.status,
			ls.our_rank,
			ls.our_mean_normalized_payout,
			ls.our_p_top1,
			ls.our_p_in_money,
			ls.total_simulations,
			sc.starting_state_key,
			sc.excluded_entry_name,
			sc.notes,
			sc.metadata_json,
			sc.created_at,
			sc.updated_at
		FROM derived.synthetic_calcuttas sc
		LEFT JOIN LATERAL (
			SELECT
				sr.status,
				sr.our_rank,
				sr.our_mean_normalized_payout,
				sr.our_p_top1,
				sr.our_p_in_money,
				sr.total_simulations
			FROM derived.simulation_runs sr
			WHERE sr.synthetic_calcutta_id = sc.id
				AND sr.deleted_at IS NULL
			ORDER BY sr.created_at DESC
			LIMIT 1
		) ls ON TRUE
		WHERE sc.deleted_at IS NULL
			AND ($1::uuid IS NULL OR sc.cohort_id = $1::uuid)
			AND ($2::uuid IS NULL OR sc.calcutta_id = $2::uuid)
		ORDER BY sc.created_at DESC
		LIMIT $3::int
		OFFSET $4::int
	`, uuidParamOrNil(cohortID), uuidParamOrNil(calcuttaID), limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]synthetic_scenarios.SyntheticCalcuttaListItem, 0)
	for rows.Next() {
		var it synthetic_scenarios.SyntheticCalcuttaListItem
		if err := rows.Scan(
			&it.ID,
			&it.CohortID,
			&it.CalcuttaID,
			&it.CalcuttaSnapshotID,
			&it.HighlightedEntryID,
			&it.FocusStrategyGenerationID,
			&it.FocusEntryName,
			&it.LatestSimulationStatus,
			&it.OurRank,
			&it.OurMeanNormalizedPayout,
			&it.OurPTop1,
			&it.OurPInMoney,
			&it.TotalSimulations,
			&it.StartingStateKey,
			&it.ExcludedEntryName,
			&it.Notes,
			&it.Metadata,
			&it.CreatedAt,
			&it.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *SyntheticScenariosRepository) GetSyntheticCalcutta(ctx context.Context, id string) (*synthetic_scenarios.SyntheticCalcuttaListItem, error) {
	var it synthetic_scenarios.SyntheticCalcuttaListItem
	if err := r.pool.QueryRow(ctx, `
		SELECT
			sc.id::text,
			sc.cohort_id::text,
			sc.calcutta_id::text,
			sc.calcutta_snapshot_id::text,
			sc.highlighted_snapshot_entry_id::text,
			sc.focus_strategy_generation_run_id::text,
			sc.focus_entry_name,
			sc.starting_state_key,
			sc.excluded_entry_name,
			sc.notes,
			sc.metadata_json,
			sc.created_at,
			sc.updated_at
		FROM derived.synthetic_calcuttas sc
		WHERE sc.id = $1::uuid
			AND sc.deleted_at IS NULL
		LIMIT 1
	`, id).Scan(
		&it.ID,
		&it.CohortID,
		&it.CalcuttaID,
		&it.CalcuttaSnapshotID,
		&it.HighlightedEntryID,
		&it.FocusStrategyGenerationID,
		&it.FocusEntryName,
		&it.StartingStateKey,
		&it.ExcludedEntryName,
		&it.Notes,
		&it.Metadata,
		&it.CreatedAt,
		&it.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, synthetic_scenarios.ErrSyntheticCalcuttaNotFound
		}
		return nil, err
	}
	return &it, nil
}

func (r *SyntheticScenariosRepository) PatchSyntheticCalcutta(ctx context.Context, id string, p synthetic_scenarios.PatchSyntheticCalcuttaParams) error {
	var snapshotID *string
	var existingHighlighted *string
	var existingNotes *string
	var existingMetadata json.RawMessage
	if err := r.pool.QueryRow(ctx, `
		SELECT
			calcutta_snapshot_id::text,
			highlighted_snapshot_entry_id::text,
			notes,
			metadata_json
		FROM derived.synthetic_calcuttas
		WHERE id = $1::uuid
			AND deleted_at IS NULL
		LIMIT 1
	`, id).Scan(&snapshotID, &existingHighlighted, &existingNotes, &existingMetadata); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return synthetic_scenarios.ErrSyntheticCalcuttaNotFound
		}
		return err
	}
	if snapshotID == nil || strings.TrimSpace(*snapshotID) == "" {
		return synthetic_scenarios.ErrSyntheticCalcuttaHasNoSnapshot
	}

	newHighlighted := existingHighlighted
	if p.HighlightedEntryID != nil {
		v := strings.TrimSpace(*p.HighlightedEntryID)
		if v == "" {
			newHighlighted = nil
		} else {
			var exists bool
			if err := r.pool.QueryRow(ctx, `
				SELECT EXISTS(
					SELECT 1
					FROM core.calcutta_snapshot_entries
					WHERE id = $1::uuid
						AND calcutta_snapshot_id = $2::uuid
						AND deleted_at IS NULL
				)
			`, v, *snapshotID).Scan(&exists); err != nil {
				return err
			}
			if !exists {
				return synthetic_scenarios.ErrHighlightedEntryDoesNotBelong
			}
			newHighlighted = &v
		}
	}

	newNotes := existingNotes
	if p.Notes != nil {
		v := strings.TrimSpace(*p.Notes)
		if v == "" {
			newNotes = nil
		} else {
			newNotes = &v
		}
	}

	newMetadata := existingMetadata
	if p.Metadata != nil {
		b := []byte(*p.Metadata)
		if len(b) == 0 {
			newMetadata = json.RawMessage([]byte("{}"))
		} else {
			newMetadata = json.RawMessage(b)
		}
	}
	if len(newMetadata) == 0 {
		newMetadata = json.RawMessage([]byte("{}"))
	}

	highlightedParam := any(nil)
	if newHighlighted != nil && strings.TrimSpace(*newHighlighted) != "" {
		highlightedParam = *newHighlighted
	}
	notesParam := any(nil)
	if newNotes != nil {
		notesParam = *newNotes
	}

	_, err := r.pool.Exec(ctx, `
		UPDATE derived.synthetic_calcuttas
		SET highlighted_snapshot_entry_id = $2::uuid,
			notes = $3,
			metadata_json = $4::jsonb,
			updated_at = NOW()
		WHERE id = $1::uuid
			AND deleted_at IS NULL
	`, id, highlightedParam, notesParam, []byte(newMetadata))
	return err
}

func (r *SyntheticScenariosRepository) CreateSyntheticCalcutta(ctx context.Context, p synthetic_scenarios.CreateSyntheticCalcuttaParams) (string, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return "", err
	}
	committed := false
	defer func() {
		if committed {
			return
		}
		_ = tx.Rollback(ctx)
	}()

	snapshotID := (*string)(nil)
	if p.CalcuttaSnapshotID != nil && strings.TrimSpace(*p.CalcuttaSnapshotID) != "" {
		v := strings.TrimSpace(*p.CalcuttaSnapshotID)
		snapshotID = &v
	} else {
		focusRunID := ""
		if p.FocusStrategyGenerationID != nil {
			focusRunID = strings.TrimSpace(*p.FocusStrategyGenerationID)
		}
		created, err := createSyntheticCalcuttaSnapshot(ctx, tx, p.CalcuttaID, p.ExcludedEntryName, focusRunID, p.FocusEntryName)
		if err != nil {
			return "", err
		}
		snapshotID = &created
	}

	focusRunParam := any(nil)
	if p.FocusStrategyGenerationID != nil && strings.TrimSpace(*p.FocusStrategyGenerationID) != "" {
		focusRunParam = strings.TrimSpace(*p.FocusStrategyGenerationID)
	}

	var syntheticID string
	if err := tx.QueryRow(ctx, `
		UPDATE derived.synthetic_calcuttas
		SET calcutta_snapshot_id = $3::uuid,
			focus_strategy_generation_run_id = $4::uuid,
			focus_entry_name = $5,
			starting_state_key = $6,
			excluded_entry_name = $7,
			updated_at = NOW(),
			deleted_at = NULL
		WHERE cohort_id = $1::uuid
			AND calcutta_id = $2::uuid
			AND deleted_at IS NULL
		RETURNING id::text
	`, p.CohortID, p.CalcuttaID, snapshotID, focusRunParam, p.FocusEntryName, p.StartingStateKey, p.ExcludedEntryName).Scan(&syntheticID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			if err := tx.QueryRow(ctx, `
				INSERT INTO derived.synthetic_calcuttas (
					cohort_id,
					calcutta_id,
					calcutta_snapshot_id,
					focus_strategy_generation_run_id,
					focus_entry_name,
					starting_state_key,
					excluded_entry_name
				)
				VALUES ($1::uuid, $2::uuid, $3::uuid, $4::uuid, $5, $6, $7)
				RETURNING id::text
			`, p.CohortID, p.CalcuttaID, snapshotID, focusRunParam, p.FocusEntryName, p.StartingStateKey, p.ExcludedEntryName).Scan(&syntheticID); err != nil {
				return "", err
			}
		} else {
			return "", err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return "", err
	}
	committed = true
	return syntheticID, nil
}

func (r *SyntheticScenariosRepository) ListSyntheticEntries(ctx context.Context, syntheticCalcuttaID string) ([]synthetic_scenarios.SyntheticEntryListItem, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{AccessMode: pgx.ReadOnly})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var snapshotID *string
	if err := tx.QueryRow(ctx, `
		SELECT calcutta_snapshot_id::text
		FROM derived.synthetic_calcuttas
		WHERE id = $1::uuid
			AND deleted_at IS NULL
		LIMIT 1
	`, syntheticCalcuttaID).Scan(&snapshotID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, synthetic_scenarios.ErrSyntheticCalcuttaNotFound
		}
		return nil, err
	}
	if snapshotID == nil || strings.TrimSpace(*snapshotID) == "" {
		return nil, synthetic_scenarios.ErrSyntheticCalcuttaHasNoSnapshot
	}

	rows, err := tx.Query(ctx, `
		WITH latest_eval AS (
			SELECT sr.calcutta_evaluation_run_id
			FROM derived.simulation_runs sr
			WHERE sr.synthetic_calcutta_id = $2::uuid
				AND sr.deleted_at IS NULL
				AND sr.calcutta_evaluation_run_id IS NOT NULL
			ORDER BY sr.created_at DESC
			LIMIT 1
		),
		perf AS (
			SELECT
				ROW_NUMBER() OVER (ORDER BY COALESCE(ep.mean_normalized_payout, 0.0) DESC)::int AS rank,
				ep.entry_name,
				COALESCE(ep.mean_normalized_payout, 0.0)::double precision AS mean_normalized_payout,
				COALESCE(ep.p_top1, 0.0)::double precision AS p_top1,
				COALESCE(ep.p_in_money, 0.0)::double precision AS p_in_money
			FROM derived.entry_performance ep
			JOIN latest_eval le
				ON le.calcutta_evaluation_run_id = ep.calcutta_evaluation_run_id
			WHERE ep.deleted_at IS NULL
		)
		SELECT
			scc.id::text,
			scc.candidate_id::text,
			scc.snapshot_entry_id::text,
			e.entry_id::text,
			e.display_name,
			e.is_synthetic,
			e.created_at,
			e.updated_at,
			et.team_id::text,
			et.bid_points,
			p.rank,
			p.mean_normalized_payout,
			p.p_top1,
			p.p_in_money
		FROM derived.synthetic_calcutta_candidates scc
		JOIN derived.candidates c
			ON c.id = scc.candidate_id
			AND c.deleted_at IS NULL
		JOIN core.calcutta_snapshot_entries e
			ON e.id = scc.snapshot_entry_id
			AND e.calcutta_snapshot_id = $1::uuid
			AND e.deleted_at IS NULL
		LEFT JOIN core.calcutta_snapshot_entry_teams et
			ON et.calcutta_snapshot_entry_id = e.id
			AND et.deleted_at IS NULL
		LEFT JOIN perf p
			ON p.entry_name = e.display_name
		WHERE scc.synthetic_calcutta_id = $2::uuid
			AND scc.deleted_at IS NULL
		ORDER BY scc.created_at ASC, et.bid_points DESC
	`, *snapshotID, syntheticCalcuttaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	byID := make(map[string]*synthetic_scenarios.SyntheticEntryListItem)
	order := make([]string, 0)
	for rows.Next() {
		var attachmentID string
		var candidateID string
		var snapshotEntryID string
		var sourceEntryID *string
		var displayName string
		var isSynthetic bool
		var createdAt time.Time
		var updatedAt time.Time
		var teamID *string
		var bidPoints *int
		var rank *int
		var mean *float64
		var pTop1 *float64
		var pInMoney *float64

		if err := rows.Scan(
			&attachmentID,
			&candidateID,
			&snapshotEntryID,
			&sourceEntryID,
			&displayName,
			&isSynthetic,
			&createdAt,
			&updatedAt,
			&teamID,
			&bidPoints,
			&rank,
			&mean,
			&pTop1,
			&pInMoney,
		); err != nil {
			return nil, err
		}

		it, ok := byID[attachmentID]
		if !ok {
			it = &synthetic_scenarios.SyntheticEntryListItem{
				ID:            attachmentID,
				CandidateID:   candidateID,
				SnapshotEntry: snapshotEntryID,
				EntryID:       sourceEntryID,
				DisplayName:   displayName,
				IsSynthetic:   isSynthetic,
				Rank:          rank,
				Mean:          mean,
				PTop1:         pTop1,
				PInMoney:      pInMoney,
				Teams:         make([]synthetic_scenarios.SyntheticEntryTeam, 0),
				CreatedAt:     createdAt,
				UpdatedAt:     updatedAt,
			}
			byID[attachmentID] = it
			order = append(order, attachmentID)
		}
		if teamID != nil && bidPoints != nil {
			it.Teams = append(it.Teams, synthetic_scenarios.SyntheticEntryTeam{TeamID: *teamID, BidPoints: *bidPoints})
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	items := make([]synthetic_scenarios.SyntheticEntryListItem, 0, len(order))
	for _, id := range order {
		items = append(items, *byID[id])
	}
	for i := range items {
		sort.Slice(items[i].Teams, func(a, b int) bool { return items[i].Teams[a].BidPoints > items[i].Teams[b].BidPoints })
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *SyntheticScenariosRepository) CreateSyntheticEntry(ctx context.Context, p synthetic_scenarios.CreateSyntheticEntryParams) (string, error) {
	var snapshotID *string
	if err := r.pool.QueryRow(ctx, `
		SELECT calcutta_snapshot_id::text
		FROM derived.synthetic_calcuttas
		WHERE id = $1::uuid
			AND deleted_at IS NULL
		LIMIT 1
	`, p.SyntheticCalcuttaID).Scan(&snapshotID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", synthetic_scenarios.ErrSyntheticCalcuttaNotFound
		}
		return "", err
	}
	if snapshotID == nil || strings.TrimSpace(*snapshotID) == "" {
		return "", synthetic_scenarios.ErrSyntheticCalcuttaHasNoSnapshot
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return "", err
	}
	committed := false
	defer func() {
		if committed {
			return
		}
		_ = tx.Rollback(ctx)
	}()

	metadata := map[string]any{"teams": p.Teams}
	metadataJSON, _ := json.Marshal(metadata)

	var candidateID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO derived.candidates (source_kind, source_entry_artifact_id, display_name, metadata_json)
		VALUES ('manual', NULL, $1, $2::jsonb)
		RETURNING id::text
	`, p.DisplayName, string(metadataJSON)).Scan(&candidateID); err != nil {
		return "", err
	}

	var snapshotEntryID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO core.calcutta_snapshot_entries (calcutta_snapshot_id, entry_id, display_name, is_synthetic)
		VALUES ($1::uuid, NULL, $2, true)
		RETURNING id::text
	`, *snapshotID, p.DisplayName).Scan(&snapshotEntryID); err != nil {
		return "", err
	}

	for _, t := range p.Teams {
		if _, err := tx.Exec(ctx, `
			INSERT INTO core.calcutta_snapshot_entry_teams (calcutta_snapshot_entry_id, team_id, bid_points)
			VALUES ($1::uuid, $2::uuid, $3::int)
		`, snapshotEntryID, t.TeamID, t.BidPoints); err != nil {
			return "", err
		}
	}

	var attachmentID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO derived.synthetic_calcutta_candidates (synthetic_calcutta_id, candidate_id, snapshot_entry_id)
		VALUES ($1::uuid, $2::uuid, $3::uuid)
		RETURNING id::text
	`, p.SyntheticCalcuttaID, candidateID, snapshotEntryID).Scan(&attachmentID); err != nil {
		return "", err
	}

	if err := tx.Commit(ctx); err != nil {
		return "", err
	}
	committed = true
	return attachmentID, nil
}

func (r *SyntheticScenariosRepository) ImportSyntheticEntry(ctx context.Context, p synthetic_scenarios.ImportSyntheticEntryParams) (string, int, error) {
	strategyGenerationRunID := ""
	artifactKind := ""
	if err := r.pool.QueryRow(ctx, `
		SELECT run_id::text, artifact_kind
		FROM derived.run_artifacts
		WHERE id = $1::uuid
			AND run_kind = 'strategy_generation'
			AND deleted_at IS NULL
		LIMIT 1
	`, p.EntryArtifactID).Scan(&strategyGenerationRunID, &artifactKind); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", 0, synthetic_scenarios.ErrEntryArtifactNotFound
		}
		return "", 0, err
	}
	strategyGenerationRunID = strings.TrimSpace(strategyGenerationRunID)
	if strategyGenerationRunID == "" {
		return "", 0, synthetic_scenarios.ErrEntryArtifactHasNoRunID
	}
	if strings.TrimSpace(artifactKind) != "metrics" {
		return "", 0, synthetic_scenarios.ErrEntryArtifactNotMetrics
	}

	var snapshotID *string
	if err := r.pool.QueryRow(ctx, `
		SELECT calcutta_snapshot_id::text
		FROM derived.synthetic_calcuttas
		WHERE id = $1::uuid
			AND deleted_at IS NULL
		LIMIT 1
	`, p.SyntheticCalcuttaID).Scan(&snapshotID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", 0, synthetic_scenarios.ErrSyntheticCalcuttaNotFound
		}
		return "", 0, err
	}
	if snapshotID == nil || strings.TrimSpace(*snapshotID) == "" {
		return "", 0, synthetic_scenarios.ErrSyntheticCalcuttaHasNoSnapshot
	}

	displayName := ""
	if p.DisplayName != nil {
		displayName = strings.TrimSpace(*p.DisplayName)
	}
	if displayName == "" {
		var resolved string
		_ = r.pool.QueryRow(ctx, `
			SELECT COALESCE(name, ''::text)
			FROM derived.strategy_generation_runs
			WHERE id = $1::uuid
				AND deleted_at IS NULL
			LIMIT 1
		`, strategyGenerationRunID).Scan(&resolved)
		resolved = strings.TrimSpace(resolved)
		if resolved == "" {
			displayName = "Imported Strategy"
		} else {
			displayName = resolved
		}
	}

	rows, err := r.pool.Query(ctx, `
		SELECT team_id::text, bid_points::int
		FROM derived.strategy_generation_run_bids
		WHERE strategy_generation_run_id = $1::uuid
			AND deleted_at IS NULL
		ORDER BY bid_points DESC
	`, strategyGenerationRunID)
	if err != nil {
		return "", 0, err
	}
	defer rows.Close()

	teams := make([]synthetic_scenarios.SyntheticEntryTeam, 0)
	for rows.Next() {
		var t synthetic_scenarios.SyntheticEntryTeam
		if err := rows.Scan(&t.TeamID, &t.BidPoints); err != nil {
			return "", 0, err
		}
		teams = append(teams, t)
	}
	if err := rows.Err(); err != nil {
		return "", 0, err
	}
	if len(teams) == 0 {
		return "", 0, synthetic_scenarios.ErrNoStrategyGenerationRunBids
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return "", 0, err
	}
	committed := false
	defer func() {
		if committed {
			return
		}
		_ = tx.Rollback(ctx)
	}()

	var candidateID string
	if err := tx.QueryRow(ctx, `
		SELECT id::text
		FROM derived.candidates
		WHERE source_kind = 'entry_artifact'
			AND source_entry_artifact_id = $1::uuid
			AND deleted_at IS NULL
		LIMIT 1
	`, p.EntryArtifactID).Scan(&candidateID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			if err := tx.QueryRow(ctx, `
				INSERT INTO derived.candidates (source_kind, source_entry_artifact_id, display_name, metadata_json)
				VALUES ('entry_artifact', $1::uuid, $2, '{}'::jsonb)
				RETURNING id::text
			`, p.EntryArtifactID, displayName).Scan(&candidateID); err != nil {
				return "", 0, err
			}
		} else {
			return "", 0, err
		}
	}

	var existingAttachmentID string
	if err := tx.QueryRow(ctx, `
		SELECT id::text
		FROM derived.synthetic_calcutta_candidates
		WHERE synthetic_calcutta_id = $1::uuid
			AND candidate_id = $2::uuid
			AND deleted_at IS NULL
		LIMIT 1
	`, p.SyntheticCalcuttaID, candidateID).Scan(&existingAttachmentID); err == nil {
		if err := tx.Commit(ctx); err != nil {
			return "", 0, err
		}
		committed = true
		return existingAttachmentID, len(teams), nil
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return "", 0, err
	}

	var snapshotEntryID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO core.calcutta_snapshot_entries (calcutta_snapshot_id, entry_id, display_name, is_synthetic)
		VALUES ($1::uuid, NULL, $2, true)
		RETURNING id::text
	`, *snapshotID, displayName).Scan(&snapshotEntryID); err != nil {
		return "", 0, err
	}

	for _, t := range teams {
		if _, err := tx.Exec(ctx, `
			INSERT INTO core.calcutta_snapshot_entry_teams (calcutta_snapshot_entry_id, team_id, bid_points)
			VALUES ($1::uuid, $2::uuid, $3::int)
		`, snapshotEntryID, t.TeamID, t.BidPoints); err != nil {
			return "", 0, err
		}
	}

	var attachmentID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO derived.synthetic_calcutta_candidates (synthetic_calcutta_id, candidate_id, snapshot_entry_id)
		VALUES ($1::uuid, $2::uuid, $3::uuid)
		RETURNING id::text
	`, p.SyntheticCalcuttaID, candidateID, snapshotEntryID).Scan(&attachmentID); err != nil {
		return "", 0, err
	}

	if err := tx.Commit(ctx); err != nil {
		return "", 0, err
	}
	committed = true
	return attachmentID, len(teams), nil
}

func (r *SyntheticScenariosRepository) PatchSyntheticEntry(ctx context.Context, p synthetic_scenarios.PatchSyntheticEntryParams) error {
	tx, err := r.pool.Begin(ctx)
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

	var candidateID string
	var snapshotEntryID string
	var sourceKind string
	if err := tx.QueryRow(ctx, `
		SELECT
			scc.candidate_id::text,
			scc.snapshot_entry_id::text,
			c.source_kind
		FROM derived.synthetic_calcutta_candidates scc
		JOIN derived.candidates c
			ON c.id = scc.candidate_id
			AND c.deleted_at IS NULL
		WHERE scc.id = $1::uuid
			AND scc.deleted_at IS NULL
		LIMIT 1
	`, p.AttachmentID).Scan(&candidateID, &snapshotEntryID, &sourceKind); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return synthetic_scenarios.ErrSyntheticEntryNotFound
		}
		return err
	}

	if p.DisplayName != nil {
		if _, err := tx.Exec(ctx, `
			UPDATE derived.candidates
			SET display_name = $2,
				updated_at = NOW()
			WHERE id = $1::uuid
				AND deleted_at IS NULL
		`, candidateID, *p.DisplayName); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `
			UPDATE core.calcutta_snapshot_entries
			SET display_name = $2,
				updated_at = NOW()
			WHERE id = $1::uuid
				AND deleted_at IS NULL
		`, snapshotEntryID, *p.DisplayName); err != nil {
			return err
		}
	}

	if p.Teams != nil {
		if sourceKind != "manual" {
			return synthetic_scenarios.ErrOnlyManualCandidatesCanBeEdited
		}

		teams := *p.Teams
		metadata := map[string]any{"teams": teams}
		metadataJSON, _ := json.Marshal(metadata)
		if _, err := tx.Exec(ctx, `
			UPDATE derived.candidates
			SET metadata_json = $2::jsonb,
				updated_at = NOW()
			WHERE id = $1::uuid
				AND deleted_at IS NULL
		`, candidateID, string(metadataJSON)); err != nil {
			return err
		}

		if _, err := tx.Exec(ctx, `
			DELETE FROM core.calcutta_snapshot_entry_teams
			WHERE calcutta_snapshot_entry_id = $1::uuid
		`, snapshotEntryID); err != nil {
			return err
		}

		for _, t := range teams {
			if _, err := tx.Exec(ctx, `
				INSERT INTO core.calcutta_snapshot_entry_teams (calcutta_snapshot_entry_id, team_id, bid_points)
				VALUES ($1::uuid, $2::uuid, $3::int)
			`, snapshotEntryID, t.TeamID, t.BidPoints); err != nil {
				return err
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	committed = true
	return nil
}

func (r *SyntheticScenariosRepository) DeleteSyntheticEntry(ctx context.Context, p synthetic_scenarios.DeleteSyntheticEntryParams) error {
	var snapshotEntryID string
	if err := r.pool.QueryRow(ctx, `
		SELECT snapshot_entry_id::text
		FROM derived.synthetic_calcutta_candidates
		WHERE id = $1::uuid
			AND deleted_at IS NULL
		LIMIT 1
	`, p.AttachmentID).Scan(&snapshotEntryID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return synthetic_scenarios.ErrSyntheticEntryNotFound
		}
		return err
	}

	tx, err := r.pool.Begin(ctx)
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

	if _, err := tx.Exec(ctx, `
		UPDATE derived.synthetic_calcutta_candidates
		SET deleted_at = NOW(),
			updated_at = NOW()
		WHERE id = $1::uuid
			AND deleted_at IS NULL
	`, p.AttachmentID); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `
		UPDATE core.calcutta_snapshot_entries
		SET deleted_at = NOW(),
			updated_at = NOW()
		WHERE id = $1::uuid
			AND deleted_at IS NULL
	`, snapshotEntryID); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `
		UPDATE derived.synthetic_calcuttas
		SET highlighted_snapshot_entry_id = NULL,
			updated_at = NOW()
		WHERE highlighted_snapshot_entry_id = $1::uuid
			AND deleted_at IS NULL
	`, snapshotEntryID); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	committed = true
	return nil
}
func createSyntheticCalcuttaSnapshot(ctx context.Context, tx pgx.Tx, calcuttaID string, excludedEntryName *string, focusStrategyGenerationRunID string, focusEntryName *string) (string, error) {
	var snapshotID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO core.calcutta_snapshots (base_calcutta_id, snapshot_type, description)
		VALUES ($1::uuid, 'synthetic_calcutta', 'Synthetic calcutta snapshot')
		RETURNING id
	`, calcuttaID).Scan(&snapshotID); err != nil {
		return "", err
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO core.calcutta_snapshot_payouts (calcutta_snapshot_id, position, amount_cents)
		SELECT $2, position, amount_cents
		FROM core.payouts
		WHERE calcutta_id = $1
			AND deleted_at IS NULL
	`, calcuttaID, snapshotID); err != nil {
		return "", err
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO core.calcutta_snapshot_scoring_rules (calcutta_snapshot_id, win_index, points_awarded)
		SELECT $2, win_index, points_awarded
		FROM core.calcutta_scoring_rules
		WHERE calcutta_id = $1
			AND deleted_at IS NULL
	`, calcuttaID, snapshotID); err != nil {
		return "", err
	}

	excluded := ""
	if excludedEntryName != nil {
		excluded = *excludedEntryName
	}

	entryRows, err := tx.Query(ctx, `
		SELECT id::text, name
		FROM core.entries
		WHERE calcutta_id = $1::uuid
			AND deleted_at IS NULL
			AND (name != $2 OR $2 = '')
		ORDER BY created_at ASC
	`, calcuttaID, excluded)
	if err != nil {
		return "", err
	}
	defer entryRows.Close()

	type entryRow struct {
		id   string
		name string
	}
	entries := make([]entryRow, 0)
	for entryRows.Next() {
		var id, name string
		if err := entryRows.Scan(&id, &name); err != nil {
			return "", err
		}
		entries = append(entries, entryRow{id: id, name: name})
	}
	if err := entryRows.Err(); err != nil {
		return "", err
	}

	for _, e := range entries {
		var snapshotEntryID string
		if err := tx.QueryRow(ctx, `
			INSERT INTO core.calcutta_snapshot_entries (calcutta_snapshot_id, entry_id, display_name, is_synthetic)
			VALUES ($1::uuid, $2::uuid, $3, false)
			RETURNING id
		`, snapshotID, e.id, e.name).Scan(&snapshotEntryID); err != nil {
			return "", err
		}

		if _, err := tx.Exec(ctx, `
			INSERT INTO core.calcutta_snapshot_entry_teams (calcutta_snapshot_entry_id, team_id, bid_points)
			SELECT $1::uuid, team_id, bid_points
			FROM core.entry_teams
			WHERE entry_id = $2::uuid
				AND deleted_at IS NULL
		`, snapshotEntryID, e.id); err != nil {
			return "", err
		}
	}

	if focusStrategyGenerationRunID != "" {
		name := "Our Strategy"
		if focusEntryName != nil && strings.TrimSpace(*focusEntryName) != "" {
			name = strings.TrimSpace(*focusEntryName)
		}

		var snapshotEntryID string
		if err := tx.QueryRow(ctx, `
			INSERT INTO core.calcutta_snapshot_entries (calcutta_snapshot_id, entry_id, display_name, is_synthetic)
			VALUES ($1::uuid, NULL, $2, true)
			RETURNING id
		`, snapshotID, name).Scan(&snapshotEntryID); err != nil {
			return "", err
		}

		ct, err := tx.Exec(ctx, `
			INSERT INTO core.calcutta_snapshot_entry_teams (calcutta_snapshot_entry_id, team_id, bid_points)
			SELECT $1::uuid, reb.team_id, reb.bid_points
			FROM derived.strategy_generation_run_bids reb
			WHERE reb.strategy_generation_run_id = $2::uuid
				AND reb.deleted_at IS NULL
		`, snapshotEntryID, focusStrategyGenerationRunID)
		if err != nil {
			return "", err
		}
		if ct.RowsAffected() == 0 {
			return "", pgx.ErrNoRows
		}
	}

	return snapshotID, nil
}
