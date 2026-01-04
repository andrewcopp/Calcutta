# User invites + historical entries claim plan

## Goals

- [ ] Allow everyone to log in and see their historical Calcutta entries
- [ ] Support inviting historical players via email for the 2026 Calcutta
- [ ] Keep the implementation lightweight (small pool, ~70 people)
- [ ] Preserve a manual/admin recovery path for the inevitable edge cases

## Non-goals

- [ ] Perfect identity resolution across all time (we will do a one-time cleanup + keep an admin fix path)
- [ ] General-purpose “account linking” across multiple auth providers
- [ ] High-scale deliverability/marketing email features

## Key decisions to confirm

- [x] **Invite vs magic-link**
  - [x] Option A: invite email contains a token that leads to a “set password” page
  - [ ] Option B: invite email is a pure magic-link login (no password required)
- [ ] **Token TTL**
  - [x] Choose default expiry: 7 days
  - [x] Resend behavior: invalidate previous tokens on resend
- [x] **User state naming**
  - [x] Use `invited` / `requires_password_setup` semantics (not `password_reset`)
- [x] **Initial admin bootstrap**
  - [x] Use env-based bootstrap (Option 1)
  - [x] Support a single email (`BOOTSTRAP_ADMIN_EMAIL`)
- [x] **Email uniqueness**
  - [x] Confirm `users.email` is unique
  - [x] Collision policy: skip + report

## Data model / schema

### A. Users

- [x] Ensure the `users` table supports an “invited but not yet claimed” state
- [x] Add fields to support invite/claim tokens:
  - [x] `invite_token_hash` (store hash only)
  - [x] `invite_expires_at`
  - [x] `invite_consumed_at` (or `invite_claimed_at`)
  - [x] `invited_at`
  - [x] `last_invite_sent_at`
  - [x] `status` (or equivalent) supports `invited`, `requires_password_setup`, and `active`

### B. Preserve historical traceability on entries

- [ ] Ensure each historical entry is attached to a `user_id`
- [ ] Keep a copy of the original sheet name on entries (or ensure it already exists):
  - [ ] `legacy_entry_name` (or equivalent)

### C. Optional supporting tables (only if needed)

- [ ] If name cleanup needs aliasing, introduce a `user_aliases` table:
  - [ ] `user_id`, `alias_name`, `created_at`

## Migrations

- [x] Add migration(s) for invite/claim columns
- [ ] Add migration(s) for `legacy_entry_name` if not already present
- [ ] Add indexes:
  - [ ] `users(email)` unique
  - [x] `users(invite_expires_at)` (optional)
  - [x] `users(status)` (optional)
- [ ] Add constraints:
  - [ ] Ensure `invite_token_hash` cannot be reused after `invite_consumed_at` is set

## Backfill / import process (one-time)

- [ ] Identify distinct historical people (canonical names)
- [ ] Collect their emails from the organizer
- [ ] Create one `user` per person/email
  - [ ] Set status to `invited` (or `requires_password_setup`)
  - [ ] No password set initially (or password unusable)
- [ ] Attach each historical entry to the correct `user_id`
  - [ ] Preserve the original entry name in `legacy_entry_name`
- [ ] Produce a one-time verification report
  - [ ] For each user: years played, # entries
  - [ ] Flag any entry without user assignment

## Invite / claim (“magic link”) implementation

### Token design

- [x] Use a cryptographically secure random token (at least 128 bits; 256 bits is fine)
- [x] Store only a hash in the DB (e.g. SHA-256(token))
- [x] Enforce:
  - [x] Expiration check
  - [x] Single-use (`invite_consumed_at IS NULL`)
  - [x] Invalidate previous invites on resend

### Flows

- [ ] **Admin send invite**
  - [x] Generates token + saves hash/expiry
  - [ ] Sends email with claim URL
- [ ] **Accept invite**
  - [x] Validates token
  - [x] Marks token consumed
  - [x] Transitions user to “claim in progress” or directly to “active” depending on UX
- [ ] **Set password** (if using “invite -> set password”)
  - [x] User sets password after accepting invite
  - [x] Transition user to `active`

## API endpoints

- [ ] Admin-only:
  - [x] `POST /admin/users/{id}/invite` (create + send invite)
  - [x] `POST /admin/users/{id}/invite/resend`
  - [ ] `POST /admin/entries/{id}/reassign` (manual fix)
  - [x] `GET /admin/users?status=invited` (list)
- [ ] Public:
  - [x] `POST /auth/invite/accept` (token -> session or next step)
  - [ ] `POST /auth/password/set` (token/session -> password)

## Email delivery

- [ ] Decide email provider approach:
  - [ ] SMTP
  - [ ] Transactional provider (SendGrid/Mailgun/etc.)
- [ ] Create templates:
  - [ ] Initial invite (“invited to 2026 Calcutta + view historical Calcuttas”)
  - [ ] Resend invite
- [ ] Add basic operational controls:
  - [ ] Rate limiting resend
  - [ ] Log send attempts + errors

## Auth / authorization

- [x] Ensure invited users cannot access protected resources until claimed/active (or decide what they can see pre-claim)
- [x] Ensure historical entries access is by `user_id` linkage only (no name fallback for auth)
- [ ] Admin bootstrap
  - [x] Add `BOOTSTRAP_ADMIN_EMAIL` support
  - [x] On startup, ensure there is a global admin grant for that email
  - [ ] If the user does not exist, create them in `requires_password_setup` and send an invite
  - [x] Remove or guard any “first signup becomes admin” behavior so prod bootstrap is explicit

## Admin UI (minimal)

- [ ] Page: “Invited / Unclaimed users”
  - [ ] List users with status `invited`
  - [ ] Resend invite
  - [ ] View associated entries/years
- [ ] Page/controls: “Entry ownership fixes”
  - [ ] Search entry by year/name
  - [ ] Reassign to a different user

## Testing

- [ ] Unit tests for token generation and hashing
- [ ] Unit tests for token validation:
  - [ ] Expired token rejected
  - [ ] Consumed token rejected
  - [ ] Unknown token rejected
- [ ] Unit tests for state transitions:
  - [ ] invited -> active
- [ ] Authorization tests:
  - [ ] user can only see entries where `entries.user_id = current_user_id`

## Observability / audit

- [ ] Log invite creation and consumption (user id, timestamps)
- [ ] Log resend events
- [ ] Record admin actor for admin endpoints (if available)

## Runbook

- [ ] Steps to run the one-time import
- [ ] Steps to send invites (dry run option)
- [ ] Steps to bootstrap initial admin (`BOOTSTRAP_ADMIN_EMAIL`)
- [ ] Troubleshooting: bounced emails, wrong assignment, user requests a new email
- [ ] Manual recovery: reassign entries, update email, resend invite
