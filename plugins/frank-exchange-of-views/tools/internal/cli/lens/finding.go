package lens

import (
	"errors"
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
		`a lens finding (the tool assigns the L{role}-F{N} label): --key <your local F1> --severity <g> --likelihood <g> --impact <g> --reason "..." --quote "<the exact sentence, quoted from the report and nothing else>". `+
			`A FINDING IS A DEFECT IN THE TEXT — whether what the report SAYS stands up as written. It is not a verdict on the world. `+
			"For a claim resting on NO citation, that distinction decides the verb: where the source is one you could go and read, `corroborate` "+
			`puts it on the record and the finding then says what it showed; where there is nothing to fetch — the source cannot be obtained, no `+
			`source exists to look for, or the defect IS the writing (a figure with no derivation, an internal contradiction, an unfalsifiable `+
			"universal) — raise it here. `verify` and `reproduce` judge evidence blue already produced. "+
			`NOTE THE ASYMMETRY RATHER THAN LETTING IT CHOOSE: this verb cannot fail, needing no fetch and risking no `+"`absent`"+`, and the `+
			`evidence verbs can.`,
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
			location := seat.Str(cmd, flags.Quote)
			if strings.TrimSpace(location) == "" {
				return nil, fmt.Errorf("lens finding requires --quote: the EXACT text you are flagging, quoted from the report and nothing else — it is matched against blue/report.md to place the marker")
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
			if err := record.MutateBlueReport(s.RunDir, func(old []byte) ([]byte, error) {
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

			p := record.NewPayload().Set("label", label).Set("finding_id", findingID)
			seat.Set(cmd, p, "finding_key", flags.Key)
			seat.SetGrade(p, "severity", &severity)
			seat.SetGrade(p, "likelihood", &likelihood)
			seat.SetGrade(p, "impact", &impact)
			seat.Set(cmd, p, "location", flags.Quote)
			p.Set("reason", text)
			if _, err := record.Append(s.Identity(), "finding", p); err != nil {
				return nil, err
			}
			// The anchor event is EXPECTED for the immortal-marker detector: "finding
			// <id> has a marker at <location>". Keyed on the id (idempotent per finding).
			ap := record.NewPayload().Set("id", findingID).Set("location", location)
			if _, err := record.Append(s.Identity(), "anchor", ap); err != nil {
				return nil, err
			}
			// The LABEL leads: it is the run-unique identity a gap's found_by names.
			return findingResult{Label: label, FindingID: findingID}, nil
		}))

	c.Flags().String(flags.Key, "", flags.DescKey+"; the TOOL assigns the run-unique label L{role}-F{N}")
	c.Flags().Var(&severity, flags.Severity, flags.GradeUsage("how bad this is"))
	c.Flags().Var(&likelihood, flags.Likelihood, "how likely the CONSEQUENCE is — never how likely the defect is to BE there, which is what one grade meant before v2 split them")
	c.Flags().Var(&impact, flags.Impact, "how bad the consequence is if it lands")
	c.Flags().String(flags.Quote, "", "REQUIRED — "+flags.DescQuote+". The finding-marker is placed there")
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
