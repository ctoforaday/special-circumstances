package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ctoforaday/special-circumstances/scripts/internal/gitx"
)

// failureGuidance is what an author sees when the gate fails. It states the evidence
// rather than the rule, because the rule on its own reads as bureaucracy and the evidence
// does not.
const failureGuidance = `
Every protocol patch that shipped before this check recurred in adjacent form:
lens labels patched, blue-lane footnotes bit next run; blue-reads-transcript
patched, the lossy gap summary bit next run; grade enums widened, the mass
mapping distorted next run; friction-to-file shipped, lens seats still had no
write path. Four for four. The fixes were correct; they were INSTANCE fixes.

Add to any commit in this branch:

  Rule-Class: <slug from feov-memory/protocol-class-registry.md>
  Sibling-Sweep: <which adjacent seats or surfaces carry this same duty, and
                  what you found at each — including "none do", if that is true>
`

func main() {
	base := flag.String("base", "origin/main", "base ref to compare against")
	flag.Parse()

	root, err := gitx.Root()
	if err != nil {
		fmt.Fprintln(os.Stderr, "rule-sweep:", err)
		os.Exit(1)
	}

	files, messages, dirty, err := collect(*base, root)
	if err != nil {
		// A gate that cannot read its range must FAIL, not pass. The .mjs original
		// returned "" for any git error, so a bad base ref produced zero changed files
		// and a cheerful "no protocol surface touched".
		fmt.Fprintf(os.Stderr, "rule-sweep: cannot read the range against %q: %v\n", *base, err)
		os.Exit(1)
	}

	registry, err := os.ReadFile(filepath.Join(root, "feov-memory", "protocol-class-registry.md"))
	if err != nil {
		// An absent registry means every named class is "unknown", which would fail
		// every protocol patch for a reason the author cannot act on. Say so plainly.
		fmt.Fprintf(os.Stderr, "rule-sweep: cannot read feov-memory/protocol-class-registry.md: %v\n", err)
		os.Exit(1)
	}

	result := evaluate(input{files: files, messages: messages, dirty: dirty, classes: knownClasses(string(registry))})

	if result.skipped {
		fmt.Println("rule-sweep: no protocol surface touched — nothing to sweep")
		return
	}

	// EXIT 3 = "this gate did not measure anything", distinct from 0 (measured, clean)
	// and 1 (measured, failed). It is a code rather than a phrase on purpose: a runner
	// keying on prose is the shape this whole gate set keeps finding (facts-are-fields).
	// It cannot fire in CI, where the tree is always clean.
	if result.unmeasurable {
		fmt.Fprintf(os.Stderr, "rule-sweep: NOT MEASURED — %s\n\n", result.reason)
		for _, f := range result.touched {
			fmt.Fprintf(os.Stderr, "  %s\n", f)
		}
		os.Exit(3)
	}
	fmt.Printf("rule-sweep: %d protocol surface(s) changed:\n", len(result.touched))
	for _, f := range result.touched {
		fmt.Printf("  %s\n", f)
	}

	if result.ok {
		fmt.Printf("\nrule-sweep: OK — class %s\n", strings.Join(result.named, ", "))
		for _, s := range result.sweeps {
			fmt.Printf("  sweep: %s\n", s)
		}
		return
	}

	fmt.Fprint(os.Stderr, "\nrule-sweep: FAILED — a rule patch must name its class and sweep its siblings.\n\n")
	for _, p := range result.problems {
		fmt.Fprintf(os.Stderr, "  - %s\n", p)
	}
	fmt.Fprint(os.Stderr, failureGuidance)
	os.Exit(1)
}
