package importer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/bundles"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Report struct {
	StartedAt       time.Time `json:"started_at"`
	FinishedAt      time.Time `json:"finished_at"`
	DryRun          bool      `json:"dry_run"`
	Schools         int       `json:"schools"`
	Tournaments     int       `json:"tournaments"`
	TournamentTeams int       `json:"tournament_teams"`
	Calcuttas       int       `json:"calcuttas"`
	Entries         int       `json:"entries"`
	Bids            int       `json:"bids"`
	Payouts         int       `json:"payouts"`
	Rounds          int       `json:"rounds"`
}

type Options struct {
	DryRun bool
}

func ImportFromDir(ctx context.Context, pool *pgxpool.Pool, inDir string, opts Options) (Report, error) {
	report := Report{StartedAt: time.Now().UTC(), DryRun: opts.DryRun}

	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return report, err
	}
	committed := false
	defer func() {
		if committed {
			return
		}
		_ = tx.Rollback(ctx)
	}()

	counts, err := importAll(ctx, tx, inDir)
	if err != nil {
		return report, err
	}
	report.Schools = counts.schools
	report.Tournaments = counts.tournaments
	report.TournamentTeams = counts.teams
	report.Calcuttas = counts.calcuttas
	report.Entries = counts.entries
	report.Bids = counts.bids
	report.Payouts = counts.payouts
	report.Rounds = counts.rounds

	if opts.DryRun {
		report.FinishedAt = time.Now().UTC()
		return report, nil
	}
	if err := tx.Commit(ctx); err != nil {
		return report, err
	}
	committed = true
	report.FinishedAt = time.Now().UTC()
	return report, nil
}

type importCounts struct {
	schools     int
	tournaments int
	teams       int
	calcuttas   int
	entries     int
	bids        int
	payouts     int
	rounds      int
}

func importAll(ctx context.Context, tx pgx.Tx, inDir string) (importCounts, error) {
	var c importCounts

	sc, err := importSchools(ctx, tx, inDir)
	if err != nil {
		return c, err
	}
	c.schools = sc

	tc, teams, err := importTournaments(ctx, tx, inDir)
	if err != nil {
		return c, err
	}
	c.tournaments = tc
	c.teams = teams

	cc, entries, bids, payouts, rounds, err := importCalcuttas(ctx, tx, inDir)
	if err != nil {
		return c, err
	}
	c.calcuttas = cc
	c.entries = entries
	c.bids = bids
	c.payouts = payouts
	c.rounds = rounds

	return c, nil
}

func readJSON(path string, v any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}

func importSchools(ctx context.Context, tx pgx.Tx, inDir string) (int, error) {
	path := filepath.Join(inDir, "schools.json")
	var b bundles.SchoolsBundle
	if err := readJSON(path, &b); err != nil {
		return 0, err
	}
	for _, s := range b.Schools {
		_, err := tx.Exec(ctx, `
			INSERT INTO schools (slug, name)
			VALUES ($1, $2)
			ON CONFLICT (slug) WHERE deleted_at IS NULL
			DO UPDATE SET name = EXCLUDED.name, updated_at = NOW()
		`, s.Slug, s.Name)
		if err != nil {
			return 0, err
		}
	}
	return len(b.Schools), nil
}

func importTournaments(ctx context.Context, tx pgx.Tx, inDir string) (int, int, error) {
	paths, err := filepath.Glob(filepath.Join(inDir, "tournaments", "*.json"))
	if err != nil {
		return 0, 0, err
	}
	sort.Strings(paths)

	teamsInserted := 0
	for _, path := range paths {
		var b bundles.TournamentBundle
		if err := readJSON(path, &b); err != nil {
			return 0, 0, err
		}

		var tournamentID string
		err := tx.QueryRow(ctx, `
			INSERT INTO tournaments (import_key, name, rounds, final_four_top_left, final_four_bottom_left, final_four_top_right, final_four_bottom_right)
			VALUES ($1, $2, $3, NULLIF($4, ''), NULLIF($5, ''), NULLIF($6, ''), NULLIF($7, ''))
			ON CONFLICT (import_key) WHERE deleted_at IS NULL
			DO UPDATE SET
				name = EXCLUDED.name,
				rounds = EXCLUDED.rounds,
				final_four_top_left = EXCLUDED.final_four_top_left,
				final_four_bottom_left = EXCLUDED.final_four_bottom_left,
				final_four_top_right = EXCLUDED.final_four_top_right,
				final_four_bottom_right = EXCLUDED.final_four_bottom_right,
				updated_at = NOW()
			RETURNING id
		`, b.Tournament.ImportKey, b.Tournament.Name, b.Tournament.Rounds, b.Tournament.FinalFourTopLeft, b.Tournament.FinalFourBottomLeft, b.Tournament.FinalFourTopRight, b.Tournament.FinalFourBottomRight).Scan(&tournamentID)
		if err != nil {
			return 0, 0, err
		}

		for _, team := range b.Teams {
			var schoolID string
			err := tx.QueryRow(ctx, `
				SELECT id
				FROM schools
				WHERE slug = $1 AND deleted_at IS NULL
			`, team.SchoolSlug).Scan(&schoolID)
			if err != nil {
				return 0, 0, fmt.Errorf("school slug %s not found: %w", team.SchoolSlug, err)
			}

			var tournamentTeamID string
			err = tx.QueryRow(ctx, `
				INSERT INTO tournament_teams (tournament_id, school_id, seed, region, byes, wins, eliminated)
				VALUES ($1, $2, $3, $4, $5, $6, $7)
				ON CONFLICT (tournament_id, school_id) WHERE deleted_at IS NULL
				DO UPDATE SET
					seed = EXCLUDED.seed,
					region = EXCLUDED.region,
					byes = EXCLUDED.byes,
					wins = EXCLUDED.wins,
					eliminated = EXCLUDED.eliminated,
					updated_at = NOW()
				RETURNING id
			`, tournamentID, schoolID, team.Seed, team.Region, team.Byes, team.Wins, team.Eliminated).Scan(&tournamentTeamID)
			if err != nil {
				return 0, 0, err
			}

			if team.KenPom != nil {
				_, err := tx.Exec(ctx, `
					INSERT INTO tournament_team_kenpom_stats (tournament_team_id, net_rtg, o_rtg, d_rtg, adj_t)
					VALUES ($1, $2, $3, $4, $5)
					ON CONFLICT (tournament_team_id)
					DO UPDATE SET
						net_rtg = EXCLUDED.net_rtg,
						o_rtg = EXCLUDED.o_rtg,
						d_rtg = EXCLUDED.d_rtg,
						adj_t = EXCLUDED.adj_t,
						updated_at = NOW()
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

func importCalcuttas(ctx context.Context, tx pgx.Tx, inDir string) (int, int, int, int, int, error) {
	root := filepath.Join(inDir, "calcuttas")

	var paths []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".json") {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		if os.IsNotExist(err) {
			return 0, 0, 0, 0, 0, nil
		}
		return 0, 0, 0, 0, 0, err
	}
	sort.Strings(paths)

	entryCount := 0
	bidCount := 0
	payoutCount := 0
	roundCount := 0

	for _, path := range paths {
		var b bundles.CalcuttaBundle
		if err := readJSON(path, &b); err != nil {
			return 0, 0, 0, 0, 0, err
		}

		var tournamentID string
		err := tx.QueryRow(ctx, `
			SELECT id
			FROM tournaments
			WHERE import_key = $1 AND deleted_at IS NULL
		`, b.Tournament.ImportKey).Scan(&tournamentID)
		if err != nil {
			return 0, 0, 0, 0, 0, fmt.Errorf("tournament import_key %s not found: %w", b.Tournament.ImportKey, err)
		}

		ownerID, err := ensureUser(ctx, tx, b.Calcutta.Owner, b.Calcutta.Key)
		if err != nil {
			return 0, 0, 0, 0, 0, err
		}

		var calcuttaID string
		err = tx.QueryRow(ctx, `
			INSERT INTO calcuttas (tournament_id, owner_id, name, key)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (tournament_id, key) WHERE deleted_at IS NULL AND key IS NOT NULL
			DO UPDATE SET
				owner_id = EXCLUDED.owner_id,
				name = EXCLUDED.name,
				updated_at = NOW()
			RETURNING id
		`, tournamentID, ownerID, b.Calcutta.Name, b.Calcutta.Key).Scan(&calcuttaID)
		if err != nil {
			return 0, 0, 0, 0, 0, err
		}

		for _, r := range b.Rounds {
			_, err := tx.Exec(ctx, `
				INSERT INTO calcutta_rounds (calcutta_id, round, points)
				VALUES ($1, $2, $3)
				ON CONFLICT (calcutta_id, round) WHERE deleted_at IS NULL
				DO UPDATE SET points = EXCLUDED.points, updated_at = NOW()
			`, calcuttaID, r.Round, r.Points)
			if err != nil {
				return 0, 0, 0, 0, 0, err
			}
			roundCount++
		}

		for _, p := range b.Payouts {
			_, err := tx.Exec(ctx, `
				INSERT INTO calcutta_payouts (calcutta_id, position, amount_cents)
				VALUES ($1, $2, $3)
				ON CONFLICT (calcutta_id, position)
				DO UPDATE SET amount_cents = EXCLUDED.amount_cents, updated_at = NOW()
			`, calcuttaID, p.Position, p.AmountCents)
			if err != nil {
				return 0, 0, 0, 0, 0, err
			}
			payoutCount++
		}

		entryIDByKey := make(map[string]string, len(b.Entries))
		for _, e := range b.Entries {
			var entryUserID *string
			if e.UserEmail != nil {
				uid, err := ensureUserByEmail(ctx, tx, *e.UserEmail, e.UserName)
				if err != nil {
					return 0, 0, 0, 0, 0, err
				}
				entryUserID = &uid
			}

			var entryID string
			err := tx.QueryRow(ctx, `
				INSERT INTO calcutta_entries (calcutta_id, name, user_id, key)
				VALUES ($1, $2, $3, $4)
				ON CONFLICT (calcutta_id, key) WHERE deleted_at IS NULL AND key IS NOT NULL
				DO UPDATE SET name = EXCLUDED.name, user_id = EXCLUDED.user_id, updated_at = NOW()
				RETURNING id
			`, calcuttaID, e.Name, entryUserID, e.Key).Scan(&entryID)
			if err != nil {
				return 0, 0, 0, 0, 0, err
			}
			entryIDByKey[e.Key] = entryID
			entryCount++
		}

		for _, bid := range b.Bids {
			entryID := entryIDByKey[bid.EntryKey]
			if entryID == "" {
				return 0, 0, 0, 0, 0, fmt.Errorf("bid references unknown entry_key %s", bid.EntryKey)
			}

			var teamID string
			err := tx.QueryRow(ctx, `
				SELECT tt.id
				FROM tournament_teams tt
				JOIN schools s ON s.id = tt.school_id
				WHERE tt.tournament_id = $1 AND s.slug = $2 AND tt.deleted_at IS NULL AND s.deleted_at IS NULL
			`, tournamentID, bid.SchoolSlug).Scan(&teamID)
			if err != nil {
				return 0, 0, 0, 0, 0, fmt.Errorf("tournament team not found for tournament %s school %s: %w", b.Tournament.ImportKey, bid.SchoolSlug, err)
			}

			_, err = tx.Exec(ctx, `
				INSERT INTO calcutta_entry_teams (entry_id, team_id, bid)
				VALUES ($1, $2, $3)
				ON CONFLICT (entry_id, team_id) WHERE deleted_at IS NULL
				DO UPDATE SET bid = EXCLUDED.bid, updated_at = NOW()
			`, entryID, teamID, bid.Bid)
			if err != nil {
				return 0, 0, 0, 0, 0, err
			}
			bidCount++
		}
	}

	return len(paths), entryCount, bidCount, payoutCount, roundCount, nil
}

func ensureUser(ctx context.Context, tx pgx.Tx, u *bundles.UserRef, fallbackKey string) (string, error) {
	email := ""
	first := ""
	last := ""
	if u != nil {
		if u.Email != nil {
			email = strings.TrimSpace(*u.Email)
		}
		if u.FirstName != nil {
			first = strings.TrimSpace(*u.FirstName)
		}
		if u.LastName != nil {
			last = strings.TrimSpace(*u.LastName)
		}
	}
	if email == "" {
		email = fmt.Sprintf("owner+%s@local", strings.ReplaceAll(fallbackKey, " ", "-"))
	}
	if first == "" {
		first = "Unknown"
	}
	if last == "" {
		last = ""
	}
	full := strings.TrimSpace(first + " " + last)
	var fullName *string
	if full != "" {
		fullName = &full
	}
	return ensureUserByEmail(ctx, tx, email, fullName)
}

func ensureUserByEmail(ctx context.Context, tx pgx.Tx, email string, fullName *string) (string, error) {
	first := ""
	last := ""
	if fullName != nil {
		parts := strings.Fields(*fullName)
		if len(parts) > 0 {
			first = parts[0]
		}
		if len(parts) > 1 {
			last = strings.Join(parts[1:], " ")
		}
	}
	if first == "" {
		first = "Unknown"
	}
	if last == "" {
		last = ""
	}

	var id string
	err := tx.QueryRow(ctx, `
		INSERT INTO users (email, first_name, last_name)
		VALUES ($1, $2, $3)
		ON CONFLICT (email)
		DO UPDATE SET first_name = EXCLUDED.first_name, last_name = EXCLUDED.last_name, updated_at = NOW()
		RETURNING id
	`, email, first, last).Scan(&id)
	return id, err
}
