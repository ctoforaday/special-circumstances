package record

import (
	"strconv"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// THE CONTRADICTION RULE, PER CLAIM (gblock, 2026-10-06), driven through the fold every surface
// reads (contradictionsOf): a reading opens a contradiction whose id is its events.id, a later
// reading of the same claim joins it, a finding answers it by reference or by quoting the claim,
// and a reading after the answer opens a new one.

func referenceFinding(id, ref string) *recordpb.Finding {
	return &recordpb.Finding{Id: proto.String(id), Text: proto.String("the source says otherwise"),
		AboutKind: recordpb.AboutKind_ABOUT_KIND_CONTRADICTION.Enum(), AboutRef: proto.String(ref)}
}

// contradictionsT is the fold over the record as the write path stores it.
func contradictionsT(t *testing.T, evs ...*Event) []contradiction {
	t.Helper()
	m := loadedT(t, evs...)
	return contradictionsOf(m.Events, m.At.IDs(m.Events))
}

func TestAContradictionIsPerClaim(t *testing.T) {
	const claim = "the sky is green"
	read := func(seat string) *Event { return recordtest.Event(t, seat, contradictingVerify(claim)) }

	// (i) Two lenses read the claim, then one answer: answered, and both readings are covered.
	cs := contradictionsT(t, read(evLens), read("red-lens-logic"))
	if len(cs) != 1 || cs[0].reader != "red-lens-logic" || cs[0].since == cs[0].id {
		t.Fatalf("two readings of one claim = %+v, want one contradiction owned by the later reader, since its reading", cs)
	}
	id := cs[0].idWord()
	cs = contradictionsT(t, read(evLens), read("red-lens-logic"), recordtest.Event(t, evLens, referenceFinding("F-00000001", id)))
	if len(cs) != 1 || cs[0].answeredBy != "F-00000001" {
		t.Errorf("an answer by reference = %+v, want contradiction %s answered by F-00000001", cs, id)
	}

	// (ii) A reading after the answer opens a NEW contradiction with its own id, owned by its reader.
	cs = contradictionsT(t, read(evLens), recordtest.Event(t, evLens, referenceFinding("F-00000001", id)), read("red-lens-logic"))
	if len(cs) != 2 || cs[0].answeredBy == "" || cs[1].answeredBy != "" || cs[1].id == cs[0].id || cs[1].reader != "red-lens-logic" {
		t.Errorf("a reading after the answer = %+v, want the answered one and a new open one owned by red-lens-logic", cs)
	}

	// The quote answers too — but only what was read before it.
	quote := &recordpb.Finding{Id: proto.String("F-00000002"), Location: proto.String(claim), Text: proto.String("t")}
	if cs = contradictionsT(t, read(evLens), recordtest.Event(t, evLens, quote)); len(cs) != 1 || cs[0].answeredBy != "F-00000002" {
		t.Errorf("a finding quoting the claim after the reading = %+v, want it answered", cs)
	}
	if cs = contradictionsT(t, recordtest.Event(t, evLens, quote), read(evLens)); len(cs) != 1 || cs[0].answeredBy != "" {
		t.Errorf("a finding quoting the claim BEFORE the reading = %+v, want the reading open: the finding was not an answer to it", cs)
	}
}

// (iii)-(v) THE WRITE REFUSES what the rule does not admit: a reference that names no open
// contradiction, one already answered (naming the answer), one that is not an id, and a gap about
// a contradiction — at validate, with no CLI involved.
func TestTheWriteRefusesAReferenceTheRuleDoesNotAdmit(t *testing.T) {
	b := reopenedBoard(t).add(evLens, contradictingVerify("the sky is green"))
	m := loadedT(t, b.evs...)
	open := unansweredContradictions(m.Events, m.At.IDs(m.Events))
	if len(open) != 1 {
		t.Fatalf("fixture: %d open contradictions, want 1", len(open))
	}
	id := open[0].idWord()
	run := b.seed()
	for _, c := range []struct {
		ref, says string
	}{
		{"not-a-number", "is not a contradiction's id"},
		{strconv.FormatInt(open[0].id+999, 10), "names no contradiction on this record (1 open)"},
	} {
		if _, err := Append(Identity{Run: run, SeatID: evLens}, referenceFinding("F-00000009", c.ref)); err == nil || !strings.Contains(err.Error(), c.says) {
			t.Errorf("a finding about contradiction %q: %v, want a refusal saying %q", c.ref, err, c.says)
		}
	}
	if _, err := Append(Identity{Run: run, SeatID: evLens}, referenceFinding("F-00000001", id)); err != nil {
		t.Fatalf("the answer by reference was refused: %v", err)
	}
	if _, err := Append(Identity{Run: run, SeatID: evLens}, referenceFinding("F-00000002", id)); err == nil || !strings.Contains(err.Error(), "already answered, by finding F-00000001") {
		t.Errorf("a second answer to contradiction %s: %v, want a refusal naming F-00000001", id, err)
	}
	mint := &recordpb.Mint{GapId: proto.String("G-00000001"), Class: proto.String("scope-creep"), Problem: proto.String("p"),
		CheckKind: recordpb.CheckKind_CHECK_KIND_DOCUMENT.Enum(), AcceptanceCheck: proto.String("c"),
		Severity: recordpb.Grade_GRADE_LOW.Enum(), Likelihood: recordpb.Grade_GRADE_LOW.Enum(), Impact: recordpb.Grade_GRADE_LOW.Enum(),
		AboutKind: recordpb.AboutKind_ABOUT_KIND_CONTRADICTION.Enum(), AboutRef: proto.String(id)}
	if _, err := Append(Identity{Run: run, SeatID: evLens}, mint); err == nil || !strings.Contains(err.Error(), "a gap is not about a contradiction") {
		t.Errorf("a mint about a contradiction: %v, want the write's refusal", err)
	}
}

// AN EDITED-AWAY CONTRADICTION IS ANSWERED BY REFERENCE: once blue edits the sentence a finding can
// no longer quote it, and the id still names the reading. Every surface names the blocker by
// `contradiction <id>` — a bare decimal would match any digit of a refusal.
func TestAnEditedAwayContradictionIsAnsweredByReference(t *testing.T) {
	b := reopenedBoard(t).add(evLens, contradictingVerify("the sky is green")).
		add("blue-respond", &recordpb.BlueEdit{Old: proto.String("before"), New: proto.String("after"), Text: proto.String("why")})
	run := b.seed()
	bs := mustBlockers(t, run)
	var held *Blocker
	for i := range bs {
		if bs[i].Kind == BlockerContradiction {
			held = &bs[i]
		}
	}
	if held == nil || held.Detail != "the sky is green" || held.Owner != evLens {
		t.Fatalf("blockers = %+v, want a contradiction on the claim owned by %s", bs, evLens)
	}
	named := "contradiction " + held.Subject
	if item := held.WorkItem(); !strings.Contains(item, named) || !strings.Contains(item, "`finding --about-kind contradiction --about "+held.Subject+"`") {
		t.Errorf("the work item does not name %q and its answer: %s", named, item)
	}
	if why := kindOf(held.Kind).why(*held); !strings.Contains(why, named) {
		t.Errorf("the plan's reason does not name %q: %s", named, why)
	}
	if err := passRefusalOver(t, run, BlockerContradiction); err == nil || !strings.Contains(err.Error(), named) {
		t.Errorf("the PASS refusal does not name %q: %v", named, err)
	}
	if _, err := Append(Identity{Run: run, SeatID: evLens}, referenceFinding("F-00000001", held.Subject)); err != nil {
		t.Fatalf("the answer by reference to an edited-away claim was refused: %v", err)
	}
	for _, b := range mustBlockers(t, run) {
		if b.Kind == BlockerContradiction {
			t.Errorf("contradiction %s still holds the PASS after the finding answered it", b.Subject)
		}
	}
}
