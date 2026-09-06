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

// SUBJECT PROSE IS NOT A LEAK, and this is the assertion that keeps the list from becoming a
// censor. The advisory does not block and the lens argues, precisely because a pattern cannot tell
// a report narrating itself from a report quoting a source that narrates something.
func TestOrdinarySubjectProseIsClean(t *testing.T) {
	for _, s := range []string{
		"Loh's method is completing the square composed with x = -B/2 + z.",
		"The small root is recovered as C/r to avoid catastrophic cancellation.",
		"Savage 1989 is known only through the interested party's summary.",
	} {
		if got := Find(s); len(got) != 0 {
			t.Errorf("subject prose flagged as process voice: %q matched %v", s, got[0].Match)
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
