package httpserver

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
)

func (s *Server) handleGetGameOutcomesAlgorithmCoverage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	rows, err := s.pool.Query(ctx, `
		WITH total AS (
			SELECT COUNT(*)::int AS total
			FROM core.tournaments
			WHERE deleted_at IS NULL
		), covered AS (
			SELECT algorithm_id, COUNT(DISTINCT tournament_id)::int AS covered
			FROM derived.game_outcome_runs
			WHERE deleted_at IS NULL
			GROUP BY algorithm_id
		)
		SELECT
			a.id::text,
			a.name,
			a.description,
			COALESCE(c.covered, 0)::int AS covered,
			t.total::int AS total
		FROM derived.algorithms a
		CROSS JOIN total t
		LEFT JOIN covered c ON c.algorithm_id = a.id
		WHERE a.kind = 'game_outcomes'
			AND a.deleted_at IS NULL
		ORDER BY a.created_at DESC
	`)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, name string
		var desc *string
		var covered, total int
		if err := rows.Scan(&id, &name, &desc, &covered, &total); err != nil {
			writeErrorFromErr(w, r, err)
			return
		}
		items = append(items, map[string]interface{}{
			"id":          id,
			"name":        name,
			"description": desc,
			"covered":     covered,
			"total":       total,
		})
	}
	if err := rows.Err(); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"items": items,
		"count": len(items),
	})
}

func (s *Server) handleGetMarketShareAlgorithmCoverage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	rows, err := s.pool.Query(ctx, `
		WITH total AS (
			SELECT COUNT(*)::int AS total
			FROM core.calcuttas
			WHERE deleted_at IS NULL
		), covered AS (
			SELECT algorithm_id, COUNT(DISTINCT calcutta_id)::int AS covered
			FROM derived.market_share_runs
			WHERE deleted_at IS NULL
			GROUP BY algorithm_id
		)
		SELECT
			a.id::text,
			a.name,
			a.description,
			COALESCE(c.covered, 0)::int AS covered,
			t.total::int AS total
		FROM derived.algorithms a
		CROSS JOIN total t
		LEFT JOIN covered c ON c.algorithm_id = a.id
		WHERE a.kind = 'market_share'
			AND a.deleted_at IS NULL
		ORDER BY a.created_at DESC
	`)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, name string
		var desc *string
		var covered, total int
		if err := rows.Scan(&id, &name, &desc, &covered, &total); err != nil {
			writeErrorFromErr(w, r, err)
			return
		}
		items = append(items, map[string]interface{}{
			"id":          id,
			"name":        name,
			"description": desc,
			"covered":     covered,
			"total":       total,
		})
	}
	if err := rows.Err(); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"items": items,
		"count": len(items),
	})
}

func (s *Server) handleGetGameOutcomesAlgorithmCoverageDetail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	algorithmID := vars["id"]
	if algorithmID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Missing algorithm ID", "id")
		return
	}

	var algID, algName string
	var algDesc *string
	err := s.pool.QueryRow(ctx, `
		SELECT id::text, name, description
		FROM derived.algorithms
		WHERE id = $1::uuid
			AND kind = 'game_outcomes'
			AND deleted_at IS NULL
		LIMIT 1
	`, algorithmID).Scan(&algID, &algName, &algDesc)
	if err != nil {
		if err == pgx.ErrNoRows {
			writeError(w, r, http.StatusNotFound, "not_found", "Algorithm not found", "id")
			return
		}
		writeErrorFromErr(w, r, err)
		return
	}

	rows, err := s.pool.Query(ctx, `
		SELECT
			t.id::text,
			t.name,
			t.starting_at,
			MAX(r.created_at) AS last_run_at
		FROM core.tournaments t
		LEFT JOIN derived.game_outcome_runs r
			ON r.tournament_id = t.id
			AND r.algorithm_id = $1::uuid
			AND r.deleted_at IS NULL
		WHERE t.deleted_at IS NULL
		GROUP BY t.id, t.name, t.starting_at
		ORDER BY t.starting_at DESC NULLS LAST, t.name DESC
	`, algorithmID)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	defer rows.Close()

	tournaments := make([]map[string]interface{}, 0)
	covered := 0
	total := 0
	for rows.Next() {
		total++
		var tid, name string
		var startingAt *time.Time
		var lastRunAt *time.Time
		if err := rows.Scan(&tid, &name, &startingAt, &lastRunAt); err != nil {
			writeErrorFromErr(w, r, err)
			return
		}
		if lastRunAt != nil {
			covered++
		}
		tournaments = append(tournaments, map[string]interface{}{
			"tournament_id":   tid,
			"tournament_name": name,
			"starting_at":     startingAt,
			"last_run_at":     lastRunAt,
		})
	}
	if err := rows.Err(); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"algorithm": map[string]interface{}{
			"id":          algID,
			"name":        algName,
			"description": algDesc,
		},
		"covered": covered,
		"total":   total,
		"items":   tournaments,
		"count":   len(tournaments),
	})
}

func (s *Server) handleGetMarketShareAlgorithmCoverageDetail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	algorithmID := vars["id"]
	if algorithmID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Missing algorithm ID", "id")
		return
	}

	var algID, algName string
	var algDesc *string
	err := s.pool.QueryRow(ctx, `
		SELECT id::text, name, description
		FROM derived.algorithms
		WHERE id = $1::uuid
			AND kind = 'market_share'
			AND deleted_at IS NULL
		LIMIT 1
	`, algorithmID).Scan(&algID, &algName, &algDesc)
	if err != nil {
		if err == pgx.ErrNoRows {
			writeError(w, r, http.StatusNotFound, "not_found", "Algorithm not found", "id")
			return
		}
		writeErrorFromErr(w, r, err)
		return
	}

	rows, err := s.pool.Query(ctx, `
		SELECT
			c.id::text,
			c.name,
			c.tournament_id::text,
			t.name,
			t.starting_at,
			MAX(r.created_at) AS last_run_at
		FROM core.calcuttas c
		JOIN core.tournaments t
			ON t.id = c.tournament_id
			AND t.deleted_at IS NULL
		LEFT JOIN derived.market_share_runs r
			ON r.calcutta_id = c.id
			AND r.algorithm_id = $1::uuid
			AND r.deleted_at IS NULL
		WHERE c.deleted_at IS NULL
		GROUP BY c.id, c.name, c.tournament_id, t.name, t.starting_at
		ORDER BY t.starting_at DESC NULLS LAST, c.created_at DESC
	`, algorithmID)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	covered := 0
	total := 0
	for rows.Next() {
		total++
		var calcuttaID, calcuttaName, tournamentID, tournamentName string
		var startingAt *time.Time
		var lastRunAt *time.Time
		if err := rows.Scan(&calcuttaID, &calcuttaName, &tournamentID, &tournamentName, &startingAt, &lastRunAt); err != nil {
			writeErrorFromErr(w, r, err)
			return
		}
		if lastRunAt != nil {
			covered++
		}
		items = append(items, map[string]interface{}{
			"calcutta_id":     calcuttaID,
			"calcutta_name":   calcuttaName,
			"tournament_id":   tournamentID,
			"tournament_name": tournamentName,
			"starting_at":     startingAt,
			"last_run_at":     lastRunAt,
		})
	}
	if err := rows.Err(); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"algorithm": map[string]interface{}{
			"id":          algID,
			"name":        algName,
			"description": algDesc,
		},
		"covered": covered,
		"total":   total,
		"items":   items,
		"count":   len(items),
	})
}
