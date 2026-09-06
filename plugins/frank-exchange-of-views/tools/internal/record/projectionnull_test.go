package record

import (
	"encoding/json"
	"strings"
	"testing"
)

// NO PROJECTION LIST MAY MARSHAL AS `null`.
//
// Dropping `omitempty` stopped an unset field from vanishing. A nil SLICE re-creates the same
// ambiguity one token along: a reader cannot tell "no ancestors" from "not computed". The record
// already states the rule at the write path — a gap with no ancestors records `supersedes: []`,
// because an absent key would read as lineage UNKNOWN where the truth is lineage NONE — and the
// projection owes the same answer.
//
// `null` is CORRECT for a scalar or a struct pointer: an ungraded gap genuinely has no severity,
// and a run with no sitting genuinely has none. Only lists are policed here.
func TestNoProjectionListMarshalsAsNull(t *testing.T) {
	listFields := map[string]bool{
		"open": true, "closed": true, "observations": true, "anomalies": true,
		"found_by": true, "supersedes": true, "regrades": true, "findings": true,
		"closed_index": true, "outstanding": true, "available": true,
		"motions": true, "proofs": true, "citations": true, "violations": true,
	}
	b := &Board{}
	for name, v := range map[string]any{
		"BoardJSON": BoardJSONOf(b),
		"WorkJSON":  workJSONOfGaps(workStatesOfBoardT(b)),
	} {
		raw, err := json.Marshal(v)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		var generic map[string]any
		if err := json.Unmarshal(raw, &generic); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		for k, val := range generic {
			if listFields[k] && val == nil {
				t.Errorf("%s.%s marshalled as null — emit [] so \"none\" is distinguishable "+
					"from \"not computed\"", name, k)
			}
		}
		if strings.Contains(string(raw), `"anomalies":null`) {
			t.Errorf("%s: anomalies is null", name)
		}
	}
}
