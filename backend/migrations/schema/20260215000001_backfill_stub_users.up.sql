-- Create stub users from entry names and backfill entries.user_id
-- Handles entries that either have no user_id OR all share the same generic admin user_id

-- 1. Insert stub users from distinct entry names
-- We create stubs for entries where the name doesn't match the linked user's actual name
INSERT INTO core.users (id, email, first_name, last_name, status, created_at, updated_at)
SELECT
    gen_random_uuid(),
    NULL,
    CASE
        WHEN position(' ' in e.name) > 0 THEN left(e.name, length(e.name) - length(reverse(split_part(reverse(e.name), ' ', 1))) - 1)
        ELSE e.name
    END,
    CASE
        WHEN position(' ' in e.name) > 0 THEN reverse(split_part(reverse(e.name), ' ', 1))
        ELSE ''
    END,
    'stub',
    NOW(),
    NOW()
FROM (
    SELECT DISTINCT e.name
    FROM core.entries e
    LEFT JOIN core.users u ON u.id = e.user_id
    WHERE e.deleted_at IS NULL
      -- Include entries with no user, or where the entry name doesn't match the user's name
      AND (e.user_id IS NULL
           OR (u.first_name || ' ' || u.last_name) <> e.name)
) e
-- Don't create stubs if a user with this name already exists
WHERE NOT EXISTS (
    SELECT 1 FROM core.users existing
    WHERE existing.deleted_at IS NULL
      AND existing.first_name = CASE
          WHEN position(' ' in e.name) > 0 THEN left(e.name, length(e.name) - length(reverse(split_part(reverse(e.name), ' ', 1))) - 1)
          ELSE e.name
      END
      AND existing.last_name = CASE
          WHEN position(' ' in e.name) > 0 THEN reverse(split_part(reverse(e.name), ' ', 1))
          ELSE ''
      END
);

-- 2. Backfill entries.user_id by matching name to the correct user (stub or active)
-- Only update entries where the current user doesn't match the entry name
UPDATE core.entries e
SET user_id = u.id
FROM core.users u
WHERE e.deleted_at IS NULL
  AND u.deleted_at IS NULL
  AND e.name = (u.first_name || ' ' || u.last_name)
  -- Only update if current user doesn't match entry name
  AND (e.user_id IS NULL
       OR NOT EXISTS (
           SELECT 1 FROM core.users cu
           WHERE cu.id = e.user_id
             AND (cu.first_name || ' ' || cu.last_name) = e.name
       ));
