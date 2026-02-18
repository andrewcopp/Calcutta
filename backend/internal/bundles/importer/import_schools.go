package importer

import (
	"context"
	"path/filepath"

	"github.com/andrewcopp/Calcutta/backend/internal/bundles"
	"github.com/jackc/pgx/v5"
)

func importSchools(ctx context.Context, tx pgx.Tx, inDir string) (int, error) {
	path := filepath.Join(inDir, "schools.json")
	var b bundles.SchoolsBundle
	if err := bundles.ReadJSON(path, &b); err != nil {
		return 0, err
	}
	for _, s := range b.Schools {
		_, err := tx.Exec(ctx, `
			INSERT INTO core.schools (slug, name)
			VALUES ($1, $2)
			ON CONFLICT (slug) WHERE deleted_at IS NULL
			DO UPDATE SET name = EXCLUDED.name, updated_at = NOW(), deleted_at = NULL
		`, s.Slug, s.Name)
		if err != nil {
			return 0, err
		}
	}
	return len(b.Schools), nil
}
