package record

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode"
)

// THE REAL-DATA HALF OF THE #344 COMPAT CHECK, and it exists because the other half is a FIXTURE.
//
// `pre-motion-run` is produced by TestFuzzDebate: mechanically complete, and mechanically shaped.
// The plan's §V is explicit that the two artifacts have different jobs — fixtures prove logic,
// while only real data surfaces data-shaped defects, and a harness-produced record is precisely
// the category that cannot surface a harness sentinel. The sensitivity is live in this repo by
// its own validation loop: the node suite is the ONLY check that catches a backtick nested in a
// template literal, and `node --check` passes on it.
//
// So this record was written by real model-driven seats, through the real pre-#344 binary
// (0.47.0, which has no motion verbs at all), and it carries prose the fuzz does not generate:
// backticks, angle brackets, pipes, bracketed citations, em dashes and other non-ASCII, and a
// single opinion field of 13,445 characters. Five seats across two rounds — blue contesting three
// grades, the merge ruling them and six proposed directions, the bench hearing a petition and
// granting it in part.
//
// It CANNOT be regenerated once the retiring verbs are gone. Recover it from git history.
//
// ONE THING IT DOES NOT CARRY, said here so nobody assumes otherwise: `contests_ruling` appears
// ZERO times in it. The merge ruled a line too-thin, blue argued against that reasoning at the
// leaf, and then declined the line anyway — and the field cannot represent disagreeing and
// yielding, because it is a side effect of moving a line to `pursued` rather than a channel of its
// own. That absence is a finding (see #344), not a hole in the artifact; the fuzz fixture carries
// the field, and premotion_test.go asserts on it there.
const preMotionRealRun = "testdata/pre-motion-real-run"

// The five types the collapse retires, with what each one IS — named individually so a failure
// says WHICH exchange stopped being visible rather than "something is missing".
var retiringTypes = map[string]string{
	"dispute":         "blue contesting a grade",
	"dispute-respond": "the merge answering that contest",
	"petition":        "a seat's ethical/safety/integrity filing",
	"petition-rule":   "the bench's ruling on it",
	"avenue-rule":     "the merge's ruling on a proposed direction",
}

// THE ARTIFACT MUST KEEP CARRYING WHAT IT WAS MADE FOR.
//
// Runs first and independently: a truncated or replaced artifact fails HERE with a reason, rather
// than as a confusing pass somewhere downstream on a record that no longer exercises anything.
func TestThePreMotionRealRunStillCarriesTheRetiringVocabulary(t *testing.T) {
	entries, err := os.ReadDir(filepath.Join(preMotionRealRun, "records"))
	if err != nil {
		t.Fatalf("the real pre-motion record is missing: %v — it CANNOT be reproduced after #344 (no binary will still write these types); recover it from git history", err)
	}
	var joined strings.Builder
	for _, e := range entries {
		b, rerr := os.ReadFile(filepath.Join(preMotionRealRun, "records", e.Name()))
		if rerr != nil {
			t.Fatal(rerr)
		}
		joined.Write(b)
	}
	text := joined.String()
	for typ, what := range retiringTypes {
		if !strings.Contains(text, `"type":"`+typ+`"`) {
			t.Errorf("the real record no longer carries %q (%s) — without it this artifact proves nothing about that exchange", typ, what)
		}
	}
}

// A REAL PRE-COLLAPSE RECORD STILL REPLAYS, and every exchange is still visible.
func TestARealPreCollapseRecordStillReplays(t *testing.T) {
	b, err := BoardState(preMotionRealRun)
	if err != nil {
		t.Fatalf("a real pre-collapse record must still replay: %v", err)
	}
	if len(b.Events) == 0 {
		t.Fatal("the real record replayed to ZERO events — a plausible zero, which is the exact failure this guards")
	}
	counts := map[string]int{}
	for _, e := range b.Events {
		counts[e.Type]++
	}
	for typ, what := range retiringTypes {
		if counts[typ] == 0 {
			t.Errorf("%s (%s) is not visible in the replayed record — a reader sees an empty section, which reads identically to a run that never had one", typ, what)
		}
	}
}

// THE DUAL-READ, ON REAL SEAT PROSE.
//
// This asserts SHAPE, not verdicts. The fixture test can assert `ruling == "rejected"` because a
// harness wrote it; here five real seats decided what to argue, and pinning their conclusions
// would make this test a transcript of one afternoon's opinions rather than a compat check. What
// must hold is that every ask came back WITH its answer and WITH the prose the seat wrote.
func TestTheDualReadRecoversTheRealExchanges(t *testing.T) {
	b, err := BoardState(preMotionRealRun)
	if err != nil {
		t.Fatal(err)
	}
	ms := AllMotions(b)
	if len(ms) == 0 {
		t.Fatal("the dual-read returned NO motions for a record carrying all five retiring types — the plausible zero this guards against")
	}

	bySubject := map[string][]*Motion{}
	for _, m := range ms {
		bySubject[m.Subject] = append(bySubject[m.Subject], m)
	}
	// The record's actual shape: 3 grade disputes, 1 petition, 6 direction rulings.
	for subject, want := range map[string]int{"grade": 3, "petition": 1, "direction": 6} {
		if got := len(bySubject[subject]); got != want {
			t.Errorf("%s: recovered %d motion(s), want %d — a join that drops or duplicates an exchange is the #312 defect the id exists to close", subject, got, want)
		}
	}

	for _, m := range ms {
		if m.Basis == "" && m.Subject != "direction" {
			// A direction's ask is the avenue's `line`, recovered from the proposal; the other
			// two carry the filer's own words on the filing event.
			t.Errorf("motion %s (%s) came back with no ASK — the filer's argument is the substance, and a row without it is a verdict nobody can contest", m.ID, m.Subject)
		}
		if !m.Ruled() {
			t.Errorf("motion %s (%s) came back UNRULED, but every exchange in this record was answered — a ruling lost at replay renders as an ask nobody addressed", m.ID, m.Subject)
			continue
		}
		if m.Opinion == "" {
			t.Errorf("motion %s (%s) kept its ruling and lost its REASON — an unreasoned ruling is the decoration the filer cannot contest", m.ID, m.Subject)
		}
		if m.Fields["legacy"] != "true" {
			t.Errorf("motion %s came from a pre-collapse record and is not marked legacy; a synthesised id must never read as one the record carried", m.ID)
		}
	}
}

// THE PROSE SURVIVES THE ROUND TRIP BYTE FOR BYTE, and this is the assertion only real data can
// make.
//
// A dual-read that JSON-decodes, replays, joins and re-renders has several places to mangle text,
// and every one of them is invisible against harness prose — `fuzz: ruling on the contested grade`
// survives anything. Real seat prose does not: it carries backticks, angle brackets, pipes,
// bracketed references, em dashes and other non-ASCII, and one opinion here is 13,445 characters.
//
// The test finds the longest field the record actually holds rather than naming a sentence: a
// quoted excerpt would rot the moment the artifact is regenerated, and this artifact is meant to
// outlive everyone who remembers what was in it.
func TestRealSeatProseSurvivesTheDualRead(t *testing.T) {
	b, err := BoardState(preMotionRealRun)
	if err != nil {
		t.Fatal(err)
	}
	ms := AllMotions(b)

	var longest string
	var nonASCII, backticks, angles bool
	for _, m := range ms {
		for _, s := range []string{m.Basis, m.Opinion, m.Relief, m.AppealReason} {
			if len(s) > len(longest) {
				longest = s
			}
			backticks = backticks || strings.Contains(s, "`")
			angles = angles || strings.ContainsAny(s, "<>")
			for _, r := range s {
				if r > unicode.MaxASCII {
					nonASCII = true
				}
			}
		}
	}

	// 4000 is well under the 13,445-character opinion this record holds and well over anything a
	// truncation would leave. A ceiling somewhere in the replay path is the failure mode: it
	// silently keeps the first N characters and the row still renders.
	if len(longest) < 4000 {
		t.Errorf("the longest prose field recovered is %d characters; this record holds one of 13,445. Something in the replay path is truncating, and a truncated ruling still renders as a ruling", len(longest))
	}
	if !backticks {
		t.Error("no backtick survived the round trip — the one character class this repo's own validation loop says is caught by exactly one check")
	}
	if !angles {
		t.Error("no angle bracket survived the round trip")
	}
	if !nonASCII {
		t.Error("no non-ASCII survived the round trip — em dashes and typographic quotes are what real seats write and what encoding bugs eat")
	}
}
