package db

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type IdempotencyRecord struct {
	Key            string
	UserID         string
	ResponseStatus int
	ResponseBody   json.RawMessage
}

type IdempotencyRepository struct {
	pool *pgxpool.Pool
}

func NewIdempotencyRepository(pool *pgxpool.Pool) *IdempotencyRepository {
	return &IdempotencyRepository{pool: pool}
}

// Get returns a cached response for the given key+user, or nil if not found.
func (r *IdempotencyRepository) Get(ctx context.Context, key, userID string) (*IdempotencyRecord, error) {
	var rec IdempotencyRecord
	err := r.pool.QueryRow(ctx,
		`SELECT key, user_id::text, response_status, response_body
		 FROM core.idempotency_keys
		 WHERE key = $1 AND user_id = $2 AND response_status IS NOT NULL`,
		key, userID,
	).Scan(&rec.Key, &rec.UserID, &rec.ResponseStatus, &rec.ResponseBody)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &rec, nil
}

// Reserve inserts a placeholder row so concurrent requests can detect duplication.
// Returns true if successfully reserved, false if the key already exists.
func (r *IdempotencyRepository) Reserve(ctx context.Context, key, userID string) (bool, error) {
	tag, err := r.pool.Exec(ctx,
		`INSERT INTO core.idempotency_keys (key, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		key, userID,
	)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() == 1, nil
}

// Complete stores the response for a previously reserved key.
func (r *IdempotencyRepository) Complete(ctx context.Context, key, userID string, status int, body json.RawMessage) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE core.idempotency_keys SET response_status = $3, response_body = $4 WHERE key = $1 AND user_id = $2`,
		key, userID, status, body,
	)
	return err
}
