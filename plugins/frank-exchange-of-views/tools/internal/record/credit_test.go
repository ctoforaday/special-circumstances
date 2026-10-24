package record

import (
	"encoding/json"
	"slices"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

func creditOf(gap, finding string) *recordpb.Credit {
	return &recordpb.Credit{GapId: proto.String(gap), FoundBy: proto.String(finding)}
}

// A CONTRADICTION IS ANSWERED BY THE OPEN GAP THAT HOLDS ITS DEFECT (gblock, 2026-10-24). The lens
// that raised it credits that gap with its raising finding: the gap's found_by gains the finding, the
// contradiction reads answered, and no duplicate is minted. The duplicate screen's refusal of the
// answering mint names the gap and the credit. Refused: a closed gap, another seat's finding, a
// finding that raised nothing, a contradiction already answered. A credit frees no mint.
func TestACreditAnswersARaiseWithTheOpenGap(t *testing.T) {
	const logic = "red-lens-logic"
	const sky = "the report says the sky is green, which its own source refutes outright"
	closed := corrMint("G-00000009")
	b := reopenedBoard(t, logic).dispatch(2, logic).register(logic).
		add(evLens, contradictingVerify("the sky is green")).
		add(logic, contradictingVerify("water is dry")).
		add(logic, closed).
		add(logic, &recordpb.Close{GapId: proto.String("G-00000009"), ClosureClass: recordtest.P(recordpb.Disposition_DISPOSITION_REPAIRED), Prose: proto.String("fixed")})
	run := b.seed()
	md := "# Report\n\nA paragraph that cites nothing."
	withRenderer(t, renders(&md))
	ids := map[string]string{}
	for _, c := range unansweredOf(t, run) {
		ids[c.Claim] = c.ID
	}
	ev, lg := Identity{Run: run, SeatID: evLens}, Identity{Run: run, SeatID: logic}

	held := corrMint("G-00000001")
	held.Problem = proto.String(sky)
	mustAppend(t, lg, held)
	mustAppend(t, ev, referenceFinding("F-00000001", ids["the sky is green"]))
	plain := referenceFinding("F-00000002", "")
	plain.AboutKind, plain.AboutRef = nil, nil
	mustAppend(t, ev, plain)
	mustAppend(t, lg, referenceFinding("F-00000003", ids["water is dry"]))

	t.Run("the screen refuses the answering mint and names the credit", func(t *testing.T) {
		dup := raisingGap("G-00000002", "F-00000001")
		dup.Problem = proto.String(sky)
		_, err := Append(ev, dup)
		mustRefuse(t, err, "would answer contradiction "+ids["the sky is green"], "your finding F-00000001 raised",
			"G-00000001 (open, minted by "+logic, "`credit --id G-00000001 --found-by F-00000001`")
	})
	t.Run("a closed gap answers nothing", func(t *testing.T) {
		_, err := Append(ev, creditOf("G-00000009", "F-00000001"))
		mustRefuse(t, err, "G-00000009 is closed")
	})
	t.Run("another seat's finding is not the actor's to credit", func(t *testing.T) {
		_, err := Append(lg, creditOf("G-00000001", "F-00000001"))
		mustRefuse(t, err, "finding F-00000001 is "+evLens+"'s, not yours")
	})
	t.Run("a finding that raised nothing answers nothing", func(t *testing.T) {
		_, err := Append(ev, creditOf("G-00000001", "F-00000002"))
		mustRefuse(t, err, "finding F-00000002 raised no contradiction")
	})

	mustAppend(t, ev, creditOf("G-00000001", "F-00000001"))
	for _, c := range unansweredOf(t, run) {
		if c.ID == ids["the sky is green"] {
			t.Fatalf("contradiction %s stands after its raiser credited G-00000001: %+v", c.ID, c)
		}
	}
	t.Run("an answered contradiction is not credited again", func(t *testing.T) {
		_, err := Append(ev, creditOf("G-00000001", "F-00000001"))
		mustRefuse(t, err, "is already answered, by gap G-00000001")
	})
	t.Run("a credit of the lens's own gap frees no mint", func(t *testing.T) {
		mustAppend(t, lg, creditOf("G-00000001", "F-00000003"))
		w, err := WorkOfSeat(run, "lens", logic)
		if err != nil {
			t.Fatal(err)
		}
		if got := w.Sitting.MintBudget; got == nil || got.Open != 1 || got.Answering != 0 {
			t.Errorf("%s's mint_budget after crediting its own gap = %+v, want open 1, answering 0", logic, got)
		}
	})

	board, err := BoardJSONOfRun(run)
	if err != nil {
		t.Fatal(err)
	}
	var foundBy []string
	for _, g := range board.Open {
		if g.ID == "G-00000001" {
			foundBy = g.FoundBy
		}
	}
	if !slices.Equal(foundBy, []string{"F-00000001", "F-00000003"}) {
		t.Errorf("G-00000001's found_by on the board = %v, want the two credited findings in order", foundBy)
	}
	raw, err := FindingsJSONBytes(run)
	if err != nil {
		t.Fatal(err)
	}
	var fj FindingsJSON
	if err := json.Unmarshal(raw, &fj); err != nil {
		t.Fatal(err)
	}
	for _, f := range fj.Findings {
		if f.ID == "F-00000001" && !slices.Equal(f.MintedAs, []string{"G-00000001"}) {
			t.Errorf("F-00000001's minted_as = %v, want the gap credited with it", f.MintedAs)
		}
	}
	blockers, err := PassBlockers(run)
	if err != nil {
		t.Fatal(err)
	}
	for _, bl := range blockers {
		if bl.Kind == BlockerContradiction {
			t.Errorf("a contradiction holds the PASS after both raises were credited: %+v", bl)
		}
	}
}
