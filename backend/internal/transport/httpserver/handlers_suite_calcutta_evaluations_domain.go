package httpserver

import (
	"context"
	"time"

	appcalcutta "github.com/andrewcopp/Calcutta/backend/internal/app/calcutta"
	"github.com/andrewcopp/Calcutta/backend/pkg/models"
)

func (s *Server) computeHypotheticalFinishByEntryNameForStrategyRun(ctx context.Context, calcuttaID string, strategyGenerationRunID string) (map[string]*suiteCalcuttaEvalFinish, bool, error) {
	if calcuttaID == "" || strategyGenerationRunID == "" {
		return nil, false, nil
	}

	// Load payouts
	payoutRows, err := s.pool.Query(ctx, `
		SELECT position::int, amount_cents::int
		FROM core.payouts
		WHERE calcutta_id = $1::uuid
			AND deleted_at IS NULL
		ORDER BY position ASC
	`, calcuttaID)
	if err != nil {
		return nil, false, err
	}
	defer payoutRows.Close()

	payouts := make([]*models.CalcuttaPayout, 0)
	for payoutRows.Next() {
		var pos int
		var cents int
		if err := payoutRows.Scan(&pos, &cents); err != nil {
			return nil, false, err
		}
		payouts = append(payouts, &models.CalcuttaPayout{CalcuttaID: calcuttaID, Position: pos, AmountCents: cents})
	}
	if err := payoutRows.Err(); err != nil {
		return nil, false, err
	}

	// Load team points (actual wins/byes)
	teamRows, err := s.pool.Query(ctx, `
		WITH t AS (
			SELECT tournament_id
			FROM core.calcuttas
			WHERE id = $1::uuid
				AND deleted_at IS NULL
			LIMIT 1
		)
		SELECT
			team.id::text,
			core.calcutta_points_for_progress($1::uuid, team.wins, team.byes)::float8
		FROM core.teams team
		JOIN t ON t.tournament_id = team.tournament_id
		WHERE team.deleted_at IS NULL
	`, calcuttaID)
	if err != nil {
		return nil, false, err
	}
	defer teamRows.Close()

	teamPoints := make(map[string]float64)
	for teamRows.Next() {
		var teamID string
		var pts float64
		if err := teamRows.Scan(&teamID, &pts); err != nil {
			return nil, false, err
		}
		teamPoints[teamID] = pts
	}
	if err := teamRows.Err(); err != nil {
		return nil, false, err
	}

	// Load real entries and their bids
	rows, err := s.pool.Query(ctx, `
		SELECT
			e.id::text,
			e.name,
			e.created_at,
			et.team_id::text,
			et.bid_points::int
		FROM core.entries e
		JOIN core.entry_teams et ON et.entry_id = e.id AND et.deleted_at IS NULL
		WHERE e.calcutta_id = $1::uuid
			AND e.deleted_at IS NULL
		ORDER BY e.created_at ASC
	`, calcuttaID)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()

	entryByID := make(map[string]*models.CalcuttaEntry)
	entryBids := make(map[string]map[string]float64)
	existingTotalBid := make(map[string]float64)
	for rows.Next() {
		var entryID, name, teamID string
		var createdAt time.Time
		var bid int
		if err := rows.Scan(&entryID, &name, &createdAt, &teamID, &bid); err != nil {
			return nil, false, err
		}
		if _, ok := entryByID[entryID]; !ok {
			entryByID[entryID] = &models.CalcuttaEntry{ID: entryID, Name: name, CalcuttaID: calcuttaID, Created: createdAt}
			entryBids[entryID] = make(map[string]float64)
		}
		entryBids[entryID][teamID] += float64(bid)
		existingTotalBid[teamID] += float64(bid)
	}
	if err := rows.Err(); err != nil {
		return nil, false, err
	}

	// Load our strategy bids
	ourBids := make(map[string]float64)
	ourRows, err := s.pool.Query(ctx, `
		SELECT team_id::text, bid_points::int
		FROM derived.recommended_entry_bids
		WHERE strategy_generation_run_id = $1::uuid
			AND deleted_at IS NULL
	`, strategyGenerationRunID)
	if err != nil {
		return nil, false, err
	}
	defer ourRows.Close()
	for ourRows.Next() {
		var teamID string
		var bid int
		if err := ourRows.Scan(&teamID, &bid); err != nil {
			return nil, false, err
		}
		ourBids[teamID] += float64(bid)
	}
	if err := ourRows.Err(); err != nil {
		return nil, false, err
	}
	if len(ourBids) == 0 {
		return nil, false, nil
	}

	entries := make([]*models.CalcuttaEntry, 0, len(entryByID)+1)
	for entryID, e := range entryByID {
		total := 0.0
		for teamID, bid := range entryBids[entryID] {
			pts, ok := teamPoints[teamID]
			if !ok {
				continue
			}
			den := existingTotalBid[teamID] + ourBids[teamID]
			if den <= 0 {
				continue
			}
			total += pts * (bid / den)
		}
		e.TotalPoints = total
		entries = append(entries, e)
	}

	ourID := "our_strategy"
	ourCreated := time.Now()
	ourTotal := 0.0
	for teamID, bid := range ourBids {
		pts, ok := teamPoints[teamID]
		if !ok {
			continue
		}
		den := existingTotalBid[teamID] + bid
		if den <= 0 {
			continue
		}
		ourTotal += pts * (bid / den)
	}
	entries = append(entries, &models.CalcuttaEntry{ID: ourID, Name: "Our Strategy", CalcuttaID: calcuttaID, TotalPoints: ourTotal, Created: ourCreated})

	_, results := appcalcutta.ComputeEntryPlacementsAndPayouts(entries, payouts)
	out := make(map[string]*suiteCalcuttaEvalFinish)
	for _, e := range entries {
		res, ok := results[e.ID]
		if !ok {
			continue
		}
		out[e.Name] = &suiteCalcuttaEvalFinish{
			FinishPosition: res.FinishPosition,
			IsTied:         res.IsTied,
			InTheMoney:     res.InTheMoney,
			PayoutCents:    res.PayoutCents,
			TotalPoints:    e.TotalPoints,
		}
	}
	return out, true, nil
}
