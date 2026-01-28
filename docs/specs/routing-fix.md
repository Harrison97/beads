# Spec: Routing Improvements (Rebase From Upstream Main)

Owner: Harrison97 + Codex
Status: Draft
Target branch: `upstream/main` (base)

## Goal
Bring routing UX, observability, and correctness up to the standard of the learnings in `origin/main` without dragging along unrelated changes. This spec defines the exact behaviors to implement and the test/doc updates required, starting cleanly from `upstream/main`.

## Non‑Goals
- Do not introduce pattern/priority routing or any `bd routes` CLI.
- Do not change routing to use labels or titles for matching.
- Do not change backend selection logic outside the routing changes described here.

## Background / Learnings from Origin Main
The `origin/main` branch has a coherent set of routing fixes:
- Auto‑routing for creates based on explicit ID prefix *and* configured prefix.
- Debug logging and warning for malformed `routes.jsonl`.
- Documentation that matches the actual prefix‑based routing implementation.
- Tests covering error paths and edge cases for routing utilities.

Two potential gaps to address while re‑implementing:
1) Auto‑routing should **not** override an explicit `--repo` selection.
2) Town‑root discovery for auto‑routing should not depend on CWD if `--db` points elsewhere.

This spec includes both improvements.

## Definitions
- **Routing**: prefix‑based lookup using `.beads/routes.jsonl` lines of `{ "prefix": "bd-", "path": "beads/mayor/rig" }`.
- **Town root**: directory containing `mayor/town.json` at or above the current rig(s).
- **Routes file**: `.beads/routes.jsonl` at the town root.

## User Stories
1) As a user working in a multi‑rig town, I can create `bd-123` and have it go to the correct rig without explicitly specifying `--rig`.
2) As a user with misconfigured routes, I can see *why* routes are failing by enabling `BD_DEBUG_ROUTING=1`.
3) As a user, I never have my explicit `--repo` override silently ignored.
4) As a user running commands from any subdir or via `--db`, auto‑routing still works reliably.

## Functional Requirements

### FR1: Auto‑routing via explicit ID
If `--id` is provided and it includes a prefix, and neither `--rig` nor `--prefix` are provided:
- Find the town routes file.
- If the prefix matches a route whose `path` is **not** `"."`, route creation to that rig.

### FR2: Auto‑routing via configured prefix
If no explicit ID routing happened, and neither `--rig` nor `--prefix` are provided:
- Resolve the configured prefix (DB config keys `issue-prefix` or `issue_prefix`, then `config.yaml`).
- If configured prefix exists and a town route maps it to a different rig, route creation to that rig.

### FR3: Respect explicit `--repo`
If `--repo` is passed explicitly, **skip** auto‑routing (both explicit‑ID routing and configured‑prefix routing). `--repo` is always authoritative.

### FR4: Town discovery should not depend on CWD
When doing auto‑routing based on configured prefix, prefer town discovery using the **beads directory derived from the active db** rather than CWD. This avoids false negatives when `--db` points elsewhere.

### FR5: Debug logging for routes
When `BD_DEBUG_ROUTING=1`:
- Log which routes file is being loaded and whether it was missing.
- Log malformed JSON lines with line numbers.
- Log empty `prefix` or `path` entries.
- Log count of parsed vs skipped routes.
- Log auto‑routing decisions (explicit ID and configured prefix cases).

### FR6: Warning for completely broken routes
If `routes.jsonl` exists, has malformed lines, and results in **0 valid routes**, emit a warning to stderr:
- Include a hint about `BD_DEBUG_ROUTING=1` and suppression via `BD_QUIET_ROUTING=1`.

## Behavioral Edge Cases
- `path: "."` should map to town root `.beads` directory.
- Symlinked `.beads` should still find the correct town root.
- Unknown prefix or missing routes file must fall back to local creation (no error).

## API/Code Changes

### 1) `internal/routing/routes.go`
- Add package‑level documentation describing Gas Town layout and prefix routing.
- Extend `LoadRoutes` with debug logging and malformed‑routes warning.
- Add debug logging around `findTownRootFromCWD` to help trace symlink issues.
- Add helper `AutoDetectTargetRig(currentBeadsDir, configuredPrefix) (rigName string, shouldRoute bool, err error)`.

### 2) `cmd/bd/create.go`
- Add auto‑routing flow (explicit ID, then configured prefix) before `--rig/--prefix` handling.
- Respect `--repo` by **skipping** auto‑routing when `--repo` was explicitly passed.
- Use `dbPath` to derive current beads dir for auto‑routing configured prefix; do not rely on CWD.

### 3) `internal/routing/routing_test.go`
Add coverage for:
- `LoadRoutes` malformed JSON lines are skipped.
- `LoadRoutes` empty prefix/path are skipped.
- `LoadRoutes` empty file, comments‑only file, missing file (nil slice, no error).
- `LoadRoutes` mixed content.
- `ResolveBeadsDirForRig` (non‑existent rig, non‑existent target dir, “.” path, redirects, all input formats).
- `ResolveBeadsDirForID` (unknown prefix, no prefix, no routes, non‑existent target dir, “.” path).
- Warning behavior for completely broken routes, and suppression via `BD_QUIET_ROUTING=1`.

### 4) `website/docs/multi-agent/routing.md`
Rewrite to reflect prefix‑based routing:
- Remove references to pattern‑based routing and `bd routes` CLI.
- Document routes.jsonl with `prefix`/`path` fields.
- Document multi‑rig Gas Town structure and `--rig` usage.
- Document `BD_DEBUG_ROUTING` and warning behavior.
- Document redirect file and symlinked `.beads` behavior.

## Acceptance Criteria
- Auto‑routing triggers for explicit ID and configured prefix, but not when `--repo` is explicitly passed.
- Debug logs appear with `BD_DEBUG_ROUTING=1` and are silent otherwise.
- Warning triggers only when all routes are malformed (and can be suppressed).
- Routing docs match implementation (prefix/path; no `bd routes`).
- All new routing tests pass on upstream main base.

## Test Plan
- Unit: `go test ./internal/routing/...` (covers new routing tests).
- Smoke:
  - `BD_DEBUG_ROUTING=1 bd create "test" --dry-run`
  - `bd create --id=bd-abc123 --dry-run` with routes.jsonl present
  - `bd create --repo /tmp/repo --dry-run` to ensure no auto‑routing occurs

## Rollout / Migration
- No migrations required.
- This is a behavioral improvement; no data format changes.

## Risks
- Over‑routing due to ambiguous prefixes; mitigated by explicit `--repo` override and `--rig/--prefix` flags.
- Debug logs are noisy if `BD_DEBUG_ROUTING=1` is widely enabled; off by default.

## Implementation Checklist
- [ ] Add routing docs and debug logging in `internal/routing/routes.go`.
- [ ] Add `AutoDetectTargetRig` helper.
- [ ] Add auto‑routing in `cmd/bd/create.go` with `--repo` guard.
- [ ] Add routing tests in `internal/routing/routing_test.go`.
- [ ] Rewrite routing docs in `website/docs/multi-agent/routing.md`.
- [ ] Run `go test ./internal/routing/...`.

