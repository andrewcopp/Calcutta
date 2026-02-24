-- Migration: add_favorites_total_points
-- Created: 2026-02-24 04:19:41 UTC

SET search_path = '';
SET lock_timeout = '5s';
SET statement_timeout = '30s';

-- Add nullable favorites_total_points column to predicted_team_values
ALTER TABLE compute.predicted_team_values
    ADD COLUMN IF NOT EXISTS favorites_total_points double precision;

-- Add CHECK constraint for non-negative values
ALTER TABLE compute.predicted_team_values
    ADD CONSTRAINT ck_compute_predicted_team_values_favorites_total_points_gte0
    CHECK (favorites_total_points >= 0);
