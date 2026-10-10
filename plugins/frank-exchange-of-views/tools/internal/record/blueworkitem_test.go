package record

import (
	"encoding/json"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

const (
	quotedGap  = "G-00000001" // its mint prescribes exact text over a sentence the report holds
	proseGap   = "G-00000002" // its mint prescribes none
	closedGap  = "G-00000003" // its lens closes it between blue's dispatch and blue's register
	blueSeatID = "blue-respond"
)

// blueWorkStage is a run in which blue-respond is dispatched on three gaps and sits: one whose mint
// proposed exact text over a sentence carrying the gap's anchor, one whose mint proposed none, and
// one its lens closed after the dispatch and before blue registered.
func blueWorkStage(t *testing.T) Run {
	t.Helper()
	mint := func(id string, extra func(*recordpb.Mint)) *recordpb.Mint {
		m := &recordpb.Mint{GapId: proto.String(id), Class: proto.String("x"), Problem: proto.String("p"),
			AcceptanceCheck: proto.String("c"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT),
			Severity: recordtest.P(recordpb.Grade_GRADE_HIGH), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM),
			Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM), FixBasis: proto.String("proposed")}
		if extra != nil {
			extra(m)
		}
		return m
	}
	// The record package links no renderer; the report is the one a mint over this sentence leaves.
	was := reportRenderer
	t.Cleanup(func() { reportRenderer = was })
	reportRenderer = func(string, bool, []ReportOp) (string, error) {
		return "# R\n\nThe sky is teal at noon." + anchor.Token(quotedGap) + " Next one.\n", nil
	}
	b := newStage(t).cast(evLens, "red-chair", blueSeatID).ingest().
		register(evLens).
		add(evLens, mint(quotedGap, func(m *recordpb.Mint) {
			m.FixBasis, m.Location, m.FixNew = proto.String("verified"), proto.String("The sky is teal at noon."), proto.String("The sky is blue at noon.")
		})).
		add(evLens, mint(proseGap, nil)).
		add(evLens, mint(closedGap, nil)).
		bracket("red-chair", "chair-a").dispatch(1, blueSeatID, quotedGap, proseGap, closedGap).
		closeGap(evLens, closedGap).
		registerAs(blueSeatID, "blue-a")
	return b.seed()
}

func workItemT(t *testing.T, w WorkJSON, id string) WorkGapJSON {
	t.Helper()
	for _, g := range w.Open {
		if g.ID == id {
			return g
		}
	}
	t.Fatalf("the work list holds no open item %s: %+v", id, w.Open)
	return WorkGapJSON{}
}

// BLUE'S WORK ITEM CARRIES THE FIX AS A QUOTED SPAN, AND IT IS THE BOARD'S PAIR. A gap whose mint
// prescribed exact text reads fix_basis, fix_old and fix_new on the work item, byte for byte what
// the board serves for that gap — the pair an accepted edit sends.
func TestTheWorkItemCarriesTheQuotedFixTheBoardServes(t *testing.T) {
	run := blueWorkStage(t)
	w, err := WorkOfSeat(run, "blue", blueSeatID)
	if err != nil {
		t.Fatal(err)
	}
	board, err := BoardJSONOfRun(run)
	if err != nil {
		t.Fatal(err)
	}
	var onBoard GapJSON
	for _, g := range board.Open {
		if g.ID == quotedGap {
			onBoard = g
		}
	}
	got := workItemT(t, w, quotedGap)
	if got.FixBasis != "verified" || got.FixNew == "" {
		t.Fatalf("work item %s = fix_basis %q, fix_new %q; its mint prescribed exact text on a verified basis", quotedGap, got.FixBasis, got.FixNew)
	}
	if got.FixBasis != onBoard.FixBasis || got.FixOld != onBoard.FixOld || got.FixNew != onBoard.FixNew {
		t.Errorf("work item fix = (%q, %q, %q), the board's = (%q, %q, %q) — one computation, two carriers",
			got.FixBasis, got.FixOld, got.FixNew, onBoard.FixBasis, onBoard.FixOld, onBoard.FixNew)
	}
	// The pair is the one read over the report: the old half runs through the gap's own anchor and
	// the new half carries it, which the recorded pair does not.
	if tok := anchor.Token(quotedGap); !strings.Contains(got.FixOld, tok) || !strings.Contains(got.FixNew, tok) {
		t.Errorf("work item fix = (%q, %q), want both halves carrying the gap's anchor as the report holds it", got.FixOld, got.FixNew)
	}
}

// A GAP WHOSE MINT PRESCRIBED NO TEXT SAYS SO, and says it differently from a field nobody
// computed: its basis reads `proposed` and neither half of a pair is on the item.
func TestAWorkItemWithNoProposedTextStatesItsBasisAndCarriesNoPair(t *testing.T) {
	w, err := WorkOfSeat(blueWorkStage(t), "blue", blueSeatID)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(workItemT(t, w, proseGap))
	if err != nil {
		t.Fatal(err)
	}
	var keys map[string]any
	if err := json.Unmarshal(raw, &keys); err != nil {
		t.Fatal(err)
	}
	if keys["fix_basis"] != "proposed" {
		t.Errorf("fix_basis = %v, want \"proposed\" — the mint's own word for a demand with no exact text", keys["fix_basis"])
	}
	for _, k := range []string{"fix_old", "fix_new"} {
		if v, present := keys[k]; present {
			t.Errorf("%s = %v on a gap whose mint prescribed no text; an absent pair is the answer", k, v)
		}
	}
}

// THE WORK LIST NAMES THE ENGAGED GAPS BLUE FOUND CLOSED, off the sitting the position duty and
// capture's record-parity read (BlueSittings): each gap of the dispatch a close preceded blue's
// register on, with who closed it and the closure's fate.
func TestBluesWorkListNamesTheEngagedGapsItFoundClosed(t *testing.T) {
	run := blueWorkStage(t)
	w, err := WorkOfSeat(run, "blue", blueSeatID)
	if err != nil {
		t.Fatal(err)
	}
	e := w.Engaged
	if e == nil {
		t.Fatal("blue-respond's work list has a null engaged; its dispatch is on the record")
	}
	if got := strings.Join(e.GapIDs, " "); got != quotedGap+" "+proseGap+" "+closedGap {
		t.Errorf("engaged gap_ids = %q, want the dispatch's three", got)
	}
	want := FoundClosedJSON{ID: closedGap, ClosedBySeat: evLens, Fate: "repaired"}
	if len(e.FoundClosed) != 1 || e.FoundClosed[0] != want {
		t.Errorf("found_closed = %+v, want exactly %+v", e.FoundClosed, want)
	}
	// The same sitting, the same answer: what is not found closed is what the sitting owes on.
	m, err := MergedEvents(run)
	if err != nil {
		t.Fatal(err)
	}
	ss := BlueSittings(m.Events, m.At, WhileRunning)
	if n := len(ss[len(ss)-1].Engaged) - len(ss[len(ss)-1].Open); n != len(e.FoundClosed) {
		t.Errorf("found_closed names %d gap(s); the sitting the position duty reads found %d closed", len(e.FoundClosed), n)
	}

	// Nothing found closed is an empty list, never an absent one.
	open, err := WorkOfSeat(newStage(t).cast(evLens, "red-chair", blueSeatID).ingest().register(evLens).mint(evLens, quotedGap, "high").
		bracket("red-chair", "chair-a").dispatch(1, blueSeatID, quotedGap).registerAs(blueSeatID, "blue-a").seed(), "blue", blueSeatID)
	if err != nil {
		t.Fatal(err)
	}
	if e := open.Engaged; e == nil || e.FoundClosed == nil || len(e.FoundClosed) != 0 {
		t.Errorf("engaged = %+v, want the block with an empty found_closed list", e)
	}

	// Every other seat is dispatched to no gap another seat closes: its list says null, and says it.
	raw, err := WorkJSONBytes(run, "lens", evLens)
	if err != nil {
		t.Fatal(err)
	}
	var lens map[string]any
	if err := json.Unmarshal(raw, &lens); err != nil {
		t.Fatal(err)
	}
	if v, present := lens["engaged"]; !present || v != nil {
		t.Errorf("a lens's work list has engaged = %v (present %v), want the key with null", v, present)
	}
}
