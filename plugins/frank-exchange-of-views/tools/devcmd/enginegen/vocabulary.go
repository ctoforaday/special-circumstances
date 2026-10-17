package main

import (
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

const (
	vocabBegin = "// BEGIN GENERATED ENVELOPE VOCABULARY — go -C plugins/frank-exchange-of-views/tools run ./devcmd/enginegen writes this from the record's enums and record.Plan. DO NOT EDIT BY HAND."
	vocabEnd   = "// END GENERATED ENVELOPE VOCABULARY"
)

// vocabulary is every closed set a seat's envelope offers and the shape of the plan the chair
// relays, as Go holds them.
//
// AN ENVELOPE WORD THE RECORD REFUSES IS A PATH THAT CANNOT WORK (#329): the petition-ruling
// schema once offered `halt`, the record's enum did not hold it, and a judge following the schema
// ran a write the tool refused while the engine halted off the envelope. Each set here is the
// record's own — the list its write path resolves against — so the envelope cannot offer a word
// the tool refuses or omit one it accepts.
type vocabulary struct {
	grades          []string // recordpb.GradeWords
	dimensions      []string // a grade motion's `dimension`
	petitionRulings []string // a petition motion's ruling
	reliefBinds     []string // a petition ruling's `binds`
	dispositions    []string // a docket motion's ruling
	verdicts        []string // the chair's verdict, in capitals
	outcomes        []string // the run's outcome, in capitals

	// benchSeat and lensOrBlue are the seats a plan's party can name (record.PartySeats), in the
	// two shapes a party takes; dispatchOccasions is what the bench's party can carry.
	benchSeat         string
	lensOrBlue        []string
	dispatchOccasions []string

	// plan is record.Plan's relayed fields, and party the struct a party is.
	plan  []record.PlanField
	party reflect.Type
}

func vocabularyFromTheRecord() (vocabulary, error) {
	v := vocabulary{
		grades:            recordpb.GradeWords(),
		dimensions:        seatWords(record.MotionFields["grade"]["dimension"]),
		petitionRulings:   seatWords(record.MotionVerdicts["petition"]),
		reliefBinds:       seatWords(record.MotionFields["petition"]["binds"]),
		dispositions:      seatWords(record.MotionVerdicts["docket"]),
		dispatchOccasions: record.DispatchOccasions(),
		plan:              record.PlanFields(),
		party:             reflect.TypeFor[record.Party](),
	}
	var err error
	if v.verdicts, err = eventEnum("verdict", "verdict"); err != nil {
		return v, err
	}
	if v.outcomes, err = eventEnum("outcome", "verdict"); err != nil {
		return v, err
	}
	v.benchSeat, v.lensOrBlue = record.PartySeats()
	return v, nil
}

// seatWords is the words of a set a seat may return: the record's, without the ones only the tool
// writes.
func seatWords(vs []record.EnumValue) []string { return record.Names(record.SeatFilable(vs)) }

// eventEnum is the closed set an event type carries under a key (record.EnumFields).
func eventEnum(typ, key string) ([]string, error) {
	for _, e := range record.EnumFields[typ] {
		if e.Key == key {
			return seatWords(e.Values), nil
		}
	}
	return nil, fmt.Errorf("the record declares no closed set %s.%s", typ, key)
}

// renderVocabulary is the vocabulary region's lines, between the markers and not including them.
//
// IT REFUSES WHAT A HARNESS COULD NOT VALIDATE AGAINST. An empty set renders as an enum that
// accepts nothing, and a plan field no schema can carry renders as a relay the chair's harness
// refuses forever; each is a defect of the declaration, and writing it out would move the failure
// into a run.
func renderVocabulary(v vocabulary) ([]byte, error) {
	var b strings.Builder
	sets := []struct {
		name, what string
		words      []string
	}{
		{"GRADE", "the grade scale", v.grades},
		{"DISPUTE_DIMENSION", "a grade motion's dimensions", v.dimensions},
		{"RULING_ON_PETITION", "a petition's rulings", v.petitionRulings},
		{"RELIEF_BINDS", "whom a petition's relief binds", v.reliefBinds},
		{"DISPOSITION", "the docket's dispositions", v.dispositions},
		{"VERDICT", "the chair's verdicts", v.verdicts},
		{"RUN_OUTCOME", "the run's outcomes", v.outcomes},
	}
	for _, s := range sets {
		if len(s.words) == 0 {
			return nil, fmt.Errorf("the record holds no word for %s", s.what)
		}
		typ, enum := js("string"), words(s.words)
		if s.name == "VERDICT" {
			// NULL IS "RECORDED NONE", as the field's absence is: a chair sitting that records no
			// verdict writes no verdict event, and its envelope says so with null or by omission.
			typ, enum = "['string', 'null']", list(append(each(s.words, js), "null"))
		}
		fmt.Fprintf(&b, "const %s = %s\n", s.name, object(prop{"type", typ}, prop{"enum", enum}))
	}

	if v.benchSeat == "" || len(v.lensOrBlue) == 0 {
		return nil, fmt.Errorf("the record names no seat a plan's party can be")
	}
	if len(v.dispatchOccasions) == 0 {
		return nil, fmt.Errorf("the record convenes the bench's party for no occasion")
	}
	fmt.Fprintf(&b, "const BENCH_PARTY_OCCASIONS = %s\n", words(v.dispatchOccasions))

	if err := v.renderPlan(&b); err != nil {
		return nil, err
	}
	if strings.Contains(b.String(), unwritable) {
		return nil, fmt.Errorf("a vocabulary word or a plan field's description holds a character a script string cannot carry")
	}
	return []byte(b.String()), nil
}

// renderPlan writes the two party shapes, the plan's schema, and the two type tables the
// engine's own check of a relay walks.
func (v vocabulary) renderPlan(b *strings.Builder) error {
	if err := sound(v.plan); err != nil {
		return err
	}
	var party []record.PlanField
	for _, f := range v.plan {
		if f.Of == v.party {
			party = f.Items
		}
	}
	if party == nil {
		return fmt.Errorf("record.Plan carries no list of parties")
	}
	if err := v.renderParties(b, party); err != nil {
		return err
	}

	var required, props, types, itemTypes []prop
	for _, f := range v.plan {
		props = append(props, prop{f.JSON, v.schemaOf(f, f.Desc)})
		if f.Optional {
			continue
		}
		required = append(required, prop{val: js(f.JSON)})
		types = append(types, prop{f.JSON, js(f.Type)})
		if f.Of != nil {
			var of []prop
			for _, i := range f.Items {
				if !i.Optional {
					of = append(of, prop{i.JSON, js(i.Type)})
				}
			}
			itemTypes = append(itemTypes, prop{f.JSON, object(of...)})
		}
	}
	if len(props) == 0 {
		return fmt.Errorf("record.Plan relays no field")
	}
	rule, emptyOnlyWith, err := emptyOnlyWithOf(v.plan)
	if err != nil {
		return err
	}
	var extra []prop
	if rule != "" {
		extra = append(extra, prop{"anyOf", rule})
	}
	b.WriteString(block("PLAN", required, props, extra...))
	fmt.Fprintf(b, "const PLAN_FIELD_TYPES = %s\n", object(types...))
	fmt.Fprintf(b, "const PLAN_EMPTY_ONLY_WITH = %s\n", object(emptyOnlyWith...))
	b.WriteString("const PLAN_ITEM_TYPES = {\n")
	for _, t := range itemTypes {
		fmt.Fprintf(b, "  %s: %s,\n", t.key, t.val)
	}
	b.WriteString("}\n")
	return nil
}

// emptyOnlyWithOf is the plan's one EmptyOnlyWith rule as a schema clause — the list non-empty, or
// the boolean true — and as the table the engine's own check walks. A rule naming no boolean of
// the plan, or held by a nested field or by a second field, is refused: the clause is written for
// one rule on the plan itself.
func emptyOnlyWithOf(fs []record.PlanField) (string, []prop, error) {
	var rules []prop
	var list, flag string
	for _, f := range fs {
		for _, i := range f.Items {
			if i.EmptyOnlyWith != "" {
				return "", nil, fmt.Errorf("record.Plan's %s.%s states an emptyonlywith rule, and the schema states one only on the plan's own fields", f.JSON, i.JSON)
			}
		}
		if f.EmptyOnlyWith == "" {
			continue
		}
		at := slices.IndexFunc(fs, func(g record.PlanField) bool { return g.JSON == f.EmptyOnlyWith })
		if f.Type != "array" || f.Optional || at < 0 || fs[at].Type != "boolean" || fs[at].Optional {
			return "", nil, fmt.Errorf("record.Plan's %s is empty only with %q, and the schema states that rule only for a required list and a required boolean of the plan", f.JSON, f.EmptyOnlyWith)
		}
		rules = append(rules, prop{f.JSON, js(f.EmptyOnlyWith)})
		list, flag = f.JSON, f.EmptyOnlyWith
	}
	switch len(rules) {
	case 0:
		return "", nil, nil
	case 1:
	default:
		return "", nil, fmt.Errorf("record.Plan states %d emptyonlywith rules, and the schema states one", len(rules))
	}
	// EACH BRANCH'S PROPERTY STATES ITS TYPE beside the keyword that needs one. A branch's property is
	// a new data location, so the plan's `type` does not reach it, and the harness's strict-types
	// compile (ajv's strictTypes) warns at launch on a `minItems` with no `type: "array"`. The enum
	// states its type for the same reading.
	rule := "[" + object(prop{"required", words([]string{flag})}, prop{"properties", object(prop{flag, object(prop{"type", js("boolean")}, prop{"enum", "[true]"})})}) +
		", " + object(prop{"required", words([]string{list})}, prop{"properties", object(prop{list, object(prop{"type", js("array")}, prop{"minItems", "1"})})}) + "]"
	return rule, rules, nil
}

// renderParties writes a party's two shapes from record.Party's fields.
//
// ONE STRUCT, TWO SHAPES, AND THE SEAT SAYS WHICH — the rule requireDispatchOccasions holds a
// dispatch row to. The bench's party names the bench's seat, states every field, and carries at
// least one of the occasions the chair convenes it for. A lens's or blue's names one of their
// seats and carries no occasion: the verb omits the field for them, and an empty list is the same
// state. A field with a `bench` sentence carries it in the bench's shape and its `desc` in the
// other.
func (v vocabulary) renderParties(b *strings.Builder, fields []record.PlanField) error {
	const seat, occasions = "SeatID", "Occasions"
	has := map[string]bool{}
	var benchRequired, benchProps, otherRequired, otherProps []prop
	for _, f := range fields {
		has[f.Name] = true
		benchRequired = append(benchRequired, prop{val: js(f.JSON)})
		if !f.Optional {
			otherRequired = append(otherRequired, prop{val: js(f.JSON)})
		}
		switch f.Name {
		case seat:
			benchProps = append(benchProps, prop{f.JSON, object(prop{"type", js(f.Type)}, prop{"enum", words([]string{v.benchSeat})})})
			otherProps = append(otherProps, prop{f.JSON, object(prop{"type", js(f.Type)}, prop{"enum", words(v.lensOrBlue)})})
		case occasions:
			if f.Type != "array" || f.Elem != "string" || !f.Optional {
				return fmt.Errorf("record.Party's %s is not an optional list of words, and the two party shapes are written for one", f.Name)
			}
			benchProps = append(benchProps, prop{f.JSON, object(described([]prop{{"type", js(f.Type)}, {"minItems", "1"},
				{"items", object(prop{"type", js(f.Elem)}, prop{"enum", "BENCH_PARTY_OCCASIONS"})}}, f.Bench)...)})
			otherProps = append(otherProps, prop{f.JSON, object(described([]prop{{"type", js(f.Type)}, {"maxItems", "0"}}, f.Desc)...)})
		default:
			desc := f.Bench
			if desc == "" {
				desc = f.Desc
			}
			benchProps = append(benchProps, prop{f.JSON, v.schemaOf(f, desc)})
			otherProps = append(otherProps, prop{f.JSON, v.schemaOf(f, f.Desc)})
		}
	}
	for _, name := range []string{seat, occasions} {
		if !has[name] {
			return fmt.Errorf("record.Party has no field %s, and the two party shapes are told apart by it", name)
		}
	}
	b.WriteString(block("BENCH_PARTY", benchRequired, benchProps))
	b.WriteString(block("LENS_OR_BLUE_PARTY", otherRequired, otherProps))
	return nil
}

// sound refuses a plan field no schema can carry, or whose key the engine could not read as one.
func sound(fs []record.PlanField) error {
	for _, f := range fs {
		if f.Defect != "" {
			return fmt.Errorf("record.Plan's %s cannot be relayed under a schema: %s", f.JSON, f.Defect)
		}
		if !identifier.MatchString(f.JSON) {
			return fmt.Errorf("record.Plan's key %q is not a word the engine can read as a key", f.JSON)
		}
		if err := sound(f.Items); err != nil {
			return err
		}
	}
	return nil
}

// described adds a description to a schema's keys where the field states one.
func described(ps []prop, desc string) []prop {
	if desc == "" {
		return ps
	}
	return append(ps, prop{"description", text(desc)})
}

// schemaOf is one plan field's schema, under the description given.
func (v vocabulary) schemaOf(f record.PlanField, desc string) string {
	ps := []prop{{"type", js(f.Type)}}
	if f.Min != nil {
		ps = append(ps, prop{"minimum", strconv.Itoa(*f.Min)})
	}
	switch {
	case f.Of == v.party:
		ps = described(append(ps, prop{"items", object(prop{"anyOf", "[BENCH_PARTY, LENS_OR_BLUE_PARTY]"})}), desc)
	case f.Of != nil:
		var required, props []prop
		for _, i := range f.Items {
			props = append(props, prop{i.JSON, v.schemaOf(i, i.Desc)})
			if !i.Optional {
				required = append(required, prop{val: js(i.JSON)})
			}
		}
		ps = append(described(ps, desc), prop{"items", object(prop{"type", js("object")}, prop{"required", list(vals(required))}, prop{"additionalProperties", "false"}, prop{"properties", object(props...)})})
	default:
		if f.Type == "array" {
			ps = append(ps, prop{"items", object(prop{"type", js(f.Elem)})})
		}
		if f.Enum != nil {
			ps = append(ps, prop{"enum", words(f.Enum)})
		}
		if f.Pattern != "" {
			ps = append(ps, prop{"pattern", text(f.Pattern)})
		}
		ps = described(ps, desc)
	}
	return object(ps...)
}

// prop is one key of a script object and its value, already written as script source.
type prop struct{ key, val string }

func vals(ps []prop) []string {
	out := make([]string, len(ps))
	for i, p := range ps {
		out[i] = p.val
	}
	return out
}

func object(ps ...prop) string {
	out := make([]string, len(ps))
	for i, p := range ps {
		out[i] = p.key + ": " + p.val
	}
	return "{ " + strings.Join(out, ", ") + " }"
}

func list(vs []string) string { return "[" + strings.Join(vs, ", ") + "]" }

func words(ws []string) string { return list(each(ws, js)) }

// block is an object schema declared as a constant, one property to a line, then any keys of the
// schema itself that follow its properties.
//
// CLOSED: A KEY THE DECLARATION DOES NOT HOLD IS REFUSED. The verb prints exactly the keys the
// declaration walks (TestThePlanWalkListsTheKeysTheVerbPrints), so a key it does not hold was added
// on the way. Open, the schema took o3's re-sent relay — its first, nested inside a `plan` key, with
// the required fields added beside it and emptied — and the engine dispatched nobody on it.
func block(name string, required, props []prop, extra ...prop) string {
	var b strings.Builder
	fmt.Fprintf(&b, "const %s = {\n  type: 'object',\n  required: %s,\n  additionalProperties: false,\n  properties: {\n", name, list(vals(required)))
	for _, p := range props {
		fmt.Fprintf(&b, "    %s: %s,\n", p.key, p.val)
	}
	b.WriteString("  },\n")
	for _, p := range extra {
		fmt.Fprintf(&b, "  %s: %s,\n", p.key, p.val)
	}
	b.WriteString("}\n")
	return b.String()
}

// text is a sentence or a pattern as a script string: single-quoted, or double-quoted where it
// holds an apostrophe. One holding both quotes, a backtick or a line break is refused by
// renderVocabulary rather than escaped by a rule only that sentence would exercise.
func text(v string) string {
	if strings.ContainsAny(v, "\n\r`"+unwritable) {
		return unwritable
	}
	v = strings.ReplaceAll(v, `\`, `\\`)
	switch {
	case !strings.Contains(v, "'"):
		return "'" + v + "'"
	case !strings.Contains(v, `"`):
		return `"` + v + `"`
	}
	return unwritable
}
