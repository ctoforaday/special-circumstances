package fuzz

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/claimcount"
)

// anchorEvidence renders the two sides of the anchor bijection into the failure text.
//
// #645 IS THE CASE THIS EXISTS FOR. The consistency oracle reported "the record carries finding
// id f-005dff8e and report.md has no marker for it" — a true statement that names one id and
// nothing else. Answering the only question that follows (was the marker never written, or
// written and then removed?) needs the report and the record side by side, and the sweep's
// promise to keep the failing run directory is worth nothing in the environment where a
// 1-in-160 violation is actually seen: a CI container that is reclaimed before anyone reads
// the log.
//
// So the evidence goes where the log goes. It is deliberately compact — the marker ids present
// in the report, not the report — because the whole point is that it survives being pasted into
// an issue.
func anchorEvidence(runDir string) string {
	rep, err := os.ReadFile(filepath.Join(runDir, "blue", "report.md"))
	if err != nil {
		return fmt.Sprintf("\n  [anchor evidence unavailable: %v]", err)
	}
	ids := claimcount.ProtectedAnchorIDs(string(rep))
	sort.Strings(ids)
	var b strings.Builder
	fmt.Fprintf(&b, "\n  report.md carries %d protected anchor(s)", len(ids))
	if len(ids) > 0 {
		fmt.Fprintf(&b, ": %s", strings.Join(ids, " "))
	}
	fmt.Fprintf(&b, "\n  report.md is %d bytes at %s", len(rep), filepath.Join(runDir, "blue", "report.md"))
	return b.String()
}
