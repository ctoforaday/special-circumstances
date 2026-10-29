package record

import (
	"encoding/json"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// A LENS CAN ALWAYS ANSWER WHAT IT RAISED (gblock, 2026-10-20). A contradiction is raised by a
// finding and answered only by the gap minted from that finding; the raiser's answering gap takes no
// mint of its budget, is admitted at the cap, and does not count while it is open. Every other mint
// counts — a second gap from an answered raise, a gap crediting another lens's raise, a gap crediting
// a finding that raised nothing.
func TestALensCanAlwaysAnswerWhatItRaised(t *testing.T) {
	const logic = "red-lens-logic"
	b := reopenedBoard(t, logic).dispatch(2, logic).register(logic).
		add(evLens, contradictingVerify("the sky is green")).
		add(evLens, contradictingVerify("grass is blue")).
		add(logic, contradictingVerify("water is dry"))
	writeRunConfig(t, b.runDir, `{"mintBudget":1}`)
	run := b.seed()
	md := "# Report\n\nA paragraph that cites nothing."
	withRenderer(t, renders(&md))
	ids := map[string]string{}
	for _, c := range unansweredOf(t, run) {
		ids[c.Claim] = c.ID
	}
	ev, lg := Identity{Run: run, SeatID: evLens}, Identity{Run: run, SeatID: logic}

	standing := func(t *testing.T, id Identity, open, answering, left int) {
		t.Helper()
		w, err := WorkOfSeat(run, "lens", id.SeatID)
		if err != nil {
			t.Fatal(err)
		}
		got := w.Sitting.MintBudget
		if got == nil || got.Open != open || got.Answering != answering || got.Left != left {
			t.Fatalf("%s's mint_budget = %+v, want open %d, answering %d, left %d", id.SeatID, got, open, answering, left)
		}
		if answering > 0 && !mentions(w.Sitting.Open, "more that answer a contradiction you raised or a check of yours that failed") {
			t.Errorf("%s's work list does not say what its answering gaps cost: %v", id.SeatID, hows(w.Sitting.Open))
		}
	}
	answering := func(gap, finding string) *recordpb.Mint {
		m := raisingGap(gap, finding)
		m.Problem = proto.String(map[string]string{
			"G-00000002": "the sky is not green", "G-00000005": "grass is not blue", "G-00000006": "the moon is not cheese",
			"G-00000007": "rivers run uphill", "G-00000008": "salt is sweet", "G-0000000b": "water is wet",
		}[gap])
		return m
	}

	// The evidence lens's budget is max(1, ceil(1 citation × 3 / 2)) = 2.
	mustAppend(t, ev, budgetMint("G-00000001"))
	standing(t, ev, 1, 0, 1)
	mustAppend(t, ev, referenceFinding("F-00000001", ids["the sky is green"]))
	mustAppend(t, ev, answering("G-00000002", "F-00000001"))
	standing(t, ev, 1, 1, 1) // the answering gap took no slot
	mustAppend(t, ev, budgetMint("G-00000003"))
	standing(t, ev, 2, 1, 0)
	_, err := Append(ev, budgetMint("G-00000004"))
	mustRefuse(t, err, "holds 2 gap(s) open and its mint budget admits no more", "takes no mint")

	t.Run("at the cap, a raise of its own is answered", func(t *testing.T) {
		mustAppend(t, ev, referenceFinding("F-00000002", ids["grass is blue"]))
		mustAppend(t, ev, answering("G-00000005", "F-00000002"))
		standing(t, ev, 2, 2, 0)
	})
	t.Run("a raise already answered answers nothing again", func(t *testing.T) {
		_, err := Append(ev, answering("G-00000006", "F-00000001"))
		mustRefuse(t, err, "admits no more")
	})
	t.Run("another lens's raise launders nothing", func(t *testing.T) {
		mustAppend(t, lg, referenceFinding("F-00000003", ids["water is dry"]))
		_, err := Append(ev, answering("G-00000007", "F-00000003"))
		mustRefuse(t, err, "admits no more")
	})
	t.Run("a finding that raised nothing launders nothing", func(t *testing.T) {
		f := referenceFinding("F-00000004", "")
		f.AboutKind, f.AboutRef = nil, nil
		mustAppend(t, ev, f)
		_, err := Append(ev, answering("G-00000008", "F-00000004"))
		mustRefuse(t, err, "admits no more")
	})
	t.Run("the raiser answers at its own cap", func(t *testing.T) {
		// The logic lens's budget is max(1, 0 claims) = 1.
		mustAppend(t, lg, budgetMint("G-00000009"))
		_, err := Append(lg, budgetMint("G-0000000a"))
		mustRefuse(t, err, "admits no more")
		mustAppend(t, lg, answering("G-0000000b", "F-00000003"))
		standing(t, lg, 1, 1, 0)
	})
	if left := unansweredOf(t, run); len(left) != 0 {
		t.Errorf("%d contradiction(s) stand after every raise was answered: %+v", len(left), left)
	}
}

// unansweredOf is the run's unanswered contradictions, as `show evidence` lists them.
func unansweredOf(t *testing.T, run Run) []EvidenceContradictionJSON {
	t.Helper()
	raw, err := EvidenceJSONBytes(run)
	if err != nil {
		t.Fatal(err)
	}
	var ev EvidenceJSON
	if err := json.Unmarshal(raw, &ev); err != nil {
		t.Fatal(err)
	}
	return ev.UnansweredContradictions
}
