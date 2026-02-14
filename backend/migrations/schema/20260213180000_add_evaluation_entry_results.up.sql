-- Add table to store per-entry results from evaluations
CREATE TABLE IF NOT EXISTS lab.evaluation_entry_results (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    evaluation_id UUID NOT NULL REFERENCES lab.evaluations(id),
    entry_name TEXT NOT NULL,
    mean_normalized_payout DOUBLE PRECISION,
    p_top1 DOUBLE PRECISION,
    p_in_money DOUBLE PRECISION,
    rank INTEGER,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(evaluation_id, entry_name)
);

CREATE INDEX idx_evaluation_entry_results_evaluation_id
ON lab.evaluation_entry_results(evaluation_id);
