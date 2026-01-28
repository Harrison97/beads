package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/steveyegge/beads/internal/config"
	"github.com/steveyegge/beads/internal/routing"
	"github.com/steveyegge/beads/internal/rpc"
	"github.com/steveyegge/beads/internal/storage"
)

func townBeadsDirFromCwd() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("working directory: %w", err)
	}

	beadsDir := filepath.Join(cwd, ".beads")
	return routing.FindTownBeadsDir(beadsDir)
}

func resolveTargetRigBeadsDir(rigName string) (string, string, error) {
	townBeadsDir, err := townBeadsDirFromCwd()
	if err != nil {
		return "", "", err
	}

	return routing.ResolveBeadsDirForRig(rigName, townBeadsDir)
}

func resolveConfiguredPrefix(ctx context.Context, daemonClient *rpc.Client, store storage.Storage) string {
	// Database config takes precedence over config.yaml.
	if daemonClient != nil {
		resp, err := daemonClient.GetConfig(&rpc.GetConfigArgs{Key: "issue-prefix"})
		if err == nil && resp.Value != "" {
			return resp.Value
		}
		resp, err = daemonClient.GetConfig(&rpc.GetConfigArgs{Key: "issue_prefix"})
		if err == nil && resp.Value != "" {
			return resp.Value
		}
	}

	if store != nil {
		dbPrefix, _ := store.GetConfig(ctx, "issue-prefix")
		if dbPrefix != "" {
			return dbPrefix
		}
		dbPrefix, _ = store.GetConfig(ctx, "issue_prefix")
		if dbPrefix != "" {
			return dbPrefix
		}
	}

	return config.GetString("issue-prefix")
}
