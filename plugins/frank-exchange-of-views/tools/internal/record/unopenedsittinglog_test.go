package record

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// toolRefusal is the entry the tool writes for a call it refused the seat.
func (b *stage) toolRefusal(seat string) *stage {
	return b.add(seat, &recordpb.Log{Text: proto.String("refused `edit` with --quote: the quoted span was not found"),
		Type: recordpb.LogType_LOG_TYPE_REFUSAL.Enum(), Source: recordpb.LogSource_LOG_SOURCE_TOOL.Enum()})
}

// logItems is the log items of a list.
func logItems(s SittingJSON) []Item {
	var out []Item
	for _, it := range s.Open {
		if it.Kind == ItemLog {
			out = append(out, it)
		}
	}
	return out
}

// refusedCallItems is the log items of a list that count refused or failed calls.
func refusedCallItems(s SittingJSON) []Item {
	var out []Item
	for _, it := range logItems(s) {
		if strings.Contains(it.What, "refused or failed call(s)") {
			out = append(out, it)
		}
	}
	return out
}

// A SITTING THAT HAS NOT OPENED HOLDS NO REFUSED CALL. blue-respond's agent type seats more than
// one seat, so the hook's bracket names none and its sitting opens at its register. Dispatched
// again and reading its list before that register, it is told the sitting is not on the record —
// and the log item on the same list counts no call of the sitting before as this one's. On the
// 2026-11-10 development run three sittings read "1 refused or failed call(s) of yours this
// sitting" before their first call, and one filed an entry explaining a refusal it never met.
func TestASittingNotYetOpenedCountsNoRefusedCallOfTheOneBefore(t *testing.T) {
	const blue = "blue-respond"
	dispatched := func() *stage {
		b := newStage(t).cast(evLens, "red-chair", blue, "judge").ingest().
			register("red-chair").dispatch(2, blue).
			bracket("", "blue-a").registerAs(blue, "blue-a").toolRefusal(blue).stop("blue-a").
			register("red-chair").dispatch(2, blue).
			bracket("", "blue-b")
		return b
	}

	before := sittingOfRunT(t, dispatched().seed(), "blue", blue)
	if len(owedItems(before)) != 1 {
		t.Fatalf("the fixture is not a dispatched sitting awaiting its register: open = %+v", before.Open)
	}
	if got := refusedCallItems(before); len(got) != 0 {
		t.Errorf("before its register the list counts a refused call of the sitting before as this one's: %q", got[0].What)
	}
	// The sitting before recorded no act, and its emptiness is not this sitting's: the log stays open.
	if got := logItems(before); len(got) != 1 || !strings.HasPrefix(got[0].What, "the log is open — ") {
		t.Errorf("before its register the log item is not the open channel: %+v", got)
	}

	registered := sittingOfRunT(t, dispatched().registerAs(blue, "blue-b").seed(), "blue", blue)
	if got := refusedCallItems(registered); len(got) != 0 {
		t.Errorf("a registered sitting with no refused call counts one: %q", got[0].What)
	}

	refused := sittingOfRunT(t, dispatched().registerAs(blue, "blue-b").toolRefusal(blue).seed(), "blue", blue)
	if got := refusedCallItems(refused); len(got) != 1 || !strings.Contains(got[0].What, "recorded 1 refused or failed call(s)") {
		t.Errorf("a sitting with one refused call of its own: log items = %+v, want one counting 1", got)
	}
}

// THE CHAIR'S SITTING OPENS AT ITS REGISTER TOO when the workflow comes back to it, and the list it
// reads before that register counts no refused call of its sitting before.
func TestAChairSittingNotYetOpenedCountsNoRefusedCallOfTheOneBefore(t *testing.T) {
	const blue = "blue-respond"
	returned := func() *stage {
		return newStage(t).cast(evLens, "red-chair", blue, "judge").ingest().
			register("red-chair").toolRefusal("red-chair").dispatch(2, blue).
			bracket("", "blue-a").registerAs(blue, "blue-a").stop("blue-a")
	}

	before := sittingOfRunT(t, returned().seed(), "chair", "red-chair")
	if len(owedItems(before)) != 1 {
		t.Fatalf("the fixture is not a chair sitting awaiting its register: open = %+v", before.Open)
	}
	if got := refusedCallItems(before); len(got) != 0 {
		t.Errorf("before its register the chair's list counts a refused call of the sitting before as this one's: %q", got[0].What)
	}

	registered := sittingOfRunT(t, returned().register("red-chair").toolRefusal("red-chair").seed(), "chair", "red-chair")
	if got := refusedCallItems(registered); len(got) != 1 || !strings.Contains(got[0].What, "recorded 1 refused or failed call(s)") {
		t.Errorf("a chair sitting with one refused call of its own: log items = %+v, want one counting 1", got)
	}
}
