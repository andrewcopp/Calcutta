-- Create prediction_batches table (parent table for prediction runs)
CREATE TABLE derived.prediction_batches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tournament_id UUID NOT NULL REFERENCES core.tournaments(id),
    probability_source_key TEXT NOT NULL DEFAULT 'kenpom',
    game_outcome_spec_json JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_prediction_batches_tournament_id ON derived.prediction_batches(tournament_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_prediction_batches_created_at ON derived.prediction_batches(created_at DESC) WHERE deleted_at IS NULL;

COMMENT ON TABLE derived.prediction_batches IS 'Stores metadata for prediction generation runs (analogous to simulated_tournaments for simulations)';
COMMENT ON COLUMN derived.prediction_batches.probability_source_key IS 'Identifier for the probability model used (e.g., kenpom)';
COMMENT ON COLUMN derived.prediction_batches.game_outcome_spec_json IS 'Parameters for the game outcome model (e.g., {"kind": "kenpom", "sigma": 10.0})';

-- Create predicted_team_values table (child table with per-team predictions)
CREATE TABLE derived.predicted_team_values (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    prediction_batch_id UUID NOT NULL REFERENCES derived.prediction_batches(id),
    tournament_id UUID NOT NULL REFERENCES core.tournaments(id),
    team_id UUID NOT NULL REFERENCES core.teams(id),
    expected_points FLOAT NOT NULL,
    variance_points FLOAT,
    std_points FLOAT,
    p_round_1 FLOAT,
    p_round_2 FLOAT,
    p_round_3 FLOAT,
    p_round_4 FLOAT,
    p_round_5 FLOAT,
    p_round_6 FLOAT,
    p_round_7 FLOAT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_predicted_team_values_batch_id ON derived.predicted_team_values(prediction_batch_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_predicted_team_values_tournament_id ON derived.predicted_team_values(tournament_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_predicted_team_values_team_id ON derived.predicted_team_values(team_id) WHERE deleted_at IS NULL;

COMMENT ON TABLE derived.predicted_team_values IS 'Stores predicted expected points and round probabilities for each team (analogous to simulated_teams for simulations)';
COMMENT ON COLUMN derived.predicted_team_values.expected_points IS 'Expected tournament points for 100% ownership of this team';
COMMENT ON COLUMN derived.predicted_team_values.p_round_1 IS 'Probability of winning first game (First Four or Round of 64)';
COMMENT ON COLUMN derived.predicted_team_values.p_round_2 IS 'Probability of reaching Round of 32';
COMMENT ON COLUMN derived.predicted_team_values.p_round_3 IS 'Probability of reaching Sweet 16';
COMMENT ON COLUMN derived.predicted_team_values.p_round_4 IS 'Probability of reaching Elite 8';
COMMENT ON COLUMN derived.predicted_team_values.p_round_5 IS 'Probability of reaching Final Four';
COMMENT ON COLUMN derived.predicted_team_values.p_round_6 IS 'Probability of reaching Championship game';
COMMENT ON COLUMN derived.predicted_team_values.p_round_7 IS 'Probability of winning Championship';
