package blue

import (
	"fmt"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchortext"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// THE FREEZE'S EDIT ARM (#1695 §III.B14, L45): every statement of text standing on a gap — red's
// closing's road, from the closing until the answer of the docketing or proposal it argues, and a
// standing supersede proposal's stored quote and fix — holds until its ruling, so the bench's answer
// for that side applies it. Blue's edit and retire, and the text the bench's answer for red applies
// (the edit or retire of another docketing's road, in its ruling's transaction), re-plan each standing
// road on the report they leave, with the accept's own planner (PlanAccept), and an act after which one
// no longer plans is refused, naming the docketing or proposal: an edit over the span, and an edit
// elsewhere that makes the quote ambiguous, alike. An act that leaves every road planning lands.

// roadsHold is the freeze's guard on blue's edit and retire, asked of the record with blue's act on it.
var roadsHold = record.CheckLanded(func(l *record.Landing) error {
	return requireRoadsHold(l, "edit elsewhere, or argue in your closing on %[1]s")
})

// BenchRoadsHold is the freeze's guard on the text the bench's answer applies, asked of the record with
// the applied act and its ruling on it: the ruled docketing's own road has ended, and every other road
// standing must still plan.
var BenchRoadsHold = record.CheckLanded(func(l *record.Landing) error {
	return requireRoadsHold(l, "answer this docketing for the other side, or rule %[2]s first")
})

// requireRoadsHold refuses where a standing road no longer plans on the report l renders. remedy is
// what the refused seat may do instead, formatted with the road's gap and the docketing or proposal
// that holds it.
func requireRoadsHold(l *record.Landing, remedy string) error {
	roads, err := l.StandingRoads()
	if err != nil || len(roads) == 0 {
		return err
	}
	report, err := l.Report()
	if err != nil {
		return err
	}
	for _, r := range roads {
		if err := roadPlans(l, report, r); err != nil {
			what := "red's text in the closing argued on it"
			if r.Successor != "" {
				what = "the quote of the successor " + r.Successor + " it proposes"
			}
			// THE PLANNER'S OWN REFUSAL IS WORDED FOR A SEAT TYPING A PAIR, which the refused seat is not:
			// the refusal names the span the statement quotes and what this act did to it.
			return feov.Wrap(feov.Conflict, err,
				"%s stands on gap %s and holds %s, which quotes %q: this act changes that span, or makes it ambiguous, so the statement no longer applies to the report it leaves. "+
					"Every statement standing on %s holds until its ruling, so nothing was recorded; %s",
				r.By, r.GapID, what, r.Location, r.GapID, fmt.Sprintf(remedy, r.GapID, r.By))
		}
	}
	return nil
}

// roadPlans is nil where road r still plans on report: a pair as the accept applies it — a successor's
// with its own marker placed at its quote, as `lens mint` verifies a fix — or a retire. AN ACT THAT DOES
// WHAT RED'S TEXT DOES LEAVES THE ROAD STANDING: a retire road always stands, and a pair road stands
// where an edit replaced exactly its quoted words with exactly its own and the report still holds them
// at the gap's place (record.RedTextApplied); the answer for red
// then applies nothing more (record.requireRoadApplied).
func roadPlans(l *record.Landing, report string, r record.Road) error {
	switch {
	case r.Retire != "":
		return nil
	case r.Successor != "":
		return SuccessorPlans(report, r.Successor, r.Location, r.FixNew)
	}
	_, err := PlanAccept(report, r.GapID, r.Location, r.FixNew)
	if err != nil {
		if done, derr := l.RedTextApplied(report, r.GapID, r.Location, r.FixNew); derr != nil || done {
			return derr
		}
	}
	return err
}

// SuccessorPlans is nil where a proposed successor's quote places its marker on report and its fix, if
// it states one, is a pair the accept applies there — the questions `lens mint`'s Places guard and its
// verified fix ask of the mint the bench writes (L48).
func SuccessorPlans(report, successor, loc, fix string) error {
	placed, err := anchortext.Attach(report, successor, loc)
	if err != nil {
		return anchortext.Refusal(err)
	}
	if fix == "" {
		return nil
	}
	_, err = PlanAccept(placed, successor, loc, fix)
	return err
}
