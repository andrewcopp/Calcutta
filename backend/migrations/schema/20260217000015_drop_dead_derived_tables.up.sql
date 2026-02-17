-- derived.simulation_runs: hollowed out by 4 prior migrations, only touched by deleted retain tool
DROP TABLE IF EXISTS derived.simulation_runs CASCADE;

-- derived.optimized_entries: never written to, reads removed in code cleanup
DROP TABLE IF EXISTS derived.optimized_entries CASCADE;

-- lab.investment_models.is_benchmark: never read or written
ALTER TABLE lab.investment_models DROP COLUMN IF EXISTS is_benchmark;

-- lab.entries.state: dead state machine, always 'complete', never queried by value
DROP INDEX IF EXISTS idx_lab_entries_state;
ALTER TABLE lab.entries DROP COLUMN IF EXISTS state;
