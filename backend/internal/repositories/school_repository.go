package repositories

import (
	"calcutta/internal/models"
	"context"
	"database/sql"
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
