CREATE TABLE core.idempotency_keys (
    key text NOT NULL,
    user_id uuid NOT NULL,
    created_at timestamptz DEFAULT now() NOT NULL,
    response_status integer,
    response_body jsonb,
    PRIMARY KEY (key, user_id)
);

-- Auto-expire old keys (index for cleanup queries)
CREATE INDEX idx_idempotency_keys_created ON core.idempotency_keys (created_at);
