# Calcutta Membership System

This document describes the unified system for inviting users to Calcuttas and managing their participation lifecycle.

## Overview

When a Calcutta admin invites an email address:
1. If user exists → create CalcuttaInvitation (pending)
2. If user doesn't exist → create User (invited status) + CalcuttaInvitation (pending)
3. User sets password (if needed) → User becomes active
4. User accepts invitation → Entry created (draft status)
5. User submits bids → Entry becomes final

## Schema

### CalcuttaInvitation Status
| Status | Meaning |
|--------|---------|
| `pending` | Waiting for user to accept |
| `accepted` | User accepted, Entry created |
| `revoked` | Admin explicitly revoked |

The `revoked_at` timestamp tracks when revocation occurred. Revoked invitations remain visible for audit purposes (not soft-deleted).

### Entry Status
| Status | Meaning |
|--------|---------|
| `draft` | User has entry, bids not finalized |
| `final` | User submitted bids |

Bidding deadline is a Calcutta property. When deadline passes, `final` entries participate; `draft` entries miss the train.

### Calcutta Visibility
| Visibility | Listed | Access |
|------------|--------|--------|
| `public` | Yes | Anyone can join |
| `unlisted` | No | Anyone with link can join |
| `private` | No | Invitation required |

For private calcuttas, having an accepted CalcuttaInvitation IS the whitelist.

## User Lifecycle

```
stub → (admin assigns email) → invited → (user sets password) → active
```

- **stub**: Historical user, no email (imported from old data)
- **invited**: Has email, no password yet
- **active**: Has password, can log in

## Key Design Decisions

1. **CalcuttaInvitation uses user_id, not email** — If email doesn't exist, create user first
2. **No "declined" status** — User can always accept later; admin revokes if needed
3. **Entry created on accept** — Not before, not after
4. **Draft/final, not locked** — Locked is a Calcutta property (bidding deadline)
5. **Revoked ≠ deleted** — Revoked invitations stay visible for history

## Migration

Migration `20260220000002_membership_system` adds:
- `core.calcuttas.visibility` (default: 'private')
- `core.entries.status` (default: 'draft')
- `core.calcutta_invitations.revoked_at`
- `core.users.invited_by` (optional audit field)
