package record

import (
	"slices"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// A FINDING SAYS WHETHER IT BECAME ANYTHING (#747).
//
// `found_by` answers "which findings became this gap" from the GAP. Nothing answered "did this
// finding become anything" from the FINDING, so a finding that produced no gap looked exactly
// like one that did unless a reader cross-indexed every gap row by hand.
//
// Measured on 2026-08-23_research-loop-counterparts: 20 findings, 16 gaps, and three findings —
// L2-F1, L6-F8, L6-F11 — that no gap credits. One alleged a fabricated verbatim quote in the very
// text the merge closed a gap on in that same sitting; nothing on the docket said so, because the
// docket is gap-shaped and that fact is finding-shaped.
//
// This is the same defect the `Anchor` field was, one join over: a key living where nothing could
// reach it rather than where it was written.
func TestAFindingSaysWhichGapsCreditIt(t *testing.T) {
	find := func(label string) *Event {
		return recordtest.Event(t, "red-lens-r1-L5", 1, &recordpb.Finding{
			FindingId: proto.String("f-" + label), Label: proto.String(label),
			Text: proto.String("a finding"), Severity: recordtest.P(recordpb.Grade_GRADE_MEDIUM),
		})
	}
	mint := func(gap string, foundBy ...string) *Event {
		return recordtest.Event(t, "red-merge-r1", 1, &recordpb.Mint{
			GapId: proto.String(gap), Class: proto.String("self-attestation"),
			Problem: proto.String("p"), RequiredFix: proto.String("f"), AcceptanceCheck: proto.String("a"),
			CheckKind: recordpb.CheckKind_CHECK_KIND_DOCUMENT.Enum(), FoundBy: foundBy,
		})
	}
	evs := []*Event{
		find("L5-F1"), find("L6-F1"), // folded together into one gap
		find("L2-F1"), // read and never minted — the case this exists for
		mint("R1-1", "L5-F1", "L6-F1"),
	}
	by := map[string][]string{}
	for _, f := range FindingsJSONOf(evs).Findings {
		by[f.Label] = f.MintedAs
	}

	// THE FOLD IS VISIBLE FROM BOTH SIDES: two findings, one gap, each naming it.
	for _, l := range []string{"L5-F1", "L6-F1"} {
		if !slices.Contains(by[l], "R1-1") {
			t.Errorf("%s was folded into R1-1 and does not say so: %v", l, by[l])
		}
	}
	// AND THE DROP IS AN EMPTY LIST, NOT A MISSING KEY. `[]` means read and not minted, which is
	// the state that used to be invisible; omitting the field would put it back.
	if got, ok := by["L2-F1"]; !ok || len(got) != 0 {
		t.Errorf("L2-F1 produced no gap and the view does not say so plainly: %v (present=%v)", got, ok)
	}
	if by["L2-F1"] == nil {
		t.Error("L2-F1's minted_as is nil rather than an empty list — it marshals as null, which reads " +
			"as 'not computed' where the honest answer is 'nothing minted this'")
	}
}

// THE TYPED READ IS THE JOIN, and this is the trap the design can fall into.
//
// FindingsJSONOf computes minted_as from the mint bodies in the SAME event slice. Hand it only
// finding events and every finding reports `[]` — "no gap credits any finding" — in exactly the
// bytes the view uses for a genuine drop. That is the plausible zero, inside the fix for a
// plausible zero, so it is pinned here rather than trusted: FindingsJSONBytes must fetch MINT
// alongside FINDING.
func TestWithoutTheMintEventsEveryFindingReadsAsDropped(t *testing.T) {
	finding := recordtest.Event(t, "red-lens-r1-L5", 1, &recordpb.Finding{
		FindingId: proto.String("f-1"), Label: proto.String("L5-F1"), Text: proto.String("x"),
	})
	minted := recordtest.Event(t, "red-merge-r1", 1, &recordpb.Mint{
		GapId: proto.String("R1-1"), Class: proto.String("self-attestation"),
		Problem: proto.String("p"), RequiredFix: proto.String("f"), AcceptanceCheck: proto.String("a"),
		CheckKind: recordpb.CheckKind_CHECK_KIND_DOCUMENT.Enum(), FoundBy: []string{"L5-F1"},
	})
	withMint := FindingsJSONOf([]*Event{finding, minted}).Findings
	if len(withMint) != 1 || len(withMint[0].MintedAs) != 1 {
		t.Fatalf("with the mint present the credit must be seen: %+v", withMint)
	}
	// The same finding, same label, mint withheld: the view cannot tell this from a real drop.
	withoutMint := FindingsJSONOf([]*Event{finding}).Findings
	if len(withoutMint[0].MintedAs) != 0 {
		t.Fatalf("unexpected credit with no mint in the stream: %+v", withoutMint)
	}
	// So the RUN PATH has to ask for both. This is the assertion that keeps the typed read honest.
	if !slices.Contains(findingsViewEventTypes(), recordpb.EventType_EVENT_TYPE_MINT) {
		t.Error("FindingsJSONBytes does not fetch MINT, so minted_as is empty for every finding on " +
			"every run — reporting 'nothing was minted from any finding' in the same bytes as a drop")
	}
}
