package cli

import (
	"fmt"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
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
	// THE REPORT IS PART OF THE WORLD THESE VERBS OPERATE IN. A supporting `lens corroborate`
	// splices a citation anchor at the claim, so the claim must be a real span of the live
	// document — the same rule blue's cite has always been held to. Without this the corroborate
	// case was refused for quoting a sentence that existed nowhere, which reads as the verb being
	// broken rather than as the fixture being a placeholder.
	seedBlueReport(t, runDir)
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
		"--id", "R1-3", "--as", "repaired", "--verified-by", "L1", "--verified-with", "go test",
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
		typ    recordpb.EventType
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
			// THE QUOTE IS A REAL SPAN of the seeded report. A supporting corroboration splices a
			// citation anchor at the claim, so the claim must be in the live document — the same
			// rule blue's cite is held to. `"the claim"` was a placeholder and is now refused.
			args: []string{"--quote", "the parser accepts an empty body in this line.",
				"--url", "https://example.test/a", "--title", "Example A",
				"--as", "supports", "--confidence", "high", "--reason", "read at the leaf",
				"--access-date", "2026-07-18"},
			typ: recordpb.EventType_EVENT_TYPE_VERIFY,
			// The flag is --access-date; the payload key is access_date, and the
			// citation render reads the payload key. --as lands under `outcome`.
			want: map[string]string{"claim": "the parser accepts an empty body in this line.", "url": "https://example.test/a", "title": "Example A",
				// `text`, not `reason`: --reason is the flag, and a verify stores its reading as
				// `text`. The old key named a field Verify does not carry, so the row asserted
				// nothing until fieldText started failing on an unknown name.
				"outcome": "supports", "confidence": "high", "text": "read at the leaf", "access_date": "2026-07-18"},
			says: "corroborating source Example A verified: supports",
		},
		{
			name: "lens corroborate without an access date leaves the key absent",
			path: []string{"corroborate"}, seatID: "red-lens-r1-L1",
			args: []string{"--quote", "c", "--url", "https://example.test/b", "--title", "Example B",
				"--as", "weak", "--confidence", "low", "--reason", "it gestures at it"},
			typ:    recordpb.EventType_EVENT_TYPE_VERIFY,
			want:   map[string]string{"url": "https://example.test/b", "title": "Example B", "outcome": "weak", "confidence": "low"},
			absent: []string{"access_date"},
			says:   "corroborating source Example B verified: weak",
		},
		{
			name: "motion grade rule records the merge's answer",
			path: []string{"motion", "grade", "rule"}, seatID: "red-merge-r1",
			args: []string{"--id", "M1", "--as", "rejected", "--reason", "the evidence does not reach it"},
			typ:  recordpb.EventType_EVENT_TYPE_MOTION_RULE,
			// THE RULING IS A ONEOF ARM, not a `ruling` field: `grade` carries GradeRuling,
			// `petition` PetitionRuling, `direction` DirectionRuling. That separation is what
			// makes a ruling from the wrong subject's vocabulary unrepresentable, and it is why
			// the old key could never match. The ruler's argument is `opinion`.
			want: map[string]string{"motion_id": "M1", "subject": "grade", "grade": "rejected",
				"opinion": "the evidence does not reach it"},
			says: "motion M1 ruled rejected",
		},
		{
			name: "motion grade file contests a grade through the accounted channel",
			path: []string{"motion", "grade", "file"}, seatID: "blue-lane-1",
			args: []string{"--id", "R1-1", "--dimension", "severity", "--proposed", "low", "--reason", "§4 says otherwise"},
			typ:  recordpb.EventType_EVENT_TYPE_MOTION,
			want: map[string]string{"gap_id": "R1-1", "dimension": "severity",
				"proposed": "low", "basis": "§4 says otherwise", "subject": "grade"},
			says: "filed (grade)",
		},
		{
			name: "blue manifest-row records the receipt",
			role: "blue", seatID: "blue-lane-1",
			args: []string{"--id", "R1-2", "--reason", "figures recomputed; acceptance check run: pass"},
			typ:  recordpb.EventType_EVENT_TYPE_MANIFEST_ROW,
			want: map[string]string{"gap_id": "R1-2", "row": "figures recomputed; acceptance check run: pass"},
			says: "manifest row recorded for R1-2",
		},
		{
			name: "blue retire records what left and why",
			role: "blue", seatID: "blue-lane-1",
			args: []string{"--quote", "the claim as it stood", "--reason", "refuted", "--new", "the replacement claim"},
			typ:  recordpb.EventType_EVENT_TYPE_RETIRE,
			want: map[string]string{"claim": "the claim as it stood", "reason": "refuted",
				"superseded_by": "the replacement claim"},
			says: "retired: the claim as it stood",
		},
		{
			name: "blue line of inquiry propose records the direction and its hypothesis",
			path: []string{"line-of-inquiry", "propose"}, seatID: "blue-lane-1",
			args: []string{"--reason", "search the offline archive", "--method", "full-text search",
				"--hypothesis", "the 1997 proceedings are scanned"},
			typ: recordpb.EventType_EVENT_TYPE_AVENUE,
			want: map[string]string{"line": "search the offline archive", "status": "proposed",
				"reason": "search the offline archive", "method": "full-text search",
				"hypothesis": "the 1997 proceedings are scanned"},
			says: "line of inquiry Q1 recorded (proposed): search the offline archive",
		},
		{
			name: "motion petition rule records the ruling and its opinion",
			path: []string{"motion", "petition", "rule"}, seatID: "judge-petition-red-merge-r1",
			args: []string{"--id", "M2", "--as", "granted", "--reason", "the written opinion"},
			typ:  recordpb.EventType_EVENT_TYPE_MOTION_RULE,
			want: map[string]string{"motion_id": "M2", "subject": "petition",
				// `opinion` is the ruler's argument — the one prose channel MotionRule carries.
				"petition": "granted", "opinion": "the written opinion"},
			says: "motion M2 ruled granted",
		},
		{
			name: "bench halt is the safety boundary",
			role: "bench", seatID: "judge-terminal",
			args: []string{"--reason", "the run must stop, and here is why"},
			typ:  recordpb.EventType_EVENT_TYPE_HALT,
			want: map[string]string{"opinion": "the run must stop, and here is why"},
			says: "JUDICIAL HALT recorded — capture relays this verbatim",
		},
		{
			name: "bench certify is the run-end statement",
			role: "bench", seatID: "assemble",
			args: []string{"--reason", "what I would want a human to re-examine"},
			typ:  recordpb.EventType_EVENT_TYPE_CERTIFY,
			want: map[string]string{"statement": "what I would want a human to re-examine"},
			says: "certification recorded",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runDir := newRun(t)
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
				// The verb word is the event type's word with the schema's underscores rendered as
				// the hyphens a COMMAND uses — `manifest_row` is typed `manifest-row`. A case
				// whose path differs beyond that states it in `path`, as three already do.
				// NO ROLE SEGMENT. The surface is scoped to the seat, so a verb sits at the ROOT
				// of its own tree; prefixing the role reproduces the old grouped path and is
				// answered with "no command named \"bench\" exists". `tc.role` still selects
				// which surface the case belongs to; it is not part of what a seat types.
				path = []string{strings.ReplaceAll(recordpb.Word(tc.typ), "_", "-")}
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
			body, ok := recordpb.Body(ev)
			if !ok {
				t.Fatalf("%s wrote an event with no body", tc.name)
			}
			for k, want := range tc.want {
				if got := fieldText(t, body, k); got != want {
					t.Errorf("%s = %q, want %q", k, got, want)
				}
			}
			keys := setFields(body)
			for _, k := range tc.absent {
				if keys[k] {
					t.Errorf("the body carries %q though the seat never passed it", k)
				}
			}
			if ev.GetSeatId() != tc.seatID {
				t.Errorf("seatId = %q, want %q", ev.GetSeatId(), tc.seatID)
			}
		})
	}
}

// spot-check's --ids is a CSV field, so it is ALWAYS present as an array — the
// same "lineage none, not lineage unknown" rule the mint lists follow.
func TestSpotCheckIdsAreAlwaysAnArray(t *testing.T) {
	t.Run("with ids", func(t *testing.T) {
		runDir := newRun(t)
		seedReferents(t, runDir)
		out, err := run(t, "spot-check", "--run", runDir, "--seat-id", "red-merge-r1",
			"--ids", "R1-3", "--reason", "it still holds")
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(out, "spot-checked R1-3") {
			t.Errorf("stdout = %q", out)
		}
		ev := lastBody(t, runDir, &recordpb.SpotCheck{})
		got := ev.GetIds()
		if len(got) != 1 || got[0] != "R1-3" {
			t.Errorf("ids = %q, want [R1-3] — the only CLOSED gap, which is what a spot-check samples", got)
		}
		if ev.GetReason() != "it still holds" {
			t.Errorf("reason = %q — what the sample found lands in the one prose channel", ev.GetReason())
		}
	})

	// THE THREE STATES, AND WHERE THE DISTINCTION LIVES NOW.
	//
	// This asserted that a spot-check with no ids still carried `ids` as a present-but-empty key,
	// because "an absent list reads as NOT CHECKED rather than CHECKED NOTHING". A payload map
	// could hold that; a proto message cannot — a repeated field has no presence, so an empty list
	// and an unset one are the same bytes, and the storage writes no rows for either.
	//
	// The distinction is real and it did not disappear; it is carried by two other facts the
	// record does hold. NOT CHECKED is the ABSENCE OF THE EVENT — a seat that never ran the verb
	// has no spot_check at all, and the sitting gate reads exactly that. CHECKED NOTHING is the
	// event with `none` set, which is a bool with presence and says so explicitly.
	//
	// So the states are asserted where they live rather than through a key that cannot carry them.
	t.Run("the three states are distinguishable", func(t *testing.T) {
		// 1. Never checked: no event.
		runDir := t.TempDir()
		if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_SPOT_CHECK); n != 0 {
			t.Fatalf("%d spot-checks before any were run", n)
		}
		// 2. Checked nothing, explicitly: --none, with the reason that distinguishes it from a
		// skipped duty.
		if _, err := run(t, "spot-check", "--run", runDir, "--seat-id", "red-merge-r1",
			"--none", "--reason", "the archive was empty at round start"); err != nil {
			t.Fatal(err)
		}
		ev := lastBody(t, runDir, &recordpb.SpotCheck{})
		if !ev.GetNone() {
			t.Error("the explicit empty form did not set `none` — without it, checked-nothing and " +
				"sampled-nothing-in-particular are the same event")
		}
		if ev.GetReason() == "" {
			t.Error("the explicit empty form lost its reason, which is the only thing distinguishing " +
				"it from a duty nobody discharged")
		}
		if len(ev.GetIds()) != 0 {
			t.Errorf("the empty form sampled %v", ev.GetIds())
		}
	})
}

// spot-check is a singleton per seat: the round's duty is discharged ONCE, and a second is
// REFUSED rather than silently replacing the first.
//
// The shard record deduped on read — two events, one key, one discarded — so a seat that ran the
// verb twice learned nothing about which sample stood. `events.key` is UNIQUE now, so the second
// write fails and the seat is told why.
func TestSpotCheckIsASingleton(t *testing.T) {
	runDir := newRun(t)
	seedReferents(t, runDir)
	if _, err := run(t, "spot-check", "--run", runDir, "--seat-id", "red-merge-r1", "--ids", "R1-3",
		"--reason", "re-read the closure record"); err != nil {
		t.Fatal(err)
	}
	_, err := run(t, "spot-check", "--run", runDir, "--seat-id", "red-merge-r1", "--ids", "R1-3",
		"--reason", "re-read it again")
	if err == nil {
		t.Fatal("a second spot-check was accepted — the round's duty would have two discharges")
	}
	if !strings.Contains(err.Error(), "once-per-sitting") {
		t.Errorf("the refusal does not teach what was wrong:\n%v", err)
	}
	n := 0
	for _, e := range events(t, runDir) {
		if e.GetType() == recordpb.EventType_EVENT_TYPE_SPOT_CHECK {
			n++
		}
	}
	if n != 1 {
		t.Errorf("%d spot-check events survived, want 1", n)
	}
}

// regrade moves only the grades it carries, and refuses without its basis.
func TestRegradeMovesOnlyThePassedGrades(t *testing.T) {
	runDir := newRun(t)
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
	ev := lastBody(t, runDir, &recordpb.Regrade{})
	if got := ev.GetSeverity(); got != recordpb.Grade_GRADE_CERTAIN {
		t.Errorf("severity = %q", got)
	}
	if got := ev.GetBasis(); got != "new evidence in §4" {
		t.Errorf("basis = %q", got)
	}
	// The grades NOT passed must be absent, so the replay leaves them alone.
	keys := setFields(ev)
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
	if g.Severity != recordpb.Grade_GRADE_CERTAIN {
		t.Errorf("board severity = %v, want certain", g.Severity)
	}
	if g.Likelihood != recordpb.Grade_GRADE_LOW || g.Impact != recordpb.Grade_GRADE_LOW {
		t.Errorf("a grade the regrade did not carry moved: likelihood=%v impact=%v", g.Likelihood, g.Impact)
	}
}

// The prose channel is available on the verbs that declare it, and --file is the
// documented path for anything above trivial size.
func TestProseVerbsAcceptAFile(t *testing.T) {
	// THE FIELD, NOT THE FLAG. Every row said `reason` — the word a seat types — while the schema
	// spells the prose field per verb: a halt stores `opinion`, a certification `statement`, a
	// revision and a closing `text`. Read against a payload map the wrong name returned "" and the
	// assertion compared "" to the file's content, so it failed honestly; read against the
	// descriptor it names the field the record actually holds.
	cases := []struct {
		role, verb, seatID, field string
		typ                       recordpb.EventType
		extra                     []string
	}{
		{"bench", "halt", "judge-terminal", "opinion", recordpb.EventType_EVENT_TYPE_HALT, nil},
		{"bench", "certify", "assemble", "statement", recordpb.EventType_EVENT_TYPE_CERTIFY, nil},
		{"blue", "revision", "blue-lane-1", "text", recordpb.EventType_EVENT_TYPE_REVISION, nil},
		{"merge", "closing", "red-merge-r1", "text", recordpb.EventType_EVENT_TYPE_CLOSING, []string{"--id", "R1-1"}},
		{"blue", "manifest-row", "blue-lane-1", "row", recordpb.EventType_EVENT_TYPE_MANIFEST_ROW, []string{"--id", "R1-1"}},
	}
	body := "a multi-line payload\nwith unicode — ✓ 日本語\nand <angle> & entities\n"
	for _, tc := range cases {
		t.Run(tc.seatID+"/"+tc.verb, func(t *testing.T) {
			runDir := newRun(t)
			seedReferents(t, runDir)
			args := append([]string{tc.verb, "--run", runDir, "--seat-id", tc.seatID,
				"--reason-file", writeTemp(t, body)}, tc.extra...)
			if _, err := run(t, args...); err != nil {
				t.Fatal(err)
			}
			// BY TYPE, not "the last event in the log" — the fixture seeds referents, so the
			// tail of the slice belongs to whichever act happened last, not to this verb.
			last := lastOfType(t, runDir, tc.typ)
			lb, ok := recordpb.Body(last)
			if !ok {
				t.Fatalf("%s/%s wrote an event with no body", tc.role, tc.verb)
			}
			// Less the file's terminating newline: that is a line terminator every editor
			// appends, not content the seat chose to record.
			if got := fieldText(t, lb, tc.field); got != strings.TrimRight(body, "\n") {
				t.Errorf("%s = %q, want the file's content without its terminator", tc.field, got)
			}
		})
	}
}
