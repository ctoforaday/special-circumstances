package record

import (
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// A RETIRE IS FIRST AND PERMANENT FOR EVERY READER. The live tool can neither reopen a retired anchor
// (an edit invents no anchor) nor retire one twice (the retire reads it off the report), and no
// archive holds either; a record that did is read the one way by the Go ledger, the evidence table
// and the SQL views: C-0000000a retired then named by an edit's `reopened` stays retired, and
// C-0000000b retired twice is retired by its first retire.
func TestARetireIsFirstAndPermanentForEveryReader(t *testing.T) {
	const reopened, twice = "C-0000000a", "C-0000000b"
	run := newStage(t).cite(reopened).cite(twice).
		retire(reopened).
		add("blue-respond", &recordpb.BlueEdit{Old: proto.String("the claim"), New: proto.String("the claim, reworded"),
			Text: proto.String("reworded"), Reopened: []string{reopened}}).
		retire(twice).retire(twice).seed()
	fam, err := MergedEvents(run)
	if err != nil {
		t.Fatal(err)
	}
	first := map[string]int64{}
	for _, e := range fam.Events {
		if r, ok := recordpb.BodyAs[*recordpb.Retire](e); ok {
			for _, a := range r.GetAnchors() {
				if _, seen := first[a]; !seen {
					first[a] = fam.At.Of(e).ID
				}
			}
		}
	}
	retired := RetiredAnchors(fam.Events, fam.At)
	ev := EvidenceJSONOf(fam.Events, fam.At)
	db, err := openRunForRead(run)
	if err != nil {
		t.Fatal(err)
	}
	for _, a := range []string{reopened, twice} {
		if !retired[a] {
			t.Errorf("the anchor ledger reads %s live", a)
		}
		var at int64
		for _, r := range ev.Retired {
			if r.Anchor() == a {
				at = r.Retired.Event
			}
		}
		if at != first[a] {
			t.Errorf("the evidence table's retire of %s is event %d, want its first, %d", a, at, first[a])
		}
		var view int64
		if err := db.QueryRow(`SELECT COALESCE("retired_event", 0) FROM "evidence_anchor" WHERE "anchor_id" = ?`, a).Scan(&view); err != nil || view != first[a] {
			t.Errorf("evidence_anchor's retire of %s = %d (%v), want %d", a, view, err, first[a])
		}
		if r, found, err := AnchorRetiredAt(run, a); err != nil || !found || r.Event != first[a] {
			t.Errorf("AnchorRetiredAt(%s) = %+v, %v, %v, want event %d", a, r, found, err, first[a])
		}
	}
}
