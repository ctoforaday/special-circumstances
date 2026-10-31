package cli

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// THE SMOKE'S G1, DONE PROPERLY. Red asked blue to test the protocol on a false claim
// ("is 9 prime"). Blue answered in prose asserting the test had happened, and G2 refused
// it for showing no evidence. These drive the verb that makes the assertion unnecessary.

func computeSeat(t *testing.T, runDir, body string) string {
	t.Helper()
	writeReport(t, runDir, body)
	const seat = "blue-respond"
	if _, err := run(t, "register", "--run", runDir, "--seat-id", seat); err != nil {
		t.Fatal(err)
	}
	return seat
}

// script writes a one-file bundle: name, alone in scratch/<name without its extension>/ under the
// run directory, and returns that bundle directory as --dir takes it — relative to the run.
func script(t *testing.T, runDir, name, body string) string {
	t.Helper()
	dir := filepath.Join("scratch", strings.TrimSuffix(name, filepath.Ext(name)))
	if err := os.MkdirAll(filepath.Join(runDir, dir), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(runDir, dir, name), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return filepath.ToSlash(dir)
}

// ahead adds flags to an argument list ahead of its `--`, where compute's command stays last; a
// list with no `--` takes them at its end.
func ahead(args []string, flags ...string) []string {
	out := append([]string{}, args...)
	for i, a := range out {
		if a == "--" {
			return append(append(out[:i:i], flags...), args[i:]...)
		}
	}
	return append(out, flags...)
}

// A computation settles the claim, anchors at the sentence it backs, and leaves an artifact.
func TestComputeAnchorsAndRecordsTheComputation(t *testing.T) {
	runDir := newRun(t)
	seat := computeSeat(t, runDir, "# H\n\nNine is composite, so the protocol rejects a false claim.\n")
	s := script(t, runDir, "nine.js", "let d=[];for(let i=2;i<9;i++)if(9%i===0)d.push(i);console.log('divisors:',d.join(','));")

	out, err := run(t, "compute", "--run", runDir, "--seat-id", seat,
		"--quote", "Nine is composite, so the protocol rejects a false claim.",
		"--dir", s, "--reason", "trial division settles it", "--", "node", "nine.js")
	if err != nil {
		t.Fatalf("compute: %v", err)
	}
	if !strings.Contains(out, "reproducible") {
		t.Errorf("a deterministic computation was not graded reproducible: %q", out)
	}
	// The anchor binds the computation to the sentence, like a citation.
	if !strings.Contains(readReport(t, runDir), "<!--computation:K-") {
		t.Error("no computation anchor was spliced, so nothing connects the script to the claim")
	}
	ev := lastBody(t, runDir, &recordpb.Computation{})
	if ev.GetBasis() != recordpb.Basis_BASIS_REPRODUCIBLE {
		t.Errorf("basis = %q", ev.GetBasis())
	}
	// THE OUTPUT IS IN THE CACHE, NOT ON THE EVENT, and that is a decision rather than a drop —
	// compute.go states it: "content is not a fact about the debate", the output is addressed by
	// `bundle_sha`, and report/computations.go renders script, exit, basis and drift. This asserted
	// `text` carried the answer, which is the field the SEAT's prose lands in.
	//
	// What must hold is the JOIN: the event names a sha, and the sha addresses the output an
	// auditor re-runs against. A sha on the event with nothing behind it is a citation to
	// evidence that does not exist.
	if ev.GetBundleSha() == "" {
		t.Fatal("the computation event names no sha — nothing addresses the output an auditor would re-run against")
	}
	if !strings.Contains(ev.GetText(), "trial division settles it") {
		t.Errorf("the seat's own prose did not reach the event: %q", ev.GetText())
	}
}

// THE COMPUTATION ANCHOR IS IMMORTAL, exactly as a finding marker and a citation anchor are: an
// edit may carry it across but never drop it. Evidence a seat produced by RUNNING something
// must not be removable by rewriting a sentence.
func TestBlueEditCannotDropAComputationAnchor(t *testing.T) {
	runDir := newRun(t)
	// Text FOLLOWS the anchored sentence so the anchor sits MID-span: it splices after the
	// quoted sentence, so a span ending at that sentence never crosses it. The replacement rewrites
	// that sentence, so the tool has nowhere to put the anchor back.
	seat := computeSeat(t, runDir, "# H\n\nNine is composite by trial division. The protocol therefore rejects it.\n")
	s := script(t, runDir, "n.js", "console.log('composite');")
	if _, err := run(t, "compute", "--run", runDir, "--seat-id", seat,
		"--quote", "Nine is composite by trial division.", "--dir", s, "--reason", "r", "--", "node", "n.js"); err != nil {
		t.Fatal(err)
	}
	_, err := run(t, "edit", "--run", runDir, "--seat-id", seat,
		"--key", "E1", "--quote", "by trial division. The protocol therefore rejects it",
		"--new", "by sieving. The protocol rejects it", "--reason", "shorter")
	if err == nil {
		t.Fatal("an edit dropped a computation anchor; the computation backing the claim would vanish silently")
	}
	if !strings.Contains(err.Error(), "computation anchor") {
		t.Errorf("the refusal must name the anchor CLASS so the seat knows what it disturbed: %v", err)
	}
}

// RED RE-RUNS RATHER THAN RE-READS. This is the audit no citation can offer.
func TestRedReproducesAComputation(t *testing.T) {
	runDir := newRun(t)
	seat := computeSeat(t, runDir, "# H\n\nSeven has no divisor between two and six.\n")
	s := script(t, runDir, "seven.js", "console.log('no divisors in 2..6');")
	if _, err := run(t, "compute", "--run", runDir, "--seat-id", seat,
		"--quote", "Seven has no divisor between two and six.", "--dir", s, "--reason", "r", "--", "node", "seven.js"); err != nil {
		t.Fatal(err)
	}
	computationAnchor := lastBody(t, runDir, &recordpb.Computation{}).GetComputationId()

	// --as and --reason are REQUIRED (#343): re-running measures DETERMINISM, and a script that
	// prints "7 is prime" reproduces forever. The soundness verdict is red's, from reading it.
	out, err := run(t, "reproduce", "--run", runDir, "--seat-id", "red-lens-evidence", "--anchor", computationAnchor,
		"--as", "sound", "--reason", "trial division to sqrt(n); it computes primality rather than asserting it")
	if err != nil {
		t.Fatalf("reproduce: %v", err)
	}
	// The verdict is on the RECORD now, not only in red's head.
	if ev := lastBody(t, runDir, &recordpb.Reproduce{}); ev.GetSoundness() != recordpb.Soundness_SOUNDNESS_SOUND {
		t.Errorf("the soundness judgement must be recorded, got %q", ev.GetSoundness())
	}
	if !strings.Contains(out, "REPRODUCES") {
		t.Errorf("red re-ran the computation and did not confirm it: %q", out)
	}
}

// A MOVING RESULT IS RECORDED ONLY WITH WHAT HOLDS ACROSS IT. Output that differs between the two
// runs is refused until the seat makes it deterministic or states the invariant with --variance;
// with one it is recorded `variant`, and --variance over runs that agreed is refused.
func TestAMovingComputationIsRecordedVariantOnlyWithItsInvariant(t *testing.T) {
	runDir := newRun(t)
	seat := computeSeat(t, runDir, "# H\n\nThe sampled value varies between runs.\n")
	s := script(t, runDir, "moves.js", "console.log(Math.random());")
	args := []string{"compute", "--run", runDir, "--seat-id", seat, "--quote", "The sampled value varies between runs.", "--dir", s, "--reason", "sampling a live system"}
	command := []string{"--", "node", "moves.js"}

	before := countType(t, runDir, recordpb.EventType_EVENT_TYPE_COMPUTATION)
	out, err := run(t, append(args, command...)...)
	if err == nil || !strings.Contains(err.Error()+out, "--variance") {
		t.Fatalf("a moving computation with no invariant was recorded, or its refusal does not name --variance: %v\n%s", err, out)
	}
	if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_COMPUTATION); n != before {
		t.Errorf("the refused computation recorded %d event(s)", n-before)
	}
	out, err = run(t, append(append(args, "--variance", "a number in [0, 1)"), command...)...)
	if err != nil || !strings.Contains(out, "variant") {
		t.Fatalf("a moving computation with its invariant stated was not recorded variant: %v\n%s", err, out)
	}
	if c := lastBody(t, runDir, &recordpb.Computation{}); c.GetBasis() != recordpb.Basis_BASIS_VARIANT || c.GetInvariant() != "a number in [0, 1)" || c.GetDrift() == "" {
		t.Errorf("the record holds basis=%s invariant=%q drift=%q", c.GetBasis(), c.GetInvariant(), c.GetDrift())
	}

	steady := script(t, runDir, "steady.js", "console.log(7);")
	out, err = run(t, "compute", "--run", runDir, "--seat-id", seat, "--quote", "The sampled value varies between runs.", "--dir", steady, "--reason", "r", "--variance", "seven", "--", "node", "steady.js")
	if err == nil {
		t.Errorf("--variance over runs that agreed was recorded:\n%s", out)
	}
}

// A computation that cannot run is not evidence, and the failure is a capability signal — the same
// treatment `blue cite` gives an unreachable source. Here the command names a program this
// environment does not have on PATH.
func TestAnUnrunnableComputationIsRefusedAndLogged(t *testing.T) {
	runDir := newRun(t)
	seat := computeSeat(t, runDir, "# H\n\nA sentence to anchor to.\n")
	s := script(t, runDir, "mystery.rb", "puts 1")

	if _, err := run(t, "compute", "--run", runDir, "--seat-id", seat,
		"--quote", "A sentence to anchor to.", "--dir", s, "--reason", "r", "--", "feov-no-such-interpreter", "mystery.rb"); err == nil {
		t.Fatal("a command this environment does not have was accepted as evidence")
	}
	if countType(t, runDir, recordpb.EventType_EVENT_TYPE_LOG) == 0 {
		t.Error("the refusal wrote no log entry, so the missing capability is invisible to the retool loop")
	}
	if countType(t, runDir, recordpb.EventType_EVENT_TYPE_COMPUTATION) != 0 {
		t.Error("a refused computation still landed on the record")
	}
}

// A computation must anchor to text that is actually there — the same rule a citation follows.
func TestComputeRefusesAMisquotedLocation(t *testing.T) {
	runDir := newRun(t)
	seat := computeSeat(t, runDir, "# H\n\nA sentence to anchor to.\n")
	s := script(t, runDir, "ok.js", "console.log('x');")
	if _, err := run(t, "compute", "--run", runDir, "--seat-id", seat,
		"--quote", "a sentence that is not in the report", "--dir", s, "--reason", "r", "--", "node", "ok.js"); err == nil {
		t.Fatal("a computation anchored to text the report does not contain was accepted")
	}
}

// A COMPUTATION'S KEY NAMES ONE COMPUTATION, not one program: a retry under it states the sentence, its note and
// the gap it answers again, and a retry stating any of them otherwise is refused naming the stored
// statement, recording nothing.
func TestAComputationRetryIsHeldToEverythingItStated(t *testing.T) {
	runDir := newRun(t)
	const first, second = "Nine is composite.", "Ten is composite."
	seat := computeSeat(t, runDir, "# H\n\n"+first+"\n\n"+second+"\n")
	s := script(t, runDir, "nine.js", "console.log(9%3)")
	args := []string{"compute", "--run", runDir, "--seat-id", seat, "--key", "P1", "--dir", s, "--quote", first, "--reason", "trial division", "--", "node", "nine.js"}
	if _, err := run(t, args...); err != nil {
		t.Fatalf("compute: %v", err)
	}
	if out, err := run(t, args...); err != nil || !strings.Contains(out, repeatedWords) {
		t.Errorf("an identical retry = %q, %v; want the stored computation, idempotent", out, err)
	}
	for flag, swap := range map[string][2]string{
		"--quote":  {first, second},
		"--reason": {"trial division", "a sieve"},
	} {
		retry := append([]string(nil), args...)
		for i := range retry {
			if retry[i] == swap[0] {
				retry[i] = swap[1]
			}
		}
		before := countType(t, runDir, recordpb.EventType_EVENT_TYPE_COMPUTATION)
		_, err := run(t, retry...)
		if err == nil || !strings.Contains(err.Error(), flag+` "`+swap[0]+`"`) {
			t.Errorf("a retry with another %s = %v; want a conflict naming the stored %s", flag, err, flag)
		}
		if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_COMPUTATION); n != before {
			t.Errorf("the refused retry recorded %d computation(s)", n-before)
		}
	}
}

// --expect-error IS A CLAIM THAT THE RUN FAILS: on a run that failed nothing it is refused rather
// than accepted and used for nothing, and nothing is recorded.
func TestExpectErrorOnACleanRunIsRefused(t *testing.T) {
	runDir := newRun(t)
	seat := computeSeat(t, runDir, "# H\n\nNine is composite.\n")
	s := script(t, runDir, "nine.js", "console.log(9%3)")
	_, err := run(t, "compute", "--run", runDir, "--seat-id", seat, "--quote", "Nine is composite.",
		"--dir", s, "--reason", "trial division", "--expect-error", "--", "node", "nine.js")
	if err == nil || !strings.Contains(err.Error(), "drop --expect-error") {
		t.Errorf("--expect-error on a clean run = %v; want it refused", err)
	}
	if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_COMPUTATION); n != 0 {
		t.Errorf("the refused computation recorded %d computation(s)", n)
	}
}
