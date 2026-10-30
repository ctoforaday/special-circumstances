package lens

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"strings"

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

// reproduce: red re-RUNS a proof instead of re-reading it.
//
// THIS IS THE STRONGEST AUDIT THE ENGINE HAS. Every other check ends in reading: a cited
// source is fetched and read, a claim is compared against prose, an anchor is located. All of
// them end with the auditor believing bytes someone else chose. A proof ends with red
// executing the same program and comparing the result — the one place the audit does not
// depend on trusting the audited.
//
// It is the direct answer to the smoke's R2-2, where blue asserted a test had happened and
// red could only observe that no evidence was shown. With a proof there is nothing to assert:
// the script is on disk, and either it still says what blue recorded or it does not.
//
// IT RECORDS ITS VERDICT NOW (#343). It used to record nothing, on the reasoning that a read
// leaves its result to the reader — which held for count-claims, and did not hold here.
//
// The asymmetry was the tell: the citation pair records BOTH halves (blue cites, red verifies),
// while the proof pair recorded only blue's. So the newest and strongest evidence axis was the
// LEAST audited: "red re-ran the computation and it held" existed nowhere but in red's head, and
// a proof nobody re-ran was indistinguishable from one that reproduced.
//
// TWO AXES, AND ONLY ONE OF THEM IS COMPUTABLE.
//
//   - REPRODUCED: the tool re-runs the script and compares bytes. A measurement; not claimable.
//   - SOUND: does the script actually establish the claim it is anchored to? A judgement, and
//     red must READ THE SCRIPT to make it.
//
// Reproducing is not proving. `print("7 is prime")` reproduces perfectly forever, and a proof
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
		// reads. The script's sha is a fact of the proof the anchor names, so asking the seat for it
		// asked it to look up what the record already joins.
		p, err := record.ProofAt(run, strings.TrimSpace(seat.Str(cmd, flags.Anchor)))
		if err != nil {
			return nil, err
		}
		soundness := seat.Str(cmd, flags.As)
		if soundness == "" {
			return nil, feov.Errorf(feov.MissingField, "requires --as sound|unsound: re-running proves the script is DETERMINISTIC, not that it establishes the claim — `print(\"7 is prime\")` reproduces forever. This call records your verdict on the sentence at --anchor, so read the script FIRST — `show evidence --anchor %s` prints it — then call this once and say whether it computes what THAT sentence says; nothing is recorded by this call", p.Label)
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
			return nil, feov.Errorf(feov.MissingField, "requires --reason: say what the script ACTUALLY COMPUTES, in your words, and whether that is what the sentence at %s says. A soundness verdict with no reading behind it is the assertion this verb exists to replace — read the script FIRST (`show evidence --anchor %s` prints it), then call this once with the verdict and what the reading showed; nothing is recorded by this call", p.Label, p.Label)
		}
		body := &recordpb.Reproduce{
			ProofId:   proto.String(p.Label),
			ProofSha:  proto.String(p.SHA),
			Soundness: &snd,
			Note:      proto.String(note),
		}
		res := reproduceResult{Anchor: p.Label, SHA: p.SHA}
		// A SCRIPT IS RE-RUN ONCE, AND THAT RE-RUN STANDS FOR EVERY PROOF ON IT WHATEVER IT SHOWED
		// (gblock, 2026-10-17: "Why keep reproducing it when it wasn't modified?"). The script is
		// addressed by its sha256 and never changes, so a second run of the same bytes measures the
		// environment, not the proof — and a re-run that differed or called the tool was re-run on
		// every `reproduce`. Where the script has a re-run, this act judges its own sentence on that
		// re-run and runs nothing; only a changed script, a new sha, is a new re-run.
		var o rerunOutcome
		if p.Rerun != nil {
			o = rerunOutcome{rerun: p.Rerun.Rerun, recorded: p.Rerun.Recorded, got: p.Rerun.Observed, by: p.Rerun.SeatID}
		} else {
			rr, err := compute.Reproduce(run.Dir(), p.SHA)
			if err != nil {
				return nil, err
			}
			o = rerunOutcome{rerun: recordpb.Rerun_RERUN_DIFFERS, recorded: rr.Want, got: rr.Got, toolCalls: rr.ToolCalls}
			switch {
			case len(rr.ToolCalls) > 0:
				// A SCRIPT THAT CALLS THE TOOL CANNOT RUN IN A CHECKER'S HANDS, and nothing it printed is
				// a result: the tool refused every call, and the script may have turned the refusal
				// into an ordinary-looking number.
				o.rerun = recordpb.Rerun_RERUN_CALLS_THE_TOOL
			case rr.Matches:
				o.rerun = recordpb.Rerun_RERUN_REPRODUCED
			}
		}
		if err := o.refuseSound(p, snd); err != nil {
			return nil, err
		}
		res.Reproduced, res.Got, res.Recorded, res.ToolCalls = o.rerun == recordpb.Rerun_RERUN_REPRODUCED, o.got, o.recorded, o.toolCalls
		if o.by != "" {
			body.Rerun = recordpb.Rerun_RERUN_NOT_RUN.Enum()
			res.By, res.Standing = o.by, recordpb.Word(o.rerun)
		} else {
			body.Rerun = o.rerun.Enum()
			switch o.rerun {
			case recordpb.Rerun_RERUN_CALLS_THE_TOOL:
				body.ObservedOutput = proto.String(truncateOutput(o.got))
			case recordpb.Rerun_RERUN_DIFFERS:
				// Only on a MISMATCH, and truncated: the recorded output already lives in the
				// proof artifact, and copying it wholesale into the event would put a second
				// copy of the same bytes on the record. What the reader needs is what CHANGED.
				body.RecordedOutput = proto.String(truncateOutput(o.recorded))
				body.ObservedOutput = proto.String(truncateOutput(o.got))
			}
		}
		res.Rerun = recordpb.Word(body.GetRerun())
		if _, err := record.Append(s.Identity(), body); err != nil {
			return nil, err
		}
		return res, nil
	}))
	c.Flags().Var(flags.ProofAnchor().WithCheck(record.ProofExists), flags.Anchor, "the id of the proof whose sentence you judge, from the report's `"+anchor.Token("P-…")+"` token — `show evidence` lists every proof by it. The tool re-runs that proof's script unless the script already has a re-run, which stands whatever it showed")
	enumhelp.Flag(c, flags.As, record.MustEnum("reproduce", "soundness"), "having READ the script: does it actually establish the sentence at --anchor?")
	seat.Require(c, flags.Anchor, flags.As)
	// The handler refuses a verdict with no reading behind it; the marker says so where the seat reads.
	return seat.SaysRequired(c, flags.Reason)
}

// rerunOutcome is what a script's re-run showed — this act's own, or the standing one it judges on,
// where by names the seat whose re-run that is.
type rerunOutcome struct {
	rerun         recordpb.Rerun
	recorded, got string
	toolCalls     []string
	by            string
}

// refuseSound is the one check of a `sound` verdict against what the re-run showed, whoever ran it:
// a script that calls the tool is never sound, and a `reproducible` proof whose re-run did not
// reproduce is not sound either.
func (o rerunOutcome) refuseSound(p record.Proof, snd recordpb.Soundness) error {
	if snd != recordpb.Soundness_SOUNDNESS_SOUND {
		return nil
	}
	// The run this verdict stands on, in the words of the refusal.
	ran := "run now"
	if o.by != "" {
		ran = "as " + o.by + " re-ran it — the re-run that stands for the script, so nothing was re-run now —"
	}
	switch o.rerun {
	case recordpb.Rerun_RERUN_CALLS_THE_TOOL:
		calls := "calls feov-record"
		if len(o.toolCalls) > 0 {
			calls += " (" + o.toolCalls[0] + ")"
		}
		return feov.Errorf(feov.Conflict, "proof %s's script, %s, %s, and the tool refuses every call made while a checker runs a script — so the re-run checked nothing, and `--as sound` is not recorded. "+
			"Nothing is recorded. A script no checker can run establishes nothing a checker can check: that is `--as unsound`, with a reason that says the script reads the debate's tool rather than computing its claim", p.Label, ran, calls)
	case recordpb.Rerun_RERUN_DIFFERS:
		// A SOUND VERDICT OVER A RE-RUN THAT DID NOT REPRODUCE IS REFUSED, where the proof is
		// recorded `reproducible`. The verdict and its reason are typed before the run, so on a
		// failed run both are a judgement of output the seat has not seen: on universe-m20 a
		// stored script that could not run for any red seat took five `sound` verdicts, two of
		// them with a note saying the output matched. Refused, the seat reads what the script
		// printed and calls again with what the run did.
		//
		// ONLY `reproducible`. A proof graded `observed` or `not_portable` is recorded as one
		// whose output moves, and a re-run that differs is that proof behaving as recorded.
		if p.Basis != compute.Reproducible {
			return nil
		}
		read := "Your verdict and reason were written before this output existed: read what the script printed, then call again with what the run did. "
		if o.by != "" {
			read = "Read what that re-run printed, above, then call again with what the run did. "
		}
		return feov.Errorf(feov.Conflict, "the re-run did NOT reproduce, so `--as sound` is not recorded. Proof %s is recorded `%s`, and its stored script, %s, did not print the recorded output.\n"+
			"  recorded: %s\n  re-run:   %s\n"+
			"Nothing is recorded. "+read+
			"A script the seat auditing it cannot run to its recorded output establishes nothing that seat can check — that is `--as unsound`, with a reason that says how the run went",
			p.Label, compute.Reproducible, ran, strings.TrimSpace(truncateOutput(o.recorded)), strings.TrimSpace(truncateOutput(o.got)))
	}
	return nil
}

// truncateOutput keeps a mismatch legible without copying a whole run's output onto the record.
func truncateOutput(s string) string {
	const max = 400
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}

type reproduceResult struct {
	Anchor string `json:"anchor"`
	SHA    string `json:"sha256"`
	// Rerun is what this act did with the script, in the record's word: reproduced | differs |
	// calls_the_tool, or not_run where the script already had a re-run and this act judged its
	// sentence on that re-run.
	Rerun string `json:"rerun"`
	// Reproduced is whether the script's standing re-run printed the recorded output — the record's
	// own word for it, on every result, true or false. It was `matches`: a lens on universe-m20
	// filtered the envelope for `ok`, `reproduced` and `sound`, read `ok: true` beside the nulls of
	// two keys nothing carried, and took a failed re-run for a recorded success. The word a seat
	// reaches for is the field, and a `sound` verdict over a failed re-run of a reproducible proof
	// never reaches an `ok: true` at all.
	Reproduced bool   `json:"reproduced"`
	Got        string `json:"got"`
	Recorded   string `json:"recorded"`
	// By is the seat whose re-run this judgement stands on, and Standing what that re-run showed
	// (reproduced | differs | calls_the_tool), where this act ran nothing.
	By       string `json:"rerun_by,omitempty"`
	Standing string `json:"standing_rerun,omitempty"`
	// ToolCalls is every feov-record call the script made under the re-run, each refused.
	ToolCalls []string `json:"tool_calls,omitempty"`
}

func (r reproduceResult) Human() string {
	short := r.SHA
	if len(short) > 12 {
		short = short[:12]
	}
	switch r.Rerun {
	case recordpb.Word(recordpb.Rerun_RERUN_NOT_RUN):
		showed := map[string]string{
			recordpb.Word(recordpb.Rerun_RERUN_REPRODUCED):     "REPRODUCED",
			recordpb.Word(recordpb.Rerun_RERUN_DIFFERS):        "did NOT reproduce",
			recordpb.Word(recordpb.Rerun_RERUN_CALLS_THE_TOOL): "CALLED THE TOOL, so it checked nothing",
		}[r.Standing]
		return "proof " + r.Anchor + ": its script " + short + " was already re-run by " + r.By + ", and " + showed + " — that re-run stands for every proof on the script whatever it showed, so nothing was re-run. Your verdict on this sentence is recorded"
	case recordpb.Word(recordpb.Rerun_RERUN_CALLS_THE_TOOL):
		return "proof " + r.Anchor + " CANNOT BE CHECKED: its script called feov-record (" + strings.Join(r.ToolCalls, "; ") + "), which refuses every call made while a checker runs a script, so nothing it printed is a result.\n  now: " + r.Got +
			"\n  Recorded as calls_the_tool, with your verdict: the script reads the debate's tool rather than computing its claim"
	case recordpb.Word(recordpb.Rerun_RERUN_REPRODUCED):
		return "proof " + r.Anchor + " (script " + short + ") REPRODUCES — re-running it gives the recorded output byte for byte, for every proof on that script. Your verdict on this sentence is recorded"
	}
	return "proof " + r.Anchor + " (script " + short + ") DOES NOT REPRODUCE.\n  recorded: " + r.Recorded + "\n  now:      " + r.Got +
		"\n  Either the computation depends on something that moved, or the recorded output is not what the script produces. Both are findings; neither is recorded for you."
}
