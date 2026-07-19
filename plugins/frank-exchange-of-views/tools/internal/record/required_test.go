package record

import "testing"

// THE TABLE MUST BE TRUE.
//
// RequiredFields is what the CLI marks REQUIRED in a seat's help, and help is the seat's
// whole contract. A table that claims a field is required when validate would accept it
// missing teaches the seat to pass something it does not need; a table that omits a real
// requirement leaves the seat to discover it by failing. Either way the contract lies.
//
// So this is behavioural, not structural: for every field the table declares, a payload
// missing exactly that field must actually be REFUSED by validate.

// runWithGap is a run directory in which R1-1 actually exists. Since references are now
// checked at write time, a fixture that names a gap must create it — which is the point.
func runWithGap(t *testing.T) string {
	t.Helper()
	runDir := t.TempDir()
	if _, _, err := RegisterSeat(runDir, "red-merge-r1"); err != nil {
		t.Fatal(err)
	}
	id, err := MintGapID(runDir, 1)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Append(runDir, "red-merge-r1", "mint", NewPayload().Set("gap_id", id).
		Set("acceptance_check", "c").Set("class", "x").
		Set("likelihood", "medium").Set("impact", "medium")); err != nil {
		t.Fatal(err)
	}
	return runDir
}

func TestEveryDeclaredRequiredFieldIsActuallyEnforced(t *testing.T) {
	// full is a payload that satisfies each verb, from which one field is removed at a
	// time. Values are placeholders — validate checks presence, not meaning.
	full := map[string]map[string]any{
		"mint":    {"acceptance_check": "c", "class": "scope-creep"},
		"close":   {"gap_id": "R1-1", "anchor_seat": "L1", "anchor_tool": "t", "anchor_target": "x"},
		"dispose": {"disposition": "declined", "reason": "r"},
		"regrade": {"basis": "b"},
		"retire":  {"claim": "c", "reason": "r"},
		"avenue":  {"status": "pursued", "line": "l"},
		"opinion": {"gap_id": "R1-1", "disposition": "carried", "principle": "p", "tension": "t", "review_flag": "no"},
	}

	for typ, required := range RequiredFields {
		base, ok := full[typ]
		if !ok {
			t.Errorf("%s is declared in RequiredFields but this test has no complete payload for it — the table grew and the check did not", typ)
			continue
		}
		for _, missing := range required {
			t.Run(typ+"/without_"+missing, func(t *testing.T) {
				p := NewPayload()
				for k, v := range base {
					if k != missing {
						p.Set(k, v)
					}
				}
				dir := t.TempDir()
				if typ == "opinion" {
					dir = runWithGap(t)
				}
				if err := validate(dir, typ, p); err == nil {
					t.Errorf("RequiredFields says %s.%s is required, but validate ACCEPTED a payload without it — the help would mark a flag REQUIRED that the tool does not require", typ, missing)
				}
			})
		}
	}
}

// And the complete payloads must be ACCEPTED, or the test above would pass for the wrong
// reason: a validate that rejects everything satisfies every "missing field is refused"
// case while telling us nothing.
func TestTheCompletePayloadsAreAccepted(t *testing.T) {
	for typ, p := range map[string]*Payload{
		"dispose": NewPayload().Set("disposition", "declined").Set("reason", "r"),
		"regrade": NewPayload().Set("basis", "b"),
		"retire":  NewPayload().Set("claim", "c").Set("reason", "r"),
		"avenue":  NewPayload().Set("status", "pursued").Set("line", "l"),
		"opinion": NewPayload().Set("gap_id", "R1-1").Set("disposition", "carried").
			Set("principle", "p").Set("tension", "t").Set("review_flag", "no"),
	} {
		dir := t.TempDir()
		if typ == "opinion" {
			dir = runWithGap(t)
		}
		if err := validate(dir, typ, p); err != nil {
			t.Errorf("%s rejected a payload carrying every required field: %v", typ, err)
		}
	}
}

// review_flag is required by PRESENCE, not by being non-empty, and that distinction is
// load-bearing: `--review-flag false` is a legitimate ruling ("no, a human need not look
// at this"), and a generic present-and-non-empty check would refuse it. Three separate
// defects in this codebase have come from treating a falsy value as an absent one.
func TestAFalsyReviewFlagSatisfiesTheRequirement(t *testing.T) {
	p := NewPayload().Set("gap_id", "R1-1").Set("disposition", "carried").
		Set("principle", "p").Set("tension", "t").Set("review_flag", false)
	if err := validate(runWithGap(t), "opinion", p); err != nil {
		t.Errorf("a legitimately falsy review_flag was treated as missing: %v", err)
	}
}

// A CARRY IS A LINEAGE CLAIM. --carried-from was presence-only, so any value satisfied it
// and satisfying it skipped the anchor — a fresh, unverified closure wearing a carry's
// clothes, counted as closed by every projection and by anchored_closures_pct. The help
// offers it exactly where a seat that cannot produce an anchor will read it.
func TestCarriedFromCannotLaunderAnUnanchoredFirstClosure(t *testing.T) {
	runDir := t.TempDir()
	if _, _, err := RegisterSeat(runDir, "red-merge-r1"); err != nil {
		t.Fatal(err)
	}
	id, err := MintGapID(runDir, 1)
	if err != nil {
		t.Fatal(err)
	}
	mint := NewPayload().Set("gap_id", id).Set("acceptance_check", "c").Set("class", "x").
		Set("likelihood", "medium").Set("impact", "medium")
	if _, err := Append(runDir, "red-merge-r1", "mint", mint); err != nil {
		t.Fatal(err)
	}

	// No prior closure exists, so a carry is a false claim about the record.
	p := NewPayload().Set("gap_id", id).Set("carried_from", "1")
	if err := validate(runDir, "close", p); err == nil {
		t.Error("an unanchored FIRST closure was accepted as a carry — that is the laundering path: no verification, no lineage, and it scores as closed")
	}

	// Anchored, it goes through — the escape hatch is closed, not the door.
	anchored := NewPayload().Set("gap_id", id).
		Set("anchor_seat", "L1").Set("anchor_tool", "go test").Set("anchor_target", "./x")
	if err := validate(runDir, "close", anchored); err != nil {
		t.Errorf("an anchored closure must still be accepted: %v", err)
	}
}

// And a GENUINE carry still works: close once with an anchor, then restate it.
func TestAGenuineCarryIsStillAccepted(t *testing.T) {
	runDir := t.TempDir()
	if _, _, err := RegisterSeat(runDir, "red-merge-r1"); err != nil {
		t.Fatal(err)
	}
	id, err := MintGapID(runDir, 1)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Append(runDir, "red-merge-r1", "mint", NewPayload().Set("gap_id", id).
		Set("acceptance_check", "c").Set("class", "x").
		Set("likelihood", "medium").Set("impact", "medium")); err != nil {
		t.Fatal(err)
	}
	if _, err := Append(runDir, "red-merge-r1", "close", NewPayload().Set("gap_id", id).
		Set("anchor_seat", "L1").Set("anchor_tool", "go test").Set("anchor_target", "./x")); err != nil {
		t.Fatal(err)
	}
	if err := validate(runDir, "close", NewPayload().Set("gap_id", id).Set("carried_from", "1")); err != nil {
		t.Errorf("a carry restating a real earlier closure must be accepted: %v", err)
	}
}

// likelihood and impact MULTIPLY into GapMass, and an absent grade contributes zero — so
// an ungraded gap reads as harmless rather than ungraded. Severity and cx are reported
// rather than multiplied, so their absence is visible and they stay optional.
func TestMintRequiresTheGradesThatMultiplyIntoMass(t *testing.T) {
	base := func() *Payload {
		return NewPayload().Set("acceptance_check", "c").Set("class", "scope-creep").
			Set("likelihood", "medium").Set("impact", "medium")
	}
	for _, missing := range []string{"likelihood", "impact"} {
		p := NewPayload()
		for _, k := range base().Keys() {
			if k != missing {
				v, _ := base().Get(k)
				p.Set(k, v)
			}
		}
		if err := validate(t.TempDir(), "mint", p); err == nil {
			t.Errorf("mint without --%s was accepted; its mass computes to ZERO and the gap sinks to the bottom of every ranking as though it were harmless", missing)
		}
	}
	// Severity and cx remain optional: absent, they are SHOWN absent.
	if err := validate(t.TempDir(), "mint", base()); err != nil {
		t.Errorf("severity and cx must stay optional — their absence is visible, not silently zero: %v", err)
	}
	if GapMass("", "medium") != 0 {
		t.Error("the premise of this rule has changed: an absent grade no longer contributes zero")
	}
}
