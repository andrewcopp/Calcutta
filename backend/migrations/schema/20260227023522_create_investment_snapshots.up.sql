-- Migration: create_investment_snapshots
-- Created: 2026-02-27 02:35:22 UTC

SET search_path = '';
SET lock_timeout = '5s';
SET statement_timeout = '30s';

CREATE TABLE IF NOT EXISTS core.investment_snapshots (
    id UUID PRIMARY KEY DEFAULT public.uuid_generate_v4(),
    portfolio_id UUID NOT NULL,
    changed_by UUID NOT NULL,
    reason TEXT NOT NULL DEFAULT '',
    investments JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- updated_at trigger
CREATE TRIGGER trg_core_investment_snapshots_updated_at
    BEFORE UPDATE ON core.investment_snapshots
    FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();

-- Foreign key constraints
ALTER TABLE core.investment_snapshots
    ADD CONSTRAINT investment_snapshots_portfolio_id_fkey
    FOREIGN KEY (portfolio_id) REFERENCES core.portfolios(id);

ALTER TABLE core.investment_snapshots
    ADD CONSTRAINT investment_snapshots_changed_by_fkey
    FOREIGN KEY (changed_by) REFERENCES core.users(id);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_investment_snapshots_portfolio_id
    ON core.investment_snapshots(portfolio_id) WHERE (deleted_at IS NULL);
