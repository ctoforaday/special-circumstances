package cli

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/proof"
)

// A PAGE STATES WHAT ITS FLAGS DO THERE, AND THE CALL AGREES. Each test below is a guess a seat
// made off a page in the wave B smoke: the page said something that was not so, or left out the
// one fact the guess needed. Each reads the page's statement and then makes the call the statement
// is about, so the sentence cannot outlive the behaviour.

// `--id` IS ON THE ONE VIEW IT SCOPES. Declared on the `show` group it was printed in every view's
// flag block, and two lenses typed `show board --id <gap>` off that block.
func TestOnlyTheChangesViewTakesAGapScope(t *testing.T) {
	root := NewRootFor("red-lens-evidence")
	show := cmdAt(root, []string{"show"})
	if f := show.PersistentFlags().Lookup("id"); f != nil {
		t.Errorf("the show group declares --id for every view beneath it: %s", f.Usage)
	}
	for _, view := range show.Commands() {
		page := seat.HelpText(view)
		has := view.Flags().Lookup("id") != nil
		if want := view.Name() == "changes"; has != want {
			t.Errorf("show %s takes --id = %v, want %v", view.Name(), has, want)
		}
		if has != strings.Contains(page, "--id gap-id") {
			t.Errorf("show %s: the page lists --id = %v and the view takes it = %v", view.Name(), !has, has)
		}
	}

	runDir := actFixture(t)
	gap := onlyID(t, runDir, "G")
	_, stderr, err := runSeen(t, "show", "board", "--id", gap, "--run", runDir, "--seat-id", "red-lens-evidence")
	if err == nil {
		t.Fatal("`show board --id` was accepted: a board read as if it were scoped to one gap")
	}
	if !strings.Contains(stderr, "unknown flag: --id") || !strings.Contains(stderr, "`show changes`") {
		t.Errorf("the refusal does not say --id is not the board's and that `show changes` takes it:\n%s", stderr)
	}
	// THE BARE GROUP IS A VIEW TOO — the seat's work list — and it took --id and ignored it.
	if _, _, err := runSeen(t, "show", "--id", gap, "--run", runDir, "--seat-id", "red-lens-evidence"); err == nil {
		t.Error("bare `show --id` was accepted and answered with the unscoped work list")
	}
	if out, _, err := runSeen(t, "show", "changes", "--id", gap, "--run", runDir, "--seat-id", "red-lens-evidence"); err != nil || !strings.Contains(out, gap) {
		t.Errorf("`show changes --id %s` did not answer for that gap (%v):\n%s", gap, err, out)
	}
}

// THE REPORT'S PAGE SAYS WHICH SELECTORS COMBINE, beside each one, and each statement is the call's.
func TestTheReportPageSaysWhichSelectorsCombine(t *testing.T) {
	report := cmdAt(NewRootFor("red-lens-evidence"), []string{"show", "report"})
	usage := func(name string) string { return report.Flags().Lookup(name).Usage }
	for name, says := range map[string]string{
		"window":    "ONLY with --anchor, and refused without it",
		"anchor":    "--match, --quote and --assembled are each refused with it",
		"assembled": "--match and --quote select its sections and print them in the --json form, with or without --json; --anchor and --window are refused beside it in either form",
		"match":     "Refused beside --anchor. Beside --assembled it selects SECTIONS, printed in the --json form with or without --json",
		"quote":     "Refused beside --anchor. Beside --assembled it selects SECTIONS, printed in the --json form with or without --json",
	} {
		if !strings.Contains(usage(name), says) {
			t.Errorf("show report --%s does not say %q:\n%s", name, says, usage(name))
		}
	}
	// THE BOUND IS THE REPORT'S: no other view has an anchor read or an assembled form.
	board := cmdAt(NewRootFor("red-lens-evidence"), []string{"show", "board"})
	if u := board.Flags().Lookup("match").Usage; strings.Contains(u, "--assembled") {
		t.Errorf("show board --match carries the report's bound:\n%s", u)
	}

	runDir := actFixture(t)
	call := func(args ...string) (string, error) {
		out, stderr, err := runSeen(t, append([]string{"show", "report", "--run", runDir, "--seat-id", "red-lens-evidence"}, args...)...)
		return out + stderr, err
	}
	for _, refused := range [][]string{
		{"--window", "2"},
		{"--assembled", "--window", "2"},
	} {
		if out, err := call(refused...); err == nil {
			t.Errorf("show report %s is accepted, and the page says it is refused:\n%.200s", strings.Join(refused, " "), out)
		}
	}
	for _, taken := range [][]string{
		{"--assembled", "--json", "--match", "parser"},
		{"--assembled", "--match", "parser"},
		{"--assembled", "--quote", "parser"},
		{"--assembled", "--json"},
		{"--match", "parser"},
		{"--json", "--quote", "parser"},
	} {
		if out, err := call(taken...); err != nil {
			t.Errorf("show report %s is refused, and the page says it is taken: %v\n%.300s", strings.Join(taken, " "), err, out)
		}
	}

	// `heading` IN ITS TWO SPELLINGS, said under the field tree that lists it.
	page := seat.HelpText(report)
	for _, says := range []string{"`heading` IS SPELLED TWO WAYS", "bare name — `Research areas`, no `#` marks", "`## Research areas`"} {
		if !strings.Contains(page, says) {
			t.Errorf("show report's page does not say %q", says)
		}
	}
	assembled, err := call("--assembled", "--json")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(assembled, `"heading": "Research areas"`) || strings.Contains(assembled, `"heading": "## `) {
		t.Errorf("the assembled form does not carry bare section names, which the page says it does:\n%.400s", assembled)
	}
	lines, err := call("--json")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(lines, `"heading": "#`) {
		t.Errorf("the report's lines do not carry their heading as written, which the page says they do:\n%.400s", lines)
	}
}

// THE EVIDENCE PAGE SAYS WHERE A PROOF'S VERDICT LIVES. Two chair sittings rebuilt
// `{reproduced, sound}` from the proof ROW of a nested `verified` object, read null on every proof,
// and recorded verified proofs as unverified while `counts.proofs_unverified` on the same output
// said 0. The verdict is flat on the row now, in its two halves.
func TestTheEvidencePageSaysWhereAProofsVerdictLives(t *testing.T) {
	page := seat.HelpText(cmdAt(NewRootFor("red-chair"), []string{"show", "evidence"}))
	for _, says := range []string{
		"A PROOF'S VERDICT IS FLAT ON ITS ROW, IN TWO HALVES", "`rerun` is the SCRIPT's", "`soundness` is THIS ANCHOR's sentence",
		"`verified` is true when the script reproduced AND this sentence is judged sound",
		"`counts.proofs_unverified` is how many proofs are not verified",
		"`--about-kind contradiction --about <id>`",
	} {
		if !strings.Contains(page, says) {
			t.Errorf("show evidence's page does not say %q", says)
		}
	}
	// AND THE PATHS IT NAMES ARE THE SHAPE'S, read off the field tree the same page generates.
	for _, shape := range []string{"proofs:[{anchor,sha256,basis,cites,drift,seat_id,epoch,rerun,soundness,verified,checks:[{rerun,soundness,standing,",
		"unanswered_contradictions:[{id,claim,reader,anchor,raised_by}]", "counts:{sources,proofs,proofs_unverified,"} {
		if !strings.Contains(page, shape) {
			t.Errorf("the page names a path its own field tree does not hold (%q)", shape)
		}
	}
}

// A PROOF'S SCRIPT IS DRAFTED IN THE RUN'S scratch/ AND KEPT IN ITS PROOF STORE. The page names
// the relative path a seat passes for a draft there, and the stored copy is the proof: the draft can
// go and the proof still reads. An absolute path is still read where it lies.
func TestProveTakesAScriptFromTheRunsScratch(t *testing.T) {
	runDir := actFixture(t)
	body := []byte("echo 0\n")
	if err := os.MkdirAll(filepath.Join(runDir, "scratch"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(runDir, "scratch", "blue-respond-remainder.sh"), body, 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := runAt(t, "prove", "--run", runDir, "--seat-id", "blue-respond",
		"--quote", "the parser accepts an empty body in this line.", "--script", "./scratch/blue-respond-remainder.sh", "--reason", "an empty body parses to nothing")
	if err != nil || !strings.Contains(out, "recorded") {
		t.Fatalf("a script drafted in the run's scratch/ was not proved (%v): %s", err, out)
	}
	if err := os.RemoveAll(filepath.Join(runDir, "scratch")); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(body)
	stored, err := os.ReadFile(filepath.Join(runDir, proof.StoredScript(hex.EncodeToString(sum[:]), "remainder.sh")))
	if err != nil || !bytes.Equal(stored, body) {
		t.Fatalf("with scratch/ gone, the proof store does not hold the script's bytes (%v): %q", err, stored)
	}

	pad := filepath.Join(t.TempDir(), "blue-respond-elsewhere.sh")
	if err := os.WriteFile(pad, []byte("echo 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err = runAt(t, "prove", "--run", runDir, "--seat-id", "blue-respond",
		"--quote", "the parser accepts an empty body in this line.", "--script", pad, "--reason", "an empty body parses to one")
	if err != nil || !strings.Contains(out, "recorded") {
		t.Fatalf("a script at an absolute path outside the run directory was not proved (%v): %s", err, out)
	}

	u := cmdAt(NewRootFor("blue-respond"), []string{"prove"}).Flags().Lookup("script").Usage
	for _, says := range []string{"A relative path is read from the run directory", "drafted in the run's scratch/ is passed as scratch/", "an absolute path is read where it lies", "the proof does not depend on the draft"} {
		if !strings.Contains(u, says) {
			t.Errorf("prove --script does not say %q:\n%s", says, u)
		}
	}
}

// A LENS'S GRADE FLAG PRINTS WHAT THE AXIS MEANS AND STOPS. Why complexity is the axis to contest
// is said to the seat choosing an axis to contest, on the grade motion's menu.
func TestTheContestClauseIsOnTheMotionsMenuAlone(t *testing.T) {
	const contest = "the axis to contest when the fix is worth more than the defect"
	lens := NewRootFor("red-lens-evidence")
	for _, verb := range []string{"mint", "regrade"} {
		u := cmdAt(lens, []string{verb}).Flags().Lookup("complexity").Usage
		if !strings.HasSuffix(u, "what fixing it costs, on the same scale") {
			t.Errorf("%s --complexity does not read \"what fixing it costs, on the same scale\":\n%s", verb, u)
		}
		if page := seat.HelpText(cmdAt(lens, []string{verb})); strings.Contains(page, contest) {
			t.Errorf("the %s page carries the grade motion's contest clause", verb)
		}
	}
	for _, s := range []string{"blue-respond", "red-lens-evidence"} {
		page := seat.HelpText(cmdAt(NewRootFor(s), []string{"motion", "grade", "file"}))
		if !strings.Contains(page, "complexity  what fixing it costs — "+contest+", which is what makes defect_accepted arguable") {
			t.Errorf("%s: the grade motion's menu does not say why complexity is the axis to contest:\n%s", s, page)
		}
	}
}

// THE FINDING PAGE SAYS WHICH AXES A FINDING CARRIES, and the flags are those three.
func TestTheFindingPageSaysWhichAxesItCarries(t *testing.T) {
	finding := cmdAt(NewRootFor("red-lens-evidence"), []string{"finding"})
	if page := seat.HelpText(finding); !strings.Contains(page, "A FINDING IS GRADED ON THREE AXES — severity, likelihood, impact — and not on complexity: that axis is what a FIX costs") {
		t.Errorf("the finding page does not say which axes a finding carries:\n%s", page)
	}
	for name, want := range map[string]bool{"severity": true, "likelihood": true, "impact": true, "complexity": false} {
		if got := finding.Flags().Lookup(name) != nil; got != want {
			t.Errorf("finding takes --%s = %v, and its page says %v", name, got, want)
		}
	}
}
