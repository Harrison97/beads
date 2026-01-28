package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/steveyegge/beads/internal/routing"
)

func townBeadsDirFromCwd() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("working directory: %w", err)
	}

	beadsDir := filepath.Join(cwd, ".beads")
	return routing.FindTownBeadsDir(beadsDir)
}
