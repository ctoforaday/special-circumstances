package record

import (
	"encoding/json"
	"fmt"
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

// THE AUTHOR'S ACTS ARE ITS HALF OF THE EXCHANGE, AND ITS TURN COMES FIRST AT EVERY READER (X17, T-B47):
// on G2, a gap about G1, the author is G1's minter, a lens. Its regrade of G1 naming G2 is its act on G2
// (Regrade.answers): the raiser's turn comes (lensTurnOn), a dispatch of the raiser after it is not
// early (EarlyLensDispatches), the round the raiser's audit ends is one the author acted in and moved,
// and after a remand of G2 it answers the remand. At impasse the plan dockets G2 awaiting the author's
// closing, not blue's. Each reader keyed on blue's seat instead reads none of this.
func TestTheAuthorsActIsItsHalfOfTheExchange(t *testing.T) {
	const author, raiser = evLens, outsideLens
	authorRegrades := func(b *stage) *stage {
		agent := fmt.Sprintf("author-acts-%d", b.n)
		return b.register("red-chair").dispatch(2, author, "G2").registerAs(author, agent).
			add(author, &recordpb.Regrade{GapId: proto.String("G1"), Answers: proto.String("G2"),
				Severity: recordpb.Grade_GRADE_HIGH.Enum(), Basis: proto.String("G2 is right: the grade understated it")}).stop(agent)
	}
	t.Run("the raiser's turn and its audit", func(t *testing.T) {
		acted := authorRegrades(authorStage(t))
		x, err := Exchanges(acted.register("red-chair").seed(), DefaultParams)
		if err != nil {
			t.Fatal(err)
		}
		if g := x["G2"]; g == nil || !g.lensTurn || !g.pending.acted {
			t.Fatalf("after the author's act on G2 the raiser's turn has not come, or the open round reads no act: %+v", g)
		}
		dispatched := acted.more().register("red-chair").dispatch(2, raiser, "G2")
		early, err := EarlyLensDispatches(dispatched.seed())
		if err != nil {
			t.Fatal(err)
		}
		if slices.ContainsFunc(early, func(d EarlyLensDispatch) bool { return d.GapID == "G2" }) {
			t.Errorf("capture names a dispatch on G2 early, after the author's act: %+v", early)
		}
		x, err = Exchanges(authorRegrades(authorStage(t)).register("red-chair").dispatch(2, raiser, "G2").sitClosed(raiser).register("red-chair").seed(), DefaultParams)
		if err != nil {
			t.Fatal(err)
		}
		if g := x["G2"]; g == nil || g.Exchanges != 1 || g.Stalled != 0 || len(g.rounds) != 1 || !g.rounds[0].acted || !g.rounds[0].audited {
			t.Errorf("the raiser's audit after the author's act ends one round the author acted in and moved: %+v", g)
		}
	})
	t.Run("the remand's exchange", func(t *testing.T) {
		st := authorStage(t)
		for range 2 {
			st.register("red-chair").dispatch(2, author, "G2").sitClosed(author).dispatch(2, raiser, "G2").sitClosed(raiser)
		}
		st.docketAndRemandWith("M1", "G2", remandDirection)
		x, err := Exchanges(authorRegrades(st).register("red-chair").seed(), DefaultParams)
		if err != nil {
			t.Fatal(err)
		}
		if g := x["G2"]; g == nil || !g.answeredRemand {
			t.Errorf("the author's act on G2 after its remand does not answer the remand: %+v", g)
		}
	})
	t.Run("impasse awaits the author's closing", func(t *testing.T) {
		st := authorStage(t)
		for range 2 {
			st.register("red-chair").dispatch(2, author, "G2").sitClosed(author).dispatch(2, raiser, "G2").sitClosed(raiser)
		}
		plan := planOfT(t, st.register("red-chair").seed())
		if !whyMentions(plan, "G2: at impasse") {
			t.Fatalf("G2 is not docketed at impasse: why %q", plan.Why)
		}
		if _, ok := partyOf(plan, author); !ok || !whyMentions(plan, closingsAwaitedWhy("G2", []string{raiser, author})) {
			t.Errorf("the docketing of G2 does not await its raiser's and its author's closings: parties %+v, why %q", plan.Parties, plan.Why)
		}
	})
}
