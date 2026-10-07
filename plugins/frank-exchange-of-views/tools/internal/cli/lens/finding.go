package lens

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchortext"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/enumhelp"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/reportproj"
)

// finding: a lens's graded observation, for the chair to dispose.
//
// The label is TOOL-assigned — L{role}-F{N}, run-unique per role, the role read
// from the seat id. A lens no longer invents it: hand-numbered labels collided
// (four L5-F1s in one round of run 3), and the label is now the identity a gap's
// found_by names, so it must be unambiguous run-wide. The lens passes a stable
// local --key (its own F1/F2) purely as a crash-retry handle; a retry returns the
// existing label rather than minting a duplicate (the mint --key pattern).
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
			return nil, fmt.Errorf("lens finding requires --reason: the explanation red re-audits the repair against")
		}
		location := seat.Str(cmd, flags.Quote)
		aboutKind, aboutRef := seat.Str(cmd, flags.AboutKind), seat.Str(cmd, flags.About)
		about, aboutRefP, aerr := record.ResolveAbout("lens finding", run, aboutKind, aboutRef)
		aboutSet := about != nil
		if aerr != nil {
			return nil, aerr
		}
		// EXACTLY ONE ANCHOR. A finding anchored to nothing cannot be found; one anchored to both
		// a sentence and a section is claiming two subjects and the gap it becomes inherits the
		// ambiguity.
		switch {
		case strings.TrimSpace(location) == "" && !aboutSet:
			return nil, fmt.Errorf("lens finding needs an anchor: --quote for text that IS in the report, " +
				"or --about-kind/--about for something that is not.\n\n" +
				"An ABSENCE has no sentence to quote. Borrowing an innocent one as a handle — a missing avenue " +
				"pinned to a sentence the finding itself calls fine — lands a reader of the gap list " +
				"on good prose. Anchor it to the section it " +
				"is missing from, the avenue whose reason you are arguing against, or the gap it is about")
		case strings.TrimSpace(location) != "" && aboutSet:
			return nil, fmt.Errorf("lens finding takes --quote OR --about, not both: a finding has one subject, " +
				"and the gap it becomes would inherit the ambiguity")
		}
		// Crash-retry idempotency: a prior finding under this --key returns its
		// label, no second event AND no second marker (BEFORE any write).
		key := seat.Str(cmd, flags.Key)
		if prior, priorID, err := record.FindingByKey(run, s.SeatID, key); err != nil {
			return nil, err
		} else if prior != "" {
			// THE PAIR MAY BE HALF-APPENDED: the finding and its anchor are two appends, so a crash
			// between them leaves the finding recorded and its anchor out of the report. The retry
			// finishes the pair at the location the finding stored.
			if err := placeOwed(s, run, priorID, func(err error) error { return placementRefusal("lens finding", "finding", err) }); err != nil {
				return nil, err
			}
			return findingResult{Label: prior, Idempotent: true}, nil
		}
		label, err := record.NextFindingLabel(run, s.SeatID)
		if err != nil {
			return nil, err
		}
		// Mint the id UP FRONT: it forms the marker. THE ANCHOR EVENT (below) IS THE MARKER — it
		// carries the quote, and reportproj.Render re-places the marker at replay. No file is
		// spliced, so there is no torn-splice window; the --key retry above is idempotent and
		// reconciles a half-appended pair.
		findingID := record.NewFindingID()

		// VALIDATE the placement against the current render: NOT FOUND -> reject (a mis-quote),
		// in-fence -> reject. Nothing is recorded on a refusal. On success the bytes are discarded —
		// the Finding + Anchor events below are what the report is replayed from.
		//
		// SKIPPED ENTIRELY FOR AN ABSENCE. There is no placement to validate when the subject is
		// something the report does NOT contain, and running this anyway is what forced a lens to
		// borrow a live sentence as a handle: the only way past this check was to name text that
		// exists, whatever the finding was actually about.
		if !aboutSet {
			current, rerr := reportproj.RenderFromRecord(run)
			if rerr != nil {
				return nil, rerr
			}
			if _, aerr := anchortext.Attach(current, findingID, location); aerr != nil {
				return nil, placementRefusal("lens finding", "finding", aerr)
			}
		}

		body := &recordpb.Finding{
			Label:      proto.String(label),
			FindingId:  proto.String(findingID),
			FindingKey: proto.String(seat.Str(cmd, flags.Key)),
			Location:   proto.String(seat.Str(cmd, flags.Quote)),
			Text:       proto.String(text),
			AboutKind:  about,
			AboutRef:   aboutRefP,
			Severity:   seat.GradeOrNil(&severity),
			Likelihood: seat.GradeOrNil(&likelihood),
			Impact:     seat.GradeOrNil(&impact),
		}
		if _, err := record.Append(s.Identity(), body); err != nil {
			return nil, err
		}
		// NO ANCHOR FOR AN ABSENCE, and that is the point rather than an omission: a finding about
		// something NOT in the report has no location to mark, and placing one would put an anchor on
		// the innocent prose this change exists to stop borrowing.
		if !aboutSet {
			ap := &recordpb.Anchor{Id: proto.String(findingID), Location: proto.String(location)}
			if _, err := record.Append(s.Identity(), ap); err != nil {
				return nil, err
			}
		}
		// The LABEL leads: it is the run-unique identity a gap's found_by names.
		return findingResult{Label: label, FindingID: findingID}, nil
	}))

	c.Flags().String(flags.Key, "", flags.DescKey+"; the TOOL assigns the run-unique label <area>-F<n> (evidence-F1)")
	c.Flags().Var(&severity, flags.Severity, flags.GradeUsage("how bad this is"))
	c.Flags().Var(&likelihood, flags.Likelihood, flags.DescLikelihood)
	c.Flags().Var(&impact, flags.Impact, flags.DescImpact)
	// A finding anchors by --quote OR by --about-kind/--about, and the handler refuses one with
	// neither — so the marker states the condition, not the bare word.
	flags.Text(c, flags.Quote, "REQUIRED unless --about-kind/--about name the subject — "+flags.DescQuote+". The finding anchor is placed there")
	enumhelp.Flag(c, flags.AboutKind, record.MustEnum("finding", "about_kind"),
		"anchor this finding to something that is NOT report text — a section for what is missing from it, an avenue, or a gap already on the board; use instead of --quote. A finding about a gap reaches the seat that minted it, the one seat that can act on it")
	flags.Text(c, flags.About, "the reference --about-kind names: a section heading, an avenue id (Q1), or a gap id. It is CHECKED against the record")
	// The handler refuses a finding with no explanation; the marker says so where the seat reads.
	return seat.SaysRequired(c, flags.Reason)
}

// placementRefusal is the refusal `finding` and `mint` give for a quote Attach will not place; noun
// is what the quote anchors.
func placementRefusal(verb, noun string, err error) error {
	switch {
	case errors.Is(err, anchortext.ErrMisQuote):
		return fmt.Errorf("%s: --quote was not found in report.md.\n\nIt is matched LITERALLY against the report, so it must be the quoted text ALONE. A section heading in front of it (\"Findings: …\", \"## Method — …\") is the common cause and makes it match nothing — measured, four times in one sitting with four different separators. Name the section in --reason instead.\n\nA quote may not cross a blank line: a %s anchors ONE passage", verb, noun)
	case errors.Is(err, anchortext.ErrInFence):
		return fmt.Errorf("%s: the quote resolves inside a code fence — anchor a prose sentence, not code", verb)
	}
	return anchortext.Refusal(verb, err)
}

// placeOwed appends the Anchor a retried act owes, at the location its first call stored, once
// Attach accepts that location against the report as it stands; otherwise it returns Attach's
// refusal through refuse and the act stays unplaced.
func placeOwed(s seat.Context, run record.Run, id string, refuse func(error) error) error {
	loc, err := record.UnplacedLocation(run, id)
	if err != nil || loc == "" {
		return err
	}
	current, err := reportproj.RenderFromRecord(run)
	if err != nil {
		return err
	}
	if _, err := anchortext.Attach(current, id, loc); err != nil {
		return refuse(err)
	}
	_, err = record.Append(s.Identity(), &recordpb.Anchor{Id: proto.String(id), Location: proto.String(loc)})
	return err
}

type findingResult struct {
	Label      string `json:"label"`
	FindingID  string `json:"finding_id,omitempty"`
	Idempotent bool   `json:"idempotent,omitempty"`
}

func (r findingResult) Human() string {
	if r.Idempotent {
		return "finding " + r.Label + " (idempotent retry — existing label returned)"
	}
	return "finding recorded: " + r.Label + " — the run-unique label a gap's found_by names (id " + r.FindingID + ")"
}
