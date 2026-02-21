CREATE FUNCTION core.calcutta_slugify(input text) RETURNS text
    LANGUAGE sql IMMUTABLE
    AS $$
    SELECT trim(both '-' from regexp_replace(lower(input), '[^a-z0-9]+', '-', 'g'))
$$;
