# Routing Improvements Checklist

## Planning
- [ ] Confirm scope: prefix-based routing only (no patterns, no CLI commands)
- [ ] Confirm `--repo` override semantics

## Code Changes
- [ ] Add/restore AutoDetectTargetRig helper in `internal/routing/routes.go`
- [ ] Add debug logging + malformed routes warning in `LoadRoutes`
- [ ] Improve town-root discovery logging for symlinked `.beads`
- [ ] Add auto-routing in `cmd/bd/create.go`
- [ ] Skip auto-routing when `--repo` is explicitly passed

## Tests
- [ ] Add LoadRoutes malformed/empty/missing file tests
- [ ] Add ResolveBeadsDirForRig error-path tests
- [ ] Add ResolveBeadsDirForID edge-case tests
- [ ] Add warning/suppression tests for malformed routes

## Docs
- [ ] Update `website/docs/multi-agent/routing.md` to prefix-based description
- [ ] Ensure troubleshooting mentions `BD_DEBUG_ROUTING` and `BD_QUIET_ROUTING`

## Verification
- [ ] `go test ./internal/routing/...`
- [ ] Optional smoke: `BD_DEBUG_ROUTING=1 bd create --dry-run`

