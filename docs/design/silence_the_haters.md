# silence_the_haters

## Security / secrets

- [x] **Verify no secrets are tracked:** `git status --ignored`, `git ls-files | grep -E '\\.env(\\.|$)'`
- [ ] **Rotate any dev tokens/keys** that were ever shared outside the local machine
- [x] **Scan git history for leaked secrets** (commands below)
- [x] **Remove insecure default DB password fallbacks** in data-science code/docs
- [ ] **Document rotations** in the “Findings / actions taken” section below

## Go backend: DX + correctness

- [x] **Remove hard exits from non-`main` packages** (e.g. `log.Fatalf` inside constructors) (`6600a17`)
- [x] **Make `cmd/*` CLIs consistent** (flags, help text, exit codes) (`76ea23e`)
- [x] **Centralize config for all Go commands** (one loader; support `.env` + `.env.local`) (`6309cdf`)
- [x] **Fix `cmd/workers` to run multiple workers concurrently** (no longer blocks on first worker)

## Simulation / business logic

- [x] **Eliminate TODOs in business logic** (e.g. First Four / play-in handling) (`4fe745e`)
- [ ] **Add/expand unit tests for key logic paths** (deterministic, one assertion per test)

## OSS polish

- [x] **Add a `CONTRIBUTING.md`** with “how to run locally” for backend, frontend, data-science
- [x] **Add a `SECURITY.md`** with reporting guidelines + note on secrets handling
- [ ] **Ensure generated artifacts are not committed** (`out/`, `.venv/`, build output)
- [x] **Add a one-command developer bootstrap** (Makefile targets or scripts) (`56940c4`)

## Git history secret-scan commands (read-only)

- [x] **Scan for common secret keys:** `git grep -n -I "JWT_SECRET" $(git rev-list --all)`
- [x] **Scan for generic key names:** `git grep -n -I "API_KEY" $(git rev-list --all)`
- [x] **Scan for suspicious filenames in history:** `git log --all --name-only --pretty=format: | sort | uniq | grep -E '\\.(env|pem|key)$'`
- [ ] **Optional: run a dedicated scanner** (gitleaks/trufflehog) and review findings

## Findings / actions taken

- [ ] **Secret scan findings:**
  - `git status --ignored` shows `.env`, `.env.local`, and `.envrc` are ignored (not tracked).
  - `git ls-files | grep -E '\\.env(\\.|$)'` shows tracked examples only:
    - `.env.example`
    - `data-science/.env.example`
  - `git grep -n -I "JWT_SECRET" $(git rev-list --all)` results appear to be documentation/example references only (e.g. `.env.example` placeholder + README text). No evidence of an actual secret value in history from this scan.
  - `git grep -n -I "API_KEY" $(git rev-list --all)` results appear to be:
    - `.env.example` placeholder (`CALCUTTA_API_KEY=your-api-key-here`)
    - data-science docs/scripts that read `CALCUTTA_API_KEY` from the environment or CLI args
  - JWT-like token scan: `git grep -n -I -E 'eyJ[A-Za-z0-9_-]{10,}\\.[A-Za-z0-9_-]{10,}\\.[A-Za-z0-9_-]{10,}' $(git rev-list --all)` returned no matches.
  - Generic secret/token/password scan returned matches that appear non-sensitive:
    - GitHub workflow permissions (`id-token: write` for OIDC)
    - runtime code references (e.g. `TokenManager{secret: []byte(secret)}`)
    - default local DB password values in data-science scripts (not a leak; cleaned up in `75b4c06`)
  - Suspicious-filenames scan (`\\.(env|pem|key)$`) did not report any tracked history hits in the pasted output.
  - AWS key pattern scan (`AKIA[0-9A-Z]{16}`) returned no matches.
- [ ] **Rotations performed:**
  - (Fill these out as you rotate; do not paste secret values into this repo.)

### Rotation checklist (what should be rotated if this repo was shared)

- [ ] **Backend auth:** `JWT_SECRET` (if using `AUTH_MODE != cognito`)
- [ ] **API auth:** `CALCUTTA_API_KEY` (any key used for `Authorization: Bearer`)
- [ ] **Database credentials:** `DB_PASSWORD`, `CALCUTTA_ANALYTICS_DB_PASSWORD`, connection URLs that embed passwords
- [ ] **Cognito / OAuth:** any client secrets (if applicable)
- [ ] **CI/CD:** GitHub Actions secrets, deploy tokens, cloud provider credentials (if applicable)

### Secrets source of truth (recommended)

- **Source of truth:** a password manager (recommended: 1Password) holding the *current* values for:
  - `JWT_SECRET`
  - `CALCUTTA_API_KEY`
  - DB passwords / URLs (avoid embedding passwords in URLs if you can)
- **Local dev:** copy from password manager into local `.env` / `.env.local` (gitignored).
- **Repo:** `.env.example` contains placeholders only.
- **Later (when deployed):** use AWS Secrets Manager as the deployed source of truth.
  - [ ] TODO: define secret names/paths and required IAM permissions
  - [ ] TODO: wire runtime config to read from Secrets Manager in deployed environments
  - [ ] TODO: document the deploy-time workflow (how secrets get created/updated)
  - [ ] TODO: treat password manager as break-glass / human vault

### Rotations performed (template)

- [ ] **Rotation entry**
  - **Secret name:**
  - **Scope/system:** (local dev / staging / prod / GitHub / AWS / etc)
  - **When:** (YYYY-MM-DD)
  - **How rotated:** (where you generated it; e.g. AWS console, GitHub settings, password manager)
  - **Where stored now:** (1Password item name, env var manager, GitHub Actions secret name)
  - **Where updated:**
    - [ ] local `.env`
    - [ ] deployment env vars
    - [ ] CI secrets
    - [ ] other:
  - **Old secret invalidated:** (yes/no/unknown)
  - **Verification:** (what you did to confirm new secret works)

- [ ] **Rotation entry**
  - **Secret name:** (process) secrets source-of-truth established
  - **Scope/system:** local dev
  - **When:** (YYYY-MM-DD)
  - **How rotated:** created entries in password manager; removed any previous ad-hoc copies
  - **Where stored now:** (password manager item name)
  - **Where updated:**
    - [ ] local `.env`
    - [ ] other:
  - **Old secret invalidated:** n/a
  - **Verification:** local `go run` / `docker compose` succeeds using `.env` without hardcoding secrets
