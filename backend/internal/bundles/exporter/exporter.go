package exporter

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/bundles"
	"github.com/jackc/pgx/v5/pgxpool"
)

var nonAlnum = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = nonAlnum.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}

func uniquifyKey(base string, used map[string]int) string {
	if base == "" {
		base = "unknown"
	}
	if _, ok := used[base]; !ok {
		used[base] = 1
		return base
	}
	used[base]++
	return fmt.Sprintf("%s-%d", base, used[base])
}

func ExportToDir(ctx context.Context, pool *pgxpool.Pool, outDir string, generatedAt time.Time) error {
	if err := exportSchools(ctx, pool, outDir, generatedAt); err != nil {
		return err
	}
	if err := exportTournaments(ctx, pool, outDir, generatedAt); err != nil {
		return err
	}
	if err := exportPools(ctx, pool, outDir, generatedAt); err != nil {
		return err
	}
	return nil
}

func exportSchools(ctx context.Context, pool *pgxpool.Pool, outDir string, generatedAt time.Time) error {
	r, err := pool.Query(ctx, `SELECT slug, name FROM core.schools WHERE deleted_at IS NULL ORDER BY name ASC`)
	if err != nil {
		return err
	}
	defer r.Close()

	out := bundles.SchoolsBundle{Version: 1, GeneratedAt: generatedAt, Schools: make([]bundles.SchoolEntry, 0)}
	for r.Next() {
		var slug, name string
		if err := r.Scan(&slug, &name); err != nil {
			return err
		}
		out.Schools = append(out.Schools, bundles.SchoolEntry{Slug: slug, Name: name})
	}
	if err := r.Err(); err != nil {
		return err
	}

	return bundles.WriteJSON(filepath.Join(outDir, "schools.json"), out)
}

func exportTournaments(ctx context.Context, pool *pgxpool.Pool, outDir string, generatedAt time.Time) error {
	r, err := pool.Query(ctx, `
		SELECT t.id, t.import_key,
		       comp.name || ' (' || seas.year || ')' AS name,
		       t.rounds, t.starting_at,
		       COALESCE(t.final_four_top_left, ''),
		       COALESCE(t.final_four_bottom_left, ''),
		       COALESCE(t.final_four_top_right, ''),
		       COALESCE(t.final_four_bottom_right, '')
		FROM core.tournaments t
		JOIN core.competitions comp ON comp.id = t.competition_id
		JOIN core.seasons seas ON seas.id = t.season_id
		WHERE t.deleted_at IS NULL
		ORDER BY seas.year ASC
	`)
	if err != nil {
		return err
	}
	defer r.Close()

	for r.Next() {
		var tournamentID, importKey, name string
		var rounds int
		var startingAt *time.Time
		var f1, f2, f3, f4 string
		if err := r.Scan(&tournamentID, &importKey, &name, &rounds, &startingAt, &f1, &f2, &f3, &f4); err != nil {
			return err
		}

		teams, err := loadTournamentTeams(ctx, pool, tournamentID)
		if err != nil {
			return err
		}

		bundle := bundles.TournamentBundle{
			Version:     1,
			GeneratedAt: generatedAt,
			Tournament: bundles.TournamentRecord{
				ImportKey:            importKey,
				Name:                 name,
				Rounds:               rounds,
				StartingAt:           startingAt,
				FinalFourTopLeft:     f1,
				FinalFourBottomLeft:  f2,
				FinalFourTopRight:    f3,
				FinalFourBottomRight: f4,
			},
			Teams: teams,
		}

		path := filepath.Join(outDir, "tournaments", fmt.Sprintf("%s.json", importKey))
		if err := bundles.WriteJSON(path, bundle); err != nil {
			return err
		}
	}
	return r.Err()
}

func loadTournamentTeams(ctx context.Context, pool *pgxpool.Pool, tournamentID string) ([]bundles.TeamRecord, error) {
	r, err := pool.Query(ctx, `
		SELECT
			t.id,
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
		ORDER BY t.created_at ASC
	`, tournamentID)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	out := make([]bundles.TeamRecord, 0)
	for r.Next() {
		var teamID, slug, name, region string
		var seed, byes, wins int
		var isEliminated bool
		var net, o, d, adj *float64
		var hasKP bool
		if err := r.Scan(&teamID, &slug, &name, &seed, &region, &byes, &wins, &isEliminated, &net, &o, &d, &adj, &hasKP); err != nil {
			return nil, err
		}

		tr := bundles.TeamRecord{SchoolSlug: slug, SchoolName: name, Seed: seed, Region: region, Byes: byes, Wins: wins, IsEliminated: isEliminated}
		if hasKP {
			tr.KenPom = &bundles.KenPomRecord{NetRTG: bundles.DerefFloat64(net), ORTG: bundles.DerefFloat64(o), DRTG: bundles.DerefFloat64(d), AdjT: bundles.DerefFloat64(adj)}
		}
		out = append(out, tr)
	}
	if err := r.Err(); err != nil {
		return nil, err
	}

	sort.Slice(out, func(i, j int) bool { return out[i].SchoolSlug < out[j].SchoolSlug })
	return out, nil
}

func exportPools(ctx context.Context, pool *pgxpool.Pool, outDir string, generatedAt time.Time) error {
	r, err := pool.Query(ctx, `
		SELECT
			p.id,
			p.name,
			p.owner_id,
			t.import_key,
			comp.name || ' (' || seas.year || ')' AS tournament_name,
			COALESCE(u.email, ''),
			COALESCE(u.first_name, ''),
			COALESCE(u.last_name, '')
		FROM core.pools p
		JOIN core.tournaments t ON t.id = p.tournament_id
		JOIN core.competitions comp ON comp.id = t.competition_id
		JOIN core.seasons seas ON seas.id = t.season_id
		JOIN core.users u ON u.id = p.owner_id
		WHERE p.deleted_at IS NULL AND t.deleted_at IS NULL AND u.deleted_at IS NULL
		ORDER BY seas.year ASC, p.created_at ASC
	`)
	if err != nil {
		return err
	}
	defer r.Close()

	usedPoolKeysByTournament := make(map[string]map[string]int)

	for r.Next() {
		var poolID, poolName, ownerID, tournamentKey, tournamentName string
		var email, first, last string
		if err := r.Scan(&poolID, &poolName, &ownerID, &tournamentKey, &tournamentName, &email, &first, &last); err != nil {
			return err
		}

		if usedPoolKeysByTournament[tournamentKey] == nil {
			usedPoolKeysByTournament[tournamentKey] = make(map[string]int)
		}
		poolKey := uniquifyKey("pool-"+slugify(poolName), usedPoolKeysByTournament[tournamentKey])

		rounds, err := loadScoringRules(ctx, pool, poolID)
		if err != nil {
			return err
		}
		payouts, err := loadPoolPayouts(ctx, pool, poolID)
		if err != nil {
			return err
		}
		portfolios, investments, err := loadPoolPortfoliosAndInvestments(ctx, pool, poolID, poolKey)
		if err != nil {
			return err
		}

		owner := &bundles.UserRef{}
		if email != "" {
			owner.Email = &email
		}
		if first != "" {
			owner.FirstName = &first
		}
		if last != "" {
			owner.LastName = &last
		}

		bundle := bundles.PoolBundle{
			Version:     1,
			GeneratedAt: generatedAt,
			Tournament:  bundles.TournamentRef{ImportKey: tournamentKey, Name: tournamentName},
			Pool: bundles.PoolRecord{
				Key:   poolKey,
				Name:  poolName,
				Owner: owner,
			},
			Rounds:      rounds,
			Payouts:     payouts,
			Portfolios:  portfolios,
			Investments: investments,
		}

		path := filepath.Join(outDir, "pools", tournamentKey, fmt.Sprintf("%s.json", poolKey))
		if err := bundles.WriteJSON(path, bundle); err != nil {
			return err
		}
	}
	return r.Err()
}

func loadScoringRules(ctx context.Context, pool *pgxpool.Pool, poolID string) ([]bundles.RoundRecord, error) {
	r, err := pool.Query(ctx, `SELECT win_index AS round, points_awarded AS points FROM core.pool_scoring_rules WHERE pool_id = $1 AND deleted_at IS NULL ORDER BY win_index ASC`, poolID)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	out := make([]bundles.RoundRecord, 0)
	for r.Next() {
		var round, points int
		if err := r.Scan(&round, &points); err != nil {
			return nil, err
		}
		out = append(out, bundles.RoundRecord{Round: round, Points: points})
	}
	return out, r.Err()
}

func loadPoolPayouts(ctx context.Context, pool *pgxpool.Pool, poolID string) ([]bundles.PayoutRecord, error) {
	r, err := pool.Query(ctx, `SELECT position, amount_cents FROM core.payouts WHERE pool_id = $1 AND deleted_at IS NULL ORDER BY position ASC`, poolID)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	out := make([]bundles.PayoutRecord, 0)
	for r.Next() {
		var pos, cents int
		if err := r.Scan(&pos, &cents); err != nil {
			return nil, err
		}
		out = append(out, bundles.PayoutRecord{Position: pos, AmountCents: cents})
	}
	return out, r.Err()
}

func loadPoolPortfoliosAndInvestments(ctx context.Context, pool *pgxpool.Pool, poolID string, poolKey string) ([]bundles.PortfolioRecord, []bundles.InvestmentRecord, error) {
	r, err := pool.Query(ctx, `
		SELECT
			p.id,
			p.name,
			COALESCE(u.email, ''),
			COALESCE(u.first_name, ''),
			COALESCE(u.last_name, '')
		FROM core.portfolios p
		LEFT JOIN core.users u ON u.id = p.user_id
		WHERE p.pool_id = $1 AND p.deleted_at IS NULL
		ORDER BY p.created_at ASC
	`, poolID)
	if err != nil {
		return nil, nil, err
	}
	defer r.Close()

	usedPortfolioKeys := make(map[string]int)
	portfolioKeyByPortfolioID := make(map[string]string)

	portfolios := make([]bundles.PortfolioRecord, 0)
	for r.Next() {
		var portfolioID, name string
		var email, first, last string
		if err := r.Scan(&portfolioID, &name, &email, &first, &last); err != nil {
			return nil, nil, err
		}

		base := "portfolio-" + slugify(name)
		portfolioKey := uniquifyKey(base, usedPortfolioKeys)
		portfolioKey = fmt.Sprintf("%s:%s", poolKey, portfolioKey)
		portfolioKeyByPortfolioID[portfolioID] = portfolioKey

		var userName *string
		if first != "" || last != "" {
			full := strings.TrimSpace(strings.TrimSpace(first) + " " + strings.TrimSpace(last))
			if full != "" {
				userName = &full
			}
		}

		var userEmail *string
		if email != "" {
			userEmail = &email
		}

		portfolios = append(portfolios, bundles.PortfolioRecord{Key: portfolioKey, Name: name, UserName: userName, UserEmail: userEmail})
	}
	if err := r.Err(); err != nil {
		return nil, nil, err
	}

	investments, err := loadPoolInvestments(ctx, pool, poolID, portfolioKeyByPortfolioID)
	if err != nil {
		return nil, nil, err
	}

	sort.Slice(portfolios, func(i, j int) bool { return portfolios[i].Key < portfolios[j].Key })
	return portfolios, investments, nil
}

func loadPoolInvestments(ctx context.Context, pool *pgxpool.Pool, poolID string, portfolioKeyByPortfolioID map[string]string) ([]bundles.InvestmentRecord, error) {
	r, err := pool.Query(ctx, `
		SELECT
			inv.portfolio_id,
			inv.credits,
			s.slug
		FROM core.investments inv
		JOIN core.portfolios p ON p.id = inv.portfolio_id
		JOIN core.teams t ON t.id = inv.team_id
		JOIN core.schools s ON s.id = t.school_id
		WHERE p.pool_id = $1 AND inv.deleted_at IS NULL AND t.deleted_at IS NULL AND s.deleted_at IS NULL AND p.deleted_at IS NULL
		ORDER BY p.created_at ASC, s.name ASC
	`, poolID)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	out := make([]bundles.InvestmentRecord, 0)
	for r.Next() {
		var portfolioID, schoolSlug string
		var credits int
		if err := r.Scan(&portfolioID, &credits, &schoolSlug); err != nil {
			return nil, err
		}
		portfolioKey := portfolioKeyByPortfolioID[portfolioID]
		out = append(out, bundles.InvestmentRecord{PortfolioKey: portfolioKey, SchoolSlug: schoolSlug, Credits: credits})
	}
	if err := r.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
