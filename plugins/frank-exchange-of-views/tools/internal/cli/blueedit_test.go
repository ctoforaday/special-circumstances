package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// blue edit is the ONLY write path to blue/report.md for a response seat. These drive the
// real command tree so the lockdown's write path is pinned where it is enforced: an exact
// span replace that PRESERVES red's invisible finding-markers, rejects a marker-spanning
// edit, records an append-only diff-stack op, and reconciles idempotently after a crash.

func writeReport(t *testing.T, runDir, body string) {
	t.Helper()
	dir := filepath.Join(runDir, "blue")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "report.md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readReport(t *testing.T, runDir string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(runDir, "blue", "report.md"))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func countType(t *testing.T, runDir, typ string) int {
	t.Helper()
	m, err := record.MergedEvents(runDir)
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	for _, e := range m.Events {
		if e.Type == typ {
			n++
		}
	}
	return n
}

const blueSeat = "blue-respond-r2"

func registerBlue(t *testing.T, runDir string) {
	t.Helper()
	if _, err := run(t, "blue", "register", "--run", runDir, "--seat-id", blueSeat); err != nil {
		t.Fatal(err)
	}
}

func TestBlueEditReplacesSpanPreservingMarker(t *testing.T) {
	runDir := t.TempDir()
	// A trailing finding-marker after "time"; a footnote after "grows".
	writeReport(t, runDir, "# Findings\n\nThe cost is high and rising over time<!--fx:f-abc123-->. Volume grows[^v] steadily.\n")
	registerBlue(t, runDir)

	out, err := run(t, "blue", "edit", "--run", runDir, "--seat-id", blueSeat,
		"--key", "E1", "--old", "rising over time", "--new", "climbing sharply", "--reason", "sharper phrasing")
	if err != nil {
		t.Fatalf("blue edit: %v (out %q)", err, out)
	}
	rep := readReport(t, runDir)
	if !strings.Contains(rep, "climbing sharply") || strings.Contains(rep, "rising over time") {
		t.Errorf("span not replaced: %q", rep)
	}
	if !strings.Contains(rep, "<!--fx:f-abc123-->") {
		t.Errorf("finding-marker was dropped: %q", rep)
	}
	ev := lastOfType(t, runDir, "blue_edit")
	if ev.Payload.Str("old") != "rising over time" || ev.Payload.Str("new") != "climbing sharply" {
		t.Errorf("blue_edit op payload wrong: old=%q new=%q", ev.Payload.Str("old"), ev.Payload.Str("new"))
	}
	if ev.Payload.Str("reason") != "sharper phrasing" {
		t.Errorf("reason not recorded: %q", ev.Payload.Str("reason"))
	}
}

func TestBlueEditRejectsMarkerSpanningEdit(t *testing.T) {
	runDir := t.TempDir()
	writeReport(t, runDir, "# H\n\nThe value is context<!--fx:f-1-->: important here.\n")
	registerBlue(t, runDir)

	out, err := run(t, "blue", "edit", "--run", runDir, "--seat-id", blueSeat,
		"--key", "E1", "--old", "context: important", "--new", "context: vital", "--reason", "x")
	if err == nil {
		t.Fatalf("expected a reject; out %q", out)
	}
	// An edit whose span holds an anchor is refused only when --new DROPS it. The message must
	// name the marker and tell blue how to succeed — carry the token across — rather than the old
	// "edit around it", which deadlocked against the uniqueness guard whenever the disambiguating
	// context was the anchored text itself.
	if !strings.Contains(err.Error(), "f-1") || !strings.Contains(err.Error(), "<!--fx:f-1-->") {
		t.Errorf("reject message should name the marker and quote the token to reproduce: %v", err)
	}
	if strings.Contains(readReport(t, runDir), "vital") {
		t.Error("report was mutated on a rejected marker-spanning edit")
	}
	if n := countType(t, runDir, "blue_edit"); n != 0 {
		t.Errorf("a rejected edit recorded %d stack ops, want 0", n)
	}
}

func TestBlueEditRejectsAbsentOld(t *testing.T) {
	runDir := t.TempDir()
	writeReport(t, runDir, "# H\n\nThe scheduler is preemptive.\n")
	registerBlue(t, runDir)

	_, err := run(t, "blue", "edit", "--run", runDir, "--seat-id", blueSeat,
		"--key", "E1", "--old", "the scheduler is cooperative", "--new", "x", "--reason", "y")
	if err == nil {
		t.Fatal("a mis-quote must be rejected")
	}
	if n := countType(t, runDir, "blue_edit"); n != 0 {
		t.Errorf("a mis-quote recorded %d stack ops, want 0 (no phantom event)", n)
	}
}

func TestBlueEditIdempotentRetry(t *testing.T) {
	runDir := t.TempDir()
	writeReport(t, runDir, "# H\n\nThe cost is rising fast.\n")
	registerBlue(t, runDir)
	args := []string{"blue", "edit", "--run", runDir, "--seat-id", blueSeat,
		"--key", "E1", "--old", "rising fast.", "--new", "climbing fast.", "--reason", "r"}
	if _, err := run(t, args...); err != nil {
		t.Fatal(err)
	}
	// Same --key again: old is now gone, new present. Reconcile → no-op, no second op.
	out, err := run(t, args...)
	if err != nil {
		t.Fatalf("idempotent retry errored: %v (%q)", err, out)
	}
	if !strings.Contains(out, "idempotent") {
		t.Errorf("retry should report idempotent: %q", out)
	}
	if n := countType(t, runDir, "blue_edit"); n != 1 {
		t.Errorf("retry recorded %d ops, want exactly 1", n)
	}
	if strings.Count(readReport(t, runDir), "climbing fast") != 1 {
		t.Error("retry double-applied the edit")
	}
}

// Crash after the event append, before the write lands (event-first ordering): the op is
// on the stack but the report still holds `old`. A retry under the same key reconciles
// FORWARD — applies the write — and appends no second op. No wedge.
func TestBlueEditReconcilesEventWithoutWrite(t *testing.T) {
	runDir := t.TempDir()
	writeReport(t, runDir, "# H\n\nThe cost is rising fast.\n")
	registerBlue(t, runDir)
	// Simulate the crash window: the intent event exists, the write never happened.
	p := record.NewPayload()
	p.Set("edit_key", "E1")
	p.Set("old", "rising fast")
	p.Set("new", "climbing fast")
	p.Set("reason", "r")
	if _, err := record.Append(record.Identity{RunDir: runDir, SeatID: blueSeat, Round: record.RoundOf(blueSeat)}, "blue_edit", p); err != nil {
		t.Fatal(err)
	}
	// Retry with the same key → reconcile forward.
	if _, err := run(t, "blue", "edit", "--run", runDir, "--seat-id", blueSeat,
		"--key", "E1", "--old", "rising fast.", "--new", "climbing fast.", "--reason", "r"); err != nil {
		t.Fatalf("reconcile retry errored: %v", err)
	}
	if !strings.Contains(readReport(t, runDir), "climbing fast") {
		t.Error("reconcile did not apply the pending write")
	}
	if n := countType(t, runDir, "blue_edit"); n != 1 {
		t.Errorf("reconcile appended a second op (%d), want 1", n)
	}
}

// ---- --answers: provenance as a join key (#267) ----

// The happy path: --answers lands on the event as the join key every #267 measurement reads.
func TestBlueEditRecordsTheGapItAnswers(t *testing.T) {
	runDir := t.TempDir()
	writeReport(t, runDir, "# H\n\nFive independent verification approaches agree.\n")
	gap := mintGap(t, runDir, "G1", "overclaimed-independence")
	registerBlue(t, runDir)

	if _, err := run(t, "blue", "edit", "--run", runDir, "--seat-id", blueSeat,
		"--key", "E1", "--old", "Five independent verification", "--new", "Five verification",
		"--answers", gap, "--reason", "drop the independence claim"); err != nil {
		t.Fatalf("blue edit --answers: %v", err)
	}
	if got := lastOfType(t, runDir, "blue_edit").Payload.Str("answers"); got != gap {
		t.Errorf("answers = %q, want %q — without it required_fix and the actual change cannot be joined", got, gap)
	}
}

// A gap no mint created is refused HERE, while the seat is still there to fix it. An
// unchecked reference is accepted at write time and DROPPED at replay (refs.go).
func TestBlueEditRefusesAnUnknownGap(t *testing.T) {
	runDir := t.TempDir()
	writeReport(t, runDir, "# H\n\nSome text to change here.\n")
	mintGap(t, runDir, "G1", "overclaimed-independence")
	registerBlue(t, runDir)

	_, err := run(t, "blue", "edit", "--run", runDir, "--seat-id", blueSeat,
		"--key", "E1", "--old", "text to change", "--new", "prose to revise",
		"--answers", "R9-99", "--reason", "why")
	if err == nil {
		t.Fatal("an edit answering a gap no mint created was accepted")
	}
	if !strings.Contains(err.Error(), "R9-99") {
		t.Errorf("the refusal must name the dangling id: %v", err)
	}
	if n := countType(t, runDir, "blue_edit"); n != 0 {
		t.Errorf("a refused edit still recorded %d op(s) — validation must precede the append", n)
	}
	if strings.Contains(readReport(t, runDir), "prose to revise") {
		t.Error("a refused edit still wrote to report.md")
	}
}

// THE CONVENTION IS REFUSED, NOT DEPRECATED. 19 of 26 smoke edits opened --reason with the
// gap id and 7 did not; a 73% link looks like a key and is not one.
func TestBlueEditRefusesAGapIDHidingInTheReason(t *testing.T) {
	runDir := t.TempDir()
	writeReport(t, runDir, "# H\n\nSome text to change here.\n")
	gap := mintGap(t, runDir, "G1", "overclaimed-independence")
	registerBlue(t, runDir)

	_, err := run(t, "blue", "edit", "--run", runDir, "--seat-id", blueSeat,
		"--key", "E1", "--old", "text to change", "--new", "prose to revise",
		"--reason", gap+": acknowledge the shared definition")
	if err == nil {
		t.Fatal("the free-text convention was accepted while --answers was empty")
	}
	if !strings.Contains(err.Error(), "--answers") || !strings.Contains(err.Error(), gap) {
		t.Errorf("the refusal must name both the gap and the flag that fixes it: %v", err)
	}
	if n := countType(t, runDir, "blue_edit"); n != 0 {
		t.Errorf("a refused edit recorded %d op(s)", n)
	}
}

// The guard matches the BOARD, not a pattern: prose that merely looks id-shaped, or names
// a gap this run never minted, is blue's business and passes.
func TestBlueEditAllowsProseThatNamesNoRealGap(t *testing.T) {
	runDir := t.TempDir()
	writeReport(t, runDir, "# H\n\nSome text to change here.\n")
	mintGap(t, runDir, "G1", "overclaimed-independence")
	registerBlue(t, runDir)

	if _, err := run(t, "blue", "edit", "--run", runDir, "--seat-id", blueSeat,
		"--key", "E1", "--old", "text to change", "--new", "prose to revise",
		"--reason", "tightened per R9-99 and section 3-2 of the style note"); err != nil {
		t.Fatalf("prose naming no real gap was refused: %v", err)
	}
	if countType(t, runDir, "blue_edit") != 1 {
		t.Error("the edit did not land")
	}
}

// An edit that answers no gap — blue's own authorial work, or the punctuation repair that
// was 6 of the smoke's 26 edits — stays legal with --answers absent.
func TestBlueEditWithoutAnswersIsStillLegal(t *testing.T) {
	runDir := t.TempDir()
	writeReport(t, runDir, "# H\n\nSome text to change here.\n")
	mintGap(t, runDir, "G1", "overclaimed-independence")
	registerBlue(t, runDir)

	if _, err := run(t, "blue", "edit", "--run", runDir, "--seat-id", blueSeat,
		"--key", "E1", "--old", "text to change", "--new", "prose to revise",
		"--reason", "clearer phrasing, self-directed"); err != nil {
		t.Fatalf("a self-directed edit was refused: %v", err)
	}
	if got := lastOfType(t, runDir, "blue_edit").Payload.Str("answers"); got != "" {
		t.Errorf("answers = %q, want empty", got)
	}
}

// The help IS the seat's contract (contract_test.go's premise). A flag whose description
// drifts out of the help teaches nothing, and the seat's only teacher is what it can read.
func TestBlueEditHelpTeachesTheProvenanceFlag(t *testing.T) {
	h := help(t, "blue", "edit", "--help")
	if !strings.Contains(h, "--answers ") {
		t.Fatalf("blue edit --help never names --answers, so the join key is undiscoverable:\n%s", h)
	}
	for _, want := range []string{"gap id", "provenance"} {
		if !strings.Contains(h, want) {
			t.Errorf("blue edit --help does not say %q — a seat cannot tell what --answers takes:\n%s", want, h)
		}
	}
}

// ---- concrete proposals: fix_basis is DERIVED, never claimed (#267 stage 3) ----

// Prose alone stays legal and is honestly labelled `proposed`. This is the substantive
// channel — "research X", "enumerate the residual risks" — and it must not be second-class.
func TestMintWithoutAConcreteProposalIsBasisProposed(t *testing.T) {
	runDir := t.TempDir()
	writeReport(t, runDir, "# H\n\nFive independent verification approaches agree.\n")
	mintGap(t, runDir, "G1", "overclaim")

	ev := lastOfType(t, runDir, "mint")
	if got := ev.Payload.Str("fix_basis"); got != "proposed" {
		t.Errorf("fix_basis = %q, want %q — a prose fix red did not check against the document is proposed", got, "proposed")
	}
}

// A validated pair EARNS `verified`. The basis is computed from the check passing, so red
// cannot assert it: there is no flag to type.
func TestConcreteProposalEarnsBasisVerified(t *testing.T) {
	runDir := t.TempDir()
	writeReport(t, runDir, "# H\n\nFive independent verification approaches agree.\n")
	if _, err := run(t, "merge", "mint", "--run", runDir, "--seat-id", "red-merge-r1",
		"--key", "G1", "--class", "overclaim", "--class-new", "--definition", "d", "--neighbor", "n",
		"--distinguisher", "x", "--location", "Five independent verification approaches agree.", "--problem", "the defect",
		"--fix", "drop the independence claim", "--check-kind", "document", "--check", "the section no longer claims independence",
		"--severity", "medium", "--likelihood", "medium", "--impact", "medium", "--cx", "low",
		"--fix-old", "Five independent verification", "--fix-new", "Five verification"); err != nil {
		t.Fatalf("a legal concrete proposal was refused: %v", err)
	}
	ev := lastOfType(t, runDir, "mint")
	if got := ev.Payload.Str("fix_basis"); got != "verified" {
		t.Errorf("fix_basis = %q, want %q — a pair validated against the live report is verified", got, "verified")
	}
	if ev.Payload.Str("fix_old") == "" || ev.Payload.Str("fix_new") == "" {
		t.Error("the proposal itself was not recorded, so blue cannot apply what red proposed")
	}
}

// There is NO --fix-basis flag. A seat asked to self-report verified|proposed reports the
// flattering one; the whole axis depends on the basis being unclaimable.
func TestThereIsNoWayToClaimAVerifiedBasis(t *testing.T) {
	runDir := t.TempDir()
	writeReport(t, runDir, "# H\n\nSome text.\n")
	_, err := run(t, "merge", "mint", "--run", runDir, "--seat-id", "red-merge-r1",
		"--key", "G1", "--class", "x", "--check-kind", "document", "--check", "c", "--problem", "p",
		"--likelihood", "medium", "--impact", "medium", "--fix-basis", "verified")
	if err == nil {
		t.Fatal("--fix-basis was accepted — the basis must be derived, never asserted")
	}
	if !strings.Contains(err.Error(), "fix-basis") {
		t.Errorf("the refusal should name the unknown flag: %v", err)
	}
}

// Half a proposal is not a proposal: blue could not apply it, so it is refused at the source
// rather than recorded as something blue must interpret.
func TestHalfAProposalIsRefused(t *testing.T) {
	runDir := t.TempDir()
	writeReport(t, runDir, "# H\n\nFive independent verification approaches agree.\n")
	_, err := run(t, "merge", "mint", "--run", runDir, "--seat-id", "red-merge-r1",
		"--key", "G1", "--class", "x", "--check-kind", "document", "--check", "c", "--problem", "p",
		"--likelihood", "medium", "--impact", "medium",
		"--fix-old", "Five independent verification")
	if err == nil {
		t.Fatal("--fix-old without --fix-new was accepted")
	}
	if !strings.Contains(err.Error(), "--fix-new") {
		t.Errorf("the refusal must name the missing half: %v", err)
	}
	if n := countType(t, runDir, "mint"); n != 0 {
		t.Errorf("a refused mint still recorded %d event(s)", n)
	}
}

// A proposal red could not have written without reading the document is the point of the
// axis; one quoting text that is not there proves the opposite, and is refused.
func TestAProposalAgainstTextThatIsNotThereIsRefused(t *testing.T) {
	runDir := t.TempDir()
	writeReport(t, runDir, "# H\n\nFive independent verification approaches agree.\n")
	_, err := run(t, "merge", "mint", "--run", runDir, "--seat-id", "red-merge-r1",
		"--key", "G1", "--class", "x", "--check-kind", "document", "--check", "c", "--problem", "p",
		"--likelihood", "medium", "--impact", "medium",
		"--fix-old", "a sentence the report never contained", "--fix-new", "anything")
	if err == nil {
		t.Fatal("a proposal against absent text was accepted, so nothing forced red to read the report")
	}
	if n := countType(t, runDir, "mint"); n != 0 {
		t.Errorf("a refused mint still recorded %d event(s)", n)
	}
}

// ---- estoppel: red is bound by the fix it prescribed (#267 stage 4) ----

const prescribedText = "Five verification approaches agree, all sharing one definition of primality."

// mintWithProposal mints a gap carrying a concrete proposal and returns its id.
func mintWithProposal(t *testing.T, runDir, key, fixOld, fixNew string) string {
	t.Helper()
	out, err := run(t, "merge", "mint", "--run", runDir, "--seat-id", "red-merge-r1",
		"--key", key, "--class", "overclaim", "--location", "Five independent verification approaches agree.", "--problem", "the defect",
		"--fix", "drop the independence claim", "--check-kind", "document", "--check", "the section no longer claims it",
		"--likelihood", "medium", "--impact", "medium",
		"--fix-old", fixOld, "--fix-new", fixNew)
	if err != nil {
		t.Fatalf("mint with proposal: %v", err)
	}
	return gapID(out)
}

// seedProposalApplied mints a proposal and has blue apply it VERBATIM, which is the only
// state that estops red.
func seedProposalApplied(t *testing.T, runDir string) string {
	t.Helper()
	// TWO sentences: the second is UNRELATED text the estoppel case points at.
	writeReport(t, runDir, "# H\n\nFive independent verification approaches agree.\n\nSieve costs grow with the bound.\n")
	// A first gap so the class exists, then the one that carries the proposal.
	mintGap(t, runDir, "G0", "overclaim")
	gap := mintWithProposal(t, runDir, "G1", "Five independent verification approaches agree", prescribedText)
	registerBlue(t, runDir)
	if _, err := run(t, "blue", "edit", "--run", runDir, "--seat-id", blueSeat,
		"--key", "E1", "--answers", gap,
		"--old", "Five independent verification approaches agree", "--new", prescribedText,
		"--reason", "applying red's proposed text verbatim"); err != nil {
		t.Fatalf("blue applying red's proposal: %v", err)
	}
	return gap
}

// Applying red's exact text is recorded by the TOOL comparing bytes — never claimed.
func TestApplyingRedsProposalVerbatimIsRecorded(t *testing.T) {
	runDir := t.TempDir()
	seedProposalApplied(t, runDir)
	if !lastOfType(t, runDir, "blue_edit").Payload.Bool("applied_verbatim") {
		t.Error("an edit identical to red's proposal was not recorded as verbatim, so nothing estops red")
	}
}

// A counter-edit does NOT set the flag: blue's right to disagree stays real, and the text it
// authored stays auditable.
func TestACounterEditIsNotRecordedAsVerbatim(t *testing.T) {
	runDir := t.TempDir()
	writeReport(t, runDir, "# H\n\nFive independent verification approaches agree.\n")
	mintGap(t, runDir, "G0", "overclaim")
	gap := mintWithProposal(t, runDir, "G1", "Five independent verification approaches agree", prescribedText)
	registerBlue(t, runDir)
	if _, err := run(t, "blue", "edit", "--run", runDir, "--seat-id", blueSeat,
		"--key", "E1", "--answers", gap,
		"--old", "Five independent verification approaches agree", "--new", "Five approaches agree, on one shared definition",
		"--reason", "red's wording overstates it; mine is tighter"); err != nil {
		t.Fatalf("counter-edit: %v", err)
	}
	if lastOfType(t, runDir, "blue_edit").Payload.Bool("applied_verbatim") {
		t.Error("a counter-edit was recorded as verbatim application — blue's own text would then estop red")
	}
}

// THE GUARD. A fresh gap located in text red prescribed and blue applied verbatim is
// refused, and the refusal is LOGGED AS FRICTION so the pathology is countable rather than
// silently prevented.
func TestEstoppelRefusesAFreshGapAgainstRedsOwnPrescription(t *testing.T) {
	runDir := t.TempDir()
	prior := seedProposalApplied(t, runDir)

	before := countType(t, runDir, "mint")
	_, err := run(t, "merge", "mint", "--run", runDir, "--seat-id", "red-merge-r2",
		"--key", "G2", "--class", "overclaim", "--location", prescribedText,
		"--problem", "this sentence overclaims", "--check-kind", "document", "--check", "c",
		"--likelihood", "medium", "--impact", "medium")
	if err == nil {
		t.Fatal("red opened a fresh gap against text it prescribed itself")
	}
	for _, want := range []string{"estoppel", prior, "--supersedes"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal must name %q so red knows where to argue it: %v", want, err)
		}
	}
	if n := countType(t, runDir, "mint"); n != before {
		t.Errorf("a refused mint still landed on the board (%d -> %d)", before, n)
	}
	// A guard that silently blocks is invisible.
	if countType(t, runDir, "friction") == 0 {
		t.Error("the estoppel rejection logged no friction — the block is unmeasurable, which is how a dead guard survives")
	}
}

// RED IS NOT SILENCED. Declaring the lineage IS the amendment the guard asks for, so a mint
// naming the estopping gap in --supersedes goes through.
func TestEstoppelLetsAnAmendmentThroughWhenLineageIsDeclared(t *testing.T) {
	runDir := t.TempDir()
	prior := seedProposalApplied(t, runDir)

	if _, err := run(t, "merge", "mint", "--run", runDir, "--seat-id", "red-merge-r2",
		"--key", "G2", "--class", "overclaim", "--location", prescribedText,
		"--problem", "my own fix turned out to contradict §3", "--check-kind", "document", "--check", "c",
		"--likelihood", "medium", "--impact", "medium",
		"--supersedes", prior); err != nil {
		t.Fatalf("red was blocked from AMENDING its own prescription, which the guard must allow: %v", err)
	}
}

// Text blue authored is auditable normally — the guard is not a shield over the report.
func TestEstoppelDoesNotBlockAGapAgainstUnrelatedText(t *testing.T) {
	runDir := t.TempDir()
	seedProposalApplied(t, runDir)

	if _, err := run(t, "merge", "mint", "--run", runDir, "--seat-id", "red-merge-r2",
		// UNRELATED, but PRESENT. Since 0.63.0 a mint's --location is matched against the report,
		// so "text the guard should not cover" can no longer mean "text that does not exist".
		"--key", "G2", "--class", "overclaim", "--location", "Sieve costs grow with the bound.",
		"--problem", "unsupported", "--check-kind", "document", "--check", "c",
		"--likelihood", "medium", "--impact", "medium"); err != nil {
		t.Fatalf("an unrelated finding was estopped — the guard is over-broad: %v", err)
	}
}
