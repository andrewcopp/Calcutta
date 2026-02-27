-- Migration: remove_portfolio_status
-- Created: 2026-02-27 05:10:56 UTC

SET search_path = '';
SET lock_timeout = '5s';
SET statement_timeout = '30s';

-- Drop the status check constraint and column from portfolios
ALTER TABLE core.portfolios
    DROP CONSTRAINT IF EXISTS ck_portfolios_status;

ALTER TABLE core.portfolios
    DROP COLUMN IF EXISTS status;

-- Fix investment_snapshots FK to add ON DELETE CASCADE
ALTER TABLE core.investment_snapshots
    DROP CONSTRAINT IF EXISTS investment_snapshots_portfolio_id_fkey;

ALTER TABLE core.investment_snapshots
    ADD CONSTRAINT investment_snapshots_portfolio_id_fkey
    FOREIGN KEY (portfolio_id) REFERENCES core.portfolios(id) ON DELETE CASCADE;
