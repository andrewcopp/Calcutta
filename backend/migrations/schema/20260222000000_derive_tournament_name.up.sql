-- Rename competition from "NCAA Men's" to "NCAA Tournament"
UPDATE core.competitions SET name = 'NCAA Tournament' WHERE name = 'NCAA Men''s';

-- Drop the redundant name column from tournaments
ALTER TABLE core.tournaments DROP COLUMN name;
