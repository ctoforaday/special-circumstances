package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// A COMPUTATION'S SCRIPT IS READ THROUGH THE RECORD'S PROJECTION.
//
// The lens owes a soundness verdict that rests on READING the script, and `reproduce` records that
// verdict in the act that re-runs — so the read has to come first and has to be somewhere. Wave A
// (2026-10-09) had no such place: six log entries from five lenses say the evidence view gave only
// the sha256, and each found the file by listing a `computations/` directory nothing had told it about,
// against a prompt that says every read is a projection. One filed three placeholder verdicts
// instead.

// provenRun records one computation and returns its anchor and sha256 as the unscoped table lists them.
// Its cast seats every lens; the computations are the computation lens's.
func provenRun(t *testing.T, body string) (runDir, anchor, sha string) {
	t.Helper()
	runDir = newRun(t)
	cast, lanes := record.CastFor(1)
	if _, err := record.Append(record.Identity{Run: runtest.Open(t, runDir), SeatID: record.HarnessSeat},
		&recordpb.Cast{SeatIds: cast, LaneSeatIds: lanes}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	seat := computeSeat(t, runDir, "# H\n\nSeven has no divisor between two and six.\n")
	s := script(t, runDir, "seven.js", body)
	if _, err := run(t, "compute", "--run", runDir, "--seat-id", seat,
		"--quote", "Seven has no divisor between two and six.", "--dir", s, "--reason", "r", "--", "node", "seven.js"); err != nil {
		t.Fatal(err)
	}
	p := lastBody(t, runDir, &recordpb.Computation{})
	for _, lens := range []string{"red-lens-evidence", "red-lens-computation"} {
		if _, err := run(t, "register", "--run", runDir, "--seat-id", lens); err != nil {
			t.Fatal(err)
		}
	}
	return runDir, p.GetComputationId(), p.GetBundleSha()
}

func TestAComputationReadAtItsAnchorCarriesItsBundleAndRecordedOutput(t *testing.T) {
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

	// THE WORK ITEM THAT SENDS A LENS TO THE COMPUTATION SAYS WHERE ITS SCRIPT IS READ, before it names
	// the act that re-runs it and records the verdict.
	work, err := run(t, "show", "--run", runDir, "--seat-id", "red-lens-computation", "work")
	if err != nil {
		t.Fatalf("show work: %v", err)
	}
	read, rerun := strings.Index(work, "show evidence --anchor "+anchor), strings.Index(work, "reproduce --anchor "+anchor)
	if read < 0 || rerun < 0 || read > rerun {
		t.Errorf("the lens's work item for an un-re-run computation must name the read at its anchor BEFORE the act that re-runs it (read at %d, re-run at %d):\n%s", read, rerun, work)
	}

	// BOTH PARTIES READ IT: the lens that judges the script, and blue, who wrote it and answers for it.
	for _, seat := range []string{"red-lens-computation", "blue-respond"} {
		out, err := run(t, "show", "--run", runDir, "--seat-id", seat, "evidence", "--anchor", anchor)
		if err != nil {
			t.Fatalf("%s: show evidence --anchor %s: %v", seat, anchor, err)
		}
		var got struct {
			Anchor      string `json:"anchor"`
			Computation *struct {
				Anchor string `json:"anchor"`
				Sha256 string `json:"sha256"`
				Files  []struct {
					Path string `json:"path"`
					Body string `json:"body"`
				} `json:"files"`
				Output *string `json:"output"`
			} `json:"computation"`
			Source json.RawMessage `json:"source"`
		}
		if err := json.Unmarshal([]byte(out), &got); err != nil {
			t.Fatalf("%s: the scoped read is not JSON: %v\n%s", seat, err, out)
		}
		if got.Computation == nil {
			t.Fatalf("%s: a computation's anchor resolved to no computation:\n%s", seat, out)
		}
		if got.Anchor != anchor || got.Computation.Anchor != anchor || got.Computation.Sha256 != sha {
			t.Errorf("%s: asked for %s (%s), got anchor %q / computation %q (%s)", seat, anchor, sha, got.Anchor, got.Computation.Anchor, got.Computation.Sha256)
		}
		// THE BYTES `reproduce` EXECUTES, not a summary of them.
		// The file keeps the name it has in the bundle: it is what the command names.
		if fs := got.Computation.Files; len(fs) != 1 || fs[0].Path != "seven.js" || fs[0].Body != body {
			t.Errorf("%s: the bundle is not the stored bytes:\n got %+v\nwant seven.js %q", seat, fs, body)
		}
		if got.Computation.Output == nil || !strings.Contains(*got.Computation.Output, "divisors of 7 in 2..6: 0") {
			t.Errorf("%s: the recorded output is not the stored bytes: %v", seat, got.Computation.Output)
		}
		if len(got.Source) != 0 {
			t.Errorf("%s: a computation's anchor also returned a source: %s", seat, got.Source)
		}
	}
}

// A MISSING ARTIFACT IS A REFUSAL, NEVER AN EMPTY BUNDLE. `"files": []` reads exactly like a
// program that does nothing, and a lens would judge it unsound — a verdict about the store,
// recorded as a verdict about blue's computation.
func TestAComputationWhoseArtifactIsMissingIsRefusedNotShownEmpty(t *testing.T) {
	for _, gone := range []string{"dir/seven.js", "acts/<id>/output"} {
		t.Run(gone, func(t *testing.T) {
			runDir, anchor, sha := provenRun(t, "console.log('no divisors in 2..6');")
			path := filepath.Join(runDir, "computations", sha, filepath.FromSlash(strings.Replace(gone, "<id>", anchor, 1)))
			if err := os.Remove(path); err != nil {
				t.Fatalf("the fixture's artifact is not where the store puts it: %v", err)
			}
			out, err := run(t, "show", "--run", runDir, "--seat-id", "red-lens-evidence", "evidence", "--anchor", anchor)
			if err == nil {
				t.Fatalf("a computation with no %s in the store was shown as if whole:\n%s", gone, out)
			}
			if strings.TrimSpace(out) != "" {
				t.Errorf("the refusal also printed a document a pipeline would read as the answer:\n%s", out)
			}
			for _, want := range []string{"not found", filepath.Join(runDir, "computations", sha), sha[:12]} {
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
	runDir, _, _ := provenRun(t, "console.log('no divisors in 2..6');")

	if out, err := run(t, "show", "--run", runDir, "--seat-id", "red-lens-evidence", "evidence", "--anchor", "K-0000dead"); err == nil {
		t.Errorf("an anchor on no entry was answered:\n%s", out)
	} else if !strings.Contains(err.Error(), "K-0000dead") {
		t.Errorf("the refusal does not name the anchor it was given: %v", err)
	}
}

// ONE ANCHOR FLAG, ON BOTH READS THAT TAKE AN ANCHOR. The report read and the evidence read take the
// same id, so an id of the wrong shape — an old-case prefix, a computation's sha256, a word — is refused by
// the one declaration in the same words on each, and a token pasted whole from the report is the id
// inside it on each.
func TestBothAnchorReadsRefuseAMalformedIDInTheSameWords(t *testing.T) {
	runDir, id, sha := provenRun(t, "console.log('no divisors in 2..6');")
	for _, bad := range []string{"p-0000dead", sha, "the-computation"} {
		shared := flags.AnchorID().Set(bad)
		if shared == nil {
			t.Fatalf("the fixture id %q is a well-formed anchor", bad)
		}
		for _, view := range []string{"report", "evidence"} {
			out, err := run(t, "show", "--run", runDir, "--seat-id", "red-lens-evidence", view, "--anchor", bad)
			if err == nil {
				t.Errorf("show %s --anchor %s was answered:\n%s", view, bad, out)
			} else if !strings.Contains(err.Error(), shared.Error()) {
				t.Errorf("show %s --anchor %s is not refused in the shared flag's words\n got: %v\nwant: %v", view, bad, err, shared)
			}
		}
	}
	out, err := run(t, "show", "--run", runDir, "--seat-id", "red-lens-evidence", "evidence", "--anchor", anchor.Token(id))
	if err != nil {
		t.Fatalf("the computation's token, pasted whole, is refused: %v", err)
	}
	if !strings.Contains(out, `"anchor": "`+id+`"`) {
		t.Errorf("the token did not read the entry at %s:\n%s", id, out)
	}
}

// A SOURCE'S ANCHOR READS ITS ONE ROW. The flag is the document anchor, and the evidence layer
// holds two kinds of them; an anchor that scoped computations and was refused for citations would be a
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
		Computation json.RawMessage `json:"computation"`
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("not JSON: %v\n%s", err, out)
	}
	if got.Source == nil || got.Source.Anchor != label || got.Source.URL != "https://example.org/primes" {
		t.Errorf("the citation's anchor did not resolve to its row:\n%s", out)
	}
	if len(got.Computation) != 0 {
		t.Errorf("a citation's anchor also returned a computation: %s", got.Computation)
	}
}
