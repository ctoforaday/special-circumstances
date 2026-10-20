package record

import (
	"errors"
)

// Guard is a verb's own check on a write, asked of the record the write lands on (Landing) inside the
// write's transaction, before the write path's validation. It refuses by returning an error, which
// rolls the transaction back with nothing written, and it may fill fields of the body the verb is
// writing from what it reads there — a fact derived from the record is derived from the record the
// event joins. Check makes one; CheckOnReport makes one that reads the report.
//
// A VERB'S CHECK MADE BEFORE Append IS A CHECK OF A RECORD ANOTHER SEAT CAN MOVE (#1252). Seats write
// in parallel: an edit validated against the report as it stood, then appended after another seat's
// anchor landed inside its span, is an edit replay cannot apply, and every render of the report fails
// from then on. A guard runs under the write lock the insert holds, so what it admits is the record
// the event lands on.
type Guard struct {
	ask func(*Landing) error
	// readsReport is the guard's declaration that it reads Landing.Report: the write renders the report
	// before it takes the lock (landingReport), and the guard is handed that render wherever no other
	// write has moved the report since.
	readsReport bool
	// last is the guard's declaration that it is asked AFTER the write path's validation (CheckLast).
	last bool
}

// Check is a guard that does not read the report.
func Check(ask func(*Landing) error) Guard { return Guard{ask: ask} }

// CheckOnReport is a guard that reads the report (Landing.Report). The declaration is what keeps the
// render out of the lock: a write whose guards declared none, asked for the report, is refused as an
// internal error rather than rendering it under the lock.
func CheckOnReport(ask func(*Landing) error) Guard { return Guard{ask: ask, readsReport: true} }

// CheckLast is a guard asked after the write path's own validation rather than before it: one whose
// refusal tells the seat to record the act again, which is false of an act the write path refuses for
// itself — a motion on a gap already disposed of is refused whatever the seat reads.
func CheckLast(ask func(*Landing) error) Guard { return Guard{ask: ask, last: true} }

// ErrWritesNothing is a guard's answer that the write is not owed on the record it would land on — the
// run is over, the anchor is already placed. The transaction rolls back with nothing written, and the
// caller that asked for the guard decides what "nothing" means to its seat.
var ErrWritesNothing = errors.New("the record this write would land on does not take it, so nothing was written")

// Landing is the record as a write lands on it: the writing transaction and the run's files. It is
// what a Guard asks, and every answer is the transaction's.
//
// IT HOLDS NO Run, for writeView's reason: the run's handle is one connection and the transaction
// holds it, so a read through the handle from inside would wait forever. A reader a guard needs is
// added here in its Querier form, never by handing the guard the run.
type Landing struct {
	w writeView
}

// guardsAt runs the guards whose last is last, in order, against w.
func guardsAt(w writeView, guards []Guard, last bool) error {
	if len(guards) == 0 {
		return nil
	}
	l := &Landing{w: w}
	for _, g := range guards {
		if g.last != last {
			continue
		}
		if err := g.ask(l); err != nil {
			return err
		}
	}
	return nil
}

// guardsReadReport is whether any of guards declared that it reads the report.
func guardsReadReport(guards []Guard) bool {
	for _, g := range guards {
		if g.readsReport {
			return true
		}
	}
	return false
}

// Report is the report as the record the write lands on renders it: the render made before the lock
// where no write has moved the report since, else a render under it (landingReport). A guard asks it
// only when it was made by CheckOnReport.
func (l *Landing) Report() (string, error) { return l.w.report() }

// Family is FamilyOf asked of the landing record.
func (l *Landing) Family() (Family, error) { return familyAt(l.w.q) }

// Verdict is DeriveVerdict asked of the landing record.
func (l *Landing) Verdict() (verdict, why string, ok bool, err error) {
	r, err := verdictReadsAt(l.w.q)
	if err != nil {
		return "", "", false, err
	}
	return verdictOf(l.w.files, r)
}

// Outcome is RecordedOutcome asked of the landing record.
func (l *Landing) Outcome() (string, error) { return recordedOutcomeAt(l.w.q) }

// Prescription is the location and text gapID's mint recorded, as the landing record holds them.
func (l *Landing) Prescription(gapID string) (loc, fix string, found bool, err error) {
	return prescriptionAt(l.w.q, gapID)
}

// ProposalAppliedVerbatim is record.ProposalAppliedVerbatim asked of the landing record.
func (l *Landing) ProposalAppliedVerbatim(gapID, report, old, new string) (bool, error) {
	return proposalAppliedVerbatimAt(l.w.q, gapID, report, old, new)
}

// EditSpans is every recorded edit's (old, new) on the landing record, in record order.
func (l *Landing) EditSpans() ([]EditSpan, error) { return editSpansAt(l.w.q) }

// UnplacedLocation is record.UnplacedLocation asked of the landing record.
func (l *Landing) UnplacedLocation(id string) (string, error) { return unplacedLocationAt(l.w.q, id) }

// UntilTheOutcome is the guard of a write the run's end closes — a log entry the tool writes for a
// seat: on a record that holds an outcome it writes nothing (ErrWritesNothing). Asked under the write
// lock, because the bench's outcome can commit between a check made before it and the entry, which
// then lands after the documents the run assembled.
var UntilTheOutcome = Check(func(l *Landing) error {
	v, err := l.Outcome()
	if err != nil {
		return err
	}
	if v != "" {
		return ErrWritesNothing
	}
	return nil
})
