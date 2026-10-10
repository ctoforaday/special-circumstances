package surface

import (
	"os"
	"regexp"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/repotree"
)

// envelopeRestated is a schema property — `name: { … }` or `name: SCHEMA_CONST,` — for a fact the
// record carries: a petition is a motion on the record, what stands unruled is the plan's blockers,
// a seat's log entry is a log event, the report's claim count is counted from the report, and what
// a sitting put on the record is read off the record. An envelope field for any of them is a
// second channel for one fact, and the engine reading it is a dual reader (#1203, #1267, #1298).
var envelopeRestated = regexp.MustCompile(`(?m)^\s*(petitions|unruled_motions|log|claim_count|sitting_record_appended):\s*(\{|[A-Z_]+,)`)

// envelopeRelayNamed is one of the two relayed facts named ANYWHERE in the engine script — a
// schema's `required` list, a read, a result field, a prompt sentence asking a seat to return it.
// The script has no use for either word, so a hit is the channel coming back by any door.
var envelopeRelayNamed = regexp.MustCompile(`claim_count|blue_claims|sitting_record_appended|sitting_record_unresolved|ensureSittingRecord`)

// envelopeRelayRead is a Go reader indexing either relayed key out of a decoded envelope or a
// journal result: `r["claim_count"]`, `env["sitting_record_appended"]`. The count's own JSON key
// on `count-claims` is a struct tag, never an index, so the tool's own figure does not match.
var envelopeRelayRead = regexp.MustCompile(`\[\s*"(claim_count|sitting_record_appended)"\s*\]`)

// envelopeLogRead is the engine indexing a `log` member of anything: `env.log`, `env['log']`.
// The script's own `log(…)` call is a function, never a member, so the script holds no such
// access at all and any one of them is a returned envelope's log being read.
var envelopeLogRead = regexp.MustCompile("\\.log\\b|\\[\\s*['\"`]log['\"`]\\s*\\]")

// NO ENVELOPE RESTATES A FACT THE RECORD CARRIES. The petition channel was an envelope field that
// lenses — whose returns are prose — could never fill, so their petitions reached no bench; the
// terminal sitting fired on the chair's own count of unruled motions, which nothing audited; and
// the envelope's log was a copy of entries every seat writes to the record, which nine of thirteen
// seat types could not return. Petitions and unruled motions come from the record through the
// chair's plan; the log is read from the record and relayed by nobody.
//
// TWO MORE WERE A SEAT'S WORD ABOUT WHAT THE TOOL ALREADY KNEW: the claim count blue typed into
// its envelope beside the tool's own count, and the attestation that its sitting was on the record
// — which the engine answered with a re-prompt, and a scorecard row scored. The count is counted
// from the report by the reader that wants it, and a sitting short of what it owes is named by the
// record's own readers. This is the gate against any of the five channels being added back, as a
// schema property, as a read by the engine, or as a read by a Go consumer of the envelopes.
func TestNoEnvelopeRestatesARecordFact(t *testing.T) {
	path, err := repotree.DebateJS()
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("cannot read the engine script: %v", err)
	}
	for _, m := range envelopeRestated.FindAllStringSubmatch(string(b), -1) {
		t.Errorf("debate.js declares an envelope field %q — a fact the record carries; a seat's envelope restating it is a second channel the engine must not read", m[1])
	}
	for _, m := range envelopeLogRead.FindAllString(string(b), -1) {
		t.Errorf("debate.js reads a `log` member (%q) — a seat's log entries are on the record, and an envelope copy is a second channel the engine must not read", m)
	}
	for _, m := range envelopeRelayNamed.FindAllString(string(b), -1) {
		t.Errorf("debate.js names %q — the record computes that fact, and no seat relays or attests it through an envelope", m)
	}
	srcs, err := repotree.ToolSources()
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range srcs {
		src, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		for _, m := range envelopeRelayRead.FindAllString(string(src), -1) {
			t.Errorf("%s reads %s out of an envelope — a seat's relay of a fact the record computes", p, m)
		}
	}
	// The patterns must still see a schema property and a read, or a reshaped script reads as a
	// clean board.
	for _, line := range []string{"    petitions: {\n", "    unruled_motions: { type: 'integer' },\n", "    log: { type: 'array', items: { type: 'string' } },\n",
		"    claim_count: { type: 'number' },\n", "    sitting_record_appended: { type: 'boolean' },\n"} {
		if !envelopeRestated.MatchString(line) {
			t.Fatalf("the schema pattern no longer matches %q — this gate is measuring nothing", line)
		}
	}
	for _, read := range []string{"if (env && env.log) for (const f of env.log)", "env['log']", `env["log"]`} {
		if !envelopeLogRead.MatchString(read) {
			t.Fatalf("the read pattern no longer matches %q — this gate is measuring nothing", read)
		}
	}
	for _, named := range []string{"required: ['claim_count', 'saturation_reached', 'sitting_record_appended'],", "blueEnv2.claim_count", "env.sitting_record_appended === true", "sitting_record_unresolved: sittingRecordUnresolved,"} {
		if !envelopeRelayNamed.MatchString(named) {
			t.Fatalf("the relay pattern no longer matches %q — this gate is measuring nothing", named)
		}
	}
	for _, read := range []string{`if c, ok := r["claim_count"].(float64); ok {`, `if _, present := r["sitting_record_appended"]; present {`} {
		if !envelopeRelayRead.MatchString(read) {
			t.Fatalf("the Go read pattern no longer matches %q — this gate is measuring nothing", read)
		}
	}
	if envelopeRelayRead.MatchString("ClaimCount int `json:\"claim_count\"`") {
		t.Fatal("the Go read pattern matches count-claims' own JSON key — the tool's figure is not a relay")
	}
	if envelopeLogRead.MatchString("log(`epoch ${epoch}: dispatching`)") {
		t.Fatal("the read pattern matches the script's own log() call — it would fail every script")
	}
}
