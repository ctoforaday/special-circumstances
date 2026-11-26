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

// AN ITEM THAT LEFT THE QUEUE AND IS DEMANDED AGAIN ARRIVES AFRESH, and the oracle holds the head the
// writer holds. Release fuzz seed 27 (#1943): a lens's regrade discharged its one item, the writer told
// it the item left and nothing more was queued (prose, no current item), and blue's later edit queued
// the item again behind a fresh citation. The writer handed the citation, the head of the queue; the
// oracle still held the discharged item as worked and failed the handing as no flap.
func TestTheQueueOracleForgetsAnItemThatLeftTheQueue(t *testing.T) {
	const sitting = "red-lens-evidence#agent#168"
	const worked, cited = "reaudit:gap:G-1", "unchecked:anchor:C-1"
	r := &runner{}
	quiet := func(before map[string]bool) queueReads {
		return queueReads{
			before: before, beforeID: 10, writerID: 10,
			after:    func() (map[string]bool, int64, error) { return before, 10, nil },
			delivery: func() string { return "stored" },
			between:  func(a, b int64) string { return "[]" },
		}
	}
	r.headMovesOnlyOnDischarge("red-lens-evidence", sitting, "handed\n{\"current\": {\"key\": \""+worked+"\"}}", quiet(map[string]bool{worked: true}))
	r.headMovesOnlyOnDischarge("red-lens-evidence", sitting, worked+" left your list, and nothing more is queued for you: whether anything blocks you finishing is `complete` on your work list (now true).\n", quiet(map[string]bool{}))
	r.headMovesOnlyOnDischarge("red-lens-evidence", sitting, "handed\n{\"current\": {\"key\": \""+cited+"\"}}", quiet(map[string]bool{cited: true, worked: true}))
	if r.actErr != "" {
		t.Errorf("the oracle failed the head of a queue its worked item had left: %q", r.actErr)
	}
	if r.lensHead[sitting] != cited {
		t.Errorf("the head the oracle holds is %q, want %q", r.lensHead[sitting], cited)
	}
}

// AN ITEM STILL QUEUED STAYS THE HEAD THROUGH A HANDING THAT NAMES NONE: a line saying another key left
// moves nothing, and the writer moving off it after is failed.
func TestTheQueueOracleHoldsAnItemThatNeverLeft(t *testing.T) {
	const sitting = "red-lens-evidence#agent#168"
	const worked, cited = "reaudit:gap:G-1", "unchecked:anchor:C-1"
	r := &runner{}
	queued := map[string]bool{worked: true, cited: true}
	reads := queueReads{
		before: queued, beforeID: 10, writerID: 10,
		after:    func() (map[string]bool, int64, error) { return queued, 10, nil },
		delivery: func() string { return "stored" },
		between:  func(a, b int64) string { return "[]" },
	}
	r.headMovesOnlyOnDischarge("red-lens-evidence", sitting, "handed\n{\"current\": {\"key\": \""+worked+"\"}}", reads)
	r.headMovesOnlyOnDischarge("red-lens-evidence", sitting, "[\"reaudit:gap:G-2\"] left your list; your current item is still "+worked+".\n", reads)
	r.headMovesOnlyOnDischarge("red-lens-evidence", sitting, "handed\n{\"current\": {\"key\": \""+cited+"\"}}", reads)
	if !strings.Contains(r.actErr, "not a flap") {
		t.Errorf("the oracle let the head move off a worked item that never left: %q", r.actErr)
	}
}
