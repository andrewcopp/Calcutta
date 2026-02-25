package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/andrewcopp/Calcutta/backend/internal/bundles"
)

func main() {
	if err := run(); err != nil {
		slog.Error("cmd_failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	inDir := flag.String("in", "./exports/bundles", "input bundles directory")
	outDir := flag.String("out", "./migrations/schema", "output migrations directory")
	startTS := flag.Int("start-ts", 20260225200000, "starting timestamp (YYYYMMDDHHMMSS)")
	flag.Parse()

	ts := *startTS

	// Generate schools seed migration
	schoolsPath := filepath.Join(*inDir, "schools.json")
	var schoolsBundle bundles.SchoolsBundle
	if err := bundles.ReadJSON(schoolsPath, &schoolsBundle); err != nil {
		return fmt.Errorf("reading schools.json: %w", err)
	}

	upSQL, downSQL := generateSchoolsMigration(schoolsBundle)
	if err := writeMigration(*outDir, ts, "seed_schools", upSQL, downSQL); err != nil {
		return err
	}
	slog.Info("generated", "migration", fmt.Sprintf("%d_seed_schools", ts), "schools", len(schoolsBundle.Schools))
	ts++

	// Generate tournament seed migrations
	tournamentPaths, err := filepath.Glob(filepath.Join(*inDir, "tournaments", "*.json"))
	if err != nil {
		return fmt.Errorf("globbing tournaments: %w", err)
	}
	sort.Strings(tournamentPaths)

	for _, path := range tournamentPaths {
		var tb bundles.TournamentBundle
		if err := bundles.ReadJSON(path, &tb); err != nil {
			return fmt.Errorf("reading %s: %w", path, err)
		}

		upSQL, downSQL := generateTournamentMigration(tb)

		// Extract year from import key for the migration name
		year := extractYear(tb.Tournament.ImportKey)
		name := fmt.Sprintf("seed_ncaa_tournament_%d", year)

		if err := writeMigration(*outDir, ts, name, upSQL, downSQL); err != nil {
			return err
		}
		slog.Info("generated", "migration", fmt.Sprintf("%d_%s", ts, name), "teams", len(tb.Teams))
		ts++
	}

	return nil
}

func extractYear(importKey string) int {
	var year int
	fmt.Sscanf(importKey, "ncaa-tournament-%d", &year)
	return year
}

func generateSchoolsMigration(b bundles.SchoolsBundle) (string, string) {
	var up strings.Builder
	up.WriteString("SET search_path = '';\n\n")
	up.WriteString("INSERT INTO core.schools (name, slug) VALUES\n")

	for i, s := range b.Schools {
		comma := ","
		if i == len(b.Schools)-1 {
			comma = ""
		}
		up.WriteString(fmt.Sprintf("  (%s, %s)%s\n", sqlQuote(s.Name), sqlQuote(s.Slug), comma))
	}
	up.WriteString("ON CONFLICT (slug) WHERE deleted_at IS NULL DO NOTHING;\n")

	down := "SET search_path = '';\n\n-- Schools are reference data; down migration is intentionally empty.\n-- Deleting schools would cascade-break teams in tournament seed migrations.\n"

	return up.String(), down
}

func generateTournamentMigration(b bundles.TournamentBundle) (string, string) {
	var up strings.Builder
	importKey := b.Tournament.ImportKey
	year := extractYear(importKey)

	up.WriteString("SET search_path = '';\n\n")

	// Season
	up.WriteString(fmt.Sprintf("INSERT INTO core.seasons (year) VALUES (%d)\nON CONFLICT (year) DO NOTHING;\n\n", year))

	// Tournament
	startingAt := "NULL"
	if b.Tournament.StartingAt != nil {
		startingAt = sqlQuote(b.Tournament.StartingAt.Format("2006-01-02T15:04:05-07:00"))
	}
	up.WriteString(fmt.Sprintf(`INSERT INTO core.tournaments (competition_id, season_id, import_key, rounds, starting_at,
  final_four_top_left, final_four_bottom_left, final_four_top_right, final_four_bottom_right)
SELECT c.id, s.id, %s, %d, %s,
  NULLIF(%s, ''), NULLIF(%s, ''), NULLIF(%s, ''), NULLIF(%s, '')
FROM core.competitions c, core.seasons s
WHERE c.name = 'NCAA Tournament' AND s.year = %d
ON CONFLICT (import_key) WHERE deleted_at IS NULL DO NOTHING;
`,
		sqlQuote(importKey),
		b.Tournament.Rounds,
		startingAt,
		sqlQuote(b.Tournament.FinalFourTopLeft),
		sqlQuote(b.Tournament.FinalFourBottomLeft),
		sqlQuote(b.Tournament.FinalFourTopRight),
		sqlQuote(b.Tournament.FinalFourBottomRight),
		year,
	))

	// Teams - bulk insert via VALUES + JOIN
	up.WriteString(fmt.Sprintf(`
INSERT INTO core.teams (tournament_id, school_id, seed, region, byes, wins, is_eliminated)
SELECT t.id, s.id, v.seed, v.region, v.byes, v.wins, v.is_eliminated
FROM (VALUES
`))

	for i, team := range b.Teams {
		comma := ","
		if i == len(b.Teams)-1 {
			comma = ""
		}
		up.WriteString(fmt.Sprintf("  (%s, %d, %s, %d, %d, %t)%s\n",
			sqlQuote(team.SchoolSlug), team.Seed, sqlQuote(team.Region),
			team.Byes, team.Wins, team.IsEliminated, comma))
	}

	up.WriteString(fmt.Sprintf(`) AS v(school_slug, seed, region, byes, wins, is_eliminated)
JOIN core.tournaments t ON t.import_key = %s AND t.deleted_at IS NULL
JOIN core.schools s ON s.slug = v.school_slug AND s.deleted_at IS NULL
ON CONFLICT (tournament_id, school_id) WHERE deleted_at IS NULL DO NOTHING;
`, sqlQuote(importKey)))

	// KenPom stats - only if any teams have kenpom data
	hasKenPom := false
	for _, team := range b.Teams {
		if team.KenPom != nil {
			hasKenPom = true
			break
		}
	}

	if hasKenPom {
		up.WriteString(fmt.Sprintf(`
INSERT INTO core.team_kenpom_stats (team_id, net_rtg, o_rtg, d_rtg, adj_t)
SELECT tm.id, v.net_rtg, v.o_rtg, v.d_rtg, v.adj_t
FROM (VALUES
`))

		first := true
		for _, team := range b.Teams {
			if team.KenPom == nil {
				continue
			}
			if !first {
				up.WriteString(",\n")
			}
			up.WriteString(fmt.Sprintf("  (%s, %s, %s, %s, %s)",
				sqlQuote(team.SchoolSlug),
				formatFloat(team.KenPom.NetRTG),
				formatFloat(team.KenPom.ORTG),
				formatFloat(team.KenPom.DRTG),
				formatFloat(team.KenPom.AdjT)))
			first = false
		}

		up.WriteString(fmt.Sprintf(`
) AS v(school_slug, net_rtg, o_rtg, d_rtg, adj_t)
JOIN core.tournaments t ON t.import_key = %s AND t.deleted_at IS NULL
JOIN core.schools s ON s.slug = v.school_slug AND s.deleted_at IS NULL
JOIN core.teams tm ON tm.tournament_id = t.id AND tm.school_id = s.id AND tm.deleted_at IS NULL
ON CONFLICT (team_id) DO NOTHING;
`, sqlQuote(importKey)))
	}

	// Down migration
	var down strings.Builder
	down.WriteString("SET search_path = '';\n\n")

	// Delete in reverse FK order
	down.WriteString(fmt.Sprintf(`-- Delete kenpom stats for this tournament's teams
DELETE FROM core.team_kenpom_stats
WHERE team_id IN (
  SELECT tm.id FROM core.teams tm
  JOIN core.tournaments t ON t.id = tm.tournament_id
  WHERE t.import_key = %s AND t.deleted_at IS NULL
);

`, sqlQuote(importKey)))

	down.WriteString(fmt.Sprintf(`-- Delete teams for this tournament
DELETE FROM core.teams
WHERE tournament_id IN (
  SELECT id FROM core.tournaments WHERE import_key = %s AND deleted_at IS NULL
);

`, sqlQuote(importKey)))

	down.WriteString(fmt.Sprintf(`-- Delete tournament
DELETE FROM core.tournaments WHERE import_key = %s;
`, sqlQuote(importKey)))

	return up.String(), down.String()
}

func writeMigration(outDir string, ts int, name, upSQL, downSQL string) error {
	upPath := filepath.Join(outDir, fmt.Sprintf("%d_%s.up.sql", ts, name))
	downPath := filepath.Join(outDir, fmt.Sprintf("%d_%s.down.sql", ts, name))

	if err := os.WriteFile(upPath, []byte(upSQL), 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", upPath, err)
	}
	if err := os.WriteFile(downPath, []byte(downSQL), 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", downPath, err)
	}
	return nil
}

func sqlQuote(s string) string {
	escaped := strings.ReplaceAll(s, "'", "''")
	return "'" + escaped + "'"
}

func formatFloat(f float64) string {
	return fmt.Sprintf("%g", f)
}
