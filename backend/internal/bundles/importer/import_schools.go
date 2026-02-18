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
			SELECT $1, $2
			WHERE NOT EXISTS (SELECT 1 FROM core.schools WHERE slug = $1 AND deleted_at IS NULL)
			ON CONFLICT DO NOTHING
		`, s.Slug, s.Name)
		if err != nil {
			return 0, err
		}
		_, err = tx.Exec(ctx, `
			UPDATE core.schools SET name = $2, updated_at = NOW(), deleted_at = NULL
			WHERE slug = $1
		`, s.Slug, s.Name)
		if err != nil {
			return 0, err
		}
	}
	return len(b.Schools), nil
}
