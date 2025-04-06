CREATE TABLE tournament_games (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tournament_id UUID NOT NULL REFERENCES tournaments(id),
    team1_id UUID REFERENCES tournament_teams(id),
    team2_id UUID REFERENCES tournament_teams(id),
    tipoff_time TIMESTAMP WITH TIME ZONE,
    sort_order INTEGER NOT NULL,
    team1_score INTEGER,
    team2_score INTEGER,
    next_game_id UUID REFERENCES tournament_games(id),
    next_game_slot INTEGER CHECK (next_game_slot IN (1, 2)),
    is_final BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT valid_next_game CHECK (
        (next_game_id IS NULL AND next_game_slot IS NULL) OR
        (next_game_id IS NOT NULL AND next_game_slot IS NOT NULL)
    )
); 