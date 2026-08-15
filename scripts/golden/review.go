package main

// review is the mode that makes accepting a golden a DECISION rather than a side effect.
//
// THE HOLE `-update` LEFT. The header above names review erosion as the problem and offers
// three countermeasures, and they are real — but the update path still goes straight from
// "I think something changed" to "every golden on disk now says what the code says", and
// hands back a `--stat`. A `--stat` is a COUNT, not a review: "3 files changed, 218
// insertions" is compatible with a deliberate format change and with a semantics regression,
// and it reads the same either way. Nothing in that path ever shows the author a failing
// test, so the failure — the one artifact that says which behaviour moved — is skipped.
//
// Measured on this repository: the session validation loop recorded
// `UPDATE_GOLDENS=1 go test ./internal/difftest -run TestGolden` as check 11. That is the
// REGENERATE command written down as the CHECK. A check that rewrites its own expectations
// cannot fail, so it had been passing by construction for as long as it had been listed.
//
// WHAT REVIEW DOES DIFFERENTLY, in the order that matters:
//
//  1. VERIFY FIRST. If the goldens are current it says so and stops — you never reach the
//     rewrite without a failure to justify it.
//  2. Materialise the proposals, then show the actual DIFF per file, not a stat line.
//  3. ACCEPT NOTHING BY DEFAULT. The tree is left exactly as found unless you name what to
//     keep. Reviewing is free; changing the contract is the deliberate act.
//  4. Re-verify afterwards and report honestly — a partial accept leaves the suite RED, and
//     saying so is the point.
//
// The revert in step 3 is destructive, so it is fenced: the run refuses to start if any
// golden is already dirty, and it only ever reverts paths whose state it changed itself
// during this run. It cannot touch work it did not create.

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ctoforaday/special-circumstances/scripts/internal/gitx"
)

// diffBudget caps the per-file diff. A review that prints ten thousand lines is the same
// rubber stamp by another route — the cap is a prompt to open the file, not a summary of it.
const diffBudget = 120

// goldenState is what git says about the goldens at one instant.
type goldenState struct {
	modified  map[string]bool
	untracked map[string]bool
}

func (s goldenState) dirty() []string {
	var all []string
	for f := range s.modified {
		all = append(all, f)
	}
	for f := range s.untracked {
		all = append(all, f)
	}
	sort.Strings(all)
	return all
}

// readGoldenState asks git which goldens are modified and which are new.
//
// Errors are ERRORS, never an empty set: a state read that silently returned "nothing is
// dirty" would let the precondition below pass on exactly the trees it exists to refuse
// (gitx's package comment is about this same failure).
func readGoldenState(root string) (goldenState, error) {
	mod, err := gitx.Run(root, "diff", "--name-only", "--", "*.golden")
	if err != nil {
		return goldenState{}, fmt.Errorf("reading modified goldens: %w", err)
	}
	unt, err := gitx.Run(root, "ls-files", "--others", "--exclude-standard", "--", "*.golden")
	if err != nil {
		return goldenState{}, fmt.Errorf("reading new goldens: %w", err)
	}
	s := goldenState{modified: map[string]bool{}, untracked: map[string]bool{}}
	for _, f := range gitx.Lines(mod) {
		s.modified[f] = true
	}
	for _, f := range gitx.Lines(unt) {
		s.untracked[f] = true
	}
	return s, nil
}

// proposal is one golden the update run moved.
type proposal struct {
	path  string
	isNew bool
}

// proposalsBetween reports what THIS RUN changed, by set difference against the state
// captured before it. Anything already dirty beforehand is excluded — the tool must never
// claim, revert, or delete an edit it did not make.
func proposalsBetween(before, after goldenState) []proposal {
	var out []proposal
	for f := range after.modified {
		if !before.modified[f] {
			out = append(out, proposal{path: f})
		}
	}
	for f := range after.untracked {
		if !before.untracked[f] {
			out = append(out, proposal{path: f, isNew: true})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].path < out[j].path })
	return out
}

// renderDiff shows what a single golden proposes to become.
func renderDiff(root string, p proposal) string {
	if p.isNew {
		body, err := os.ReadFile(filepath.Join(root, p.path))
		if err != nil {
			return fmt.Sprintf("  (new golden; could not be read: %v)\n", err)
		}
		return budget(fmt.Sprintf("  NEW FILE, %d line(s):\n%s",
			len(gitx.Lines(string(body))), indent(string(body))))
	}
	// --no-ext-diff so a developer's difftool cannot turn a review into an interactive
	// session; -U2 because a golden's context is other golden lines, not code.
	d, err := gitx.Run(root, "diff", "--no-ext-diff", "-U2", "--", p.path)
	if err != nil {
		return fmt.Sprintf("  (diff unavailable: %v)\n", err)
	}
	return budget(indent(d))
}

func indent(s string) string {
	var b strings.Builder
	for _, l := range strings.Split(strings.TrimRight(s, "\n"), "\n") {
		fmt.Fprintf(&b, "  %s\n", l)
	}
	return b.String()
}

func budget(s string) string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	if len(lines) <= diffBudget {
		return strings.Join(lines, "\n") + "\n"
	}
	kept := strings.Join(lines[:diffBudget], "\n")
	return kept + fmt.Sprintf("\n  … %d more line(s) — this one is too big to review here; "+
		"open it with `git diff -- <path>` before accepting\n", len(lines)-diffBudget)
}

// revert undoes one proposal: a tracked golden goes back to HEAD, a new one is removed.
func revert(root string, p proposal) error {
	if p.isNew {
		return os.Remove(filepath.Join(root, p.path))
	}
	_, err := gitx.Run(root, "checkout", "--", p.path)
	return err
}

// acceptSet turns the -accept flag into a matcher. A bare filename matches any path ending
// in it, so a reviewer can type `debate.golden` rather than the full testdata path — but an
// entry matching NOTHING is an error, never a silent no-op, because "I accepted it and it
// reverted anyway" is the failure this whole mode exists to prevent.
func acceptSet(spec string, props []proposal) (map[string]bool, error) {
	keep := map[string]bool{}
	if strings.TrimSpace(spec) == "" {
		return keep, nil
	}
	for _, want := range strings.Split(spec, ",") {
		want = strings.TrimSpace(want)
		if want == "" {
			continue
		}
		hits := 0
		for _, p := range props {
			if p.path == want || strings.HasSuffix(p.path, "/"+want) {
				keep[p.path] = true
				hits++
			}
		}
		if hits == 0 {
			return nil, fmt.Errorf("-accept %q matched none of the %d proposed golden(s); "+
				"nothing was kept, so re-run with a name from the list above", want, len(props))
		}
	}
	return keep, nil
}

// reviewPlan is the decision, separated from the doing so it can be tested without a
// git tree: given proposals and the accept flags, what is kept and what goes back.
func reviewPlan(props []proposal, keep map[string]bool, all bool) (accepted, reverted []proposal) {
	for _, p := range props {
		if all || keep[p.path] {
			accepted = append(accepted, p)
			continue
		}
		reverted = append(reverted, p)
	}
	return accepted, reverted
}

// reviewMode is the whole posture in one function. Returns the process exit code.
func reviewMode(root, accept string, acceptAll bool, reason string) int {
	if acceptAll && strings.TrimSpace(reason) == "" {
		fmt.Fprintln(os.Stderr, "golden: -accept-all requires -reason. Accepting every golden at once is the "+
			"move most likely to wave a regression through, so it costs one sentence.")
		return 2
	}

	// ---- precondition: the tool must be able to tell its own writes from yours ----
	before, err := readGoldenState(root)
	if err != nil {
		fmt.Fprintln(os.Stderr, "golden:", err)
		return 1
	}
	if d := before.dirty(); len(d) > 0 {
		fmt.Fprintf(os.Stderr, "golden: %d golden(s) are already modified or untracked:\n", len(d))
		for _, f := range d {
			fmt.Fprintf(os.Stderr, "  %s\n", f)
		}
		fmt.Fprintln(os.Stderr, "\nReview reverts everything it is not told to keep, and it must never revert an "+
			"edit it did not make. Commit or stash these first.")
		return 2
	}

	// ---- 1. VERIFY. No failure, no rewrite. ----
	fmt.Print("══ VERIFY ══\n")
	if !runLegs(root, false) {
		fmt.Print("\ngolden: OK — every golden matches. Nothing to review.\n")
		return 0
	}

	// ---- 2. materialise what the code would record ----
	fmt.Print("\n══ PROPOSE ══ (re-recording so the diff can be read; nothing is kept yet)\n")
	runLegs(root, true)

	after, err := readGoldenState(root)
	if err != nil {
		fmt.Fprintln(os.Stderr, "golden:", err)
		return 1
	}
	props := proposalsBetween(before, after)
	if len(props) == 0 {
		// The verify leg failed and yet no golden moved. That is a REAL test failure, not
		// staleness, and calling it "nothing to review" would send the author to the wrong
		// problem entirely.
		fmt.Fprint(os.Stderr, "\ngolden: the suite FAILED but no golden changed.\n"+
			"That is a genuine test failure, not a stale contract — read the verify output above.\n")
		return 1
	}

	// ---- 3. show the diffs ----
	fmt.Printf("\n══ REVIEW ══ %d golden(s) propose to change\n", len(props))
	for _, p := range props {
		kind := "modified"
		if p.isNew {
			kind = "NEW"
		}
		fmt.Printf("\n── %s (%s)\n%s", p.path, kind, renderDiff(root, p))
	}

	keep, err := acceptSet(accept, props)
	if err != nil {
		// Revert before refusing: a mistyped -accept must not leave the tree rewritten.
		for _, p := range props {
			_ = revert(root, p)
		}
		fmt.Fprintln(os.Stderr, "\ngolden:", err)
		fmt.Fprintln(os.Stderr, "the tree has been restored; nothing was kept.")
		return 2
	}
	accepted, reverted := reviewPlan(props, keep, acceptAll)

	// ---- 4. keep only what was named ----
	for _, p := range reverted {
		if err := revert(root, p); err != nil {
			fmt.Fprintf(os.Stderr, "golden: could not revert %s: %v\n", p.path, err)
			return 1
		}
	}

	fmt.Print("\n══ RESULT ══\n")
	if len(accepted) == 0 {
		fmt.Printf("accepted 0 of %d — the tree is exactly as you found it.\n", len(props))
		fmt.Print("Re-run with -accept <name>[,<name>…] for the ones you have read, " +
			"or -accept-all -reason \"…\" if the whole batch is one deliberate change.\n")
		fmt.Print("\ngolden: STALE — goldens still do not match the code.\n")
		return 1
	}
	fmt.Printf("accepted %d, reverted %d.\n", len(accepted), len(reverted))
	for _, p := range accepted {
		fmt.Printf("  kept    %s\n", p.path)
	}
	for _, p := range reverted {
		fmt.Printf("  reverted %s\n", p.path)
	}
	if acceptAll {
		fmt.Printf("\nreason: %s\n", strings.TrimSpace(reason))
	}

	// ---- 5. re-verify, and be honest about a partial accept ----
	fmt.Print("\n══ RE-VERIFY ══\n")
	if runLegs(root, false) {
		fmt.Fprintf(os.Stderr, "\ngolden: STILL FAILING — expected while %d proposal(s) stay reverted.\n"+
			"Accept them too, or fix the code so they stop being proposed.\n", len(reverted))
		return 1
	}
	fmt.Print("\ngolden: OK — every accepted golden is recorded and the suite is green.\n" +
		"Commit the testdata change ON ITS OWN, separately from the code that caused it.\n")
	return 0
}
