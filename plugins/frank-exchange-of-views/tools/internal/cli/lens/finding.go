package lens

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
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

	c := seat.Prose(seat.New("finding",
		`a lens finding (the tool assigns the L{role}-F{N} label): --key <your local F1> --severity <g> --likelihood <g> --impact <g> --reason "..." --location "<section + quoted sentence>"`,
		func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
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
			location := seat.Str(cmd, flags.Location)
			if strings.TrimSpace(location) == "" {
				return nil, fmt.Errorf("lens finding requires --location: a section heading plus a quoted sentence to anchor the finding")
			}
			// Crash-retry idempotency: a prior finding under this --key returns its
			// label, no second event AND no second marker (BEFORE any write).
			key := seat.Str(cmd, flags.Key)
			if prior, err := record.ExistingFindingByKey(s.RunDir, s.SeatID, key); err != nil {
				return nil, err
			} else if prior != "" {
				return findingResult{Label: prior, Idempotent: true}, nil
			}
			label, err := record.NextFindingLabel(s.RunDir, s.SeatID)
			if err != nil {
				return nil, err
			}
			// Mint the id UP FRONT: it forms the marker, so it must exist before the
			// report write. Append will not re-mint (finding_id already present).
			findingID := record.NewFindingID()
			marker := "[^" + findingID + "]"

			// Insert the marker at the located quote UNDER THE LOCK, atomically. NOT
			// FOUND -> reject (a mis-quote): the transform returns an error, nothing is
			// written, and the finding is not recorded. The events are appended ONLY
			// after a confirmed write (so a failed write leaves no id in EXPECTED).
			quote := extractQuote(location)
			if err := record.MutateBlueReport(s.RunDir, func(old []byte) ([]byte, error) {
				end := locateEnd(string(old), quote)
				if end < 0 {
					return nil, fmt.Errorf("lens finding: the quoted content was not found in report.md — quote the EXACT text you are flagging (via --location); check the section and wording")
				}
				if insideFence(string(old), end) {
					return nil, fmt.Errorf("lens finding: the quote resolves inside a code fence — anchor a prose sentence, not code")
				}
				return insertMarker(old, end, marker), nil
			}); err != nil {
				return nil, err
			}

			p := record.NewPayload().Set("label", label).Set("finding_id", findingID)
			seat.Set(cmd, p, "finding_key", flags.Key)
			seat.SetGrade(p, "severity", &severity)
			seat.SetGrade(p, "likelihood", &likelihood)
			seat.SetGrade(p, "impact", &impact)
			seat.SetSame(cmd, p, flags.Location)
			p.Set("text", text)
			if _, err := record.Append(s.RunDir, s.SeatID, "finding", p); err != nil {
				return nil, err
			}
			// The anchor event is EXPECTED for the immortal-marker detector: "finding
			// <id> has a marker at <location>". Keyed on the id (idempotent per finding).
			ap := record.NewPayload().Set("id", findingID).Set("location", location)
			if _, err := record.Append(s.RunDir, s.SeatID, "anchor", ap); err != nil {
				return nil, err
			}
			// The LABEL leads: it is the run-unique identity a gap's found_by names.
			return findingResult{Label: label, FindingID: findingID}, nil
		}))

	c.Flags().String(flags.Key, "", "a stable local handle (your own F1, F2 …) making a retried finding idempotent; the TOOL assigns the run-unique label L{role}-F{N}")
	c.Flags().Var(&severity, flags.Severity, flags.GradeUsage("how bad this is"))
	c.Flags().Var(&likelihood, flags.Likelihood, "how likely the CONSEQUENCE is (v2 grades consequence only, never existence)")
	c.Flags().Var(&impact, flags.Impact, "how bad the consequence is if it lands")
	c.Flags().String(flags.Location, "", "REQUIRED — where the defect lives: a section heading plus a QUOTED sentence (the exact text from the report; it anchors the finding-marker and is rejected if not found)")
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
