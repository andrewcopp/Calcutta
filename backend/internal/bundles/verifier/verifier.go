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
	mismatches = append(mismatches, verifyPools(ctx, pool, inDir)...)

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

func verifyPools(ctx context.Context, pool *pgxpool.Pool, inDir string) []Mismatch {
	root := filepath.Join(inDir, "pools")

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
		return []Mismatch{{Where: "pools", What: err.Error()}}
	}
	sort.Strings(paths)

	var out []Mismatch
	for _, path := range paths {
		var b bundles.PoolBundle
		if err := bundles.ReadJSON(path, &b); err != nil {
			out = append(out, Mismatch{Where: "pools", What: err.Error()})
			continue
		}

		var tournamentID string
		err := pool.QueryRow(ctx, `SELECT id FROM core.tournaments WHERE import_key = $1 AND deleted_at IS NULL`, b.Tournament.ImportKey).Scan(&tournamentID)
		if err != nil {
			out = append(out, Mismatch{Where: "pools:" + b.Pool.Key, What: "tournament missing"})
			continue
		}

		var poolID string
		var poolName string
		var ownerEmail string
		err = pool.QueryRow(ctx, `
			SELECT p.id, p.name, COALESCE(u.email, '')
			FROM core.pools p
			JOIN core.users u ON u.id = p.owner_id
			WHERE p.name = $1 AND p.tournament_id = $2 AND p.deleted_at IS NULL AND u.deleted_at IS NULL
		`, b.Pool.Name, tournamentID).Scan(&poolID, &poolName, &ownerEmail)
		if err != nil {
			out = append(out, Mismatch{Where: "pools:" + b.Pool.Key, What: "missing in db"})
			continue
		}

		if poolName != b.Pool.Name {
			out = append(out, Mismatch{Where: "pools:" + b.Pool.Key, What: fmt.Sprintf("name mismatch db=%q bundle=%q", poolName, b.Pool.Name)})
		}
		if b.Pool.Owner != nil && b.Pool.Owner.Email != nil {
			if ownerEmail != *b.Pool.Owner.Email {
				out = append(out, Mismatch{Where: "pools:" + b.Pool.Key, What: fmt.Sprintf("owner email mismatch db=%q bundle=%q", ownerEmail, *b.Pool.Owner.Email)})
			}
		}

		out = append(out, verifyScoringRules(ctx, pool, poolID, b.Pool.Key, b.Rounds)...)
		out = append(out, verifyPoolPayouts(ctx, pool, poolID, b.Pool.Key, b.Payouts)...)
		out = append(out, verifyPoolPortfoliosAndInvestments(ctx, pool, poolID, b.Pool.Key, b.Portfolios, b.Investments)...)
	}

	return out
}

func verifyScoringRules(ctx context.Context, pool *pgxpool.Pool, poolID, poolKey string, rounds []bundles.RoundRecord) []Mismatch {
	rows, err := pool.Query(ctx, `SELECT win_index AS round, points_awarded AS points FROM core.pool_scoring_rules WHERE pool_id = $1 AND deleted_at IS NULL`, poolID)
	if err != nil {
		return []Mismatch{{Where: "pool_scoring_rules:" + poolKey, What: err.Error()}}
	}
	defer rows.Close()

	inDB := map[int]int{}
	for rows.Next() {
		var r, p int
		if err := rows.Scan(&r, &p); err != nil {
			return []Mismatch{{Where: "pool_scoring_rules:" + poolKey, What: err.Error()}}
		}
		inDB[r] = p
	}
	if err := rows.Err(); err != nil {
		return []Mismatch{{Where: "pool_scoring_rules:" + poolKey, What: err.Error()}}
	}

	var out []Mismatch
	for _, rr := range rounds {
		p, ok := inDB[rr.Round]
		if !ok {
			out = append(out, Mismatch{Where: "pool_scoring_rules:" + poolKey, What: fmt.Sprintf("missing round %d", rr.Round)})
			continue
		}
		if p != rr.Points {
			out = append(out, Mismatch{Where: "pool_scoring_rules:" + poolKey, What: fmt.Sprintf("round %d points mismatch db=%d bundle=%d", rr.Round, p, rr.Points)})
		}
	}
	return out
}

func verifyPoolPayouts(ctx context.Context, pool *pgxpool.Pool, poolID, poolKey string, payouts []bundles.PayoutRecord) []Mismatch {
	rows, err := pool.Query(ctx, `SELECT position, amount_cents FROM core.payouts WHERE pool_id = $1 AND deleted_at IS NULL`, poolID)
	if err != nil {
		return []Mismatch{{Where: "pool_payouts:" + poolKey, What: err.Error()}}
	}
	defer rows.Close()

	inDB := map[int]int{}
	for rows.Next() {
		var pos, cents int
		if err := rows.Scan(&pos, &cents); err != nil {
			return []Mismatch{{Where: "pool_payouts:" + poolKey, What: err.Error()}}
		}
		inDB[pos] = cents
	}
	if err := rows.Err(); err != nil {
		return []Mismatch{{Where: "pool_payouts:" + poolKey, What: err.Error()}}
	}

	var out []Mismatch
	for _, p := range payouts {
		c, ok := inDB[p.Position]
		if !ok {
			out = append(out, Mismatch{Where: "pool_payouts:" + poolKey, What: fmt.Sprintf("missing payout position %d", p.Position)})
			continue
		}
		if c != p.AmountCents {
			out = append(out, Mismatch{Where: "pool_payouts:" + poolKey, What: fmt.Sprintf("payout %d cents mismatch db=%d bundle=%d", p.Position, c, p.AmountCents)})
		}
	}
	return out
}

func verifyPoolPortfoliosAndInvestments(ctx context.Context, pool *pgxpool.Pool, poolID, poolKey string, portfolios []bundles.PortfolioRecord, investments []bundles.InvestmentRecord) []Mismatch {
	rows, err := pool.Query(ctx, `
		SELECT p.id, p.name, COALESCE(u.email, '')
		FROM core.portfolios p
		LEFT JOIN core.users u ON u.id = p.user_id
		WHERE p.pool_id = $1 AND p.deleted_at IS NULL
	`, poolID)
	if err != nil {
		return []Mismatch{{Where: "pool_portfolios:" + poolKey, What: err.Error()}}
	}
	defer rows.Close()

	// Map by portfolio name for lookup (since we no longer have legacy IDs)
	inDBByName := map[string]struct {
		id    string
		email string
	}{}
	for rows.Next() {
		var id, n, e string
		if err := rows.Scan(&id, &n, &e); err != nil {
			return []Mismatch{{Where: "pool_portfolios:" + poolKey, What: err.Error()}}
		}
		inDBByName[n] = struct {
			id    string
			email string
		}{id: id, email: e}
	}
	if err := rows.Err(); err != nil {
		return []Mismatch{{Where: "pool_portfolios:" + poolKey, What: err.Error()}}
	}

	// Build portfolio name to ID mapping for investment verification
	portfolioIDByName := make(map[string]string, len(portfolios))
	var out []Mismatch
	for _, pf := range portfolios {
		db, ok := inDBByName[pf.Name]
		if !ok {
			out = append(out, Mismatch{Where: "pool_portfolios:" + poolKey, What: fmt.Sprintf("missing portfolio_key %s", pf.Key)})
			continue
		}
		portfolioIDByName[pf.Name] = db.id
		if pf.UserEmail != nil {
			if db.email != *pf.UserEmail {
				out = append(out, Mismatch{Where: "pool_portfolios:" + poolKey, What: fmt.Sprintf("portfolio %s email mismatch db=%q bundle=%q", pf.Key, db.email, *pf.UserEmail)})
			}
		}
	}

	rows, err = pool.Query(ctx, `
		SELECT inv.portfolio_id, s.slug, inv.credits
		FROM core.investments inv
		JOIN core.portfolios p ON p.id = inv.portfolio_id
		JOIN core.teams t ON t.id = inv.team_id
		JOIN core.schools s ON s.id = t.school_id
		WHERE p.pool_id = $1 AND inv.deleted_at IS NULL AND p.deleted_at IS NULL AND t.deleted_at IS NULL AND s.deleted_at IS NULL
	`, poolID)
	if err != nil {
		return append(out, Mismatch{Where: "pool_investments:" + poolKey, What: err.Error()})
	}
	defer rows.Close()

	investmentDB := map[string]int{}
	for rows.Next() {
		var portfolioID, schoolSlug string
		var credits int
		if err := rows.Scan(&portfolioID, &schoolSlug, &credits); err != nil {
			return append(out, Mismatch{Where: "pool_investments:" + poolKey, What: err.Error()})
		}
		investmentDB[portfolioID+":"+schoolSlug] = credits
	}
	if err := rows.Err(); err != nil {
		return append(out, Mismatch{Where: "pool_investments:" + poolKey, What: err.Error()})
	}

	// Build portfolio key to name mapping
	portfolioNameByKey := make(map[string]string, len(portfolios))
	for _, pf := range portfolios {
		portfolioNameByKey[pf.Key] = pf.Name
	}

	for _, inv := range investments {
		portfolioName, ok := portfolioNameByKey[inv.PortfolioKey]
		if !ok || portfolioName == "" {
			out = append(out, Mismatch{Where: "pool_investments:" + poolKey, What: fmt.Sprintf("investment references unknown portfolio_key %s", inv.PortfolioKey)})
			continue
		}
		portfolioID, ok := portfolioIDByName[portfolioName]
		if !ok || portfolioID == "" {
			out = append(out, Mismatch{Where: "pool_investments:" + poolKey, What: fmt.Sprintf("investment portfolio %s not found in db", portfolioName)})
			continue
		}
		k := portfolioID + ":" + inv.SchoolSlug
		v, ok := investmentDB[k]
		if !ok {
			out = append(out, Mismatch{Where: "pool_investments:" + poolKey, What: fmt.Sprintf("missing investment %s", k)})
			continue
		}
		if v != inv.Credits {
			out = append(out, Mismatch{Where: "pool_investments:" + poolKey, What: fmt.Sprintf("investment mismatch %s db=%d bundle=%d", k, v, inv.Credits)})
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
