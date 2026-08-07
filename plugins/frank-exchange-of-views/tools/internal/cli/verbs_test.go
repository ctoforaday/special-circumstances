package cli

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// The remaining verbs, driven through the real command tree. Each case asserts
// what the verb RECORDS, not merely that it exited zero: the payload key names
// are the record format, and a verb that writes the right message with the wrong
// key is a verb whose events no projection will find.

// seedReferents creates the entities these cases NAME: two gaps and an observation.
//
// Every cross-reference is checked at write time now, so a case that disposes O1 or rules
// on R1-1 must have an O1 and an R1-1 to point at. Before the checks landed these were
// invented ids that resolved to nothing, which is exactly the state the checks exist to
// refuse — the fixtures were demonstrating the bug.
func seedReferents(t *testing.T, runDir string) {
	t.Helper()
	for i := 0; i < 2; i++ {
		if _, err := run(t, "merge", "mint", "--run", runDir, "--seat-id", "red-merge-r1",
			"--key", fmt.Sprintf("seed-%d", i), "--class", "x", "--check-kind", "document", "--check", "c",
			"--likelihood", "medium", "--impact", "medium", "--problem", "p"); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := run(t, "lens", "observe", "--run", runDir, "--seat-id", "red-lens-r1-L1",
		"--label", "SEED-O1", "--kind", "note", "--reason", "a seeded observation"); err != nil {
		t.Fatal(err)
	}
	// STATE, not just referents. A dispute-respond needs a dispute to answer, and a
	// spot-check samples the ARCHIVE, so R1-3 is minted and closed to put something in
	// it. Verbs are refused on the wrong state now, so the fixture has to build the
	// world each verb actually operates in.
	if _, err := run(t, "blue", "dispute", "--run", runDir, "--seat-id", "blue-respond-r1",
		"--id", "R1-1", "--dimension", "severity", "--proposed", "low",
		"--reason", "the seeded dispute this fixture answers"); err != nil {
		t.Fatal(err)
	}
	if _, err := run(t, "merge", "mint", "--run", runDir, "--seat-id", "red-merge-r1",
		"--key", "seed-archived", "--class", "x", "--check-kind", "document", "--check", "c",
		"--likelihood", "medium", "--impact", "medium", "--problem", "p"); err != nil {
		t.Fatal(err)
	}
	if _, err := run(t, "merge", "close", "--run", runDir, "--seat-id", "red-merge-r1",
		"--id", "R1-3", "--as", "closed", "--anchor-seat", "L1", "--anchor-tool", "go test",
		"--anchor-target", "./x", "--reason", "closed so the archive is not empty"); err != nil {
		t.Fatal(err)
	}
}

func TestVerbPayloads(t *testing.T) {
	cases := []struct {
		name   string
		role   string
		seatID string
		args   []string
		typ    string
		// want is the payload the event must carry, by key.
		want map[string]string
		// absent keys must not appear at all.
		absent []string
		// stdout must contain this.
		says string
	}{
		{
			name: "lens observe defaults its kind to note",
			role: "lens", seatID: "red-lens-r1-L1",
			args: []string{"--label", "O1", "--reason", "a below-bar note"},
			typ:  "observe",
			want: map[string]string{"kind": "note", "label": "O1", "text": "a below-bar note"},
			says: "observation recorded",
		},
		{
			name: "lens observe takes an explicit kind",
			role: "lens", seatID: "red-lens-r1-L1",
			args: []string{"--kind", "checked-held", "--label", "O2", "--reason", "checked and held"},
			typ:  "observe",
			want: map[string]string{"kind": "checked-held", "label": "O2"},
			says: "awaiting merge disposition",
		},
		{
			name: "lens cite records the access date under its payload name",
			role: "lens", seatID: "red-lens-r1-L1",
			args: []string{"--claim", "the claim", "--reference", "https://example.test/a",
				"--confidence", "high", "--access-date", "2026-07-18"},
			typ: "cite",
			// The flag is --access-date; the payload key is access_date, and the
			// citation render reads the payload key.
			want: map[string]string{"claim": "the claim", "reference": "https://example.test/a",
				"confidence": "high", "access_date": "2026-07-18"},
			says: "citation recorded: https://example.test/a",
		},
		{
			name: "lens cite without an access date leaves the key absent",
			role: "lens", seatID: "red-lens-r1-L1",
			args:   []string{"--claim", "c", "--reference", "https://example.test/b", "--confidence", "low"},
			typ:    "cite",
			want:   map[string]string{"reference": "https://example.test/b"},
			absent: []string{"access_date"},
			says:   "citation recorded",
		},
		{
			name: "merge dispose gives an observation its fate",
			role: "merge", seatID: "red-merge-r1",
			args: []string{"--observation", "SEED-O1", "--as", "folded-into", "--into", "R1-2", "--reason", "same root cause"},
			typ:  "dispose",
			want: map[string]string{"observation": "SEED-O1", "disposition": "folded-into",
				"into": "R1-2", "reason": "same root cause"},
			says: "disposed SEED-O1: folded-into",
		},
		{
			name: "merge dispute-respond records red's answer",
			role: "merge", seatID: "red-merge-r1",
			args: []string{"--id", "R1-1", "--as", "rejected", "--reason", "the evidence does not reach it"},
			typ:  "dispute-respond",
			want: map[string]string{"gap_id": "R1-1", "response": "rejected",
				"rationale": "the evidence does not reach it"},
			says: "dispute on R1-1: rejected",
		},
		{
			name: "blue dispute contests a grade through the accounted channel",
			role: "blue", seatID: "blue-lane-1",
			args: []string{"--id", "R1-1", "--dimension", "severity", "--proposed", "low", "--reason", "§4 says otherwise"},
			typ:  "dispute",
			want: map[string]string{"gap_id": "R1-1", "dimension": "severity",
				"proposed": "low", "evidence": "§4 says otherwise"},
			says: "dispute filed on R1-1.severity",
		},
		{
			name: "blue manifest-row records the receipt",
			role: "blue", seatID: "blue-lane-1",
			args: []string{"--id", "R1-2", "--row", "figures recomputed; acceptance check run: pass"},
			typ:  "manifest-row",
			want: map[string]string{"gap_id": "R1-2", "row": "figures recomputed; acceptance check run: pass"},
			says: "manifest row recorded for R1-2",
		},
		{
			name: "blue retire records what left and why",
			role: "blue", seatID: "blue-lane-1",
			args: []string{"--claim", "the claim as it stood", "--reason", "refuted", "--superseded-by", "the replacement claim"},
			typ:  "retire",
			want: map[string]string{"claim": "the claim as it stood", "reason": "refuted",
				"superseded_by": "the replacement claim"},
			says: "retired: the claim as it stood",
		},
		{
			name: "blue avenue records a dead end and what killed it",
			role: "blue", seatID: "blue-lane-1",
			args: []string{"--line", "search the offline archive", "--status", "abandoned",
				"--reason", "the archive is unreachable", "--method", "full-text search"},
			typ: "avenue",
			want: map[string]string{"line": "search the offline archive", "status": "abandoned",
				"reason": "the archive is unreachable", "method": "full-text search"},
			says: "avenue A1 recorded (abandoned): search the offline archive",
		},
		{
			name: "blue confidence records the calibration substrate",
			role: "blue", seatID: "blue-lane-1",
			// The flag words are --claim/--confidence; the payload keys stay label/grade.
			args: []string{"--claim", "C1", "--confidence", "medium"},
			typ:  "confidence",
			want: map[string]string{"label": "C1", "grade": "medium"},
			says: "confidence recorded for C1",
		},
		{
			name: "bench petition-rule records the ruling and its opinion",
			role: "bench", seatID: "judge-petition",
			args: []string{"--petitioner", "red-lens-r1-L1", "--petition-class", "safety",
				"--as", "granted", "--reason", "the written opinion"},
			typ: "petition-rule",
			want: map[string]string{"petitioner": "red-lens-r1-L1", "class": "safety",
				"ruling": "granted", "opinion": "the written opinion"},
			says: "petition granted (red-lens-r1-L1)",
		},
		{
			name: "bench halt is the safety boundary",
			role: "bench", seatID: "judge-terminal",
			args: []string{"--reason", "the run must stop, and here is why"},
			typ:  "halt",
			want: map[string]string{"opinion": "the run must stop, and here is why"},
			says: "JUDICIAL HALT recorded — capture relays this verbatim",
		},
		{
			name: "bench certify is the run-end statement",
			role: "bench", seatID: "assemble",
			args: []string{"--reason", "what I would want a human to re-examine"},
			typ:  "certify",
			want: map[string]string{"statement": "what I would want a human to re-examine"},
			says: "certification recorded",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runDir := t.TempDir()
			seedReferents(t, runDir)
			verb := strings.SplitN(tc.name, " ", 3)[1]
			args := append([]string{tc.role, verb, "--run", runDir, "--seat-id", tc.seatID}, tc.args...)
			out, err := run(t, args...)
			if err != nil {
				t.Fatalf("%v: %v", args, err)
			}
			if !strings.Contains(out, tc.says) {
				t.Errorf("stdout = %q, want it to contain %q", out, tc.says)
			}
			ev := lastOfType(t, runDir, tc.typ)
			for k, want := range tc.want {
				if got := ev.Payload.Str(k); got != want {
					t.Errorf("payload[%q] = %q, want %q", k, got, want)
				}
			}
			keys := payloadKeys(ev)
			for _, k := range tc.absent {
				if keys[k] {
					t.Errorf("payload carries %q though the seat never passed it", k)
				}
			}
			if ev.SeatID != tc.seatID {
				t.Errorf("seatId = %q, want %q", ev.SeatID, tc.seatID)
			}
		})
	}
}

// spot-check's --ids is a CSV field, so it is ALWAYS present as an array — the
// same "lineage none, not lineage unknown" rule the mint lists follow.
func TestSpotCheckIdsAreAlwaysAnArray(t *testing.T) {
	t.Run("with ids", func(t *testing.T) {
		runDir := t.TempDir()
		seedReferents(t, runDir)
		out, err := run(t, "merge", "spot-check", "--run", runDir, "--seat-id", "red-merge-r1",
			"--ids", "R1-3", "--notes", "it still holds")
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(out, "spot-checked R1-3") {
			t.Errorf("stdout = %q", out)
		}
		ev := lastOfType(t, runDir, "spot-check")
		got := ev.Payload.StrList("ids")
		if len(got) != 1 || got[0] != "R1-3" {
			t.Errorf("ids = %q, want [R1-3] — the only CLOSED gap, which is what a spot-check samples", got)
		}
		if ev.Payload.Str("notes") != "it still holds" {
			t.Errorf("notes = %q", ev.Payload.Str("notes"))
		}
	})

	t.Run("with no ids at all the key is still an empty array", func(t *testing.T) {
		runDir := t.TempDir()
		if _, err := run(t, "merge", "spot-check", "--run", runDir, "--seat-id", "red-merge-r1"); err != nil {
			t.Fatal(err)
		}
		ev := lastOfType(t, runDir, "spot-check")
		if !payloadKeys(ev)["ids"] {
			t.Fatal("the ids key is absent; an absent list reads as \"not checked\" rather than \"checked nothing\"")
		}
		b, err := json.Marshal(ev.Payload)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(b), `"ids":[]`) {
			t.Errorf("empty ids did not serialize as []: %s", b)
		}
	})
}

// spot-check is a singleton per seat: the round's duty is discharged once.
func TestSpotCheckIsASingleton(t *testing.T) {
	runDir := t.TempDir()
	seedReferents(t, runDir)
	for _, ids := range []string{"R1-3", "R1-3"} {
		if _, err := run(t, "merge", "spot-check", "--run", runDir, "--seat-id", "red-merge-r1", "--ids", ids); err != nil {
			t.Fatal(err)
		}
	}
	n := 0
	for _, e := range events(t, runDir) {
		if e.Type == "spot-check" {
			n++
		}
	}
	if n != 1 {
		t.Errorf("%d spot-check events survived, want 1", n)
	}
}

// regrade moves only the grades it carries, and refuses without its basis.
func TestRegradeMovesOnlyThePassedGrades(t *testing.T) {
	runDir := t.TempDir()
	seatID := "red-merge-r1"
	if _, err := run(t, "merge", "mint", "--run", runDir, "--seat-id", seatID,
		"--class", "x", "--check-kind", "document", "--check", "c", "--problem", "p",
		"--severity", "low", "--likelihood", "low", "--impact", "low"); err != nil {
		t.Fatal(err)
	}
	if _, err := run(t, "merge", "regrade", "--run", runDir, "--seat-id", seatID,
		"--id", "R1-1", "--severity", "certain", "--reason", "new evidence in §4"); err != nil {
		t.Fatal(err)
	}
	ev := lastOfType(t, runDir, "regrade")
	if got := ev.Payload.Str("severity"); got != "certain" {
		t.Errorf("severity = %q", got)
	}
	if got := ev.Payload.Str("basis"); got != "new evidence in §4" {
		t.Errorf("basis = %q", got)
	}
	// The grades NOT passed must be absent, so the replay leaves them alone.
	keys := payloadKeys(ev)
	for _, k := range []string{"likelihood", "impact", "complexity_cost"} {
		if keys[k] {
			t.Errorf("regrade carries %q though it was not passed — the replay would overwrite it", k)
		}
	}
	// The board reflects the move without losing the untouched grades.
	board, err := boardState(t, runDir)
	if err != nil {
		t.Fatal(err)
	}
	g := board.Gaps["R1-1"]
	if g.Severity != "certain" {
		t.Errorf("board severity = %v, want certain", g.Severity)
	}
	if g.Likelihood != "low" || g.Impact != "low" {
		t.Errorf("a grade the regrade did not carry moved: likelihood=%v impact=%v", g.Likelihood, g.Impact)
	}
}

// The prose channel is available on the verbs that declare it, and --file is the
// documented path for anything above trivial size.
func TestProseVerbsAcceptAFile(t *testing.T) {
	cases := []struct {
		role, verb, seatID, key string
		extra                   []string
	}{
		{"bench", "halt", "judge-terminal", "opinion", nil},
		{"bench", "certify", "assemble", "statement", nil},
		{"blue", "revision", "blue-lane-1", "text", nil},
		{"lens", "observe", "red-lens-r1-L1", "text", []string{"--label", "PROSE-O1"}},
		{"merge", "closing", "red-merge-r1", "text", []string{"--id", "R1-1"}},
		{"blue", "manifest-row", "blue-lane-1", "row", []string{"--id", "R1-1"}},
	}
	body := "a multi-line payload\nwith unicode — ✓ 日本語\nand <angle> & entities\n"
	for _, tc := range cases {
		t.Run(tc.role+"/"+tc.verb, func(t *testing.T) {
			runDir := t.TempDir()
			seedReferents(t, runDir)
			args := append([]string{tc.role, tc.verb, "--run", runDir, "--seat-id", tc.seatID,
				"--reason-file", writeTemp(t, body)}, tc.extra...)
			if _, err := run(t, args...); err != nil {
				t.Fatal(err)
			}
			// BY TYPE, not "the last event in the log". That only worked while the run
			// dir was otherwise empty: MergedEvents is not time-ordered, so once the
			// fixture seeds referents the tail of the slice is whichever shard sorts
			// last, not whichever event happened last.
			last := lastOfType(t, runDir, tc.verb)
			// Less the file's terminating newline: that is a line terminator every
			// editor appends, not content the seat chose to record.
			if got := last.Payload.Str(tc.key); got != strings.TrimRight(body, "\n") {
				t.Errorf("payload[%q] = %q, want the file's content without its terminator", tc.key, got)
			}
		})
	}
}
