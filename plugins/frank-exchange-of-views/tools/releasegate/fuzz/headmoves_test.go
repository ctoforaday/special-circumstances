package fuzz

import (
	"strings"
	"testing"
)

// THE QUEUE ORACLE HOLDS THE WRITER TO ITS OWN READ. The writer hands the head of the queue when the
// item it handed before is gone from the read it hands from, and stores that read beside the item: an
// item that left between the oracle's reads and came back is the writer's contract kept, and an item the
// writer's read held is the head moved with the item queued, whatever the oracle's own reads saw.
func TestTheQueueOracleHoldsTheWriterToItsOwnRead(t *testing.T) {
	const sitting = "red-lens-evidence#agent#7"
	handedNext := "handed\n{\"current\": {\"key\": \"unchecked:anchor:C-2\"}}"
	for _, c := range []struct {
		name    string
		writer  *handedRead
		wantErr bool
	}{
		{"the worked item left the writer's read", &handedRead{key: "unchecked:anchor:C-2", queue: []string{"unchecked:anchor:C-2"}}, false},
		{"the worked item was in the writer's read", &handedRead{key: "unchecked:anchor:C-2", queue: []string{"unchecked:anchor:C-2", "reading_debt:sentence:S"}}, true},
	} {
		t.Run(c.name, func(t *testing.T) {
			r := &runner{lensHead: map[string]string{sitting: "reading_debt:sentence:S"}}
			// The oracle's own reads before and after the writer both hold the worked item.
			queued := map[string]bool{"reading_debt:sentence:S": true, "unchecked:anchor:C-2": true}
			r.headMovesOnlyOnDischarge("red-lens-evidence", sitting, queued, handedNext, c.writer)
			if got := r.actErr != ""; got != c.wantErr {
				t.Errorf("oracle failed the handing = %v (%q), want %v", got, r.actErr, c.wantErr)
			}
			if c.wantErr && !strings.Contains(r.actErr, "the item it was working, is still queued") {
				t.Errorf("the failure does not name the worked item still queued: %q", r.actErr)
			}
			if !c.wantErr && r.lensHead[sitting] != "unchecked:anchor:C-2" {
				t.Errorf("the head the oracle holds is %q, want the item handed", r.lensHead[sitting])
			}
		})
	}
}
