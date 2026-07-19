package cli

import (
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

// THE INVARIANT: stdout and the file are the same bytes.
//
// If show re-derived the markdown instead of rendering and reading, it would be a SECOND
// reader of one artifact — the defect class this tool exists to remove, reintroduced at
// the read surface, and invisible until a seat and a human read the same run differently.
func TestShowPrintsExactlyWhatTheProjectionFileContains(t *testing.T) {
	runDir := seatRun(t)
	id := mintGap(t, runDir, "shown-gap", "read-surface")
	if _, err := run(t, "merge", "close", "--run", runDir, "--seat-id", "red-merge-r1",
		"--id", id, "--as", "closed",
		"--anchor-seat", "L1", "--anchor-tool", "go test", "--anchor-target", "./internal/x",
		"--text", "the check passes at the named site"); err != nil {
		t.Fatalf("close: %v", err)
	}

	for _, view := range []struct{ name, file string }{
		{"ledger", "ledger.md"},
		{"archive", "archive.md"},
		{"debate", "debate.md"},
		{"changelog", "CHANGELOG.md"},
		{"citation-ledger", "citation-ledger.md"},
		{"lines-of-inquiry", "lines-of-inquiry.md"},
	} {
		t.Run(view.name, func(t *testing.T) {
			out, err := run(t, "merge", "show", "--run", runDir, "--seat-id", "red-merge-r1", "--view", view.name)
			if err != nil {
				t.Fatalf("show --view %s: %v", view.name, err)
			}
			want := readProjection(t, runDir, view.file)
			if out != want {
				t.Errorf("show --view %s does not match %s byte for byte. A read path that re-derives its answer is a second reader of one artifact, and the two WILL drift.\nstdout (%d bytes):\n%s\nfile (%d bytes):\n%s",
					view.name, view.file, len(out), out, len(want), want)
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

func firstLines(s string, n int) string {
	parts := strings.SplitN(s, "\n", n+1)
	if len(parts) > n {
		parts = parts[:n]
	}
	return strings.Join(parts, "\n")
}
