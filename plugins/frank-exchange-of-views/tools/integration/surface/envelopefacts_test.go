package surface

import (
	"os"
	"regexp"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/repotree"
)

// envelopeRestated is a schema property — `name: { … }` or `name: SCHEMA_CONST,` — for a fact the
// record carries: a petition is a motion on the record, what stands unruled is the plan's blockers,
// and a seat's log entry is a log event. An envelope field for any of them is a second channel for
// one fact, and the engine reading it is a dual reader (#1203, #1267).
var envelopeRestated = regexp.MustCompile(`(?m)^\s*(petitions|unruled_motions|log):\s*(\{|[A-Z_]+,)`)

// envelopeLogRead is the engine indexing a `log` member of anything: `env.log`, `env['log']`.
// The script's own `log(…)` call is a function, never a member, so the script holds no such
// access at all and any one of them is a returned envelope's log being read.
var envelopeLogRead = regexp.MustCompile("\\.log\\b|\\[\\s*['\"`]log['\"`]\\s*\\]")

// NO ENVELOPE RESTATES A FACT THE RECORD CARRIES. The petition channel was an envelope field that
// lenses — whose returns are prose — could never fill, so their petitions reached no bench; the
// terminal sitting fired on the chair's own count of unruled motions, which nothing audited; and
// the envelope's log was a copy of entries every seat writes to the record, which nine of thirteen
// seat types could not return. Petitions and unruled motions come from the record through the
// chair's plan; the log is read from the record and relayed by nobody. This is the gate against
// any of the three channels being added back, as a schema property or as a read.
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
	// The patterns must still see a schema property and a read, or a reshaped script reads as a
	// clean board.
	for _, line := range []string{"    petitions: {\n", "    unruled_motions: { type: 'integer' },\n", "    log: { type: 'array', items: { type: 'string' } },\n"} {
		if !envelopeRestated.MatchString(line) {
			t.Fatalf("the schema pattern no longer matches %q — this gate is measuring nothing", line)
		}
	}
	for _, read := range []string{"if (env && env.log) for (const f of env.log)", "env['log']", `env["log"]`} {
		if !envelopeLogRead.MatchString(read) {
			t.Fatalf("the read pattern no longer matches %q — this gate is measuring nothing", read)
		}
	}
	if envelopeLogRead.MatchString("log(`epoch ${epoch}: dispatching`)") {
		t.Fatal("the read pattern matches the script's own log() call — it would fail every script")
	}
}
