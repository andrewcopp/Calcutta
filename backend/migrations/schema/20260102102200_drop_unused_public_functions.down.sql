-- Recreate legacy helper functions that were previously present in public.
-- These were historically used for the pre-core schema and are not expected to be needed by the current runtime.

CREATE OR REPLACE FUNCTION public.get_entry_portfolio(p_run_id character varying, p_entry_key character varying)
RETURNS TABLE(team_key character varying, school_name character varying, seed integer, region character varying, bid_amount integer)
LANGUAGE plpgsql
AS $$
BEGIN
    IF p_entry_key = 'our_strategy' THEN
        RETURN QUERY
        SELECT 
            t.team_key,
            t.school_name,
            t.seed,
            t.region,
            reb.bid_amount_points as bid_amount
        FROM gold_recommended_entry_bids reb
        JOIN bronze_teams t ON reb.team_key = t.team_key
        WHERE reb.run_id = p_run_id
        ORDER BY reb.bid_amount_points DESC;
    ELSE
        RETURN QUERY
        SELECT 
            t.team_key,
            t.school_name,
            t.seed,
            t.region,
            eb.bid_amount
        FROM bronze_entry_bids eb
        JOIN bronze_teams t ON eb.team_key = t.team_key
        JOIN gold_optimization_runs r ON eb.calcutta_key = r.calcutta_key
        WHERE r.run_id = p_run_id AND eb.entry_key = p_entry_key
        ORDER BY eb.bid_amount DESC;
    END IF;
END;
$$;

CREATE OR REPLACE FUNCTION public.set_schools_slug()
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

CREATE OR REPLACE FUNCTION public.set_tournaments_import_key()
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
