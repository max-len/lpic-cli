# Agent Core Conventions

Minimal set of rules to persist across sessions. Keep this file short.

## Commit Messages
* Conventional Commits: `type(scope): subject`
* Subject line: <= 72 chars, imperative ("add", "fix", "refactor")
* No redundant history commentary (only what this commit changes)

## Interactive Git Operations
Before running something that opens an editor (rebase -i, amend, etc.):
1. State goal
2. Show command (do **not** run yet)
3. Describe expected buffer & required edits
4. Provide save / abort instructions (`:wq` / `:q!`)
5. Wait for explicit "proceed"

## Data Sensitivity
* Treat `test*.json` as confidential; only derive structure if needed

## Assistant Behavior
* Do not push without an explicit request
* Enforce 72‑char subject rule
* Keep diffs minimal & relevant
* After code changes: ensure `go build ./...` passes
* Prefer `make build` (places binary in `bin/`); plain `go build ./cmd/client` may create a stray `./client` not used in workflow

---
End of core conventions.
