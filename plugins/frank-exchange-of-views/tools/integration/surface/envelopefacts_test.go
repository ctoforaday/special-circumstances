package surface

import (
	"os"
	"regexp"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/repotree"
)

// envelopeRestated is a schema property — `name: { … }` or `name: SCHEMA_CONST,` — for a fact the
// record carries and the plan relays: a petition is a motion on the record, and what stands unruled
// is the plan's blockers. An envelope field for either is a second channel for one fact, and the
// engine reading it is a dual reader (#1203).
var envelopeRestated = regexp.MustCompile(`(?m)^\s*(petitions|unruled_motions):\s*(\{|[A-Z_]+,)`)

// NO ENVELOPE RESTATES A FACT THE PLAN RELAYS FROM THE RECORD. The petition channel was an envelope
// field that lenses — whose returns are prose — could never fill, so their petitions reached no
// bench; and the terminal sitting fired on the chair's own count of unruled motions, which nothing
// audited. Both now come from the record through the chair's plan. This is the gate against either
// channel being added back.
func TestNoEnvelopeRestatesAFactThePlanRelays(t *testing.T) {
	path, err := repotree.DebateJS()
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("cannot read the engine script: %v", err)
	}
	for _, m := range envelopeRestated.FindAllStringSubmatch(string(b), -1) {
		t.Errorf("debate.js declares an envelope field %q — the record carries this fact and the chair's plan relays it; a seat's envelope restating it is a second channel the engine must not read", m[1])
	}
	// The pattern must still see a schema property, or a reshaped schema reads as a clean board.
	if !envelopeRestated.MatchString("    petitions: {\n") || !envelopeRestated.MatchString("    unruled_motions: { type: 'integer' },\n") {
		t.Fatal("the pattern no longer matches a schema property — this gate is measuring nothing")
	}
}
