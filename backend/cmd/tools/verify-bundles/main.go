package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/platform"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
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

type mismatch struct {
	Where string
	What  string
}

func main() {
	inDir := flag.String("in", "./exports/bundles", "input bundles directory")
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

	var mismatches []mismatch

	mismatches = append(mismatches, verifySchools(ctx, pool, *inDir)...)
	mismatches = append(mismatches, verifyTournaments(ctx, pool, *inDir)...)
	mismatches = append(mismatches, verifyCalcuttas(ctx, pool, *inDir)...)

	if len(mismatches) > 0 {
		for i, m := range mismatches {
			if i >= 50 {
				fmt.Printf("... (%d more)\n", len(mismatches)-i)
				break
			}
			fmt.Printf("%s: %s\n", m.Where, m.What)
		}
		log.Fatalf("verify failed: %d mismatches", len(mismatches))
	}

	fmt.Println("verify ok")
}

func verifySchools(ctx context.Context, pool *pgxpool.Pool, inDir string) []mismatch {
	path := filepath.Join(inDir, "schools.json")
	var b SchoolsBundle
	if err := readJSON(path, &b); err != nil {
		return []mismatch{{Where: "schools", What: err.Error()}}
	}

	rows, err := pool.Query(ctx, `SELECT slug, name FROM schools WHERE deleted_at IS NULL`)
	if err != nil {
		return []mismatch{{Where: "schools", What: err.Error()}}
	}
	defer rows.Close()

	inDB := map[string]string{}
	for rows.Next() {
		var slug, name string
		if err := rows.Scan(&slug, &name); err != nil {
			return []mismatch{{Where: "schools", What: err.Error()}}
		}
		inDB[slug] = name
	}
	if err := rows.Err(); err != nil {
		return []mismatch{{Where: "schools", What: err.Error()}}
	}

	var out []mismatch
	for _, s := range b.Schools {
		name, ok := inDB[s.Slug]
		if !ok {
			out = append(out, mismatch{Where: "schools:" + s.Slug, What: "missing in db"})
			continue
		}
		if name != s.Name {
			out = append(out, mismatch{Where: "schools:" + s.Slug, What: fmt.Sprintf("name mismatch db=%q bundle=%q", name, s.Name)})
		}
	}
	return out
}

func verifyTournaments(ctx context.Context, pool *pgxpool.Pool, inDir string) []mismatch {
	paths, err := filepath.Glob(filepath.Join(inDir, "tournaments", "*.json"))
	if err != nil {
		return []mismatch{{Where: "tournaments", What: err.Error()}}
	}
	sort.Strings(paths)

	var out []mismatch
	for _, path := range paths {
		var b TournamentBundle
		if err := readJSON(path, &b); err != nil {
			out = append(out, mismatch{Where: "tournaments", What: err.Error()})
			continue
		}

		var id, name string
		var rounds int
		var f1, f2, f3, f4 *string
		err := pool.QueryRow(ctx, `
			SELECT id, name, rounds, final_four_top_left, final_four_bottom_left, final_four_top_right, final_four_bottom_right
			FROM tournaments
			WHERE import_key = $1 AND deleted_at IS NULL
		`, b.Tournament.ImportKey).Scan(&id, &name, &rounds, &f1, &f2, &f3, &f4)
		if err != nil {
			out = append(out, mismatch{Where: "tournaments:" + b.Tournament.ImportKey, What: "missing in db"})
			continue
		}

		if name != b.Tournament.Name {
			out = append(out, mismatch{Where: "tournaments:" + b.Tournament.ImportKey, What: fmt.Sprintf("name mismatch db=%q bundle=%q", name, b.Tournament.Name)})
		}
		if rounds != b.Tournament.Rounds {
			out = append(out, mismatch{Where: "tournaments:" + b.Tournament.ImportKey, What: fmt.Sprintf("rounds mismatch db=%d bundle=%d", rounds, b.Tournament.Rounds)})
		}

		if deref(f1) != b.Tournament.FinalFourTopLeft {
			out = append(out, mismatch{Where: "tournaments:" + b.Tournament.ImportKey, What: "final_four_top_left mismatch"})
		}
		if deref(f2) != b.Tournament.FinalFourBottomLeft {
			out = append(out, mismatch{Where: "tournaments:" + b.Tournament.ImportKey, What: "final_four_bottom_left mismatch"})
		}
		if deref(f3) != b.Tournament.FinalFourTopRight {
			out = append(out, mismatch{Where: "tournaments:" + b.Tournament.ImportKey, What: "final_four_top_right mismatch"})
		}
		if deref(f4) != b.Tournament.FinalFourBottomRight {
			out = append(out, mismatch{Where: "tournaments:" + b.Tournament.ImportKey, What: "final_four_bottom_right mismatch"})
		}

		out = append(out, verifyTournamentTeams(ctx, pool, id, b.Tournament.ImportKey, b.Teams)...)
	}
	return out
}

func verifyTournamentTeams(ctx context.Context, pool *pgxpool.Pool, tournamentID, tournamentKey string, teams []TeamRecord) []mismatch {
	rows, err := pool.Query(ctx, `
		SELECT
			s.slug,
			s.name,
			tt.seed,
			tt.region,
			tt.byes,
			tt.wins,
			tt.eliminated,
			k.net_rtg,
			k.o_rtg,
			k.d_rtg,
			k.adj_t,
			(k.tournament_team_id IS NOT NULL)
		FROM tournament_teams tt
		JOIN schools s ON s.id = tt.school_id
		LEFT JOIN tournament_team_kenpom_stats k ON k.tournament_team_id = tt.id AND k.deleted_at IS NULL
		WHERE tt.tournament_id = $1 AND tt.deleted_at IS NULL AND s.deleted_at IS NULL
	`, tournamentID)
	if err != nil {
		return []mismatch{{Where: "tournament_teams:" + tournamentKey, What: err.Error()}}
	}
	defer rows.Close()

	tInDB := map[string]TeamRecord{}
	kpInDB := map[string]*KenPomRecord{}
	for rows.Next() {
		var slug, name, region string
		var seed, byes, wins int
		var eliminated bool
		var net, o, d, adj *float64
		var hasKP bool
		if err := rows.Scan(&slug, &name, &seed, &region, &byes, &wins, &eliminated, &net, &o, &d, &adj, &hasKP); err != nil {
			return []mismatch{{Where: "tournament_teams:" + tournamentKey, What: err.Error()}}
		}
		tInDB[slug] = TeamRecord{SchoolSlug: slug, SchoolName: name, Seed: seed, Region: region, Byes: byes, Wins: wins, Eliminated: eliminated}
		if hasKP {
			kpInDB[slug] = &KenPomRecord{NetRTG: derefF(net), ORTG: derefF(o), DRTG: derefF(d), AdjT: derefF(adj)}
		}
	}
	if err := rows.Err(); err != nil {
		return []mismatch{{Where: "tournament_teams:" + tournamentKey, What: err.Error()}}
	}

	var out []mismatch
	for _, t := range teams {
		dbT, ok := tInDB[t.SchoolSlug]
		if !ok {
			out = append(out, mismatch{Where: "tournament_teams:" + tournamentKey + ":" + t.SchoolSlug, What: "missing in db"})
			continue
		}
		if dbT.Seed != t.Seed || dbT.Region != t.Region || dbT.Byes != t.Byes || dbT.Wins != t.Wins || dbT.Eliminated != t.Eliminated {
			out = append(out, mismatch{Where: "tournament_teams:" + tournamentKey + ":" + t.SchoolSlug, What: "field mismatch"})
		}

		dbKP := kpInDB[t.SchoolSlug]
		if t.KenPom == nil {
			if dbKP != nil {
				out = append(out, mismatch{Where: "tournament_teams:" + tournamentKey + ":" + t.SchoolSlug, What: "kenpom present in db but missing in bundle"})
			}
			continue
		}
		if dbKP == nil {
			out = append(out, mismatch{Where: "tournament_teams:" + tournamentKey + ":" + t.SchoolSlug, What: "kenpom missing in db"})
			continue
		}
		if !approxEq(dbKP.NetRTG, t.KenPom.NetRTG) || !approxEq(dbKP.ORTG, t.KenPom.ORTG) || !approxEq(dbKP.DRTG, t.KenPom.DRTG) || !approxEq(dbKP.AdjT, t.KenPom.AdjT) {
			out = append(out, mismatch{Where: "tournament_teams:" + tournamentKey + ":" + t.SchoolSlug, What: "kenpom mismatch"})
		}
	}
	return out
}

func verifyCalcuttas(ctx context.Context, pool *pgxpool.Pool, inDir string) []mismatch {
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
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return []mismatch{{Where: "calcuttas", What: err.Error()}}
	}
	sort.Strings(paths)

	var out []mismatch
	for _, path := range paths {
		var b CalcuttaBundle
		if err := readJSON(path, &b); err != nil {
			out = append(out, mismatch{Where: "calcuttas", What: err.Error()})
			continue
		}

		var tournamentID string
		err := pool.QueryRow(ctx, `SELECT id FROM tournaments WHERE import_key = $1 AND deleted_at IS NULL`, b.Tournament.ImportKey).Scan(&tournamentID)
		if err != nil {
			out = append(out, mismatch{Where: "calcuttas:" + b.Calcutta.Key, What: "tournament missing"})
			continue
		}

		var calcuttaID string
		var calcuttaName string
		var ownerEmail string
		err = pool.QueryRow(ctx, `
			SELECT c.id, c.name, COALESCE(u.email, '')
			FROM calcuttas c
			JOIN users u ON u.id = c.owner_id
			WHERE c.tournament_id = $1 AND c.key = $2 AND c.deleted_at IS NULL AND u.deleted_at IS NULL
		`, tournamentID, b.Calcutta.Key).Scan(&calcuttaID, &calcuttaName, &ownerEmail)
		if err != nil {
			out = append(out, mismatch{Where: "calcuttas:" + b.Calcutta.Key, What: "missing in db"})
			continue
		}

		if calcuttaName != b.Calcutta.Name {
			out = append(out, mismatch{Where: "calcuttas:" + b.Calcutta.Key, What: fmt.Sprintf("name mismatch db=%q bundle=%q", calcuttaName, b.Calcutta.Name)})
		}
		if b.Calcutta.Owner != nil && b.Calcutta.Owner.Email != nil {
			if ownerEmail != *b.Calcutta.Owner.Email {
				out = append(out, mismatch{Where: "calcuttas:" + b.Calcutta.Key, What: fmt.Sprintf("owner email mismatch db=%q bundle=%q", ownerEmail, *b.Calcutta.Owner.Email)})
			}
		}

		out = append(out, verifyCalcuttaRounds(ctx, pool, calcuttaID, b.Calcutta.Key, b.Rounds)...)
		out = append(out, verifyCalcuttaPayouts(ctx, pool, calcuttaID, b.Calcutta.Key, b.Payouts)...)
		out = append(out, verifyCalcuttaEntriesAndBids(ctx, pool, tournamentID, calcuttaID, b.Calcutta.Key, b.Entries, b.Bids)...)
	}

	return out
}

func verifyCalcuttaRounds(ctx context.Context, pool *pgxpool.Pool, calcuttaID, calcuttaKey string, rounds []RoundRecord) []mismatch {
	rows, err := pool.Query(ctx, `SELECT round, points FROM calcutta_rounds WHERE calcutta_id = $1 AND deleted_at IS NULL`, calcuttaID)
	if err != nil {
		return []mismatch{{Where: "calcutta_rounds:" + calcuttaKey, What: err.Error()}}
	}
	defer rows.Close()

	inDB := map[int]int{}
	for rows.Next() {
		var r, p int
		if err := rows.Scan(&r, &p); err != nil {
			return []mismatch{{Where: "calcutta_rounds:" + calcuttaKey, What: err.Error()}}
		}
		inDB[r] = p
	}
	if err := rows.Err(); err != nil {
		return []mismatch{{Where: "calcutta_rounds:" + calcuttaKey, What: err.Error()}}
	}

	var out []mismatch
	for _, rr := range rounds {
		p, ok := inDB[rr.Round]
		if !ok {
			out = append(out, mismatch{Where: "calcutta_rounds:" + calcuttaKey, What: fmt.Sprintf("missing round %d", rr.Round)})
			continue
		}
		if p != rr.Points {
			out = append(out, mismatch{Where: "calcutta_rounds:" + calcuttaKey, What: fmt.Sprintf("round %d points mismatch db=%d bundle=%d", rr.Round, p, rr.Points)})
		}
	}
	return out
}

func verifyCalcuttaPayouts(ctx context.Context, pool *pgxpool.Pool, calcuttaID, calcuttaKey string, payouts []PayoutRecord) []mismatch {
	rows, err := pool.Query(ctx, `SELECT position, amount_cents FROM calcutta_payouts WHERE calcutta_id = $1 AND deleted_at IS NULL`, calcuttaID)
	if err != nil {
		return []mismatch{{Where: "calcutta_payouts:" + calcuttaKey, What: err.Error()}}
	}
	defer rows.Close()

	inDB := map[int]int{}
	for rows.Next() {
		var pos, cents int
		if err := rows.Scan(&pos, &cents); err != nil {
			return []mismatch{{Where: "calcutta_payouts:" + calcuttaKey, What: err.Error()}}
		}
		inDB[pos] = cents
	}
	if err := rows.Err(); err != nil {
		return []mismatch{{Where: "calcutta_payouts:" + calcuttaKey, What: err.Error()}}
	}

	var out []mismatch
	for _, p := range payouts {
		c, ok := inDB[p.Position]
		if !ok {
			out = append(out, mismatch{Where: "calcutta_payouts:" + calcuttaKey, What: fmt.Sprintf("missing payout position %d", p.Position)})
			continue
		}
		if c != p.AmountCents {
			out = append(out, mismatch{Where: "calcutta_payouts:" + calcuttaKey, What: fmt.Sprintf("payout %d cents mismatch db=%d bundle=%d", p.Position, c, p.AmountCents)})
		}
	}
	return out
}

func verifyCalcuttaEntriesAndBids(ctx context.Context, pool *pgxpool.Pool, tournamentID, calcuttaID, calcuttaKey string, entries []EntryRecord, bids []EntryTeamBid) []mismatch {
	rows, err := pool.Query(ctx, `
		SELECT ce.key, ce.name, COALESCE(u.email, '')
		FROM calcutta_entries ce
		LEFT JOIN users u ON u.id = ce.user_id
		WHERE ce.calcutta_id = $1 AND ce.deleted_at IS NULL
	`, calcuttaID)
	if err != nil {
		return []mismatch{{Where: "calcutta_entries:" + calcuttaKey, What: err.Error()}}
	}
	defer rows.Close()

	inDB := map[string]struct {
		name  string
		email string
	}{}
	for rows.Next() {
		var k, n, e string
		if err := rows.Scan(&k, &n, &e); err != nil {
			return []mismatch{{Where: "calcutta_entries:" + calcuttaKey, What: err.Error()}}
		}
		inDB[k] = struct {
			name  string
			email string
		}{name: n, email: e}
	}
	if err := rows.Err(); err != nil {
		return []mismatch{{Where: "calcutta_entries:" + calcuttaKey, What: err.Error()}}
	}

	var out []mismatch
	for _, e := range entries {
		db, ok := inDB[e.Key]
		if !ok {
			out = append(out, mismatch{Where: "calcutta_entries:" + calcuttaKey, What: fmt.Sprintf("missing entry_key %s", e.Key)})
			continue
		}
		if db.name != e.Name {
			out = append(out, mismatch{Where: "calcutta_entries:" + calcuttaKey, What: fmt.Sprintf("entry %s name mismatch db=%q bundle=%q", e.Key, db.name, e.Name)})
		}
		if e.UserEmail != nil {
			if db.email != *e.UserEmail {
				out = append(out, mismatch{Where: "calcutta_entries:" + calcuttaKey, What: fmt.Sprintf("entry %s email mismatch db=%q bundle=%q", e.Key, db.email, *e.UserEmail)})
			}
		}
	}

	rows, err = pool.Query(ctx, `
		SELECT ce.key, s.slug, cet.bid
		FROM calcutta_entry_teams cet
		JOIN calcutta_entries ce ON ce.id = cet.entry_id
		JOIN tournament_teams tt ON tt.id = cet.team_id
		JOIN schools s ON s.id = tt.school_id
		WHERE ce.calcutta_id = $1 AND cet.deleted_at IS NULL AND ce.deleted_at IS NULL AND tt.deleted_at IS NULL AND s.deleted_at IS NULL
	`, calcuttaID)
	if err != nil {
		return append(out, mismatch{Where: "calcutta_bids:" + calcuttaKey, What: err.Error()})
	}
	defer rows.Close()

	bidDB := map[string]int{}
	for rows.Next() {
		var entryKey, schoolSlug string
		var bid int
		if err := rows.Scan(&entryKey, &schoolSlug, &bid); err != nil {
			return append(out, mismatch{Where: "calcutta_bids:" + calcuttaKey, What: err.Error()})
		}
		bidDB[entryKey+"|"+schoolSlug] = bid
	}
	if err := rows.Err(); err != nil {
		return append(out, mismatch{Where: "calcutta_bids:" + calcuttaKey, What: err.Error()})
	}

	_ = tournamentID
	for _, b := range bids {
		k := b.EntryKey + "|" + b.SchoolSlug
		dbBid, ok := bidDB[k]
		if !ok {
			out = append(out, mismatch{Where: "calcutta_bids:" + calcuttaKey, What: fmt.Sprintf("missing bid %s", k)})
			continue
		}
		if dbBid != b.Bid {
			out = append(out, mismatch{Where: "calcutta_bids:" + calcuttaKey, What: fmt.Sprintf("bid mismatch %s db=%d bundle=%d", k, dbBid, b.Bid)})
		}
	}
	return out
}

func deref(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func derefF(p *float64) float64 {
	if p == nil {
		return 0
	}
	return *p
}

func approxEq(a, b float64) bool {
	if math.IsNaN(a) || math.IsNaN(b) {
		return false
	}
	return math.Abs(a-b) < 1e-9
}

var _ = pgx.ErrNoRows
