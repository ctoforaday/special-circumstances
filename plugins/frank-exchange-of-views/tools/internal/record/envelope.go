package record

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

// EnvelopeRefusal is the refusal of the envelope a seat is returning while a blocking item on its own
// work list stands — every seat, by one rule (gblock's ruling, 2026-10-24) — naming each item
// verbatim; "" when none stands. The PreToolUse hook denies the call that returns the envelope with
// it, so the sitting cannot end until each is cleared.
//
// EVERY SEAT WHOSE WORK ARRIVES AS ITEMS RETURNS AN ENVELOPE, so the rule reaches each one: the
// engine dispatches each with an envelope schema, and the harness ends a sitting dispatched with
// one only through the call this refuses. A seat dispatched with none ends with text, and nothing
// here is ever asked. The lens's envelope is LensEnvelope. The seats whose work never arrives as
// items (WorkArrivesAsItems) can hold no item to refuse on.
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

// LensEnvelope is what a lens returns to end its sitting, and it holds nothing. What a lens's
// sitting did — its findings, mints, regrades, closes, judgements, closings and petitions — is on
// the record as events, which every reader reads there, and the engine threads nothing from a lens
// to the seats after it. A field here would be a second account of the record or a value no reader
// reads. The envelope exists so that the sitting ends through the call EnvelopeRefusal answers.
//
// THE ENGINE'S SCHEMA FOR IT IS GENERATED FROM THIS DECLARATION (LensEnvelopeFields), closed, so a
// key it does not declare is refused by the harness and handed back to the lens.
type LensEnvelope struct{}

// LensEnvelopeFields is LensEnvelope's fields, as the walk the plan's schema is generated from
// reads them.
func LensEnvelopeFields() []PlanField { return fieldsOf(reflect.TypeFor[LensEnvelope]()) }

// WorkArrivesAsItems reports whether a seat's work reaches it as items on its work list: every seat
// but the blue seats no plan names — the frontier, a lane and the synthesizer — which are dispatched
// to write the inquiries, a part of the report and the report, and owe nothing that is ever an item.
func WorkArrivesAsItems(role, seatID string) bool { return role != "blue" || seatID == blueRespondSeat }
