// ingest: the one-time act that turns the report from a file into a record (issue #709).
//
// Blue authors the round-0 report as a markdown file — the natural medium. `blue ingest` reads
// it VERBATIM into a single BaseIngest event, PROVES the record renders back to exactly those
// bytes, and then DELETES the file. After it there is no file: the report exists only as the base
// plus its append-only diff-stack, reads go through the render, and every change is an event.
//
// THREE GUARANTEES, each keyed on the record rather than a marker or a permission bit:
//   - WRITE-ONCE: a second ingest is refused and redirected to `blue edit`. The base already
//     exists on the record, so it cannot be overwritten — asked-and-answered by the events.
//   - AUTHOR-ONLY: only the seat that authored the report (blue-synthesize) may ingest it. A
//     response or red seat has no business freezing the base. (True surface-invisibility to other
//     seats is a follow-up needing the synthesize/respond surface split; this is the runtime gate.)
//   - VERIFY-BEFORE-DELETE: the file is removed ONLY after VerifyReproduction confirms the record
//     reproduces it byte-for-byte. If it does not, the file is KEPT and the error tells blue to
//     STOP — it is a tooling failure, not something an edit can fix.
package blue

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/reportproj"
)

// authorSeat is the ONE seat allowed to ingest — the round-0 report's author.
const authorSeat = "blue-synthesize"

func newIngest() *cobra.Command {
	return seat.Prose(seat.New("ingest", func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
		run, err := s.Run()
		if err != nil {
			return nil, err
		}

		// AUTHOR-ONLY.
		if s.SeatID != authorSeat {
			return nil, fmt.Errorf("blue ingest is for the report's author (%s) only — a %s seat does not freeze the base. Change the report through `blue edit`", authorSeat, s.SeatID)
		}

		// WRITE-ONCE — refuse-and-redirect if a base already exists (record-state, not a marker).
		already, err := reportproj.BaseIngested(run)
		if err != nil {
			return nil, err
		}
		if already {
			return nil, fmt.Errorf("blue ingest: this run's report is already ingested — the base is frozen and cannot be re-ingested or overwritten. Make every change through `blue edit`, which appends to the diff-stack")
		}

		reportPath := filepath.Join(run.Dir(), "blue", "report.md")
		content, err := os.ReadFile(reportPath)
		if err != nil {
			return nil, fmt.Errorf("blue ingest: reading the report to ingest: %w", err)
		}
		report := string(content)

		// Record the base, then PROVE the record reproduces it before touching the file.
		if _, err := record.Append(s.Identity(), &recordpb.BaseIngest{Text: proto.String(report)}); err != nil {
			return nil, err
		}
		rendered, err := reportproj.RenderFromRecord(run)
		if err != nil {
			return nil, fmt.Errorf("blue ingest: rendering the freshly-ingested base: %w — the file is UNTOUCHED", err)
		}
		if err := VerifyReproduction(report, rendered); err != nil {
			// The base is on the record but the render did not reproduce it. The file is KEPT; the
			// error already tells blue to STOP and escalate, not to edit.
			return nil, err
		}

		// Proven. Remove the file — from here the record is the only truth.
		if err := os.Remove(reportPath); err != nil {
			return nil, fmt.Errorf("blue ingest: the base is recorded and verified, but removing the file failed: %w — remove blue/report.md by hand; the record is authoritative", err)
		}

		return seat.Msg{Message: fmt.Sprintf("blue ingest: report frozen into the record (%d bytes), verified byte-for-byte, and the file removed. The report is now the base plus its diff-stack; read it with `show report`, change it with `blue edit`.", len(report))}, nil
	}))
}
