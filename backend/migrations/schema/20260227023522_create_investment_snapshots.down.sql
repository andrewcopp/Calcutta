-- Rollback: create_investment_snapshots
-- Created: 2026-02-27 02:35:22 UTC

SET search_path = '';
SET lock_timeout = '5s';
SET statement_timeout = '30s';

ALTER TABLE IF EXISTS core.investment_snapshots
    DROP CONSTRAINT IF EXISTS investment_snapshots_changed_by_fkey;
ALTER TABLE IF EXISTS core.investment_snapshots
    DROP CONSTRAINT IF EXISTS investment_snapshots_portfolio_id_fkey;
DROP TRIGGER IF EXISTS trg_core_investment_snapshots_updated_at ON core.investment_snapshots;
DROP TABLE IF EXISTS core.investment_snapshots;
