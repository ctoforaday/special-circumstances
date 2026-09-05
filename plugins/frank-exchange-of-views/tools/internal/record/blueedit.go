package record

// AnchorIDs (the distinct `anchor`-event ids EXPECTED in the report) lived here — it fed the
// blue-report lockdown's PostToolUse dropped-marker backstop, which is gone with report-as-record
// (#709). The report places only recorded markers now, so the EXPECTED⊄PRESENT question it answered
// cannot have a non-empty answer.

// ExistingBlueEditByKey reports whether this seat already recorded a `blue_edit` event
// under --key. A crash-retried `blue edit` uses it to RECONCILE (re-apply the write
// idempotently) rather than append a second stack op — the event-first ordering keeps the
// stack durable across the crash window between the event append and the report write.
func ExistingBlueEditByKey(run Run, seatID, key string) (bool, error) {
	if key == "" {
		return false, nil
	}
	// key is non-empty (guarded above), so a row whose edit_key is NULL can never match it —
	// the same property the fold's absent-key zero value had, held by the comparison itself.
	return recordHas(run, `SELECT 1 FROM "blue_edit" b JOIN "events" e ON e."id" = b."event_id"
	  WHERE e."seat_id" = ? AND b."edit_key" = ? LIMIT 1`, seatID, key)
}
