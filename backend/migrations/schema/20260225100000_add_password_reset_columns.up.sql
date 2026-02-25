ALTER TABLE core.users
  ADD COLUMN reset_token_hash text,
  ADD COLUMN reset_expires_at timestamp with time zone,
  ADD COLUMN reset_consumed_at timestamp with time zone,
  ADD COLUMN last_reset_sent_at timestamp with time zone;
