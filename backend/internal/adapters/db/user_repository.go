package db

import (
	"context"
	"errors"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/adapters/db/sqlc"
	"github.com/andrewcopp/Calcutta/backend/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	pool *pgxpool.Pool
	q    *sqlc.Queries
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool, q: sqlc.New(pool)}
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	if user.Status == "" {
		return errors.New("user status is required")
	}

	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	return r.q.CreateUser(ctx, sqlc.CreateUserParams{
		ID:                 user.ID,
		Email:              user.Email,
		FirstName:          user.FirstName,
		LastName:           user.LastName,
		Status:             user.Status,
		PasswordHash:       user.PasswordHash,
		ExternalProvider:   user.ExternalProvider,
		ExternalProviderID: user.ExternalProviderID,
		CreatedAt:          pgtype.Timestamptz{Time: user.CreatedAt, Valid: true},
		UpdatedAt:          pgtype.Timestamptz{Time: user.UpdatedAt, Valid: true},
	})
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	row, err := r.q.GetUserByEmail(ctx, &email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return userFromRow(row.ID, row.Email, row.FirstName, row.LastName, row.Status, row.PasswordHash, row.ExternalProvider, row.ExternalProviderID, row.CreatedAt, row.UpdatedAt, row.DeletedAt), nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	row, err := r.q.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return userFromRow(row.ID, row.Email, row.FirstName, row.LastName, row.Status, row.PasswordHash, row.ExternalProvider, row.ExternalProviderID, row.CreatedAt, row.UpdatedAt, row.DeletedAt), nil
}

func (r *UserRepository) GetByExternalProvider(ctx context.Context, provider, providerID string) (*models.User, error) {
	row, err := r.q.GetUserByExternalProvider(ctx, sqlc.GetUserByExternalProviderParams{
		ExternalProvider:   &provider,
		ExternalProviderID: &providerID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return userFromRow(row.ID, row.Email, row.FirstName, row.LastName, row.Status, row.PasswordHash, row.ExternalProvider, row.ExternalProviderID, row.CreatedAt, row.UpdatedAt, row.DeletedAt), nil
}

func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	if user.Status == "" {
		return errors.New("user status is required")
	}

	now := time.Now()
	user.UpdatedAt = now

	var deletedAt pgtype.Timestamptz
	if user.DeletedAt != nil {
		deletedAt = pgtype.Timestamptz{Time: *user.DeletedAt, Valid: true}
	}

	return r.q.UpdateUser(ctx, sqlc.UpdateUserParams{
		ID:                 user.ID,
		Email:              user.Email,
		FirstName:          user.FirstName,
		LastName:           user.LastName,
		Status:             user.Status,
		PasswordHash:       user.PasswordHash,
		ExternalProvider:   user.ExternalProvider,
		ExternalProviderID: user.ExternalProviderID,
		UpdatedAt:          pgtype.Timestamptz{Time: user.UpdatedAt, Valid: true},
		DeletedAt:          deletedAt,
	})
}

func userFromRow(id string, email *string, firstName, lastName, status string, passwordHash, externalProvider, externalProviderID *string, createdAt, updatedAt, deletedAt pgtype.Timestamptz) *models.User {
	var deleted *time.Time
	if deletedAt.Valid {
		t := deletedAt.Time
		deleted = &t
	}
	return &models.User{
		ID:                 id,
		Email:              email,
		FirstName:          firstName,
		LastName:           lastName,
		Status:             status,
		PasswordHash:       passwordHash,
		ExternalProvider:   externalProvider,
		ExternalProviderID: externalProviderID,
		CreatedAt:          createdAt.Time,
		UpdatedAt:          updatedAt.Time,
		DeletedAt:          deleted,
	}
}
