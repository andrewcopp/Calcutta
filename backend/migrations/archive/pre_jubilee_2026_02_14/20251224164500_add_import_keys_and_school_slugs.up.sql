ALTER TABLE schools
    ADD COLUMN slug TEXT;

ALTER TABLE tournaments
    ADD COLUMN import_key TEXT;

CREATE OR REPLACE FUNCTION calcutta_slugify(input TEXT)
RETURNS TEXT
LANGUAGE SQL
IMMUTABLE
AS $$
    SELECT trim(both '-' from regexp_replace(lower(input), '[^a-z0-9]+', '-', 'g'))
$$;

UPDATE schools
SET slug = calcutta_slugify(name)
WHERE slug IS NULL;

WITH dupes AS (
    SELECT slug
    FROM schools
    WHERE deleted_at IS NULL
    GROUP BY slug
    HAVING COUNT(*) > 1
)
UPDATE schools s
SET slug = s.slug || '-' || left(md5(s.name), 6)
FROM dupes d
WHERE s.slug = d.slug AND s.deleted_at IS NULL;

ALTER TABLE schools
    ALTER COLUMN slug SET NOT NULL;

CREATE UNIQUE INDEX uq_schools_slug ON schools (slug)
WHERE deleted_at IS NULL;

UPDATE tournaments
SET import_key = calcutta_slugify(name)
WHERE import_key IS NULL;

WITH dupes AS (
    SELECT import_key
    FROM tournaments
    WHERE deleted_at IS NULL
    GROUP BY import_key
    HAVING COUNT(*) > 1
)
UPDATE tournaments t
SET import_key = t.import_key || '-' || left(md5(t.name), 6)
FROM dupes d
WHERE t.import_key = d.import_key AND t.deleted_at IS NULL;

ALTER TABLE tournaments
    ALTER COLUMN import_key SET NOT NULL;

CREATE UNIQUE INDEX uq_tournaments_import_key ON tournaments (import_key)
WHERE deleted_at IS NULL;

CREATE OR REPLACE FUNCTION set_schools_slug()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    IF NEW.slug IS NULL THEN
        NEW.slug := calcutta_slugify(NEW.name);
    END IF;
    RETURN NEW;
END;
$$;

CREATE OR REPLACE FUNCTION set_tournaments_import_key()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    IF NEW.import_key IS NULL THEN
        NEW.import_key := calcutta_slugify(NEW.name);
    END IF;
    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS trg_schools_set_slug ON schools;
CREATE TRIGGER trg_schools_set_slug
BEFORE INSERT OR UPDATE OF name ON schools
FOR EACH ROW
EXECUTE FUNCTION set_schools_slug();

DROP TRIGGER IF EXISTS trg_tournaments_set_import_key ON tournaments;
CREATE TRIGGER trg_tournaments_set_import_key
BEFORE INSERT OR UPDATE OF name ON tournaments
FOR EACH ROW
EXECUTE FUNCTION set_tournaments_import_key();
