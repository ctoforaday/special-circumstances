package surface

import (
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/repotree"
)

// AN ENVELOPE THAT DECLARES A VALUE THE RECORD REFUSES IS A PATH THAT CANNOT WORK.
//
// MEASURED (#329). debate.js's petition-ruling schema declared `ruling: ['granted','denied',
// 'halt']`. The record's petition-rule enum is `['granted','denied']` — deliberately, because a
// halt is the bench's own first-class terminal act rather than a petition disposition. So the
// schema told the judge a halt was a ruling, the prompt told it to record rulings with
// `petition-rule`, and the tool REFUSED the write. The engine halted off the envelope while the
// record carried no halt event: the report never said the bench halted, and the halt opinion —
// which must reach the human VERBATIM — was on no record anywhere.
//
// It survived because every existing gate was looking one level away. The prompt-verb gate checks
// that a named VERB exists, not that a named VALUE does. The enum-coverage gate checks the fuzz
// drives every record enum value, and `halt` is not a record enum value at all. And the fuzz's own
// judge already did the right thing — it drove `bench halt` — so THE FAKE WAS CORRECT WHILE
// PRODUCTION WAS NOT, and 60 green runs a sweep said nothing about the path a real judge would take.
//
// The vocabulary collision was the cause, not the symptom: while `halt` sat in that enum, calling
// `petition-rule --as halt` was the natural thing to write. This gate makes the collision
// unwriteable — every enum a seat-facing envelope declares must equal the record's enum for the
// same field, or be registered as engine-only with its reason.

// schemaEnum matches a FIELD enum in debate.js: `foo: { type: 'string', enum: ['a','b'] }`.
var schemaEnum = regexp.MustCompile(`(\w+)\s*:\s*\{[^}]*enum:\s*\[([^\]]*)\]`)

// constEnum matches a whole schema constant that IS an enum: `const GRADE = { type: 'string',
// enum: [...] }`. The field regex cannot see these — it needs a `name: {` — so the first draft of
// this gate silently skipped GRADE and DISPUTE_DIMENSION entirely while reporting a pass.
var constEnum = regexp.MustCompile(`^const ([A-Z_]+) = \{[^}]*enum:\s*\[([^\]]*)\]`)

// envelopeEnumBinding maps an envelope enum to the record enum it must agree with, keyed
// SCHEMA.field.
//
// Keyed by its schema, not by the bare field name: `class` means the petition class in one schema
// and the closure class in another, and a name-only key silently bound the first to the second —
// this gate's own first draft did exactly that, which is the same collision it exists to catch.
// engineOnly and recordOnly are the DECLARED asymmetries: values one side legitimately has and
// the other does not. Both are named rather than exempted wholesale, so a fourth of either kind
// fails this gate until someone says what it is. See JUDGE_ENVELOPE.resolution for the case that
// forced them.
type enumBind struct {
	typ, key   string
	engineOnly []string // in the envelope, deliberately not a record value
	recordOnly []string // in the record, with no envelope word — a KNOWN gap, not a blessing
}

var envelopeEnumBinding = map[string]enumBind{
	// After #344 the adjudication vocabularies are keyed on (SUBJECT, key) rather than by event
	// type, so these bind to a motion SUBJECT. The `typ` field carries `motion:<subject>` for
	// those, which recordEnumValues resolves against record.MotionVerdicts / record.MotionFields.
	// The envelope still speaks in per-exchange schemas — PETITIONS, PETITION_RULING,
	// DISPUTE_DIMENSION — because the ENGINE routes them separately; what collapsed is the
	// RECORD's vocabulary, and this gate is precisely the check that the two still agree.
	"PETITIONS.class":        {typ: "motion:petition", key: "class"},
	"PETITION_RULING.ruling": {typ: "motion:petition", key: "ruling"},
	// WHO GRANTED RELIEF BINDS. The engine routes the relief into that party's prompt, and the
	// record stores the same word on the ruling — so a value the record would refuse must not be
	// one the schema invites. This binding is the reason `binds` was added to MotionFields rather
	// than left as a free string the engine alone understood.
	"PETITION_RULING.binds":    {typ: "motion:petition", key: "binds"},
	"DISPUTE_DIMENSION.<self>": {typ: "motion:grade", key: "dimension"},

	// THE BENCH'S DISPOSITION, AND THE TWO VOCABULARIES ARE ONE AGAIN (#847).
	//
	// This was EXEMPT, on a reason that was true — the engine's vocabulary is its own, the
	// recorded form is checked at its own write path — and that is why nobody looked. Exempting
	// the FIELD meant nothing checked what the exemption protected, and the two drifted to five
	// words of ten disagreeing.
	//
	// Bound with no declared asymmetry in either direction, which is the strongest form this
	// entry can take: every word the envelope offers is a word the record accepts, and every word
	// the record accepts has an envelope word and therefore a stated duty for the other party.
	// A value added to either side now fails this gate until it is added to both or declared.
	//
	// Getting here took three changes rather than a mapping: `moot` became a record disposition
	// (it asserts neither the argument not_a_defect claims nor the verification repaired claims);
	// `grade_adjusted` left, because it is a GRADE MOTION's outcome and this envelope already
	// carries grade_disputes for that; `unresolved` left as a duplicate of `carried`.
	"JUDGE_ENVELOPE.resolution": {typ: "motion:docket", key: "ruling"},
}

// envelopeEnumExempt are envelope enums with no record counterpart, each with its reason. These
// are ENGINE vocabularies — values the script routes on that never become a payload field.
var envelopeEnumExempt = map[string]string{
	"CHAIR_ENVELOPE.verdict": "the chair's PASS|FAIL to the engine, restated — the `verdict` event carries the same two values and is the original; this field is the script's loop condition and is not written as a payload",
	"GRADE.<self>":           "the grade scale, shared by every graded axis and validated at the record's write path against record.MASS rather than by a per-field enum",
}

func TestEveryEnvelopeEnumAgreesWithTheRecord(t *testing.T) {
	path, err := repotree.DebateJS()
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("cannot read the engine script: %v", err)
	}
	// Walk the file tracking the schema constant each enum sits inside, so every enum has a
	// UNIQUE address. `class` means the petition class in one schema and the closure class in
	// another, and keying on the bare name silently bound the first to the second — this gate's
	// own first draft did that, which is the same collision it exists to catch.
	schemaDecl := regexp.MustCompile(`^const ([A-Z_]+) = \{`)
	current := ""
	found := 0
	seen := map[string]bool{}
	for _, line := range strings.Split(string(b), "\n") {
		if m := schemaDecl.FindStringSubmatch(line); m != nil {
			current = m[1]
		}
		key, raw := "", ""
		if m := constEnum.FindStringSubmatch(line); m != nil {
			key, raw = m[1]+".<self>", m[2]
		} else if m := schemaEnum.FindStringSubmatch(line); m != nil {
			key, raw = current+"."+m[1], m[2]
		} else {
			continue
		}
		found++
		if seen[key] {
			continue
		}
		seen[key] = true

		bind, bound := envelopeEnumBinding[key]
		if !bound {
			if envelopeEnumExempt[key] == "" {
				t.Errorf("envelope enum %q is bound to no record enum and has no stated exemption.\n"+
					"Register it (which record enum must it equal?) or say why it is engine-only. An unregistered\n"+
					"envelope vocabulary is one nobody has checked the record will accept — #329 exactly.", key)
			}
			continue
		}
		want := recordEnumValues(t, bind.typ, bind.key)
		got := parseJSValues(raw)
		// EACH ASYMMETRY MUST BE DECLARED. A value the envelope has and the record does not is
		// legal only if named in engineOnly; a value the record has and the envelope does not is
		// legal only if named in recordOnly. Anything else is drift, and the message says which
		// direction it went so the reader is not left diffing two lists by eye.
		if extra := missing(got, append(append([]string{}, want...), bind.engineOnly...)); len(extra) > 0 {
			t.Errorf("envelope enum %q offers %v, which the record's %s.%s does not accept and which is\n"+
				"not declared in engineOnly. A seat told it may return a value the tool refuses writes NOTHING\n"+
				"when it tries — and the engine, reading the envelope, carries on as though it had (#329).",
				key, extra, bind.typ, bind.key)
		}
		if absent := missing(want, append(append([]string{}, got...), bind.recordOnly...)); len(absent) > 0 {
			t.Errorf("the record's %s.%s accepts %v, which envelope enum %q does not offer and which is not\n"+
				"declared in recordOnly. A bench recording one of these has no envelope word for it, so the\n"+
				"other party is handed a different fate's duty or the UNMAPPED FATE default.",
				bind.typ, bind.key, absent, key)
		}
	}
	// AN ENTRY FOR AN ENUM THAT NO LONGER EXISTS IS A CLAIM ABOUT NOTHING, and it reads exactly
	// like coverage. The first draft of these tables carried five such entries — bindings and
	// exemptions for fields the scanner never produced — while the gate reported a pass.
	for k := range envelopeEnumBinding {
		if !seen[k] {
			t.Errorf("binding %q matches no enum in debate.js — a binding for a field that does not exist is checked coverage of nothing", k)
		}
	}
	for k := range envelopeEnumExempt {
		if !seen[k] {
			t.Errorf("exemption %q matches no enum in debate.js — remove it; a reason for an absent thing reads as a checked one", k)
		}
	}
	if found == 0 {
		// A regex that matches nothing reports a clean board. Refuse that outright.
		t.Fatal("found NO envelope schema enums in debate.js — the schema's shape changed and this gate is measuring nothing, which reads exactly like a pass")
	}
}

func recordEnumValues(t *testing.T, typ, key string) []string {
	t.Helper()
	// `motion:<subject>` resolves against the (subject, key) tables, which EnumFields cannot
	// express — one `motion-rule` carries granted|denied for a petition and accepted|rejected
	// for a grade.
	if subject, ok := strings.CutPrefix(typ, "motion:"); ok {
		var values []record.EnumValue
		if key == "ruling" {
			values = record.MotionVerdicts[subject]
		} else {
			values = record.MotionFields[subject][key]
		}
		if len(values) == 0 {
			t.Fatalf("no motion enum %s.%s — the binding table names a subject field the record does not have", subject, key)
		}
		out := record.Names(values)
		sort.Strings(out)
		return out
	}
	for _, e := range record.EnumFields[typ] {
		if e.Key == key {
			out := record.Names(e.Values)
			sort.Strings(out)
			return out
		}
	}
	t.Fatalf("no record enum %s.%s — the binding table names a field the record does not have", typ, key)
	return nil
}

func parseJSValues(raw string) []string {
	var out []string
	for _, p := range strings.Split(raw, ",") {
		if v := strings.Trim(strings.TrimSpace(p), "'\""); v != "" {
			out = append(out, v)
		}
	}
	sort.Strings(out)
	return out
}

func equalSets(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// missing returns the members of got that allowed does not contain — the one direction of a set
// difference, so each side of an asymmetry is reported in its own words rather than as two lists
// for the reader to diff by eye.
func missing(got, allowed []string) []string {
	have := map[string]bool{}
	for _, a := range allowed {
		have[a] = true
	}
	var out []string
	for _, g := range got {
		if !have[g] {
			out = append(out, g)
		}
	}
	sort.Strings(out)
	return out
}
