package migrate_test

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migrate"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// A COMPUTATION WHOSE METHOD CITED A LENS-PLACED LABEL MIGRATES CITING NOTHING. A lens's labelled
// reading of a source it found placed a citation; the label is dropped and its Anchor written as no
// event (EventSchema 59), so a computation whose --cites named it would be refused at the write —
// and with it every reproduce, mint and plan naming the computation, until a later edit quotes text
// a refused one never wrote and the report does not render. The computation comes forward whole,
// its cites empty.
func TestAComputationCitingADroppedLabelCitesNothing(t *testing.T) {
	const label, sentence = "C-e6aeeeda", "The cost is rising over time."
	var o oldStream
	o.add(record.HarnessSeat, "cast", "", map[string]any{})
	o.evs[len(o.evs)-1].Lists = map[string][]string{"seat_ids": {"red-lens-voice", "blue-respond"}}
	o.add(record.HarnessSeat, "report_part", "", map[string]any{"text": "# T\n\n" + sentence + "\n"})
	o.add("red-lens-voice", "verify", "", map[string]any{"claim": sentence, "label": label,
		"url": "https://example.invalid/found", "title": "Found", "outcome": "supports", "confidence": "high",
		"text": "it says so", "access_date": "2026-10-01"})
	o.add("red-lens-voice", "anchor", "", map[string]any{"id": label, "location": sentence})
	o.add("blue-respond", "computation", "", map[string]any{"computation_id": "K-4d27a6ce", "bundle_sha": "s1",
		"basis": "reproducible", "cites": label})
	o.evs[len(o.evs)-1].Lists = map[string][]string{"argv": {"python3", "script.py"}}

	res, dst := replaySlice(t, o.evs, migrate.Options{})
	if len(res.Refusals) != 0 {
		t.Fatalf("refusals: %+v — a computation citing a dropped label is refused, and every act on it with it", res.Refusals)
	}
	m, err := record.MergedEvents(dst)
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	for _, e := range m.Events {
		if c, ok := recordpb.BodyAs[*recordpb.Computation](e); ok {
			n++
			if c.Cites != nil {
				t.Errorf("computation %s cites %q; the label it cited is dropped, so it cites nothing", c.GetComputationId(), c.GetCites())
			}
		}
		if a, ok := recordpb.BodyAs[*recordpb.Anchor](e); ok && a.GetId() == label {
			t.Errorf("the migrated record holds the dropped label's anchor %s", label)
		}
	}
	if n != 1 {
		t.Errorf("%d computation(s) written, want 1", n)
	}
}
