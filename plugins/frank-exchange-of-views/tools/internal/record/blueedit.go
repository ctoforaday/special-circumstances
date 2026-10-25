package record

import "github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"

// AnchorIDs (the distinct `anchor`-event ids EXPECTED in the report) lived here — it fed the
// blue-report lockdown's PostToolUse dropped-marker backstop, which is gone with report-as-record
// (#709). The report places only recorded markers now, so the EXPECTED⊄PRESENT question it answered
// cannot have a non-empty answer.

// ExistingBlueEditByKey is the `blue_edit` this seat already recorded under --key, nil for none. A
// crash-retried `blue edit` answers with it rather than append a second stack op, and is held to
// the change it stores.
func ExistingBlueEditByKey(run Run, seatID, key string) (*recordpb.BlueEdit, error) {
	// key is non-empty (keyedAct returns nil for ""), so a row whose edit_key is NULL never matches.
	ev, err := keyedAct(run, `SELECT b."event_id" FROM "blue_edit" b JOIN "events" e ON e."id" = b."event_id"
	  WHERE e."seat_id" = ? AND b."edit_key" = ? ORDER BY b."event_id" LIMIT 1`, seatID, key)
	return ev.GetBlueEdit(), err
}
