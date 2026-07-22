package merge

import (
	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// mint: put a gap on the board.
//
// The id is TOOL-assigned, sequential per round. Letting seats choose produced
// four different "R5-1"s in one round of run 3, and the collision class dies here
// rather than being policed downstream.
func newMint() *cobra.Command {
	var severity, likelihood, impact, cx flags.GradeValue
	var supersedes, foundBy flags.CSV

	c := seat.Prose(seat.New("mint",
		`mint a board gap (id is TOOL-assigned; --key <stable-label> makes retries idempotent): --class <slug>|--class-new <slug> --definition --neighbor --distinguisher, --location "..." --problem "..."|--reason --fix "..." --check "<acceptance check red runs at re-audit>" --severity/--likelihood/--impact/--cx <grade> [--supersedes R1-2,R1-7] [--found-by L5-F3,L6-F2]`,
		func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
			// Crash-retry idempotency: --key (the stable local label, e.g. the source
			// lens finding) makes a retried mint return the EXISTING id.
			prior, err := record.ExistingMintByKey(s.RunDir, s.SeatID, seat.Str(cmd, flags.Key))
			if err != nil {
				return nil, err
			}
			if prior != "" {
				return mintResult{GapID: prior, Idempotent: true}, nil
			}
			gapID, err := record.MintGapID(s.RunDir, record.RoundOf(s.SeatID))
			if err != nil {
				return nil, err
			}
			text, err := seat.Reason(cmd)
			if err != nil {
				return nil, err
			}
			problem := seat.Str(cmd, flags.Problem)
			if problem == "" {
				problem = text
			}

			p := record.NewPayload().Set("gap_id", gapID)
			seat.Set(cmd, p, "mint_key", flags.Key)
			// --class-new both names the class and mints it, so it wins over --class.
			if seat.Given(cmd, flags.ClassNew) {
				p.Set("class", seat.Str(cmd, flags.ClassNew))
			} else {
				seat.Set(cmd, p, "class", flags.Class)
			}
			p.Set("class_new", seat.Given(cmd, flags.ClassNew))
			seat.SetSame(cmd, p, flags.Definition, flags.Neighbor, flags.Distinguisher, flags.Location)
			p.Set("problem", problem)
			seat.Set(cmd, p, "required_fix", flags.Fix)
			seat.Set(cmd, p, "acceptance_check", flags.Check)
			seat.SetGrade(p, "severity", &severity)
			seat.SetGrade(p, "likelihood", &likelihood)
			seat.SetGrade(p, "impact", &impact)
			seat.SetGrade(p, "complexity_cost", &cx)
			seat.SetList(p, "supersedes", &supersedes)
			seat.SetList(p, "found_by", &foundBy)

			if _, err := record.Append(s.RunDir, s.SeatID, "mint", p); err != nil {
				return nil, err
			}
			if seat.Given(cmd, flags.ClassNew) {
				cn := record.NewPayload().Set("slug", seat.Str(cmd, flags.ClassNew))
				seat.SetSame(cmd, cn, flags.Definition, flags.Neighbor, flags.Distinguisher)
				if _, err := record.Append(s.RunDir, s.SeatID, "class-new", cn); err != nil {
					return nil, err
				}
			}
			return mintResult{GapID: gapID}, nil
		}))

	c.Flags().String(flags.Key, "", "a stable local label (e.g. the source lens finding) making a retried mint idempotent")
	c.Flags().String(flags.Class, "", "the gap's class slug from the registry — what KIND of defect this is")
	c.Flags().String(flags.ClassNew, "", "mint a new class slug; requires --definition, --neighbor and --distinguisher")
	c.Flags().String(flags.Definition, "", "what the new class is, in one line")
	c.Flags().String(flags.Neighbor, "", "the existing class this one sits closest to")
	c.Flags().String(flags.Distinguisher, "", "the tie-break question that tells the two apart")
	c.Flags().String(flags.Location, "", "where the defect lives: a section heading plus a quoted sentence")
	c.Flags().String(flags.Problem, "", "what is wrong (or pass it via --reason)")
	c.Flags().String(flags.Fix, "", "the required fix")
	c.Flags().String(flags.Check, "", "the acceptance check red will RUN at re-audit — the pre-agreed contract, not a description")
	c.Flags().Var(&severity, flags.Severity, flags.GradeUsage("how bad this is"))
	c.Flags().Var(&likelihood, flags.Likelihood, "how likely the CONSEQUENCE is (v2 grades consequence only, never existence)")
	c.Flags().Var(&impact, flags.Impact, "how bad the consequence is if it lands")
	c.Flags().Var(&cx, flags.Complexity, "complexity_cost — what fixing it costs, on the same scale")
	c.Flags().Var(&supersedes, flags.Supersedes, "comma-separated ancestor ids this gap replaces; lineage is never dropped")
	c.Flags().Var(&foundBy, flags.FoundBy, "comma-separated lens findings that surfaced it (L5-F3,L6-F2)")
	return c
}

// mintResult reports the tool-assigned gap id; a retry returns the existing one.
type mintResult struct {
	GapID      string `json:"gap_id"`
	Idempotent bool   `json:"idempotent,omitempty"`
}

func (r mintResult) Human() string {
	if r.Idempotent {
		return "minted " + r.GapID + " (idempotent retry — existing id returned)"
	}
	return "minted " + r.GapID
}
