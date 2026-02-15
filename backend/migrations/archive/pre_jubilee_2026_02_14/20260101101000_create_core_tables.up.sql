-- Core identity tables
CREATE TABLE core.seasons (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    year INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT uq_core_seasons_year UNIQUE (year)
);

CREATE TABLE core.competitions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT uq_core_competitions_name UNIQUE (name)
);

CREATE TABLE core.tournaments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    competition_id UUID NOT NULL REFERENCES core.competitions(id),
    season_id UUID NOT NULL REFERENCES core.seasons(id),
    name VARCHAR(255) NOT NULL,
    import_key TEXT NOT NULL,
    rounds INTEGER NOT NULL,
    starting_at TIMESTAMP WITH TIME ZONE,
    final_four_top_left VARCHAR(50),
    final_four_bottom_left VARCHAR(50),
    final_four_top_right VARCHAR(50),
    final_four_bottom_right VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE UNIQUE INDEX uq_core_tournaments_import_key ON core.tournaments (import_key)
WHERE deleted_at IS NULL;

CREATE INDEX idx_core_tournaments_season_id ON core.tournaments (season_id);
CREATE INDEX idx_core_tournaments_competition_id ON core.tournaments (competition_id);

-- Core gameplay tables
CREATE TABLE core.schools (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    slug TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE UNIQUE INDEX uq_core_schools_name ON core.schools (name)
WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX uq_core_schools_slug ON core.schools (slug)
WHERE deleted_at IS NULL;

CREATE TABLE core.teams (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tournament_id UUID NOT NULL REFERENCES core.tournaments(id),
    school_id UUID NOT NULL REFERENCES core.schools(id),
    seed INTEGER NOT NULL,
    region VARCHAR(50) NOT NULL,
    byes INTEGER NOT NULL DEFAULT 0,
    wins INTEGER NOT NULL DEFAULT 0,
    eliminated BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_core_teams_tournament_id ON core.teams (tournament_id);
CREATE INDEX idx_core_teams_school_id ON core.teams (school_id);

CREATE TABLE core.team_kenpom_stats (
    team_id UUID PRIMARY KEY REFERENCES core.teams(id) ON DELETE CASCADE,
    net_rtg DOUBLE PRECISION,
    o_rtg DOUBLE PRECISION,
    d_rtg DOUBLE PRECISION,
    adj_t DOUBLE PRECISION,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_core_team_kenpom_stats_team_id ON core.team_kenpom_stats(team_id);

CREATE TABLE core.calcuttas (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tournament_id UUID NOT NULL REFERENCES core.tournaments(id),
    owner_id UUID NOT NULL REFERENCES public.users(id),
    name VARCHAR(255) NOT NULL,
    min_teams INTEGER NOT NULL DEFAULT 3,
    max_teams INTEGER NOT NULL DEFAULT 10,
    max_bid INTEGER NOT NULL DEFAULT 50,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_core_calcuttas_tournament_id ON core.calcuttas (tournament_id);
CREATE INDEX idx_core_calcuttas_owner_id ON core.calcuttas (owner_id);

CREATE TABLE core.calcutta_scoring_rules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    calcutta_id UUID NOT NULL REFERENCES core.calcuttas(id) ON DELETE CASCADE,
    win_index INTEGER NOT NULL,
    points_awarded INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT uq_core_calcutta_scoring_rules UNIQUE (calcutta_id, win_index)
);

CREATE INDEX idx_core_calcutta_scoring_rules_calcutta_id ON core.calcutta_scoring_rules (calcutta_id);

CREATE TABLE core.entries (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    user_id UUID REFERENCES public.users(id),
    calcutta_id UUID NOT NULL REFERENCES core.calcuttas(id),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_core_entries_calcutta_id ON core.entries (calcutta_id);
CREATE INDEX idx_core_entries_user_id ON core.entries (user_id);

CREATE TABLE core.entry_teams (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    entry_id UUID NOT NULL REFERENCES core.entries(id) ON DELETE CASCADE,
    team_id UUID NOT NULL REFERENCES core.teams(id),
    bid_points INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_core_entry_teams_entry_id ON core.entry_teams(entry_id);
CREATE INDEX idx_core_entry_teams_team_id ON core.entry_teams(team_id);

CREATE TABLE core.payouts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    calcutta_id UUID NOT NULL REFERENCES core.calcuttas(id) ON DELETE CASCADE,
    position INTEGER NOT NULL,
    amount_cents INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT uq_core_payouts_calcutta_position UNIQUE (calcutta_id, position)
);

CREATE INDEX idx_core_payouts_calcutta_id ON core.payouts (calcutta_id);

-- updated_at triggers
DROP TRIGGER IF EXISTS trg_core_seasons_updated_at ON core.seasons;
CREATE TRIGGER trg_core_seasons_updated_at
BEFORE UPDATE ON core.seasons
FOR EACH ROW
EXECUTE FUNCTION core.set_updated_at();

DROP TRIGGER IF EXISTS trg_core_competitions_updated_at ON core.competitions;
CREATE TRIGGER trg_core_competitions_updated_at
BEFORE UPDATE ON core.competitions
FOR EACH ROW
EXECUTE FUNCTION core.set_updated_at();

DROP TRIGGER IF EXISTS trg_core_tournaments_updated_at ON core.tournaments;
CREATE TRIGGER trg_core_tournaments_updated_at
BEFORE UPDATE ON core.tournaments
FOR EACH ROW
EXECUTE FUNCTION core.set_updated_at();

DROP TRIGGER IF EXISTS trg_core_schools_updated_at ON core.schools;
CREATE TRIGGER trg_core_schools_updated_at
BEFORE UPDATE ON core.schools
FOR EACH ROW
EXECUTE FUNCTION core.set_updated_at();

DROP TRIGGER IF EXISTS trg_core_teams_updated_at ON core.teams;
CREATE TRIGGER trg_core_teams_updated_at
BEFORE UPDATE ON core.teams
FOR EACH ROW
EXECUTE FUNCTION core.set_updated_at();

DROP TRIGGER IF EXISTS trg_core_team_kenpom_stats_updated_at ON core.team_kenpom_stats;
CREATE TRIGGER trg_core_team_kenpom_stats_updated_at
BEFORE UPDATE ON core.team_kenpom_stats
FOR EACH ROW
EXECUTE FUNCTION core.set_updated_at();

DROP TRIGGER IF EXISTS trg_core_calcuttas_updated_at ON core.calcuttas;
CREATE TRIGGER trg_core_calcuttas_updated_at
BEFORE UPDATE ON core.calcuttas
FOR EACH ROW
EXECUTE FUNCTION core.set_updated_at();

DROP TRIGGER IF EXISTS trg_core_calcutta_scoring_rules_updated_at ON core.calcutta_scoring_rules;
CREATE TRIGGER trg_core_calcutta_scoring_rules_updated_at
BEFORE UPDATE ON core.calcutta_scoring_rules
FOR EACH ROW
EXECUTE FUNCTION core.set_updated_at();

DROP TRIGGER IF EXISTS trg_core_entries_updated_at ON core.entries;
CREATE TRIGGER trg_core_entries_updated_at
BEFORE UPDATE ON core.entries
FOR EACH ROW
EXECUTE FUNCTION core.set_updated_at();

DROP TRIGGER IF EXISTS trg_core_entry_teams_updated_at ON core.entry_teams;
CREATE TRIGGER trg_core_entry_teams_updated_at
BEFORE UPDATE ON core.entry_teams
FOR EACH ROW
EXECUTE FUNCTION core.set_updated_at();

DROP TRIGGER IF EXISTS trg_core_payouts_updated_at ON core.payouts;
CREATE TRIGGER trg_core_payouts_updated_at
BEFORE UPDATE ON core.payouts
FOR EACH ROW
EXECUTE FUNCTION core.set_updated_at();
