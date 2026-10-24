package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchortext"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/report"
)

// newEngineStopped is the operator's record of the engine's own stop: the error it threw, written
// to the log.
//
// # Why the lead writes it
//
// The engine stops a run by throwing — a seat that returned nothing, a relay it refuses. It runs
// in the Workflow sandbox, which reaches no file and no process, so it cannot write the stop
// itself; and the workflow runs in the background, so its failure reaches the lead as a task
// notification and no PostToolUseFailure hook fires for it. The lead holds the error and is the
// only party that can put it on the record.
//
// # Why it is a log entry and not an outcome
//
// An outcome closes the record to seat acts, and a run the engine stopped is one an operator can
// still resume. Without an entry, its record reads exactly as a run cut off mid-sitting does: the
// last sitting, then nothing.
//
// # Why `failure` under the harness seat, and why its own verb
//
// `failure` is the type the tool already writes for a call that failed where no verb could see
// it, and the entry is the harness's, as the cast and the sitting spans are: no seat's call
// failed. A `failure` under record.HarnessSeat is therefore the stop, as fields — never a phrase
// recovered from the text.
//
// NOT `log`. That word is the seats' write verb, and the suite holds it to one meaning: its
// `--type` set is what a seat may file, its `--reason` is a seat's conclusion about the tooling,
// and the operator's root carries no `log` so that the word never names a read. An operator `log`
// taking the one type no seat may file would be the same word in a second sense.
func newEngineStopped() *cobra.Command {
	failure := recordpb.LogType_LOG_TYPE_FAILURE
	c := &cobra.Command{
		Use:   "engine-stopped",
		Short: "record that the engine stopped the run on an error (operator; writes the error to the log and NO outcome, so the run stays open to resume)",
		Long: "engine-stopped --reason \"<the error>\" writes the error the engine stopped on to the run's log, as the tool's own `" +
			recordpb.Word(failure) + "` entry under `" + record.HarnessSeat + "`.\n\n" +
			"Run it when the Workflow tool reports that the debate FAILED after seats sat. Give the error verbatim: it is the " +
			"only account of why the run stopped, and with it on the record a run the engine stopped reads differently from one " +
			"that was cut off.\n\n" +
			"IT RECORDS NO OUTCOME. An outcome closes the record to seat acts; this leaves it open, so the same workflow run " +
			"resumes against it.\n\n" +
			"The answer says which of three states the run is in:\n" +
			"  - no outcome: the run is open to its seats, and the same workflow run resumes against it;\n" +
			"  - the outcome, and no assembled documents: the assembly sitting stamped the outcome and stopped before " +
			"the bench's `assemble` wrote the run's documents, so the assembly is unfinished and the same workflow run resumes it — " +
			"the bench sits again to assemble;\n" +
			"  - the outcome, and the assembled documents: the assembly finished, there is nothing to resume, and the run " +
			"is captured.\n" +
			"The assembly's mark is its documents, because the bench's `assemble` writes them and no event: the assembly sitting's " +
			"stop is no mark, since the sitting stops whether or not the assembly ran. Capture re-assembles the documents from " +
			"the record at the close. A run directory that cannot be read is said so, and is never read as finished.\n\n" +
			"`ops log` reads the entry back, beside what the seats filed.",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			run, err := seat.Of(cmd).Run()
			if err != nil {
				return err
			}
			text, err := flags.ReadPayload(cmd)
			if err != nil {
				return err
			}
			// AN ERROR CAN NAME AN ANCHOR, and the entry names it by its id (anchortext.LogLine).
			if text = strings.TrimSpace(anchortext.LogLine(text)); text == "" {
				return feov.Errorf(feov.MissingField, "requires --reason: the error the engine stopped on, verbatim — an entry without it says a run stopped and not why")
			}
			if _, err := record.Append(record.Identity{Run: run, SeatID: record.HarnessSeat}, &recordpb.Log{
				Text:   proto.String(text),
				Type:   failure.Enum(),
				Source: recordpb.LogSource_LOG_SOURCE_TOOL.Enum(),
			}); err != nil {
				return err
			}
			// THE RECORD SAYS WHETHER THERE IS ANYTHING TO RESUME, so the lead relays it rather than
			// inferring it from the error's wording.
			fmt.Fprintf(cmd.OutOrStdout(), "the engine's stop is on the record — %s\n", stoppedState(run))
			return nil
		},
	}
	flags.RegisterPayload(c)
	c.Flags().Lookup(flags.Reason).Usage = "the error the engine stopped on, verbatim"
	return c
}

// stoppedState is what a stopped run leaves to do, read from the record and the assembly's
// documents.
//
// THE OUTCOME IS NOT THE END OF THE RUN. The assembly sitting stamps it first and then assembles
// (debate.js, the assembly prompt), so an engine stop after the outcome is an assembly sitting that
// returned nothing — and whether it assembled before it stopped is what decides resume or capture.
// The record cannot say: the bench's `assemble` writes documents and no event, and the sitting's span
// closes whether or not it assembled. So the documents are read, by the same fact capture reads
// (report.AssemblyWritten), and a directory that cannot be read is said so rather than read as
// either answer.
func stoppedState(run record.Run) string {
	v, err := record.RecordedOutcome(run)
	switch {
	case err != nil:
		return "whether the record holds an outcome could not be read: " + err.Error()
	case v == "":
		return "the record holds NO outcome: the run is open to its seats, and the same workflow run resumes against it"
	}
	// The seat's spelling, as every other operator surface shows it.
	held := "the record holds the outcome " + strings.ToUpper(v) + ", stamped by the assembly sitting"
	switch written, err := report.AssemblyWritten(run); {
	case err != nil:
		return held + ", and whether it assembled the run's documents could not be read: " + err.Error() +
			" — the assembly is NOT shown finished, and the run is neither resumed nor captured on a guess"
	case !written:
		return held + ", and the run's documents were never assembled: the assembly sitting stopped before it finished, " +
			"and the same workflow run resumes it — the bench sits again to assemble"
	}
	return held + ", and the run's documents are assembled: the assembly finished, there is nothing to resume, and the run is captured"
}
