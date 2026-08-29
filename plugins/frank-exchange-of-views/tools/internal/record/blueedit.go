package record

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// AnchorIDs returns the distinct finding-marker ids EXPECTED in blue/report.md — the id of
// every `anchor` event on the record, in first-seen order. The blue-report lockdown's
// PostToolUse backstop compares this to the ids actually present to catch a dropped marker.
func AnchorIDs(run Run) ([]string, error) {
	m, err := MergedEvents(run)
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	var out []string
	for i := range m.Events {
		body, ok := recordpb.Body(m.Events[i])
		if !ok {
			// No body at all, so no anchor id. An anchor event that carries nothing cannot
			// name a marker, and the old code reached the same answer by a different route:
			// Payload.Str on an absent key returned "", which the id != "" filter dropped.
			continue
		}
		a, isAnchor := body.(*recordpb.Anchor)
		if !isAnchor {
			continue
		}
		if id := a.GetId(); id != "" && !seen[id] {
			seen[id] = true
			out = append(out, id)
		}
	}
	return out, nil
}

// ExistingBlueEditByKey reports whether this seat already recorded a `blue_edit` event
// under --key. A crash-retried `blue edit` uses it to RECONCILE (re-apply the write
// idempotently) rather than append a second stack op — the event-first ordering keeps the
// stack durable across the crash window between the event append and the report write.
func ExistingBlueEditByKey(run Run, seatID, key string) (bool, error) {
	if key == "" {
		return false, nil
	}
	m, err := MergedEvents(run)
	if err != nil {
		return false, err
	}
	for i := range m.Events {
		e := m.Events[i]
		if e.GetSeatId() != seatID {
			continue
		}
		body, ok := recordpb.Body(e)
		if !ok {
			// No body: nothing can carry an edit_key, so this event cannot be the prior
			// record of THIS edit. Reporting "not found" here is the safe direction — the
			// caller re-applies the write idempotently rather than skipping it.
			continue
		}
		// key is non-empty (guarded above), so GetEditKey()'s zero value for an ABSENT
		// edit_key can never match it. The old `Payload.Str("edit_key") == key` had the
		// same property for the same reason.
		if be, isBlueEdit := body.(*recordpb.BlueEdit); isBlueEdit && be.GetEditKey() == key {
			return true, nil
		}
	}
	return false, nil
}
