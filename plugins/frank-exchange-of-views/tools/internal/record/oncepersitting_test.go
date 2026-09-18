package record

import (
	"sort"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// onceAct is one row of the once-per-sitting rule: the seat that files this act and a body of it.
type onceAct struct {
	seat string
	body proto.Message
}

// onceActs IS THE ENUMERATION OF THE RULE, one row per type in `singleton`. The set assertion below
// fails when a type joins that map without a row here, so a new once-per-sitting act cannot ship
// without someone deciding what it looks like across a repair.
var onceActs = map[recordpb.EventType]onceAct{
	recordpb.EventType_EVENT_TYPE_POSITION: {"blue-respond", &recordpb.Position{Text: proto.String("the position")}},
	recordpb.EventType_EVENT_TYPE_REVISION: {"blue-respond", &recordpb.Revision{Text: proto.String("the sitting's edits")}},
	recordpb.EventType_EVENT_TYPE_VERDICT:  {"red-chair", &recordpb.Gate{Verdict: recordtest.P(recordpb.Verdict_VERDICT_PASS)}},
	recordpb.EventType_EVENT_TYPE_SPOT_CHECK: {"red-chair", &recordpb.SpotCheck{Ids: []string{"G1"},
		Reason: proto.String("the anchor still resolves")}},
}

// words names a set of event types in the words a seat reads, sorted, for a comparable message.
func words(types map[recordpb.EventType]bool) []string {
	var out []string
	for typ := range types {
		out = append(out, recordpb.Word(typ))
	}
	sort.Strings(out)
	return out
}

// EVERY ONCE-PER-SITTING ACT IS REFUSED A SECOND TIME IN ONE SITTING, AND A REPAIR IS ONE SITTING
// (#1026). The duty used to be enforced only by deriveKey's ordinal colliding, and that ordinal
// counts TURNS: a sitting-record repair is a turn of its own, so inside one the key was free and
// nothing refused the second act. The duty now asks record.ActClock's question — which sitting is
// this act attributed to — which is what makes the refusal's "this sitting" true.
//
// THE REPAIR ARM IS FORGED FOR THE CHAIR'S TWO ACTS, deliberately. checkRepair admits a repair only
// from a blue seat, so `verdict` and `spot_check` cannot reach that shape through any verb; the row
// holds the WINDOW rather than a reachable run, and the reachable half of their rule is the first
// arm. The end-to-end proof on the shape a run can actually produce is the cli package's
// TestASingletonActIsRefusedASecondTimeInsideARepair.
func TestTheOncePerSittingActsAreRefusedASecondTimeInOneSitting(t *testing.T) {
	covered := map[recordpb.EventType]bool{}
	for typ := range onceActs {
		covered[typ] = true
	}
	if got, want := strings.Join(words(covered), " "), strings.Join(words(singleton), " "); got != want {
		t.Fatalf("the rule covers [%s] and this table exercises [%s] — every once-per-sitting act needs a row", want, got)
	}

	for typ, c := range onceActs {
		t.Run(recordpb.Word(typ), func(t *testing.T) {
			for _, arm := range []struct {
				name    string
				second  func(*stage, string)
				refused bool
			}{
				{"the same turn", func(*stage, string) {}, true},
				{"across a repair", func(b *stage, opened string) { b.repairs(c.seat, "agent-b", opened) }, true},
				{"a new sitting", func(b *stage, _ string) { b.register(c.seat) }, false},
			} {
				t.Run(arm.name, func(t *testing.T) {
					b := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
						registerAs(c.seat, "agent-a")
					opened := b.lastKey()
					b.add(c.seat, c.body)
					arm.second(b, opened)
					db, err := openRunForRead(b.seed())
					if err != nil {
						t.Fatal(err)
					}
					err = requireOncePerSitting(db, c.seat, typ, c.body)
					if !arm.refused {
						if err != nil {
							t.Fatalf("a %s in a NEW sitting was refused: %v", recordpb.Word(typ), err)
						}
						return
					}
					if err == nil {
						t.Fatalf("a second %s attributed to the same sitting was admitted", recordpb.Word(typ))
					}
					want := c.seat + " has already recorded a " + recordpb.Word(typ) + " this sitting"
					if !strings.Contains(err.Error(), want) {
						t.Errorf("the refusal does not say what was wrong:\n%v", err)
					}
				})
			}
		})
	}
}
