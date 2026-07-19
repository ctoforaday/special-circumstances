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
				if err := validate(t.TempDir(), typ, p); err == nil {
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
		if err := validate(t.TempDir(), typ, p); err != nil {
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
	if err := validate(t.TempDir(), "opinion", p); err != nil {
		t.Errorf("a legitimately falsy review_flag was treated as missing: %v", err)
	}
}
