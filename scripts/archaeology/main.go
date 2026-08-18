// Command archaeology fails a change that ADDS prose about a thing that no longer exists to a
// surface an agent reads.
//
// Dev tooling for this repository only. Nothing here ships to an installing project.
//
// See sweep.go for the rule, the scope and what this deliberately does not reach.
//
// Usage:
//
//	go run ./archaeology              scan the diff against origin/main
//	go run ./archaeology -base main   scan against another ref
//	go run ./archaeology -selftest    prove the matcher can still fire
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ctoforaday/special-circumstances/scripts/internal/gitx"
)

const guidance = `
The registry class is archaeology-in-live-surface, and its sweep question decides
each site:

  could a seat ACT on this sentence today?

If the thing it names is gone, no action exists and the git history is where it
belongs. Keep the rule, drop the obituary — "X used to lose the reason twice
over" becomes "X loses the reason twice over", which is the form a reader acts on.

There is deliberately NO allowlist. A guard whose own exemptions are hand-kept
reproduces the defect one level up (facts-are-fields). If a phrase is genuinely
needed, the sentence can be written without it.
`

func main() {
	base := flag.String("base", "origin/main", "base ref to compare against")
	selftest := flag.Bool("selftest", false, "prove the matcher fires on known archaeology and stays quiet on live prose")
	flag.Parse()

	if *selftest {
		if err := runSelftest(); err != nil {
			fmt.Fprintln(os.Stderr, "archaeology: SELFTEST FAILED —", err)
			os.Exit(1)
		}
		fmt.Println("archaeology: selftest OK — the matcher fires on archaeology and is quiet on live prose")
		return
	}

	root, err := gitx.Root()
	if err != nil {
		fmt.Fprintln(os.Stderr, "archaeology:", err)
		os.Exit(1)
	}

	// A gate that cannot read its range must FAIL, not pass — a bad base ref returning "" would
	// yield zero added lines and a cheerful clean report, which is the plausible zero this
	// repository keeps finding.
	diff, err := gitx.Run(root, "diff", "--unified=0", *base+"...HEAD")
	if err != nil {
		fmt.Fprintf(os.Stderr, "archaeology: cannot read the range against %q: %v\n", *base, err)
		os.Exit(1)
	}
	// The working tree too, so an uncommitted change is caught before it is committed rather
	// than after. Both are ADDED-line scans, so a pre-existing site never fails a change that
	// merely touched its file.
	dirty, err := gitx.Run(root, "diff", "--unified=0", "HEAD")
	if err != nil {
		fmt.Fprintf(os.Stderr, "archaeology: cannot read the working tree: %v\n", err)
		os.Exit(1)
	}

	found := append(scan(diff), scan(dirty)...)
	if len(found) == 0 {
		fmt.Println("archaeology: OK — no archaeology added to an agent-facing surface")
		return
	}

	fmt.Fprintf(os.Stderr, "archaeology: FAILED — %d line(s) add prose about something that no longer exists:\n\n", len(found))
	for _, f := range found {
		fmt.Fprintf(os.Stderr, "  - %s\n", f)
	}
	fmt.Fprint(os.Stderr, guidance)
	os.Exit(1)
}

// runSelftest asks the question the gate cannot ask about itself: can it still fire?
//
// A matcher whose patterns stop matching reports zero findings, which is byte-identical to a
// clean diff. That is the exact defect class this gate polices, so it is checked here rather
// than assumed — the same argument `mutate -selftest` makes one directory over.
func runSelftest() error {
	const bad = `+++ b/plugins/x/agents/seat.md
+the ` + "`blue confidence`" + ` verb is RETIRED (0.54.0), so hold it as a discipline.
+It used to be armed only by passing --bin-dir, which nothing did.
+That mode no longer exists and the file it wrote was removed.
`
	got := scan(bad)
	if len(got) < 3 {
		return fmt.Errorf("fired on %d of 3 planted archaeology lines — the patterns have gone stale, and a stale matcher reports a clean board", len(got))
	}

	// LIVE PROSE MUST STAY QUIET, or the gate trains authors to route around it. Each of these
	// states a current fact and a seat can act on every one.
	const good = `+++ b/plugins/x/agents/seat.md
+A realized risk is no longer a probability: it contributes 0 to mass.
+The report does not carry a retired claim; substance leaves through the retire verb.
+Read the whole list before your first act.
`
	if got := scan(good); len(got) != 0 {
		return fmt.Errorf("fired on live prose (%d false positive(s)): %v", len(got), got)
	}

	// OUT-OF-SCOPE FILES MUST STAY QUIET. Go comments are deliberately not policed; a gate that
	// silently widened its own scope would fail changes for a rule nobody stated.
	const code = `+++ b/plugins/x/tools/internal/record/record.go
+// This used to re-derive the round by regex, returning 0 on a miss.
`
	if got := scan(code); len(got) != 0 {
		return fmt.Errorf("fired on a Go comment, which is out of scope by design: %v", got)
	}
	return nil
}
