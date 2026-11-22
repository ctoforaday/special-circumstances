package blue

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchortext"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// THE FREEZE'S EDIT ARM (#1695 §III.B14, L45): every statement of text standing on a gap — red's
// closing's road, from the closing until the answer of the docketing or proposal it argues, and a
// standing supersede proposal's stored quote and fix — holds until its ruling, so the bench's answer
// for that side applies it. Blue's edit and retire re-plan each standing road on the report they leave,
// with the accept's own planner (PlanAccept), and an act after which one no longer plans is refused,
// naming the docketing or proposal: an edit over the span, and an edit elsewhere that makes the quote
// ambiguous, alike. An act that leaves every road planning lands.

// roadsHold is the freeze's guard, asked of the record with blue's act on it.
var roadsHold = record.CheckLanded(requireRoadsHold)

// requireRoadsHold refuses where a standing road no longer plans on the report l renders.
func requireRoadsHold(l *record.Landing) error {
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
			return feov.Wrap(feov.Conflict, err,
				"%s stands on gap %s and holds %s, which no longer applies to the report this act leaves: %v. "+
					"Every statement standing on %s holds until its ruling, so nothing was recorded; edit elsewhere, or argue in your closing on %s",
				r.By, r.GapID, what, err, r.GapID, r.GapID)
		}
	}
	return nil
}

// roadPlans is nil where road r still plans on report: a pair as the accept applies it — a successor's
// with its own marker placed at its quote, as `lens mint` verifies a fix — or a retire whose anchor no
// retire has taken.
func roadPlans(l *record.Landing, report string, r record.Road) error {
	switch {
	case r.Retire != "":
		live, err := l.AnchorLive(r.Retire)
		if err != nil || live {
			return err
		}
		return feov.Errorf(feov.Conflict, "its road retires %s, which is retired", r.Retire)
	case r.Successor != "":
		return SuccessorPlans(report, r.Successor, r.Location, r.FixNew)
	}
	_, err := PlanAccept(report, r.GapID, r.Location, r.FixNew)
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
