# silence_the_haters

- [ ] Verify no secrets are tracked: `git status --ignored`, `git ls-files | grep -E '\\.env(\\.|$)'`
- [ ] Rotate any dev tokens/keys that were ever shared outside the local machine
- [ ] Scan git history for leaked secrets (see checklist commands below)
- [ ] Remove hard exits from non-`main` packages (e.g. `log.Fatalf` inside constructors)
- [ ] Make `cmd/*` CLIs consistent (flags, help text, exit codes)
- [ ] Centralize config for all Go commands (one loader; support `.env` + `.env.local`)
- [ ] Eliminate TODOs in business logic (e.g. First Four / play-in handling)
- [ ] Add/expand unit tests for key logic paths (deterministic, single-assert tests)
- [ ] Add a CONTRIBUTING.md with “how to run locally” for backend, frontend, data-science
- [ ] Add a SECURITY.md with reporting guidelines + note on secrets handling
- [ ] Ensure generated artifacts are not committed (`out/`, `.venv/`, build output)
- [ ] Add a one-command developer bootstrap (Makefile targets or scripts)

## Git history secret-scan commands (read-only)

- [ ] `git grep -n -I "JWT_SECRET" $(git rev-list --all)`
- [ ] `git grep -n -I "API_KEY" $(git rev-list --all)`
- [ ] `git log --all --name-only --pretty=format: | sort | uniq | grep -E '\\.(env|pem|key)$'`
- [ ] (Optional) run `gitleaks` or `trufflehog` locally and review findings
