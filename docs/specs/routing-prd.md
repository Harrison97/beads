# PRD: Routing Improvements (Prefix-Based Auto-Routing)

Owner: Harrison97 + Codex
Status: Draft
Branch: routing

## Problem
Routing is documented as prefix-based and intended to be transparent for multi-rig setups, but upstream main lacks:
- Automatic routing on create based on configured prefix.
- Clear, actionable debug output for malformed routes.
- Comprehensive error-path test coverage.
- Documentation that precisely matches implementation.

These gaps lead to silent failures and confusion in multi-rig setups.

## Goals
1) Restore and clarify prefix-based auto-routing for `bd create`.
2) Add observability for route parsing and routing decisions.
3) Lock down routing behavior with unit tests.
4) Align docs with actual behavior.

## Non-Goals
- No pattern/priority routing.
- No `bd routes` CLI.
- No changes to routing model beyond prefix/path.

## Users
- Contributors in Gas Town-style multi-rig environments.
- Maintainers troubleshooting routing issues.

## Key Behaviors
- Auto-route on `bd create` when:
  - An explicit ID with prefix is provided, or
  - A configured prefix is detected, and routes.jsonl points elsewhere.
- Explicit `--repo` always overrides auto-routing.
- Debug logs show route parsing and routing decisions when `BD_DEBUG_ROUTING=1`.
- Warn on completely broken routes.jsonl unless `BD_QUIET_ROUTING=1`.

## Success Criteria
- Auto-routing works without `--rig` or `--prefix`.
- `BD_DEBUG_ROUTING=1` provides enough info to diagnose route failures.
- Routing tests cover malformed lines, missing files, dot-path, redirects.
- Docs describe prefix/path routing accurately.

## Risks
- Auto-routing might surprise users who expect `--repo` to win. Mitigation: skip auto-routing if `--repo` explicitly set.
- Town discovery could be brittle; mitigation: derive from `dbPath` when available.

## Metrics (Qualitative)
- Fewer routing “silent failure” reports.
- Reduced confusion from docs mismatch.

