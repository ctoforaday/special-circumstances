package record

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// mintOnGap is raiser's gap gap about the open gap about (ABOUT_KIND_GAP): a defect in about's own
// wording or grades, whose author is about's minter. Seeded, as no mint writes one before the gap on a
// gap lands (requireMintSubject).
func (b *stage) mintOnGap(raiser, gap, about string) *stage {
	return b.add(raiser, &recordpb.Mint{AboutKind: recordpb.AboutKind_ABOUT_KIND_GAP.Enum(), AboutRef: proto.String(about), GapId: proto.String(gap),
		Class: recordpb.GapClass_GAP_CLASS_SPEC_UNDERSPECIFICATION.Enum(), Problem: proto.String("p"), AcceptanceCheck: proto.String("c"),
		CheckKind: recordpb.CheckKind_CHECK_KIND_DOCUMENT.Enum(), Severity: recordpb.Grade_GRADE_MEDIUM.Enum(),
		Likelihood: recordpb.Grade_GRADE_MEDIUM.Enum(), Impact: recordpb.Grade_GRADE_MEDIUM.Enum()})
}

// authorStage is a run whose evidence lens minted G1 about a section of the report, and whose logic
// lens minted G2 about G1: G1's author is blue-respond and G2's is the evidence lens, G1's minter.
func authorStage(t *testing.T) *stage {
	return newStage(t).cast(evLens, outsideLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").sit(2, evLens).sit(2, outsideLens).
		mint(evLens, "G1", "medium").mintOnGap(outsideLens, "G2", "G1")
}

// T-B47: THE AUTHOR IS THE SUBJECT'S AUTHOR (authorOf), at every party site, per subject kind: the
// closing parties a docketing owes and whose closing the write admits, whose closing states the
// author's word and may not carry a road, the seat the dispatch engages on the gap at the author's
// turn and that it awaits a closing from, whose acts the exchange counts as the author's half, whose
// row says it is its turn, and which side a closing renders on. On a report subject that is
// blue-respond; on a gap about a gap it is that gap's minter, and blue is no party.
func TestTheAuthorIsTheSubjectsAuthor(t *testing.T) {
	for _, c := range []struct {
		name, gap, raiser, author, outsider string
	}{
		{"a report subject", "G1", evLens, "blue-respond", outsideLens},
		{"a gap about a gap", "G2", outsideLens, evLens, "blue-respond"},
	} {
		t.Run(c.name, func(t *testing.T) {
			if got, err := AuthorOfGap(authorStage(t).seed(), c.gap); err != nil || got != c.author {
				t.Fatalf("AuthorOfGap(%s) = %q, %v; want %s", c.gap, got, err, c.author)
			}

			// THE DISPATCH ENGAGES THE AUTHOR at its turn on a fresh gap, and nobody else on it.
			plan := planOfT(t, authorStage(t).register("red-chair").seed())
			if p, ok := partyOf(plan, c.author); !ok || !slices.Contains(p.GapIDs, c.gap) {
				t.Errorf("the plan does not engage %s, %s's author, on it: %+v", c.author, c.gap, plan.Parties)
			}
			if p, ok := partyOf(plan, c.outsider); ok && slices.Contains(p.GapIDs, c.gap) {
				t.Errorf("the plan engages %s, no party to %s, on it: %+v", c.outsider, c.gap, plan.Parties)
			}
			// AND ITS ROW SAYS IT IS THE AUTHOR'S TURN, on the author's list and on nobody else's.
			run := authorStage(t).register("red-chair").seed()
			for seat, want := range map[string]bool{c.author: true, c.outsider: false} {
				role := "blue"
				if seat != "blue-respond" {
					role = "lens"
				}
				w, err := WorkOfSeat(run, role, seat)
				if err != nil {
					t.Fatal(err)
				}
				for _, row := range w.Open {
					if row.ID == c.gap && row.YourTurn != want {
						t.Errorf("%s's row on %s says your_turn %v, want %v", seat, c.gap, row.YourTurn, want)
					}
				}
			}

			// THE AUTHOR'S SITTING ON THE GAP IS THE AUTHOR'S HALF OF THE EXCHANGE: dispatched on it, it
			// sits and acts nothing, its lens audits, and one exchange is counted.
			ex := authorStage(t).register("red-chair").dispatch(2, c.author, c.gap).sitClosed(c.author).
				register("red-chair").dispatch(2, c.raiser, c.gap).sitClosed(c.raiser).register("red-chair").seed()
			evs, win := mustEvents(t, ex)
			sittings := gapSittingsOf(evs, win.IDs(evs), win, AfterTheRun, assigneeOf(evs))[c.gap]
			if !slices.ContainsFunc(sittings, func(s partySitting) bool { return !s.red }) {
				t.Errorf("%s's sitting on %s is not read as the author's: %+v", c.author, c.gap, sittings)
			}

			// A DOCKETING OWES ITS CLOSINGS FROM THE RAISER AND THE AUTHOR.
			docketed := func() *stage { return authorStage(t).register("red-chair").docketMotion("", "D1", c.gap) }
			if got := closingOwners(t, docketed().seed()); !slices.Equal(got, []string{c.raiser, c.author}) {
				t.Errorf("closings on %s owed by %v, want %s and %s", c.gap, got, c.raiser, c.author)
			}
			plan = planOfT(t, docketed().seed())
			for _, seat := range []string{c.raiser, c.author} {
				if _, ok := partyOf(plan, seat); !ok {
					t.Errorf("%s owes its closing on %s and the plan does not ready it: %+v", seat, c.gap, plan.Parties)
				}
			}
			// THE WRITE ADMITS THE AUTHOR'S CLOSING and refuses the outsider's; the author's closing
			// states a word and carries no road.
			run = docketed().seed()
			closingOn := func(text string) *recordpb.Closing {
				return &recordpb.Closing{GapId: proto.String(c.gap), Text: proto.String(text)}
			}
			if err := validateRun(run, c.author, recordpb.EventType_EVENT_TYPE_CLOSING, closingOn("the author's case")); err != nil {
				t.Errorf("%s is %s's author and its closing is refused: %v", c.author, c.gap, err)
			}
			if err := validateRun(run, c.outsider, recordpb.EventType_EVENT_TYPE_CLOSING, closingOn("no party's case")); err == nil || !strings.Contains(err.Error(), "is party to none") {
				t.Errorf("%s is no party to %s: want its closing refused, got %v", c.outsider, c.gap, err)
			}
			road := closingOn("the author's text")
			road.NoText = proto.Bool(true)
			if err := validateRun(run, c.author, recordpb.EventType_EVENT_TYPE_CLOSING, road); err == nil || !strings.Contains(err.Error(), "states its word, never text") {
				t.Errorf("%s's closing on %s stating a road = %v; want refused, the author states a word", c.author, c.gap, err)
			}

			// THE AUTHOR'S CLOSING STATES THE AUTHOR'S WORD, and renders on the author's side.
			word := recordpb.Disposition_DISPOSITION_NOT_A_DEFECT
			argued := docketed().add(c.author, &recordpb.Closing{GapId: proto.String(c.gap), Text: proto.String("the author's case"), ClosureClass: &word}).
				add(c.raiser, &recordpb.Closing{GapId: proto.String(c.gap), Text: proto.String("the raiser's case"), NoText: proto.Bool(true)}).seed()
			pos, found, err := PositionsOf(argued, "D1")
			if err != nil || !found {
				t.Fatalf("positions of D1: %v (found %v)", err, found)
			}
			if pos.BlueWord == nil || *pos.BlueWord != word {
				t.Errorf("the author's word on %s reads %v, want %s", c.gap, pos.BlueWord, recordpb.Word(word))
			}
			b, err := DebateJSONBytes(argued)
			if err != nil {
				t.Fatal(err)
			}
			var debate struct {
				Epochs []struct {
					RedClosings  []DebateClosingJSON `json:"red_closings"`
					BlueClosings []DebateClosingJSON `json:"blue_closings"`
				} `json:"epochs"`
			}
			if err := json.Unmarshal(b, &debate); err != nil {
				t.Fatal(err)
			}
			var blue, red []string
			for _, e := range debate.Epochs {
				for _, cl := range e.BlueClosings {
					blue = append(blue, cl.Text)
				}
				for _, cl := range e.RedClosings {
					red = append(red, cl.Text)
				}
			}
			if !slices.Equal(blue, []string{"the author's case"}) || !slices.Equal(red, []string{"the raiser's case"}) {
				t.Errorf("the debate renders the author's closings %q and the raiser's %q; want each on its own side", blue, red)
			}
		})
	}
}
