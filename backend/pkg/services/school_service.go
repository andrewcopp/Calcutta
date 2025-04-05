package services

import (
	"context"
	"database/sql"

	"github.com/andrewcopp/Calcutta/backend/pkg/models"
)

// SchoolRepository handles database operations for schools
type SchoolRepository struct {
	db *sql.DB
}

// NewSchoolRepository creates a new SchoolRepository
func NewSchoolRepository(db *sql.DB) *SchoolRepository {
	return &SchoolRepository{db: db}
}

// GetAll returns all schools
func (r *SchoolRepository) GetAll(ctx context.Context) ([]models.School, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name
		FROM schools
		WHERE deleted_at IS NULL
		ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schools []models.School
	for rows.Next() {
		var school models.School
		if err := rows.Scan(&school.ID, &school.Name); err != nil {
			return nil, err
		}
		schools = append(schools, school)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return schools, nil
}

// GetByID returns a school by ID
func (r *SchoolRepository) GetByID(ctx context.Context, id string) (models.School, error) {
	var school models.School
	err := r.db.QueryRowContext(ctx, `
		SELECT id, name
		FROM schools
		WHERE id = $1 AND deleted_at IS NULL
	`, id).Scan(&school.ID, &school.Name)

	if err == sql.ErrNoRows {
		return models.School{}, nil
	}
	if err != nil {
		return models.School{}, err
	}

	return school, nil
}
