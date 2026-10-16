package record

import (
	"errors"
)

// Guard is a verb's own check on a write, asked of the record the write lands on (Landing) inside the
// write's transaction, before the write path's validation. It refuses by returning an error, which
// rolls the transaction back with nothing written, and it may fill fields of the body the verb is
// writing from what it reads there — a fact derived from the record is derived from the record the
// event joins.
//
// A VERB'S CHECK MADE BEFORE Append IS A CHECK OF A RECORD ANOTHER SEAT CAN MOVE (#1252). Seats write
// in parallel: an edit validated against the report as it stood, then appended after another seat's
// anchor landed inside its span, is an edit replay cannot apply, and every render of the report fails
// from then on. A guard runs under the write lock the insert holds, so what it admits is the record
// the event lands on.
type Guard func(*Landing) error

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
	// report is the report rendered once from the transaction's view and kept for the guards after
	// it: nothing a guard does writes, so the render cannot move under them.
	report    string
	reportErr error
	rendered  bool
}

// guardsAt runs guards, in order, against w.
func guardsAt(w writeView, guards []Guard) error {
	if len(guards) == 0 {
		return nil
	}
	l := &Landing{w: w}
	for _, g := range guards {
		if err := g(l); err != nil {
			return err
		}
	}
	return nil
}

// Report is the report as the record the write lands on renders it.
func (l *Landing) Report() (string, error) {
	if !l.rendered {
		l.report, l.reportErr = renderProjection(ReportProjectionAt(l.w.q))
		l.rendered = true
	}
	return l.report, l.reportErr
}

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

// Proposal is record.Proposal asked of the landing record, over report.
func (l *Landing) Proposal(gapID, report string) (old, fixNew string, found bool, err error) {
	return proposalAt(l.w.q, gapID, report)
}

// ProposalAppliedVerbatim is record.ProposalAppliedVerbatim asked of the landing record.
func (l *Landing) ProposalAppliedVerbatim(gapID, report, old, new string) (bool, error) {
	return proposalAppliedVerbatimAt(l.w.q, gapID, report, old, new)
}

// ClaimAppearsInAnEdit is record.ClaimAppearsInAnEdit asked of the landing record.
func (l *Landing) ClaimAppearsInAnEdit(claim string) bool {
	return claimAppearsInAnEditAt(l.w.q, claim)
}

// EditSpans is every recorded edit's (old, new) on the landing record, in record order.
func (l *Landing) EditSpans() ([]EditSpan, error) { return editSpansAt(l.w.q) }

// UnplacedLocation is record.UnplacedLocation asked of the landing record.
func (l *Landing) UnplacedLocation(id string) (string, error) { return unplacedLocationAt(l.w.q, id) }

// ClassCoinedInRun is record.ClassCoinedInRun asked of the landing record.
func (l *Landing) ClassCoinedInRun(slug string) bool { return classCoinedAt(l.w.q, slug) }

// UntilTheOutcome is the guard of a write the run's end closes — a log entry the tool writes for a
// seat: on a record that holds an outcome it writes nothing (ErrWritesNothing). Asked under the write
// lock, because the bench's outcome can commit between a check made before it and the entry, which
// then lands after the documents the run assembled.
func UntilTheOutcome(l *Landing) error {
	v, err := l.Outcome()
	if err != nil {
		return err
	}
	if v != "" {
		return ErrWritesNothing
	}
	return nil
}
