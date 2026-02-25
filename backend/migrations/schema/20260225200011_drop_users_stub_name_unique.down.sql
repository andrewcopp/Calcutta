SET search_path = '';

CREATE UNIQUE INDEX users_stub_name_unique
    ON core.users USING btree (first_name, last_name)
    WHERE ((status = 'stub'::text) AND (deleted_at IS NULL));
