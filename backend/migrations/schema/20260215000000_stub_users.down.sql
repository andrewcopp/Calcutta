-- Revert stub users schema changes

DROP INDEX IF EXISTS core.users_stub_name_unique;
DROP INDEX IF EXISTS core.users_email_unique;

-- Re-add original unique constraint on email
ALTER TABLE core.users ADD CONSTRAINT users_email_key UNIQUE (email);

-- Make email NOT NULL again (will fail if nulls exist)
ALTER TABLE core.users ALTER COLUMN email SET NOT NULL;

-- Revert status check
ALTER TABLE core.users DROP CONSTRAINT IF EXISTS users_status_check;
ALTER TABLE core.users ADD CONSTRAINT users_status_check
    CHECK (status = ANY (ARRAY['active'::text, 'invited'::text, 'requires_password_setup'::text]));
