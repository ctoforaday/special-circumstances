package lens

import (
	"errors"
	"fmt"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"strings"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchortext"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/enumhelp"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// finding: a lens's graded observation, for the chair to dispose.
//
// The id is TOOL-minted. A lens does not invent it: hand-numbered names collided
// (four L5-F1s in one round of run 3), and the id is the identity a gap's
// found_by names, so it must be unambiguous run-wide. The lens passes a stable
// local --key purely as a crash-retry handle; a retry returns the existing
// finding rather than minting a duplicate (the mint --key pattern).
func newFinding() *cobra.Command {
	var severity, likelihood, impact flags.GradeValue

	c := seat.Prose(seat.New("finding", func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
		// One resolution, at the door. Reading the path off the context is what let a refused
		// run reach a verb as if it had resolved.
		run, err := s.Run()
		if err != nil {
			return nil, err
		}
		text, err := seat.Reason(cmd)
		if err != nil {
			return nil, err
		}
		// --reason and --location are REQUIRED (slice 1b): the location's quoted
		// sentence is the marker anchor + the snapshot red re-audits against; the
		// reason is the explanation. A finding without either cannot be anchored.
		if strings.TrimSpace(text) == "" {
			return nil, feov.Errorf(feov.MissingField, "requires --reason: the explanation red re-audits the repair against")
		}
		location := seat.Location(cmd)
		aboutKind, aboutRef := seat.Str(cmd, flags.AboutKind), seat.Str(cmd, flags.About)
		about, aboutRefP, aerr := record.ResolveAbout(run, record.MustEnum("finding", "about_kind"), aboutKind, aboutRef)
		aboutSet := about != nil
		if aerr != nil {
			return nil, aerr
		}
		// EXACTLY ONE ANCHOR. A finding anchored to nothing cannot be found; one anchored to both
		// a sentence and a section is claiming two subjects and the gap it becomes inherits the
		// ambiguity.
		switch {
		case strings.TrimSpace(location) == "" && !aboutSet:
			return nil, feov.Errorf(feov.MissingField, "needs an anchor: --quote for text that IS in the report, "+
				"or --about-kind/--about for something that is not.\n\n"+
				"An ABSENCE has no sentence to quote. Borrowing an innocent one as a handle — a missing inquiry "+
				"pinned to a sentence the finding itself calls fine — lands a reader of the gap list "+
				"on good prose. Anchor it to the section it "+
				"is missing from, the inquiry whose reason you are arguing against, or the gap it is about")
		case strings.TrimSpace(location) != "" && aboutSet:
			return nil, feov.Errorf(feov.Validation, "takes --quote OR --about, not both: a finding has one subject, "+
				"and the gap it becomes would inherit the ambiguity")
		}
		key := seat.Str(cmd, flags.Key)
		area := record.AreaOf(s.SeatID)
		body := &recordpb.Finding{
			FindingKey: proto.String(key),
			Location:   proto.String(location),
			Text:       proto.String(text),
			AboutKind:  about,
			AboutRef:   aboutRefP,
			Severity:   seat.GradeOrNil(&severity),
			Likelihood: seat.GradeOrNil(&likelihood),
			Impact:     seat.GradeOrNil(&impact),
		}
		// Crash-retry idempotency: a prior finding under this --key returns its id, no second event
		// AND no second marker (BEFORE any write) — when the retry states the finding the key names.
		if prior, err := record.FindingByKey(run, s.SeatID, key); err != nil {
			return nil, err
		} else if prior != nil {
			if err := seat.RepeatsKeyed(key, "finding", prior.GetId(), prior, body, findingAsserted...); err != nil {
				return nil, err
			}
			// THE PAIR MAY BE HALF ON THE RECORD: the finding and its anchor are one write, but a
			// record migrate brought forward from an archive can hold a finding whose run crashed
			// before its anchor. The retry finishes the pair at the location the finding stored.
			if err := seat.PlaceOwed(s, prior.GetId(), key); err != nil {
				return nil, err
			}
			return findingResult{ID: prior.GetId(), Area: area, Idempotent: true}, nil
		}
		// Mint the id UP FRONT: it forms the marker. THE ANCHOR EVENT (below) IS THE MARKER — it
		// carries the quote, and reportproj.Render re-places the marker at replay. No file is
		// spliced, so there is no torn-splice window; the --key retry above is idempotent and
		// finishes a pair the record holds half of.
		findingID := record.NewID("finding")
		body.Id = proto.String(findingID)
		// NO ANCHOR FOR AN ABSENCE, and that is the point rather than an omission: a finding about
		// something NOT in the report has no location to mark, and placing one would put an anchor on
		// the innocent prose this change exists to stop borrowing.
		// THE PLACEMENT IS VALIDATED against the report the Finding and its Anchor land on: NOT FOUND
		// -> reject (a mis-quote), in-fence -> reject. Nothing is recorded on a refusal. The Finding
		// and Anchor events are what the report is replayed from.
		//
		// SKIPPED ENTIRELY FOR AN ABSENCE: an --about finding has no quote, so it places nothing and
		// nothing is checked. Checking anyway is what forced a lens to borrow a live sentence as a
		// handle: the only way past the check was to name text that exists, whatever the finding was
		// actually about.
		//
		// A FINDING ABOUT A GAP IS REFUSED WHILE THE GAP HAS MOVED SINCE THIS SITTING OPENED (gblock
		// ruling B, 2026-10-19): lenses sit concurrently, and a dispute with a gap's grade, fix or
		// standing written against the board the sitting opened with is a dispute with a gap that may
		// already be closed, answered or regraded. The refusal shows what changed.
		write := func(guards ...record.Guard) error {
			return seat.AppendPlaced(s, body, findingID, location, findingRefusal, guards...)
		}
		if about != nil && *about == recordpb.AboutKind_ABOUT_KIND_GAP {
			err = seat.AppendFresh(s, *aboutRefP, write)
		} else {
			err = write()
		}
		if err != nil {
			return nil, err
		}
		return findingResult{ID: findingID, Area: area}, nil
	}))

	c.Flags().String(flags.Key, "", flags.DescKey+"; the TOOL mints the finding's id. A repeat under the handle with another --reason, grade or --about-kind/--about is refused: a handle names one finding")
	seat.Supplies(c, "id", "the tool mints it (record.NewID), at random — an id a seat chose would name a finding the record does not hold")
	c.Flags().Var(&severity, flags.Severity, flags.GradeUsage(record.AxisMeans(recordpb.GradeDimension_GRADE_DIMENSION_SEVERITY)))
	c.Flags().Var(&likelihood, flags.Likelihood, record.AxisMeans(recordpb.GradeDimension_GRADE_DIMENSION_LIKELIHOOD))
	c.Flags().Var(&impact, flags.Impact, record.AxisMeans(recordpb.GradeDimension_GRADE_DIMENSION_IMPACT))
	// A finding anchors by --quote OR by --about-kind/--about, and the handler refuses one with
	// neither — so the marker states the condition, not the bare word.
	flags.Text(c, flags.Quote, flags.DescQuote+". The finding anchor is placed there")
	seat.RequiredUnless(c, flags.Quote, "unless --about-kind/--about name the subject")
	enumhelp.Flag(c, flags.AboutKind, record.MustEnum("finding", "about_kind"),
		"anchor this finding to something that is NOT report text — a section for what is missing from it, an inquiry, a gap already on the board, or a contradiction red recorded; use instead of --quote. A finding about a gap reaches the seat that minted it, the one seat that can act on it. A finding about an inquiry reaches no seat's work list; a gap anchors to an inquiry the same way, and a material one is on blue's. A finding about a contradiction raises it whether or not the report still holds the sentence, and the gap minted from the finding answers it")
	flags.Text(c, flags.About, "the reference --about-kind names: a section heading, an inquiry id, a gap id, or a contradiction's id, which show evidence lists under unanswered_contradictions. It is CHECKED against the record")
	// The handler refuses a finding with no explanation; the marker says so where the seat reads.
	return seat.SaysRequired(c, flags.Reason)
}

// findingAsserted is what a finding states, and so what a retry under its key must state again:
// its text, its grades and its subject. The quote is not among them — a retry is finished at the
// location the finding stored.
var findingAsserted = []seat.Asserted{
	{Flag: flags.Reason, Field: "text"},
	{Flag: flags.Severity, Field: "severity"},
	{Flag: flags.Likelihood, Field: "likelihood"},
	{Flag: flags.Impact, Field: "impact"},
	{Flag: flags.AboutKind, Field: "about_kind"},
	{Flag: flags.About, Field: "about_ref"},
}

func findingRefusal(err error) error { return placementRefusal("finding", err) }

// placementRefusal is the refusal `finding` and `mint` give for a quote Attach will not place; noun
// is what the quote anchors. Both verbs name a section by --about-kind, so a quote on a heading is
// told to.
func placementRefusal(noun string, err error) error {
	switch {
	case errors.Is(err, anchortext.ErrMisQuote):
		return feov.Errorf(feov.Validation, "--quote was not found in report.md.\n\nIt is matched LITERALLY against the report, so it must be the quoted text ALONE. A section heading in front of it (\"Findings: …\", \"## Method — …\") is the common cause and makes it match nothing — measured, four times in one sitting with four different separators. Name the section in --reason instead.\n\nA quote may not cross a blank line: a %s anchors ONE passage", noun)
	case errors.Is(err, anchortext.ErrInFence):
		return feov.Errorf(feov.Validation, "the quote resolves inside a code fence — anchor a prose sentence, not code")
	case errors.Is(err, anchortext.ErrOnHeading):
		return fmt.Errorf("%w; for a %s about the whole section, name the section with --about-kind section", anchortext.Refusal(err), noun)
	}
	return anchortext.Refusal(err)
}

type findingResult struct {
	ID         string `json:"finding_id"`
	Area       string `json:"area"`
	Idempotent bool   `json:"idempotent,omitempty"`
}

func (r findingResult) Human() string {
	ref := record.FindingRef(r.ID, r.Area)
	if r.Idempotent {
		return "finding " + ref + " (idempotent retry — existing finding returned)"
	}
	return "finding recorded: " + ref + " — the id a gap's found_by names"
}
