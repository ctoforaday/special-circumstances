package record

import (
	"slices"
	"testing"
)

// THE QUEUE IS THE LIST IN ORDER, KEYED BY FIELDS. The audit comes first while the lens has not been
// served the report since it was behind, then what the seat owes in list order, then the gaps the
// dispatch readies it on — material first, then by mass, then in mint order — with the items about
// each folded in, then the regrades and findings left over. Nothing else is queued.
func TestTheQueueOrdersTheListByItsFields(t *testing.T) {
	gap := func(id string, material, turn bool, l, i string) WorkGapJSON {
		return WorkGapJSON{ID: id, Material: material, YourTurn: turn, Likelihood: l, Impact: i, MintedBy: "red-lens-logic", RegradesOwed: []OwedRegradeJSON{}}
	}
	w := WorkJSON{
		Open: []WorkGapJSON{
			gap("G-1", false, true, "high", "high"),
			gap("G-2", true, true, "low", "low"),
			gap("G-3", true, true, "high", "high"),
			gap("G-4", false, false, "high", "high"), // another lens's, or not readied: index only
			gap("G-5", false, true, "low", "low"),
		},
		Sitting: SittingJSON{
			Role: "lens", LastRead: &LastReadJSON{Kind: "behind", Pin: 3, Head: 9},
			Open: []Item{
				{Kind: ItemLog, What: "the log is open"},
				{Kind: ItemAudit, Subject: about(SubjectReport, "9"), What: "audit"},
				gapItem(ItemReaudit, "G-3", "G-3 was answered"),
				gapItem(ItemRegradeOwed, "G-3", "G-3 regrade owed"),
				{Kind: ItemKind(BlockerFailedCheck), Subject: about(SubjectAnchor, "c-7"), Blocks: true, What: "c-7 read nothing"},
				gapItem(ItemNews, "G-9", "G-9's answers are retired"),
				gapItem(ItemFindingOnGap, "G-4", "a finding about G-4"),
				{Kind: ItemKind(BlockerUnchecked), Subject: about(SubjectAnchor, "c-8"), What: "c-8 is unchecked — no check can reach it"},
				{Kind: ItemLaw, What: "law"},
			},
		},
	}
	q := QueueOf(w)
	var keys []string
	for _, e := range q {
		keys = append(keys, e.Key)
	}
	want := []string{
		"audit:report:9",
		"failed_check:anchor:c-7",
		"reaudit:gap:G-3", // material, mass 1.0
		"reaudit:gap:G-2", // material, low mass
		"reaudit:gap:G-1", // not material, high mass
		"reaudit:gap:G-5",
		"finding_on_gap:gap:G-4",
	}
	if !slices.Equal(keys, want) {
		t.Fatalf("queue = %q\nwant    %q", keys, want)
	}
	g3 := q[2]
	if g3.Gap == nil || g3.Gap.ID != "G-3" || len(g3.Items) != 2 || g3.Items[0].Kind != ItemReaudit || g3.Items[1].Kind != ItemRegradeOwed {
		t.Errorf("the readied gap G-3 carries %+v, want its row and its reaudit and regrade items folded in", g3)
	}
	if !q[1].Blocks || q[2].Blocks {
		t.Errorf("an entry blocks exactly when an item it carries does: %v %v", q[1].Blocks, q[2].Blocks)
	}
	if len(q[3].Items) != 0 || q[3].Items == nil {
		t.Errorf("a readied gap no item names carries %#v, want []", q[3].Items)
	}

	// THE AUDIT LEAVES THE QUEUE ONCE THE SITTING IS SERVED THE REPORT, and the digest of an entry
	// moves exactly when what it hands the seat moves.
	w.Sitting.LastRead.Kind = "served"
	q2 := QueueOf(w)
	if q2[0].Key == "audit:report:9" {
		t.Error("a served lens still has the audit queued")
	}
	if q2[1].Digest != q[2].Digest {
		t.Error("the digest of an unchanged entry moved")
	}
	w.Open[2].EditedSince = []GapEdit{{}}
	if QueueOf(w)[1].Digest == q[2].Digest {
		t.Error("the digest did not move when the entry's gap row did")
	}
}

// EVERY ITEM THE SEAT OWES IS QUEUED. An owed item off the queue is one the seat is never handed,
// and its envelope is refused over it.
func TestEveryOwedItemIsQueued(t *testing.T) {
	w := WorkJSON{Sitting: SittingJSON{Open: []Item{
		{Kind: ItemRegisterOwed, Blocks: true, What: "register"},
		{Kind: ItemKind(BlockerClosingOwed), Subject: about(SubjectGap, "G-1"), Blocks: true, What: "closing"},
		{Kind: ItemKind(BlockerClosingOwed), Subject: about(SubjectGap, "G-1"), Blocks: true, What: "closing again"},
		{Kind: ItemKind(BlockerMaterialGap), Subject: about(SubjectGap, "G-2"), Blocks: true, AnotherSeatClears: true, What: "board"},
	}}}
	assertOwedQueued(t, w)
	q := QueueOf(w)
	if len(q) != 3 || q[1].Key == q[2].Key {
		t.Errorf("queue %+v: want the three owed items, each under its own key", q)
	}
}

// assertOwedQueued fails where an item of w's Owed() is on no entry of QueueOf(w).
func assertOwedQueued(t *testing.T, w WorkJSON) {
	t.Helper()
	owed, _ := w.Sitting.Owed()
	q := QueueOf(w)
	for _, it := range owed {
		found := false
		for _, e := range q {
			for _, qi := range e.Items {
				if qi.What == it.What && qi.Kind == it.Kind {
					found = true
				}
			}
		}
		if !found {
			t.Errorf("owed item %q (%s) is not queued", it.What, it.Kind)
		}
	}
}

// EVERY KIND IS ONE WORD, AND NONE IS A SECOND WORD FOR A BLOCKER.
func TestTheItemKindsAreDistinct(t *testing.T) {
	seen := map[ItemKind]bool{}
	for _, k := range ItemKinds() {
		if k == "" || seen[k] {
			t.Errorf("item kind %q is empty or listed twice", k)
		}
		seen[k] = true
	}
	for _, r := range blockerKinds {
		if r.subject == "" {
			t.Errorf("blocker kind %s names no subject kind", r.kind)
		}
	}
}
