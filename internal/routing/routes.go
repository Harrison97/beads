// Package routing provides prefix-based routing for multi-repository beads setups.
//
// # Multi-repo Architecture
//
// A multi-repo setup allows multiple independent projects to share a single
// town root directory. Each repo maintains its own .beads directory with its own
// database, while a shared routes.jsonl at the town level enables cross-repo references.
//
// Example structure:
//
//	~/town/                       # Town root (contains mayor/town.json)
//	├── mayor/
//	│   └── town.json             # Town configuration
//	├── .beads/
//	│   └── routes.jsonl          # Routing configuration
//	├── project-a/                # Rig with prefix "pa-"
//	│   └── .beads/
//	└── project-b/                # Rig with prefix "pb-"
//	    └── .beads/
//
// The routes.jsonl file maps prefixes to rig paths:
//
//	{"prefix": "pa-", "path": "project-a"}
//	{"prefix": "pb-", "path": "project-b"}
//
// # Symlink Handling
//
// The routing system correctly handles symlinked .beads directories. When .beads
// is a symlink, functions like findTownRootFromCWD use the current working directory
// rather than the resolved symlink path to determine the actual town root.
package routing

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// RoutesFileName is the name of the routes configuration file
const RoutesFileName = "routes.jsonl"

// Route represents a prefix-to-path routing rule
type Route struct {
	Prefix string `json:"prefix"` // Issue ID prefix (e.g., "gt-")
	Path   string `json:"path"`   // Relative path to .beads directory
}

// LoadRoutes loads routes from routes.jsonl in the given beads directory.
// Returns an empty slice if the file doesn't exist.
func LoadRoutes(beadsDir string) ([]Route, error) {
	routesPath := filepath.Join(beadsDir, RoutesFileName)
	debugRouting := os.Getenv("BD_DEBUG_ROUTING") != ""

	if debugRouting {
		fmt.Fprintf(os.Stderr, "[routing] LoadRoutes: loading from %s\n", routesPath)
	}

	file, err := os.Open(routesPath) //nolint:gosec // routesPath is constructed from known beadsDir
	if err != nil {
		if os.IsNotExist(err) {
			if debugRouting {
				fmt.Fprintf(os.Stderr, "[routing] LoadRoutes: file does not exist (not an error)\n")
			}
			return nil, nil // No routes file is not an error
		}
		if debugRouting {
			fmt.Fprintf(os.Stderr, "[routing] LoadRoutes: failed to open file: %v\n", err)
		}
		return nil, err
	}
	defer file.Close()

	var routes []Route
	var lineNum int
	var skippedLines int
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue // Skip empty lines and comments
		}

		var route Route
		if err := json.Unmarshal([]byte(line), &route); err != nil {
			if debugRouting {
				fmt.Fprintf(os.Stderr, "[routing] LoadRoutes: skipping malformed line %d: %s (error: %v)\n", lineNum, line, err)
			}
			skippedLines++
			continue
		}
		if route.Prefix != "" && route.Path != "" {
			routes = append(routes, route)
		} else if debugRouting {
			fmt.Fprintf(os.Stderr, "[routing] LoadRoutes: skipping line %d with empty prefix or path: %s\n", lineNum, line)
			skippedLines++
		}
	}

	if debugRouting {
		fmt.Fprintf(os.Stderr, "[routing] LoadRoutes: parsed %d valid routes, skipped %d lines\n", len(routes), skippedLines)
	}

	// Warn if routes.jsonl exists but has no valid routes after parsing
	// This catches completely broken configs without being noisy for minor issues.
	// Can be disabled with BD_QUIET_ROUTING=1.
	if skippedLines > 0 && len(routes) == 0 && os.Getenv("BD_QUIET_ROUTING") == "" {
		fmt.Fprintf(os.Stderr, "warning: %s has %d malformed line(s) and no valid routes\n", routesPath, skippedLines)
		fmt.Fprintf(os.Stderr, "  hint: set BD_DEBUG_ROUTING=1 for details, or BD_QUIET_ROUTING=1 to suppress this warning\n")
	}

	return routes, scanner.Err()
}

func ExtractPrefix(id string) string {
	prefix := ExtractIssuePrefix(id)
	if prefix == "" {
		return ""
	}
	return prefix + "-"
}

// NormalizePrefix trims a trailing hyphen to normalize comparisons.
func NormalizePrefix(prefix string) string {
	return strings.TrimSuffix(prefix, "-")
}

// PrefixesEqual compares two prefixes, ignoring a trailing hyphen.
// Returns false if either prefix is empty after normalization.
func PrefixesEqual(a, b string) bool {
	na := NormalizePrefix(a)
	nb := NormalizePrefix(b)
	if na == "" || nb == "" {
		return false
	}
	return na == nb
}
