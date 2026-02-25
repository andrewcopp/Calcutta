SET search_path = '';

-- Schools are reference data; down migration is intentionally empty.
-- Deleting schools would cascade-break teams in tournament seed migrations.
