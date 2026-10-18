package checkpointrestore

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/checkpoint"
)

// A heading that nearly names the loop reaches the digest as an unresolved
// surface, so the session is told its loop is unread rather than shown nothing.
func TestANearMissLoopHeadingReachesTheDigest(t *testing.T) {
	n := checkpoint.Parse("---\nschema: 3\n---\n## validation-loop\n1. `go test ./...` · re-armed by: tools/\n")
	paths, unresolved := watchTargets(n, t.TempDir())
	if len(paths) != 0 || len(unresolved) != 1 || !strings.Contains(unresolved[0], `"## validation-loop"`) {
		t.Errorf("paths %q, unresolved %q", paths, unresolved)
	}
}

// Low-value sections are recognised by the same heading rule: a suffixed
// "Files touched" is still left out of the fallback digest.
func TestASuffixedLowValueSectionIsStillLowValue(t *testing.T) {
	if !lowValue("Files touched ← mid-edit state") || lowValue("Files") {
		t.Error("lowValue does not follow the heading rule")
	}
}
