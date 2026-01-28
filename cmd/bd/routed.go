package main

import (
	"context"
	"path/filepath"

	"github.com/steveyegge/beads/internal/routing"
	"github.com/steveyegge/beads/internal/storage"
	"github.com/steveyegge/beads/internal/storage/factory"
	"github.com/steveyegge/beads/internal/types"
	"github.com/steveyegge/beads/internal/utils"
)

// RoutedResult contains the result of a routed issue lookup
type RoutedResult struct {
	Issue      *types.Issue
	Store      storage.Storage // The store that contains this issue (may be routed)
	Routed     bool            // true if the issue was found via routing
	ResolvedID string          // The resolved (full) issue ID
	closeFn    func()          // Function to close routed storage (if any)
}

// Close closes any routed storage. Safe to call if Routed is false.
func (r *RoutedResult) Close() {
	if r.closeFn != nil {
		r.closeFn()
	}
}

// resolveAndGetIssueWithRouting resolves a partial ID and gets the issue,
// using routes.jsonl for prefix-based routing if needed.
// This enables cross-repo issue lookups (e.g., `bd show gt-xyz` from ~/gt).
//
// The resolution happens in the correct store based on the ID prefix.
// Returns a RoutedResult containing the issue, resolved ID, and the store to use.
// The caller MUST call result.Close() when done to release any routed storage.
func resolveAndGetIssueWithRouting(ctx context.Context, localStore storage.Storage, id string) (*RoutedResult, error) {
	return getIssueWithRouting(ctx, localStore, id, true)
}

// resolveAndGetFromStore resolves a partial ID and gets the issue from a specific store.
func resolveAndGetFromStore(ctx context.Context, s storage.Storage, id string, routed bool) (*RoutedResult, error) {
	// First, resolve the partial ID
	resolvedID, err := utils.ResolvePartialID(ctx, s, id)
	if err != nil {
		return nil, err
	}

	// Then get the issue
	issue, err := s.GetIssue(ctx, resolvedID)
	if err != nil {
		return nil, err
	}
	if issue == nil {
		return nil, nil
	}

	return &RoutedResult{
		Issue:      issue,
		Store:      s,
		Routed:     routed,
		ResolvedID: resolvedID,
	}, nil
}

// getIssueWithRouting tries to get an issue from the local store first,
// then falls back to checking routes.jsonl for prefix-based routing.
// This enables cross-repo issue lookups (e.g., `bd show gt-xyz` from ~/gt).
//
// Returns a RoutedResult containing the issue and the store to use for related queries.
// The caller MUST call result.Close() when done to release any routed storage.
func getIssueWithRouting(ctx context.Context, localStore storage.Storage, id string, resolvePartial bool) (*RoutedResult, error) {
	// Step 1: If routing is available, try routed store first when resolving partial IDs.
	if dbPath == "" {
		return getLocalIssue(ctx, localStore, id, resolvePartial)
	}

	beadsDir := filepath.Dir(dbPath)
	// Use GetRoutedStorageWithOpener with factory to respect backend configuration (bd-m2jr)
	routedStorage, routeErr := routing.GetRoutedStorageWithOpener(ctx, id, beadsDir, factory.NewFromConfig)
	if routeErr != nil || routedStorage == nil {
		// No routing found or error - fall back to local store
		return getLocalIssue(ctx, localStore, id, resolvePartial)
	}

	// Step 2: Try the routed storage
	result, err := resolveAndGetFromStore(ctx, routedStorage.Storage, id, true)
	if err != nil || result == nil {
		_ = routedStorage.Close()
		if err != nil {
			return nil, err
		}
		// Fall back to local store if not found in routed store.
		return getLocalIssue(ctx, localStore, id, resolvePartial)
	}
	result.closeFn = func() { _ = routedStorage.Close() }
	return result, nil
}

func getLocalIssue(ctx context.Context, localStore storage.Storage, id string, resolvePartial bool) (*RoutedResult, error) {
	if resolvePartial {
		return resolveAndGetFromStore(ctx, localStore, id, false)
	}
	issue, err := localStore.GetIssue(ctx, id)
	if err == nil && issue != nil {
		return &RoutedResult{
			Issue:      issue,
			Store:      localStore,
			Routed:     false,
			ResolvedID: id,
		}, nil
	}
	return &RoutedResult{
		Issue:      issue,
		Store:      localStore,
		Routed:     false,
		ResolvedID: id,
	}, err
}

// getRoutedStoreForID returns a storage connection for an issue ID if routing is needed.
// Returns nil if no routing is needed (issue should be in local store).
// The caller is responsible for closing the returned storage.
func getRoutedStoreForID(ctx context.Context, id string) (*routing.RoutedStorage, error) {
	if dbPath == "" {
		return nil, nil
	}

	beadsDir := filepath.Dir(dbPath)
	// Use GetRoutedStorageWithOpener with factory to respect backend configuration (bd-m2jr)
	return routing.GetRoutedStorageWithOpener(ctx, id, beadsDir, factory.NewFromConfig)
}

// needsRouting checks if an ID would be routed to a different beads directory.
// This is used to decide whether to bypass the daemon for cross-repo lookups.
func needsRouting(id string) bool {
	if dbPath == "" {
		return false
	}

	beadsDir := filepath.Dir(dbPath)
	targetDir, routed, err := routing.ResolveBeadsDirForID(context.Background(), id, beadsDir)
	if err != nil || !routed {
		return false
	}

	// Check if the routed directory is different from the current one
	return targetDir != beadsDir
}
