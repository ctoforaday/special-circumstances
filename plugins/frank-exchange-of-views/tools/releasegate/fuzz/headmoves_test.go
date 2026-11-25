package fuzz

import (
	"strings"
	"testing"
)

// THE QUEUE ORACLE FAILS EVERY HANDING THAT MOVES THE HEAD OFF A WORKED ITEM STILL QUEUED, and says which
// failure it is: with nothing landed between its reads, no flap; with events between them, which ones. An item queued before the writer and gone after it is an item that left where events
// landed between the reads, and the work list's defect where none did.
func TestTheQueueOracleNamesWhatItFound(t *testing.T) {
	const sitting = "red-lens-evidence#agent#7"
	const worked, next = "reading_debt:sentence:S", "unchecked:anchor:C-2"
	handedNext := "handed\n{\"current\": {\"key\": \"" + next + "\"}}"
	for _, c := range []struct {
		name     string
		after    map[string]bool
		afterID  int64
		wantErr  string // "" for no failure
		wantHead string
	}{
		{"the worked item is still queued and nothing landed", map[string]bool{worked: true, next: true}, 10, "not a flap", ""},
		{"the worked item is still queued and events landed", map[string]bool{worked: true, next: true}, 12, "events landed between the oracle's reads: [11", ""},
		{"the worked item left as events landed", map[string]bool{next: true}, 12, "", next},
		{"the worked item is missing and nothing landed", map[string]bool{next: true}, 10, "the work list's defect", ""},
	} {
		t.Run(c.name, func(t *testing.T) {
			r := &runner{lensHead: map[string]string{sitting: worked}}
			reads := queueReads{
				before:   map[string]bool{worked: true, next: true},
				beforeID: 10, writerID: 10,
				after:    func() (map[string]bool, int64, error) { return c.after, c.afterID, nil },
				delivery: func() string { return `{"key":"` + next + `"}` },
				between: func(a, b int64) string {
					return "[11 EVENT_TYPE_CLOSE red-lens-logic; 12 EVENT_TYPE_MINT red-lens-logic]"
				},
			}
			r.headMovesOnlyOnDischarge("red-lens-evidence", sitting, handedNext, reads)
			switch {
			case c.wantErr == "" && r.actErr != "":
				t.Errorf("the oracle failed the handing: %q", r.actErr)
			case c.wantErr != "" && !strings.Contains(r.actErr, c.wantErr):
				t.Errorf("the oracle's failure is %q, want it to say %q", r.actErr, c.wantErr)
			case c.wantErr != "" && !strings.Contains(r.actErr, `the writer stored as handed: {"key":"`+next+`"}`):
				t.Errorf("the failure does not show what the writer stored: %q", r.actErr)
			case c.wantHead != "" && r.lensHead[sitting] != c.wantHead:
				t.Errorf("the head the oracle holds is %q, want %q", r.lensHead[sitting], c.wantHead)
			}
		})
	}
}
