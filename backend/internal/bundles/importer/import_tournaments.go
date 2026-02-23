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

func importTournaments(ctx context.Context, tx pgx.Tx, inDir string) (int, int, []string, error) {
	paths, err := filepath.Glob(filepath.Join(inDir, "tournaments", "*.json"))
	if err != nil {
		return 0, 0, nil, err
	}
	sort.Strings(paths)

	var tournamentIDs []string
	teamsInserted := 0
	for _, path := range paths {
		var b bundles.TournamentBundle
		if err := bundles.ReadJSON(path, &b); err != nil {
			return 0, 0, nil, err
		}

		// Ensure competition exists
		var competitionID string
		err := tx.QueryRow(ctx, `
			INSERT INTO core.competitions (name)
			VALUES ('NCAA Tournament')
			ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
			RETURNING id
		`).Scan(&competitionID)
		if err != nil {
			return 0, 0, nil, fmt.Errorf("failed to upsert competition: %w", err)
		}

		// Extract year from tournament name and ensure season exists
		year := 0
		fmt.Sscanf(b.Tournament.ImportKey, "ncaa-tournament-%d", &year)
		if year == 0 {
			// Try extracting from name
			fmt.Sscanf(b.Tournament.Name, "NCAA Tournament %d", &year)
		}
		var seasonID string
		err = tx.QueryRow(ctx, `
			INSERT INTO core.seasons (year)
			VALUES ($1)
			ON CONFLICT (year) DO UPDATE SET year = EXCLUDED.year
			RETURNING id
		`, year).Scan(&seasonID)
		if err != nil {
			return 0, 0, nil, fmt.Errorf("failed to upsert season for year %d: %w", year, err)
		}

		// Check if tournament exists
		var tournamentID string
		err = tx.QueryRow(ctx, `
			SELECT id FROM core.tournaments WHERE import_key = $1 AND deleted_at IS NULL
		`, b.Tournament.ImportKey).Scan(&tournamentID)
		if err != nil {
			// Tournament doesn't exist, insert it
			err = tx.QueryRow(ctx, `
				INSERT INTO core.tournaments (
					id, competition_id, season_id, import_key, rounds, starting_at,
					final_four_top_left, final_four_bottom_left, final_four_top_right, final_four_bottom_right
				)
				VALUES ($1::uuid, $2, $3, $4, $5, $6, NULLIF($7, ''), NULLIF($8, ''), NULLIF($9, ''), NULLIF($10, ''))
				RETURNING id
			`, uuid.New().String(), competitionID, seasonID, b.Tournament.ImportKey,
				b.Tournament.Rounds, b.Tournament.StartingAt, b.Tournament.FinalFourTopLeft, b.Tournament.FinalFourBottomLeft,
				b.Tournament.FinalFourTopRight, b.Tournament.FinalFourBottomRight).Scan(&tournamentID)
			if err != nil {
				return 0, 0, nil, fmt.Errorf("failed to insert tournament %s: %w", b.Tournament.ImportKey, err)
			}
		} else {
			// Tournament exists, update it
			_, err = tx.Exec(ctx, `
				UPDATE core.tournaments SET
					rounds = $2, starting_at = $3,
					final_four_top_left = NULLIF($4, ''), final_four_bottom_left = NULLIF($5, ''),
					final_four_top_right = NULLIF($6, ''), final_four_bottom_right = NULLIF($7, ''),
					updated_at = NOW(), deleted_at = NULL
				WHERE id = $1
			`, tournamentID, b.Tournament.Rounds, b.Tournament.StartingAt,
				b.Tournament.FinalFourTopLeft, b.Tournament.FinalFourBottomLeft,
				b.Tournament.FinalFourTopRight, b.Tournament.FinalFourBottomRight)
			if err != nil {
				return 0, 0, nil, fmt.Errorf("failed to update tournament %s: %w", b.Tournament.ImportKey, err)
			}
		}

		tournamentIDs = append(tournamentIDs, tournamentID)

		deriveIsEliminated(b.Teams)

		for _, team := range b.Teams {
			var schoolID string
			err := tx.QueryRow(ctx, `
				SELECT id
				FROM core.schools
				WHERE slug = $1 AND deleted_at IS NULL
			`, team.SchoolSlug).Scan(&schoolID)
			if err != nil {
				return 0, 0, nil, fmt.Errorf("school slug %s not found: %w", team.SchoolSlug, err)
			}

			var tournamentTeamID string
			err = tx.QueryRow(ctx, `
				INSERT INTO core.teams (id, tournament_id, school_id, seed, region, byes, wins, is_eliminated)
				VALUES ($1::uuid, $2, $3, $4, $5, $6, $7, $8)
				ON CONFLICT (id)
				DO UPDATE SET
					tournament_id = EXCLUDED.tournament_id,
					school_id = EXCLUDED.school_id,
					seed = EXCLUDED.seed,
					region = EXCLUDED.region,
					byes = EXCLUDED.byes,
					wins = EXCLUDED.wins,
					is_eliminated = EXCLUDED.is_eliminated,
					updated_at = NOW(),
					deleted_at = NULL
				RETURNING id
			`, uuid.New().String(), tournamentID, schoolID, team.Seed, team.Region, team.Byes, team.Wins, team.IsEliminated).Scan(&tournamentTeamID)
			if err != nil {
				return 0, 0, nil, err
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
					return 0, 0, nil, err
				}
			}
			teamsInserted++
		}
	}

	return len(paths), teamsInserted, tournamentIDs, nil
}

// deriveIsEliminated sets IsEliminated on each team based on single-elimination
// logic: if any team has reached progress M (wins+byes), all teams with progress
// less than M have been eliminated.
func deriveIsEliminated(teams []bundles.TeamRecord) {
	maxProgress := 0
	for _, t := range teams {
		if p := t.Wins + t.Byes; p > maxProgress {
			maxProgress = p
		}
	}
	if maxProgress == 0 {
		return // tournament hasn't started
	}
	for i := range teams {
		if teams[i].Wins+teams[i].Byes < maxProgress {
			teams[i].IsEliminated = true
		}
	}
}
