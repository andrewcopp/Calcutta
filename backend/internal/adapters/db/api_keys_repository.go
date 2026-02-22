package db

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
	"github.com/andrewcopp/Calcutta/backend/internal/ports"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ ports.APIKeyReader = (*APIKeysRepository)(nil)

type APIKeysRepository struct {
	pool *pgxpool.Pool
}

func NewAPIKeysRepository(pool *pgxpool.Pool) *APIKeysRepository {
	return &APIKeysRepository{pool: pool}
}

func HashAPIKey(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func (r *APIKeysRepository) Create(ctx context.Context, userID, keyHash string, label *string, now time.Time) (*models.APIKey, error) {
	if now.IsZero() {
		now = time.Now().UTC()
	}

	var row models.APIKey
	row.UserID = userID
	row.Label = label

	err := r.pool.QueryRow(ctx, `
		INSERT INTO core.api_keys (user_id, key_hash, label, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`, userID, keyHash, label, now).Scan(&row.ID, &row.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *APIKeysRepository) GetActiveByHash(ctx context.Context, keyHash string, now time.Time) (*models.APIKey, error) {
	if now.IsZero() {
		now = time.Now().UTC()
	}

	var row models.APIKey
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, label, created_at, revoked_at, last_used_at
		FROM core.api_keys
		WHERE key_hash = $1
		  AND revoked_at IS NULL
	`, keyHash).Scan(&row.ID, &row.UserID, &row.Label, &row.CreatedAt, &row.RevokedAt, &row.LastUsedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	_, _ = r.pool.Exec(ctx, `
		UPDATE core.api_keys
		SET last_used_at = $2
		WHERE id = $1
	`, row.ID, now)

	row.LastUsedAt = &now
	return &row, nil
}

func (r *APIKeysRepository) ListByUser(ctx context.Context, userID string) ([]models.APIKey, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, label, created_at, revoked_at, last_used_at
		FROM core.api_keys
		WHERE user_id = $1
			AND revoked_at IS NULL
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.APIKey
	for rows.Next() {
		var rrow models.APIKey
		if err := rows.Scan(&rrow.ID, &rrow.UserID, &rrow.Label, &rrow.CreatedAt, &rrow.RevokedAt, &rrow.LastUsedAt); err != nil {
			return nil, err
		}
		out = append(out, rrow)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *APIKeysRepository) Revoke(ctx context.Context, id, userID string, now time.Time) error {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	_, err := r.pool.Exec(ctx, `
		UPDATE core.api_keys
		SET revoked_at = $3
		WHERE id = $1
		  AND user_id = $2
		  AND revoked_at IS NULL
	`, id, userID, now)
	return err
}
