# Routing Improvements Checklist

## Planning
- [x] Confirm scope: prefix-based routing only (no patterns, no CLI commands)
- [x] Confirm `--repo` override semantics

## Code Changes
- [x] Add/restore AutoDetectTargetRig helper in `internal/routing/routes.go`
- [x] Add debug logging + malformed routes warning in `LoadRoutes`
- [x] Improve town-root discovery logging for symlinked `.beads`
- [x] Add auto-routing in `cmd/bd/create.go`
- [x] Skip auto-routing when `--repo` is explicitly passed

## Tests
- [x] Add LoadRoutes malformed/empty/missing file tests
- [x] Add ResolveBeadsDirForRig error-path tests
- [x] Add ResolveBeadsDirForID edge-case tests
- [x] Add warning/suppression tests for malformed routes

## Docs
- [x] Update `website/docs/multi-agent/routing.md` to prefix-based description
- [x] Ensure troubleshooting mentions `BD_DEBUG_ROUTING` and `BD_QUIET_ROUTING`

## Verification
- [x] `go test ./internal/routing/...`
- [ ] Optional smoke: `BD_DEBUG_ROUTING=1 bd create --dry-run`
