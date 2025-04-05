-- Create UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) NOT NULL UNIQUE,
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Schools table
CREATE TABLE schools (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Tournaments table
CREATE TABLE tournaments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    rounds INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Tournament Teams table
CREATE TABLE tournament_teams (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tournament_id UUID NOT NULL REFERENCES tournaments(id),
    school_id UUID NOT NULL REFERENCES schools(id),
    seed INTEGER NOT NULL,
    region VARCHAR(50) NOT NULL,
    byes INTEGER NOT NULL DEFAULT 0,
    wins INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Calcuttas table
CREATE TABLE calcuttas (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tournament_id UUID NOT NULL REFERENCES tournaments(id),
    owner_id UUID NOT NULL REFERENCES users(id),
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Calcutta Rounds table
CREATE TABLE calcutta_rounds (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    calcutta_id UUID NOT NULL REFERENCES calcuttas(id),
    round INTEGER NOT NULL,
    points INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Calcutta Entries table
CREATE TABLE calcutta_entries (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    user_id UUID REFERENCES users(id),
    calcutta_id UUID NOT NULL REFERENCES calcuttas(id),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Calcutta Entry Teams table
CREATE TABLE calcutta_entry_teams (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    entry_id UUID NOT NULL REFERENCES calcutta_entries(id),
    team_id UUID NOT NULL REFERENCES tournament_teams(id),
    bid DECIMAL(10,2) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Calcutta Portfolios table
CREATE TABLE calcutta_portfolios (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    entry_id UUID NOT NULL REFERENCES calcutta_entries(id),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Calcutta Portfolio Teams table
CREATE TABLE calcutta_portfolio_teams (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    portfolio_id UUID NOT NULL REFERENCES calcutta_portfolios(id),
    team_id UUID NOT NULL REFERENCES tournament_teams(id),
    ownership_percentage DECIMAL(5,4) NOT NULL,
    points_earned DECIMAL(10,2) NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes
CREATE INDEX idx_tournament_teams_tournament_id ON tournament_teams(tournament_id);
CREATE INDEX idx_calcutta_rounds_calcutta_id ON calcutta_rounds(calcutta_id);
CREATE INDEX idx_calcutta_entries_calcutta_id ON calcutta_entries(calcutta_id);
CREATE INDEX idx_calcutta_entries_user_id ON calcutta_entries(user_id);
CREATE INDEX idx_calcutta_entry_teams_entry_id ON calcutta_entry_teams(entry_id);
CREATE INDEX idx_calcutta_entry_teams_team_id ON calcutta_entry_teams(team_id);
CREATE INDEX idx_calcutta_portfolios_entry_id ON calcutta_portfolios(entry_id);
CREATE INDEX idx_calcutta_portfolio_teams_portfolio_id ON calcutta_portfolio_teams(portfolio_id);
CREATE INDEX idx_calcutta_portfolio_teams_team_id ON calcutta_portfolio_teams(team_id); 