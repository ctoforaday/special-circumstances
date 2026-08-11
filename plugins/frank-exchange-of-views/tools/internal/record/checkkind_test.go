package record

import "testing"

// check_kind IS THE FIELD WITH TEETH, AND IT REACHED NO VIEW.
//
// `computation` means the gap cannot be closed on prose at all — only a `blue prove --answers`
// settles it, and the merge is refused if it tries anything else. That gate works. But it fires
// at the MERGE, at close time, in the following round, and blue — the only seat that can satisfy
// it — could not see which gaps carried it.
//
// Measured: `prove` was invoked zero times across eighteen probed seat dispatches on boards built
// to demand it. One seat summed twelve integers inside its own reasoning, wrote the answer, and
// was satisfied. The answer was right, which is why nothing caught it.
func TestCheckKindReachesTheSeatThatMustSatisfyIt(t *testing.T) {
	runDir := t.TempDir()
	if _, _, err := RegisterSeat(runDir, "red-merge-r1"); err != nil {
		t.Fatal(err)
	}
	for id, kind := range map[string]string{"R1-1": "computation", "R1-2": "document"} {
		if _, err := Append(runDir, "red-merge-r1", "mint", NewPayload().
			Set("gap_id", id).Set("class", "self-attestation").
			Set("location", "L").Set("problem", "p").Set("required_fix", "f").
			Set("acceptance_check", "c").Set("check_kind", kind).
			Set("severity", "medium").Set("likelihood", "medium").Set("impact", "medium").
			Set("complexity_cost", "low").Set("existence", "verified")); err != nil {
			t.Fatal(err)
		}
	}
	b, err := BoardState(runDir)
	if err != nil {
		t.Fatal(err)
	}

	got := map[string]string{}
	for _, g := range BoardJSONOf(b).Open {
		got[g.ID] = g.CheckKind
	}
	if got["R1-1"] != "computation" || got["R1-2"] != "document" {
		t.Errorf("board view check_kind = %v — a seat cannot know which gaps demand a program", got)
	}

	// The worklist is the SCANNING read a seat plans its sitting from, so the demand's TYPE has
	// to be on it even though the rest of the acceptance check deliberately is not.
	got = map[string]string{}
	for _, g := range WorklistJSONOf(b).Open {
		got[g.ID] = g.CheckKind
	}
	if got["R1-1"] != "computation" || got["R1-2"] != "document" {
		t.Errorf("worklist check_kind = %v — the read a seat plans from cannot say which gaps prose will not close", got)
	}
}

// AN EMPTY FRICTION LOG IS TWO DIFFERENT RUNS, and only one of them is fine.
func TestTheFrictionViewSeparatesSilenceFromAnAttestation(t *testing.T) {
	runDir := t.TempDir()
	if _, _, err := RegisterSeat(runDir, "blue-respond-r1"); err != nil {
		t.Fatal(err)
	}
	b, err := BoardState(runDir)
	if err != nil {
		t.Fatal(err)
	}
	j := FrictionJSONOf(b)
	if j.Counts.Total != 0 || j.Counts.Attested != 0 {
		t.Fatalf("a silent run: total=%d attested=%d, want 0/0", j.Counts.Total, j.Counts.Attested)
	}

	if _, err := Append(runDir, "blue-respond-r1", "friction-none",
		NewPayload().Set("text", "read the board and my verb list; every refusal was my own error")); err != nil {
		t.Fatal(err)
	}
	b, _ = BoardState(runDir)
	j = FrictionJSONOf(b)
	// The counts must now DIFFER from the silent run. Same total, different meaning — which is
	// the whole point: zero-with-an-attestation is a statement someone can be wrong about,
	// zero-alone is the absence of one.
	if j.Counts.Total != 0 {
		t.Errorf("an attestation must not count as a complaint: total=%d", j.Counts.Total)
	}
	if j.Counts.Attested != 1 || len(j.NothingBlocked) != 1 {
		t.Fatalf("the attestation did not reach the view: attested=%d entries=%d", j.Counts.Attested, len(j.NothingBlocked))
	}
	if j.NothingBlocked[0].SeatID != "blue-respond-r1" {
		t.Error("the attestation must name the seat that made it — an unattributed one cannot be weighed")
	}
}
