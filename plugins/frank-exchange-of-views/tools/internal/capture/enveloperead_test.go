package capture

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// EVERY READ THE ENGINE ACTED ON IS RECOMPUTED AT THE EVENT IT NAMES. The one thing the engine cannot
// tell, a value of the right shape nothing wrote, is what this finds: a read that differs from the
// record there, a read naming an event the record does not hold, and an envelope carrying none. No
// journal is not a pass, and neither is a journal holding no envelope.
func TestEveryReadInTheJournalIsRecomputedAtItsEvent(t *testing.T) {
	run, _ := storedRun(t, 1, true)
	read, err := record.ReadRecord(run)
	if err != nil {
		t.Fatal(err)
	}
	if read.Event == 0 || len(read.BenchOwed) != 1 || read.BenchOwed[0] != "petition" {
		t.Fatalf("fixture: the read = %+v, want a petition owed so a false read has something to differ from", read)
	}
	carrying := func(t *testing.T, r record.RecordRead) map[string]any {
		t.Helper()
		b, err := json.Marshal(r)
		if err != nil {
			t.Fatal(err)
		}
		var doc map[string]any
		if err := json.Unmarshal(b, &doc); err != nil {
			t.Fatal(err)
		}
		return map[string]any{"record_read": doc}
	}
	early := read
	early.Event = read.Event - 1
	early.BenchOwed = []record.BenchHearing{}
	notOwed, halted, outcome, elsewhere, placeholder := read, read, read, read, read
	notOwed.BenchOwed = []record.BenchHearing{}
	halted.Halted = true
	outcome.Outcome = "VERIFIED"
	elsewhere.Event = read.Event + 1000
	placeholder.Event = 0
	placeholder.BenchOwed = []record.BenchHearing{}

	for _, tc := range []struct {
		name    string
		results []map[string]any
		verdict string
		detail  string
	}{
		{"the read as it stands", []map[string]any{carrying(t, read)}, "PASS", "1 read(s)"},
		{"owed dropped", []map[string]any{carrying(t, notOwed)}, "FAIL", "the record there says bench owed [petition]"},
		{"a halt nothing recorded", []map[string]any{carrying(t, halted)}, "FAIL", "halted true"},
		{"an outcome nothing stamped", []map[string]any{carrying(t, outcome)}, "FAIL", "outcome VERIFIED"},
		{"an event the record does not hold", []map[string]any{carrying(t, elsewhere)}, "FAIL", "names event"},
		{"the placeholder a seat types", []map[string]any{carrying(t, placeholder)}, "FAIL", "names event 0"},
		{"an envelope with none", []map[string]any{{"notes": "typed"}}, "FAIL", "envelope 1 of the journal carries no record_read"},
		{"a value that is not a read", []map[string]any{{"record_read": "Not halted"}}, "FAIL", "does not read as one"},
		{"no envelope", nil, "SKIP", "NOT MEASURED"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a := EnvelopeReadAudit(run, tc.results, true)
			if a.Verdict != tc.verdict || !strings.Contains(a.Detail, tc.detail) {
				t.Errorf("= %s: %s; want %s naming %q", a.Verdict, a.Detail, tc.verdict, tc.detail)
			}
		})
	}
	if a := EnvelopeReadAudit(run, nil, false); a.Verdict != "SKIP" || !strings.Contains(a.Detail, "no workflow journal") {
		t.Errorf("no journal = %s: %s", a.Verdict, a.Detail)
	}
}
