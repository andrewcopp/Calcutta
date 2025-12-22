package main

import (
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/andrewcopp/Calcutta/backend/pkg/common"
	"github.com/andrewcopp/Calcutta/backend/pkg/models"
	"github.com/andrewcopp/Calcutta/backend/pkg/services"
)

type kenPomRow struct {
	Team   string
	NetRtg float64
	ORtg   float64
	DRtg   float64
	AdjT   float64
}

func buildKenPomLookup(rows []kenPomRow) map[string]kenPomRow {
	byName := make(map[string]kenPomRow, len(rows)*2)
	for _, r := range rows {
		parsed := normalizeKenPomTeamName(r.Team)
		standard := common.GetStandardizedSchoolName(parsed)

		if _, ok := byName[parsed]; !ok {
			byName[parsed] = r
		}
		if _, ok := byName[standard]; !ok {
			byName[standard] = r
		}
		if _, ok := byName[r.Team]; !ok {
			byName[r.Team] = r
		}
	}
	return byName
}

func normalizeOverridesForSchoolLookup(m map[string]string) map[string]string {
	if m == nil {
		return map[string]string{}
	}

	out := make(map[string]string, len(m)*2)
	for k, v := range m {
		out[k] = v
	}
	for k, v := range m {
		if _, ok := out[v]; !ok {
			out[v] = k
		}
	}
	return out
}

var seedSuffixRe = regexp.MustCompile(`\s+(?:1[0-6]|[1-9])$`)

func main() {
	var (
		dataDir      = flag.String("data-dir", "../../../data/kenpom", "Directory containing KenPom CSVs named like {year}.csv")
		year         = flag.String("year", "", "Only import a single year (e.g. 2024)")
		dryRun       = flag.Bool("dry-run", true, "If true, do not write to database")
		overrides    = flag.String("overrides", "", "Optional path to a JSON file containing a name mapping (recommended: school name -> KenPom team name)")
		unmatchedOut = flag.String("unmatched-out", "", "Optional output path for unmatched report (defaults to {data-dir}/unmatched_{year}.csv)")
	)
	flag.Parse()
	if *dryRun {
		log.Printf("Running in dry-run mode (no database writes). Pass --dry-run=false to persist stats.")
	}

	connString := os.Getenv("DATABASE_URL")
	if connString == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	db, err := sql.Open("pgx", connString)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database connection: %v", err)
		}
	}()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	overrideMap, err := readOverrides(*overrides)
	if err != nil {
		log.Fatalf("Error reading overrides: %v", err)
	}
	nameOverrides := normalizeOverridesForSchoolLookup(overrideMap)

	files, err := kenPomCSVFiles(*dataDir, *year)
	if err != nil {
		log.Fatalf("Error locating KenPom CSVs: %v", err)
	}
	if len(files) == 0 {
		log.Fatalf("No KenPom CSV files found in %s", *dataDir)
	}

	repo := services.NewTournamentRepository(db)

	for _, file := range files {
		y, err := yearFromFilename(file)
		if err != nil {
			log.Printf("Skipping %s: %v", file, err)
			continue
		}

		log.Printf("Importing KenPom CSV for year %s: %s", y, file)

		tournamentID, err := findTournamentIDByYear(context.Background(), db, y)
		if err != nil {
			log.Fatalf("Error finding tournament for year %s: %v", y, err)
		}

		tournamentTeams, err := repo.GetTeams(context.Background(), tournamentID)
		if err != nil {
			log.Fatalf("Error loading tournament teams for year %s: %v", y, err)
		}

		rows, err := readKenPomCSV(file)
		if err != nil {
			log.Fatalf("Error reading KenPom CSV %s: %v", file, err)
		}
		kenpomByName := buildKenPomLookup(rows)

		var unmatched []unmatchedRow
		matched := 0
		considered := 0
		upserted := 0
		for _, tt := range tournamentTeams {
			if tt == nil || tt.School == nil {
				continue
			}
			considered++
			schoolName := tt.School.Name
			schoolNameStandard := common.GetStandardizedSchoolName(schoolName)

			lookupName := common.GetKenPomTeamName(schoolName)
			if v, ok := nameOverrides[schoolName]; ok {
				lookupName = v
			}
			lookupName = normalizeKenPomTeamName(lookupName)

			row, ok := kenpomByName[schoolName]
			if !ok {
				row, ok = kenpomByName[schoolNameStandard]
			}
			if !ok {
				row, ok = kenpomByName[common.GetStandardizedSchoolName(lookupName)]
			}
			if !ok {
				row, ok = kenpomByName[lookupName]
			}
			if !ok {
				unmatched = append(unmatched, unmatchedRow{
					Year:                 y,
					TournamentSchoolName: schoolName,
					KenPomLookupName:     lookupName,
				})
				continue
			}

			matched++
			if *dryRun {
				continue
			}

			stats := &models.KenPomStats{
				NetRtg: floatPtr(row.NetRtg),
				ORtg:   floatPtr(row.ORtg),
				DRtg:   floatPtr(row.DRtg),
				AdjT:   floatPtr(row.AdjT),
			}
			if err := repo.UpsertTournamentTeamKenPomStats(context.Background(), tt.ID, stats); err != nil {
				log.Fatalf("Error upserting KenPom stats for year %s, school %s: %v", y, schoolName, err)
			}
			upserted++
		}

		if *dryRun {
			log.Printf("Year %s: matched %d/%d tournament teams (dry-run: 0 upserted)", y, matched, considered)
		} else {
			log.Printf("Year %s: matched %d/%d tournament teams (%d upserted)", y, matched, considered, upserted)
		}
		if len(unmatched) > 0 {
			outPath := *unmatchedOut
			if outPath == "" {
				outPath = filepath.Join(*dataDir, fmt.Sprintf("unmatched_%s.csv", y))
			}
			if err := writeUnmatchedCSV(outPath, unmatched); err != nil {
				log.Fatalf("Error writing unmatched report: %v", err)
			}
			log.Printf("Year %s: wrote unmatched report to %s (%d rows)", y, outPath, len(unmatched))
		}
	}
}

func kenPomCSVFiles(dir, onlyYear string) ([]string, error) {
	if onlyYear != "" {
		path := filepath.Join(dir, fmt.Sprintf("%s.csv", onlyYear))
		if _, err := os.Stat(path); err != nil {
			return nil, err
		}
		return []string{path}, nil
	}

	matches, err := filepath.Glob(filepath.Join(dir, "*.csv"))
	if err != nil {
		return nil, err
	}
	sort.Strings(matches)
	return matches, nil
}

func yearFromFilename(path string) (string, error) {
	base := filepath.Base(path)
	if !strings.HasSuffix(base, ".csv") {
		return "", errors.New("not a csv")
	}
	year := strings.TrimSuffix(base, ".csv")
	if len(year) != 4 {
		return "", fmt.Errorf("expected {year}.csv, got %s", base)
	}
	if _, err := strconv.Atoi(year); err != nil {
		return "", fmt.Errorf("invalid year %s", year)
	}
	return year, nil
}

func normalizeKenPomTeamName(team string) string {
	team = strings.TrimSpace(team)
	team = strings.TrimSuffix(team, "*")
	team = seedSuffixRe.ReplaceAllString(team, "")
	team = strings.TrimSpace(team)
	return team
}

func readKenPomCSV(path string) ([]kenPomRow, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.FieldsPerRecord = -1

	header, err := r.Read()
	if err != nil {
		return nil, err
	}

	idx := make(map[string]int, len(header))
	for i, h := range header {
		idx[strings.TrimSpace(h)] = i
	}
	required := []string{"Team", "NetRtg", "ORtg", "DRtg", "AdjT"}
	for _, k := range required {
		if _, ok := idx[k]; !ok {
			return nil, fmt.Errorf("missing required column %s", k)
		}
	}

	rows := make([]kenPomRow, 0, 400)
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		team := strings.TrimSpace(rec[idx["Team"]])
		if team == "" {
			continue
		}

		netRtg, err := parseFloat(rec[idx["NetRtg"]])
		if err != nil {
			return nil, fmt.Errorf("team %s: NetRtg: %w", team, err)
		}
		oRtg, err := parseFloat(rec[idx["ORtg"]])
		if err != nil {
			return nil, fmt.Errorf("team %s: ORtg: %w", team, err)
		}
		dRtg, err := parseFloat(rec[idx["DRtg"]])
		if err != nil {
			return nil, fmt.Errorf("team %s: DRtg: %w", team, err)
		}
		adjT, err := parseFloat(rec[idx["AdjT"]])
		if err != nil {
			return nil, fmt.Errorf("team %s: AdjT: %w", team, err)
		}

		rows = append(rows, kenPomRow{
			Team:   team,
			NetRtg: netRtg,
			ORtg:   oRtg,
			DRtg:   dRtg,
			AdjT:   adjT,
		})
	}

	return rows, nil
}

type unmatchedRow struct {
	Year                 string
	TournamentSchoolName string
	KenPomLookupName     string
}

func writeUnmatchedCSV(path string, rows []unmatchedRow) error {
	if len(rows) == 0 {
		return nil
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	if err := w.Write([]string{"Year", "TournamentSchoolName", "KenPomLookupName"}); err != nil {
		return err
	}
	for _, r := range rows {
		if err := w.Write([]string{r.Year, r.TournamentSchoolName, r.KenPomLookupName}); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}

func parseFloat(s string) (float64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, errors.New("empty")
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}
	return v, nil
}

func floatPtr(v float64) *float64 {
	vv := v
	return &vv
}

func readOverrides(path string) (map[string]string, error) {
	if path == "" {
		return map[string]string{}, nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m map[string]string
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	if m == nil {
		m = map[string]string{}
	}
	return m, nil
}

func findTournamentIDByYear(ctx context.Context, db *sql.DB, year string) (string, error) {
	name := fmt.Sprintf("NCAA Tournament %s", year)
	var id string
	err := db.QueryRowContext(ctx, `
		SELECT id
		FROM tournaments
		WHERE name = $1 AND deleted_at IS NULL
		LIMIT 1
	`, name).Scan(&id)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("tournament not found for name %q", name)
	}
	if err != nil {
		return "", err
	}
	if id == "" {
		return "", errors.New("tournament id empty")
	}
	return id, nil
}
