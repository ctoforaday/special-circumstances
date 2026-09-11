// ingest: the one-time act that turns the report from a file into a record (issue #709).
//
// Blue authors the round-0 report as a markdown file — the natural medium. `blue ingest` reads
// it VERBATIM into a single BaseIngest event, PROVES the record renders back to exactly those
// bytes, and then DELETES the file. After it there is no file: the report exists only as the base
// plus its append-only diff-stack, reads go through the render, and every change is an event.
//
// THREE GUARANTEES, each keyed on the record rather than a marker or a permission bit:
//   - WRITE-ONCE: a second ingest is refused and redirected to `edit`. The base already
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
	"strings"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/reportproj"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/reportvoice"
)

// authorSeat is the ONE seat allowed to ingest — the round-0 report's author.
const authorSeat = "blue-synthesize"

// NO PROSE CHANNEL. BaseIngest carries the report and nothing else, so a --reason here had nowhere
// to go: it was registered (seat.Prose), accepted, and recorded nowhere. The report IS the
// artifact; a verb with no field for an argument does not accept one.
func newIngest() *cobra.Command {
	return seat.New("ingest", func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
		run, err := s.Run()
		if err != nil {
			return nil, err
		}

		// AUTHOR-ONLY.
		if s.SeatID != authorSeat {
			return nil, fmt.Errorf("blue ingest is for the report's author (%s) only — a %s seat does not freeze the base. Change the report through `edit`", authorSeat, s.SeatID)
		}

		// WRITE-ONCE — refuse-and-redirect if a base already exists (record-state, not a marker).
		already, err := reportproj.BaseIngested(run)
		if err != nil {
			return nil, err
		}
		if already {
			return nil, fmt.Errorf("blue ingest: this run's report is already ingested — the base is frozen and cannot be re-ingested or overwritten. Make every change through `edit`, which appends to the diff-stack")
		}

		reportPath := filepath.Join(run.Dir(), "blue", "report.md")
		content, err := os.ReadFile(reportPath)
		if err != nil {
			return nil, fmt.Errorf("blue ingest: reading the report to ingest: %w", err)
		}
		report := string(content)

		// THE VOICE CENSUS, over the WHOLE BASE. This is the largest authored artifact in a run —
		// 16,874 characters on the measured one — and until #873 it was the only report text that
		// reached the record without ever meeting the advisory. Every inline lane tag in that
		// run's finished report came in through here.
		//
		// FindAll, not Find: `edit` advises on one span, where knowing a tell is PRESENT is
		// the whole signal. A document needs the count and the places, or an author is told once
		// that a lane tag exists and cannot reach the other five.
		//
		// It refuses nothing, and must not start. A pattern cannot tell a report narrating its own
		// construction from a report quoting a source that narrates something, and a gate that
		// cannot tell those apart would silently cost the second one. Red's voice lens holds the
		// judgement; this says where to look.
		census := voiceCensus(report)

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

		return ingestResult{
			Bytes:      len(report),
			VoiceTells: census,
		}, nil
	})
}

// voiceCensus renders the whole-document census as lines an author can act on.
//
// Grouped by class with the lines named, because that is how the work is done: an author fixing
// six lane tags wants one item naming six places, not six items. The cap exists so a badly-voiced
// base cannot bury the rest of the result, and it STATES that it capped — a truncated list
// presented as a whole one is the defect this package is about.
func voiceCensus(report string) []string {
	occ := reportvoice.FindAll(report)
	if len(occ) == 0 {
		return nil
	}
	byClass := map[reportvoice.Class][]reportvoice.Occurrence{}
	var order []reportvoice.Class
	for _, o := range occ {
		if _, seen := byClass[o.Class]; !seen {
			order = append(order, o.Class)
		}
		byClass[o.Class] = append(byClass[o.Class], o)
	}
	const showPerClass = 8
	var out []string
	for _, c := range order {
		os := byClass[c]
		lines := make([]string, 0, len(os))
		shown := os
		if len(shown) > showPerClass {
			shown = shown[:showPerClass]
		}
		for _, o := range shown {
			lines = append(lines, fmt.Sprintf("%d", o.Line))
		}
		tail := ""
		if len(os) > len(shown) {
			tail = fmt.Sprintf(" (+%d more)", len(os)-len(shown))
		}
		out = append(out, fmt.Sprintf("%s ×%d at line %s%s — %s: %s",
			c, len(os), strings.Join(lines, ", "), tail, quoteFirst(os), os[0].Redirect))
	}
	return out
}

func quoteFirst(os []reportvoice.Occurrence) string {
	return fmt.Sprintf("first is %q", os[0].Match)
}

// ingestResult carries the census back with the confirmation.
type ingestResult struct {
	Bytes int `json:"bytes"`
	// VoiceTells is ADVICE and the base is already frozen by the time it renders. Ingest is
	// WRITE-ONCE and has just deleted the file, so unlike `edit` the author cannot re-do the
	// act it advises on — the route from here is `edit`, and the message says so rather than
	// implying a re-ingest that the record would refuse.
	VoiceTells []string `json:"voice_tells,omitempty"`
}

func (r ingestResult) Human() string {
	head := fmt.Sprintf("blue ingest: report frozen into the record (%d bytes), verified byte-for-byte, and the file removed. The report is now the base plus its diff-stack; read it with `show report`, change it with `edit`.", r.Bytes)
	if len(r.VoiceTells) == 0 {
		return head
	}
	return head + "\n\nNOTE — the base sounds in places like the run rather than the subject. It is\nrecorded and this is not a refusal; it may be wrong. Change any of it with\n`edit` — the base itself is frozen and cannot be re-ingested:\n  - " +
		strings.Join(r.VoiceTells, "\n  - ") +
		"\n\nSEPARATION, NEVER DELETION: where a tell carries a real limit on the CONCLUSION,\nit stays and is re-voiced as a limit on the subject — only the fact about the run\ngoes. Red's voice lens holds that judgement; these are the literal tells, and the\nleaks that matter most are the ones no pattern catches."
}
