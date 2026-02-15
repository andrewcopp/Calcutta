ALTER TABLE IF EXISTS derived.candidates
    ADD COLUMN IF NOT EXISTS calcutta_id UUID REFERENCES core.calcuttas(id),
    ADD COLUMN IF NOT EXISTS tournament_id UUID REFERENCES core.tournaments(id),
    ADD COLUMN IF NOT EXISTS strategy_generation_run_id UUID REFERENCES derived.strategy_generation_runs(id),
    ADD COLUMN IF NOT EXISTS market_share_run_id UUID REFERENCES derived.market_share_runs(id),
    ADD COLUMN IF NOT EXISTS market_share_artifact_id UUID REFERENCES derived.run_artifacts(id),
    ADD COLUMN IF NOT EXISTS advancement_run_id UUID REFERENCES derived.game_outcome_runs(id),
    ADD COLUMN IF NOT EXISTS optimizer_key TEXT,
    ADD COLUMN IF NOT EXISTS starting_state_key TEXT,
    ADD COLUMN IF NOT EXISTS excluded_entry_name TEXT,
    ADD COLUMN IF NOT EXISTS git_sha TEXT;

CREATE INDEX IF NOT EXISTS idx_derived_candidates_calcutta_id
ON derived.candidates(calcutta_id)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_derived_candidates_tournament_id
ON derived.candidates(tournament_id)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_derived_candidates_strategy_generation_run_id
ON derived.candidates(strategy_generation_run_id)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_derived_candidates_market_share_artifact_id
ON derived.candidates(market_share_artifact_id)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_derived_candidates_advancement_run_id
ON derived.candidates(advancement_run_id)
WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_derived_candidates_lab_config
ON derived.candidates(
    calcutta_id,
    optimizer_key,
    market_share_artifact_id,
    advancement_run_id,
    starting_state_key,
    COALESCE(excluded_entry_name, ''::text)
)
WHERE deleted_at IS NULL
  AND calcutta_id IS NOT NULL
  AND optimizer_key IS NOT NULL
  AND market_share_artifact_id IS NOT NULL
  AND advancement_run_id IS NOT NULL
  AND starting_state_key IS NOT NULL;
