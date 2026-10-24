package record

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

// EnvelopeRefusal is the refusal of the envelope a seat is returning while a blocking item on its own
// work list stands — every seat, by one rule (gblock's ruling, 2026-10-24) — naming each item
// verbatim; "" when none stands. The PreToolUse hook denies the call that returns the envelope with
// it, so the sitting cannot end until each is cleared.
//
// THE WORK LIST IS THE ONE SOURCE OF WHAT IS BLOCKING: this reads WorkOfSeat, the call `show work`
// makes, and refuses on the items it marks blocking whose clearing is this seat's own act
// (SittingJSON.Owed). Nothing here re-derives a seat's duties, so the list and the refusal cannot
// disagree about what holds a sitting. An item a seat cannot discharge in the sitting it is in is
// therefore never blocking on its list — that is the list's rule, not an exemption here.
//
// THE CHAIR'S BLOCKING ITEMS HOLD A PASS, NOT ITS SITTING (SittingJSON.HoldsPass): its verdict is
// recorded in the one sitting the board permits it, and the PASS gate's blockers are cleared once the
// plan permits a PASS. The verdict gate refuses on them, so its envelope is refused on none.
//
// AN ENVELOPE WHOSE CHECK CANNOT RUN IS REFUSED, and the refusal names why (gblock's standing rule:
// an unreadable record refuses). A missing record is unreadable: the seat asking is bound to a run.
func EnvelopeRefusal(run Run, seatID string) (string, error) {
	role := seatclass.RoleOfSeat(seatID)
	if role == "" {
		return "", fmt.Errorf("%q is no seat this binary knows, so its work list cannot be read", seatID)
	}
	if err := recordPresent(run); err != nil {
		return "", err
	}
	w, err := WorkOfSeat(run, role, seatID)
	if err != nil {
		return "", err
	}
	if w.Sitting.HoldsPass() {
		return "", nil
	}
	owed, _ := w.Sitting.Owed()
	if len(owed) == 0 {
		return "", nil
	}
	named := make([]string, 0, len(owed))
	for i, it := range owed {
		named = append(named, fmt.Sprintf("(%d) %s", i+1, it.What))
	}
	return fmt.Sprintf("your envelope is refused: %d blocking item(s) on your work list stand — %s. "+
		"Your sitting cannot end while any stands: clear each, then return your envelope",
		len(owed), strings.Join(named, "; ")), nil
}

// AgentEnvelopeRefusal is EnvelopeRefusal for the seat the agent is bound to (SeatOfAgent), with that
// seat; "" and no seat for an agent bound to none, which returns no seat's envelope. A run whose
// record is absent cannot say whether the agent is a seat, so it is an error, not "no seat".
func AgentEnvelopeRefusal(run Run, agentID string) (refusal, seatID string, err error) {
	if err := recordPresent(run); err != nil {
		return "", "", err
	}
	seatID, found, err := SeatOfAgent(run, agentID)
	if err != nil || !found {
		return "", "", err
	}
	refusal, err = EnvelopeRefusal(run, seatID)
	return refusal, seatID, err
}

// recordPresent refuses a run whose record database is not there to read.
func recordPresent(run Run) error {
	if _, err := os.Stat(filepath.Join(run.Records(), dbName)); err != nil {
		return fmt.Errorf("the run's record is not readable at %s: %v", run.Records(), err)
	}
	return nil
}
