package fuzz

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/sittingwrite"
)

// subjectlessOwed are the kinds of owed item about the run rather than one thing on it: the sitting a
// seat owes, the chair's verdict, and blue's engage on every successor it found.
var subjectlessOwed = map[record.ItemKind]bool{
	record.ItemRegisterOwed: true, record.ItemVerdictOwed: true, record.ItemSuccessorOwed: true,
}

// queueHolds holds a lens's work list, on whatever board the sweep reached, to its delivery one item
// at a time: every item carries a kind; every item it owes carries a subject (bar the run-scoped
// kinds) and is on its queue (record.QueueOf) under a key no other entry holds; its delivery at
// dispatch fits sittingwrite.DeliveryBudget, in bytes as the writer measures it. The item the writer
// hands it as current, after each of its calls to the record, is held by headMovesOnlyOnDischarge.
func (r *runner) queueHolds(seatID string) {
	w, err := record.WorkOfSeat(r.run(), "lens", seatID)
	if err != nil {
		r.noteActErr("queue oracle: " + seatID + "'s work list does not read: " + err.Error())
		return
	}
	q := record.QueueOf(w)
	keys := map[string]bool{}
	for _, e := range q {
		if keys[e.Key] {
			r.noteActErr(fmt.Sprintf("queue oracle: %s's queue holds key %s twice", seatID, e.Key))
			return
		}
		keys[e.Key] = true
	}
	owed, _ := w.Sitting.Owed()
owed:
	for _, it := range owed {
		if it.Kind == "" || (it.Subject == nil && !subjectlessOwed[it.Kind]) {
			r.noteActErr(fmt.Sprintf("queue oracle: %s owes %q, which carries kind %q and subject %v", seatID, it.What, it.Kind, it.Subject))
			return
		}
		for _, e := range q {
			for _, qi := range e.Items {
				if qi.What == it.What {
					continue owed
				}
			}
		}
		r.noteActErr(fmt.Sprintf("queue oracle: %s owes %q and it is not queued", seatID, it.What))
		return
	}
	if text := sittingwrite.RenderDispatchDelivery(w); len(text) > sittingwrite.DeliveryBudget {
		r.noteActErr(fmt.Sprintf("queue oracle: %s's delivery is %d bytes, over the budget of %d (%s)", seatID, len(text), sittingwrite.DeliveryBudget, sectionBytes(text)))
		return
	}
}

// queueReads is what the queue oracle reads around one call of the writer's (handNext): the lens's
// queue before it runs and the record's last event id before that read, the last event id when the
// writer returns, and — made only where the oracle needs them — the queue after it with the last event
// id after that read, the writer's stored delivery, and the events between two ids. The ids and the
// delivery only say what a failure was; whether it fails reads the two queues.
type queueReads struct {
	before   map[string]bool
	beforeID int64
	writerID int64
	idErr    error
	after    func() (map[string]bool, int64, error)
	delivery func() string
	between  func(a, b int64) string
}

// headMovesOnlyOnDischarge holds the current item the writer named in text, what it handed the lens
// after one of its calls to the record (handNext), to the one handed before while that one is still
// queued. sitting keys what was handed: a new sitting starts from the head of its queue.
//
// STILL QUEUED IS THE QUEUE READ BEFORE THE WRITER AND THE ONE READ AFTER IT: a seat sitting beside this
// one (blue fating an inquiry, a sibling's gap) can take the item off the queue in between, and the
// writer then rightly hands the next one. An item that leaves the queue between the reads with no event
// landed between them is not an item that left: the same record read twice gave two lists, the work
// list's defect, and the run fails on it; a worked item still queued with nothing landed around the
// writer is no flap either. Every failure names the three event ids, the events between
// the oracle's reads, and what the writer stored as handed. A worked item off the queue read before
// the writer is no longer worked: once it left, a demand that queues it again is a new arrival.
func (r *runner) headMovesOnlyOnDischarge(seatID, sitting, text string, reads queueReads) {
	current, handed := "", false
	if i := strings.Index(text, "{"); i >= 0 {
		var d struct {
			Current *struct {
				Key string `json:"key"`
			} `json:"current"`
		}
		if json.Unmarshal([]byte(text[i:]), &d) == nil && d.Current != nil {
			current, handed = d.Current.Key, true
		}
	}
	r.mu.Lock()
	prevHead := r.lensHead[sitting]
	r.mu.Unlock()
	fail := func(what string, afterID int64) {
		landed := "no event landed between the oracle's reads"
		switch {
		case reads.idErr != nil:
			landed = "the record's last event id did not read: " + reads.idErr.Error()
		case afterID != reads.beforeID && reads.between != nil:
			landed = "events landed between the oracle's reads: " + reads.between(reads.beforeID, afterID)
		}
		stored := "not read"
		if reads.delivery != nil {
			stored = reads.delivery()
		}
		r.noteActErr(fmt.Sprintf("queue oracle: %s: %s; event ids: %d before the writer, %d when it returned, %d after the read after it; %s; the writer stored as handed: %s",
			seatID, what, reads.beforeID, reads.writerID, afterID, landed, stored))
	}
	if handed && prevHead != "" && prevHead != current && reads.before[prevHead] {
		after, afterID, err := reads.after()
		quiet := reads.idErr == nil && afterID == reads.beforeID
		switch {
		case err != nil:
			fail(fmt.Sprintf("its work list does not read after the writer handed %s in place of %s, the item it was working: %v", current, prevHead, err), afterID)
			return
		case after[prevHead] && quiet:
			fail(fmt.Sprintf("not a flap: it was handed %s as current while %s, the item it was working, is still queued, and nothing changed the record around the writer", current, prevHead), afterID)
			return
		case after[prevHead]:
			fail(fmt.Sprintf("it was handed %s as current while %s, the item it was working, is still queued", current, prevHead), afterID)
			return
		case quiet:
			fail(fmt.Sprintf("the work list's defect: %s, the item it was working, is queued in the read before the writer and missing from the read after it, and nothing changed the record between them", prevHead), afterID)
			return
		}
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.lensHead == nil {
		r.lensHead = map[string]string{}
	}
	switch {
	case handed:
		r.lensHead[sitting] = current
	case prevHead != "" && !reads.before[prevHead]:
		// THE WORKED ITEM LEFT THE QUEUE, and the writer handed no item in its place: the queue it read
		// held none, or nothing moved. Demanded again, it arrives afresh behind whatever is queued, as
		// the writer hands it (WriteNextItem), so the oracle holds nothing as worked until the next
		// handing names an item.
		delete(r.lensHead, sitting)
	}
}

// sectionBytes is the size of each part of a rendered delivery, for a failure to say what was large.
func sectionBytes(text string) string {
	i := strings.Index(text, "{")
	if i < 0 {
		return "no document"
	}
	var top map[string]json.RawMessage
	if json.Unmarshal([]byte(text[i:]), &top) != nil {
		return "the document does not parse"
	}
	var parts []string
	for k, v := range top {
		parts = append(parts, fmt.Sprintf("%s=%d", k, len(v)))
	}
	var cur map[string]json.RawMessage
	if json.Unmarshal(top["current"], &cur) == nil {
		var gap map[string]json.RawMessage
		if json.Unmarshal(cur["gap"], &gap) == nil {
			for k, v := range gap {
				if len(v) > 200 {
					parts = append(parts, fmt.Sprintf("current.gap.%s=%d", k, len(v)))
				}
			}
		}
	}
	sort.Strings(parts)
	return strings.Join(parts, " ")
}
