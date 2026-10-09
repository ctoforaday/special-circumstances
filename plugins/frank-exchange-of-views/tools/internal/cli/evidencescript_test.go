package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// A PROOF'S SCRIPT IS READ THROUGH THE RECORD'S PROJECTION.
//
// The lens owes a soundness verdict that rests on READING the script, and `reproduce` records that
// verdict in the act that re-runs — so the read has to come first and has to be somewhere. Wave A
// (2026-10-09) had no such place: six log entries from five lenses say the evidence view gave only
// the sha256, and each found the file by listing a `proofs/` directory nothing had told it about,
// against a prompt that says every read is a projection. One filed three placeholder verdicts
// instead.

// provenRun records one proof and returns its anchor and sha256 as the unscoped table lists them.
func provenRun(t *testing.T, body string) (runDir, anchor, sha string) {
	t.Helper()
	runDir = newRun(t)
	seat := proveSeat(t, runDir, "# H\n\nSeven has no divisor between two and six.\n")
	s := script(t, runDir, "seven.js", body)
	if _, err := run(t, "prove", "--run", runDir, "--seat-id", seat,
		"--quote", "Seven has no divisor between two and six.", "--script", s, "--reason", "r"); err != nil {
		t.Fatal(err)
	}
	p := lastBody(t, runDir, &recordpb.Proof{})
	if _, err := run(t, "register", "--run", runDir, "--seat-id", "red-lens-evidence"); err != nil {
		t.Fatal(err)
	}
	return runDir, p.GetProofId(), p.GetProofSha()
}

func TestAProofReadAtItsAnchorCarriesItsScriptAndRecordedOutput(t *testing.T) {
	const body = "let d=[];for(let i=2;i<7;i++)if(7%i===0)d.push(i);\nconsole.log('divisors of 7 in 2..6:', d.length);\n"
	runDir, anchor, sha := provenRun(t, body)

	// THE UNSCOPED VIEW STAYS A TABLE. Every seat pulls it, and a run's scripts in every seat's
	// context is a cost nobody asked for: 15 of wave A's 39 seats were over half their window.
	table, err := run(t, "show", "--run", runDir, "--seat-id", "red-lens-evidence", "evidence")
	if err != nil {
		t.Fatalf("show evidence: %v", err)
	}
	if strings.Contains(table, "divisors of 7") || strings.Contains(table, `"script"`) {
		t.Errorf("the unscoped evidence table carries a script; it is a lookup table, and the script is read at one anchor:\n%s", table)
	}

	// THE WORK ITEM THAT SENDS A LENS TO THE PROOF SAYS WHERE ITS SCRIPT IS READ, before it hands
	// over the id the re-run takes.
	work, err := run(t, "show", "--run", runDir, "--seat-id", "red-lens-evidence", "work")
	if err != nil {
		t.Fatalf("show work: %v", err)
	}
	read, rerun := strings.Index(work, "show evidence --anchor "+anchor), strings.Index(work, "--id "+sha)
	if read < 0 || rerun < 0 || read > rerun {
		t.Errorf("the lens's work item for an un-re-run proof must name the read at its anchor BEFORE the re-run's id (read at %d, re-run at %d):\n%s", read, rerun, work)
	}

	// BOTH PARTIES READ IT: the lens that judges the script, and blue, who wrote it and answers for it.
	for _, seat := range []string{"red-lens-evidence", "blue-respond"} {
		out, err := run(t, "show", "--run", runDir, "--seat-id", seat, "evidence", "--anchor", anchor)
		if err != nil {
			t.Fatalf("%s: show evidence --anchor %s: %v", seat, anchor, err)
		}
		var got struct {
			Anchor string `json:"anchor"`
			Proof  *struct {
				Anchor         string  `json:"anchor"`
				Sha256         string  `json:"sha256"`
				ScriptFile     string  `json:"script_file"`
				Script         *string `json:"script"`
				RecordedOutput *string `json:"recorded_output"`
			} `json:"proof"`
			Source json.RawMessage `json:"source"`
		}
		if err := json.Unmarshal([]byte(out), &got); err != nil {
			t.Fatalf("%s: the scoped read is not JSON: %v\n%s", seat, err, out)
		}
		if got.Proof == nil {
			t.Fatalf("%s: a proof's anchor resolved to no proof:\n%s", seat, out)
		}
		if got.Anchor != anchor || got.Proof.Anchor != anchor || got.Proof.Sha256 != sha {
			t.Errorf("%s: asked for %s (%s), got anchor %q / proof %q (%s)", seat, anchor, sha, got.Anchor, got.Proof.Anchor, got.Proof.Sha256)
		}
		// THE BYTES `reproduce` EXECUTES, not a summary of them.
		if got.Proof.Script == nil || *got.Proof.Script != body {
			t.Errorf("%s: the script is not the stored bytes:\n got %v\nwant %q", seat, got.Proof.Script, body)
		}
		if got.Proof.RecordedOutput == nil || !strings.Contains(*got.Proof.RecordedOutput, "divisors of 7 in 2..6: 0") {
			t.Errorf("%s: the recorded output is not the stored bytes: %v", seat, got.Proof.RecordedOutput)
		}
		if got.Proof.ScriptFile != "script.js" {
			t.Errorf("%s: script_file = %q; the extension is what says which interpreter runs it", seat, got.Proof.ScriptFile)
		}
		if len(got.Source) != 0 {
			t.Errorf("%s: a proof's anchor also returned a source: %s", seat, got.Source)
		}
	}
}

// A MISSING ARTIFACT IS A REFUSAL, NEVER AN EMPTY SCRIPT. `"script": ""` reads exactly like a
// script that does nothing, and a lens would judge it unsound — a verdict about the store,
// recorded as a verdict about blue's computation.
func TestAProofWhoseArtifactIsMissingIsRefusedNotShownEmpty(t *testing.T) {
	for _, gone := range []string{"script.js", "output"} {
		t.Run(gone, func(t *testing.T) {
			runDir, anchor, sha := provenRun(t, "console.log('no divisors in 2..6');")
			path := filepath.Join(runDir, "proofs", sha, gone)
			if err := os.Remove(path); err != nil {
				t.Fatalf("the fixture's artifact is not where the store puts it: %v", err)
			}
			out, err := run(t, "show", "--run", runDir, "--seat-id", "red-lens-evidence", "evidence", "--anchor", anchor)
			if err == nil {
				t.Fatalf("a proof with no %s in the store was shown as if whole:\n%s", gone, out)
			}
			if strings.TrimSpace(out) != "" {
				t.Errorf("the refusal also printed a document a pipeline would read as the answer:\n%s", out)
			}
			for _, want := range []string{"not found", filepath.Join(runDir, "proofs", sha), sha[:12]} {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("the refusal does not say %q: %v", want, err)
				}
			}
		})
	}
}

// AN ANCHOR NOTHING IS AT IS A REFUSAL TOO, for the reason `--id` on a view that cannot scope is:
// answering it with the whole table hands back a different question's answer.
func TestEvidenceAtAnAnchorNothingIsAtIsRefused(t *testing.T) {
	runDir, anchor, sha := provenRun(t, "console.log('no divisors in 2..6');")

	if out, err := run(t, "show", "--run", runDir, "--seat-id", "red-lens-evidence", "evidence", "--anchor", "p-0000dead"); err == nil {
		t.Errorf("an anchor on no entry was answered:\n%s", out)
	} else if !strings.Contains(err.Error(), "p-0000dead") {
		t.Errorf("the refusal does not name the anchor it was given: %v", err)
	}

	// THE SHA IS WHAT `reproduce` TAKES AND NOT WHAT THIS TAKES. A seat holding it is one lookup
	// from the anchor, and the refusal is that lookup.
	_, err := run(t, "show", "--run", runDir, "--seat-id", "red-lens-evidence", "evidence", "--anchor", sha)
	if err == nil {
		t.Fatal("a sha256 was accepted as a document anchor — two names for one read")
	}
	if !strings.Contains(err.Error(), anchor) {
		t.Errorf("the refusal for a proof's sha256 does not name that proof's anchor %s: %v", anchor, err)
	}
}

// A SOURCE'S ANCHOR READS ITS ONE ROW. The flag is the document anchor, and the evidence layer
// holds two kinds of them; an anchor that scoped proofs and was refused for citations would be a
// lookup table that answers half its keys.
func TestASourceReadAtItsAnchorIsItsOneRow(t *testing.T) {
	runDir := citedRun(t, "https://example.org/primes")
	label := citeSentence(t, runDir, "The sky is blue.", "https://example.org/primes")
	out, err := run(t, "show", "--run", runDir, "--seat-id", blueSeat, "evidence", "--anchor", label)
	if err != nil {
		t.Fatalf("show evidence --anchor %s: %v", label, err)
	}
	var got struct {
		Anchor string `json:"anchor"`
		Source *struct {
			Anchor string `json:"anchor"`
			URL    string `json:"url"`
		} `json:"source"`
		Proof json.RawMessage `json:"proof"`
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("not JSON: %v\n%s", err, out)
	}
	if got.Source == nil || got.Source.Anchor != label || got.Source.URL != "https://example.org/primes" {
		t.Errorf("the citation's anchor did not resolve to its row:\n%s", out)
	}
	if len(got.Proof) != 0 {
		t.Errorf("a citation's anchor also returned a proof: %s", got.Proof)
	}
}
