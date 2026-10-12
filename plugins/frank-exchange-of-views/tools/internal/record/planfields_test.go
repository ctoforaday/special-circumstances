package record

import (
	"encoding/json"
	"reflect"
	"slices"
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
