package migrate

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// stubPlan is a plan as the verb prints one, with head and the given lines; every list present.
func stubPlan(head int64, why ...string) record.Plan {
	return record.Plan{Head: head, Parties: []record.Party{}, Docket: []string{}, RemandOwed: []record.RemandOwed{},
		Exchanges: []record.ExchangeCount{}, StaleAreas: []record.StaleArea{}, Why: why, Blockers: []record.PlanBlocker{}}
}

func storedOf(p record.Plan) *recordpb.Plan {
	return &recordpb.Plan{Printed: proto.String(p.WithKey().Printed())}
}

// A STORED PLAN NAMING MOVED EVENTS IS TRANSLATED, AND ITS SEAT WORDS ARE NOT (F8, F12). newID moves
// 412 to 411, as a source event before it that writes none would, and 637 to 636. The head, a stale
// area's pin, a contradiction blocker and every id phrase of a line follow it; a quoted claim, the
// remand line's tail after its prefix (an unclosed quote and an id phrase in it) and the remand's
// direction are carried as written; a content id's hex is not a numeral. A phrase that is not
// whole tokens (`spin`, `reopened at`), a moved id outside every phrase, a quote that does not close
// and a head that wrote no event are each refused.
func TestAStoredPlanNamingMovedEventsIsTranslated(t *testing.T) {
	newID := map[int64]int64{412: 411, 637: 636, 10: 10}
	const (
		direction = `re-read pin 412 and the claim "head 412`
		remand    = `G-1: at impasse (2 exchange(s) (0 stalled)) and remanded by the bench — its minting lens and blue are ready for the one more exchange the remand grants, on the ruling's direction: ` + direction
		claim     = `red-lens-evidence read a source contradicting "head 412", which blue has not answered`
		prefixed  = `red-lens-evidence read a source contradicting "x on the ruling's direction: pin 412" (contradiction 412), which blue has not answered`
		content   = `G-7b637cca: open, material, the lens's to answer`
	)
	in := stubPlan(412,
		"red-lens-logic: never sat — audits the report at head 412",
		"red-lens-voice: head 412 moved past its last sitting's pin 412",
		"red-lens-adversary: sat at the head (pin 412); retired (pin 412)",
		"A-1: changed at edit 412 after its last check; its last check is event 412",
		"red-lens-logic: retires once it owes nothing; it owes contradiction 412",
		claim, prefixed, remand, content)
	in.StaleAreas = []record.StaleArea{{SeatID: "red-lens-voice", Pin: 412}}
	in.Blockers = []record.PlanBlocker{{Kind: record.BlockerContradiction, Subject: "412", Owner: "red-lens-evidence"}}
	in.RemandOwed = []record.RemandOwed{{GapID: "G-1", Direction: direction}}
	pl := storedOf(in)
	before := pl.GetPrinted()
	if err := TranslatePlanIDs(pl, newID); err != nil {
		t.Fatalf("refused: %v", err)
	}
	got, again, err := record.ReadPlan([]byte(pl.GetPrinted()))
	if err != nil || again != pl.GetPrinted() {
		t.Fatalf("the translated plan does not read back as printed: %v", err)
	}
	if pl.GetPrinted() == before || got.Key != got.WithKey().Key {
		t.Errorf("the translated plan is not re-keyed: key %s, its own %s", got.Key, got.WithKey().Key)
	}
	if got.Head != 411 || got.StaleAreas[0].Pin != 411 || got.Blockers[0].Subject != "411" {
		t.Errorf("head %d, pin %d, contradiction blocker %q; want each 411", got.Head, got.StaleAreas[0].Pin, got.Blockers[0].Subject)
	}
	want := []string{
		"red-lens-logic: never sat — audits the report at head 411",
		"red-lens-voice: head 411 moved past its last sitting's pin 411",
		"red-lens-adversary: sat at the head (pin 411); retired (pin 411)",
		"A-1: changed at edit 411 after its last check; its last check is event 411",
		"red-lens-logic: retires once it owes nothing; it owes contradiction 411",
		claim,
		strings.Replace(prefixed, "(contradiction 412)", "(contradiction 411)", 1),
		remand, content,
	}
	for i := range want {
		if i >= len(got.Why) || got.Why[i] != want[i] {
			t.Errorf("line %d:\n got %q\nwant %q", i, got.Why[i], want[i])
		}
	}
	if len(got.RemandOwed) != 1 || got.RemandOwed[0].Direction != direction {
		t.Errorf("remand_owed = %+v; its direction is carried as written", got.RemandOwed)
	}

	for name, c := range map[string]struct {
		p    record.Plan
		frag string
	}{
		"a phrase word that is not a whole token": {stubPlan(412, "red-lens-logic: spin 412 is no phrase"), "outside every id phrase"},
		"a moved id after no phrase word":         {stubPlan(412, "red-lens-logic: retires at 412"), "outside every id phrase"},
		"a phrase that is not whole tokens":       {stubPlan(412, "red-lens-logic: reopened at 412"), "outside every id phrase"},
		"a quote that does not close":             {stubPlan(412, `red-lens-evidence read a source contradicting "head 412`), "does not close"},
		"an unlanded head":                        {stubPlan(500), "wrote no event"},
	} {
		pl := storedOf(c.p)
		was := pl.GetPrinted()
		err := TranslatePlanIDs(pl, newID)
		if err == nil || !strings.Contains(err.Error(), c.frag) {
			t.Errorf("%s: %v, want a refusal naming %q", name, err, c.frag)
		}
		if pl.GetPrinted() != was {
			t.Errorf("%s: a refused plan was rewritten", name)
		}
	}

	same := storedOf(stubPlan(10, "red-lens-logic: sat at the head (pin 10); 2 claim(s)", content))
	was := same.GetPrinted()
	if err := TranslatePlanIDs(same, newID); err != nil || same.GetPrinted() != was {
		t.Errorf("a plan naming no moved id: %v, bytes kept %v", err, same.GetPrinted() == was)
	}
}

// THE REMAND PREFIX MIGRATE READS IS THE ONE THE VERB PRINTS (F14). migrate's copy describes archived
// bytes; a rewording of the live one is a second entry here, not an edit of this one.
func TestMigrateKnowsTheVerbsRemandPrefix(t *testing.T) {
	if remandDirectionPrefix != record.RemandDirectionPrefix {
		t.Errorf("the remand prefix the verb prints changed — keep the old wording for archived bytes and add the new one (migrate %q, the verb %q)",
			remandDirectionPrefix, record.RemandDirectionPrefix)
	}
}
