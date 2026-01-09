package db

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/app/suite_scenarios"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SuiteScenariosRepository struct {
	pool *pgxpool.Pool
}

func NewSuiteScenariosRepository(pool *pgxpool.Pool) *SuiteScenariosRepository {
	return &SuiteScenariosRepository{pool: pool}
}

func (r *SuiteScenariosRepository) ListSimulatedCalcuttas(ctx context.Context, tournamentID *string, limit, offset int) ([]suite_scenarios.SimulatedCalcutta, error) {
	var tournamentParam any
	if tournamentID != nil && strings.TrimSpace(*tournamentID) != "" {
		tournamentParam = strings.TrimSpace(*tournamentID)
	} else {
		tournamentParam = nil
	}

	rows, err := r.pool.Query(ctx, `
		SELECT
			id::text,
			name,
			description,
			tournament_id::text,
			base_calcutta_id::text,
			starting_state_key,
			excluded_entry_name,
			highlighted_simulated_entry_id::text,
			metadata_json,
			created_at,
			updated_at
		FROM derived.simulated_calcuttas
		WHERE deleted_at IS NULL
			AND ($1::uuid IS NULL OR tournament_id = $1::uuid)
		ORDER BY created_at DESC
		LIMIT $2::int
		OFFSET $3::int
	`, tournamentParam, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]suite_scenarios.SimulatedCalcutta, 0)
	for rows.Next() {
		var it suite_scenarios.SimulatedCalcutta
		if err := rows.Scan(
			&it.ID,
			&it.Name,
			&it.Description,
			&it.TournamentID,
			&it.BaseCalcuttaID,
			&it.StartingStateKey,
			&it.ExcludedEntryName,
			&it.HighlightedSimulatedEntry,
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

func (r *SuiteScenariosRepository) GetSimulatedCalcutta(ctx context.Context, id string) (*suite_scenarios.SimulatedCalcutta, []suite_scenarios.SimulatedCalcuttaPayout, []suite_scenarios.SimulatedCalcuttaScoringRule, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{AccessMode: pgx.ReadOnly})
	if err != nil {
		return nil, nil, nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var it suite_scenarios.SimulatedCalcutta
	if err := tx.QueryRow(ctx, `
		SELECT
			id::text,
			name,
			description,
			tournament_id::text,
			base_calcutta_id::text,
			starting_state_key,
			excluded_entry_name,
			highlighted_simulated_entry_id::text,
			metadata_json,
			created_at,
			updated_at
		FROM derived.simulated_calcuttas
		WHERE id = $1::uuid
			AND deleted_at IS NULL
		LIMIT 1
	`, id).Scan(
		&it.ID,
		&it.Name,
		&it.Description,
		&it.TournamentID,
		&it.BaseCalcuttaID,
		&it.StartingStateKey,
		&it.ExcludedEntryName,
		&it.HighlightedSimulatedEntry,
		&it.Metadata,
		&it.CreatedAt,
		&it.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, nil, suite_scenarios.ErrSimulatedCalcuttaNotFound
		}
		return nil, nil, nil, err
	}

	payoutRows, err := tx.Query(ctx, `
		SELECT position::int, amount_cents::int
		FROM derived.simulated_calcutta_payouts
		WHERE simulated_calcutta_id = $1::uuid
			AND deleted_at IS NULL
		ORDER BY position ASC
	`, id)
	if err != nil {
		return nil, nil, nil, err
	}
	defer payoutRows.Close()

	payouts := make([]suite_scenarios.SimulatedCalcuttaPayout, 0)
	for payoutRows.Next() {
		var p suite_scenarios.SimulatedCalcuttaPayout
		if err := payoutRows.Scan(&p.Position, &p.AmountCents); err != nil {
			return nil, nil, nil, err
		}
		payouts = append(payouts, p)
	}
	if err := payoutRows.Err(); err != nil {
		return nil, nil, nil, err
	}

	ruleRows, err := tx.Query(ctx, `
		SELECT win_index::int, points_awarded::int
		FROM derived.simulated_calcutta_scoring_rules
		WHERE simulated_calcutta_id = $1::uuid
			AND deleted_at IS NULL
		ORDER BY win_index ASC
	`, id)
	if err != nil {
		return nil, nil, nil, err
	}
	defer ruleRows.Close()

	rules := make([]suite_scenarios.SimulatedCalcuttaScoringRule, 0)
	for ruleRows.Next() {
		var rr suite_scenarios.SimulatedCalcuttaScoringRule
		if err := ruleRows.Scan(&rr.WinIndex, &rr.PointsAwarded); err != nil {
			return nil, nil, nil, err
		}
		rules = append(rules, rr)
	}
	if err := ruleRows.Err(); err != nil {
		return nil, nil, nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, nil, nil, err
	}

	return &it, payouts, rules, nil
}

func (r *SuiteScenariosRepository) CreateSimulatedCalcutta(ctx context.Context, p suite_scenarios.CreateSimulatedCalcuttaParams) (string, error) {
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

	var createdID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO derived.simulated_calcuttas (
			name,
			description,
			tournament_id,
			base_calcutta_id,
			starting_state_key,
			excluded_entry_name,
			metadata_json
		)
		VALUES ($1, $2, $3::uuid, NULL, $4, $5, $6::jsonb)
		RETURNING id::text
	`, p.Name, p.Description, p.TournamentID, p.StartingStateKey, p.ExcludedEntryName, []byte(p.Metadata)).Scan(&createdID); err != nil {
		return "", err
	}

	for _, payout := range p.Payouts {
		if _, err := tx.Exec(ctx, `
			INSERT INTO derived.simulated_calcutta_payouts (
				simulated_calcutta_id,
				position,
				amount_cents
			)
			VALUES ($1::uuid, $2::int, $3::int)
		`, createdID, payout.Position, payout.AmountCents); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				return "", suite_scenarios.ErrDuplicatePayoutPosition
			}
			return "", err
		}
	}

	for _, sr := range p.ScoringRules {
		if _, err := tx.Exec(ctx, `
			INSERT INTO derived.simulated_calcutta_scoring_rules (
				simulated_calcutta_id,
				win_index,
				points_awarded
			)
			VALUES ($1::uuid, $2::int, $3::int)
		`, createdID, sr.WinIndex, sr.PointsAwarded); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				return "", suite_scenarios.ErrDuplicateScoringRuleWinIndex
			}
			return "", err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return "", err
	}
	committed = true
	return createdID, nil
}

func (r *SuiteScenariosRepository) CreateSimulatedCalcuttaFromCalcutta(ctx context.Context, p suite_scenarios.CreateSimulatedCalcuttaFromCalcuttaParams) (string, int, error) {
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

	var tournamentID string
	var calcuttaName string
	if err := tx.QueryRow(ctx, `
		SELECT tournament_id::text, name
		FROM core.calcuttas
		WHERE id = $1::uuid
			AND deleted_at IS NULL
		LIMIT 1
	`, p.CalcuttaID).Scan(&tournamentID, &calcuttaName); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", 0, suite_scenarios.ErrCalcuttaNotFound
		}
		return "", 0, err
	}

	name := "Simulated " + strings.TrimSpace(calcuttaName)
	if p.Name != nil {
		v := strings.TrimSpace(*p.Name)
		if v != "" {
			name = v
		}
	}
	var desc any
	if p.Description != nil {
		v := strings.TrimSpace(*p.Description)
		if v == "" {
			desc = nil
		} else {
			desc = v
		}
	} else {
		desc = nil
	}

	var createdID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO derived.simulated_calcuttas (
			name,
			description,
			tournament_id,
			base_calcutta_id,
			starting_state_key,
			excluded_entry_name,
			metadata_json
		)
		VALUES ($1, $2, $3::uuid, $4::uuid, $5, $6, $7::jsonb)
		RETURNING id::text
	`, name, desc, tournamentID, p.CalcuttaID, p.StartingStateKey, p.ExcludedEntryName, []byte(p.Metadata)).Scan(&createdID); err != nil {
		return "", 0, err
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO derived.simulated_calcutta_payouts (simulated_calcutta_id, position, amount_cents)
		SELECT $1::uuid, position, amount_cents
		FROM core.payouts
		WHERE calcutta_id = $2::uuid
			AND deleted_at IS NULL
	`, createdID, p.CalcuttaID); err != nil {
		return "", 0, err
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO derived.simulated_calcutta_scoring_rules (simulated_calcutta_id, win_index, points_awarded)
		SELECT $1::uuid, win_index, points_awarded
		FROM core.calcutta_scoring_rules
		WHERE calcutta_id = $2::uuid
			AND deleted_at IS NULL
	`, createdID, p.CalcuttaID); err != nil {
		return "", 0, err
	}

	entryRows, err := tx.Query(ctx, `
		SELECT id::text, name
		FROM core.entries
		WHERE calcutta_id = $1::uuid
			AND deleted_at IS NULL
		ORDER BY created_at ASC
	`, p.CalcuttaID)
	if err != nil {
		return "", 0, err
	}
	defer entryRows.Close()

	copiedEntries := 0
	for entryRows.Next() {
		var entryID string
		var entryName string
		if err := entryRows.Scan(&entryID, &entryName); err != nil {
			return "", 0, err
		}

		var simulatedEntryID string
		if err := tx.QueryRow(ctx, `
			INSERT INTO derived.simulated_entries (
				simulated_calcutta_id,
				display_name,
				source_kind,
				source_entry_id,
				source_candidate_id
			)
			VALUES ($1::uuid, $2, 'from_real_entry', $3::uuid, NULL)
			RETURNING id::text
		`, createdID, strings.TrimSpace(entryName), entryID).Scan(&simulatedEntryID); err != nil {
			return "", 0, err
		}

		teamRows, err := tx.Query(ctx, `
			SELECT team_id::text, bid_points::int
			FROM core.entry_teams
			WHERE entry_id = $1::uuid
				AND deleted_at IS NULL
			ORDER BY bid_points DESC
		`, entryID)
		if err != nil {
			return "", 0, err
		}

		for teamRows.Next() {
			var teamID string
			var bidPoints int
			if err := teamRows.Scan(&teamID, &bidPoints); err != nil {
				teamRows.Close()
				return "", 0, err
			}
			if _, err := tx.Exec(ctx, `
				INSERT INTO derived.simulated_entry_teams (
					simulated_entry_id,
					team_id,
					bid_points
				)
				VALUES ($1::uuid, $2::uuid, $3::int)
			`, simulatedEntryID, teamID, bidPoints); err != nil {
				teamRows.Close()
				return "", 0, err
			}
		}
		if err := teamRows.Err(); err != nil {
			teamRows.Close()
			return "", 0, err
		}
		teamRows.Close()

		copiedEntries++
	}
	if err := entryRows.Err(); err != nil {
		return "", 0, err
	}

	if err := tx.Commit(ctx); err != nil {
		return "", 0, err
	}
	committed = true
	return createdID, copiedEntries, nil
}

func (r *SuiteScenariosRepository) PatchSimulatedCalcutta(ctx context.Context, id string, p suite_scenarios.PatchSimulatedCalcuttaParams) error {
	var existingName string
	var existingDescription *string
	var existingHighlighted *string
	var existingMetadata json.RawMessage
	if err := r.pool.QueryRow(ctx, `
		SELECT name, description, highlighted_simulated_entry_id::text, metadata_json
		FROM derived.simulated_calcuttas
		WHERE id = $1::uuid
			AND deleted_at IS NULL
		LIMIT 1
	`, id).Scan(&existingName, &existingDescription, &existingHighlighted, &existingMetadata); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return suite_scenarios.ErrSimulatedCalcuttaNotFound
		}
		return err
	}

	newName := existingName
	if p.Name != nil {
		newName = strings.TrimSpace(*p.Name)
	}

	newDescription := existingDescription
	if p.Description != nil {
		v := strings.TrimSpace(*p.Description)
		if v == "" {
			newDescription = nil
		} else {
			newDescription = &v
		}
	}

	newHighlighted := existingHighlighted
	if p.HighlightedSimulatedEntry != nil {
		v := strings.TrimSpace(*p.HighlightedSimulatedEntry)
		if v == "" {
			newHighlighted = nil
		} else {
			var ok bool
			if err := r.pool.QueryRow(ctx, `
				SELECT EXISTS(
					SELECT 1
					FROM derived.simulated_entries e
					WHERE e.id = $1::uuid
						AND e.simulated_calcutta_id = $2::uuid
						AND e.deleted_at IS NULL
				)
			`, v, id).Scan(&ok); err != nil {
				return err
			}
			if !ok {
				return suite_scenarios.ErrHighlightedEntryDoesNotBelong
			}
			newHighlighted = &v
		}
	}

	newMetadata := existingMetadata
	if p.Metadata != nil {
		newMetadata = *p.Metadata
	}
	if len(newMetadata) == 0 {
		newMetadata = json.RawMessage([]byte("{}"))
	}

	var highlightedParam any
	if newHighlighted != nil && strings.TrimSpace(*newHighlighted) != "" {
		highlightedParam = *newHighlighted
	} else {
		highlightedParam = nil
	}

	_, err := r.pool.Exec(ctx, `
		UPDATE derived.simulated_calcuttas
		SET name = $2,
			description = $3,
			highlighted_simulated_entry_id = $4::uuid,
			metadata_json = $5::jsonb,
			updated_at = NOW()
		WHERE id = $1::uuid
			AND deleted_at IS NULL
	`, id, newName, newDescription, highlightedParam, []byte(newMetadata))
	return err
}

func (r *SuiteScenariosRepository) ReplaceSimulatedCalcuttaRules(ctx context.Context, id string, p suite_scenarios.ReplaceSimulatedCalcuttaRulesParams) error {
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

	var exists bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1
			FROM derived.simulated_calcuttas
			WHERE id = $1::uuid
				AND deleted_at IS NULL
		)
	`, id).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return suite_scenarios.ErrSimulatedCalcuttaNotFound
	}

	if _, err := tx.Exec(ctx, `
		UPDATE derived.simulated_calcutta_payouts
		SET deleted_at = NOW(),
			updated_at = NOW()
		WHERE simulated_calcutta_id = $1::uuid
			AND deleted_at IS NULL
	`, id); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `
		UPDATE derived.simulated_calcutta_scoring_rules
		SET deleted_at = NOW(),
			updated_at = NOW()
		WHERE simulated_calcutta_id = $1::uuid
			AND deleted_at IS NULL
	`, id); err != nil {
		return err
	}

	for _, payout := range p.Payouts {
		if _, err := tx.Exec(ctx, `
			INSERT INTO derived.simulated_calcutta_payouts (
				simulated_calcutta_id,
				position,
				amount_cents
			)
			VALUES ($1::uuid, $2::int, $3::int)
		`, id, payout.Position, payout.AmountCents); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				return suite_scenarios.ErrDuplicatePayoutPosition
			}
			return err
		}
	}

	for _, sr := range p.ScoringRules {
		if _, err := tx.Exec(ctx, `
			INSERT INTO derived.simulated_calcutta_scoring_rules (
				simulated_calcutta_id,
				win_index,
				points_awarded
			)
			VALUES ($1::uuid, $2::int, $3::int)
		`, id, sr.WinIndex, sr.PointsAwarded); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				return suite_scenarios.ErrDuplicateScoringRuleWinIndex
			}
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	committed = true
	return nil
}

func (r *SuiteScenariosRepository) ListSimulatedEntries(ctx context.Context, simulatedCalcuttaID string) (bool, []suite_scenarios.SimulatedEntry, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{AccessMode: pgx.ReadOnly})
	if err != nil {
		return false, nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var exists bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1
			FROM derived.simulated_calcuttas
			WHERE id = $1::uuid
				AND deleted_at IS NULL
		)
	`, simulatedCalcuttaID).Scan(&exists); err != nil {
		return false, nil, err
	}
	if !exists {
		return false, nil, nil
	}

	rows, err := tx.Query(ctx, `
		SELECT
			e.id::text,
			e.simulated_calcutta_id::text,
			e.display_name,
			e.source_kind,
			e.source_entry_id::text,
			e.source_candidate_id::text,
			e.created_at,
			e.updated_at,
			et.team_id::text,
			et.bid_points::int
		FROM derived.simulated_entries e
		LEFT JOIN derived.simulated_entry_teams et
			ON et.simulated_entry_id = e.id
			AND et.deleted_at IS NULL
		WHERE e.simulated_calcutta_id = $1::uuid
			AND e.deleted_at IS NULL
		ORDER BY e.created_at ASC, et.bid_points DESC
	`, simulatedCalcuttaID)
	if err != nil {
		return false, nil, err
	}
	defer rows.Close()

	byID := make(map[string]*suite_scenarios.SimulatedEntry)
	order := make([]string, 0)
	for rows.Next() {
		var entryID string
		var scID string
		var displayName string
		var sourceKind string
		var sourceEntryID *string
		var sourceCandidateID *string
		var createdAt, updatedAt time.Time
		var teamID *string
		var bidPoints *int

		if err := rows.Scan(
			&entryID,
			&scID,
			&displayName,
			&sourceKind,
			&sourceEntryID,
			&sourceCandidateID,
			&createdAt,
			&updatedAt,
			&teamID,
			&bidPoints,
		); err != nil {
			return false, nil, err
		}

		it, ok := byID[entryID]
		if !ok {
			it = &suite_scenarios.SimulatedEntry{
				ID:                  entryID,
				SimulatedCalcuttaID: scID,
				DisplayName:         displayName,
				SourceKind:          sourceKind,
				SourceEntryID:       sourceEntryID,
				SourceCandidateID:   sourceCandidateID,
				Teams:               make([]suite_scenarios.SimulatedEntryTeam, 0),
				CreatedAt:           createdAt,
				UpdatedAt:           updatedAt,
			}
			byID[entryID] = it
			order = append(order, entryID)
		}
		if teamID != nil && bidPoints != nil {
			it.Teams = append(it.Teams, suite_scenarios.SimulatedEntryTeam{TeamID: *teamID, BidPoints: *bidPoints})
		}
	}
	if err := rows.Err(); err != nil {
		return false, nil, err
	}

	items := make([]suite_scenarios.SimulatedEntry, 0, len(order))
	for _, id := range order {
		items = append(items, *byID[id])
	}

	if err := tx.Commit(ctx); err != nil {
		return false, nil, err
	}

	return true, items, nil
}

func (r *SuiteScenariosRepository) CreateSimulatedEntry(ctx context.Context, p suite_scenarios.CreateSimulatedEntryParams) (string, error) {
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

	var exists bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1
			FROM derived.simulated_calcuttas
			WHERE id = $1::uuid
				AND deleted_at IS NULL
		)
	`, p.SimulatedCalcuttaID).Scan(&exists); err != nil {
		return "", err
	}
	if !exists {
		return "", suite_scenarios.ErrSimulatedCalcuttaNotFound
	}

	var entryID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO derived.simulated_entries (
			simulated_calcutta_id,
			display_name,
			source_kind,
			source_entry_id,
			source_candidate_id
		)
		VALUES ($1::uuid, $2, 'manual', NULL, NULL)
		RETURNING id::text
	`, p.SimulatedCalcuttaID, p.DisplayName).Scan(&entryID); err != nil {
		return "", err
	}

	for _, t := range p.Teams {
		if _, err := tx.Exec(ctx, `
			INSERT INTO derived.simulated_entry_teams (
				simulated_entry_id,
				team_id,
				bid_points
			)
			VALUES ($1::uuid, $2::uuid, $3::int)
		`, entryID, t.TeamID, t.BidPoints); err != nil {
			return "", err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return "", err
	}
	committed = true
	return entryID, nil
}

func (r *SuiteScenariosRepository) PatchSimulatedEntry(ctx context.Context, p suite_scenarios.PatchSimulatedEntryParams) error {
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

	var existingSimulatedCalcuttaID string
	if err := tx.QueryRow(ctx, `
		SELECT simulated_calcutta_id::text
		FROM derived.simulated_entries
		WHERE id = $1::uuid
			AND deleted_at IS NULL
		LIMIT 1
	`, p.EntryID).Scan(&existingSimulatedCalcuttaID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return suite_scenarios.ErrSimulatedEntryNotFound
		}
		return err
	}
	if existingSimulatedCalcuttaID != p.SimulatedCalcuttaID {
		return suite_scenarios.ErrSimulatedEntryNotFound
	}

	if p.DisplayName != nil {
		if _, err := tx.Exec(ctx, `
			UPDATE derived.simulated_entries
			SET display_name = $2,
				updated_at = NOW()
			WHERE id = $1::uuid
				AND deleted_at IS NULL
		`, p.EntryID, *p.DisplayName); err != nil {
			return err
		}
	}

	if p.Teams != nil {
		teams := *p.Teams
		if _, err := tx.Exec(ctx, `
			UPDATE derived.simulated_entry_teams
			SET deleted_at = NOW(),
				updated_at = NOW()
			WHERE simulated_entry_id = $1::uuid
				AND deleted_at IS NULL
		`, p.EntryID); err != nil {
			return err
		}
		for _, t := range teams {
			if _, err := tx.Exec(ctx, `
				INSERT INTO derived.simulated_entry_teams (
					simulated_entry_id,
					team_id,
					bid_points
				)
				VALUES ($1::uuid, $2::uuid, $3::int)
			`, p.EntryID, t.TeamID, t.BidPoints); err != nil {
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

func (r *SuiteScenariosRepository) DeleteSimulatedEntry(ctx context.Context, p suite_scenarios.DeleteSimulatedEntryParams) error {
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

	var existingSimulatedCalcuttaID string
	if err := tx.QueryRow(ctx, `
		SELECT simulated_calcutta_id::text
		FROM derived.simulated_entries
		WHERE id = $1::uuid
			AND deleted_at IS NULL
		LIMIT 1
	`, p.EntryID).Scan(&existingSimulatedCalcuttaID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return suite_scenarios.ErrSimulatedEntryNotFound
		}
		return err
	}
	if existingSimulatedCalcuttaID != p.SimulatedCalcuttaID {
		return suite_scenarios.ErrSimulatedEntryNotFound
	}

	if _, err := tx.Exec(ctx, `
		UPDATE derived.simulated_entry_teams
		SET deleted_at = NOW(),
			updated_at = NOW()
		WHERE simulated_entry_id = $1::uuid
			AND deleted_at IS NULL
	`, p.EntryID); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `
		UPDATE derived.simulated_entries
		SET deleted_at = NOW(),
			updated_at = NOW()
		WHERE id = $1::uuid
			AND deleted_at IS NULL
	`, p.EntryID); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `
		UPDATE derived.simulated_calcuttas
		SET highlighted_simulated_entry_id = NULL,
			updated_at = NOW()
		WHERE id = $1::uuid
			AND highlighted_simulated_entry_id = $2::uuid
			AND deleted_at IS NULL
	`, p.SimulatedCalcuttaID, p.EntryID); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	committed = true
	return nil
}

func (r *SuiteScenariosRepository) ImportCandidateAsSimulatedEntry(ctx context.Context, p suite_scenarios.ImportCandidateAsSimulatedEntryParams) (string, int, error) {
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

	var exists bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1
			FROM derived.simulated_calcuttas
			WHERE id = $1::uuid
				AND deleted_at IS NULL
		)
	`, p.SimulatedCalcuttaID).Scan(&exists); err != nil {
		return "", 0, err
	}
	if !exists {
		return "", 0, suite_scenarios.ErrSimulatedCalcuttaNotFound
	}

	var candidateDisplayName string
	var strategyGenerationRunID *string
	var metadataJSON []byte
	if err := tx.QueryRow(ctx, `
		SELECT display_name, strategy_generation_run_id::text, metadata_json
		FROM derived.candidates
		WHERE id = $1::uuid
			AND deleted_at IS NULL
		LIMIT 1
	`, p.CandidateID).Scan(&candidateDisplayName, &strategyGenerationRunID, &metadataJSON); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", 0, suite_scenarios.ErrCandidateNotFound
		}
		return "", 0, err
	}

	name := strings.TrimSpace(candidateDisplayName)
	if p.DisplayName != nil {
		name = strings.TrimSpace(*p.DisplayName)
	}
	if name == "" {
		name = "Candidate"
	}

	type parsedMetadata struct {
		Teams []suite_scenarios.SimulatedEntryTeam `json:"teams"`
	}

	teams := make([]suite_scenarios.SimulatedEntryTeam, 0)
	if strategyGenerationRunID != nil && strings.TrimSpace(*strategyGenerationRunID) != "" {
		rows, err := tx.Query(ctx, `
			SELECT team_id::text, bid_points::int
			FROM derived.strategy_generation_run_bids
			WHERE strategy_generation_run_id = $1::uuid
				AND deleted_at IS NULL
			ORDER BY bid_points DESC
		`, *strategyGenerationRunID)
		if err != nil {
			return "", 0, err
		}
		for rows.Next() {
			var teamID string
			var bidPoints int
			if err := rows.Scan(&teamID, &bidPoints); err != nil {
				rows.Close()
				return "", 0, err
			}
			teams = append(teams, suite_scenarios.SimulatedEntryTeam{TeamID: teamID, BidPoints: bidPoints})
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return "", 0, err
		}
		rows.Close()
	}

	if len(teams) == 0 {
		var parsed parsedMetadata
		_ = json.Unmarshal(metadataJSON, &parsed)
		teams = parsed.Teams
	}

	if len(teams) == 0 {
		return "", 0, suite_scenarios.ErrCandidateHasNoBids
	}

	for i := range teams {
		teams[i].TeamID = strings.TrimSpace(teams[i].TeamID)
		if teams[i].TeamID == "" {
			return "", 0, suite_scenarios.ErrCandidateInvalidTeamID
		}
		if _, err := uuid.Parse(teams[i].TeamID); err != nil {
			return "", 0, suite_scenarios.ErrCandidateInvalidTeamID
		}
		if teams[i].BidPoints <= 0 {
			return "", 0, suite_scenarios.ErrCandidateInvalidBidPoints
		}
	}

	var entryID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO derived.simulated_entries (
			simulated_calcutta_id,
			display_name,
			source_kind,
			source_entry_id,
			source_candidate_id
		)
		VALUES ($1::uuid, $2, 'from_candidate', NULL, $3::uuid)
		RETURNING id::text
	`, p.SimulatedCalcuttaID, name, p.CandidateID).Scan(&entryID); err != nil {
		return "", 0, err
	}

	for _, t := range teams {
		if _, err := tx.Exec(ctx, `
			INSERT INTO derived.simulated_entry_teams (
				simulated_entry_id,
				team_id,
				bid_points
			)
			VALUES ($1::uuid, $2::uuid, $3::int)
		`, entryID, t.TeamID, t.BidPoints); err != nil {
			return "", 0, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return "", 0, err
	}
	committed = true
	return entryID, len(teams), nil
}
