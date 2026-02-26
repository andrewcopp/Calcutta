package importer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/andrewcopp/Calcutta/backend/internal/bundles"
	"github.com/jackc/pgx/v5"
)

func importPools(ctx context.Context, tx pgx.Tx, inDir string) (int, int, int, int, int, int, error) {
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
		if os.IsNotExist(err) {
			return 0, 0, 0, 0, 0, 0, nil
		}
		return 0, 0, 0, 0, 0, 0, err
	}
	sort.Strings(paths)

	portfolioCount := 0
	stubUserCount := 0
	investmentCount := 0
	payoutCount := 0
	roundCount := 0

	for _, path := range paths {
		var b bundles.PoolBundle
		if err := bundles.ReadJSON(path, &b); err != nil {
			return 0, 0, 0, 0, 0, 0, err
		}

		var tournamentID string
		err := tx.QueryRow(ctx, `
			SELECT id
			FROM core.tournaments
			WHERE import_key = $1 AND deleted_at IS NULL
		`, b.Tournament.ImportKey).Scan(&tournamentID)
		if err != nil {
			return 0, 0, 0, 0, 0, 0, fmt.Errorf("tournament import_key %q not found — have you run seed migrations? %w", b.Tournament.ImportKey, err)
		}

		ownerName := ownerDisplayName(b.Pool.Owner, b.Pool.Key)
		ownerID, err := createStubUser(ctx, tx, ownerName)
		if err != nil {
			return 0, 0, 0, 0, 0, 0, err
		}
		stubUserCount++

		var poolID string
		err = tx.QueryRow(ctx, `
			INSERT INTO core.pools (tournament_id, owner_id, created_by, name)
			VALUES ($1, $2, $2, $3)
			RETURNING id
		`, tournamentID, ownerID, b.Pool.Name).Scan(&poolID)
		if err != nil {
			return 0, 0, 0, 0, 0, 0, err
		}

		for _, r := range b.Rounds {
			_, err := tx.Exec(ctx, `
				INSERT INTO core.pool_scoring_rules (pool_id, win_index, points_awarded)
				VALUES ($1, $2, $3)
				ON CONFLICT (pool_id, win_index)
				DO UPDATE SET points_awarded = EXCLUDED.points_awarded, updated_at = NOW(), deleted_at = NULL
			`, poolID, r.Round, r.Points)
			if err != nil {
				return 0, 0, 0, 0, 0, 0, err
			}
			roundCount++
		}

		for _, p := range b.Payouts {
			_, err := tx.Exec(ctx, `
				INSERT INTO core.payouts (pool_id, position, amount_cents)
				VALUES ($1, $2, $3)
				ON CONFLICT (pool_id, position)
				DO UPDATE SET amount_cents = EXCLUDED.amount_cents, updated_at = NOW(), deleted_at = NULL
			`, poolID, p.Position, p.AmountCents)
			if err != nil {
				return 0, 0, 0, 0, 0, 0, err
			}
			payoutCount++
		}

		portfolioIDByKey := make(map[string]string, len(b.Portfolios))
		for _, pf := range b.Portfolios {
			// Determine display name: prefer UserName, fallback to portfolio Name
			displayName := pf.Name
			if pf.UserName != nil && strings.TrimSpace(*pf.UserName) != "" {
				displayName = *pf.UserName
			}

			portfolioUserID, err := createStubUser(ctx, tx, displayName)
			if err != nil {
				return 0, 0, 0, 0, 0, 0, err
			}
			stubUserCount++

			var portfolioID string
			err = tx.QueryRow(ctx, `
				INSERT INTO core.portfolios (name, user_id, pool_id)
				VALUES ($1, $2, $3)
				RETURNING id
			`, pf.Name, portfolioUserID, poolID).Scan(&portfolioID)
			if err != nil {
				return 0, 0, 0, 0, 0, 0, err
			}
			portfolioIDByKey[pf.Key] = portfolioID
			portfolioCount++
		}

		for _, inv := range b.Investments {
			portfolioID := portfolioIDByKey[inv.PortfolioKey]
			if portfolioID == "" {
				return 0, 0, 0, 0, 0, 0, fmt.Errorf("investment references unknown portfolio_key %s", inv.PortfolioKey)
			}

			var teamID string
			err := tx.QueryRow(ctx, `
				SELECT t.id
				FROM core.teams t
				JOIN core.schools s ON s.id = t.school_id
				WHERE t.tournament_id = $1 AND s.slug = $2 AND t.deleted_at IS NULL AND s.deleted_at IS NULL
			`, tournamentID, inv.SchoolSlug).Scan(&teamID)
			if err != nil {
				return 0, 0, 0, 0, 0, 0, fmt.Errorf("team not found for tournament %q school %q — have you run seed migrations? %w", b.Tournament.ImportKey, inv.SchoolSlug, err)
			}

			_, err = tx.Exec(ctx, `
				INSERT INTO core.investments (portfolio_id, team_id, credits)
				VALUES ($1, $2, $3)
			`, portfolioID, teamID, inv.Credits)
			if err != nil {
				return 0, 0, 0, 0, 0, 0, err
			}
			investmentCount++
		}
	}

	return len(paths), portfolioCount, stubUserCount, investmentCount, payoutCount, roundCount, nil
}

// createStubUser always creates a new stub user (never deduplicates).
func createStubUser(ctx context.Context, tx pgx.Tx, name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		name = "Unknown"
	}

	parts := strings.Fields(name)
	first := parts[0]
	last := ""
	if len(parts) > 1 {
		last = strings.Join(parts[1:], " ")
	}

	var id string
	err := tx.QueryRow(ctx, `
		INSERT INTO core.users (first_name, last_name, status, external_provider)
		VALUES ($1, $2, 'stub', 'historical')
		RETURNING id
	`, first, last).Scan(&id)
	return id, err
}

// ownerDisplayName extracts a display name from a UserRef, falling back to the pool key.
func ownerDisplayName(u *bundles.UserRef, fallbackKey string) string {
	if u == nil {
		return fallbackKey
	}
	first := ""
	last := ""
	if u.FirstName != nil {
		first = strings.TrimSpace(*u.FirstName)
	}
	if u.LastName != nil {
		last = strings.TrimSpace(*u.LastName)
	}
	fullName := strings.TrimSpace(first + " " + last)
	if fullName == "" {
		return fallbackKey
	}
	return fullName
}
