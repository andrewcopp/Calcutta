package importer

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"

	"github.com/andrewcopp/Calcutta/backend/internal/bundles"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func importTournaments(ctx context.Context, tx pgx.Tx, inDir string) (int, int, error) {
	paths, err := filepath.Glob(filepath.Join(inDir, "tournaments", "*.json"))
	if err != nil {
		return 0, 0, err
	}
	sort.Strings(paths)

	teamsInserted := 0
	for _, path := range paths {
		var b bundles.TournamentBundle
		if err := bundles.ReadJSON(path, &b); err != nil {
			return 0, 0, err
		}

		var tournamentID string
		err := tx.QueryRow(ctx, `
			WITH year_ctx AS (
				SELECT CAST(SUBSTRING($2 FROM '[0-9]{4}') AS INTEGER) AS year
			),
			competition_ins AS (
				INSERT INTO core.competitions (name)
				VALUES ('NCAA Men''s')
				ON CONFLICT (name) DO NOTHING
				RETURNING id
			),
			competition AS (
				SELECT id FROM competition_ins
				UNION ALL
				SELECT id FROM core.competitions WHERE name = 'NCAA Men''s'
				LIMIT 1
			),
			season_ins AS (
				INSERT INTO core.seasons (year)
				SELECT year FROM year_ctx
				ON CONFLICT (year) DO NOTHING
				RETURNING id
			),
			season AS (
				SELECT id FROM season_ins
				UNION ALL
				SELECT id FROM core.seasons WHERE year = (SELECT year FROM year_ctx)
				LIMIT 1
			)
			INSERT INTO core.tournaments (
				id,
				competition_id,
				season_id,
				name,
				import_key,
				rounds,
				final_four_top_left,
				final_four_bottom_left,
				final_four_top_right,
				final_four_bottom_right
			)
			SELECT
				$1::uuid,
				(SELECT id FROM competition),
				(SELECT id FROM season),
				$2,
				$3,
				$4,
				NULLIF($5, ''),
				NULLIF($6, ''),
				NULLIF($7, ''),
				NULLIF($8, '')
			ON CONFLICT (import_key) WHERE deleted_at IS NULL
			DO UPDATE SET
				name = EXCLUDED.name,
				rounds = EXCLUDED.rounds,
				final_four_top_left = EXCLUDED.final_four_top_left,
				final_four_bottom_left = EXCLUDED.final_four_bottom_left,
				final_four_top_right = EXCLUDED.final_four_top_right,
				final_four_bottom_right = EXCLUDED.final_four_bottom_right,
				updated_at = NOW(),
				deleted_at = NULL
			RETURNING id
		`, uuid.New().String(), b.Tournament.Name, b.Tournament.ImportKey, b.Tournament.Rounds, b.Tournament.FinalFourTopLeft, b.Tournament.FinalFourBottomLeft, b.Tournament.FinalFourTopRight, b.Tournament.FinalFourBottomRight).Scan(&tournamentID)
		if err != nil {
			return 0, 0, err
		}

		for _, team := range b.Teams {
			var schoolID string
			err := tx.QueryRow(ctx, `
				SELECT id
				FROM core.schools
				WHERE slug = $1 AND deleted_at IS NULL
			`, team.SchoolSlug).Scan(&schoolID)
			if err != nil {
				return 0, 0, fmt.Errorf("school slug %s not found: %w", team.SchoolSlug, err)
			}

			var tournamentTeamID string
			err = tx.QueryRow(ctx, `
				INSERT INTO core.teams (id, tournament_id, school_id, seed, region, byes, wins, eliminated)
				VALUES ($1::uuid, $2, $3, $4, $5, $6, $7, $8)
				ON CONFLICT (id)
				DO UPDATE SET
					tournament_id = EXCLUDED.tournament_id,
					school_id = EXCLUDED.school_id,
					seed = EXCLUDED.seed,
					region = EXCLUDED.region,
					byes = EXCLUDED.byes,
					wins = EXCLUDED.wins,
					eliminated = EXCLUDED.eliminated,
					updated_at = NOW(),
					deleted_at = NULL
				RETURNING id
			`, uuid.New().String(), tournamentID, schoolID, team.Seed, team.Region, team.Byes, team.Wins, team.Eliminated).Scan(&tournamentTeamID)
			if err != nil {
				return 0, 0, err
			}

			if team.KenPom != nil {
				_, err := tx.Exec(ctx, `
					INSERT INTO core.team_kenpom_stats (team_id, net_rtg, o_rtg, d_rtg, adj_t)
					VALUES ($1, $2, $3, $4, $5)
					ON CONFLICT (team_id)
					DO UPDATE SET
						net_rtg = EXCLUDED.net_rtg,
						o_rtg = EXCLUDED.o_rtg,
						d_rtg = EXCLUDED.d_rtg,
						adj_t = EXCLUDED.adj_t,
						updated_at = NOW(),
						deleted_at = NULL
				`, tournamentTeamID, team.KenPom.NetRTG, team.KenPom.ORTG, team.KenPom.DRTG, team.KenPom.AdjT)
				if err != nil {
					return 0, 0, err
				}
			}
			teamsInserted++
		}
	}

	return len(paths), teamsInserted, nil
}
