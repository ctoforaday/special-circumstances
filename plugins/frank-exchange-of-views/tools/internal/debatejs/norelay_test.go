package debatejs

import (
	"strings"
	"testing"
)

// THE ENGINE RE-PROMPTS NO SEAT FOR WHAT ITS SITTING PUT ON THE RECORD, AND READS NO CLAIM COUNT A
// SEAT RELAYED. The engine reads no record, so an attestation was a seat's word about the record
// and the re-prompt a second agent acting on that word; the claim count was a number the tool
// computes, typed into an envelope. The record answers both: a responding sitting that owes a
// position and has not filed one is blocked by its own work list, capture's record-parity names a
// sitting that closed without one, and the count is counted from the report.
//
// So two drives of the real script — every blue envelope bare, and every blue envelope carrying
// the old keys with the attestation FALSE — dispatch the same seats in the same order, none of
// them a re-prompt of a blue seat, and end the same way.
func TestTheEngineRePromptsNoSeatForItsSittingRecord(t *testing.T) {
	blueParty := map[string]any{"seat_id": "blue-respond", "gap_ids": []any{"G1"}}
	plans := func() []map[string]any { return []map[string]any{planWith([]any{lensParty, blueParty}, []any{})} }

	labels := func(ds []Dispatch) []string {
		var out []string
		for _, d := range ds {
			// The label ends with the run's slug, and each drive has a run directory of its own.
			label := d.Label
			if i := strings.LastIndex(label, " · "); i >= 0 {
				label = label[:i]
			}
			out = append(out, d.SeatID+" | "+label)
		}
		return out
	}
	bare, bareOut := drivePlans(t, plans(), nil)
	stale, staleOut := drivePlans(t, plans(), Envelope{"claim_count": 7, "sitting_record_appended": false})

	if got, want := strings.Join(labels(stale), "\n"), strings.Join(labels(bare), "\n"); got != want {
		t.Errorf("an envelope that does not attest its sitting changed who the engine dispatched:\n--- attesting nothing\n%s\n--- carrying sitting_record_appended:false\n%s", want, got)
	}
	if bareOut != staleOut {
		t.Errorf("the run ended %+v with bare envelopes and %+v with unattested ones", bareOut, staleOut)
	}
	seen := map[string]bool{}
	for _, d := range stale {
		seen[d.SeatID] = true
		if strings.Contains(d.Label, "sitting-record") || strings.Contains(d.Prompt, "Sitting-record repair") {
			t.Errorf("the engine re-prompted a seat for its sitting record: %s", d.Label)
		}
		for _, gone := range []string{"sitting_record_appended", "sitting_record_unresolved", "claim_count", "revision"} {
			if strings.Contains(d.Prompt, gone) {
				t.Errorf("%s: the dispatch names %q", d.Label, gone)
			}
		}
	}
	// THE DRIVE REACHED THE TWO SEATS THAT CARRIED THE ATTESTATION AND THE SEAT HANDED THE LIST.
	for _, seat := range []string{"blue-synthesize", "blue-respond", "judge"} {
		if !seen[seat] {
			t.Errorf("the drive dispatched no %s, so its dispatch was not read", seat)
		}
	}
}
