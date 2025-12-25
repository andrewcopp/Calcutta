DROP TRIGGER IF EXISTS trg_tournaments_set_import_key ON tournaments;
DROP TRIGGER IF EXISTS trg_schools_set_slug ON schools;

DROP FUNCTION IF EXISTS set_tournaments_import_key();
DROP FUNCTION IF EXISTS set_schools_slug();
DROP FUNCTION IF EXISTS calcutta_slugify(TEXT);

DROP INDEX IF EXISTS uq_tournaments_import_key;
ALTER TABLE tournaments
    DROP COLUMN IF EXISTS import_key;

DROP INDEX IF EXISTS uq_schools_slug;
ALTER TABLE schools
    DROP COLUMN IF EXISTS slug;
