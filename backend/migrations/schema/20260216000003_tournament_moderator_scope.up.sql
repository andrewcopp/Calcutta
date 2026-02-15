-- Expand grants scope_type to include 'tournament' for moderator permissions.
ALTER TABLE core.grants DROP CONSTRAINT IF EXISTS grants_scope_type_check;
ALTER TABLE core.grants ADD CONSTRAINT grants_scope_type_check
  CHECK (scope_type = ANY (ARRAY['global'::text, 'calcutta'::text, 'tournament'::text]));

-- Update scope_id check: tournament scopes also require a non-null scope_id.
ALTER TABLE core.grants DROP CONSTRAINT IF EXISTS grants_scope_id_check;
ALTER TABLE core.grants ADD CONSTRAINT grants_scope_id_check
  CHECK (
    (scope_type = 'global' AND scope_id IS NULL)
    OR (scope_type <> 'global' AND scope_id IS NOT NULL)
  );
