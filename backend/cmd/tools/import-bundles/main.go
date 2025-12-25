package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/platform"
	"github.com/jackc/pgx/v5"
)

type SchoolsBundle struct {
	Version     int           `json:"version"`
	GeneratedAt time.Time     `json:"generated_at"`
	Schools     []SchoolEntry `json:"schools"`
}

type SchoolEntry struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

type TournamentBundle struct {
	Version     int              `json:"version"`
	GeneratedAt time.Time        `json:"generated_at"`
	Tournament  TournamentRecord `json:"tournament"`
	Teams       []TeamRecord     `json:"teams"`
}

type TournamentRecord struct {
	ImportKey            string `json:"import_key"`
	Name                 string `json:"name"`
	Rounds               int    `json:"rounds"`
	FinalFourTopLeft     string `json:"final_four_top_left"`
	FinalFourBottomLeft  string `json:"final_four_bottom_left"`
	FinalFourTopRight    string `json:"final_four_top_right"`
	FinalFourBottomRight string `json:"final_four_bottom_right"`
}

type TeamRecord struct {
	SchoolSlug string        `json:"school_slug"`
	SchoolName string        `json:"school_name"`
	Seed       int           `json:"seed"`
	Region     string        `json:"region"`
	Byes       int           `json:"byes"`
	Wins       int           `json:"wins"`
	Eliminated bool          `json:"eliminated"`
	KenPom     *KenPomRecord `json:"kenpom,omitempty"`
}

type KenPomRecord struct {
	NetRTG float64 `json:"net_rtg"`
	ORTG   float64 `json:"o_rtg"`
	DRTG   float64 `json:"d_rtg"`
	AdjT   float64 `json:"adj_t"`
}

type CalcuttaBundle struct {
	Version     int            `json:"version"`
	GeneratedAt time.Time      `json:"generated_at"`
	Tournament  TournamentRef  `json:"tournament"`
	Calcutta    CalcuttaRecord `json:"calcutta"`
	Rounds      []RoundRecord  `json:"rounds"`
	Payouts     []PayoutRecord `json:"payouts"`
	Entries     []EntryRecord  `json:"entries"`
	Bids        []EntryTeamBid `json:"bids"`
}

type TournamentRef struct {
	ImportKey string `json:"import_key"`
	Name      string `json:"name"`
}

type CalcuttaRecord struct {
	Key   string   `json:"key"`
	Name  string   `json:"name"`
	Owner *UserRef `json:"owner,omitempty"`
}

type UserRef struct {
	Email     *string `json:"email,omitempty"`
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
}

type RoundRecord struct {
	Round  int `json:"round"`
	Points int `json:"points"`
}

type PayoutRecord struct {
	Position    int `json:"position"`
	AmountCents int `json:"amount_cents"`
}

type EntryRecord struct {
	Key       string  `json:"key"`
	Name      string  `json:"name"`
	UserName  *string `json:"user_name,omitempty"`
	UserEmail *string `json:"user_email,omitempty"`
}

type EntryTeamBid struct {
	EntryKey   string `json:"entry_key"`
	SchoolSlug string `json:"school_slug"`
	Bid        int    `json:"bid"`
}

func readJSON(path string, v any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}

func main() {
	inDir := flag.String("in", "./exports/bundles", "input bundles directory")
	dryRun := flag.Bool("dry-run", true, "read and validate bundles; rollback DB writes")
	flag.Parse()

	cfg, err := platform.LoadConfigFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	pool, err := platform.OpenPGXPool(ctx, cfg, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		log.Fatal(err)
	}
	committed := false
	defer func() {
		if committed {
			return
		}
		_ = tx.Rollback(ctx)
	}()

	if err := importSchools(ctx, tx, *inDir); err != nil {
		log.Fatal(err)
	}
	if err := importTournaments(ctx, tx, *inDir); err != nil {
		log.Fatal(err)
	}
	if err := importCalcuttas(ctx, tx, *inDir); err != nil {
		log.Fatal(err)
	}

	if *dryRun {
		fmt.Println("dry-run complete (rolled back)")
		return
	}
	if err := tx.Commit(ctx); err != nil {
		log.Fatal(err)
	}
	committed = true
	fmt.Println("import complete")
}

func importSchools(ctx context.Context, tx pgx.Tx, inDir string) error {
	path := filepath.Join(inDir, "schools.json")
	var b SchoolsBundle
	if err := readJSON(path, &b); err != nil {
		return err
	}
	for _, s := range b.Schools {
		_, err := tx.Exec(ctx, `
			INSERT INTO schools (slug, name)
			VALUES ($1, $2)
			ON CONFLICT (slug) WHERE deleted_at IS NULL
			DO UPDATE SET name = EXCLUDED.name, updated_at = NOW()
		`, s.Slug, s.Name)
		if err != nil {
			return err
		}
	}
	return nil
}

func importTournaments(ctx context.Context, tx pgx.Tx, inDir string) error {
	paths, err := filepath.Glob(filepath.Join(inDir, "tournaments", "*.json"))
	if err != nil {
		return err
	}
	sort.Strings(paths)

	for _, path := range paths {
		var b TournamentBundle
		if err := readJSON(path, &b); err != nil {
			return err
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
			return err
		}

		for _, team := range b.Teams {
			var schoolID string
			err := tx.QueryRow(ctx, `
				SELECT id
				FROM schools
				WHERE slug = $1 AND deleted_at IS NULL
			`, team.SchoolSlug).Scan(&schoolID)
			if err != nil {
				return fmt.Errorf("school slug %s not found: %w", team.SchoolSlug, err)
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
				return err
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
					return err
				}
			}
		}
	}

	return nil
}

func importCalcuttas(ctx context.Context, tx pgx.Tx, inDir string) error {
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
			return nil
		}
		return err
	}
	sort.Strings(paths)

	for _, path := range paths {
		var b CalcuttaBundle
		if err := readJSON(path, &b); err != nil {
			return err
		}

		var tournamentID string
		err := tx.QueryRow(ctx, `
			SELECT id
			FROM tournaments
			WHERE import_key = $1 AND deleted_at IS NULL
		`, b.Tournament.ImportKey).Scan(&tournamentID)
		if err != nil {
			return fmt.Errorf("tournament import_key %s not found: %w", b.Tournament.ImportKey, err)
		}

		ownerID, err := ensureUser(ctx, tx, b.Calcutta.Owner, b.Calcutta.Key)
		if err != nil {
			return err
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
			return err
		}

		for _, r := range b.Rounds {
			_, err := tx.Exec(ctx, `
				INSERT INTO calcutta_rounds (calcutta_id, round, points)
				VALUES ($1, $2, $3)
				ON CONFLICT (calcutta_id, round) WHERE deleted_at IS NULL
				DO UPDATE SET points = EXCLUDED.points, updated_at = NOW()
			`, calcuttaID, r.Round, r.Points)
			if err != nil {
				return err
			}
		}

		for _, p := range b.Payouts {
			_, err := tx.Exec(ctx, `
				INSERT INTO calcutta_payouts (calcutta_id, position, amount_cents)
				VALUES ($1, $2, $3)
				ON CONFLICT (calcutta_id, position)
				DO UPDATE SET amount_cents = EXCLUDED.amount_cents, updated_at = NOW()
			`, calcuttaID, p.Position, p.AmountCents)
			if err != nil {
				return err
			}
		}

		entryIDByKey := make(map[string]string, len(b.Entries))
		for _, e := range b.Entries {
			var entryUserID *string
			if e.UserEmail != nil {
				uid, err := ensureUserByEmail(ctx, tx, *e.UserEmail, e.UserName)
				if err != nil {
					return err
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
				return err
			}
			entryIDByKey[e.Key] = entryID
		}

		for _, bid := range b.Bids {
			entryID := entryIDByKey[bid.EntryKey]
			if entryID == "" {
				return fmt.Errorf("bid references unknown entry_key %s", bid.EntryKey)
			}

			var teamID string
			err := tx.QueryRow(ctx, `
				SELECT tt.id
				FROM tournament_teams tt
				JOIN schools s ON s.id = tt.school_id
				WHERE tt.tournament_id = $1 AND s.slug = $2 AND tt.deleted_at IS NULL AND s.deleted_at IS NULL
			`, tournamentID, bid.SchoolSlug).Scan(&teamID)
			if err != nil {
				return fmt.Errorf("tournament team not found for tournament %s school %s: %w", b.Tournament.ImportKey, bid.SchoolSlug, err)
			}

			_, err = tx.Exec(ctx, `
				INSERT INTO calcutta_entry_teams (entry_id, team_id, bid)
				VALUES ($1, $2, $3)
				ON CONFLICT (entry_id, team_id) WHERE deleted_at IS NULL
				DO UPDATE SET bid = EXCLUDED.bid, updated_at = NOW()
			`, entryID, teamID, bid.Bid)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func ensureUser(ctx context.Context, tx pgx.Tx, u *UserRef, fallbackKey string) (string, error) {
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
	return ensureUserByEmail(ctx, tx, email, ptrString(strings.TrimSpace(first+" "+last)))
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

func ptrString(s string) *string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return &s
}
