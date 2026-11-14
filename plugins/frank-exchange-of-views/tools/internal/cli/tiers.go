package cli

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cost"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// WHAT THIS RUN ASKED FOR, AND WHAT ANSWERED IT — side by side, because for six runs only the
// first half existed and the second was assumed to equal it.
//
// `inputs/run-config.json` records the REQUEST. Every surface downstream read it and described the
// run in its terms: the seats, the assembler, and one certified report which stated a blue/red
// pairing that had not argued the debate. What actually replied is on the record now, measured at
// each seat's `register` from that seat's own trajectory, so this is a join over two facts rather
// than a restatement of one.
//
// THE COVERAGE IS PART OF THE ANSWER. A seat whose serving model nobody measured prints NOT
// MEASURED, never the configured model and never a blank that reads as agreement — that
// substitution of a miss for a match is the whole reason this projection exists.
func newShowTiers() *cobra.Command {
	var format string
	c := &cobra.Command{
		Use:   "tiers",
		Short: "WHAT THIS RUN ASKED FOR AND WHAT ANSWERED IT: the configured tier per seat class beside the model actually served to each seat, from the record",
		Long: "tiers joins the run's CONFIGURED model tiers (inputs/run-config.json — the request) against the " +
			"model that actually answered each seat — read from that seat's own trajectory, which `register` names by " +
			"recording its agent id, and where the harness declares a substitution by naming both ends).\n\n" +
			"A seat nobody measured reads NOT MEASURED. It never reads as the configured model, and it never reads as " +
			"a blank that could be mistaken for agreement: a run where nothing looked is not a run that matched.",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			sc := seat.Of(cmd)
			// seat.Of ALREADY inferred if nothing was supplied — it ends its own resolution with
			// InferRunDir. What used to sit here was a SECOND inference, reached only when the
			// first had produced nothing, which by then meant the run had been REFUSED rather
			// than omitted. Inferring there resolved quietly to a different run than the one the
			// seat was told it could not have. Run() returns that refusal instead.
			run, err := sc.Run()
			if err != nil {
				return err
			}
			fam, err := record.FamilyOf(run)
			if err != nil {
				return err
			}
			rep := tierReport(run, fam)
			out := cmd.OutOrStdout()
			switch format {
			case "json":
				b, err := json.MarshalIndent(rep, "", "  ")
				if err != nil {
					return err
				}
				fmt.Fprintln(out, string(b))
			case "markdown":
				fmt.Fprint(out, renderTiers(rep))
			default:
				// Refused rather than defaulted, for the reason `ops diagnostics` gives: an
				// unknown --format that quietly emits JSON hands back something that parses in
				// answer to a question the tool does not have.
				return feov.Errorf(feov.Validation, "unknown --format %q (json | markdown)", format)
			}
			if rep.ConfigUnread != "" {
				return feov.Errorf(feov.Validation,
					"no seat was judged against a configured tier: %s", rep.ConfigUnread)
			}
			if rep.Substituted > 0 {
				return feov.Errorf(feov.Validation,
					"%d seat(s) were answered by a model other than their class's configured one", rep.Substituted)
			}
			return nil
		},
	}
	c.Flags().StringVar(&format, flags.Format, "json", "json (the form a reader joins on) | markdown")
	return c
}

// TierRow is one seat's request-versus-service, as fields.
type TierRow struct {
	SeatID string `json:"seatId"`
	Class  string `json:"class"`
	// Configured is the tier this seat's CLASS was dispatched on.
	Configured string `json:"configured,omitempty"`
	// Served is the model that answered. Empty means NOT MEASURED, which is why Measured exists
	// beside it rather than a reader inferring the state from the blank.
	Served   string `json:"served,omitempty"`
	Measured bool   `json:"measured"`
	// NotMeasured says why nothing answered, set exactly when Measured is false.
	NotMeasured string `json:"notMeasured,omitempty"`
	// Declared marks a substitution the harness itself announced, as against one this join
	// deduced by comparing tiers.
	Declared bool `json:"declared"`
	Matches  bool `json:"matches"`
}

// TierReport is the run's whole answer, request and service together.
type TierReport struct {
	RunDir          string `json:"runDir"`
	ConfiguredBulk  string `json:"configuredBulk,omitempty"`
	ConfiguredJudge string `json:"configuredJudgment,omitempty"`
	// ConfigUnread is why the run's configuration could not be read, or "". The served rows still
	// list — what answered each seat does not depend on the configuration — and none is judged.
	ConfigUnread string    `json:"configUnread,omitempty"`
	Seats        []TierRow `json:"seats"`
	TierBound    int       `json:"tierBound"`
	Measured     int       `json:"measured"`
	Substituted  int       `json:"substituted"`
}

func tierReport(run record.Run, b record.Family) TierReport {
	// A configuration that does not read is said, never shown as a run that configured no tier.
	var unread string
	bulk, judgment, err := cost.TierConfig(run)
	if err != nil {
		unread = err.Error()
	}
	rep := TierReport{RunDir: run.Dir(), ConfiguredBulk: bulk, ConfiguredJudge: judgment, ConfigUnread: unread}
	for _, sm := range record.SeatModels(run, b) {
		if sm.Class == "" {
			continue // the operator and anything off the roster ride no tier
		}
		// ONE JUDGMENT, capture's: a declared substitution counts on the same tier as on another.
		configured, v := cost.JudgeServed(sm, bulk, judgment)
		// MATCHES IS FALSE WHEN NOTHING WAS MEASURED, and that is the point of carrying Measured
		// separately: a reader that treats !Matches as "substituted" would over-report, and one
		// that treats an unmeasured seat as matching would rebuild the defect.
		row := TierRow{SeatID: sm.SeatID, Class: sm.Class, Configured: configured,
			Served: sm.Served, Measured: sm.Measured(), NotMeasured: sm.NotMeasured, Declared: sm.Substituted(),
			Matches: v == cost.ServedMatches}
		rep.TierBound++
		if row.Measured {
			rep.Measured++
		}
		if v == cost.ServedDeclared || v == cost.ServedOtherTier {
			rep.Substituted++
		}
		rep.Seats = append(rep.Seats, row)
	}
	sort.Slice(rep.Seats, func(i, j int) bool { return rep.Seats[i].SeatID < rep.Seats[j].SeatID })
	return rep
}

func renderTiers(r TierReport) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "# model tiers — requested and served\n\n")
	if r.ConfigUnread != "" {
		fmt.Fprintf(&sb, "configured: **NOT READ** — %s; no seat below is judged against a tier\n\n", r.ConfigUnread)
	} else {
		fmt.Fprintf(&sb, "configured: bulk=%s · judgment=%s\n\n", orUnset(r.ConfiguredBulk), orUnset(r.ConfiguredJudge))
	}
	fmt.Fprintf(&sb, "| seat | class | configured | served |\n|---|---|---|---|\n")
	// UNDER A CONFIGURATION NOT READ, NO ROW IS JUDGED: its configured cell says the configuration
	// was not read rather than that it declared nothing, and a served model is not marked a
	// mismatch against a tier nobody could read. A declared substitution is the harness's own
	// statement and is marked either way.
	unread := r.ConfigUnread != ""
	for _, s := range r.Seats {
		served := s.Served
		switch {
		case !s.Measured:
			served = "**NOT MEASURED** — " + s.NotMeasured
		case s.Declared:
			served = "**" + s.Served + "** (substitution declared by the harness)"
		case !s.Matches && !unread:
			served = "**" + s.Served + "**"
		}
		configured := orUnset(s.Configured)
		if unread {
			configured = "(not read)"
		}
		fmt.Fprintf(&sb, "| %s | %s | %s | %s |\n", s.SeatID, s.Class, configured, served)
	}
	if unread {
		fmt.Fprintf(&sb, "\nserved measured on %d of %d tier-bound seat(s); none judged against a configured tier, which was not read\n",
			r.Measured, r.TierBound)
	} else {
		fmt.Fprintf(&sb, "\nserved measured on %d of %d tier-bound seat(s); %d answered by a model other than the configured one\n",
			r.Measured, r.TierBound, r.Substituted)
	}
	if r.TierBound > 0 && r.Measured == 0 {
		fmt.Fprintf(&sb, "\nNOTHING LOOKED. This run predates the served-model field or its seats' trajectories were unreadable, which is NOT the same as a run that matched its configuration.\n")
	}
	return sb.String()
}

func orUnset(s string) string {
	if s == "" {
		return "(not declared)"
	}
	return s
}
