ALTER TABLE core.users
  DROP COLUMN IF EXISTS reset_token_hash,
  DROP COLUMN IF EXISTS reset_expires_at,
  DROP COLUMN IF EXISTS reset_consumed_at,
  DROP COLUMN IF EXISTS last_reset_sent_at;
