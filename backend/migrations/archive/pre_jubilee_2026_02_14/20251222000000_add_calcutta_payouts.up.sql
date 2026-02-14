CREATE TABLE calcutta_payouts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    calcutta_id UUID NOT NULL REFERENCES calcuttas(id),
    position INTEGER NOT NULL,
    amount_cents INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

ALTER TABLE calcutta_payouts
    ADD CONSTRAINT uq_calcutta_payouts_calcutta_position UNIQUE (calcutta_id, position);

CREATE INDEX idx_calcutta_payouts_calcutta_id ON calcutta_payouts(calcutta_id);
