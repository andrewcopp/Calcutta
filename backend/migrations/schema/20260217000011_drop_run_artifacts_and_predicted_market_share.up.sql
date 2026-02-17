-- Drop run_artifacts table (write-only, nobody reads)
DROP TABLE IF EXISTS derived.run_artifacts;

-- Drop predicted_market_share table (write-only, nobody reads)
DROP TABLE IF EXISTS derived.predicted_market_share;
