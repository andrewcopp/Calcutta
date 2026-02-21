-- Seed Data Migration
-- Inserts permissions, labels, label_permissions, and reference data

SET search_path = '';

BEGIN;

-- =============================================================================
-- PERMISSIONS (16)
-- =============================================================================

INSERT INTO core.permissions (id, key, description) VALUES
  ('06f9f5af-0285-4852-9430-106aca75100c', 'admin.analytics.export', 'Export analytics snapshots'),
  ('47cd8bba-46e6-4959-8b86-72eaed4100be', 'admin.analytics.read', 'Read analytics'),
  ('5a1ebaa1-368e-4c4e-a4aa-d12a3f9ebc21', 'admin.api_keys.write', 'Create and revoke API keys'),
  ('20d84b63-ec46-4f20-8afa-9e5b9397f171', 'admin.bundles.export', 'Export bundles'),
  ('3b29ff2e-a969-4dfe-848e-47de12a762cc', 'admin.bundles.import', 'Import bundles'),
  ('10cca9d3-32eb-4ef8-ba34-2a398ea42069', 'admin.bundles.read', 'Read bundle upload status'),
  ('a3a1a1cc-8a4d-4813-911a-601675cd25f6', 'admin.hof.read', 'Read hall of fame'),
  ('4d045f02-704b-475c-bb6a-5a88d742a07c', 'admin.users.read', 'Read users'),
  ('caeb9563-7298-4ef3-b9d2-4c13529baa20', 'admin.users.write', 'Create/update users'),
  ('a9c96ef0-d7e7-4b35-9d68-7bef3b89c9c1', 'calcutta.config.write', 'Create and configure a calcutta'),
  ('8ea2f503-a867-4f53-a6ee-a7113d60dc12', 'calcutta.invite.write', 'Invite and manage calcutta members'),
  ('074eb6a9-c3ab-4f27-9a0d-fa0779bb7bcb', 'calcutta.read', 'Read calcutta data'),
  ('a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c5d', 'entry.write', 'Edit any entry in a calcutta'),
  ('eb55e249-6edd-42eb-8fc3-41b3920de2fa', 'lab.read', 'Read lab models, entries, evaluations'),
  ('28df6d71-bf3e-44a8-b3aa-caa4a5be444c', 'lab.write', 'Create/modify lab entries, pipelines'),
  ('f8be010c-b161-4d9b-a8f2-268c08be5e43', 'tournament.game.write', 'Update tournament game results and winners')
ON CONFLICT (id) DO NOTHING;

-- =============================================================================
-- LABELS (5)
-- =============================================================================

INSERT INTO core.labels (id, key, description) VALUES
  ('7fd3956d-9df0-4c1b-b176-e7b8b6d01248', 'site_admin', 'Full site administration'),
  ('2623b7e3-4dcf-4378-a091-a2a33efe560c', 'tournament_admin', 'Update tournament game results'),
  ('49b7d3d1-e45f-49b7-9bdc-424959d0c7ab', 'calcutta_admin', 'Manage a specific calcutta'),
  ('fcbaa74f-823e-4015-8680-522f2e210fef', 'player', 'Participate in a calcutta'),
  ('b8e1c2d3-f4a5-4b6c-8d7e-9f0a1b2c3d4e', 'user_manager', 'Manage user accounts: password resets, invite resends, user admin')
ON CONFLICT (id) DO NOTHING;

-- =============================================================================
-- LABEL_PERMISSIONS (25)
-- =============================================================================

-- site_admin gets all permissions
INSERT INTO core.label_permissions (label_id, permission_id) VALUES
  ('7fd3956d-9df0-4c1b-b176-e7b8b6d01248', 'f8be010c-b161-4d9b-a8f2-268c08be5e43'),
  ('7fd3956d-9df0-4c1b-b176-e7b8b6d01248', 'a9c96ef0-d7e7-4b35-9d68-7bef3b89c9c1'),
  ('7fd3956d-9df0-4c1b-b176-e7b8b6d01248', '8ea2f503-a867-4f53-a6ee-a7113d60dc12'),
  ('7fd3956d-9df0-4c1b-b176-e7b8b6d01248', '074eb6a9-c3ab-4f27-9a0d-fa0779bb7bcb'),
  ('7fd3956d-9df0-4c1b-b176-e7b8b6d01248', '06f9f5af-0285-4852-9430-106aca75100c'),
  ('7fd3956d-9df0-4c1b-b176-e7b8b6d01248', '5a1ebaa1-368e-4c4e-a4aa-d12a3f9ebc21'),
  ('7fd3956d-9df0-4c1b-b176-e7b8b6d01248', '20d84b63-ec46-4f20-8afa-9e5b9397f171'),
  ('7fd3956d-9df0-4c1b-b176-e7b8b6d01248', '3b29ff2e-a969-4dfe-848e-47de12a762cc'),
  ('7fd3956d-9df0-4c1b-b176-e7b8b6d01248', '10cca9d3-32eb-4ef8-ba34-2a398ea42069'),
  ('7fd3956d-9df0-4c1b-b176-e7b8b6d01248', '47cd8bba-46e6-4959-8b86-72eaed4100be'),
  ('7fd3956d-9df0-4c1b-b176-e7b8b6d01248', 'a3a1a1cc-8a4d-4813-911a-601675cd25f6'),
  ('7fd3956d-9df0-4c1b-b176-e7b8b6d01248', '4d045f02-704b-475c-bb6a-5a88d742a07c'),
  ('7fd3956d-9df0-4c1b-b176-e7b8b6d01248', 'caeb9563-7298-4ef3-b9d2-4c13529baa20'),
  ('7fd3956d-9df0-4c1b-b176-e7b8b6d01248', '28df6d71-bf3e-44a8-b3aa-caa4a5be444c'),
  ('7fd3956d-9df0-4c1b-b176-e7b8b6d01248', 'eb55e249-6edd-42eb-8fc3-41b3920de2fa'),
  ('7fd3956d-9df0-4c1b-b176-e7b8b6d01248', 'a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c5d'),
  -- tournament_admin
  ('2623b7e3-4dcf-4378-a091-a2a33efe560c', 'f8be010c-b161-4d9b-a8f2-268c08be5e43'),
  ('2623b7e3-4dcf-4378-a091-a2a33efe560c', '074eb6a9-c3ab-4f27-9a0d-fa0779bb7bcb'),
  -- calcutta_admin
  ('49b7d3d1-e45f-49b7-9bdc-424959d0c7ab', 'a9c96ef0-d7e7-4b35-9d68-7bef3b89c9c1'),
  ('49b7d3d1-e45f-49b7-9bdc-424959d0c7ab', '8ea2f503-a867-4f53-a6ee-a7113d60dc12'),
  ('49b7d3d1-e45f-49b7-9bdc-424959d0c7ab', '074eb6a9-c3ab-4f27-9a0d-fa0779bb7bcb'),
  ('49b7d3d1-e45f-49b7-9bdc-424959d0c7ab', 'a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c5d'),
  -- player
  ('fcbaa74f-823e-4015-8680-522f2e210fef', '074eb6a9-c3ab-4f27-9a0d-fa0779bb7bcb'),
  -- user_manager
  ('b8e1c2d3-f4a5-4b6c-8d7e-9f0a1b2c3d4e', '4d045f02-704b-475c-bb6a-5a88d742a07c'),
  ('b8e1c2d3-f4a5-4b6c-8d7e-9f0a1b2c3d4e', 'caeb9563-7298-4ef3-b9d2-4c13529baa20')
ON CONFLICT (label_id, permission_id) DO NOTHING;

-- =============================================================================
-- COMPETITIONS
-- =============================================================================

INSERT INTO core.competitions (name) VALUES ('NCAA Tournament')
ON CONFLICT (name) DO NOTHING;

COMMIT;
