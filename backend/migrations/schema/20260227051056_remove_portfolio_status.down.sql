-- Rollback: remove_portfolio_status
-- Created: 2026-02-27 05:10:56 UTC

SET search_path = '';
SET lock_timeout = '5s';
SET statement_timeout = '30s';

-- Restore the status column with original type and default
ALTER TABLE core.portfolios
    ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'draft';

ALTER TABLE core.portfolios
    ADD CONSTRAINT ck_portfolios_status CHECK (status IN ('draft', 'submitted'));

-- Revert investment_snapshots FK back to plain (no CASCADE)
ALTER TABLE core.investment_snapshots
    DROP CONSTRAINT IF EXISTS investment_snapshots_portfolio_id_fkey;

ALTER TABLE core.investment_snapshots
    ADD CONSTRAINT investment_snapshots_portfolio_id_fkey
    FOREIGN KEY (portfolio_id) REFERENCES core.portfolios(id);
