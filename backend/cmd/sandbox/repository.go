package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type entryBid struct {
	EntryName string
	TeamID    string
	Bid       float64
}

func calcuttaYear(ctx context.Context, db *sql.DB, calcuttaID string) (int, error) {
	query := `
		SELECT COALESCE(substring(t.name from '([0-9]{4})')::int, 0) as tournament_year
		FROM calcuttas c
		JOIN tournaments t ON t.id = c.tournament_id AND t.deleted_at IS NULL
		WHERE c.deleted_at IS NULL
			AND c.id = $1::uuid
	`

	var year int
	if err := db.QueryRowContext(ctx, query, calcuttaID).Scan(&year); err != nil {
		return 0, err
	}
	if year == 0 {
		return 0, fmt.Errorf("failed to parse tournament year for calcutta-id %s", calcuttaID)
	}
	return year, nil
}

func computeSeedMeans(ctx context.Context, db *sql.DB, excludeCalcuttaID string, trainYears int, minYear int, maxYear int, excludeEntryName string) (map[int]float64, map[int]float64, error) {
	query := `
		WITH team_bids AS (
			SELECT
				c.id as calcutta_id,
				tt.seed,
				CASE (tt.wins + tt.byes)
					WHEN 0 THEN 0
					WHEN 1 THEN 0
					WHEN 2 THEN 50
					WHEN 3 THEN 150
					WHEN 4 THEN 300
					WHEN 5 THEN 500
					WHEN 6 THEN 750
					WHEN 7 THEN 1050
					ELSE 0
				END::float as team_points,
				COALESCE(SUM(
					CASE
						WHEN $5 <> '' AND ce.name = $5 THEN 0
						ELSE COALESCE(cet.bid, 0)
					END
				), 0)::float as total_bid
			FROM tournament_teams tt
			JOIN calcuttas c ON c.tournament_id = tt.tournament_id AND c.deleted_at IS NULL
			JOIN tournaments t ON t.id = c.tournament_id AND t.deleted_at IS NULL
			LEFT JOIN calcutta_entries ce ON ce.calcutta_id = c.id AND ce.deleted_at IS NULL
			LEFT JOIN calcutta_entry_teams cet ON cet.entry_id = ce.id AND cet.team_id = tt.id AND cet.deleted_at IS NULL
			WHERE tt.deleted_at IS NULL
				AND c.id <> $1::uuid
				AND (
					$2 = 0
					OR
					(
						COALESCE(substring(t.name from '([0-9]{4})')::int, 0) <> 0
						AND COALESCE(substring(t.name from '([0-9]{4})')::int, 0) >= $3
						AND COALESCE(substring(t.name from '([0-9]{4})')::int, 0) <= $4
					)
				)
			GROUP BY c.id, tt.id, tt.seed, team_points
		),
		enriched AS (
			SELECT
				calcutta_id,
				seed,
				team_points,
				CASE
					WHEN SUM(total_bid) OVER (PARTITION BY calcutta_id) > 0 THEN (total_bid / SUM(total_bid) OVER (PARTITION BY calcutta_id))
					ELSE 0
				END as bid_share
			FROM team_bids
		)
		SELECT
			seed,
			AVG(team_points) as mean_points,
			AVG(bid_share) as mean_bid_share
		FROM enriched
		GROUP BY seed
		ORDER BY seed
	`

	rows, err := db.QueryContext(ctx, query, excludeCalcuttaID, trainYears, minYear, maxYear, excludeEntryName)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	seedPoints := map[int]float64{}
	seedBidShare := map[int]float64{}
	for rows.Next() {
		var seed int
		var meanPoints, meanBidShare float64
		if err := rows.Scan(&seed, &meanPoints, &meanBidShare); err != nil {
			return nil, nil, err
		}
		seedPoints[seed] = meanPoints
		seedBidShare[seed] = meanBidShare
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	return seedPoints, seedBidShare, nil
}

func resolveSingleCalcuttaIDForYear(ctx context.Context, db *sql.DB, year int) (string, error) {
	query := `
		SELECT c.id, c.name
		FROM calcuttas c
		JOIN tournaments t ON t.id = c.tournament_id AND t.deleted_at IS NULL
		WHERE c.deleted_at IS NULL
			AND COALESCE(substring(t.name from '([0-9]{4})')::int, 0) = $1
		ORDER BY c.created_at ASC, c.id ASC
	`

	rows, err := db.QueryContext(ctx, query, year)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	type calcuttaRow struct {
		ID   string
		Name string
	}

	calcuttas := make([]calcuttaRow, 0)
	for rows.Next() {
		var r calcuttaRow
		if err := rows.Scan(&r.ID, &r.Name); err != nil {
			return "", err
		}
		calcuttas = append(calcuttas, r)
	}
	if err := rows.Err(); err != nil {
		return "", err
	}

	if len(calcuttas) == 0 {
		return "", fmt.Errorf("%w: %d", ErrNoCalcuttaForYear, year)
	}
	if len(calcuttas) > 1 {
		msg := fmt.Sprintf("found %d calcuttas for year %d; please re-run with -calcutta-id. Candidates:\n", len(calcuttas), year)
		for _, c := range calcuttas {
			msg += fmt.Sprintf("- %s (%s)\n", c.ID, c.Name)
		}
		return "", errors.New(msg)
	}

	return calcuttas[0].ID, nil
}

func queryTeamDataset(ctx context.Context, db *sql.DB, year int, calcuttaID string, excludeEntryName string) ([]TeamDatasetRow, error) {
	if year != 0 && calcuttaID != "" {
		return nil, errors.New("cannot provide both year and calcuttaID")
	}
	if year == 0 && calcuttaID == "" {
		return nil, errors.New("must provide either year or calcuttaID")
	}

	var calcuttaParam any
	if calcuttaID != "" {
		calcuttaParam = calcuttaID
	} else {
		calcuttaParam = nil
	}
	query := `
		WITH team_bids AS (
			SELECT
				t.name as tournament_name,
				COALESCE(substring(t.name from '([0-9]{4})')::int, 0) as tournament_year,
				c.id as calcutta_id,
				tt.id as team_id,
				s.name as school_name,
				tt.seed,
				tt.region,
				tt.wins,
				tt.byes,
				CASE (tt.wins + tt.byes)
					WHEN 0 THEN 0
					WHEN 1 THEN 0
					WHEN 2 THEN 50
					WHEN 3 THEN 150
					WHEN 4 THEN 300
					WHEN 5 THEN 500
					WHEN 6 THEN 750
					WHEN 7 THEN 1050
					ELSE 0
				END::float as team_points,
				COALESCE(SUM(COALESCE(cet.bid, 0)), 0)::float as total_bid,
				COALESCE(SUM(
					CASE
						WHEN $3 <> '' AND ce.name = $3 THEN 0
						ELSE COALESCE(cet.bid, 0)
					END
				), 0)::float as total_bid_excl
			FROM tournament_teams tt
			JOIN schools s ON s.id = tt.school_id AND s.deleted_at IS NULL
			JOIN calcuttas c ON c.tournament_id = tt.tournament_id AND c.deleted_at IS NULL
			JOIN tournaments t ON t.id = c.tournament_id AND t.deleted_at IS NULL
			LEFT JOIN calcutta_entries ce ON ce.calcutta_id = c.id AND ce.deleted_at IS NULL
			LEFT JOIN calcutta_entry_teams cet ON cet.entry_id = ce.id AND cet.team_id = tt.id AND cet.deleted_at IS NULL
			WHERE tt.deleted_at IS NULL
				AND (
					(c.id = $1::uuid)
					OR
					($2 <> 0 AND COALESCE(substring(t.name from '([0-9]{4})')::int, 0) = $2)
				)
			GROUP BY
				t.name,
				tournament_year,
				c.id,
				tt.id,
				s.name,
				tt.seed,
				tt.region,
				tt.wins,
				tt.byes,
				team_points
		)
		SELECT
			tournament_name,
			tournament_year,
			calcutta_id,
			team_id,
			school_name,
			seed,
			region,
			wins,
			byes,
			team_points,
			total_bid,
			SUM(total_bid) OVER (PARTITION BY calcutta_id) as calcutta_total_bid,
			CASE
				WHEN SUM(total_bid) OVER (PARTITION BY calcutta_id) > 0 THEN (total_bid / SUM(total_bid) OVER (PARTITION BY calcutta_id))
				ELSE 0
			END as normalized_bid,
			total_bid_excl,
			SUM(total_bid_excl) OVER (PARTITION BY calcutta_id) as calcutta_total_bid_excl,
			CASE
				WHEN SUM(total_bid_excl) OVER (PARTITION BY calcutta_id) > 0 THEN (total_bid_excl / SUM(total_bid_excl) OVER (PARTITION BY calcutta_id))
				ELSE 0
			END as normalized_bid_excl
		FROM team_bids
		ORDER BY seed ASC, total_bid DESC, school_name ASC
	`

	dbRows, err := db.QueryContext(ctx, query, calcuttaParam, year, excludeEntryName)
	if err != nil {
		return nil, err
	}
	defer dbRows.Close()

	results := make([]TeamDatasetRow, 0)
	for dbRows.Next() {
		var r TeamDatasetRow
		if err := dbRows.Scan(
			&r.TournamentName,
			&r.TournamentYear,
			&r.CalcuttaID,
			&r.TeamID,
			&r.SchoolName,
			&r.Seed,
			&r.Region,
			&r.Wins,
			&r.Byes,
			&r.TeamPoints,
			&r.TotalCommunityBid,
			&r.CalcuttaTotalCommunity,
			&r.NormalizedBid,
			&r.TotalCommunityBidExcl,
			&r.CalcuttaTotalExcl,
			&r.NormalizedBidExcl,
		); err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	if err := dbRows.Err(); err != nil {
		return nil, err
	}

	if len(results) == 0 {
		if year != 0 {
			return nil, fmt.Errorf("no rows returned for year %d", year)
		}
		return nil, fmt.Errorf("no rows returned for calcutta-id %s", calcuttaID)
	}

	return results, nil
}

func queryEntryBids(ctx context.Context, db *sql.DB, calcuttaID string, excludeEntryName string) ([]entryBid, error) {
	query := `
		SELECT
			ce.name,
			cet.team_id,
			SUM(cet.bid)::float as bid
		FROM calcutta_entries ce
		JOIN calcutta_entry_teams cet ON cet.entry_id = ce.id AND cet.deleted_at IS NULL
		WHERE ce.deleted_at IS NULL
			AND ce.calcutta_id = $1::uuid
			AND ($2 = '' OR ce.name <> $2)
		GROUP BY ce.name, cet.team_id
		ORDER BY ce.name ASC
	`

	rows, err := db.QueryContext(ctx, query, calcuttaID, excludeEntryName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]entryBid, 0)
	for rows.Next() {
		var r entryBid
		if err := rows.Scan(&r.EntryName, &r.TeamID, &r.Bid); err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}
