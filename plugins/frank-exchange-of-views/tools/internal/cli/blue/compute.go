package blue

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchortext"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/compute"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/reportvoice"
)

// compute: settle a claim by COMPUTING it, and leave the computation as the evidence.
//
// THE ENGINE COULD ALWAYS DO THIS AND NEVER DID. Every seat has Write, Edit and Bash; across
// six runs of recorded events no seat ever wrote a program to settle a question. Nothing tied
// its hands — nothing invited it either, and the surrounding protocol (search, cite, fetch at
// the leaf, never improvise) reads as "evidence means sources".
//
// Measured cost of that: the 2026-08-04 smoke's R1-2 asked blue to test the protocol on a
// false claim ("is 9 prime"). Blue answered in PROSE, asserting the test had happened. R2-2
// refused it — "no evidence shown" — and the epoch was spent. Three lines of trial division
// settle it and leave an artifact red can re-run.
//
// IT IS THE CITATION MODEL WITH THE LAST MILE WALKED, not a new kind of evidence. Cite the
// METHOD with `blue cite` (trial division, Miller-Rabin — someone else's proof, sourced), then
// compute the INSTANCE by running it. Several methods on one claim is several anchors on one
// sentence, which is exactly how a mathematician argues.
//
// THE BASIS IS DERIVED, never claimed: the tool runs the program twice. Two runs that wrote the
// same bytes are `reproducible` — the auditor can reach them independently. Two that differed are
// refused unless the seat states what holds across them (--variance), and then they are `variant`:
// still evidence, of a system in motion (a live call measures what an API actually does, which
// beats its documentation), and the report must not read the two alike.
func newCompute() *cobra.Command {
	c := seat.Prose(seat.New("compute", computeResult{}, func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
		run, err := s.Run()
		if err != nil {
			return nil, err
		}
		location, dir := seat.Location(cmd), seat.Str(cmd, flags.Dir)
		if strings.TrimSpace(location) == "" {
			return nil, feov.Errorf(feov.MissingField, "requires --quote: the EXACT sentence in the report (as `show report` serves it) this computation backs — a computation anchored to nothing is a program nobody can connect to a claim")
		}
		if strings.TrimSpace(dir) == "" {
			return nil, feov.Errorf(feov.MissingField, "requires --dir: the directory holding the program and the files it reads — absolute, or relative to the run directory")
		}
		argv, err := commandArgs(cmd)
		if err != nil {
			return nil, err
		}
		if !filepath.IsAbs(dir) {
			dir = filepath.Join(run.Dir(), dir)
		}
		// The bundle is read once, here, and every file is checked against the record BEFORE it is
		// opened: a read of a record file in this process releases the locks its connection holds.
		prog, err := compute.ReadProgram(dir, argv, func(path string) error {
			return record.RefuseRecordFile(run, path, "--dir", "a bundle file")
		})
		if err != nil {
			return nil, err
		}

		// CRASH-RETRY AND KEY COLLISION LOOK IDENTICAL AT THE KEY, AND ARE OPPOSITE FACTS.
		//
		// A retry is the SAME program arriving twice: returning the recorded sha is right, and
		// re-running it would splice a second anchor. A collision is a DIFFERENT program under a
		// key an earlier dispatch of this seat already burned — and this used to take the same
		// path, answering "already recorded as <sha>" and recording NOTHING. A seat that believed
		// the message shipped a report claiming a computation it does not hold, and a resumed dispatch
		// cannot see which keys its earlier self used. Measured in
		// research/2026-09-02_quadratic-formula (blue-respond, --key P1).
		//
		// THE NOTE IS PRINTED IN THE REPORT — as this computation's [^KN] footnote and its evidence.md
		// heading (report/computations.go) — so it meets the same voice advisory as edit's --new. ADVISORY,
		// COMPUTED BEFORE ANY RETURN so an idempotent retry carries it too; it refuses nothing.
		why, err := seat.Reason(cmd)
		if err != nil {
			return nil, err
		}
		tells := spanVoiceTells(why)

		// The bundle's own sha settles it — its files and its command — and costs a read rather than an
		// execution.
		key := seat.Str(cmd, flags.Key)
		if prior, err := record.ExistingComputationByKey(run, s.SeatID, key); err != nil {
			return nil, err
		} else if prior != nil {
			if now := prog.Sha; now != prior.GetBundleSha() {
				return nil, feov.Errorf(feov.Conflict, "--key %q already names computation %s on this run, and %s run as %s hashes to %s — "+
					"a key is one program's name, not a slot to reuse. This is not a retry: recording nothing here would "+
					"leave you citing a computation of something else. Give this program its own --key",
					key, compute.Short(prior.GetBundleSha()), dir, prog, compute.Short(now))
			}
			// THE SAME PROGRAM IS NOT YET THE SAME ACT: a retry states the sentence it backs, the
			// method it cites, the gap it answers and its note again, or it is a second computation reusing
			// the name.
			retry := computeRetry(cmd, why)
			if err := seat.RepeatsKeyed(key, "computation", prior.GetComputationId(), prior, retry, computeAsserted...); err != nil {
				return nil, err
			}
			// A computation on the record without its anchor — one migrate brought forward from an archive
			// whose run crashed between the two — is finished by the retry, at the location it states.
			if err := seat.PlaceOwed(s, prior.GetComputationId(), key); err != nil {
				return nil, err
			}
			return computeResult{SHA: prior.GetBundleSha(), Idempotent: true, VoiceTells: tells}, nil
		}

		res, err := compute.Run(run.Dir(), prog)
		if err != nil {
			// A computation that will not run is not evidence, and the failure is a capability
			// signal — the same treatment `blue cite` gives an unreachable source.
			msg := err.Error()
			if _, ferr := record.Append(s.Identity(), &recordpb.Log{Text: proto.String(msg), Type: recordpb.LogType_LOG_TYPE_DEFECT.Enum(), Source: recordpb.LogSource_LOG_SOURCE_TOOL.Enum()}); ferr != nil {
				return nil, ferr
			}
			return nil, record.ToolLogged(err)
		}

		// THE ENVIRONMENT FAILING IS NOT THE COMPUTATION ANSWERING, and exit 0 does not tell the
		// two apart. A program runs in a copy of its bundle; one authored against the repo root
		// prints "No such file or directory", carries on, and ends clean — which is how a report
		// came to cite an enumeration that enumerated nothing as re-runnable evidence.
		// Refused rather than graded, because the artifact is otherwise indistinguishable from a
		// real negative result, and the seat is one edit away from a correct one.
		expectError, err := cmd.Flags().GetBool(flags.ExpectError)
		if err != nil {
			return nil, err
		}
		// A BUNDLE THAT CALLS THE TOOL CANNOT BE CHECKED BY ANYONE: the tool refuses every call made
		// while a checker runs it, so the stored copy's run said nothing a reader can repeat.
		if len(res.Calls) > 0 {
			return nil, feov.Errorf(feov.Validation, "%s calls feov-record (%s), which refuses every call made while a checker runs the program — no one else could run it to this output. A computation reads its inputs from files in its bundle, never from the debate's tool", prog, res.Calls[0])
		}
		// TWO RUNS THAT DIFFERED ARE A MEASUREMENT ONLY WHEN THE SEAT SAYS WHAT HOLDS ACROSS THEM, and
		// --variance over two runs that agreed is a claim about a difference that did not happen.
		variance := strings.TrimSpace(seat.Str(cmd, flags.Variance))
		if res.Drift != "" && variance == "" {
			return nil, feov.Errorf(feov.Validation, "%s did not write the same bytes twice — %s. Make its output deterministic (seed it, sort it, drop the timestamp), or state with --variance what holds across the runs, and it is recorded as `variant`", prog, res.Drift)
		}
		if res.Drift == "" && variance != "" {
			return nil, feov.Errorf(feov.Validation, "--variance states what holds when the runs differ, and %s wrote the same bytes on every run — drop --variance and it is recorded as `reproducible`", prog)
		}
		// --expect-error IS A CLAIM ABOUT THE RUN, and a run that failed nothing makes it false: the
		// flag would be accepted and used for nothing.
		if expectError && res.Failed == "" {
			return nil, feov.Errorf(feov.Validation, "--expect-error says this computation's point is a failing command, but %s ran to exit %d with no environment error in its output — drop --expect-error, or fix the program so it captures the failure it is meant to prove", prog, res.Exit)
		}
		if res.Failed != "" && !expectError {
			return nil, feov.Errorf(feov.Validation, "%s ran to exit %d, but its output carries an ENVIRONMENT error, not a result:\n    %s\n"+
				"A computation runs in a fresh copy of its bundle, outside the run directory, and its working directory is that copy (%s in the output) — a file it reads belongs in the bundle. "+
				"Fix the program and re-run; if capturing this failure IS the point (proving a path is absent, a command missing), say so with --expect-error and it is recorded as the result it is",
				prog, res.Exit, res.Failed, compute.BundleToken)
		}

		// THE ANCHOR EVENT BELOW IS THE MARKER: it carries the quote, and reportproj.Render re-places
		// the computation anchor at replay. No file is spliced; a --key retry is idempotent (handled
		// above). Mint the id; the placement is validated against the report the Computation and its Anchor
		// land on — a mis-quote or in-fence quote is refused and nothing recorded.
		label := record.NewID("computation")

		// `output` does not survive onto the event: it is CONTENT, stored as this act's own under
		// acts/<computation_id>/ before the event is appended, and taken back if the append is
		// refused, so no act directory is left that no event names. `location` DOES survive: under
		// report-as-record the projection re-places this computation's marker by re-locating it.
		body := &recordpb.Computation{
			ComputationId: proto.String(label),
			BundleSha:     proto.String(res.Bundle),
			Argv:          res.Argv,
			Basis:         recordpb.Basis_BASIS_REPRODUCIBLE.Enum(),
			Exit:          proto.Int32(int32(res.Exit)),
		}
		if res.ToolchainPath != "" {
			body.ToolchainPath, body.ToolchainSha256 = proto.String(res.ToolchainPath), proto.String(res.ToolchainSha256)
		}
		if res.Drift != "" {
			body.Basis = recordpb.Basis_BASIS_VARIANT.Enum()
			body.Drift = proto.String(res.Drift)
			body.Invariant = proto.String(variance)
		}
		computeFields(cmd, body, why)
		// The store before the event: the bundle once under dir/, then this act's output and out/.
		if err := compute.WriteBundle(run.Dir(), res.Bundle, res.Argv, prog.Files); err != nil {
			return nil, err
		}
		if err := compute.WriteAct(run.Dir(), res.Bundle, label, []byte(res.Output), res.Out); err != nil {
			return nil, err
		}
		if _, err := seat.AppendPlaced(s, body, label, location, computeRefusal); err != nil {
			if rerr := compute.RemoveAct(run.Dir(), res.Bundle, label); rerr != nil {
				return nil, errors.Join(err, rerr)
			}
			return nil, err
		}
		// COPIES LEFT BEHIND ARE THE TOOL'S DEFECT, NOT THE COMPUTATION'S. Both runs finished and
		// every recorded field was measured before the removal, so the act stands; the leftover is
		// logged after it, best-effort, as the tool's own failure.
		if res.Leftover != "" {
			_, _ = record.Append(s.Identity(), &recordpb.Log{Text: proto.String(res.Leftover), Type: recordpb.LogType_LOG_TYPE_DEFECT.Enum(), Source: recordpb.LogSource_LOG_SOURCE_TOOL.Enum()})
		}
		return computeResult{Label: label, SHA: res.Bundle, Basis: recordpb.Word(body.GetBasis()), Exit: res.Exit, Drift: res.Drift, Leftover: res.Leftover, VoiceTells: tells}, nil
	}))

	// DescQuoteAlone: this verb's --reason is the computation note the report prints, no place for a section name.
	flags.Text(c, flags.Quote, flags.DescQuoteAlone)
	c.Flags().String(flags.Dir, "", "the directory holding the program and every file it reads — its bundle, copied whole (at most 1 MiB, no symlinks) into a fresh directory for each run. A relative path is read from the run directory, so a bundle drafted in the run's scratch/ is passed as scratch/<name>; an absolute path is read where it lies. The tool keeps its own copy with the computation, so the computation does not depend on the draft")
	seat.Require(c, flags.Quote, flags.Dir)
	c.Use = "compute [flags] -- <command> [args...]"
	c.Args = cobra.ArbitraryArgs
	c.Flags().Var(flags.CitationAnchor().WithCheck(record.CitationExists(record.ComputeCitesAdvice)), flags.Cites, "the citation id of the METHOD this applies — the source that says trial division or Miller-Rabin decides primality. The method is cited; the instance is computed")
	flags.Text(c, flags.Variance, "what holds across the two runs when they differ — the invariant the result is evidence of (\"every run reports 91 = 7 × 13; the timing line moves\"). Required when the runs differ and refused when they agree")
	c.Flags().Bool(flags.ExpectError, false, "this computation's POINT is a failing command (a path that must be absent, a tool that must be missing) — record the environment error as the result instead of refusing it")
	c.Flags().String(flags.Key, "", flags.DescKey+". A repeat under the handle with another bundle or command, --quote, --cites, --answers, --reason or --variance is refused: a handle names one computation")
	c.Flags().Var(flags.GapID().WithCheck(record.GapExists), flags.Answers, "the gap id this computation settles — REQUIRED to close a gap whose check kind is computation, which prose cannot answer")
	seat.Records(c, "computation")
	seat.Supplies(c, "basis", "the tool runs the program twice and records `reproducible` when the runs agree, `variant` when they differ")
	return c
}

// commandArgs is the command after `--`: everything there is the program's argv, and nothing may
// stand before it, where it would read as neither a flag nor the command.
func commandArgs(cmd *cobra.Command) ([]string, error) {
	args, dash := cmd.Flags().Args(), cmd.Flags().ArgsLenAtDash()
	if dash > 0 {
		return nil, feov.Errorf(feov.Validation, "%q stands before `--`, where it is neither a flag nor the command: the command goes after `--`, last, as `-- python3 count.py`", args[0])
	}
	if dash < 0 {
		return nil, feov.Errorf(feov.MissingField, "requires the command after `--`: the program and its arguments, last, as `-- python3 count.py` or `-- ./run.sh` — the tool runs exactly that, in a copy of --dir")
	}
	return args, nil
}

// computeRetry is what a retry under a computation's key states besides its bundle.
func computeRetry(cmd *cobra.Command, why string) *recordpb.Computation {
	retry := &recordpb.Computation{}
	computeFields(cmd, retry, why)
	return retry
}

// computeFields sets every field of a computation that the SEAT types.
func computeFields(cmd *cobra.Command, body *recordpb.Computation, why string) {
	body.Location = proto.String(seat.Location(cmd))
	body.Key = proto.String(seat.Str(cmd, flags.Key))
	if v := strings.TrimSpace(seat.Str(cmd, flags.Variance)); v != "" {
		body.Invariant = proto.String(v)
	}
	body.Answers = proto.String(seat.Str(cmd, flags.Answers))
	body.Cites = proto.String(seat.Str(cmd, flags.Cites))
	body.Text = proto.String(why)
}

// computeAsserted is what a computation states besides its program, and so what a retry under its key must
// state again.
var computeAsserted = []seat.Asserted{
	{Flag: flags.Quote, Field: "location"},
	{Flag: flags.Cites, Field: "cites"},
	{Flag: flags.Answers, Field: "answers"},
	{Flag: flags.Reason, Field: "text"},
	{Flag: flags.Variance, Field: "invariant"},
}

func computeRefusal(err error) error { return anchortext.Refusal(err) }

type computeResult struct {
	Label      string `json:"computation_id,omitempty"`
	SHA        string `json:"sha256"`
	Basis      string `json:"basis,omitempty"`
	Exit       int    `json:"exit"`
	Drift      string `json:"drift,omitempty"`
	Idempotent bool   `json:"idempotent,omitempty"`
	// Leftover is why the program's copies could not be removed; the computation is recorded.
	Leftover string `json:"leftover,omitempty"`
	// VoiceTells is ADVICE on the note, and the computation is already recorded by the time it renders.
	VoiceTells []string `json:"voice_tells,omitempty"`
}

func (r computeResult) Human() string {
	note := reportvoice.Note("the computation's footnote, and its heading in evidence.md", r.VoiceTells)
	if r.Idempotent {
		return "blue compute (idempotent retry — already recorded as " + compute.Short(r.SHA) + ")" + note
	}
	out := fmt.Sprintf("computation %s recorded (%s, exit %d) — anchored, program and output stored", r.Label, r.Basis, r.Exit)
	if r.Drift != "" {
		out += "\n  the runs differed: " + r.Drift + " — recorded with the invariant you stated"
	}
	if r.Leftover != "" {
		out += "\n  the tool's own defect, logged: " + r.Leftover
	}
	return out + note
}
