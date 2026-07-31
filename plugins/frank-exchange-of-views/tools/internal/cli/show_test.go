package cli

import (
	"encoding/json"
	"strings"
	"testing"
)

// READS THROUGH THE TOOL.
//
// The seats could write every act through feov-record while still READING the run by
// opening markdown at paths they learned from a prompt. That asymmetry is how a seat comes
// to trust a hand-written artifact over the event log, and it is why the two were able to
// disagree by 9 open / 9 closed against 3 open / 15 closed without anything noticing.
//
// `show` closes it. These tests hold the property that makes it safe to rely on.

// THE INVARIANT: show is a thin wrapper over the ONE shared computation.
//
// There is now a single derivation for each view (internal/view), so show cannot be a
// SECOND reader of the artifact — the defect class this tool exists to remove. This test
// holds show's bytes to exactly that shared computation, so a divergent re-derivation at
// the read surface fails loudly.
func TestShowPrintsExactlyTheSharedProjection(t *testing.T) {
	runDir := seatRun(t)
	id := mintGap(t, runDir, "shown-gap", "read-surface")
	if _, err := run(t, "merge", "close", "--run", runDir, "--seat-id", "red-merge-r1",
		"--id", id, "--as", "closed",
		"--anchor-seat", "L1", "--anchor-tool", "go test", "--anchor-target", "./internal/x",
		"--reason", "the check passes at the named site"); err != nil {
		t.Fatalf("close: %v", err)
	}

	for _, name := range []string{"ledger", "archive", "debate", "changelog", "citation-ledger", "lines-of-inquiry"} {
		t.Run(name, func(t *testing.T) {
			out, err := run(t, "merge", "show", "--run", runDir, "--seat-id", "red-merge-r1", "--view", name)
			if err != nil {
				t.Fatalf("show --view %s: %v", name, err)
			}
			want := readProjection(t, runDir, name)
			if out != want {
				t.Errorf("show --view %s does not match the shared view.Markdown computation byte for byte — a re-derivation at the read surface is a second reader.\nstdout (%d bytes):\n%s\ncomputed (%d bytes):\n%s",
					name, len(out), out, len(want), want)
			}
		})
	}
}

// Each role's bare `show` gives it the artifact it actually works against, so a seat never
// has to know the run's file layout to see its own state.
func TestBareShowGivesEachRoleItsOwnView(t *testing.T) {
	runDir := seatRun(t)
	mintGap(t, runDir, "role-views", "read-surface")

	// merge's own view is the structured BOARD, not the markdown ledger: the merge seat
	// is the one that acts on gap state, and it should get it in the form it acts on
	// rather than prose it has to parse back.
	for _, c := range []struct{ role, marker string }{
		{"merge", `"counts"`},
		{"blue", "CHANGELOG"},
		{"lens", "citation-ledger"},
		{"bench", "debate.md"},
	} {
		t.Run(c.role, func(t *testing.T) {
			seat := map[string]string{
				"merge": "red-merge-r1", "blue": "blue-respond-r1",
				"lens": "red-lens-r1-L1", "bench": "judge-r1",
			}[c.role]
			out, err := run(t, c.role, "show", "--run", runDir, "--seat-id", seat)
			if err != nil {
				t.Fatalf("%s show: %v", c.role, err)
			}
			if !strings.Contains(out, c.marker) {
				t.Errorf("a bare `%s show` returned something other than its own view (wanted a header naming %s):\n%s", c.role, c.marker, firstLines(out, 3))
			}
		})
	}
}

// An unknown view must be REFUSED and must say what the options are. A seat's whole
// contract is --help plus the error text; a bare failure teaches it nothing and it will
// improvise a file read instead, which is the behaviour this verb exists to replace.
func TestUnknownViewIsRefusedWithTheListOfViews(t *testing.T) {
	runDir := seatRun(t)
	_, err := run(t, "merge", "show", "--run", runDir, "--seat-id", "red-merge-r1", "--view", "the-board")
	if err == nil {
		t.Fatal("an unknown view was accepted; a seat would get an empty read and no signal that it asked for something that does not exist")
	}
	for _, want := range []string{"the-board", "ledger"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal does not mention %q — it must name what was asked for AND what is available: %v", want, err)
		}
	}
}

// show mutates no events. It renders (which is idempotent) and prints; a read that
// appended to the log would corrupt every metric derived from event counts.
func TestShowRecordsNothing(t *testing.T) {
	runDir := seatRun(t)
	mintGap(t, runDir, "no-write-on-read", "read-surface")
	before := len(events(t, runDir))

	for i := 0; i < 3; i++ {
		if _, err := run(t, "merge", "show", "--run", runDir, "--seat-id", "red-merge-r1"); err != nil {
			t.Fatalf("show: %v", err)
		}
	}
	if after := len(events(t, runDir)); after != before {
		t.Errorf("three reads added %d events (%d → %d); a read that writes inflates every count derived from the log", after-before, before, after)
	}
}

// --json on a read opts into a view's STRUCTURED form where one exists. `debate` is the one
// view with both a markdown transcript and a JSON form; the audits count its sections from
// the JSON instead of regexing the prose. The contract is one-way: --json is an error on a
// view that is already JSON by name, and on a markdown view with no JSON form — so there is
// exactly one way to reach each form, and a wrong guess fails loudly.
func TestDebateJSONViewAndOneWayContract(t *testing.T) {
	runDir := seatRun(t)
	mintGap(t, runDir, "debate-json", "read-surface")
	if _, err := run(t, "merge", "position", "--run", runDir, "--seat-id", "red-merge-r1",
		"--reason", "red's round narrative"); err != nil {
		t.Fatalf("position: %v", err)
	}

	// debate --json parses and carries the rounds structure.
	out, err := run(t, "merge", "show", "--run", runDir, "--seat-id", "red-merge-r1", "--view", "debate", "--json")
	if err != nil {
		t.Fatalf("show --view debate --json: %v", err)
	}
	var dj struct {
		Rounds []struct {
			Red []string `json:"red"`
		} `json:"rounds"`
	}
	if e := json.Unmarshal([]byte(out), &dj); e != nil {
		t.Fatalf("debate --json is not valid JSON (%v):\n%s", e, out)
	}
	found := false
	for _, r := range dj.Rounds {
		for _, red := range r.Red {
			if strings.Contains(red, "red's round narrative") {
				found = true
			}
		}
	}
	if !found {
		t.Errorf("debate --json did not carry the red position text:\n%s", out)
	}

	// --json on a JSON-by-name view is refused (no alias to that JSON).
	for _, v := range []string{"board", "findings", "friction"} {
		if _, err := run(t, "merge", "show", "--run", runDir, "--seat-id", "red-merge-r1", "--view", v, "--json"); err == nil {
			t.Errorf("--view %s --json was accepted; it must refuse (that view is already JSON by name)", v)
		}
	}
	// --json on a markdown view with no JSON form is refused.
	if _, err := run(t, "merge", "show", "--run", runDir, "--seat-id", "red-merge-r1", "--view", "ledger", "--json"); err == nil {
		t.Error("--view ledger --json was accepted; ledger has no JSON form and must refuse")
	}
}

func firstLines(s string, n int) string {
	parts := strings.SplitN(s, "\n", n+1)
	if len(parts) > n {
		parts = parts[:n]
	}
	return strings.Join(parts, "\n")
}
