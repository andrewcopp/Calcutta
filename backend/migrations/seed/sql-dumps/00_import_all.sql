-- Generated on: 2025-12-21 00:03:06
-- This file contains seed data for the Calcutta application

-- Master seed data import file
-- This file imports all seed data in the correct order to maintain referential integrity
--
-- Usage: This file is automatically used by the seed migration process
-- To manually import: psql -d calcutta -f 00_import_all.sql

-- Disable triggers during import to avoid constraint issues
SET session_replication_role = replica;

-- Import in dependency order
\i schools.sql
\i tournaments.sql
\i tournament_teams.sql
\i users.sql
\i calcuttas.sql
\i calcutta_rounds.sql
\i calcutta_entries.sql
\i calcutta_entry_teams.sql

-- Re-enable triggers
SET session_replication_role = DEFAULT;
