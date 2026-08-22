package cli

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// THE SMOKE'S R1-2, DONE PROPERLY. Red asked blue to test the protocol on a false claim
// ("is 9 prime"). Blue answered in prose asserting the test had happened, and R2-2 refused
// it for showing no evidence. These drive the verb that makes the assertion unnecessary.

func proveSeat(t *testing.T, runDir, body string) string {
	t.Helper()
	writeReport(t, runDir, body)
	const seat = "blue-respond-r1"
	if _, err := run(t, "blue", "register", "--run", runDir, "--seat-id", seat); err != nil {
		t.Fatal(err)
	}
	return seat
}

func script(t *testing.T, runDir, name, body string) string {
	t.Helper()
	if err := os.WriteFile(filepath.Join(runDir, name), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return name
}

// A computation settles the claim, anchors at the sentence it backs, and leaves an artifact.
func TestProveAnchorsAndRecordsTheComputation(t *testing.T) {
	runDir := t.TempDir()
	seat := proveSeat(t, runDir, "# H\n\nNine is composite, so the protocol rejects a false claim.\n")
	s := script(t, runDir, "nine.js", "let d=[];for(let i=2;i<9;i++)if(9%i===0)d.push(i);console.log('divisors:',d.join(','));")

	out, err := run(t, "blue", "prove", "--run", runDir, "--seat-id", seat,
		"--quote", "Nine is composite, so the protocol rejects a false claim.",
		"--script", s, "--reason", "trial division settles it")
	if err != nil {
		t.Fatalf("prove: %v", err)
	}
	if !strings.Contains(out, "reproducible") {
		t.Errorf("a deterministic computation was not graded reproducible: %q", out)
	}
	// The anchor binds the computation to the sentence, like a citation.
	if !strings.Contains(readReport(t, runDir), "<!--proof:p-") {
		t.Error("no proof anchor was spliced, so nothing connects the script to the claim")
	}
	ev := lastBody(t, runDir, &recordpb.Proof{})
	if ev.GetProofBasis() != "reproducible" {
		t.Errorf("basis = %q", ev.GetProofBasis())
	}
	// THE OUTPUT IS IN THE CACHE, NOT ON THE EVENT, and that is a decision rather than a drop —
	// prove.go states it: "content is not a fact about the debate", the output is addressed by
	// `proof_sha`, and report/proofs.go renders script, exit, basis and drift. This asserted
	// `text` carried the answer, which is the field the SEAT's prose lands in.
	//
	// What must hold is the JOIN: the event names a sha, and the sha addresses the output an
	// auditor re-runs against. A sha on the event with nothing behind it is a citation to
	// evidence that does not exist.
	if ev.GetProofSha() == "" {
		t.Fatal("the proof event names no sha — nothing addresses the output an auditor would re-run against")
	}
	if !strings.Contains(ev.GetText(), "trial division settles it") {
		t.Errorf("the seat's own prose did not reach the event: %q", ev.GetText())
	}
}

// THE PROOF ANCHOR IS IMMORTAL, exactly as a finding marker and a citation anchor are: an
// edit may carry it across but never drop it. Evidence a seat produced by RUNNING something
// must not be removable by rewriting a sentence.
func TestBlueEditCannotDropAProofAnchor(t *testing.T) {
	runDir := t.TempDir()
	// Text FOLLOWS the anchored sentence so the anchor sits MID-span: it splices after the
	// quoted sentence, so a span ending at that sentence never crosses it.
	seat := proveSeat(t, runDir, "# H\n\nNine is composite by trial division. The protocol therefore rejects it.\n")
	s := script(t, runDir, "n.js", "console.log('composite');")
	if _, err := run(t, "blue", "prove", "--run", runDir, "--seat-id", seat,
		"--quote", "Nine is composite by trial division.", "--script", s, "--reason", "r"); err != nil {
		t.Fatal(err)
	}
	_, err := run(t, "blue", "edit", "--run", runDir, "--seat-id", seat,
		"--key", "E1", "--quote", "by trial division. The protocol therefore rejects it",
		"--new", "by trial division. The protocol rejects it", "--reason", "shorter")
	if err == nil {
		t.Fatal("an edit dropped a proof anchor; the computation backing the claim would vanish silently")
	}
	if !strings.Contains(err.Error(), "proof anchor") {
		t.Errorf("the refusal must name the anchor CLASS so the seat knows what it disturbed: %v", err)
	}
}

// RED RE-RUNS RATHER THAN RE-READS. This is the audit no citation can offer.
func TestRedReproducesAProof(t *testing.T) {
	runDir := t.TempDir()
	seat := proveSeat(t, runDir, "# H\n\nSeven has no divisor between two and six.\n")
	s := script(t, runDir, "seven.js", "console.log('no divisors in 2..6');")
	if _, err := run(t, "blue", "prove", "--run", runDir, "--seat-id", seat,
		"--quote", "Seven has no divisor between two and six.", "--script", s, "--reason", "r"); err != nil {
		t.Fatal(err)
	}
	sha := lastBody(t, runDir, &recordpb.Proof{}).GetProofSha()

	// --as and --reason are REQUIRED (#343): re-running measures DETERMINISM, and a script that
	// prints "7 is prime" reproduces forever. The soundness verdict is red's, from reading it.
	out, err := run(t, "lens", "reproduce", "--run", runDir, "--seat-id", "red-lens-r1-L1", "--id", sha,
		"--as", "sound", "--reason", "trial division to sqrt(n); it computes primality rather than asserting it")
	if err != nil {
		t.Fatalf("reproduce: %v", err)
	}
	// The verdict is on the RECORD now, not only in red's head.
	if ev := lastBody(t, runDir, &recordpb.Reproduce{}); ev.GetSoundness() != recordpb.Soundness_SOUNDNESS_SOUND {
		t.Errorf("the soundness judgement must be recorded, got %q", ev.GetSoundness())
	}
	if !strings.Contains(out, "REPRODUCES") {
		t.Errorf("red re-ran the proof and did not confirm it: %q", out)
	}
}

// A MOVING RESULT IS RECORDED AS A MEASUREMENT, not refused and not laundered as a proof.
// The network choice makes this the routine case.
func TestAMovingProofIsGradedObserved(t *testing.T) {
	runDir := t.TempDir()
	seat := proveSeat(t, runDir, "# H\n\nThe sampled value varies between runs.\n")
	s := script(t, runDir, "moves.js", "console.log(Math.random());")

	out, err := run(t, "blue", "prove", "--run", runDir, "--seat-id", seat,
		"--quote", "The sampled value varies between runs.", "--script", s, "--reason", "sampling a live system")
	if err != nil {
		t.Fatalf("a non-deterministic computation was refused; it is a measurement, not an error: %v", err)
	}
	if !strings.Contains(out, "observed") || !strings.Contains(out, "NOT reproducible") {
		t.Errorf("the seat was not told its evidence is a measurement rather than a proof: %q", out)
	}
}

// A proof that cannot run is not evidence, and the failure is a capability signal — the same
// treatment `blue cite` gives an unreachable source.
func TestAnUnrunnableProofIsRefusedAndLogsFriction(t *testing.T) {
	runDir := t.TempDir()
	seat := proveSeat(t, runDir, "# H\n\nA sentence to anchor to.\n")
	s := script(t, runDir, "mystery.rb", "puts 1")

	if _, err := run(t, "blue", "prove", "--run", runDir, "--seat-id", seat,
		"--quote", "A sentence to anchor to.", "--script", s, "--reason", "r"); err == nil {
		t.Fatal("a script with no known interpreter was accepted as evidence")
	}
	if countType(t, runDir, recordpb.EventType_EVENT_TYPE_FRICTION) == 0 {
		t.Error("the refusal logged no friction, so the capability gap is invisible to the retool loop")
	}
	if countType(t, runDir, recordpb.EventType_EVENT_TYPE_PROOF) != 0 {
		t.Error("a refused proof still landed on the record")
	}
}

// A proof must anchor to text that is actually there — the same rule a citation follows.
func TestProveRefusesAMisquotedLocation(t *testing.T) {
	runDir := t.TempDir()
	seat := proveSeat(t, runDir, "# H\n\nA sentence to anchor to.\n")
	s := script(t, runDir, "ok.js", "console.log('x');")
	if _, err := run(t, "blue", "prove", "--run", runDir, "--seat-id", seat,
		"--quote", "a sentence that is not in the report", "--script", s, "--reason", "r"); err == nil {
		t.Fatal("a proof anchored to text the report does not contain was accepted")
	}
}
