package verifier

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/andrewcopp/Calcutta/backend/internal/bundles"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Mismatch struct {
	Where string `json:"where"`
	What  string `json:"what"`
}

type Report struct {
	OK            bool       `json:"ok"`
	MismatchCount int        `json:"mismatchCount"`
	Mismatches    []Mismatch `json:"mismatches,omitempty"`
}

func VerifyDirAgainstDB(ctx context.Context, pool *pgxpool.Pool, inDir string) (Report, error) {
	var mismatches []Mismatch

	// Schools and tournaments are now seeded via migrations, so only verify
	// if the bundle directory contains those files.
	if _, err := os.Stat(filepath.Join(inDir, "schools.json")); err == nil {
		mismatches = append(mismatches, verifySchools(ctx, pool, inDir)...)
	}
	if paths, _ := filepath.Glob(filepath.Join(inDir, "tournaments", "*.json")); len(paths) > 0 {
		mismatches = append(mismatches, verifyTournaments(ctx, pool, inDir)...)
	}
	mismatches = append(mismatches, verifyCalcuttas(ctx, pool, inDir)...)

	r := Report{OK: len(mismatches) == 0, MismatchCount: len(mismatches)}
	if len(mismatches) > 0 {
		if len(mismatches) > 50 {
			r.Mismatches = mismatches[:50]
		} else {
			r.Mismatches = mismatches
		}
	}
	return r, nil
}

func verifySchools(ctx context.Context, pool *pgxpool.Pool, inDir string) []Mismatch {
	path := filepath.Join(inDir, "schools.json")
	var b bundles.SchoolsBundle
	if err := bundles.ReadJSON(path, &b); err != nil {
		return []Mismatch{{Where: "schools", What: err.Error()}}
	}

	rows, err := pool.Query(ctx, `SELECT slug, name FROM core.schools WHERE deleted_at IS NULL`)
	if err != nil {
		return []Mismatch{{Where: "schools", What: err.Error()}}
	}
	defer rows.Close()

	inDB := map[string]string{}
	for rows.Next() {
		var slug, name string
		if err := rows.Scan(&slug, &name); err != nil {
			return []Mismatch{{Where: "schools", What: err.Error()}}
		}
		inDB[slug] = name
	}
	if err := rows.Err(); err != nil {
		return []Mismatch{{Where: "schools", What: err.Error()}}
	}

	var out []Mismatch
	for _, s := range b.Schools {
		name, ok := inDB[s.Slug]
		if !ok {
			out = append(out, Mismatch{Where: "schools:" + s.Slug, What: "missing in db"})
			continue
		}
		if name != s.Name {
			out = append(out, Mismatch{Where: "schools:" + s.Slug, What: fmt.Sprintf("name mismatch db=%q bundle=%q", name, s.Name)})
		}
	}
	return out
}

func verifyTournaments(ctx context.Context, pool *pgxpool.Pool, inDir string) []Mismatch {
	paths, err := filepath.Glob(filepath.Join(inDir, "tournaments", "*.json"))
	if err != nil {
		return []Mismatch{{Where: "tournaments", What: err.Error()}}
	}
	sort.Strings(paths)

	var out []Mismatch
	for _, path := range paths {
		var b bundles.TournamentBundle
		if err := bundles.ReadJSON(path, &b); err != nil {
			out = append(out, Mismatch{Where: "tournaments", What: err.Error()})
			continue
		}

		var id, name string
		var rounds int
		var f1, f2, f3, f4 *string
		err := pool.QueryRow(ctx, `
			SELECT id, name, rounds, final_four_top_left, final_four_bottom_left, final_four_top_right, final_four_bottom_right
			FROM core.tournaments
			WHERE import_key = $1 AND deleted_at IS NULL
		`, b.Tournament.ImportKey).Scan(&id, &name, &rounds, &f1, &f2, &f3, &f4)
		if err != nil {
			out = append(out, Mismatch{Where: "tournaments:" + b.Tournament.ImportKey, What: "missing in db"})
			continue
		}

		if name != b.Tournament.Name {
			out = append(out, Mismatch{Where: "tournaments:" + b.Tournament.ImportKey, What: fmt.Sprintf("name mismatch db=%q bundle=%q", name, b.Tournament.Name)})
		}
		if rounds != b.Tournament.Rounds {
			out = append(out, Mismatch{Where: "tournaments:" + b.Tournament.ImportKey, What: fmt.Sprintf("rounds mismatch db=%d bundle=%d", rounds, b.Tournament.Rounds)})
		}

		if bundles.DerefString(f1) != b.Tournament.FinalFourTopLeft {
			out = append(out, Mismatch{Where: "tournaments:" + b.Tournament.ImportKey, What: "final_four_top_left mismatch"})
		}
		if bundles.DerefString(f2) != b.Tournament.FinalFourBottomLeft {
			out = append(out, Mismatch{Where: "tournaments:" + b.Tournament.ImportKey, What: "final_four_bottom_left mismatch"})
		}
		if bundles.DerefString(f3) != b.Tournament.FinalFourTopRight {
			out = append(out, Mismatch{Where: "tournaments:" + b.Tournament.ImportKey, What: "final_four_top_right mismatch"})
		}
		if bundles.DerefString(f4) != b.Tournament.FinalFourBottomRight {
			out = append(out, Mismatch{Where: "tournaments:" + b.Tournament.ImportKey, What: "final_four_bottom_right mismatch"})
		}

		out = append(out, verifyTournamentTeams(ctx, pool, id, b.Tournament.ImportKey, b.Teams)...)
	}
	return out
}

func verifyTournamentTeams(ctx context.Context, pool *pgxpool.Pool, tournamentID, tournamentKey string, teams []bundles.TeamRecord) []Mismatch {
	rows, err := pool.Query(ctx, `
		SELECT
			s.slug,
			s.name,
			t.seed,
			t.region,
			t.byes,
			t.wins,
			t.is_eliminated,
			k.net_rtg,
			k.o_rtg,
			k.d_rtg,
			k.adj_t,
			(k.team_id IS NOT NULL)
		FROM core.teams t
		JOIN core.schools s ON s.id = t.school_id
		LEFT JOIN core.team_kenpom_stats k ON k.team_id = t.id AND k.deleted_at IS NULL
		WHERE t.tournament_id = $1 AND t.deleted_at IS NULL AND s.deleted_at IS NULL
	`, tournamentID)
	if err != nil {
		return []Mismatch{{Where: "tournament_teams:" + tournamentKey, What: err.Error()}}
	}
	defer rows.Close()

	tInDB := map[string]bundles.TeamRecord{}
	kpInDB := map[string]*bundles.KenPomRecord{}
	for rows.Next() {
		var slug, name, region string
		var seed, byes, wins int
		var isEliminated bool
		var net, o, d, adj *float64
		var hasKP bool
		if err := rows.Scan(&slug, &name, &seed, &region, &byes, &wins, &isEliminated, &net, &o, &d, &adj, &hasKP); err != nil {
			return []Mismatch{{Where: "tournament_teams:" + tournamentKey, What: err.Error()}}
		}
		tInDB[slug] = bundles.TeamRecord{SchoolSlug: slug, SchoolName: name, Seed: seed, Region: region, Byes: byes, Wins: wins, IsEliminated: isEliminated}
		if hasKP {
			kpInDB[slug] = &bundles.KenPomRecord{NetRTG: bundles.DerefFloat64(net), ORTG: bundles.DerefFloat64(o), DRTG: bundles.DerefFloat64(d), AdjT: bundles.DerefFloat64(adj)}
		}
	}
	if err := rows.Err(); err != nil {
		return []Mismatch{{Where: "tournament_teams:" + tournamentKey, What: err.Error()}}
	}

	var out []Mismatch
	for _, t := range teams {
		dbT, ok := tInDB[t.SchoolSlug]
		if !ok {
			out = append(out, Mismatch{Where: "tournament_teams:" + tournamentKey + ":" + t.SchoolSlug, What: "missing in db"})
			continue
		}
		if dbT.Seed != t.Seed || dbT.Region != t.Region || dbT.Byes != t.Byes || dbT.Wins != t.Wins || dbT.IsEliminated != t.IsEliminated {
			out = append(out, Mismatch{Where: "tournament_teams:" + tournamentKey + ":" + t.SchoolSlug, What: "field mismatch"})
		}

		dbKP := kpInDB[t.SchoolSlug]
		if t.KenPom == nil {
			if dbKP != nil {
				out = append(out, Mismatch{Where: "tournament_teams:" + tournamentKey + ":" + t.SchoolSlug, What: "kenpom present in db but missing in bundle"})
			}
			continue
		}
		if dbKP == nil {
			out = append(out, Mismatch{Where: "tournament_teams:" + tournamentKey + ":" + t.SchoolSlug, What: "kenpom missing in db"})
			continue
		}
		if !approxEq(dbKP.NetRTG, t.KenPom.NetRTG) || !approxEq(dbKP.ORTG, t.KenPom.ORTG) || !approxEq(dbKP.DRTG, t.KenPom.DRTG) || !approxEq(dbKP.AdjT, t.KenPom.AdjT) {
			out = append(out, Mismatch{Where: "tournament_teams:" + tournamentKey + ":" + t.SchoolSlug, What: "kenpom mismatch"})
		}
	}
	return out
}

func verifyCalcuttas(ctx context.Context, pool *pgxpool.Pool, inDir string) []Mismatch {
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
		return []Mismatch{{Where: "calcuttas", What: err.Error()}}
	}
	sort.Strings(paths)

	var out []Mismatch
	for _, path := range paths {
		var b bundles.CalcuttaBundle
		if err := bundles.ReadJSON(path, &b); err != nil {
			out = append(out, Mismatch{Where: "calcuttas", What: err.Error()})
			continue
		}

		var tournamentID string
		err := pool.QueryRow(ctx, `SELECT id FROM core.tournaments WHERE import_key = $1 AND deleted_at IS NULL`, b.Tournament.ImportKey).Scan(&tournamentID)
		if err != nil {
			out = append(out, Mismatch{Where: "calcuttas:" + b.Calcutta.Key, What: "tournament missing"})
			continue
		}

		var calcuttaID string
		var calcuttaName string
		var ownerEmail string
		err = pool.QueryRow(ctx, `
			SELECT c.id, c.name, COALESCE(u.email, '')
			FROM core.calcuttas c
			JOIN core.users u ON u.id = c.owner_id
			WHERE c.name = $1 AND c.tournament_id = $2 AND c.deleted_at IS NULL AND u.deleted_at IS NULL
		`, b.Calcutta.Name, tournamentID).Scan(&calcuttaID, &calcuttaName, &ownerEmail)
		if err != nil {
			out = append(out, Mismatch{Where: "calcuttas:" + b.Calcutta.Key, What: "missing in db"})
			continue
		}

		if calcuttaName != b.Calcutta.Name {
			out = append(out, Mismatch{Where: "calcuttas:" + b.Calcutta.Key, What: fmt.Sprintf("name mismatch db=%q bundle=%q", calcuttaName, b.Calcutta.Name)})
		}
		if b.Calcutta.Owner != nil && b.Calcutta.Owner.Email != nil {
			if ownerEmail != *b.Calcutta.Owner.Email {
				out = append(out, Mismatch{Where: "calcuttas:" + b.Calcutta.Key, What: fmt.Sprintf("owner email mismatch db=%q bundle=%q", ownerEmail, *b.Calcutta.Owner.Email)})
			}
		}

		out = append(out, verifyScoringRules(ctx, pool, calcuttaID, b.Calcutta.Key, b.Rounds)...)
		out = append(out, verifyCalcuttaPayouts(ctx, pool, calcuttaID, b.Calcutta.Key, b.Payouts)...)
		out = append(out, verifyCalcuttaEntriesAndBids(ctx, pool, calcuttaID, b.Calcutta.Key, b.Entries, b.Bids)...)
	}

	return out
}

func verifyScoringRules(ctx context.Context, pool *pgxpool.Pool, calcuttaID, calcuttaKey string, rounds []bundles.RoundRecord) []Mismatch {
	rows, err := pool.Query(ctx, `SELECT win_index AS round, points_awarded AS points FROM core.calcutta_scoring_rules WHERE calcutta_id = $1 AND deleted_at IS NULL`, calcuttaID)
	if err != nil {
		return []Mismatch{{Where: "calcutta_scoring_rules:" + calcuttaKey, What: err.Error()}}
	}
	defer rows.Close()

	inDB := map[int]int{}
	for rows.Next() {
		var r, p int
		if err := rows.Scan(&r, &p); err != nil {
			return []Mismatch{{Where: "calcutta_scoring_rules:" + calcuttaKey, What: err.Error()}}
		}
		inDB[r] = p
	}
	if err := rows.Err(); err != nil {
		return []Mismatch{{Where: "calcutta_scoring_rules:" + calcuttaKey, What: err.Error()}}
	}

	var out []Mismatch
	for _, rr := range rounds {
		p, ok := inDB[rr.Round]
		if !ok {
			out = append(out, Mismatch{Where: "calcutta_scoring_rules:" + calcuttaKey, What: fmt.Sprintf("missing round %d", rr.Round)})
			continue
		}
		if p != rr.Points {
			out = append(out, Mismatch{Where: "calcutta_scoring_rules:" + calcuttaKey, What: fmt.Sprintf("round %d points mismatch db=%d bundle=%d", rr.Round, p, rr.Points)})
		}
	}
	return out
}

func verifyCalcuttaPayouts(ctx context.Context, pool *pgxpool.Pool, calcuttaID, calcuttaKey string, payouts []bundles.PayoutRecord) []Mismatch {
	rows, err := pool.Query(ctx, `SELECT position, amount_cents FROM core.payouts WHERE calcutta_id = $1 AND deleted_at IS NULL`, calcuttaID)
	if err != nil {
		return []Mismatch{{Where: "calcutta_payouts:" + calcuttaKey, What: err.Error()}}
	}
	defer rows.Close()

	inDB := map[int]int{}
	for rows.Next() {
		var pos, cents int
		if err := rows.Scan(&pos, &cents); err != nil {
			return []Mismatch{{Where: "calcutta_payouts:" + calcuttaKey, What: err.Error()}}
		}
		inDB[pos] = cents
	}
	if err := rows.Err(); err != nil {
		return []Mismatch{{Where: "calcutta_payouts:" + calcuttaKey, What: err.Error()}}
	}

	var out []Mismatch
	for _, p := range payouts {
		c, ok := inDB[p.Position]
		if !ok {
			out = append(out, Mismatch{Where: "calcutta_payouts:" + calcuttaKey, What: fmt.Sprintf("missing payout position %d", p.Position)})
			continue
		}
		if c != p.AmountCents {
			out = append(out, Mismatch{Where: "calcutta_payouts:" + calcuttaKey, What: fmt.Sprintf("payout %d cents mismatch db=%d bundle=%d", p.Position, c, p.AmountCents)})
		}
	}
	return out
}

func verifyCalcuttaEntriesAndBids(ctx context.Context, pool *pgxpool.Pool, calcuttaID, calcuttaKey string, entries []bundles.EntryRecord, bids []bundles.EntryTeamBid) []Mismatch {
	rows, err := pool.Query(ctx, `
		SELECT e.id, e.name, COALESCE(u.email, '')
		FROM core.entries e
		LEFT JOIN core.users u ON u.id = e.user_id
		WHERE e.calcutta_id = $1 AND e.deleted_at IS NULL
	`, calcuttaID)
	if err != nil {
		return []Mismatch{{Where: "calcutta_entries:" + calcuttaKey, What: err.Error()}}
	}
	defer rows.Close()

	// Map by entry name for lookup (since we no longer have legacy IDs)
	inDBByName := map[string]struct {
		id    string
		email string
	}{}
	for rows.Next() {
		var id, n, e string
		if err := rows.Scan(&id, &n, &e); err != nil {
			return []Mismatch{{Where: "calcutta_entries:" + calcuttaKey, What: err.Error()}}
		}
		inDBByName[n] = struct {
			id    string
			email string
		}{id: id, email: e}
	}
	if err := rows.Err(); err != nil {
		return []Mismatch{{Where: "calcutta_entries:" + calcuttaKey, What: err.Error()}}
	}

	// Build entry name to ID mapping for bid verification
	entryIDByName := make(map[string]string, len(entries))
	var out []Mismatch
	for _, e := range entries {
		db, ok := inDBByName[e.Name]
		if !ok {
			out = append(out, Mismatch{Where: "calcutta_entries:" + calcuttaKey, What: fmt.Sprintf("missing entry_key %s", e.Key)})
			continue
		}
		entryIDByName[e.Name] = db.id
		if e.UserEmail != nil {
			if db.email != *e.UserEmail {
				out = append(out, Mismatch{Where: "calcutta_entries:" + calcuttaKey, What: fmt.Sprintf("entry %s email mismatch db=%q bundle=%q", e.Key, db.email, *e.UserEmail)})
			}
		}
	}

	rows, err = pool.Query(ctx, `
		SELECT et.entry_id, s.slug, et.bid_points
		FROM core.entry_teams et
		JOIN core.entries e ON e.id = et.entry_id
		JOIN core.teams t ON t.id = et.team_id
		JOIN core.schools s ON s.id = t.school_id
		WHERE e.calcutta_id = $1 AND et.deleted_at IS NULL AND e.deleted_at IS NULL AND t.deleted_at IS NULL AND s.deleted_at IS NULL
	`, calcuttaID)
	if err != nil {
		return append(out, Mismatch{Where: "calcutta_bids:" + calcuttaKey, What: err.Error()})
	}
	defer rows.Close()

	bidDB := map[string]int{}
	for rows.Next() {
		var entryID, schoolSlug string
		var bid int
		if err := rows.Scan(&entryID, &schoolSlug, &bid); err != nil {
			return append(out, Mismatch{Where: "calcutta_bids:" + calcuttaKey, What: err.Error()})
		}
		bidDB[entryID+":"+schoolSlug] = bid
	}
	if err := rows.Err(); err != nil {
		return append(out, Mismatch{Where: "calcutta_bids:" + calcuttaKey, What: err.Error()})
	}

	// Build entry key to name mapping
	entryNameByKey := make(map[string]string, len(entries))
	for _, e := range entries {
		entryNameByKey[e.Key] = e.Name
	}

	for _, b := range bids {
		entryName, ok := entryNameByKey[b.EntryKey]
		if !ok || entryName == "" {
			out = append(out, Mismatch{Where: "calcutta_bids:" + calcuttaKey, What: fmt.Sprintf("bid references unknown entry_key %s", b.EntryKey)})
			continue
		}
		entryID, ok := entryIDByName[entryName]
		if !ok || entryID == "" {
			out = append(out, Mismatch{Where: "calcutta_bids:" + calcuttaKey, What: fmt.Sprintf("bid entry %s not found in db", entryName)})
			continue
		}
		k := entryID + ":" + b.SchoolSlug
		v, ok := bidDB[k]
		if !ok {
			out = append(out, Mismatch{Where: "calcutta_bids:" + calcuttaKey, What: fmt.Sprintf("missing bid %s", k)})
			continue
		}
		if v != b.Bid {
			out = append(out, Mismatch{Where: "calcutta_bids:" + calcuttaKey, What: fmt.Sprintf("bid mismatch %s db=%d bundle=%d", k, v, b.Bid)})
		}
	}
	return out
}

func approxEq(a, b float64) bool {
	if math.IsNaN(a) || math.IsNaN(b) {
		return false
	}
	return math.Abs(a-b) < 1e-9
}
