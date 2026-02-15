-- Revert scope_type check to exclude 'tournament'.
ALTER TABLE core.grants DROP CONSTRAINT IF EXISTS grants_scope_type_check;
ALTER TABLE core.grants ADD CONSTRAINT grants_scope_type_check
  CHECK (scope_type = ANY (ARRAY['global'::text, 'calcutta'::text]));

-- Revert scope_id check.
ALTER TABLE core.grants DROP CONSTRAINT IF EXISTS grants_scope_id_check;
ALTER TABLE core.grants ADD CONSTRAINT grants_scope_id_check
  CHECK (
    (scope_type = 'global' AND scope_id IS NULL)
    OR (scope_type <> 'global' AND scope_id IS NOT NULL)
  );
