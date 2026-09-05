package blue

import (
	"fmt"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/bluedoc"
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
