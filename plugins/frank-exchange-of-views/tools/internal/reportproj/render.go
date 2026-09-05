package reportproj

import (
	"fmt"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchortext"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/bluedoc"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// THE REPORT IS THE RECORD, NOT A FILE (#709).
//
// The current report is the frozen base with every text-mutating event replayed over it, in
// record order. Five verbs mutate the report, and each replays through the SAME transform it
// applied at write time, located by the SAME rule:
//
//   - blue edit       → a splice (old→new), located by bluedoc.LocateUniqueReplacing (spliceMut).
//   - blue cite       → a citation marker spliced at the quoted sentence (insertMut).
//   - blue prove      → a proof marker, likewise.
//   - lens finding    → a finding marker, recorded as an Anchor event (its Finding sibling is
//                       metadata and inserts nothing — replaying it too would double the marker).
//   - lens corroborate→ a citation marker at the corroborated claim, on the Verify event (only
//                       when it backs the claim and so carries a c- Label; a plain verify does not).
//
// Because replay reproduces the identical running text each verb saw, the locators resolve the
// identical offsets, so the bytes match what the file used to hold. Once the base is ingested and
// the file deleted, this is the ONLY way the report exists: there is no file to raw-write,
// chattr, or bypass, and a change that is not a recorded event cannot exist.
//
// Every failure here is LOUD, never a plausible zero: an op that will not locate is a record that
// no longer describes a real sequence over this base, and a silently skipped op would render a
// report that never existed.

// ApplySplice is the pure text transform blue edit and replay share: replace report[start:end]
// with new, then tidy the two seams the splice created. NO legality guards — those (anchors
// transit, no dropped marker, present-and-unique) are write-time validation that already passed
// when the edit was recorded; replay reproduces bytes, it does not re-authorise.
func ApplySplice(report string, start, end int, new string) string {
	next := report[:start] + new + report[end:]
	// SPLICE HYGIENE (see splice.go): tidy the two seams this edit created. Deterministic, narrow,
	// silent — and it MUST run in replay too, or the replayed text drifts from what the file held.
	next, _ = tidySeam(next, start+len(new)) // trailing seam first — the later offset is stable
	next, _ = tidySeam(next, start)
	return next
}

// Op is one recorded splice from the diff-stack: the (old, new) of a BlueEdit event. Render takes
// plain strings rather than *recordpb.BlueEdit so callers testing the edit path need not build the
// record schema — the record-aware path is RenderFromRecord.
type Op struct{ Old, New string }

// Render reproduces a report from a base and an ordered splice stack. It is the edit-only surface
// blue edit's fidelity tests drive against planEdit; RenderFromRecord composes the full event
// stream (splices AND marker insertions).
func Render(base string, ops []Op) (string, error) {
	text := base
	for i, op := range ops {
		next, err := (spliceMut{old: op.Old, new: op.New}).apply(text)
		if err != nil {
			return "", fmt.Errorf("render: replaying edit %d of %d: %w", i+1, len(ops), err)
		}
		text = next
	}
	return text, nil
}

// mutation is one recorded text change, applied to the running report during replay. The two
// shapes are a splice (blue edit) and a marker insertion (cite/prove/finding); both locate against
// the running text exactly as the verb did, and both are loud when they cannot.
type mutation interface {
	apply(text string) (string, error)
	describe() string
}

type spliceMut struct{ old, new string }

func (m spliceMut) apply(text string) (string, error) {
	start, end, err := bluedoc.LocateUniqueReplacing("render", text, m.old)
	if err != nil {
		return "", err
	}
	return ApplySplice(text, start, end, m.new), nil
}

func (m spliceMut) describe() string { return fmt.Sprintf("edit %q→%q", m.old, m.new) }

type insertMut struct{ location, marker string }

func (m insertMut) apply(text string) (string, error) {
	// TORN-ANCHOR ADOPTION, REPLAYED. The verbs adopt an orphan marker rather than splice a
	// rival; if two events ever name the same marker (a crash-retry that appended twice), replay
	// must place it ONCE. Keyed on the exact marker token, so two findings on one sentence — two
	// distinct tokens — both still insert.
	if strings.Contains(text, m.marker) {
		return text, nil
	}
	next, err := anchortext.InsertAnchor([]byte(text), m.location, m.marker)
	if err != nil {
		return "", err
	}
	return string(next), nil
}

func (m insertMut) describe() string { return fmt.Sprintf("insert %s at %q", m.marker, m.location) }

// RenderFromRecord is the record-aware render: it reads the run's frozen base and the ordered
// stream of text mutations from the record — the report_op view SELECTS and orders them in SQL — and
// folds them over the base. This is the ONLY way to obtain the current report once the file is
// deleted; the former readers of blue/report.md call this instead.
//
// The FOLD is the irreducible sequential part: a splice must be located in the running text a prior
// op left, which SQL cannot carry. Everything before it — which events mutate the text, and in what
// order — is the record's own answer (record.ReportProjection over the report_op view), not a walk
// of the whole event log rebuilt here. Two structural failures stay loud, never a plausible zero:
// no base (nothing ingested) and two bases (the write-once rule broken) — both raised by the query.
func RenderFromRecord(run record.Run) (string, error) {
	base, haveBase, ops, err := record.ReportProjection(run)
	if err != nil {
		return "", err
	}
	if !haveBase {
		return "", fmt.Errorf("render: no base has been ingested for this run — there is nothing to render")
	}
	text := base
	for i, op := range ops {
		mut, err := mutationOf(op)
		if err != nil {
			return "", err
		}
		next, err := mut.apply(text)
		if err != nil {
			return "", fmt.Errorf("render: replaying mutation %d of %d (%s): %w", i+1, len(ops), mut.describe(), err)
		}
		text = next
	}
	return text, nil
}

// mutationOf turns one report_op row into the transform that applies it. "edit" is a splice
// (old→new); "insert" places Token(id) at the anchoring quote — the view already excluded a marker
// event with no location, so an insert always names a real quote here.
func mutationOf(op record.ReportOp) (mutation, error) {
	switch op.Kind {
	case "edit":
		return spliceMut{old: op.A, new: op.B}, nil
	case "insert":
		return insertMut{location: op.A, marker: anchor.Token(op.B)}, nil
	default:
		return nil, fmt.Errorf("render: unknown report_op kind %q — the report_op view emits only edit and insert", op.Kind)
	}
}

// BaseIngested reports whether this run already has a frozen base — the write-once state test that
// makes `blue ingest` refuse a second ingest. An EXISTS query, not a walk; read from the record, so
// it cannot drift from the truth.
func BaseIngested(run record.Run) (bool, error) {
	return record.ReportBaseExists(run)
}
