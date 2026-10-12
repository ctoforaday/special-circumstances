package main

import (
	"reflect"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

type sampleParty struct{}
type sampleDebt struct{}

func sampleVocabulary() vocabulary {
	zero := 0
	party := reflect.TypeFor[sampleParty]()
	return vocabulary{
		grades: []string{"low", "high"}, dimensions: []string{"severity"},
		petitionRulings: []string{"granted", "denied"}, reliefBinds: []string{"blue"},
		dispositions: []string{"repaired", "moot"}, verdicts: []string{"PASS", "FAIL"}, outcomes: []string{"VERIFIED"},
		benchSeat: "judge", lensOrBlue: []string{"blue-respond", "red-lens-evidence"},
		dispatchOccasions: []string{"docket"}, relayBlockerKinds: []string{"reopened"},
		party: party,
		plan: []record.PlanField{
			{Name: "Head", JSON: "head", Type: "integer", Min: &zero, Desc: "the report's head"},
			{Name: "Parties", JSON: "parties", Type: "array", Of: party, Items: []record.PlanField{
				{Name: "SeatID", JSON: "seat_id", Type: "string"},
				{Name: "GapIDs", JSON: "gap_ids", Type: "array", Elem: "string", Desc: "a party's gaps", Bench: "the bench's docket"},
				{Name: "Occasions", JSON: "occasions", Type: "array", Elem: "string", Optional: true, Desc: "absent", Bench: "what it sits for"},
			}},
			{Name: "Debts", JSON: "debts", Type: "array", Desc: "what the run's owed", Of: reflect.TypeFor[sampleDebt](), Items: []record.PlanField{
				{Name: "Kind", JSON: "kind", Type: "string", Enum: []string{"infra", "source"}},
				{Name: "Why", JSON: "why", Type: "string", Pattern: `\S`, Desc: "never blank"},
				{Name: "Note", JSON: "note", Type: "string", Optional: true},
			}},
			{Name: "Ceiling", JSON: "ceiling", Type: "boolean"},
			{Name: "ToFile", JSON: "to_file", Type: "array", Elem: "string", Optional: true},
		},
	}
}

// The region states each set as the record holds it, a party's two shapes, and the plan as its
// declaration reads: every key, what is required, each field's type, floor, pattern, set and
// sentence, and the two type tables — where an optional field is in the schema and in neither
// table.
func TestTheVocabularyRegionStatesEachSetAndThePlanAsDeclared(t *testing.T) {
	got, err := renderVocabulary(sampleVocabulary())
	if err != nil {
		t.Fatal(err)
	}
	want := `const GRADE = { type: 'string', enum: ['low', 'high'] }
const DISPUTE_DIMENSION = { type: 'string', enum: ['severity'] }
const RULING_ON_PETITION = { type: 'string', enum: ['granted', 'denied'] }
const RELIEF_BINDS = { type: 'string', enum: ['blue'] }
const DISPOSITION = { type: 'string', enum: ['repaired', 'moot'] }
const VERDICT = { type: ['string', 'null'], enum: ['PASS', 'FAIL', null] }
const RUN_OUTCOME = { type: 'string', enum: ['VERIFIED'] }
const BENCH_PARTY_OCCASIONS = ['docket']
const KEYED_BLOCKER_KINDS = ['reopened']
const BENCH_PARTY = {
  type: 'object',
  required: ['seat_id', 'gap_ids', 'occasions'],
  properties: {
    seat_id: { type: 'string', enum: ['judge'] },
    gap_ids: { type: 'array', items: { type: 'string' }, description: "the bench's docket" },
    occasions: { type: 'array', minItems: 1, items: { type: 'string', enum: BENCH_PARTY_OCCASIONS }, description: 'what it sits for' },
  },
}
const LENS_OR_BLUE_PARTY = {
  type: 'object',
  required: ['seat_id', 'gap_ids'],
  properties: {
    seat_id: { type: 'string', enum: ['blue-respond', 'red-lens-evidence'] },
    gap_ids: { type: 'array', items: { type: 'string' }, description: "a party's gaps" },
    occasions: { type: 'array', maxItems: 0, description: 'absent' },
  },
}
const PLAN = {
  type: 'object',
  required: ['head', 'parties', 'debts', 'ceiling'],
  properties: {
    head: { type: 'integer', minimum: 0, description: "the report's head" },
    parties: { type: 'array', items: { anyOf: [BENCH_PARTY, LENS_OR_BLUE_PARTY] } },
    debts: { type: 'array', description: "what the run's owed", items: { type: 'object', required: ['kind', 'why'], properties: { kind: { type: 'string', enum: ['infra', 'source'] }, why: { type: 'string', pattern: '\\S', description: 'never blank' }, note: { type: 'string' } } } },
    ceiling: { type: 'boolean' },
    to_file: { type: 'array', items: { type: 'string' } },
  },
}
const PLAN_FIELD_TYPES = { head: 'integer', parties: 'array', debts: 'array', ceiling: 'boolean' }
const PLAN_ITEM_TYPES = {
  parties: { seat_id: 'string', gap_ids: 'array' },
  debts: { kind: 'string', why: 'string' },
}
`
	if string(got) != want {
		t.Errorf("the region renders as:\n%s\nwant:\n%s", got, want)
	}
}

// Each declaration a harness could not validate against is refused at generation, naming it.
func TestAVocabularyAHarnessCouldNotValidateAgainstIsRefused(t *testing.T) {
	cases := []struct {
		name  string
		spoil func(*vocabulary)
		want  string
	}{
		{"an empty set", func(v *vocabulary) { v.dispositions = nil }, "no word for the docket's dispositions"},
		{"no party seat", func(v *vocabulary) { v.lensOrBlue = nil }, "no seat a plan's party can be"},
		{"no dispatch occasion", func(v *vocabulary) { v.dispatchOccasions = nil }, "for no occasion"},
		{"no relay-held blocker kind", func(v *vocabulary) { v.relayBlockerKinds = nil }, "on no kind"},
		{"a plan with no parties", func(v *vocabulary) { v.plan = v.plan[:1] }, "no list of parties"},
		{"a party with no occasions", func(v *vocabulary) { v.plan[1].Items = v.plan[1].Items[:2] }, "no field Occasions"},
		{"a party whose occasions are required", func(v *vocabulary) { v.plan[1].Items[2].Optional = false }, "not an optional list of words"},
		{"a field no schema can carry", func(v *vocabulary) { v.plan[3].Defect = "its Go type map[string]int has no JSON type here" }, "ceiling cannot be relayed under a schema"},
		{"a nested field no schema can carry", func(v *vocabulary) { v.plan[2].Items[0].Defect = "no set" }, "kind cannot be relayed under a schema"},
		{"a key that is no word", func(v *vocabulary) { v.plan[0].JSON = "report-head" }, `"report-head" is not a word`},
		{"a word a script string cannot carry", func(v *vocabulary) { v.grades = []string{"lo'w"} }, "cannot carry"},
		{"a sentence holding both quotes", func(v *vocabulary) { v.plan[0].Desc = `the report's "head"` }, "cannot carry"},
	}
	for _, c := range cases {
		v := sampleVocabulary()
		c.spoil(&v)
		out, err := renderVocabulary(v)
		if err == nil {
			t.Errorf("%s: rendered\n%s", c.name, out)
			continue
		}
		if !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: refused with %q, which does not name it (%q)", c.name, err, c.want)
		}
	}
}

// The record's own declarations render: every set holds a word, and record.Plan holds no field a
// schema cannot carry. `enginegen -check` compares the committed region to this rendering.
func TestTheRecordsVocabularyRenders(t *testing.T) {
	v, err := vocabularyFromTheRecord()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := renderVocabulary(v); err != nil {
		t.Fatal(err)
	}
}
