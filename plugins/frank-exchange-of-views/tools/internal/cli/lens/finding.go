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

// finding: a lens's graded observation, for the merge to dispose.
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
				"An ABSENCE has no sentence to quote. Borrowing an innocent one as a handle is what this " +
				"used to force — a missing line of inquiry pinned to a sentence the finding itself called " +
				"fine — and a reader of the gap list then lands on good prose. Anchor it to the section it " +
				"is missing from, the line of inquiry whose reason you are arguing against, or the gap it is about")
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
			// THE PAIR MAY BE HALF-APPENDED. The finding and its anchor event follow the splice
			// as two separate appends, so a crash between them leaves the finding recorded and
			// the anchor event missing — and this early return used to seal that state forever:
			// the retry saw the finding, answered idempotently, and the immortal-marker detector
			// never learned the marker exists. The retry now finishes the interrupted pair.
			if anchored, aerr := record.AnchorEventExists(run, priorID); aerr != nil {
				return nil, aerr
			} else if !anchored {
				ap := &recordpb.Anchor{Id: proto.String(priorID), Location: proto.String(seat.Str(cmd, flags.Quote))}
				if _, aerr := record.Append(s.Identity(), ap); aerr != nil {
					return nil, aerr
				}
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
		// An INVISIBLE HTML-comment token, not a footnote: a "[^id]" marker rendered
		// as an undefined footnote AND red audited it as one, and a finding's quoted
		// location/reason text carried the marker into the record-derived sections. A
		// comment renders as nothing and is no footnote, so no seat audits it.
		marker := "<!--fx:" + findingID + "-->"

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
			if _, aerr := anchortext.InsertAnchor([]byte(current), location, marker); aerr != nil {
				switch {
				case errors.Is(aerr, anchortext.ErrMisQuote):
					return nil, fmt.Errorf("lens finding: --quote was not found in report.md.\n\nIt is matched LITERALLY against the report, so it must be the quoted text ALONE. A section heading in front of it (\"Findings: …\", \"## Method — …\") is the common cause and makes it match nothing — measured, four times in one sitting with four different separators. Name the section in --reason instead.\n\nA quote may not cross a blank line: a finding anchors ONE passage")
				case errors.Is(aerr, anchortext.ErrInFence):
					return nil, fmt.Errorf("lens finding: the quote resolves inside a code fence — anchor a prose sentence, not code")
				}
				return nil, aerr
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
		// NO MARKER FOR AN ABSENCE, and that is the point rather than an omission. The anchor
		// event exists so the immortal-marker detector can say "finding <id> has a marker at
		// <location>"; a finding about something NOT in the report has no location to mark, and
		// splicing one would put a marker on the innocent prose this change exists to stop
		// borrowing.
		if !aboutSet {
			ap := &recordpb.Anchor{Id: proto.String(findingID), Location: proto.String(location)}
			if _, err := record.Append(s.Identity(), ap); err != nil {
				return nil, err
			}
		}
		// The LABEL leads: it is the run-unique identity a gap's found_by names.
		return findingResult{Label: label, FindingID: findingID}, nil
	}))

	c.Flags().String(flags.Key, "", flags.DescKey+"; the TOOL assigns the run-unique label L{role}-F{N}")
	c.Flags().Var(&severity, flags.Severity, flags.GradeUsage("how bad this is"))
	c.Flags().Var(&likelihood, flags.Likelihood, "how likely the CONSEQUENCE is — never how likely the defect is to BE there, which is what one grade meant before v2 split them")
	c.Flags().Var(&impact, flags.Impact, "how bad the consequence is if it lands")
	c.Flags().String(flags.Quote, "", flags.DescQuote+". The finding-marker is placed there")
	enumhelp.Flag(c, flags.AboutKind, record.MustEnum("finding", "about_kind"),
		"anchor this finding to something that is NOT report text — use instead of --quote when the defect is an ABSENCE")
	c.Flags().String(flags.About, "", "the reference --about-kind names: a section heading, an avenue id, or a gap id. "+
		"It is CHECKED against the record, which a borrowed quote never was")
	return c
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
