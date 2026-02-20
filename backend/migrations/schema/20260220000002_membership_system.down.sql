-- Remove invited_by from users
ALTER TABLE core.users 
DROP COLUMN invited_by;

-- Revert calcutta_invitations changes
ALTER TABLE core.calcutta_invitations 
DROP CONSTRAINT ck_calcutta_invitations_status;

ALTER TABLE core.calcutta_invitations 
ADD CONSTRAINT ck_calcutta_invitations_status 
CHECK (status IN ('pending', 'accepted', 'declined'));

ALTER TABLE core.calcutta_invitations 
DROP COLUMN revoked_at;

-- Remove status from entries
ALTER TABLE core.entries 
DROP CONSTRAINT ck_entries_status;

ALTER TABLE core.entries 
DROP COLUMN status;

-- Remove visibility from calcuttas
ALTER TABLE core.calcuttas 
DROP CONSTRAINT ck_calcuttas_visibility;

ALTER TABLE core.calcuttas 
DROP COLUMN visibility;
