package exporter

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
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

func writeJSON(path string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
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
	if err := exportCalcuttas(ctx, pool, outDir, generatedAt); err != nil {
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

	return writeJSON(filepath.Join(outDir, "schools.json"), out)
}

func exportTournaments(ctx context.Context, pool *pgxpool.Pool, outDir string, generatedAt time.Time) error {
	r, err := pool.Query(ctx, `
		SELECT id, import_key, name, rounds,
		       COALESCE(final_four_top_left, ''),
		       COALESCE(final_four_bottom_left, ''),
		       COALESCE(final_four_top_right, ''),
		       COALESCE(final_four_bottom_right, '')
		FROM core.tournaments
		WHERE deleted_at IS NULL
		ORDER BY name ASC
	`)
	if err != nil {
		return err
	}
	defer r.Close()

	for r.Next() {
		var tournamentID, importKey, name string
		var rounds int
		var f1, f2, f3, f4 string
		if err := r.Scan(&tournamentID, &importKey, &name, &rounds, &f1, &f2, &f3, &f4); err != nil {
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
				FinalFourTopLeft:     f1,
				FinalFourBottomLeft:  f2,
				FinalFourTopRight:    f3,
				FinalFourBottomRight: f4,
				LegacyTournamentID:   tournamentID,
			},
			Teams: teams,
		}

		path := filepath.Join(outDir, "tournaments", fmt.Sprintf("%s.json", importKey))
		if err := writeJSON(path, bundle); err != nil {
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
			t.eliminated,
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
		var eliminated bool
		var net, o, d, adj *float64
		var hasKP bool
		if err := r.Scan(&teamID, &slug, &name, &seed, &region, &byes, &wins, &eliminated, &net, &o, &d, &adj, &hasKP); err != nil {
			return nil, err
		}

		tr := bundles.TeamRecord{SchoolSlug: slug, SchoolName: name, Seed: seed, Region: region, Byes: byes, Wins: wins, Eliminated: eliminated, LegacyTeamID: teamID}
		if hasKP {
			tr.KenPom = &bundles.KenPomRecord{NetRTG: derefF(net), ORTG: derefF(o), DRTG: derefF(d), AdjT: derefF(adj)}
		}
		out = append(out, tr)
	}
	if err := r.Err(); err != nil {
		return nil, err
	}

	sort.Slice(out, func(i, j int) bool { return out[i].SchoolSlug < out[j].SchoolSlug })
	return out, nil
}

func exportCalcuttas(ctx context.Context, pool *pgxpool.Pool, outDir string, generatedAt time.Time) error {
	r, err := pool.Query(ctx, `
		SELECT
			c.id,
			c.name,
			c.owner_id,
			t.import_key,
			t.name,
			COALESCE(u.email, ''),
			COALESCE(u.first_name, ''),
			COALESCE(u.last_name, '')
		FROM core.calcuttas c
		JOIN core.tournaments t ON t.id = c.tournament_id
		JOIN core.users u ON u.id = c.owner_id
		WHERE c.deleted_at IS NULL AND t.deleted_at IS NULL AND u.deleted_at IS NULL
		ORDER BY t.name ASC, c.created_at ASC
	`)
	if err != nil {
		return err
	}
	defer r.Close()

	usedCalcuttaKeysByTournament := make(map[string]map[string]int)

	for r.Next() {
		var calcuttaID, calcuttaName, ownerID, tournamentKey, tournamentName string
		var email, first, last string
		if err := r.Scan(&calcuttaID, &calcuttaName, &ownerID, &tournamentKey, &tournamentName, &email, &first, &last); err != nil {
			return err
		}

		if usedCalcuttaKeysByTournament[tournamentKey] == nil {
			usedCalcuttaKeysByTournament[tournamentKey] = make(map[string]int)
		}
		calcuttaKey := uniquifyKey("calcutta-"+slugify(calcuttaName), usedCalcuttaKeysByTournament[tournamentKey])

		rounds, err := loadCalcuttaRounds(ctx, pool, calcuttaID)
		if err != nil {
			return err
		}
		payouts, err := loadCalcuttaPayouts(ctx, pool, calcuttaID)
		if err != nil {
			return err
		}
		entries, bids, err := loadCalcuttaEntriesAndBids(ctx, pool, calcuttaID, calcuttaKey)
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

		bundle := bundles.CalcuttaBundle{
			Version:     1,
			GeneratedAt: generatedAt,
			Tournament:  bundles.TournamentRef{ImportKey: tournamentKey, Name: tournamentName},
			Calcutta: bundles.CalcuttaRecord{
				Key:           calcuttaKey,
				Name:          calcuttaName,
				Owner:         owner,
				LegacyID:      calcuttaID,
				LegacyOwnerID: ownerID,
			},
			Rounds:  rounds,
			Payouts: payouts,
			Entries: entries,
			Bids:    bids,
		}

		path := filepath.Join(outDir, "calcuttas", tournamentKey, fmt.Sprintf("%s.json", calcuttaKey))
		if err := writeJSON(path, bundle); err != nil {
			return err
		}
	}
	return r.Err()
}

func loadCalcuttaRounds(ctx context.Context, pool *pgxpool.Pool, calcuttaID string) ([]bundles.RoundRecord, error) {
	r, err := pool.Query(ctx, `SELECT win_index AS round, points_awarded AS points FROM core.calcutta_scoring_rules WHERE calcutta_id = $1 AND deleted_at IS NULL ORDER BY win_index ASC`, calcuttaID)
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

func loadCalcuttaPayouts(ctx context.Context, pool *pgxpool.Pool, calcuttaID string) ([]bundles.PayoutRecord, error) {
	r, err := pool.Query(ctx, `SELECT position, amount_cents FROM core.payouts WHERE calcutta_id = $1 AND deleted_at IS NULL ORDER BY position ASC`, calcuttaID)
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

func loadCalcuttaEntriesAndBids(ctx context.Context, pool *pgxpool.Pool, calcuttaID string, calcuttaKey string) ([]bundles.EntryRecord, []bundles.EntryTeamBid, error) {
	r, err := pool.Query(ctx, `
		SELECT
			e.id,
			e.name,
			COALESCE(u.email, ''),
			COALESCE(u.first_name, ''),
			COALESCE(u.last_name, '')
		FROM core.entries e
		LEFT JOIN core.users u ON u.id = e.user_id
		WHERE e.calcutta_id = $1 AND e.deleted_at IS NULL
		ORDER BY e.created_at ASC
	`, calcuttaID)
	if err != nil {
		return nil, nil, err
	}
	defer r.Close()

	usedEntryKeys := make(map[string]int)
	entryKeyByLegacyID := make(map[string]string)

	entries := make([]bundles.EntryRecord, 0)
	for r.Next() {
		var entryID, name string
		var email, first, last string
		if err := r.Scan(&entryID, &name, &email, &first, &last); err != nil {
			return nil, nil, err
		}

		base := "entry-" + slugify(name)
		entryKey := uniquifyKey(base, usedEntryKeys)
		entryKey = fmt.Sprintf("%s:%s", calcuttaKey, entryKey)
		entryKeyByLegacyID[entryID] = entryKey

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

		entries = append(entries, bundles.EntryRecord{Key: entryKey, Name: name, UserName: userName, UserEmail: userEmail, LegacyID: entryID})
	}
	if err := r.Err(); err != nil {
		return nil, nil, err
	}

	bids, err := loadCalcuttaBids(ctx, pool, calcuttaID, entryKeyByLegacyID)
	if err != nil {
		return nil, nil, err
	}

	sort.Slice(entries, func(i, j int) bool { return entries[i].Key < entries[j].Key })
	return entries, bids, nil
}

func loadCalcuttaBids(ctx context.Context, pool *pgxpool.Pool, calcuttaID string, entryKeyByLegacyID map[string]string) ([]bundles.EntryTeamBid, error) {
	r, err := pool.Query(ctx, `
		SELECT
			et.id,
			et.entry_id,
			et.bid_points,
			s.slug
		FROM core.entry_teams et
		JOIN core.entries e ON e.id = et.entry_id
		JOIN core.teams t ON t.id = et.team_id
		JOIN core.schools s ON s.id = t.school_id
		WHERE e.calcutta_id = $1 AND et.deleted_at IS NULL AND t.deleted_at IS NULL AND s.deleted_at IS NULL AND e.deleted_at IS NULL
		ORDER BY e.created_at ASC, s.name ASC
	`, calcuttaID)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	out := make([]bundles.EntryTeamBid, 0)
	for r.Next() {
		var id, entryID, schoolSlug string
		var bid int
		if err := r.Scan(&id, &entryID, &bid, &schoolSlug); err != nil {
			return nil, err
		}
		entryKey := entryKeyByLegacyID[entryID]
		out = append(out, bundles.EntryTeamBid{EntryKey: entryKey, SchoolSlug: schoolSlug, Bid: bid, LegacyEntryTeamID: id})
	}
	if err := r.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func derefF(p *float64) float64 {
	if p == nil {
		return 0
	}
	return *p
}
