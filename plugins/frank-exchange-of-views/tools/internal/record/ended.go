package record

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/runlive"
)

// AN ENDED RUN TAKES NOTHING (gblock, 2026-10-24: "Ended run takes nothing"). A run ENDS when it is
// captured: capture archives the record to the one place that outlives the container, folds its
// audits into the documents, and closes the run-live row. An event written after that is in no
// archive and no document, and no seat or operator reads it — so every write is refused, loudly,
// the tool's own refusal entry included (cli.noteRefusal tells the caller its refusal is not
// recorded). Between the outcome and the end the record takes the log alone (outcomeclosed.go),
// which is where a seat that did not know the run was over shows up as friction.
//
// A FILE BESIDE THE RECORD, NOT A FIELD IN IT. Whether a run was captured is about its
// disposition, not its contents (Assess says the same), and a field for it would change the
// record's shape — the event-schema epoch's business — and the hooks, which link no record, read
// it too (runlive.StateOf). Capture writes the mark at its close (MarkEnded), after its last write to the record; the writes it makes itself are the per-turn
// measurements (AppendSeatTurns), a table of their own that no seat writes and a capture run again
// refreshes, so the mark refuses none of capture's own work.
//
// WHAT A MISS RETURNS: a run captured with no mark beside it — the write failed, which capture
// reports — reads as not ended, and takes writes as a run between its outcome and its end does.
// That is the direction of the rule before it existed, never a refusal of a live run.

// ErrRunEnded is the cause every refusal of a write to an ended run carries.
var ErrRunEnded = errors.New("the run has ended")

// RunEnded reports whether the run was captured (runlive.EndedMark) — the one reading of ENDED
// every reader in this tree takes: the write path refuses on it, and Assess reports StateEnded on
// it. An outcome on the record is not this fact: the bench has decided the run, and the assembly
// sitting and the capture may both still be ahead.
func RunEnded(run Run) (bool, error) {
	if !run.Valid() {
		return false, nil
	}
	ended, err := runlive.Ended(run.Dir())
	if err != nil {
		return false, fmt.Errorf("asking whether the run was captured: %w", err)
	}
	return ended, nil
}

// requireRunNotEnded refuses a write to a run that was captured. A mark that cannot be asked
// refuses the write too: a write admitted because the question could not be answered is one the
// record vouched for on nothing.
func requireRunNotEnded(run Run) error {
	ended, err := RunEnded(run)
	if err != nil {
		return err
	}
	if ended {
		return feov.Wrap(feov.Conflict, ErrRunEnded,
			"the run has ended — it was captured — and its record takes nothing: this write is refused and records nothing, a log entry included. "+
				"What the record holds was archived and assembled at the capture; an act now reaches no reader")
	}
	return nil
}

// MarkEnded writes the mark that the run was captured. Capture calls it at its close.
func MarkEnded(run Run, at string) error {
	return os.WriteFile(filepath.Join(run.Dir(), runlive.EndedMark), []byte(at+"\n"), 0o644)
}
