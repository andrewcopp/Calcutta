# silence_the_haters

## Security / secrets

- [ ] **Verify no secrets are tracked:** `git status --ignored`, `git ls-files | grep -E '\\.env(\\.|$)'`
- [ ] **Rotate any dev tokens/keys** that were ever shared outside the local machine
- [ ] **Scan git history for leaked secrets** (commands below)
- [ ] **Document rotations** in the “Findings / actions taken” section below

## Go backend: DX + correctness

- [ ] **Remove hard exits from non-`main` packages** (e.g. `log.Fatalf` inside constructors)
- [ ] **Make `cmd/*` CLIs consistent** (flags, help text, exit codes)
- [ ] **Centralize config for all Go commands** (one loader; support `.env` + `.env.local`)
- [x] **Fix `cmd/workers` to run multiple workers concurrently** (no longer blocks on first worker)

## Simulation / business logic

- [ ] **Eliminate TODOs in business logic** (e.g. First Four / play-in handling)
- [ ] **Add/expand unit tests for key logic paths** (deterministic, one assertion per test)

## OSS polish

- [ ] **Add a `CONTRIBUTING.md`** with “how to run locally” for backend, frontend, data-science
- [ ] **Add a `SECURITY.md`** with reporting guidelines + note on secrets handling
- [ ] **Ensure generated artifacts are not committed** (`out/`, `.venv/`, build output)
- [ ] **Add a one-command developer bootstrap** (Makefile targets or scripts)

## Git history secret-scan commands (read-only)

- [ ] **Scan for common secret keys:** `git grep -n -I "JWT_SECRET" $(git rev-list --all)`
- [ ] **Scan for generic key names:** `git grep -n -I "API_KEY" $(git rev-list --all)`
- [ ] **Scan for suspicious filenames in history:** `git log --all --name-only --pretty=format: | sort | uniq | grep -E '\\.(env|pem|key)$'`
- [ ] **Optional: run a dedicated scanner** (gitleaks/trufflehog) and review findings

## Findings / actions taken

- [ ] **Secret scan findings:**
  - 
- [ ] **Rotations performed:**
  - 
