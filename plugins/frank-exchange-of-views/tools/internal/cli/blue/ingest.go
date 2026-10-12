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
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/report"
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
			return nil, feov.Errorf(feov.RoleViolation, "this act is the report's author's (%s) only — a %s seat does not freeze the base. Change the report through `edit`", authorSeat, s.SeatID)
		}

		// WRITE-ONCE — refuse-and-redirect if a base already exists (record-state, not a marker).
		already, err := reportproj.BaseIngested(run)
		if err != nil {
			return nil, err
		}
		if already {
			return nil, feov.Errorf(feov.Conflict, "this run's report is already ingested — the base is frozen and cannot be re-ingested or overwritten. Make every change through `edit`, which appends to the diff-stack")
		}

		reportPath := filepath.Join(run.Dir(), "blue", "report.md")
		content, err := os.ReadFile(reportPath)
		if os.IsNotExist(err) {
			// NOTHING WAITS AT THIS PATH, so a report written anywhere else is a refusal naming where
			// ingest looked. Setup used to stub it, and B7's synthesizer copied its draft onto the
			// run-root stub while this froze the other — 40 bytes of heading — as the report's base.
			stray := ""
			if _, serr := os.Stat(filepath.Join(run.Dir(), "report.md")); serr == nil {
				stray = fmt.Sprintf(" A report.md exists at the run root (%s): that is the assembler's path, not this one — move it here.", filepath.Join(run.Dir(), "report.md"))
			}
			return nil, feov.Errorf(feov.NotFound, "there is no report at %s — ingest freezes that one file, so write the synthesized report there first.%s", reportPath, stray)
		}
		if err != nil {
			return nil, fmt.Errorf("reading the report to ingest at %s: %w", reportPath, err)
		}
		report := string(content)
		if err := composedRefusal("", report); err != nil {
			return nil, err
		}

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
			return nil, fmt.Errorf("rendering the freshly-ingested base: %w — the file is UNTOUCHED", err)
		}
		if err := VerifyReproduction(report, rendered); err != nil {
			// The base is on the record but the render did not reproduce it. The file is KEPT; the
			// error already tells blue to STOP and escalate, not to edit.
			return nil, err
		}

		// Proven. Remove the file — from here the record is the only truth.
		if err := os.Remove(reportPath); err != nil {
			return nil, fmt.Errorf("the base is recorded and verified, but removing the file failed: %w — remove blue/report.md by hand; the record is authoritative", err)
		}

		return ingestResult{
			Path:       reportPath,
			Bytes:      len(report),
			VoiceTells: census,
		}, nil
	})
}

func init() {
	seat.HelpValues["RefusedHeadings"] = strings.Join(report.RefusedHeadings(), ", ")
	seat.HelpValues["SubjectHeadingsNote"] = report.SubjectHeadingsNote
	seat.HelpValues["WovenNotesRule"] = report.WovenNotesRule
}

// composedRefusal refuses report text that claims what the tool writes: a section it composes
// (report.ComposedIn, which reads the assembler's own table of sections), or a footnote label of
// the shape the weave numbers its notes with (report.WovenNoteIn). It is asked of the document a
// write would leave — the file ingest is about to freeze, the report an edit has planned — against
// the report as it stood (was; empty at ingest), so a heading an edit completes from parts is
// caught as one it writes whole, and a claim the record already holds is not laid on the edit that
// leaves it or takes it out. It refuses before anything is recorded: the file stands and the edit
// is not on the stack.
//
// THE WRITE IS WHERE THE AUTHOR READS IT, AND THE REFUSAL SAYS WHERE THE CONTENT GOES. Assembly
// composes these sections itself; text filed under one has no place of its own in the set, and
// its author learns that from no later act. The row's Path is the way in: the act that puts
// content of that kind on the record, or that no act of blue's does and whose does.
func composedRefusal(was, doc string) error {
	line, s, found := report.ComposedIn(was, doc)
	if !found {
		return wovenNoteRefusal(was, doc)
	}
	return feov.Errorf(feov.Validation, "the report's text would carry %q, and %s is not the report's text to write: assembly composes it, in %s, from %s. "+
		"NOTHING IS RECORDED by this call. HOW CONTENT REACHES IT: %s "+
		"WHAT TO DO: take that line and what stands under it out of the report's text. What it says goes by the way above and by no other; "+
		"only an argument of the report's own, which that way does not carry, stays in the text, under a heading of the report's own. "+
		"`show report --assembled` lists every section of the assembled set with who composes it and how content reaches it",
		line, s.Heading, s.File, s.From, s.Path)
}

// wovenNoteRefusal refuses a footnote typed by hand under a label the weave writes. The hand note
// and the tool's ship under one label and a reader's renderer shows one of them, so the other is
// lost with nothing on the record to say so; the write is the last place its author can be told.
// The way in is the Bibliography row's Path, as it is for the section itself.
func wovenNoteRefusal(was, doc string) error {
	label, s, found := report.WovenNoteIn(was, doc)
	if !found {
		return nil
	}
	return feov.Errorf(feov.Validation, "the report's text would carry the footnote label %q, and a label of that shape is assembly's to write: it numbers the report's notes %s, in %s, from %s. "+
		"A note typed by hand under such a label ships beside the tool's note of the same label, and a reader is shown one of the two. "+
		"NOTHING IS RECORDED by this call. HOW CONTENT REACHES IT: %s "+
		"WHAT TO DO: take every reference of that shape this write adds, and the `[^…]:` line that defines it, out of the text. A source or a computation goes by the way above. %s",
		label, report.WovenNoteLabels, s.File, s.From, s.Path, strings.TrimSuffix(report.NoteInCode, "."))
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
	// Path is the file that was frozen and removed, named so an author who wrote elsewhere sees
	// which file the base came from.
	Path  string `json:"path"`
	Bytes int    `json:"bytes"`
	// VoiceTells is ADVICE and the base is already frozen by the time it renders. Ingest is
	// WRITE-ONCE and has just deleted the file, so unlike `edit` the author cannot re-do the
	// act it advises on — the route from here is `edit`, and the message says so rather than
	// implying a re-ingest that the record would refuse.
	VoiceTells []string `json:"voice_tells,omitempty"`
}

func (r ingestResult) Human() string {
	head := fmt.Sprintf("blue ingest: %s frozen into the record (%d bytes), verified byte-for-byte, and removed. The report is now the base plus its diff-stack; read it with `show report`, change it with `edit`.", r.Path, r.Bytes)
	if len(r.VoiceTells) == 0 {
		return head
	}
	return head + "\n\nNOTE — the base sounds in places like the run rather than the subject. It is\nrecorded and this is not a refusal; it may be wrong. Change any of it with\n`edit` — the base itself is frozen and cannot be re-ingested:\n  - " +
		strings.Join(r.VoiceTells, "\n  - ") +
		"\n\nSEPARATION, NEVER DELETION: where a tell carries a real limit on the CONCLUSION,\nit stays and is re-voiced as a limit on the subject — only the fact about the run\ngoes. Red's voice lens holds that judgement; these are the literal tells, and the\nleaks that matter most are the ones no pattern catches."
}
