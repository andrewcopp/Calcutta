-- Add 'stub' to users status check, make email nullable, adjust uniqueness constraints

-- 1. Drop and re-add status check to include 'stub'
ALTER TABLE core.users DROP CONSTRAINT IF EXISTS users_status_check;
ALTER TABLE core.users ADD CONSTRAINT users_status_check
    CHECK (status = ANY (ARRAY['active'::text, 'invited'::text, 'requires_password_setup'::text, 'stub'::text]));

-- 2. Make email nullable
ALTER TABLE core.users ALTER COLUMN email DROP NOT NULL;

-- 3. Replace unique email constraint with partial unique index (only for non-null, non-deleted)
ALTER TABLE core.users DROP CONSTRAINT IF EXISTS users_email_key;
CREATE UNIQUE INDEX users_email_unique ON core.users (email)
    WHERE email IS NOT NULL AND deleted_at IS NULL;

-- 4. Add partial unique index on (first_name, last_name) for stubs to prevent duplicate stubs
CREATE UNIQUE INDEX users_stub_name_unique ON core.users (first_name, last_name)
    WHERE status = 'stub' AND deleted_at IS NULL;
