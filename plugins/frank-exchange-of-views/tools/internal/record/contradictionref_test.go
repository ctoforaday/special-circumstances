package record

import (
	"encoding/json"
	"slices"
	"strconv"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// THE CONTRADICTION RULE, PER CLAIM (gblock, 2026-10-06), driven through the fold every surface
// reads (contradictionsOf): a reading opens a contradiction whose id is its events.id, a later
// reading of the same claim joins it, a finding raises it by reference or by quoting the claim,
// a reading after the raise opens a new one, and the gap minted from the raising finding answers it.

// raisingGap is a gap minted from finding, crediting it in found_by: what answers a contradiction
// the finding raised.
func raisingGap(gap, finding string) *recordpb.Mint {
	m := corrMint(gap)
	m.Problem = proto.String("the report says what its source does not")
	m.FoundBy = []string{finding}
	return m
}

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
	if len(cs) != 1 || cs[0].raisedBy != "F-00000001" || cs[0].answeredBy != "" {
		t.Errorf("a raise by reference = %+v, want contradiction %s raised by F-00000001 and unanswered", cs, id)
	}
	// THE CLASS DEFAULT WINS: only the gap minted from the raising finding answers it. A gap
	// crediting another finding answers nothing.
	cs = contradictionsT(t, read(evLens), recordtest.Event(t, evLens, referenceFinding("F-00000001", id)),
		recordtest.Event(t, evLens, raisingGap("G-00000002", "F-00000009")))
	if len(cs) != 1 || cs[0].answeredBy != "" {
		t.Errorf("a gap crediting another finding = %+v, want contradiction %s unanswered", cs, id)
	}
	cs = contradictionsT(t, read(evLens), recordtest.Event(t, evLens, referenceFinding("F-00000001", id)),
		recordtest.Event(t, evLens, raisingGap("G-00000001", "F-00000001")))
	if len(cs) != 1 || cs[0].answeredBy != "G-00000001" {
		t.Errorf("the gap minted from the raising finding = %+v, want contradiction %s answered by G-00000001", cs, id)
	}

	// (ii) A reading after the answer opens a NEW contradiction with its own id, owned by its reader.
	cs = contradictionsT(t, read(evLens), recordtest.Event(t, evLens, referenceFinding("F-00000001", id)), read("red-lens-logic"))
	if len(cs) != 2 || cs[0].raisedBy == "" || cs[1].raisedBy != "" || cs[1].id == cs[0].id || cs[1].reader != "red-lens-logic" {
		t.Errorf("a reading after the answer = %+v, want the answered one and a new open one owned by red-lens-logic", cs)
	}

	// The quote answers too — but only what was read before it.
	quote := &recordpb.Finding{Id: proto.String("F-00000002"), Location: proto.String(claim), Text: proto.String("t")}
	if cs = contradictionsT(t, read(evLens), recordtest.Event(t, evLens, quote)); len(cs) != 1 || cs[0].raisedBy != "F-00000002" {
		t.Errorf("a finding quoting the claim after the reading = %+v, want it raised", cs)
	}
	if cs = contradictionsT(t, recordtest.Event(t, evLens, quote), read(evLens)); len(cs) != 1 || cs[0].raisedBy != "" {
		t.Errorf("a finding quoting the claim BEFORE the reading = %+v, want the reading open: the finding was not an answer to it", cs)
	}
}

// (iii)-(v) THE WRITE REFUSES what the rule does not admit: a reference that names no open
// contradiction, one already answered (naming the answer), one that is not an id, and a gap about
// a contradiction — at validate, with no CLI involved.
func TestTheWriteRefusesAReferenceTheRuleDoesNotAdmit(t *testing.T) {
	b := reopenedBoard(t).add(evLens, contradictingVerify("the sky is green"))
	m := loadedT(t, b.evs...)
	open := unansweredContradictions(m.Events, m.At)
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
	if _, err := Append(Identity{Run: run, SeatID: evLens}, referenceFinding("F-00000002", id)); err == nil || !strings.Contains(err.Error(), "already raised, by finding F-00000001") || !strings.Contains(err.Error(), "--found-by F-00000001") {
		t.Errorf("a second raise of contradiction %s: %v, want a refusal naming F-00000001 and the gap that answers it", id, err)
	}
	mint := &recordpb.Mint{GapId: proto.String("G-00000001"), Class: recordpb.GapClass_GAP_CLASS_SPEC_UNDERSPECIFICATION.Enum(), Problem: proto.String("p"),
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
		t.Fatalf("the raise by reference of an edited-away claim was refused: %v", err)
	}
	// RAISED IS NOT ANSWERED: the finding alone leaves the PASS held, and the item names the gap owed.
	raised := false
	for _, b := range mustBlockers(t, run) {
		if b.Kind == BlockerContradiction {
			raised = true
			if item := b.WorkItem(); !strings.Contains(item, "finding F-00000001 raised it") || !strings.Contains(item, "`mint --found-by F-00000001`") {
				t.Errorf("a raised contradiction's work item does not name the gap owed: %s", item)
			}
		}
	}
	if !raised {
		t.Fatal("contradiction stopped holding the PASS on the finding alone — a source the report misstates is material by its class, and only its gap answers it")
	}
	if _, err := Append(Identity{Run: run, SeatID: evLens}, raisingGap("G-00000001", "F-00000001")); err != nil {
		t.Fatalf("the gap minted from the raising finding was refused: %v", err)
	}
	for _, b := range mustBlockers(t, run) {
		if b.Kind == BlockerContradiction {
			t.Errorf("contradiction %s still holds the PASS after the gap minted from its finding answered it", b.Subject)
		}
	}
}

// THE CITATION'S STANDING ANSWERS A CONTRADICTION ON IT (gblock, 2026-10-16): a retire of the
// anchor, or the anchor holding — a check that backs the sentence stands and none that does not,
// so the reading has been replaced by an edit or by its own seat. Each case is read on the PASS
// gate and on `show evidence`, which must agree.
func TestTheCitationsStandingAnswersAContradiction(t *testing.T) {
	refutes := recordpb.SourceOutcome_SOURCE_OUTCOME_REFUTES
	// rereads is a verify of C-1 against the sentence the edit wrote, which the tool reads off the
	// anchor: a refuting one opens a contradiction of its own rather than joining the first.
	rereads := func(b *stage, seat string) *stage {
		return b.add(seat, &recordpb.Verify{Anchor: proto.String("C-1"), Claim: proto.String("the claim as edited"), Outcome: refutes.Enum(),
			Confidence: recordpb.Confidence_CONFIDENCE_HIGH.Enum(), Text: proto.String("read at the leaf")})
	}
	sits := func(b *stage, seat string) *stage { return b.dispatch(b.at(), seat).register(seat) }
	// contradicted is the evidence lens's backing check of C-1 standing, then the logic lens's
	// refuting reading of it from a sitting of its own.
	contradicted := func() (*stage, string) {
		b := reopenedBoard(t, outsideLens)
		sits(b, outsideLens).verify(outsideLens, "C-1", refutes)
		return b, strconv.FormatInt(b.at(), 10)
	}
	for _, c := range []struct {
		name  string
		build func(b *stage) *stage
		open  bool
	}{
		{"another lens's backing check, with no edit since the reading, is a disagreement", func(b *stage) *stage { return b }, true},
		{"a retire of the anchor answers it", func(b *stage) *stage {
			return b.add("blue-respond", &recordpb.Retire{Claim: proto.String("the claim C-1 backs"), Reason: proto.String("wrong"), Anchors: []string{"C-1"}})
		}, false},
		{"a backing check from a sitting opened after the edit answers it", func(b *stage) *stage {
			return sits(b.reopen("C-1"), evLens).supports(evLens, "C-1")
		}, false},
		{"a backing check before the edit does not", func(b *stage) *stage {
			return sits(b, evLens).supports(evLens, "C-1").reopen("C-1")
		}, true},
		{"a backing check after the edit from a sitting opened before it does not", func(b *stage) *stage {
			return sits(b, evLens).reopen("C-1").supports(evLens, "C-1")
		}, true},
		{"a check from a later sitting that does not back the sentence does not", func(b *stage) *stage {
			return rereads(sits(b.reopen("C-1"), outsideLens), outsideLens)
		}, true},
		{"a backing check its own seat's later check replaced does not", func(b *stage) *stage {
			return rereads(sits(b.reopen("C-1"), evLens).supports(evLens, "C-1"), evLens)
		}, true},
		{"the reader's own later backing check, with no edit, answers it", func(b *stage) *stage {
			return sits(b, outsideLens).supports(outsideLens, "C-1")
		}, false},
	} {
		t.Run(c.name, func(t *testing.T) {
			b, id := contradicted()
			if got := openContradictionIDs(t, c.build(b)); slices.Contains(got, id) != c.open {
				t.Errorf("open contradictions = %v; contradiction %s open = %v, want %v", got, id, !c.open, c.open)
			}
		})
	}

	// AN EDIT BEFORE THE READING IS NOT AN EDIT OF WHAT WAS READ: the reading is of the sentence as
	// it stands, so another lens's backing check of that same sentence is still a disagreement.
	b := reopenedBoard(t, outsideLens).reopen("C-1")
	sits(b, outsideLens).verify(outsideLens, "C-1", refutes)
	id := strconv.FormatInt(b.at(), 10)
	if got := openContradictionIDs(t, sits(b, evLens).supports(evLens, "C-1")); !slices.Contains(got, id) {
		t.Errorf("open contradictions = %v, want %s open: the edit came before the reading", got, id)
	}
}

// openContradictionIDs is the contradictions holding the PASS, held equal to the ones `show
// evidence` lists as unanswered.
func openContradictionIDs(t *testing.T, b *stage) []string {
	t.Helper()
	run := b.seed()
	var held []string
	for _, bl := range mustBlockers(t, run) {
		if bl.Kind == BlockerContradiction {
			held = append(held, bl.Subject)
		}
	}
	raw, err := EvidenceJSONBytes(run)
	if err != nil {
		t.Fatal(err)
	}
	var ej EvidenceJSON
	if err := json.Unmarshal(raw, &ej); err != nil {
		t.Fatal(err)
	}
	var listed []string
	for _, c := range ej.UnansweredContradictions {
		listed = append(listed, c.ID)
	}
	if !slices.Equal(held, listed) {
		t.Errorf("the PASS gate holds contradictions %v, and show evidence lists %v unanswered", held, listed)
	}
	return held
}
