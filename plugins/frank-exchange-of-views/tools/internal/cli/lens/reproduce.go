package lens

import (
	"errors"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/enumhelp"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/compute"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// reproduce: red re-RUNS a computation instead of re-reading it.
//
// THIS IS THE STRONGEST AUDIT THE ENGINE HAS. Every other check ends in reading: a cited
// source is fetched and read, a claim is compared against prose, an anchor is located. All of
// them end with the auditor believing bytes someone else chose. A computation ends with red
// executing the same program and comparing the result — the one place the audit does not
// depend on trusting the audited.
//
// It is the direct answer to the smoke's R2-2, where blue asserted a test had happened and
// red could only observe that no evidence was shown. With a computation there is nothing to assert:
// the script is on disk, and either it still says what blue recorded or it does not.
//
// IT RECORDS ITS VERDICT NOW (#343). It used to record nothing, on the reasoning that a read
// leaves its result to the reader — which held for count-claims, and did not hold here.
//
// The asymmetry was the tell: the citation pair records BOTH halves (blue cites, red verifies),
// while the computation pair recorded only blue's. So the newest and strongest evidence axis was the
// LEAST audited: "red re-ran the computation and it held" existed nowhere but in red's head, and
// a computation nobody re-ran was indistinguishable from one that reproduced.
//
// TWO AXES, AND ONLY ONE OF THEM IS COMPUTABLE.
//
//   - REPRODUCED: the tool re-runs the script and compares bytes. A measurement; not claimable.
//   - SOUND: does the script actually establish the claim it is anchored to? A judgement, and
//     red must READ THE SCRIPT to make it.
//
// Reproducing is not proving. `print("7 is prime")` reproduces perfectly forever, and a computation
// that RE-RUNS CLEAN AND ESTABLISHES NOTHING is the dangerous cell — it looks maximally
// credible. The code need not be pretty or idiomatic; it has to compute the thing it claims,
// ideally readably enough that a human can see that it does.
//
// So --as is REQUIRED. A reproduce that records only the mechanical half would put the
// tool's authority behind a claim nobody checked.
//
// The page (help/reproduce.md) gives the instruction — the tool re-runs, the seat reads and judges
// — and leaves this argument here, with one more: running the script by hand and then calling this
// to write down what was seen does the same work twice.
func newReproduce() *cobra.Command {
	c := seat.Prose(seat.New("reproduce", reproduceResult{}, func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
		run, err := s.Run()
		if err != nil {
			return nil, err
		}
		// THE SEAT NAMES THE ANCHOR; THE TOOL DERIVES THE SHA. A verdict on soundness is about one
		// sentence, and the anchor is that sentence's name — the token the seat is holding while it
		// reads. The script's sha is a fact of the computation the anchor names, so asking the seat for it
		// asked it to look up what the record already joins.
		p, err := record.ComputationAt(run, strings.TrimSpace(seat.Str(cmd, flags.Anchor)))
		if err != nil {
			return nil, err
		}
		soundness := seat.Str(cmd, flags.As)
		if soundness == "" {
			return nil, feov.Errorf(feov.MissingField, "requires --as sound|unsound: re-running proves the program is DETERMINISTIC, not that it establishes the claim — `print(\"7 is prime\")` reproduces forever. This call records your verdict on the sentence at --anchor, so read the program FIRST — `show evidence --anchor %s` prints it — then call this once and say whether it computes what THAT sentence says; nothing is recorded by this call", p.Label)
		}
		snd, known := record.SoundnessOf(soundness)
		if !known {
			return nil, feov.Errorf(feov.Validation, "%q is not a soundness this record can carry", soundness)
		}
		note, err := seat.Reason(cmd)
		if err != nil {
			return nil, err
		}
		if note == "" {
			return nil, feov.Errorf(feov.MissingField, "requires --reason: say what the program ACTUALLY COMPUTES, in your words, and whether that is what the sentence at %s says. A soundness verdict with no reading behind it is the assertion this verb exists to replace — read the program FIRST (`show evidence --anchor %s` prints it), then call this once with the verdict and what the reading showed; nothing is recorded by this call", p.Label, p.Label)
		}
		body := &recordpb.Reproduce{
			ComputationId: proto.String(p.Label),
			BundleSha:     proto.String(p.SHA),
			Soundness:     &snd,
			Note:          proto.String(note),
		}
		res := reproduceResult{Anchor: p.Label, SHA: p.SHA}
		// ONE STANDING RE-RUN PER BUNDLE, JUDGED PER ACT (gblock, 2026-10-17: "Why keep reproducing it
		// when it wasn't modified?"; plans/compute-and-show.md §III.1.4). The bundle is addressed by
		// its sha256 and never changes, so a second execution of the same bytes measures the
		// environment, not the computation: where the bundle has a standing re-run this act executes
		// nothing (`ran` false) and compares against it. The COMPARISON is this act's own — its
		// recorded output, exit and out/ — because two acts on one bundle can have recorded
		// different results, and another act's verdict says nothing about this one's.
		//
		// WHERE IT EXECUTES, THE EXECUTION IS STAGED, judged, and only then stored and recorded: a
		// refused verdict leaves nothing under rerun/, and the rename comes before the append, so no
		// event ever names a directory the store lacks. A failed append after the rename leaves a
		// directory no event names, which nothing reads, and the next reproduce executes again.
		var o rerunOutcome
		var staged *compute.Staged
		if st := p.Standing; st != nil {
			o = rerunOutcome{by: st.SeatID, rerunSha: st.RerunSha}
			if st.Rerun == recordpb.Rerun_RERUN_CALLS_THE_TOOL {
				// AN EXECUTION THAT CALLED THE TOOL HAS NO OUTPUT TO COMPARE: nothing it printed is a
				// result, so this act records what that execution was.
				o.rerun = recordpb.Rerun_RERUN_CALLS_THE_TOOL
			} else {
				e, err := compute.ReadRerun(run.Dir(), p.SHA, st.RerunSha, st.Exit)
				if err != nil {
					return nil, err
				}
				if err := o.compare(run.Dir(), p, e); err != nil {
					return nil, err
				}
			}
		} else {
			e, err := compute.Rerun(run.Dir(), p.SHA, p.Argv)
			if err != nil {
				return nil, err
			}
			o = rerunOutcome{ran: true, exit: e.Exit, toolCalls: e.Calls, leftover: e.Leftover}
			if err := o.compare(run.Dir(), p, e); err != nil {
				return nil, err
			}
			if len(e.Calls) > 0 {
				// A BUNDLE THAT CALLS THE TOOL CANNOT RUN IN A CHECKER'S HANDS, and nothing it printed
				// is a result: the tool refused every call, and the program may have turned the
				// refusal into an ordinary-looking number.
				o.rerun, o.drift = recordpb.Rerun_RERUN_CALLS_THE_TOOL, ""
			}
			if staged, err = compute.StageRerun(run.Dir(), p.SHA, e); err != nil {
				return nil, err
			}
			o.rerunSha = staged.Sha
		}
		if err := o.refuseSound(p, snd); err != nil {
			if staged != nil {
				err = errors.Join(err, staged.Discard())
			}
			return nil, err
		}
		if staged != nil {
			if err := staged.Promote(); err != nil {
				return nil, err
			}
		}
		res.Reproduced, res.Got, res.Recorded, res.ToolCalls = o.rerun == recordpb.Rerun_RERUN_REPRODUCED, o.got, o.recorded, o.toolCalls
		body.Rerun, body.Ran, body.RerunSha = o.rerun.Enum(), proto.Bool(o.ran), proto.String(o.rerunSha)
		if o.ran {
			body.RerunExit = proto.Int32(int32(o.exit))
		}
		if o.rerun == recordpb.Rerun_RERUN_DIFFERS {
			body.Drift = proto.String(o.drift)
		}
		res.By, res.Ran, res.RerunSha = o.by, o.ran, o.rerunSha
		res.Drift = o.drift
		res.Rerun = recordpb.Word(body.GetRerun())
		if beforeAppend != nil {
			if err := beforeAppend(); err != nil {
				return nil, err
			}
		}
		if _, err := record.Append(s.Identity(), body); err != nil {
			return nil, err
		}
		// COPIES LEFT BEHIND ARE THE TOOL'S DEFECT, NOT THE RE-RUN'S: the execution finished before
		// the removal, so the act stands and the leftover is logged after it, best-effort.
		if o.leftover != "" {
			_, _ = record.Append(s.Identity(), &recordpb.Log{Text: proto.String(o.leftover), Type: recordpb.LogType_LOG_TYPE_DEFECT.Enum(), Source: recordpb.LogSource_LOG_SOURCE_TOOL.Enum()})
		}
		return res, nil
	}))
	c.Flags().Var(flags.ComputationAnchor().WithCheck(record.ComputationExists), flags.Anchor, "the id of the computation whose sentence you judge, from the report's `"+anchor.Token("K-…")+"` token — `show evidence` lists every computation by it. The tool executes that computation's bundle unless the bundle already has a standing re-run; either way it compares THIS computation's own output, exit and out/")
	enumhelp.Flag(c, flags.As, record.MustEnum("reproduce", "soundness"), "having READ the program: does it actually establish the sentence at --anchor?")
	seat.Require(c, flags.Anchor, flags.As)
	seat.Supplies(c, "ran", "the tool executes the bundle unless it already has a standing re-run, and records whether this act executed it")
	// The handler refuses a verdict with no reading behind it; the marker says so where the seat reads.
	return seat.SaysRequired(c, flags.Reason)
}

// beforeAppend runs after a reproduce stores its execution and before it appends its event; nil
// outside a test.
var beforeAppend func() error

// SetBeforeReproduceAppendForTest has every reproduce call f after it stores its execution and
// before it appends its event, and an error from f refuses the act there: the seam a test uses to
// fail the append after the rename. It returns the restore.
func SetBeforeReproduceAppendForTest(f func() error) (restore func()) {
	prev := beforeAppend
	beforeAppend = f
	return func() { beforeAppend = prev }
}

// rerunOutcome is this act's comparison against an execution of its bundle — its own (ran), or the
// standing re-run, where by names the seat that executed it.
type rerunOutcome struct {
	rerun         recordpb.Rerun
	recorded, got string
	drift         string
	toolCalls     []string
	by            string
	ran           bool
	exit          int
	rerunSha      string
	// leftover is why this act's execution could not remove its copy, or "".
	leftover string
}

// compare sets this act's own recorded result against e.
func (o *rerunOutcome) compare(runDir string, p record.Computation, e compute.Execution) error {
	c, err := compute.Compare(runDir, p.SHA, p.Label, p.Exit, e)
	if err != nil {
		return err
	}
	o.recorded, o.got, o.drift = c.Recorded, e.Output, c.Drift
	o.rerun = recordpb.Rerun_RERUN_DIFFERS
	if c.Matches {
		o.rerun = recordpb.Rerun_RERUN_REPRODUCED
	}
	return nil
}

// refuseSound is the one check of a `sound` verdict against THIS act's comparison, whoever executed
// the bundle: a bundle that calls the tool is never sound, and a `reproducible` computation whose own
// recorded result the re-run did not write is not sound either — whatever another act's comparison
// against the same re-run showed.
func (o rerunOutcome) refuseSound(p record.Computation, snd recordpb.Soundness) error {
	if snd != recordpb.Soundness_SOUNDNESS_SOUND {
		return nil
	}
	// The execution this verdict stands on, in the words of the refusal.
	ran := "run now"
	if o.by != "" {
		ran = "compared against the re-run " + o.by + " executed — the one standing for the bundle, so nothing was executed now —"
	}
	switch o.rerun {
	case recordpb.Rerun_RERUN_CALLS_THE_TOOL:
		calls := "calls feov-record"
		if len(o.toolCalls) > 0 {
			calls += " (" + o.toolCalls[0] + ")"
		}
		return feov.Errorf(feov.Conflict, "computation %s's bundle, %s, %s, and the tool refuses every call made while a checker runs a program — so the re-run checked nothing, and `--as sound` is not recorded. "+
			"Nothing is recorded. A program no checker can run establishes nothing a checker can check: that is `--as unsound`, with a reason that says the program reads the debate's tool rather than computing its claim", p.Label, ran, calls)
	case recordpb.Rerun_RERUN_DIFFERS:
		// A SOUND VERDICT OVER A RE-RUN THAT DID NOT REPRODUCE IS REFUSED, where the computation is
		// recorded `reproducible`. The verdict and its reason are typed before the run, so on a
		// failed run both are a judgement of output the seat has not seen: on universe-m20 a
		// stored program that could not run for any red seat took five `sound` verdicts, two of
		// them with a note saying the output matched. Refused, the seat reads what the program
		// wrote and calls again with what the run did.
		//
		// ONLY `reproducible`. A computation recorded `variant` is one whose output moves, and a
		// re-run that differs is that computation behaving as recorded.
		reproducible := recordpb.Word(recordpb.Basis_BASIS_REPRODUCIBLE)
		if p.Basis != reproducible {
			return nil
		}
		read := "Your verdict and reason were written before this output existed: read what the program wrote, then call again with what the run did. "
		if o.by != "" {
			read = "Read what that re-run differs by, above, then call again with what the run did. "
		}
		return feov.Errorf(feov.Conflict, "the re-run did NOT reproduce this computation's recorded result, so `--as sound` is not recorded. Computation %s is recorded `%s`, and its stored bundle, %s, did not write what this computation recorded.\n"+
			"  recorded: %s\n"+
			"  %s\n"+
			"Nothing is recorded. "+read+
			"A program the seat auditing it cannot run to its recorded output establishes nothing that seat can check — that is `--as unsound`, with a reason that says how the run went",
			p.Label, reproducible, ran, strings.TrimSpace(truncateOutput(o.recorded)), strings.TrimSpace(o.drift))
	}
	return nil
}

// truncateOutput keeps a mismatch legible without copying a whole run's output into the refusal.
func truncateOutput(s string) string {
	return compute.Excerpt(s, 400)
}

type reproduceResult struct {
	Anchor string `json:"anchor"`
	SHA    string `json:"sha256"`
	// Rerun is what this act's comparison found, in the record's word: reproduced | differs |
	// calls_the_tool. Where the bundle already had a standing re-run, this act executed nothing and By
	// names whose execution it compared against.
	Rerun string `json:"rerun"`
	// Reproduced is whether the execution wrote THIS computation's recorded output, exit and out/ —
	// the record's own word for it, on every result, true or false. It was `matches`: a lens on
	// universe-m20 filtered the envelope for `ok`, `reproduced` and `sound`, read `ok: true` beside
	// the nulls of two keys nothing carried, and took a failed re-run for a recorded success. The
	// word a seat reaches for is the field, and a `sound` verdict over a failed re-run of a
	// reproducible computation never reaches an `ok: true` at all.
	Reproduced bool   `json:"reproduced"`
	Got        string `json:"got"`
	Recorded   string `json:"recorded"`
	// Drift is what differs between this computation's recorded result and the execution, where it
	// differed.
	Drift string `json:"drift,omitempty"`
	// Ran is whether this act executed the bundle; RerunSha names the execution it compared against.
	Ran      bool   `json:"ran"`
	RerunSha string `json:"rerun_sha"`
	// By is the seat whose execution this comparison read, where this act executed nothing.
	By string `json:"rerun_by,omitempty"`
	// ToolCalls is every feov-record call the program made under the re-run, each refused.
	ToolCalls []string `json:"tool_calls,omitempty"`
}

func (r reproduceResult) Human() string {
	short := compute.Short(r.SHA)
	if r.By != "" {
		showed := map[string]string{
			recordpb.Word(recordpb.Rerun_RERUN_REPRODUCED):     "this computation's recorded output REPRODUCED",
			recordpb.Word(recordpb.Rerun_RERUN_DIFFERS):        "this computation's recorded output did NOT reproduce (" + r.Drift + ")",
			recordpb.Word(recordpb.Rerun_RERUN_CALLS_THE_TOOL): "that re-run CALLED THE TOOL, so it checked nothing",
		}[r.Rerun]
		return "computation " + r.Anchor + ": compared against its bundle " + short + "'s re-run by " + r.By + ", so nothing was executed — " + showed + ". Your verdict on this sentence is recorded"
	}
	switch r.Rerun {
	case recordpb.Word(recordpb.Rerun_RERUN_CALLS_THE_TOOL):
		return "computation " + r.Anchor + " CANNOT BE CHECKED: its bundle called feov-record (" + strings.Join(r.ToolCalls, "; ") + "), which refuses every call made while a checker runs a program, so nothing it printed is a result.\n  now: " + r.Got +
			"\n  Recorded as calls_the_tool, with your verdict: the program reads the debate's tool rather than computing its claim"
	case recordpb.Word(recordpb.Rerun_RERUN_REPRODUCED):
		return "computation " + r.Anchor + " (bundle " + short + ") REPRODUCES — executing it gives this computation's recorded output byte for byte. Your verdict on this sentence is recorded"
	}
	return "computation " + r.Anchor + " (bundle " + short + ") DOES NOT REPRODUCE.\n  recorded: " + r.Recorded + "\n  now:      " + r.Got +
		"\n  Either the computation depends on something that moved, or the recorded output is not what the program produces. Both are findings; neither is recorded for you."
}
