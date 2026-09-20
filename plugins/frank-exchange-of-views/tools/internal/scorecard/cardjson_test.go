package scorecard

import (
	"encoding/json"
	"testing"
)

// A VALUE THAT IS ALREADY JSON MUST NOT BE MARSHALLED AGAIN, and that is the only reason CardJSON
// is more than a set of struct tags.
//
// objJSON is a string type holding a rendered object. Marshalled as a string it comes back quoted
// and escaped, so a reader that asked for structured output gets a string it has to parse a second
// time — which is precisely the position `show scorecard` put seats in before it had a JSON form
// at all. The assertion is that the round trip yields an OBJECT.
func TestAnAlreadyJSONValuePassesThroughAsAnObject(t *testing.T) {
	rows, err := CardJSON([]Row{
		{Clause: "c", Metric: "citation_yield_by_role", Value: objJSON(`{"lens":3,"chair":1}`)},
		{Clause: "c", Metric: "an_integer", Value: 7},
		{Clause: "c", Metric: "a_float", Value: 0.5},
		{Clause: "c", Metric: "a_string", Value: "words"},
		{Clause: "c", Metric: "not_computed", Value: nil, Note: "envelope-derived; fills in at capture"},
	})
	if err != nil {
		t.Fatal(err)
	}
	b, err := json.Marshal(rows)
	if err != nil {
		t.Fatal(err)
	}
	var back []map[string]any
	if err := json.Unmarshal(b, &back); err != nil {
		t.Fatalf("the card does not round-trip: %v\n%s", err, b)
	}

	obj, ok := back[0]["value"].(map[string]any)
	if !ok {
		t.Fatalf("an already-JSON value came back as %T (%v) — a reader asking for structure got text to parse again",
			back[0]["value"], back[0]["value"])
	}
	if obj["lens"] != float64(3) {
		t.Errorf("object value lost its contents: %v", obj)
	}
	if back[1]["value"] != float64(7) {
		t.Errorf("an int came back as %T %v, want a number", back[1]["value"], back[1]["value"])
	}
	if back[3]["value"] != "words" {
		t.Errorf("a genuine string came back as %v", back[3]["value"])
	}
	// NULL IS AN ANSWER, NOT AN ABSENCE. A "not computed" row is honest and says why in its note;
	// dropping the key would make it indistinguishable from a row nobody defined.
	v, present := back[4]["value"]
	if !present {
		t.Error("a not-computed row dropped its value key — absent reads as undefined, not as 'not computed'")
	}
	if v != nil {
		t.Errorf("a not-computed row carries %v, want null", v)
	}
	if back[4]["note"] == "" {
		t.Error("a not-computed row lost the note that says why")
	}
}
