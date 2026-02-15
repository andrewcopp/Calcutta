CREATE TABLE tournament_team_kenpom_stats (
    tournament_team_id UUID PRIMARY KEY REFERENCES tournament_teams(id) ON DELETE CASCADE,
    net_rtg DOUBLE PRECISION,
    o_rtg DOUBLE PRECISION,
    d_rtg DOUBLE PRECISION,
    adj_t DOUBLE PRECISION,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_tournament_team_kenpom_stats_team_id ON tournament_team_kenpom_stats(tournament_team_id);
