package jobqueue

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Job represents a claimed job from the run_jobs queue.
type Job struct {
	RunID     string
	RunKey    string
	RunKind   string
	Params    json.RawMessage
	ClaimedAt time.Time
}

// Claimer claims and manages jobs from the derived.run_jobs queue.
type Claimer struct {
	pool *pgxpool.Pool
}

// NewClaimer creates a new Claimer.
func NewClaimer(pool *pgxpool.Pool) *Claimer {
	return &Claimer{pool: pool}
}

// ClaimNext atomically claims the highest-priority unclaimed job matching the
// given kinds. Jobs are ordered by priority ASC, created_at ASC. Stale running
// jobs (claimed longer ago than staleAfter with exponential backoff) are
// reclaimed. Jobs with a retry_after in the future are skipped.
func (c *Claimer) ClaimNext(ctx context.Context, kinds []string, workerID string, maxAttempts int, staleAfter time.Duration) (*Job, error) {
	now := time.Now().UTC()
	baseStaleSeconds := staleAfter.Seconds()
	if baseStaleSeconds <= 0 {
		baseStaleSeconds = 1800 // 30 min default
	}

	tx, err := c.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()

	// Fail stale jobs that exceeded max attempts
	if _, err := tx.Exec(ctx, `
		UPDATE derived.run_jobs
		SET status = 'failed',
			finished_at = NOW(),
			error_message = COALESCE(error_message, 'max_attempts_exceeded'),
			updated_at = NOW()
		WHERE run_kind = ANY($1::text[])
			AND status = 'running'
			AND claimed_at IS NOT NULL
			AND claimed_at < ($2::timestamptz - make_interval(secs => ($3 * POWER(2, GREATEST(attempt - 1, 0)))))
			AND attempt >= $4
	`, kinds, pgtype.Timestamptz{Time: now, Valid: true}, baseStaleSeconds, maxAttempts); err != nil {
		return nil, err
	}

	// Claim next job — priority ASC (lower = more urgent), then created_at ASC
	q := `
		WITH candidate AS (
			SELECT id
			FROM derived.run_jobs
			WHERE run_kind = ANY($1::text[])
				AND attempt < $4
				AND (retry_after IS NULL OR retry_after <= NOW())
				AND (
					status = 'queued'
					OR (
						status = 'running'
						AND claimed_at IS NOT NULL
						AND claimed_at < ($2::timestamptz - make_interval(secs => ($3 * POWER(2, GREATEST(attempt - 1, 0)))))
					)
				)
			ORDER BY priority ASC, created_at ASC
			LIMIT 1
			FOR UPDATE SKIP LOCKED
		)
		UPDATE derived.run_jobs j
		SET status = 'running',
			attempt = j.attempt + 1,
			claimed_at = $2,
			claimed_by = $5,
			started_at = COALESCE(j.started_at, $2),
			finished_at = NULL,
			error_message = NULL,
			updated_at = NOW()
		FROM candidate
		WHERE j.id = candidate.id
		RETURNING j.run_id::text, j.run_key::text, j.run_kind, j.params_json::text
	`

	var paramsStr string
	job := &Job{}
	if err := tx.QueryRow(ctx, q,
		kinds,
		pgtype.Timestamptz{Time: now, Valid: true},
		baseStaleSeconds,
		maxAttempts,
		workerID,
	).Scan(&job.RunID, &job.RunKey, &job.RunKind, &paramsStr); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	job.ClaimedAt = now
	job.Params = json.RawMessage([]byte(paramsStr))

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	committed = true

	return job, nil
}

// Succeed marks a job as succeeded.
func (c *Claimer) Succeed(ctx context.Context, kind string, runID string) error {
	_, err := c.pool.Exec(ctx, `
		UPDATE derived.run_jobs
		SET status = 'succeeded', finished_at = NOW(), error_message = NULL, updated_at = NOW()
		WHERE run_kind = $1 AND run_id = $2::uuid
	`, kind, runID)
	return err
}

// Fail marks a job as failed with an error message.
func (c *Claimer) Fail(ctx context.Context, kind string, runID string, errMsg string) error {
	_, err := c.pool.Exec(ctx, `
		UPDATE derived.run_jobs
		SET status = 'failed', finished_at = NOW(), error_message = $3, updated_at = NOW()
		WHERE run_kind = $1 AND run_id = $2::uuid
	`, kind, runID, errMsg)
	return err
}

// Requeue puts a job back into the queue with status='queued' and an optional
// retry_after delay. The attempt counter is decremented to undo the bump from
// claiming, since intentional requeues (e.g. waiting for dependencies) should
// not count against maxAttempts.
func (c *Claimer) Requeue(ctx context.Context, kind string, runID string, retryAfter time.Time) error {
	var retryPtr *pgtype.Timestamptz
	if !retryAfter.IsZero() {
		retryPtr = &pgtype.Timestamptz{Time: retryAfter, Valid: true}
	}
	_, err := c.pool.Exec(ctx, `
		UPDATE derived.run_jobs
		SET status = 'queued',
			attempt = GREATEST(attempt - 1, 0),
			claimed_at = NULL,
			claimed_by = NULL,
			retry_after = $3,
			updated_at = NOW()
		WHERE run_kind = $1 AND run_id = $2::uuid
	`, kind, runID, retryPtr)
	return err
}
