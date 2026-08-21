package record

import "fmt"

// A RETRY UNDER THE SAME KEY RETURNS THE FIRST RESULT INSTEAD OF ACTING TWICE — once, here.
//
// # Five copies of one scan
//
// `--key` is the seat's crash-retry handle: a dispatch that dies mid-call is re-run, and without
// the handle the second run mints a second gap, splices a second anchor, records a second finding.
// Five verbs offer it, and each carried its own scan of the event log — byte-identical apart from
// three values, and one of them returning a bool where the others returned a string.
//
// Measured 8/50 duplicate dispatches in run 5, all crash re-runs. So this path runs, and a sixth
// keyed verb copying one of the five would be the sixth copy of a scan nobody had tested as a
// unit — the same shape as the prose channel, where three verbs of fifty-one got the convention
// wrong and no gate could see it.
//
// # The spec is data
//
// A keyed verb differs in what it writes, what payload field carries the handle, and what it hands
// back to the seat. Those are three values, so they are a table rather than three functions.
type keySpec struct {
	// keyField is where the handle lives in the payload. Four of five spell it `<verb>_key`; the
	// exception is blue_edit, whose event type and verb name differ.
	keyField string
	// resultField is what the seat gets back on a repeat — the id it would otherwise have been
	// assigned. Empty for a verb with nothing to return, where the answer is only "already done".
	resultField string
}

// keyedEvents is every event type that honours --key. A verb whose type is absent here does not
// offer the handle, and asking for it is an error rather than a silent "not found" — an
// idempotency check that answers "no prior" for a type it does not know is a check that lets the
// retry act twice.
var keyedEvents = map[string]keySpec{
	"mint":      {keyField: "mint_key", resultField: "gap_id"},
	"cite":      {keyField: "cite_key", resultField: "label"},
	"proof":     {keyField: "proof_key", resultField: "sha256"},
	"finding":   {keyField: "finding_key", resultField: "label"},
	"blue_edit": {keyField: "edit_key"},
}

// ExistingByKey answers whether this seat already wrote an event of this type under this handle,
// and what it was told at the time.
//
// An empty key is not a match: the handle is optional, and a seat that passes none is asking to
// act, not to retry.
func ExistingByKey(runDir, seatID, eventType, key string) (result string, found bool, err error) {
	spec, ok := keyedEvents[eventType]
	if !ok {
		return "", false, fmt.Errorf("record: %q does not honour --key, so an idempotency check on it "+
			"would answer \"no prior\" for every retry and let the act happen twice", eventType)
	}
	if key == "" {
		return "", false, nil
	}
	m, err := MergedEvents(runDir)
	if err != nil {
		return "", false, err
	}
	for _, e := range m.Events {
		if e.Type == eventType && e.SeatID == seatID && e.Payload.Str(spec.keyField) == key {
			if spec.resultField == "" {
				return "", true, nil
			}
			return e.Payload.Str(spec.resultField), true, nil
		}
	}
	return "", false, nil
}
