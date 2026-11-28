package blue

import (
	"fmt"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// THE FREEZE'S EDIT ARM (#1695 §III.B5, L45): every statement of text standing on a gap — red's
// closing's road, from the closing until the answer of the docketing it argues — holds until its
// ruling, so the bench's answer for that side applies it. Blue's edit and retire re-plan each standing
// road on the report they leave, with the accept's own planner (PlanAccept), and an act after which one
// no longer plans is refused, naming the docketing: an edit over the span, and an edit elsewhere that
// makes the quote ambiguous, alike. An act that leaves every road planning lands. The freeze holds the
// parties' acts, never the bench's application at the docket's first: a statement that application
// breaks is checked again at its own docketing's ruling (§III.B4, the docket order).
//
// THE FREEZE HOLDS WHAT THE ACT WOULD BREAK, NOT WHAT IS ALREADY BROKEN: a road that does not plan on
// the report the act lands on — the bench's application of another gap's statement took its span — is
// no statement this act can make inapplicable, and refusing on it froze every edit and retire of blue's
// until that docketing's ruling, the remedies its own work list held it to included.

// roadsHold is the freeze's guard on every act of blue's that can break a standing statement — its edit,
// retire, cite and compute: the roads that plan on the report the act lands on, read before it, and the
// refusal of an act after which one of them no longer plans, asked of the record with the act on it.
// Each call is one act's pair of guards.
func roadsHold(remedy string) []record.Guard {
	var planned map[string]bool
	before := record.Check(func(l *record.Landing) error {
		var err error
		planned, err = plannedRoads(l)
		return err
	})
	after := record.CheckLanded(func(l *record.Landing) error {
		return requireRoadsHold(l, planned, remedy)
	})
	return []record.Guard{before, after}
}

// The remedies the freeze's refusal names, per act: what the refused seat may do instead.
const (
	editRemedy     = "edit elsewhere, or argue in your closing on %[1]s"
	evidenceRemedy = "place it once %[2]s is ruled, or argue in your closing on %[1]s"
)

// roadKey names a standing road: its docketing or proposal and its gap.
func roadKey(r record.Road) string { return r.By + "\x00" + r.GapID }

// plannedRoads is the standing roads that plan on the report l renders.
func plannedRoads(l *record.Landing) (map[string]bool, error) {
	roads, err := l.StandingRoads()
	if err != nil || len(roads) == 0 {
		return nil, err
	}
	report, err := l.Report()
	if err != nil {
		return nil, err
	}
	out := map[string]bool{}
	for _, r := range roads {
		if roadPlans(l, report, r) == nil {
			out[roadKey(r)] = true
		}
	}
	return out, nil
}

// requireRoadsHold refuses where a standing road that planned before the act (planned) no longer plans
// on the report l renders. remedy is what the refused seat may do instead, formatted with the road's
// gap and the docketing or proposal that holds it.
func requireRoadsHold(l *record.Landing, planned map[string]bool, remedy string) error {
	roads, err := l.StandingRoads()
	if err != nil || len(roads) == 0 {
		return err
	}
	report, err := l.Report()
	if err != nil {
		return err
	}
	for _, r := range roads {
		if !planned[roadKey(r)] {
			continue
		}
		if err := roadPlans(l, report, r); err != nil {
			if r.Retire != "" {
				return feov.Wrap(feov.Conflict, err,
					"%s stands on gap %s and holds red's retire of %s, stated in the closing argued on it: this act makes a computation in the report apply %s's method, and a retire of a citation a computation applies is refused (L6), so the statement no longer applies to the report it leaves. "+
						"Every statement standing on %s holds until its ruling, so nothing was recorded; %s",
					r.By, r.GapID, r.Retire, r.Retire, r.GapID, fmt.Sprintf(remedy, r.GapID, r.By))
			}
			what := "red's text in the closing argued on it"
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

// roadPlans is nil where road r still plans on report: a pair as the accept applies it, or a retire no
// computation in the report applies the method of (L6, as the retire itself is held). AN ACT THAT DOES
// WHAT RED'S TEXT DOES LEAVES THE ROAD STANDING: a pair road stands where an edit replaced exactly its
// quoted words with exactly its own and the report still holds them at the gap's place
// (record.RedTextApplied); the answer for red then applies nothing more (record.requireRoadApplied).
func roadPlans(l *record.Landing, report string, r record.Road) error {
	if r.Retire != "" {
		live, err := l.ComputationsApplying(r.Retire)
		if err != nil || len(live) == 0 {
			return err
		}
		return feov.Errorf(feov.Conflict, "computation %s applies %s's method", strings.Join(live, ", "), r.Retire)
	}
	_, err := PlanAccept(report, r.GapID, r.Location, r.FixNew)
	if err != nil {
		if done, derr := l.RedTextApplied(report, r.GapID, r.Location, r.FixNew); derr != nil || done {
			return derr
		}
	}
	return err
}
