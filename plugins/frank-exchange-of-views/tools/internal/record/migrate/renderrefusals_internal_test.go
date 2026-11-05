package migrate

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// A REPORT THAT DOES NOT RENDER AFTER A REFUSAL NAMES THE REFUSAL. A refused edit writes nothing, so
// a later edit quoting the text it wrote finds nothing and the render fails; that failure discards
// the build and its manifest, so the error is all the operator sees. It names how many events were
// refused and the first of them — the cascade's root — beside the render's own failure.
func TestARenderFailureAfterARefusalNamesTheRefusal(t *testing.T) {
	ts := func(s int) string { return "2026-10-01T10:00:0" + string(rune('0'+s)) + ".000000000Z" }
	stream := sliceOf{
		{ID: 1, SeatID: record.HarnessSeat, TS: ts(1), Word: "cast", Fields: map[string]any{},
			Lists: map[string][]string{"seat_ids": {"blue-respond"}}},
		{ID: 2, SeatID: record.HarnessSeat, TS: ts(2), Word: "report_part", Fields: map[string]any{"text": "# T\n\nThe cost is rising over time.\n"}},
		// Refused at the write: it answers a gap no mint created.
		{ID: 3, SeatID: "blue-respond", TS: ts(3), Word: "blue_edit", Fields: map[string]any{
			"old": "rising over time", "new": "climbing sharply", "text": "sharper", "answers": "G-0000dead"}},
		// Written, and quoting the text only the refused edit wrote.
		{ID: 4, SeatID: "blue-respond", TS: ts(4), Word: "blue_edit", Fields: map[string]any{
			"old": "climbing sharply", "new": "rising steadily", "text": "softer"}},
	}
	dst := runtest.New(t, recordtest.TmpRun(t))
	_, err := replayInOrder(stream, Entries(), dst, Options{})
	if err == nil {
		t.Fatal("the replay rendered a report whose last edit quotes text a refused edit never wrote")
	}
	for _, want := range []string{"does not render", "1 event(s) were refused", "source event 3 (blue_edit)", "G-0000dead", "climbing sharply"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the error does not say %q:\n%v", want, err)
		}
	}
}
