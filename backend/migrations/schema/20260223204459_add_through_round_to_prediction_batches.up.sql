-- Migration: add_through_round_to_prediction_batches
-- Created: 2026-02-23 20:44:59 UTC

SET search_path = '';
SET lock_timeout = '5s';
SET statement_timeout = '30s';

ALTER TABLE compute.prediction_batches
    ADD COLUMN IF NOT EXISTS through_round integer NOT NULL DEFAULT 0;
