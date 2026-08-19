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
// Every cross-reference is checked at write time now, so a case that rules
// on R1-1 must have an O1 and an R1-1 to point at. Before the checks landed these were
// invented ids that resolved to nothing, which is exactly the state the checks exist to
// refuse — the fixtures were demonstrating the bug.
func seedReferents(t *testing.T, runDir string) {
	t.Helper()
	for i := 0; i < 2; i++ {
		if _, err := run(t, "mint", "--run", runDir, "--seat-id", "red-merge-r1",
			"--key", fmt.Sprintf("seed-%d", i), "--class", "x", "--check-kind", "document", "--check", "c",
			"--likelihood", "medium", "--impact", "medium", "--problem", "p"); err != nil {
			t.Fatal(err)
		}
	}
	// STATE, not just referents. A dispute-respond needs a dispute to answer, and a
	// spot-check samples the ARCHIVE, so R1-3 is minted and closed to put something in
	// it. Verbs are refused on the wrong state now, so the fixture has to build the
	// world each verb actually operates in.
	// M1 and M2: the motions the ruling cases answer. A rule names the motion it answers, so
	// the filing has to exist before the ruling can be tested at all — which is the join the
	// collapse exists to make, and the reason these are seeded rather than assumed.
	if _, err := run(t, "motion", "grade", "file", "--run", runDir, "--seat-id", "blue-respond-r1",
		"--id", "R1-1", "--dimension", "severity", "--proposed", "low",
		"--reason", "the seeded grade motion this fixture answers"); err != nil {
		t.Fatal(err)
	}
	if _, err := run(t, "motion", "petition", "file", "--run", runDir, "--seat-id", "blue-respond-r1",
		"--class", "safety", "--relief", "the relief sought",
		"--reason", "the seeded petition this fixture answers"); err != nil {
		t.Fatal(err)
	}
	if _, err := run(t, "mint", "--run", runDir, "--seat-id", "red-merge-r1",
		"--key", "seed-archived", "--class", "x", "--check-kind", "document", "--check", "c",
		"--likelihood", "medium", "--impact", "medium", "--problem", "p"); err != nil {
		t.Fatal(err)
	}
	if _, err := run(t, "close", "--run", runDir, "--seat-id", "red-merge-r1",
		"--id", "R1-3", "--as", "closed", "--verified-by", "L1", "--verified-with", "go test",
		"--verified-against", "./x", "--reason", "closed so the archive is not empty"); err != nil {
		t.Fatal(err)
	}
	// THE LENS SEAT MUST HAVE SAT. `petition-rule --petitioner` refuses a seat that recorded
	// nothing in the run, and the lens's presence used to come from a seeded `observe` — retired
	// with #327. `friction` is the lens verb with no referents of its own, so it seeds presence
	// without seeding state any case then has to work around.
	if _, err := run(t, "friction", "--run", runDir, "--seat-id", "red-lens-r1-L1",
		"--reason", "seeded so the lens seat has sat"); err != nil {
		t.Fatal(err)
	}
}

func TestVerbPayloads(t *testing.T) {
	cases := []struct {
		name string
		role string
		// path is the command path when it is not <role> <second word of name>. The default
		// recovers the verb from the TEST NAME, which is a fact composed into a string and
		// split apart again — it works only for a two-level tree, and `motion grade file` is
		// three. Rather than re-derive a deeper path from a longer name, the cases that need
		// one say so.
		path   []string
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
			name: "lens corroborate records the access date under its payload name",
			path: []string{"corroborate"}, seatID: "red-lens-r1-L1",
			args: []string{"--quote", "the claim", "--url", "https://example.test/a", "--title", "Example A",
				"--as", "supports", "--confidence", "high", "--reason", "read at the leaf",
				"--access-date", "2026-07-18"},
			typ: "verify",
			// The flag is --access-date; the payload key is access_date, and the
			// citation render reads the payload key. --as lands under `outcome`.
			want: map[string]string{"claim": "the claim", "url": "https://example.test/a", "title": "Example A",
				"outcome": "supports", "confidence": "high", "reason": "read at the leaf", "access_date": "2026-07-18"},
			says: "corroborating source Example A verified: supports",
		},
		{
			name: "lens corroborate without an access date leaves the key absent",
			path: []string{"corroborate"}, seatID: "red-lens-r1-L1",
			args: []string{"--quote", "c", "--url", "https://example.test/b", "--title", "Example B",
				"--as", "weak", "--confidence", "low", "--reason", "it gestures at it"},
			typ:    "verify",
			want:   map[string]string{"url": "https://example.test/b", "title": "Example B", "outcome": "weak", "confidence": "low"},
			absent: []string{"access_date"},
			says:   "corroborating source Example B verified: weak",
		},
		{
			name: "motion grade rule records the merge's answer",
			path: []string{"motion", "grade", "rule"}, seatID: "red-merge-r1",
			args: []string{"--id", "M1", "--as", "rejected", "--reason", "the evidence does not reach it"},
			typ:  "motion-rule",
			want: map[string]string{"motion_id": "M1", "subject": "grade", "ruling": "rejected",
				"reason": "the evidence does not reach it"},
			says: "motion M1 ruled rejected",
		},
		{
			name: "motion grade file contests a grade through the accounted channel",
			path: []string{"motion", "grade", "file"}, seatID: "blue-lane-1",
			args: []string{"--id", "R1-1", "--dimension", "severity", "--proposed", "low", "--reason", "§4 says otherwise"},
			typ:  "motion",
			want: map[string]string{"gap_id": "R1-1", "dimension": "severity",
				"proposed": "low", "reason": "§4 says otherwise", "subject": "grade"},
			says: "filed (grade)",
		},
		{
			name: "blue manifest-row records the receipt",
			role: "blue", seatID: "blue-lane-1",
			args: []string{"--id", "R1-2", "--reason", "figures recomputed; acceptance check run: pass"},
			typ:  "manifest-row",
			want: map[string]string{"gap_id": "R1-2", "row": "figures recomputed; acceptance check run: pass"},
			says: "manifest row recorded for R1-2",
		},
		{
			name: "blue retire records what left and why",
			role: "blue", seatID: "blue-lane-1",
			args: []string{"--quote", "the claim as it stood", "--reason", "refuted", "--new", "the replacement claim"},
			typ:  "retire",
			want: map[string]string{"claim": "the claim as it stood", "reason": "refuted",
				"superseded_by": "the replacement claim"},
			says: "retired: the claim as it stood",
		},
		{
			name: "blue line of inquiry propose records the direction and its hypothesis",
			path: []string{"line-of-inquiry", "propose"}, seatID: "blue-lane-1",
			args: []string{"--reason", "search the offline archive", "--method", "full-text search",
				"--hypothesis", "the 1997 proceedings are scanned"},
			typ: "line-of-inquiry",
			want: map[string]string{"line": "search the offline archive", "status": "proposed",
				"reason": "search the offline archive", "method": "full-text search",
				"hypothesis": "the 1997 proceedings are scanned"},
			says: "line of inquiry Q1 recorded (proposed): search the offline archive",
		},
		{
			name: "motion petition rule records the ruling and its opinion",
			path: []string{"motion", "petition", "rule"}, seatID: "judge-petition-red-merge-r1",
			args: []string{"--id", "M2", "--as", "granted", "--reason", "the written opinion"},
			typ:  "motion-rule",
			want: map[string]string{"motion_id": "M2", "subject": "petition",
				"ruling": "granted", "reason": "the written opinion"},
			says: "motion M2 ruled granted",
		},
		{
			name: "bench halt is the safety boundary",
			role: "bench", seatID: "judge-terminal",
			args: []string{"--reason", "the run must stop, and here is why"},
			typ:  "halt",
			want: map[string]string{"reason": "the run must stop, and here is why"},
			says: "JUDICIAL HALT recorded — capture relays this verbatim",
		},
		{
			name: "bench certify is the run-end statement",
			role: "bench", seatID: "assemble",
			args: []string{"--reason", "what I would want a human to re-examine"},
			typ:  "certify",
			want: map[string]string{"reason": "what I would want a human to re-examine"},
			says: "certification recorded",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runDir := t.TempDir()
			seedReferents(t, runDir)
			// THE VERB COMES FROM `typ`, NOT FROM THE TEST'S PROSE NAME.
			//
			// It was `strings.SplitN(tc.name, " ", 3)[1]` — the second word of a human-readable
			// case name. Renaming the `avenue` verb to `line-of-inquiry` renamed the case with
			// it, and the runner then invoked `blue line`, which does not exist. The failure
			// arrived as "no verb \"line\" exists on any seat", pointing at the tool rather than
			// at the harness that had composed a verb out of a sentence.
			//
			// Every case that relies on this default has verb == typ, so the event type IS the
			// verb name and there is nothing to recover: a case whose path differs states it in
			// `path`, as three already do.
			path := tc.path
			if path == nil {
				// The verb alone: the seat id selects the tree, so the role is no longer
				// part of what a seat types.
				path = []string{tc.typ}
			}
			args := append(append([]string{}, path...), "--run", runDir, "--seat-id", tc.seatID)
			args = append(args, tc.args...)
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
		out, err := run(t, "spot-check", "--run", runDir, "--seat-id", "red-merge-r1",
			"--ids", "R1-3", "--reason", "it still holds")
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
		if ev.Payload.Str("reason") != "it still holds" {
			t.Errorf("reason = %q — what the sample found lands in the one prose channel", ev.Payload.Str("reason"))
		}
	})

	t.Run("with no ids at all the key is still an empty array", func(t *testing.T) {
		runDir := t.TempDir()
		if _, err := run(t, "spot-check", "--run", runDir, "--seat-id", "red-merge-r1",
			"--reason", "the archive was empty at round start"); err != nil {
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
		if _, err := run(t, "spot-check", "--run", runDir, "--seat-id", "red-merge-r1", "--ids", ids,
			"--reason", "re-read the closure record"); err != nil {
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
	if _, err := run(t, "mint", "--run", runDir, "--seat-id", seatID,
		"--class", "x", "--check-kind", "document", "--check", "c", "--problem", "p",
		"--severity", "low", "--likelihood", "low", "--impact", "low"); err != nil {
		t.Fatal(err)
	}
	if _, err := run(t, "regrade", "--run", runDir, "--seat-id", seatID,
		"--id", "R1-1", "--severity", "certain", "--reason", "new evidence in §4"); err != nil {
		t.Fatal(err)
	}
	ev := lastOfType(t, runDir, "regrade")
	if got := ev.Payload.Str("severity"); got != "certain" {
		t.Errorf("severity = %q", got)
	}
	if got := ev.Payload.Str("reason"); got != "new evidence in §4" {
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
	// THE ROLE LEFT THE ARGS: the seat id selects the tree, so the invocation is verb-first.
	cases := []struct {
		verb, seatID, key string
		extra             []string
	}{
		{"halt", "judge-terminal", "reason", nil},
		{"certify", "assemble", "reason", nil},
		{"revision", "blue-lane-1", "reason", nil},
		{"closing", "red-merge-r1", "reason", []string{"--id", "R1-1"}},
		{"manifest-row", "blue-lane-1", "row", []string{"--id", "R1-1"}},
	}
	body := "a multi-line payload\nwith unicode — ✓ 日本語\nand <angle> & entities\n"
	for _, tc := range cases {
		t.Run(tc.seatID+"/"+tc.verb, func(t *testing.T) {
			runDir := t.TempDir()
			seedReferents(t, runDir)
			args := append([]string{tc.verb, "--run", runDir, "--seat-id", tc.seatID,
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
