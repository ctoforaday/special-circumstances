package fuzz

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
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
// dispatch fits sittingwrite.DeliveryBudget; and the item the writer's next phase hands it as current
// moves only when the one before it has left the queue.
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
	if n := len(sittingwrite.RenderLensDelivery(w)); n > sittingwrite.DeliveryBudget {
		r.noteActErr(fmt.Sprintf("queue oracle: %s's delivery is %d bytes, over the budget of %d", seatID, n, sittingwrite.DeliveryBudget))
		return
	}
	r.headMovesOnlyOnDischarge(seatID, fmt.Sprintf("%s#%d", seatID, w.SittingID()), keys)
}

// headMovesOnlyOnDischarge asks the writer what the lens is handed next, as the PostToolUse hook does,
// and holds the current item it names to the one handed before while that one is still queued.
// sitting keys what was handed: a new sitting starts from the head of its queue.
func (r *runner) headMovesOnlyOnDischarge(seatID, sitting string, queued map[string]bool) {
	if r.writer == "" {
		return
	}
	agentType := seatclass.Seats[seatID].AgentType
	out, err := exec.Command(r.writer, "-run", r.runDir, "-phase", string(sittingwrite.Next),
		"-agent-id", "fuzz-queue-"+seatID, "-agent-type", agentType).Output()
	if err != nil {
		r.noteActErr("queue oracle: the writer's next phase failed for " + seatID + ": " + err.Error())
		return
	}
	text := string(out)
	if strings.Contains(text, "COULD NOT BE READ") {
		r.noteActErr("queue oracle: " + seatID + "'s next item could not be read: " + firstLine(text))
		return
	}
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
	defer r.mu.Unlock()
	if r.lensHead == nil {
		r.lensHead = map[string]string{}
	}
	prev, had := r.lensHead[sitting]
	if handed {
		if had && prev != "" && queued[prev] && current != prev {
			if r.actErr == "" {
				r.actErr = fmt.Sprintf("queue oracle: %s was handed %s as current while %s, the item it was working, is still queued", seatID, current, prev)
			}
			return
		}
		r.lensHead[sitting] = current
	}
}
