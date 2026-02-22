package db

import (
	"context"
	"errors"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/adapters/db/sqlc"
	"github.com/andrewcopp/Calcutta/backend/internal/models"
	"github.com/andrewcopp/Calcutta/backend/internal/ports"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ ports.AuthSessionRepository = (*AuthRepository)(nil)

type AuthRepository struct {
	pool *pgxpool.Pool
	q    *sqlc.Queries
}

func NewAuthRepository(pool *pgxpool.Pool) *AuthRepository {
	return &AuthRepository{pool: pool, q: sqlc.New(pool)}
}

func (r *AuthRepository) CreateSession(ctx context.Context, userID, refreshTokenHash, userAgent, ipAddress string, expiresAt time.Time) (string, error) {
	arg := sqlc.CreateAuthSessionParams{
		UserID:           userID,
		RefreshTokenHash: refreshTokenHash,
		ExpiresAt:        pgtype.Timestamptz{Time: expiresAt, Valid: true},
	}
	if userAgent != "" {
		ua := userAgent
		arg.UserAgent = &ua
	}
	if ipAddress != "" {
		ip := ipAddress
		arg.IpAddress = &ip
	}
	return r.q.CreateAuthSession(ctx, arg)
}

func (r *AuthRepository) GetSessionByID(ctx context.Context, id string) (*models.AuthSession, error) {
	row, err := r.q.GetAuthSessionByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return authSessionFromRow(row.ID, row.UserID, row.RefreshTokenHash, row.ExpiresAt, row.RevokedAt), nil
}

func (r *AuthRepository) GetSessionByRefreshTokenHash(ctx context.Context, refreshTokenHash string) (*models.AuthSession, error) {
	row, err := r.q.GetAuthSessionByRefreshTokenHash(ctx, refreshTokenHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return authSessionFromRow(row.ID, row.UserID, row.RefreshTokenHash, row.ExpiresAt, row.RevokedAt), nil
}

func (r *AuthRepository) RotateRefreshToken(ctx context.Context, sessionID, newRefreshTokenHash string, newExpiresAt time.Time) error {
	return r.q.RotateAuthSessionRefreshToken(ctx, sqlc.RotateAuthSessionRefreshTokenParams{
		ID:               sessionID,
		RefreshTokenHash: newRefreshTokenHash,
		ExpiresAt:        pgtype.Timestamptz{Time: newExpiresAt, Valid: true},
	})
}

func (r *AuthRepository) RevokeSession(ctx context.Context, sessionID string) error {
	return r.q.RevokeAuthSession(ctx, sessionID)
}

func (r *AuthRepository) IsUserActive(ctx context.Context, userID string) (bool, error) {
	if userID == "" {
		return false, nil
	}

	var status string
	err := r.pool.QueryRow(ctx, `
		SELECT status
		FROM core.users
		WHERE id = $1 AND deleted_at IS NULL
	`, userID).Scan(&status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return status == "active", nil
}

func authSessionFromRow(id, userID, refreshTokenHash string, expiresAt, revokedAt pgtype.Timestamptz) *models.AuthSession {
	var revoked *time.Time
	if revokedAt.Valid {
		t := revokedAt.Time
		revoked = &t
	}
	return &models.AuthSession{
		ID:               id,
		UserID:           userID,
		RefreshTokenHash: refreshTokenHash,
		ExpiresAt:        expiresAt.Time,
		RevokedAt:        revoked,
	}
}
