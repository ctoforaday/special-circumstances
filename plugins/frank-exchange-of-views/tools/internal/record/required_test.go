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
// seatFor picks a seat of the role that owns each verb, since validate now resolves
// round-scoped references and needs to know who is writing.
func seatFor(typ string) string {
	switch typ {
	case "opinion", "halt", "certify":
		return "judge-r1"
	case "retire", "line-of-inquiry", "manifest-row", "revision", "confidence":
		return "blue-respond-r1"
	default:
		return "red-merge-r1"
	}
}

func runWithGap(t *testing.T) string {
	t.Helper()
	runDir := t.TempDir()
	if _, _, err := RegisterSeat(Identity{RunDir: runDir, SeatID: "red-merge-r1", Round: RoundIn(runDir)("red-merge-r1")}); err != nil {
		t.Fatal(err)
	}
	id, err := MintGapID(runDir, 1)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Append(Identity{RunDir: runDir, SeatID: "red-merge-r1", Round: RoundIn(runDir)("red-merge-r1")}, "mint", NewPayload().Set("gap_id", id).
		Set("acceptance_check", "c").Set("check_kind", "document").Set("class", "x").
		Set("likelihood", "medium").Set("impact", "medium").Set("problem", "p")); err != nil {
		t.Fatal(err)
	}
	return runDir
}

func TestEveryDeclaredRequiredFieldIsActuallyEnforced(t *testing.T) {
	// full is a payload that satisfies each verb, from which one field is removed at a
	// time. Values are placeholders — validate checks presence, not meaning.
	full := map[string]map[string]any{
		"mint":            {"acceptance_check": "c", "class": "scope-creep"},
		"close":           {"gap_id": "R1-1", "anchor_seat": "L1", "anchor_tool": "t", "anchor_target": "x", "reason": "p"},
		"closing":         {"gap_id": "R1-1", "reason": "t"},
		"regrade":         {"reason": "b"},
		"retire":          {"claim": "c", "reason": "r"},
		"line-of-inquiry": {"status": "pursued", "line": "l"},
		"inquiry-support": {"inquiry_id": "Q1", "as": "supported", "reason": "r"},
		"opinion":         {"gap_id": "R1-1", "disposition": "carried", "principle": "p", "tension": "t", "review_flag": "no", "reason": "r"},
		"halt":            {"reason": "o"},
		"certify":         {"reason": "s"},
		"outcome":         {"verdict": "VERIFIED", "reason": "p"},
		"finding":         {"label": "L1-F1"},
		"observe":         {"label": "L1-O1"},
		// The verb that took no required flag at all: a bare `lens verify` recorded an event and
		// counted as red's audit volume. `outcome` is the payload key behind --as.
		"verify": {"claim": "c", "outcome": "supports", "confidence": "high", "reason": "what the source says"},
		// The duties a bare verb used to discharge with nothing. Two of these GATE the sitting,
		// so an empty one did not merely record badly — it ended the seat's obligations.
		"friction":      {"reason": "what I reached for and could not get"},
		"friction-none": {"reason": "what I reached for and found"},
		"position":      {"reason": "where I stand this round"},
		"revision":      {"reason": "what I changed this round"},
		"manifest-row":  {"gap_id": "R1-1", "row": "what I checked and what it showed"},
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
				if err := validate(dir, seatFor(typ), typ, p); err == nil {
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
		"regrade": NewPayload().Set("reason", "b"),
		"retire":  NewPayload().Set("claim", "c").Set("reason", "r"),
		// inquiry_id is TOOL-assigned, like a finding's label and a mint's gap_id: validate
		// requires it and no flag sets it, so it is not in RequiredFields but must be present
		// for a complete payload.
		"line-of-inquiry": NewPayload().Set("inquiry_id", "Q1").Set("status", "pursued").Set("line", "l"),
		"opinion": NewPayload().Set("gap_id", "R1-1").Set("disposition", "carried").
			Set("principle", "p").Set("tension", "t").Set("review_flag", "no").Set("reason", "r"),
	} {
		dir := t.TempDir()
		if typ == "opinion" {
			dir = runWithGap(t)
		}
		if err := validate(dir, seatFor(typ), typ, p); err != nil {
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
		Set("principle", "p").Set("tension", "t").Set("review_flag", false).Set("reason", "r")
	if err := validate(runWithGap(t), "judge-r1", "opinion", p); err != nil {
		t.Errorf("a legitimately falsy review_flag was treated as missing: %v", err)
	}
}

// A CARRY IS A LINEAGE CLAIM. --carried-from was presence-only, so any value satisfied it
// and satisfying it skipped the anchor — a fresh, unverified closure wearing a carry's
// clothes, counted as closed by every projection and by anchored_closures_pct. The help
// offers it exactly where a seat that cannot produce an anchor will read it.
func TestCarriedFromCannotLaunderAnUnanchoredFirstClosure(t *testing.T) {
	runDir := t.TempDir()
	if _, _, err := RegisterSeat(Identity{RunDir: runDir, SeatID: "red-merge-r1", Round: RoundIn(runDir)("red-merge-r1")}); err != nil {
		t.Fatal(err)
	}
	id, err := MintGapID(runDir, 1)
	if err != nil {
		t.Fatal(err)
	}
	mint := NewPayload().Set("gap_id", id).Set("acceptance_check", "c").Set("check_kind", "document").Set("class", "x").
		Set("likelihood", "medium").Set("impact", "medium").Set("problem", "p")
	if _, err := Append(Identity{RunDir: runDir, SeatID: "red-merge-r1", Round: RoundIn(runDir)("red-merge-r1")}, "mint", mint); err != nil {
		t.Fatal(err)
	}

	// No prior closure exists, so a carry is a false claim about the record.
	p := NewPayload().Set("gap_id", id).Set("carried_from", "1")
	if err := validate(runDir, "red-merge-r1", "close", p); err == nil {
		t.Error("an unanchored FIRST closure was accepted as a carry — that is the laundering path: no verification, no lineage, and it scores as closed")
	}

	// Anchored, it goes through — the escape hatch is closed, not the door.
	anchored := NewPayload().Set("gap_id", id).
		Set("anchor_seat", "L1").Set("anchor_tool", "go test").Set("anchor_target", "./x").
		Set("reason", "verified and holds")
	if err := validate(runDir, "red-merge-r1", "close", anchored); err != nil {
		t.Errorf("an anchored closure must still be accepted: %v", err)
	}
}

// And a GENUINE carry still works: close once with an anchor, then restate it.
func TestAGenuineCarryIsStillAccepted(t *testing.T) {
	runDir := t.TempDir()
	if _, _, err := RegisterSeat(Identity{RunDir: runDir, SeatID: "red-merge-r1", Round: RoundIn(runDir)("red-merge-r1")}); err != nil {
		t.Fatal(err)
	}
	id, err := MintGapID(runDir, 1)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Append(Identity{RunDir: runDir, SeatID: "red-merge-r1", Round: RoundIn(runDir)("red-merge-r1")}, "mint", NewPayload().Set("gap_id", id).
		Set("acceptance_check", "c").Set("check_kind", "document").Set("class", "x").
		Set("likelihood", "medium").Set("impact", "medium").Set("problem", "p")); err != nil {
		t.Fatal(err)
	}
	if _, err := Append(Identity{RunDir: runDir, SeatID: "red-merge-r1", Round: RoundIn(runDir)("red-merge-r1")}, "close", NewPayload().Set("gap_id", id).
		Set("anchor_seat", "L1").Set("anchor_tool", "go test").Set("anchor_target", "./x").
		Set("reason", "verified and holds")); err != nil {
		t.Fatal(err)
	}
	if err := validate(runDir, "red-merge-r1", "close", NewPayload().Set("gap_id", id).Set("carried_from", "1")); err != nil {
		t.Errorf("a carry restating a real earlier closure must be accepted: %v", err)
	}
}

// likelihood and impact MULTIPLY into GapMass, and an absent grade contributes zero — so
// an ungraded gap reads as harmless rather than ungraded. Severity and cx are reported
// rather than multiplied, so their absence is visible and they stay optional.
func TestMintRequiresTheGradesThatMultiplyIntoMass(t *testing.T) {
	base := func() *Payload {
		return NewPayload().Set("acceptance_check", "c").Set("check_kind", "document").Set("class", "scope-creep").
			Set("likelihood", "medium").Set("impact", "medium").Set("problem", "p")
	}
	for _, missing := range []string{"likelihood", "impact"} {
		p := NewPayload()
		for _, k := range base().Keys() {
			if k != missing {
				v, _ := base().Get(k)
				p.Set(k, v)
			}
		}
		if err := validate(t.TempDir(), "red-merge-r1", "mint", p); err == nil {
			t.Errorf("mint without --%s was accepted; its mass computes to ZERO and the gap sinks to the bottom of every ranking as though it were harmless", missing)
		}
	}
	// Severity and cx remain optional: absent, they are SHOWN absent.
	if err := validate(t.TempDir(), "red-merge-r1", "mint", base()); err != nil {
		t.Errorf("severity and cx must stay optional — their absence is visible, not silently zero: %v", err)
	}
	if GapMass("", "medium") != 0 {
		t.Error("the premise of this rule has changed: an absent grade no longer contributes zero")
	}
}
