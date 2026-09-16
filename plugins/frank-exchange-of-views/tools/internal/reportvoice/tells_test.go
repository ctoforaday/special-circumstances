package reportvoice

import "testing"

// THE LIST IS PINNED AGAINST THE MEASUREMENT THAT PRODUCED IT. Each class here was counted on the
// 2026-09-02 quadratic-formula report; a class removed without a new measurement is a class that
// stopped being looked for, which reads exactly like a class that stopped happening.
func TestEveryMeasuredClassIsStillLookedFor(t *testing.T) {
	want := map[Class]string{
		ProcessVoice:    "this run reached the round-5 ceiling", // 161 occurrences
		LaneAttribution: "the count is 27 [minority: lane-2/practitioner]",
		DraftHistory:    "an earlier version of this sentence omitted it",
		Apparatus:       "all six cells are printed by the checking program",
	}
	for class, sample := range want {
		found := Find(sample)
		if len(found) == 0 {
			t.Errorf("%s is no longer looked for; the sample that produced it now reads clean: %q", class, sample)
			continue
		}
		if found[0].Class != class {
			t.Errorf("%q matched %s, want %s", sample, found[0].Class, class)
		}
		if found[0].Redirect == "" {
			t.Errorf("%s says where it leaked from and not where it belongs — a tell without a destination is a complaint", class)
		}
	}
}

// EVERY SPELLING THE 2026-09-11 BASE USED IS LOOKED FOR. Five of its run-narration sentences were
// "this run"; the others said "this debate", "this sitting" or "more than one research lane", and
// the advisory named only the five. Each sample must match its own class.
func TestTheRunNarrationSpellingsAreLookedFor(t *testing.T) {
	for sample, class := range map[string]Class{
		"what this debate adds is convergent verification":        ProcessVoice,
		"could not from this run's position survey other sources": ProcessVoice,
		"not pursued this sitting":                                ProcessVoice,
		"the debate settled it":                                   ProcessVoice,
		"computed independently by more than one research lane":   LaneAttribution,
		"two research lanes agree on the count":                   LaneAttribution,
	} {
		found := Find(sample)
		if len(found) == 0 {
			t.Errorf("%q reads clean; it narrates the run", sample)
			continue
		}
		if found[0].Class != class {
			t.Errorf("%q matched %s, want %s", sample, found[0].Class, class)
		}
	}
}

// SUBJECT PROSE IS NOT A LEAK, and this is the assertion that keeps the list from becoming a
// censor. The advisory does not block and the lens argues, precisely because a pattern cannot tell
// a report narrating itself from a report quoting a source that narrates something.
//
// The id-shaped and role-shaped samples hold the REFUSED set to its bar: a G-number with no process
// word beside it, a committee's chair, a judge who is sitting and a gap measured in metres are all
// subject prose, and a refusal on any of them would cost a lens a true sentence. "the red team
// exercise" is subject prose in a security report and process voice in this one, so it is ADVICE:
// matched, and never refused.
func TestOrdinarySubjectProseIsClean(t *testing.T) {
	for _, s := range []string{
		"Loh's method is completing the square composed with x = -B/2 + z.",
		"The small root is recovered as C/r to avoid catastrophic cancellation.",
		"Savage 1989 is known only through the interested party's summary.",
		"the G20 summit",
		"the chair of the committee",
		"a sitting judge",
		"a gap 2 metres wide",
	} {
		if got := Find(s); len(got) != 0 {
			t.Errorf("subject prose flagged as process voice: %q matched %v", s, got[0].Match)
		}
	}
	const ambiguous = "the red team exercise"
	if got := Refused(ambiguous); len(got) != 0 {
		t.Errorf("%q is refused on %q; a red team is a subject a report can have", ambiguous, got[0].Match)
	}
	if got := Advised(ambiguous); len(got) == 0 {
		t.Errorf("%q is not advised; in this report's voice it names a party of the run", ambiguous)
	}
}

// THE REFUSED TELLS ARE THE UNAMBIGUOUS ONES (gblock, 2026-09-11): seat and lens ids, finding labels,
// gap ids joined to a process word, and lane tags are refused; every tell with a reading as subject
// prose is advised. Each refused sample must be refused, each ambiguous sample matched and NOT
// refused, and every refused row in the list must be the one some sample reaches — so deleting a row,
// or flipping its Refuse, fails here.
func TestRefusedTellsAreUnambiguous(t *testing.T) {
	refused := []string{
		"blue-respond overstates the bound",               // seat id
		"red-lens-evidence could not open it",             // lens id
		"the red-chair ruled it",                          // chair id
		"judge-terminal certified it",                     // bench id
		"blue-synthesize merged two drafts",               // synthesis seat id
		"blue-lane-2 found the count",                     // lane seat id
		"as evidence-F3 notes, the date is wrong",         // finding label
		"gap G3 is still open",                            // gap id + process word
		"Findings G2 and G4 agree",                        // gap id + process word, capitalised
		"G4's fix did not land",                           // gap id possessive + process word
		"closed G2 without a check",                       // process verb + gap id
		"superseded as G7",                                // process verb + as + gap id
		"the count is 27 [minority: lane-2/practitioner]", // lane tag
		"a single witness [lane-1]",                       // lane tag
	}
	ambiguous := []string{
		"this run reached the ceiling",
		"the debate settled it",
		"the red team missed it",
		"blue side conceded",
		"raised in epoch 3",
		"not pursued in sitting 2",
		"two research lanes agree",
		"corrected here",
		"the checking program prints it",
	}
	reached := map[*Tell]bool{}
	for _, s := range refused {
		got := Refused(s)
		if len(got) == 0 {
			t.Errorf("%q is not refused; it has no reading as subject prose", s)
		}
		for _, f := range got {
			for i := range tells {
				if tells[i].Pattern == f.Pattern {
					reached[&tells[i]] = true
				}
			}
		}
	}
	for _, s := range ambiguous {
		if got := Refused(s); len(got) != 0 {
			t.Errorf("%q is refused on %q; it is ambiguous and must stay advice", s, got[0].Match)
		}
		if got := Advised(s); len(got) == 0 {
			t.Errorf("%q is not advised; the tell stopped being looked for", s)
		}
	}
	for i := range tells {
		if tells[i].Refuse && !reached[&tells[i]] {
			t.Errorf("refused tell %s is reached by no refused sample — a row no test would miss", tells[i].Pattern)
		}
	}
}

// ONE SOURCE, TWO READERS. Tells() must hand back a copy: a caller that mutated the shared slice
// would change what the OTHER reader sees, which is the drift having one list exists to prevent.
func TestTellsCannotBeMutatedByACaller(t *testing.T) {
	a := Tells()
	if len(a) == 0 {
		t.Fatal("no tells")
	}
	a[0] = Tell{}
	if Tells()[0].Pattern == nil {
		t.Error("a caller mutated the shared list — the two readers would drift")
	}
}
