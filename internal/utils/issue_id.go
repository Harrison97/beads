package utils

import (
	"fmt"
	"strings"

	"github.com/steveyegge/beads/internal/routing"
)

// ExtractIssuePrefix extracts the prefix from an issue ID like "bd-123" -> "bd"
// Uses the last hyphen before a numeric or hash-like suffix:
//   - "beads-vscode-1" -> "beads-vscode" (numeric suffix)
//   - "web-app-a3f8e9" -> "web-app" (hash suffix with digits)
//   - "my-cool-app-123" -> "my-cool-app" (numeric suffix)
//   - "bd-a3f" -> "bd" (3-char hash)
//
// Falls back to first hyphen when suffix looks like an English word (4+ chars, no digits):
//   - "vc-baseline-test" -> "vc" (word-like suffix: "test" is not a hash)
//   - "bd-multi-part-id" -> "bd" (word-like suffix: "id" is too short but "part-id" path)
//
// This distinguishes hash IDs (which may contain letters but have digits or are 3 chars)
// from multi-part IDs where the suffix after the first hyphen is the entire ID.
func ExtractIssuePrefix(issueID string) string {
	return routing.ExtractIssuePrefix(issueID)
}

// ExtractIssueNumber extracts the number from an issue ID like "bd-123" -> 123
func ExtractIssueNumber(issueID string) int {
	idx := strings.LastIndex(issueID, "-")
	if idx < 0 || idx == len(issueID)-1 {
		return 0
	}
	var num int
	_, _ = fmt.Sscanf(issueID[idx+1:], "%d", &num)
	return num
}
