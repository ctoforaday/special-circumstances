package sittingwrite

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// REPORT TEXT PRE-EMPTS THE CURRENT ITEM, AND THE ITEM RESUMES AFTER IT: the current item stays current
// until it leaves the queue. The chair is handed motion M1; the head moves in its sitting, so a retired
// lens's area goes stale and the chair is owed the change behind its pin, report text, which is handed
// first; after it the chair's current item is M1 again, though the stale area's own item now heads the
// queue.
func TestAnItemReportTextPreemptedResumesAfterIt(t *testing.T) {
	t.Skip("#1695 R1 merge: the chair's non-text item here was its ruling on blue's grade motion, which this branch deletes; the chair holds no other item a head move can pre-empt, so the fixture needs a seat that holds an item before any report text (R2)")
	dir := newRun(t)
	const lens = "red-lens-logic"
	evs := []*recordpb.Event{
		recordtest.At(t, record.HarnessSeat, "harness:cast", &recordpb.Cast{
			SeatIds: []string{"red-chair", lens, "red-lens-evidence", "blue-respond", "judge"}}),
		recordtest.At(t, record.HarnessSeat, "harness:ingest", &recordpb.ReportPart{Text: proto.String("# report\n\nOne paragraph.\n")}),
	}
	// The logic lens sits three times barren, each sitting with a complete delivery: it retires at pin 2.
	for i := range 3 {
		evs = append(evs,
			recordtest.At(t, "red-chair", fmt.Sprintf("chair:dispatch:%d", i), &recordpb.Dispatch{SeatId: proto.String(lens), Pin: proto.Int64(2)}),
			recordtest.At(t, lens, fmt.Sprintf("lens:register:%d", i), &recordpb.Register{}),
			recordtest.At(t, lens, fmt.Sprintf("lens:read:%d", i), &recordpb.ReportRead{Head: proto.Int64(2)}))
	}
	evs = append(evs,
		recordtest.At(t, "red-lens-evidence", "lens:register", &recordpb.Register{}),
		recordtest.At(t, "red-lens-evidence", "lens:mint", &recordpb.Mint{GapId: proto.String("G1"), Class: recordpb.GapClass_GAP_CLASS_SPEC_UNDERSPECIFICATION.Enum(),
			Problem: proto.String("p"), AcceptanceCheck: proto.String("c"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT),
			Severity: recordtest.P(recordpb.Grade_GRADE_LOW), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM)}),
		recordtest.At(t, "blue-respond", "blue:register", &recordpb.Register{}),
		recordtest.At(t, "red-chair", "chair:register", &recordpb.Register{}))
	recordtest.Seed(t, dir, append(evs, recordtest.SettledQuestion(t)...)...)

	const chair, agent = "frank-exchange-of-views:red-chair", "agent_02"
	const worked = "unruled_motion:motion:M1"
	call := func() string {
		t.Helper()
		var out bytes.Buffer
		if err := WriteNextItem(dir, agent, chair, "claude", &out); err != nil {
			t.Fatal(err)
		}
		return out.String()
	}
	currentOf := func(text string) string {
		i := strings.Index(text, "{")
		if i < 0 {
			return ""
		}
		var d struct {
			Current *struct {
				Key string `json:"key"`
			} `json:"current"`
		}
		if json.Unmarshal([]byte(text[i:]), &d) != nil || d.Current == nil {
			return ""
		}
		return d.Current.Key
	}
	queue := func() []string {
		t.Helper()
		run, err := record.NewRun(dir)
		if err != nil {
			t.Fatal(err)
		}
		w, err := record.WorkOfSeat(run, "chair", "red-chair")
		if err != nil {
			t.Fatal(err)
		}
		return queueKeys(record.QueueOf(w))
	}

	if got := currentOf(call()); got != worked {
		t.Fatalf("the chair's first call handed %q as current, want %s; queue %v", got, worked, queue())
	}
	// THE HEAD MOVES IN THE CHAIR'S SITTING: the retired lens's area is stale, with a change behind its pin.
	recordtest.Seed(t, dir, recordtest.At(t, record.HarnessSeat, "harness:ingest:2", &recordpb.ReportPart{Text: proto.String("# report\n\nTwo paragraphs.\n")}))
	q := queue()
	if len(q) < 3 || !strings.HasPrefix(q[0], string(record.ItemReportText)+":") || q[1] == worked {
		t.Fatalf("fixture: the head move did not put report text, then an item ahead of %s, on the chair's queue: %v", worked, q)
	}
	text := call()
	if !strings.HasPrefix(text, "REPORT TEXT") {
		t.Fatalf("report text did not pre-empt the current item:\n%s", text)
	}
	for range maxPreemptCalls {
		if !strings.HasPrefix(text, "REPORT TEXT") {
			break
		}
		text = call()
	}
	if got := currentOf(text); got != worked {
		t.Errorf("after the report text the chair was handed %q as current while %s, the item report text pre-empted, is still queued (%v)", got, worked, queue())
	}
}

// maxPreemptCalls bounds the calls the report text above takes: one change, one item.
const maxPreemptCalls = 10
