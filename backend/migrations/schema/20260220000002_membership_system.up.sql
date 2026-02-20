-- Add visibility to calcuttas
ALTER TABLE core.calcuttas 
ADD COLUMN visibility text NOT NULL DEFAULT 'private';

ALTER TABLE core.calcuttas 
ADD CONSTRAINT ck_calcuttas_visibility 
CHECK (visibility IN ('public', 'unlisted', 'private'));

-- Add status to entries
ALTER TABLE core.entries 
ADD COLUMN status text NOT NULL DEFAULT 'draft';

ALTER TABLE core.entries 
ADD CONSTRAINT ck_entries_status 
CHECK (status IN ('draft', 'final'));

-- Update calcutta_invitations: add revoked_at and update status constraint
ALTER TABLE core.calcutta_invitations 
ADD COLUMN revoked_at timestamptz;

ALTER TABLE core.calcutta_invitations 
DROP CONSTRAINT ck_calcutta_invitations_status;

ALTER TABLE core.calcutta_invitations 
ADD CONSTRAINT ck_calcutta_invitations_status 
CHECK (status IN ('pending', 'accepted', 'revoked'));

-- Add invited_by to users for audit trail
ALTER TABLE core.users 
ADD COLUMN invited_by uuid REFERENCES core.users(id);
