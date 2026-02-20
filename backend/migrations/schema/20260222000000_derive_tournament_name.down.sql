-- Re-add name column to tournaments
ALTER TABLE core.tournaments ADD COLUMN name text NOT NULL DEFAULT '';

-- Backfill tournament names from competition + season
UPDATE core.tournaments t
SET name = comp.name || ' ' || seas.year
FROM core.competitions comp, core.seasons seas
WHERE comp.id = t.competition_id AND seas.id = t.season_id;

-- Rename competition back
UPDATE core.competitions SET name = 'NCAA Men''s' WHERE name = 'NCAA Tournament';
