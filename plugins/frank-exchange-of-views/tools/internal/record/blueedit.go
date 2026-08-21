package record

// AnchorIDs returns the distinct finding-marker ids EXPECTED in blue/report.md — the id of
// every `anchor` event on the record, in first-seen order. The blue-report lockdown's
// PostToolUse backstop compares this to the ids actually present to catch a dropped marker.
func AnchorIDs(runDir string) ([]string, error) {
	m, err := MergedEvents(runDir)
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	var out []string
	for _, e := range m.Events {
		if e.Type == "anchor" {
			if id := e.Payload.Str("id"); id != "" && !seen[id] {
				seen[id] = true
				out = append(out, id)
			}
		}
	}
	return out, nil
}

// ExistingBlueEditByKey reports whether this seat already recorded a `blue_edit` event
// under --key. A crash-retried `blue edit` uses it to RECONCILE (re-apply the write
// idempotently) rather than append a second stack op — the event-first ordering keeps the
// stack durable across the crash window between the event append and the report write.
func ExistingBlueEditByKey(runDir, seatID, key string) (bool, error) {
	_, found, err := ExistingByKey(runDir, seatID, "blue_edit", key)
	return found, err
}
