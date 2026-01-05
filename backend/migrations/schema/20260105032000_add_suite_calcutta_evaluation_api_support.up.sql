ALTER TABLE IF EXISTS derived.suite_calcutta_evaluations
    ADD COLUMN IF NOT EXISTS starting_state_key TEXT NOT NULL DEFAULT 'post_first_four',
    ADD COLUMN IF NOT EXISTS excluded_entry_name TEXT,
    ADD COLUMN IF NOT EXISTS claimed_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS claimed_by TEXT;

INSERT INTO permissions (key, description)
VALUES
    ('analytics.suite_calcutta_evaluations.write', 'Submit suite calcutta evaluation requests'),
    ('analytics.suite_calcutta_evaluations.read', 'Read suite calcutta evaluation requests')
ON CONFLICT (key) DO NOTHING;

INSERT INTO label_permissions (label_id, permission_id)
SELECT l.id, p.id
FROM labels l
JOIN permissions p ON p.key IN (
    'analytics.suite_calcutta_evaluations.write',
    'analytics.suite_calcutta_evaluations.read'
)
WHERE l.key IN ('global_admin')
ON CONFLICT DO NOTHING;
