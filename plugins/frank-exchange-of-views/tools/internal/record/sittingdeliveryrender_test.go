package record_test

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/view"
)

// THE ANSWERS A RUN GIVES TO EVERY QUESTION ASKED OF ITS SITTINGS' ENDS AND ITS DELIVERIES, rendered
// through the calls that take a run: which sittings failed, how each sitting ended, the dispatch plan
// (each lens's readiness, pin and reason, the stale areas, the exchange counts), the PASS gate's list,
// each lens's last_read and work list, the closing duties, each delivery's line, each spot-check's
// stamped changes, and the exchange fold. The Go folds behind them read the agent_sitting and
// report_delivery rows the whole-record loaders hold; a fold that changes where it reads them answers
// here as it did, or the golden moves.
//
// NOTHING HERE READS THE LOADED ROWS OR A FOLD THAT TAKES THEM. Every call takes the run and returns
// what a seat, the engine or capture is handed, so the text is the same whichever statement answers
// underneath. ClosingDuties is the one exception, behind closingDutiesOfRun.
//
// WHAT IT COSTS IS WHOLE-RECORD READS, nine a run and one more for each work list, so the family is
// read once for the cast, the closing duties, the spot-checks and debate.md.
//
// THIS FILE IS HELD IN TWO TEST PACKAGES, internal/record (the seeded records) and
// internal/record/migrate (the archived runs), the same bytes after the package clause. A test file
// is importable by no other package and the fixtures' builders are package record's own, so neither
// package can render for the other; TestTheSittingAndDeliveryRendererIsOneText fails a copy that
// differs.

// The questions counted over what is rendered. A sweep that renders none of one is a sweep that
// would pass with its fold answering nothing.
const (
	qFailed          = "failed sittings"
	qNeverClosed     = "sittings not closed, and over"
	qStillOpen       = "sittings not closed, the seat's latest"
	qClosedLate      = "sittings closed after the seat's next opening"
	qReadyLens       = "lenses the plan readies"
	qStaleArea       = "stale areas"
	qLastReadPin     = "last_read pins above 0"
	qWholeOwed       = "whole_report_owed blockers"
	qUpheld          = "upheld_position blockers"
	qBlockers        = "blockers on the gate's list"
	qClosingDuty     = "closing duties"
	qClosingMissing  = "closing duties missing"
	qClosingRuled    = "closing duties on a ruled docketing"
	qDelivery        = "deliveries"
	qWholeDelivery   = "deliveries of the whole report"
	qChangesDelivery = "deliveries of the changes since a pin above 0"
	qUnsatDelivery   = "deliveries in a sitting not sat through"
	qNoSitDelivery   = "deliveries in no sitting"
	qStampedChange   = "spot-check areas with stamped changes"
	qExchange        = "gaps with a counted exchange"
	qRefused         = "calls refused"
)

// sittingDeliveryQuestions is every question, in the order the totals print.
var sittingDeliveryQuestions = []string{qFailed, qNeverClosed, qStillOpen, qClosedLate, qReadyLens, qStaleArea,
	qLastReadPin, "last_read first", "last_read behind", "last_read unchanged", "last_read served", "last_read undispatched",
	qWholeOwed, qUpheld, qBlockers, qClosingDuty, qClosingMissing, qClosingRuled, qDelivery, qWholeDelivery, qChangesDelivery,
	qUnsatDelivery, qNoSitDelivery, qStampedChange, qExchange, qRefused}

// closingDutiesOfRun is ClosingDuties asked of the run, whose family fam is.
func closingDutiesOfRun(_ record.Run, fam record.Family) []record.ClosingDuty {
	return record.ClosingDuties(fam.Events, fam.At)
}

// planLines is the plan, one line for each field and for each element of a field that is a list, in
// the order the plan declares them: every field it relays, whatever fields it has. Its blockers are
// one line where they are gate, the PASS gate's list, each as its kind, subject and owner. Its key
// is last: it is the plan's other fields hashed, so a plan that differs shows first in the field
// that moved.
func planLines(plan record.Plan, gate []record.Blocker) ([]string, error) {
	raw, err := json.Marshal(plan)
	if err != nil {
		return nil, err
	}
	relayed := make([]record.PlanBlocker, 0, len(gate))
	for _, b := range gate {
		relayed = append(relayed, record.PlanBlocker{Kind: b.Kind, Subject: b.Subject, Owner: b.Owner})
	}
	dec := json.NewDecoder(strings.NewReader(string(raw)))
	if _, err := dec.Token(); err != nil {
		return nil, err
	}
	var out []string
	var keyLine string
	for dec.More() {
		key, err := dec.Token()
		if err != nil {
			return nil, err
		}
		var value json.RawMessage
		if err := dec.Decode(&value); err != nil {
			return nil, err
		}
		var list []json.RawMessage
		switch {
		case key == "key":
			keyLine = fmt.Sprintf("plan key: %s", value)
		case key == "blockers" && len(gate) > 0 && slices.Equal(plan.Blockers, relayed):
			out = append(out, fmt.Sprintf("plan blockers: the gate's %d, each as its kind, subject and owner", len(gate)))
		case json.Unmarshal(value, &list) != nil || len(list) == 0:
			out = append(out, fmt.Sprintf("plan %s: %s", key, value))
		default:
			for _, el := range list {
				out = append(out, fmt.Sprintf("plan %s: %s", key, el))
			}
		}
	}
	return append(out, keyLine), nil
}

// sittingDeliveryAnswers renders the run's answers, one line each, and adds what it rendered to
// totals. A call the record refuses is a line saying so, counted: a sweep of refusals is no sweep of
// empty answers. inFull asks every cast lens's work list and prints each item's sentence; without
// it a list is asked only of a record that holds a delivery — on one that holds none no lens has a
// pin, and the plan's reason for each lens says how it stands — and an item is its kind and subject.
//
// RENDERING WRITES NOTHING, and a render that leaves the record holding another event count than it
// found is an error.
func sittingDeliveryAnswers(run record.Run, totals map[string]int, inFull bool) (string, error) {
	db, err := sql.Open("sqlite", "file:"+filepath.ToSlash(filepath.Join(run.Records(), "record.db"))+"?mode=ro")
	if err != nil {
		return "", err
	}
	defer db.Close()
	held := func() (n int, err error) {
		return n, db.QueryRow(`SELECT COUNT(*) FROM "events"`).Scan(&n)
	}
	before, err := held()
	if err != nil {
		return "", err
	}
	var b strings.Builder
	line := func(format string, args ...any) { fmt.Fprintf(&b, format+"\n", args...) }
	refused := func(what string, err error) {
		totals[qRefused]++
		line("%s: REFUSED: %s", what, strings.ReplaceAll(err.Error(), run.Dir(), "<run>"))
	}

	// How each sitting ended: the failed ones, then every sitting an agent opened, by seat.
	if failed, err := record.FailedSittings(run); err != nil {
		refused("failed sittings", err)
	} else {
		totals[qFailed] += len(failed)
		for _, f := range failed {
			line("failed sitting: %s sitting %d, epoch %d, opened at %d, agent %s", f.SeatID, f.Sitting, f.Epoch, f.SittingID, f.AgentID)
		}
	}
	if sittings, err := record.AgentSittings(run); err != nil {
		refused("agent sittings", err)
	} else {
		var seats []string
		ends := map[string][]string{}
		for _, s := range sittings {
			if _, seen := ends[s.SeatID]; !seen {
				seats = append(seats, s.SeatID)
			}
			flags := []byte("---")
			if s.Closed {
				flags[0] = 'C'
			}
			if s.Admitted {
				flags[1] = 'A'
			}
			if s.Over {
				flags[2] = 'O'
			}
			ends[s.SeatID] = append(ends[s.SeatID], fmt.Sprintf("%d:%s", s.Sitting, flags))
			switch {
			case !s.Closed && s.Over:
				totals[qNeverClosed]++
			case !s.Closed:
				totals[qStillOpen]++
			}
		}
		for _, seat := range seats {
			line("sittings %s: %s", seat, strings.Join(ends[seat], " "))
		}
	}

	// The PASS gate's list, then the dispatch plan, whole.
	blockers, err := record.PassBlockers(run)
	if err != nil {
		refused("blockers", err)
	}
	totals[qBlockers] += len(blockers)
	for _, bl := range blockers {
		line("blocker %s: subject %q, owner %q, detail %q, about %q, since %d", bl.Kind, bl.Subject, bl.Owner, bl.Detail, bl.About, bl.Since())
		switch bl.Kind {
		case record.BlockerWholeReportOwed:
			totals[qWholeOwed]++
		case record.BlockerUpheldPosition:
			totals[qUpheld]++
		}
	}
	if plan, err := record.PlanDispatch(run); err != nil {
		refused("plan", err)
	} else {
		lines, err := planLines(plan, blockers)
		if err != nil {
			return "", err
		}
		line("%s", strings.Join(lines, "\n"))
		for _, p := range plan.Parties {
			if seatclass.RoleOfSeat(p.SeatID) == "lens" {
				totals[qReadyLens]++
			}
		}
		totals[qStaleArea] += len(plan.StaleAreas)
	}

	// The family, read once: the cast, the closing duties, the spot-checks and debate.md are its.
	fam, err := record.FamilyOf(run)
	if err != nil {
		return "", err
	}
	var cast []string
	delivered := false
	for _, e := range fam.Events {
		if c := e.GetCast(); c != nil {
			cast = c.GetSeatIds()
		}
		delivered = delivered || e.GetReportRead() != nil
	}

	// Each cast lens's last_read and work list.
	for _, seat := range cast {
		if seatclass.RoleOfSeat(seat) != "lens" || !(inFull || delivered) {
			continue
		}
		w, err := record.WorkOfSeat(run, "lens", seat)
		if err != nil {
			refused("work "+seat, err)
			continue
		}
		lr := w.Sitting.LastRead
		line("last_read %s: %s, pin %d, head %d; complete %v", seat, lr.Kind, lr.Pin, lr.Head, w.Sitting.Complete)
		totals["last_read "+lr.Kind]++
		if lr.Pin > 0 {
			totals[qLastReadPin]++
		}
		for _, it := range w.Sitting.Open {
			about := "nothing"
			if it.Subject != nil {
				about = fmt.Sprintf("%s %s", it.Subject.Kind, it.Subject.ID)
			}
			if inFull {
				line("  item %s, about %s, blocks %v: %s", it.Kind, about, it.Blocks, it.What)
			} else {
				line("  item %s, about %s, blocks %v", it.Kind, about, it.Blocks)
			}
		}
	}

	// The closing duties.
	duties := closingDutiesOfRun(run, fam)
	totals[qClosingDuty] += len(duties)
	for _, d := range duties {
		state := string(d.State)
		if d.NotOwedWhy != "" {
			state += ", " + d.NotOwedWhy
		}
		line("closing duty: gap %s, motion %s (ruled %v), party %s: %s", d.GapID, d.Motion, d.Ruled, d.Party, state)
		if d.State == record.ClosingMissing {
			totals[qClosingMissing]++
		}
		if d.Ruled {
			totals[qClosingRuled]++
		}
	}

	// Each delivery: its line in debate.md, and its row in the debate view. The pin a lens's changes
	// were cut after is carried by the line's sentence alone, so that count is keyed on its words: a
	// line reworded counts none, and a sweep that counts none fails.
	if md, err := view.MarkdownFrom(view.Input{Gaps: fam.Gaps, Events: fam.Events, At: fam.At, Run: run}, "debate", ""); err != nil {
		refused("debate.md", err)
	} else {
		lines := strings.Split(string(md), "\n")
		for i, l := range lines {
			if !strings.HasPrefix(l, "### REPORT DELIVERED") || i+1 == len(lines) {
				continue
			}
			line("delivery line: %s", lines[i+1])
			totals[qDelivery]++
			if what := lines[i+1]; strings.HasPrefix(what, "the changes since head ") && !strings.HasPrefix(what, "the changes since head 0 ") {
				totals[qChangesDelivery]++
			}
		}
	}
	if raw, err := record.DebateJSONBytes(run); err != nil {
		refused("debate view", err)
	} else {
		var debate record.DebateJSON
		if err := json.Unmarshal(raw, &debate); err != nil {
			return "", err
		}
		for _, ep := range debate.Epochs {
			for _, s := range ep.Served {
				line("delivery row: epoch %d, %s at head %d, whole %v", ep.Epoch, s.SeatID, s.Head, s.Whole)
				if s.Whole {
					totals[qWholeDelivery]++
				}
			}
		}
	}

	// Each spot-check's stamp: the pin behind each area it named and the changes after it.
	for _, sc := range record.SpotChecks(fam) {
		line("spot-check, %s sitting %d, epoch %d: %s", sc.SeatID, sc.Sitting, sc.Epoch, record.DescribeAreas(sc.Areas))
		for _, a := range sc.Areas {
			if len(a.Changes) > 0 {
				totals[qStampedChange]++
			}
		}
	}

	// The exchange fold, which no failed sitting is a turn in, and the two audits that read it.
	if params, err := record.RunParams(run); err != nil {
		refused("run terms", err)
	} else if exchanges, err := record.Exchanges(run, params); err != nil {
		refused("exchanges", err)
	} else {
		gaps := make([]string, 0, len(exchanges))
		for g := range exchanges {
			gaps = append(gaps, g)
		}
		slices.Sort(gaps)
		for _, g := range gaps {
			x := exchanges[g]
			line("exchanges %s: %d, stalled %d, unresolved %d, impasse %v", g, x.Exchanges, x.Stalled, x.Unresolved, x.Impasse)
			if x.Exchanges > 0 {
				totals[qExchange]++
			}
		}
	}
	if early, err := record.EarlyLensDispatches(run); err != nil {
		refused("early lens dispatches", err)
	} else {
		for _, d := range early {
			line("early lens dispatch: at %d, %s on %s", d.At, d.SeatID, d.GapID)
		}
	}
	if early, err := record.EarlyDockets(run); err != nil {
		refused("early dockets", err)
	} else {
		for _, d := range early {
			line("early docket: at %d, motion %s on %s", d.At, d.Motion, d.GapID)
		}
	}

	// What the record's own views hold of the arms no struct above carries: counted, never rendered.
	var late, unsat, noSitting sql.NullInt64
	if err := db.QueryRow(`SELECT SUM("closed" AND NOT "sat_through") FROM "agent_sitting"`).Scan(&late); err != nil {
		return "", err
	}
	if err := db.QueryRow(`SELECT SUM(NOT "sat_through"), SUM("sitting_id" IS NULL) FROM "report_delivery"`).Scan(&unsat, &noSitting); err != nil {
		return "", err
	}
	totals[qClosedLate] += int(late.Int64)
	totals[qUnsatDelivery] += int(unsat.Int64)
	totals[qNoSitDelivery] += int(noSitting.Int64)
	if after, err := held(); err != nil || after != before {
		return "", fmt.Errorf("rendering the run's answers wrote to its record: %d events before, %d after (%v)", before, after, err)
	}
	return b.String(), nil
}
