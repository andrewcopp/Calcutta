package db

import (
	"context"
	"errors"

	"github.com/andrewcopp/Calcutta/backend/internal/adapters/db/sqlc"
	"github.com/andrewcopp/Calcutta/backend/pkg/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SchoolRepository struct {
	pool *pgxpool.Pool
	q    *sqlc.Queries
}

func NewSchoolRepository(pool *pgxpool.Pool) *SchoolRepository {
	return &SchoolRepository{pool: pool, q: sqlc.New(pool)}
}

func (r *SchoolRepository) List(ctx context.Context) ([]models.School, error) {
	rows, err := r.q.ListSchools(ctx)
	if err != nil {
		return nil, err
	}

	schools := make([]models.School, 0, len(rows))
	for _, row := range rows {
		schools = append(schools, models.School{
			ID:      row.ID,
			Name:    row.Name,
			Created: row.CreatedAt.Time,
			Updated: row.UpdatedAt.Time,
		})
	}
	return schools, nil
}

func (r *SchoolRepository) GetByID(ctx context.Context, id string) (*models.School, error) {
	row, err := r.q.GetSchoolByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &models.School{
		ID:      row.ID,
		Name:    row.Name,
		Created: row.CreatedAt.Time,
		Updated: row.UpdatedAt.Time,
	}, nil
}
