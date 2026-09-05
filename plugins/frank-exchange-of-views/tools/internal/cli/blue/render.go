package blue

import (
	"fmt"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/bluedoc"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// THE REPORT IS THE RECORD, NOT A FILE (issue #709).
//
// The current report is base + the ordered diff-stack replayed onto it. `blue edit` records each
// splice as a BlueEdit event carrying (old, new); Render folds those splices back over the frozen
// base and reproduces exactly the bytes the file used to hold — because the splice it replays is
// the SAME transform blue edit applied (applySplice), located by the SAME rule (bluedoc). Once the
// base is ingested and the file deleted, this is the only way the report exists: there is no file
// to raw-write, chattr, or bypass, and a change that is not a recorded event cannot exist.

// applySplice is the pure text transform blue edit and Render share: replace report[start:end]
// with new, then tidy the two seams the splice created. NO legality guards — those (anchors
// transit, no dropped marker, present-and-unique) are write-time validation that already passed
// when the edit was recorded; replay reproduces bytes, it does not re-authorise.
func applySplice(report string, start, end int, new string) string {
	next := report[:start] + new + report[end:]
	// SPLICE HYGIENE (see splice.go): tidy the two seams this edit created. Deterministic, narrow,
	// silent — and it MUST run in Render too, or the replayed text drifts from what the file held.
	next, _ = tidySeam(next, start+len(new)) // trailing seam first — the later offset is stable
	next, _ = tidySeam(next, start)
	return next
}

// Op is one recorded splice from the diff-stack: the (old, new) of a BlueEdit event, in record
// order. Render takes plain strings rather than *recordpb.BlueEdit so this package stays free of
// the record schema — the caller reads the events and hands over the pairs.
type Op struct{ Old, New string }

// Render reproduces the current report from the frozen base and the ordered diff-stack. Each op is
// located in the text as it stands after the previous ops (the same running-report blue edit saw),
// spliced, and its seams tidied. A replay that cannot locate an op's old span is an integrity
// failure of the record, not a user error: it means the stack no longer describes a real sequence
// of edits over this base, so it is returned loudly rather than skipped.
func Render(base string, ops []Op) (string, error) {
	text := base
	for i, op := range ops {
		start, end, err := bluedoc.LocateUniqueReplacing("render", text, op.Old)
		if err != nil {
			return "", fmt.Errorf("render: replaying edit %d of %d: %w", i+1, len(ops), err)
		}
		text = applySplice(text, start, end, op.New)
	}
	return text, nil
}

// RenderFromRecord is the record-aware render: it reads the run's frozen base (the one BaseIngest
// event) and the ordered BlueEdit diff-stack, and folds the edits back over the base. This is the
// ONLY way to obtain the current report once the file is deleted — the ~12 former readers of
// blue/report.md call this instead.
//
// Events arrive in insertion order (MergedEvents is ORDER BY id), which is edit order, and the
// BaseIngest is written first (round 0, before any edit), so a linear pass composes the report
// correctly. Two structural failures are loud, never a plausible zero: no base (nothing was
// ingested) and two bases (the write-once rule was broken).
func RenderFromRecord(run record.Run) (string, error) {
	m, err := record.MergedEvents(run)
	if err != nil {
		return "", err
	}
	base := ""
	haveBase := false
	var ops []Op
	for _, e := range m.Events {
		body, ok := recordpb.Body(e)
		if !ok {
			continue
		}
		switch b := body.(type) {
		case *recordpb.BaseIngest:
			if haveBase {
				return "", fmt.Errorf("render: two BaseIngest events on this run — the base is written once and never rewritten")
			}
			base, haveBase = b.GetText(), true
		case *recordpb.BlueEdit:
			ops = append(ops, Op{Old: b.GetOld(), New: b.GetNew()})
		}
	}
	if !haveBase {
		return "", fmt.Errorf("render: no base has been ingested for this run — there is nothing to render")
	}
	return Render(base, ops)
}

// BaseIngested reports whether this run already has a frozen base — the record-state test that
// makes `blue ingest` write-once: a second ingest is refused, never an overwrite. The answer is
// read from the events, not a marker file, so it cannot drift from the truth.
func BaseIngested(run record.Run) (bool, error) {
	m, err := record.MergedEvents(run)
	if err != nil {
		return false, err
	}
	for _, e := range m.Events {
		if body, ok := recordpb.Body(e); ok {
			if _, isBase := body.(*recordpb.BaseIngest); isBase {
				return true, nil
			}
		}
	}
	return false, nil
}
