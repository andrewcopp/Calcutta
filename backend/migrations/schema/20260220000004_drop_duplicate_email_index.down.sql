CREATE UNIQUE INDEX IF NOT EXISTS users_email_unique
    ON core.users USING btree (email)
    WHERE email IS NOT NULL AND deleted_at IS NULL;
