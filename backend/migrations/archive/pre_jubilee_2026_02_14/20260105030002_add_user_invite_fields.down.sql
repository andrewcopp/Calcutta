DROP INDEX IF EXISTS idx_users_invite_expires_at;
DROP INDEX IF EXISTS idx_users_status;
DROP INDEX IF EXISTS uq_users_invite_token_hash;

ALTER TABLE users
    DROP CONSTRAINT IF EXISTS users_status_check;

ALTER TABLE users
    DROP COLUMN IF EXISTS last_invite_sent_at,
    DROP COLUMN IF EXISTS invited_at,
    DROP COLUMN IF EXISTS invite_consumed_at,
    DROP COLUMN IF EXISTS invite_expires_at,
    DROP COLUMN IF EXISTS invite_token_hash,
    DROP COLUMN IF EXISTS status;
