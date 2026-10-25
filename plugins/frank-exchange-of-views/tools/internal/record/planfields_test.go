package record

import (
	"encoding/json"
	"reflect"
	"slices"
	"strings"
	"testing"
)

// EVERY FIELD OF THE PLAN IS ONE A SCHEMA CAN CARRY. The engine script's plan schema is generated
// from this walk, so a field of a Go kind with no JSON type, a named string type with no declared
// set, or a floor that is no integer would be a relay no harness can validate — refused here, at
// the declaration, rather than at generation.
func TestEveryPlanFieldIsOneASchemaCanCarry(t *testing.T) {
	var walk func(fs []PlanField, path string)
	walk = func(fs []PlanField, path string) {
		for _, f := range fs {
			if f.Defect != "" {
				t.Errorf("%s%s: %s", path, f.Name, f.Defect)
			}
			walk(f.Items, path+f.Name+"[].")
		}
	}
	walk(PlanFields(), "Plan.")
	walk(LensEnvelopeFields(), "LensEnvelope.")
}

// THE WALK AND THE VERB READ ONE DECLARATION. The keys PlanFields lists are the keys
// encoding/json prints for a plan with every list populated, at every depth; a field it marks
// optional is one the verb omits when empty, and no other is.
func TestThePlanWalkListsTheKeysTheVerbPrints(t *testing.T) {
	full := Plan{
		Parties:    []Party{{SeatID: "judge", GapIDs: []string{"G-1"}, Occasions: []string{"docket"}}},
		Docket:     []string{"G-1"},
		RemandOwed: []RemandOwed{{GapID: "G-1", Direction: "d"}},
		Exchanges:  []ExchangeCount{{GapID: "G-1"}},
		ToFile:     []string{"G-1"},
		StaleAreas: []StaleArea{{SeatID: "red-lens-voice"}},
		Blockers:   []PlanBlocker{{Kind: BlockerReopened}},
		Why:        []string{"w"},
		Verdict:    "PASS",
	}
	printed := func(p Plan) map[string]any {
		b, err := json.Marshal(p)
		if err != nil {
			t.Fatal(err)
		}
		var m map[string]any
		if err := json.Unmarshal(b, &m); err != nil {
			t.Fatal(err)
		}
		return m
	}
	var agree func(fs []PlanField, full, empty map[string]any, path string)
	agree = func(fs []PlanField, full, empty map[string]any, path string) {
		var walked []string
		for _, f := range fs {
			walked = append(walked, f.JSON)
			if _, printedEmpty := empty[f.JSON]; printedEmpty == f.Optional {
				t.Errorf("%s%s: the walk says optional=%v, and the verb prints it when empty: %v", path, f.JSON, f.Optional, printedEmpty)
			}
			if f.Of == nil {
				continue
			}
			item := func(m map[string]any) map[string]any {
				list, _ := m[f.JSON].([]any)
				if len(list) == 0 {
					t.Fatalf("%s%s: the populated plan prints no entry — extend `full` with one", path, f.JSON)
				}
				e, _ := list[0].(map[string]any)
				return e
			}
			zero := reflect.New(f.Of).Elem().Interface()
			zb, _ := json.Marshal(zero)
			var zm map[string]any
			_ = json.Unmarshal(zb, &zm)
			agree(f.Items, item(full), zm, path+f.JSON+"[].")
		}
		var keys []string
		for k := range full {
			keys = append(keys, k)
		}
		slices.Sort(keys)
		slices.Sort(walked)
		if !slices.Equal(keys, walked) {
			t.Errorf("%s: the verb prints %v and the walk lists %v", path, keys, walked)
		}
	}
	agree(PlanFields(), printed(full), printed(Plan{}), "")
}

// A PLAN THAT PERMITS NO PASS NAMES WHAT HOLDS IT. `blockers` is empty only when `pass_permitted`
// is true: PassPermitted is a report ingested, nobody ready and every blocker the chair's, so a
// plan without it has a party (each readied by a blocker on the gate's list), no report (the
// no_report blocker), or a blocker another seat owns. The declaration says so on the field, and
// the schema the chair's relay is held to is written from it.
func TestThePlanDeclaresItsBlockersEmptyOnlyWhenAPassIsPermitted(t *testing.T) {
	for _, f := range PlanFields() {
		if f.JSON != "blockers" {
			continue
		}
		if f.EmptyOnlyWith != "pass_permitted" {
			t.Fatalf("Plan.Blockers declares EmptyOnlyWith %q, want pass_permitted", f.EmptyOnlyWith)
		}
		return
	}
	t.Fatal("record.Plan relays no `blockers`")
}

// A rule naming no boolean of its own struct is a defect of the declaration, refused at the walk.
func TestAnEmptyOnlyWithNamingNoBooleanIsADefect(t *testing.T) {
	type spoiled struct {
		List []string `json:"list" emptyonlywith:"missing"`
		Head int      `json:"head"`
	}
	type spoiledType struct {
		List []string `json:"list" emptyonlywith:"head"`
		Head int      `json:"head"`
	}
	type scalar struct {
		Flag bool `json:"flag" emptyonlywith:"flag"`
	}
	for name, fs := range map[string][]PlanField{
		"a key the struct does not hold": fieldsOf(reflect.TypeFor[spoiled]()),
		"a key that is no boolean":       fieldsOf(reflect.TypeFor[spoiledType]()),
		"a field that is no list":        fieldsOf(reflect.TypeFor[scalar]()),
	} {
		if fs[0].Defect == "" {
			t.Errorf("%s: the walk accepts %+v", name, fs[0])
		}
	}
}

// THE VERB NEVER PRINTS A PLAN ITS SCHEMA REFUSES. A plan that permits no PASS and names no blocker
// is refused at the source, loudly, rather than handed to a chair whose harness will refuse every
// faithful relay of it.
func TestAPlanBreakingItsDeclaredRuleIsUnrelayable(t *testing.T) {
	ok := emptyPlan()
	ok.PassPermitted = true
	if err := relayable(ok); err != nil {
		t.Errorf("a PASS-permitted plan with no blockers is refused: %v", err)
	}
	held := emptyPlan()
	held.Blockers = []PlanBlocker{{Kind: BlockerNoReport}}
	if err := relayable(held); err != nil {
		t.Errorf("a plan naming what holds the PASS is refused: %v", err)
	}
	bad := emptyPlan()
	err := relayable(bad)
	if err == nil {
		t.Fatal("a plan that permits no PASS and names no blocker is relayable")
	}
	for _, w := range []string{"blockers", "pass_permitted"} {
		if !strings.Contains(err.Error(), w) {
			t.Errorf("the refusal %q does not name %s", err, w)
		}
	}
}
