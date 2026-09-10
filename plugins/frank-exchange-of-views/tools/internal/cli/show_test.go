package cli

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
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
	if _, err := run(t, "close", "--run", runDir, "--seat-id", lensSeat,
		"--id", id, "--as", "repaired",
		"--verified-by", "L1", "--verified-with", "go test", "--verified-against", "./internal/x",
		"--reason", "the check passes at the named site"); err != nil {
		t.Fatalf("close: %v", err)
	}

	// The markdown views, which are the only ones this contract can hold: a JSON-by-name view is
	// not a view.Markdown rendering, so there is no shared computation to diverge from.
	for _, name := range []string{"debate", "lines-of-inquiry"} {
		t.Run(name, func(t *testing.T) {
			out, err := run(t, "show", "--run", runDir, "--seat-id", "red-chair", name)
			if err != nil {
				t.Fatalf("show %s: %v", name, err)
			}
			want := readProjection(t, runDir, name)
			if out != want {
				t.Errorf("show %s does not match the shared view.Markdown computation byte for byte — a re-derivation at the read surface is a second reader.\nstdout (%d bytes):\n%s\ncomputed (%d bytes):\n%s",
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

	// merge's own view is the structured WORKLIST — its shrinking working set (OPEN gaps
	// lean + a prose-free closed_index), the once-per-turn read it acts on. It is the
	// last-wins default among the merge's views (board, findings, work list all default to
	// merge; work list is registered last), so a bare `merge show` resolves here. The marker
	// is `closed_index`, which ONLY the work list carries — board/findings would also match a
	// bare `"counts"`, so pinning on the unique key is what fixes the default to work list.
	// EVERY ROLE NOW DEFAULTS TO ITS OWN PENDING WORK, which is what a bare `show` should
	// answer. It did not: blue got `changelog` — a record of what it had ALREADY done, before it
	// had done anything — the lens got `citation-ledger`, and the bench got `debate`. Asked what
	// would tell them a sitting was finished, only the merge could name a mechanism; the others
	// answered with another seat's future act, which is not observable when they must decide to
	// stop. The marker is `sitting`, the block that says what is outstanding and whether anything
	// is.
	for _, c := range []struct{ role, marker string }{
		{"merge", `"sitting"`},
		{"blue", `"sitting"`},
		{"lens", `"sitting"`},
		{"bench", `"sitting"`},
	} {
		t.Run(c.role, func(t *testing.T) {
			seat := map[string]string{
				"merge": "red-chair", "blue": "blue-respond",
				"lens": "red-lens-evidence", "bench": "judge",
			}[c.role]
			out, err := run(t, "show", "--run", runDir, "--seat-id", seat)
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
	_, err := run(t, "show", "--run", runDir, "--seat-id", "red-chair", "the-board")
	if err == nil {
		t.Fatal("an unknown view was accepted; a seat would get an empty read and no signal that it asked for something that does not exist")
	}
	for _, want := range []string{"the-board", "evidence"} {
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
		if _, err := run(t, "show", "--run", runDir, "--seat-id", "red-chair"); err != nil {
			t.Fatalf("show: %v", err)
		}
	}
	if after := len(events(t, runDir)); after != before {
		t.Errorf("three reads added %d events (%d → %d); a read that writes inflates every count derived from the log", after-before, before, after)
	}
}

// --json on a read opts into a view's STRUCTURED form where one exists. `debate` is the one
// view with both a markdown transcript and a JSON form; the audits count its sections from
// the JSON instead of regexing the prose. On a view that is already JSON by name --json is
// accepted and changes nothing — the same bytes are one form, not two — and on a markdown view
// with no JSON form it is an error, because the tool cannot give what was asked for.
func TestDebateJSONViewAndOneWayContract(t *testing.T) {
	runDir := seatRun(t)
	mintGap(t, runDir, "debate-json", "read-surface")
	if _, err := run(t, "position", "--run", runDir, "--seat-id", "red-chair",
		"--reason", "red's round narrative"); err != nil {
		t.Fatalf("position: %v", err)
	}

	// debate --json parses and carries the epochs structure.
	out, err := run(t, "show", "--run", runDir, "--seat-id", "red-chair", "debate", "--json")
	if err != nil {
		t.Fatalf("show debate --json: %v", err)
	}
	var dj struct {
		Epochs []struct {
			Red []string `json:"red"`
		} `json:"epochs"`
	}
	if e := json.Unmarshal([]byte(out), &dj); e != nil {
		t.Fatalf("debate --json is not valid JSON (%v):\n%s", e, out)
	}
	found := false
	for _, r := range dj.Epochs {
		for _, red := range r.Red {
			if strings.Contains(red, "red's round narrative") {
				found = true
			}
		}
	}
	if !found {
		t.Errorf("debate --json did not carry the red position text:\n%s", out)
	}

	// --json on a JSON-by-name view is the SAME BYTES as the bare view. It used to be refused,
	// and the refusal's envelope parsed as cleanly as the data: six seats crashed on a missing
	// key (#593), and #861's B3 spent six turns on the refusal. Asserted as byte-identity, not
	// as "no error" — an accepted flag that quietly changed the output would be a second form.
	//
	// NAMES FROM seat.JSONByNameViews(), not a hand-kept list: two hand-kept lists here checked
	// stale names (`friction`, `reason`) while `evidence` went unheld. A name cannot be here
	// unless the table says so, and no marked view can be absent.
	for _, v := range seat.JSONByNameViews() {
		bare, err := run(t, "show", "--run", runDir, "--seat-id", "red-chair", v)
		if err != nil {
			t.Fatalf("show %s: %v", v, err)
		}
		flagged, err := run(t, "show", "--run", runDir, "--seat-id", "red-chair", v, "--json")
		if err != nil {
			t.Errorf("show %s --json was refused; that view is already JSON, so the flag must change nothing: %v", v, err)
			continue
		}
		if flagged != bare {
			t.Errorf("show %s --json differs from show %s — an accepted flag that changes the bytes is a second form", v, v)
		}
	}
	// Asking for two forms at once is still refused: the board's markdown AND its JSON.
	if _, err := run(t, "show", "--run", runDir, "--seat-id", "red-chair", "board", "--json", "--format", "markdown"); err == nil {
		t.Error("show board --json --format markdown was accepted; it asks for two forms and must refuse")
	}
	// A bare `show` answers with pending work, which is JSON by name — so a bare `show --json` is
	// the same bytes too, by the same rule, and not a second way to ask for anything.
	bareShow, err := run(t, "show", "--run", runDir, "--seat-id", "red-chair")
	if err != nil {
		t.Fatalf("bare show: %v", err)
	}
	if got, err := run(t, "show", "--run", runDir, "--seat-id", "red-chair", "--json"); err != nil || got != bareShow {
		t.Errorf("a bare `show --json` must be the same bytes as a bare `show` (err=%v)", err)
	}
	// --json on a markdown view with no JSON form is refused.
	if _, err := run(t, "show", "--run", runDir, "--seat-id", "red-chair", "lines-of-inquiry", "--json"); err == nil {
		t.Error("show lines-of-inquiry --json was accepted; it has no JSON form and must refuse")
	}
}

func firstLines(s string, n int) string {
	parts := strings.SplitN(s, "\n", n+1)
	if len(parts) > n {
		parts = parts[:n]
	}
	return strings.Join(parts, "\n")
}
