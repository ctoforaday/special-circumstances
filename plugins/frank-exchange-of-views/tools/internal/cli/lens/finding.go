package lens

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
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
		if strings.TrimSpace(location) == "" {
			return nil, fmt.Errorf("lens finding requires --quote: the EXACT text you are flagging, quoted from the report and nothing else — it is matched against blue/report.md to place the marker")
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
		// A TORN SPLICE IS ADOPTED, NOT DOUBLED — the same rule as `blue cite`, through the
		// same shared walk. A crash after the splice below and before the appends leaves a
		// marker on this quote that no event names; the retry would otherwise mint a fresh id
		// and splice a rival beside the immortal orphan.
		findingID := adoptTornFindingAnchor(run, location)
		spliced := findingID != ""
		if findingID == "" {
			// Mint the id UP FRONT: it forms the marker, so it must exist before the
			// report write. Append will not re-mint (finding_id already present).
			findingID = record.NewFindingID()
		}
		// An INVISIBLE HTML-comment token, not a footnote: a "[^id]" marker rendered
		// as an undefined footnote AND red audited it as one, and a finding's quoted
		// location/reason text carried the marker into the record-derived sections. A
		// comment renders as nothing and is no footnote, so no seat audits it.
		marker := "<!--fx:" + findingID + "-->"

		// Insert the marker at the located quote UNDER THE LOCK, atomically, via the
		// shared anchor-insert (the same rule blue cite anchors a citation by). NOT
		// FOUND -> reject (a mis-quote): the transform returns an error, nothing is
		// written, and the finding is not recorded. The events are appended ONLY
		// after a confirmed write (so a failed write leaves no id in EXPECTED).
		if err := record.MutateBlueReport(run, func(old []byte) ([]byte, error) {
			if spliced {
				return old, nil // the crashed first attempt already placed this marker
			}
			next, err := InsertAnchor(old, location, marker)
			switch {
			case errors.Is(err, ErrMisQuote):
				return nil, fmt.Errorf("lens finding: --quote was not found in report.md.\n\nIt is matched LITERALLY against the report, so it must be the quoted text ALONE. A section heading in front of it (\"Findings: …\", \"## Method — …\") is the common cause and makes it match nothing — measured, four times in one sitting with four different separators. Name the section in --reason instead.\n\nA quote may not cross a blank line: a finding anchors ONE passage")
			case errors.Is(err, ErrInFence):
				return nil, fmt.Errorf("lens finding: the quote resolves inside a code fence — anchor a prose sentence, not code")
			}
			return next, err
		}); err != nil {
			return nil, err
		}

		body := &recordpb.Finding{
			Label:      proto.String(label),
			FindingId:  proto.String(findingID),
			FindingKey: proto.String(seat.Str(cmd, flags.Key)),
			Location:   proto.String(seat.Str(cmd, flags.Quote)),
			Text:       proto.String(text),
			Severity:   seat.GradeOrNil(&severity),
			Likelihood: seat.GradeOrNil(&likelihood),
			Impact:     seat.GradeOrNil(&impact),
		}
		if _, err := record.Append(s.Identity(), body); err != nil {
			return nil, err
		}
		// The anchor event is EXPECTED for the immortal-marker detector: "finding
		// <id> has a marker at <location>". Keyed on the id (idempotent per finding).
		ap := &recordpb.Anchor{Id: proto.String(findingID), Location: proto.String(location)}
		if _, err := record.Append(s.Identity(), ap); err != nil {
			return nil, err
		}
		// The LABEL leads: it is the run-unique identity a gap's found_by names.
		return findingResult{Label: label, FindingID: findingID}, nil
	}))

	c.Flags().String(flags.Key, "", flags.DescKey+"; the TOOL assigns the run-unique label L{role}-F{N}")
	c.Flags().Var(&severity, flags.Severity, flags.GradeUsage("how bad this is"))
	c.Flags().Var(&likelihood, flags.Likelihood, "how likely the CONSEQUENCE is — never how likely the defect is to BE there, which is what one grade meant before v2 split them")
	c.Flags().Var(&impact, flags.Impact, "how bad the consequence is if it lands")
	c.Flags().String(flags.Quote, "", flags.DescQuote+". The finding-marker is placed there")
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

// adoptTornFindingAnchor returns the id of a finding marker already on the located quote that no
// finding OR anchor event names — a torn splice — or "" for the ordinary fresh path.
func adoptTornFindingAnchor(run record.Run, quote string) string {
	rep, err := record.ReadBlueReport(run)
	if err != nil {
		return ""
	}
	m, err := record.MergedEvents(run)
	if err != nil {
		return ""
	}
	recorded := map[string]bool{}
	for _, e := range m.Events {
		if f, ok := recordpb.BodyAs[*recordpb.Finding](e); ok {
			recorded[f.GetFindingId()] = true
		}
		if a, ok := recordpb.BodyAs[*recordpb.Anchor](e); ok {
			recorded[a.GetId()] = true
		}
	}
	return OrphanAnchorAt(string(rep), quote, "fx", func(id string) bool { return recorded[id] })
}
