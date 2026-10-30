// ingest: a lane's act that turns its part of the report from a file into a record.
//
// A lane writes its part as a markdown file — the natural medium — at the path the tool builds from
// its seat (partPath). `ingest` reads it VERBATIM into a ReportPart event, PROVES the record renders
// it where the join rule puts it, and then DELETES the file. The report is the lanes' parts, in the
// order they were recorded, with every edit, anchor and retire after them folded on; reads go through
// the render, and every change is an event.
//
// WHO, WHEN AND WHAT, each keyed on the record rather than a marker or a permission bit:
//   - LANES ONLY: a part is a lane's, and the cast names the lanes (record.LaneSeatsOf). The rule is
//     this verb's, not the write path's, so an archived report the synthesizer ingested migrates as
//     it stands.
//   - ONCE PER LANE, AND BEFORE THE REPORT CHANGES: a lane records one part, and no part is recorded
//     once the report holds an op other than a part — a part then would land after text the synthesis
//     has reorganized. Both are asked again inside the write (Landing.PartStanding), so two writes
//     cannot both pass them.
//   - THE REPORT'S OWN SECTIONS ONLY: the title, TL;DR, Catechism and Open questions are the
//     synthesizer's (report.PartSurfaceIn), so parts cannot contend for them.
//   - VERIFY-BEFORE-DELETE: the file is removed ONLY after VerifyReproduction confirms the report, as
//     it stood once this part was recorded, ends with the part as the join rule places it. If it does
//     not, the file is KEPT and the error tells the lane to STOP — a tooling failure, not something an
//     edit can fix.
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
)

// partPath is where seat's part waits to be recorded: built from the seat, never named by it. The
// path is a location the lane writes to; nothing reads a fact back out of the name.
func partPath(run record.Run, seat string) string {
	return filepath.Join(run.Dir(), "blue", "candidates", seat+".md")
}

// NO PROSE CHANNEL. ReportPart carries the part and nothing else, so a --reason here had nowhere
// to go. The part IS the artifact; a verb with no field for an argument does not accept one.
func newIngest() *cobra.Command {
	return seat.New("ingest", ingestResult{}, func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
		run, err := s.Run()
		if err != nil {
			return nil, err
		}

		lanes, err := record.LaneSeatsOf(run)
		if err != nil {
			return nil, err
		}
		if err := laneRefusal(s.SeatID, lanes); err != nil {
			return nil, err
		}
		recorded, changed, err := record.PartStanding(run, s.SeatID)
		if err != nil {
			return nil, err
		}
		if err := partStandingRefusal(s.SeatID, recorded, changed); err != nil {
			return nil, err
		}

		path := partPath(run, s.SeatID)
		if err := record.RefuseRecordFile(run, path, filepath.Join("blue", "candidates", s.SeatID+".md"), "a part of the report"); err != nil {
			return nil, err
		}
		content, err := os.ReadFile(path)
		if os.IsNotExist(err) {
			return nil, feov.Errorf(feov.NotFound, "there is no part at %s — ingest records that one file, the one named for your seat, so write your part there first", path)
		}
		if err != nil {
			return nil, fmt.Errorf("reading the part to ingest at %s: %w", path, err)
		}
		part := string(content)
		if err := partSurfaceRefusal(part); err != nil {
			return nil, err
		}
		if err := composedRefusal("", part); err != nil {
			return nil, err
		}

		// Record the part, the lane's standing asked again under the write lock, then PROVE the record
		// reproduces it before touching the file.
		if _, err := record.Append(s.Identity(), &recordpb.ReportPart{Text: proto.String(part)},
			record.Check(func(l *record.Landing) error {
				recorded, changed, err := l.PartStanding(s.SeatID)
				if err != nil {
					return err
				}
				return partStandingRefusal(s.SeatID, recorded, changed)
			})); err != nil {
			return nil, err
		}
		rendered, err := record.ReportAtPart(run, s.SeatID)
		if err != nil {
			return nil, fmt.Errorf("rendering the report with the part just recorded: %w — the file is UNTOUCHED", err)
		}
		if err := VerifyPart(part, rendered); err != nil {
			// The part is on the record but the render did not reproduce it. The file is KEPT; the
			// error already tells the lane to STOP and escalate, not to edit.
			return nil, err
		}

		// Proven. Remove the file — from here the record is the only truth.
		if err := os.Remove(path); err != nil {
			return nil, fmt.Errorf("the part is recorded and verified, but removing the file failed: %w — remove %s by hand; the record is authoritative", err, path)
		}
		return ingestResult{Path: path, Bytes: len(part)}, nil
	})
}

// laneRefusal refuses a part from a seat the cast does not name a lane.
func laneRefusal(seatID string, lanes []string) error {
	for _, l := range lanes {
		if l == seatID {
			return nil
		}
	}
	return feov.Errorf(feov.RoleViolation, "%s records no part of the report: a part is a lane's, and this run's lanes are %s. NOTHING IS RECORDED. The report changes through edit, cite, compute and retire",
		seatID, strings.Join(lanes, ", "))
}

// partStandingRefusal refuses a lane's second part, and any part once the report holds an op other
// than a part.
func partStandingRefusal(seatID string, recorded, changed bool) error {
	switch {
	case recorded:
		return feov.Errorf(feov.Conflict, "%s has already recorded its part of the report, and a lane records one. NOTHING IS RECORDED and the file stands. Your part is in the report as recorded; the synthesizer works on it from here, so end your sitting", seatID)
	case changed:
		return feov.Errorf(feov.Conflict, "the report has already been changed, so a part recorded now would land after text the synthesis has reorganized. NOTHING IS RECORDED and the file stands. Your part is not in the report: say so in your log, with what it holds that the report lacks")
	}
	return nil
}

// partSurfaceRefusal refuses a part that holds a row of the text other than the report's own
// sections (report.PartSurfaceIn).
func partSurfaceRefusal(part string) error {
	line, row, found := report.PartSurfaceIn(report.ParseText(part))
	if !found {
		return nil
	}
	return feov.Errorf(feov.Validation, "your part would carry %q, and %s is the synthesizer's to write, not a lane's: a part holds only the report's own sections, under headings of its own. NOTHING IS RECORDED and the file stands. Take that line and what stands under it out of your part; what it says belongs in a section of your own or to the synthesizer",
		line, row)
}

func init() {
	seat.HelpValues["PartSurfaces"] = report.PartSurfaces
	seat.HelpValues["RefusedHeadings"] = strings.Join(report.RefusedHeadings(), ", ")
	seat.HelpValues["SubjectHeadingsNote"] = report.SubjectHeadingsNote
	seat.HelpValues["WovenNotesRule"] = report.WovenNotesRule
	seat.HelpValues["TextOrderRule"] = report.TextOrderRule
}

// composedRefusal refuses report text that claims what the tool writes: a section it composes
// (report.ComposedIn, which reads the assembler's own table of sections), or a footnote label of
// the shape the weave numbers its notes with (report.WovenNoteIn); and text that holds a section
// out of the order the report ships in (orderRefusal). It is asked of the document a
// write would leave — the part ingest is about to record, the report an edit has planned — against
// the report as it stood (was; empty at ingest), so a heading an edit completes from parts is
// caught as one it writes whole, and a claim the record already holds is not laid on the edit that
// leaves it or takes it out. It refuses before anything is recorded: the file stands and the edit
// is not on the stack.
//
// THE WRITE IS WHERE THE AUTHOR READS IT, AND THE REFUSAL SAYS WHERE THE CONTENT GOES. Assembly
// composes these sections itself; text filed under one has no place of its own in the set, and
// its author learns that from no later act. The row's Path is the way in: the act that puts
// content of that kind on the record, or that no act of blue's does and whose does.
//
// Each text is parsed once (report.ParseText) and every refusal reads that one parse.
func composedRefusal(wasText, docText string) error {
	was, doc := report.ParseText(wasText), report.ParseText(docText)
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
func wovenNoteRefusal(was, doc report.ParsedText) error {
	label, s, found := report.WovenNoteIn(was, doc)
	if !found {
		return orderRefusal(was, doc)
	}
	return feov.Errorf(feov.Validation, "the report's text would carry the footnote label %q, and a label of that shape is assembly's to write: it numbers the report's notes %s, in %s, from %s. "+
		"A note typed by hand under such a label ships beside the tool's note of the same label, and a reader is shown one of the two. "+
		"NOTHING IS RECORDED by this call. HOW CONTENT REACHES IT: %s "+
		"WHAT TO DO: take every reference of that shape this write adds, and the `[^…]:` line that defines it, out of the text. A source or a computation goes by the way above. %s",
		label, report.WovenNoteLabels, s.File, s.From, s.Path, strings.TrimSuffix(report.NoteInCode, "."))
}

// orderRefusal refuses report text that holds a line out of the order the report ships in
// (report.OutOfOrderIn, which reads the order off the assembler's table). The text ships as it is
// written, so a section out of place in it is out of place in the reader's report, and the write is
// the one act that can say so: every later reader takes the order as its author's. Like the
// refusals above it is for what the write adds (report.added): a record migrated with its text as
// written can hold a section out of place, and an edit that leaves it there, or moves it to its
// place, is not the one that put it there.
func orderRefusal(was, doc report.ParsedText) error {
	m, found := report.OutOfOrderIn(was, doc)
	if !found {
		return nil
	}
	side := "before"
	if m.After {
		side = "after"
	}
	beside := []rune(m.Beside)
	if len(beside) > 120 {
		beside = append(beside[:120], '…')
	}
	return feov.Errorf(feov.Validation, "the report's text would carry %q %s %q, and %s. "+
		"The report ships in the order its text is written, and its text stands in this order: %s. "+
		"NOTHING IS RECORDED by this call. "+
		"WHAT TO DO: move the section to its place, whole, with every line under it and every anchor it holds. "+
		"In an edit, quote the section and what stands between it and its place as one span, as `show report` prints them, and write the same text in --new with the section where it belongs: "+
		"every anchor the span holds stays on its sentence. Where what stands on the wrong side is text this write adds, write that text on the right side instead. "+
		"Where the section is a second one under its name, give it a heading of its own or join it to the first",
		m.Line, side, string(beside), m.Belongs, report.TextOrder)
}

// ingestResult confirms the part recorded.
type ingestResult struct {
	// Path is the file that was recorded and removed.
	Path  string `json:"path"`
	Bytes int    `json:"bytes"`
}

func (r ingestResult) Human() string {
	return fmt.Sprintf("ingest: %s recorded as your part of the report (%d bytes), verified, and removed. Read the report with `show report`.", r.Path, r.Bytes)
}
