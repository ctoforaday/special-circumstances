package lens

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"strings"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/enumhelp"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/proof"
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
// leaves its result to the reader — which held for near-match and count-claims, and did not hold
// here.
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
func newReproduce() *cobra.Command {
	c := seat.Prose(seat.New("reproduce", func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
		run, err := s.Run()
		if err != nil {
			return nil, err
		}
		sha := seat.Str(cmd, flags.ID)
		if sha == "" {
			// THE WAY THROUGH, NAMED. This message used to say the sha was "listed in the report
			// beside the sentence it backs". It is not: the report carries an opaque
			// `<!--proof:p-…-->` anchor, and the sha lives on the record. A seat reading the
			// document had the token this verb does not take and no path to the one it does.
			return nil, feov.Errorf(feov.MissingField, "requires --id: the sha256 of the proof to re-run. Reading the report and holding a `%s` anchor, resolve it with `show evidence` — every proof is listed there with its anchor, its sha256, and whether anyone has re-run it yet, and `show evidence --anchor <that id>` prints its script", anchor.Token("P-…"))
		}
		ok, got, want, err := proof.Reproduce(run.Dir(), sha)
		if err != nil {
			return nil, err
		}
		soundness := seat.Str(cmd, flags.As)
		if soundness == "" {
			return nil, feov.Errorf(feov.MissingField, "requires --as sound|unsound: re-running proves the script is DETERMINISTIC, not that it establishes the claim — `print(\"7 is prime\")` reproduces forever. This call re-runs the proof and records your verdict in one act, so read the script FIRST — `show evidence --anchor <the proof's anchor id>` prints it — then call this once and say whether it computes what it is anchored to; nothing is recorded by this call")
		}
		snd, known := record.SoundnessOf(soundness)
		if !known {
			return nil, feov.Errorf(feov.Validation, "%q is not a soundness this record can carry", soundness)
		}
		// A SOUND VERDICT OVER A RE-RUN THAT DID NOT REPRODUCE IS REFUSED, where the proof is recorded
		// `reproducible`. The verdict and its reason are typed before the run, so on a failed run
		// both are a judgement of output the seat has not seen: on universe-m20 a stored script
		// that could not run for any red seat took five `sound` verdicts, two of them with a note
		// saying the output matched. Refused, the seat reads what the script printed and calls
		// again with what the run did.
		//
		// ONLY `reproducible`. A proof graded `observed` or `not_portable` is recorded as one whose
		// output moves, and a re-run that differs is that proof behaving as recorded.
		if !ok && snd == recordpb.Soundness_SOUNDNESS_SOUND {
			proofs, err := record.RecordedProofs(run)
			if err != nil {
				return nil, err
			}
			for _, p := range proofs {
				if p.SHA == sha && p.Basis == proof.Reproducible {
					return nil, feov.Errorf(feov.Conflict, "the re-run did NOT reproduce, so `--as sound` is not recorded. Proof %s is recorded `%s`, and its stored script, run now, did not print the recorded output.\n"+
						"  recorded: %s\n  now:      %s\n"+
						"Nothing is recorded. Your verdict and reason were written before this output existed: read what the script printed, then call again with what the run did. "+
						"A script the seat auditing it cannot run to its recorded output establishes nothing that seat can check — that is `--as unsound`, with a reason that says how the run went",
						sha[:12], proof.Reproducible, strings.TrimSpace(truncateOutput(want)), strings.TrimSpace(truncateOutput(got)))
				}
			}
		}
		body := &recordpb.Reproduce{
			ProofSha:   proto.String(sha),
			Reproduced: proto.Bool(ok),
			Soundness:  &snd,
		}
		if !ok {
			// Only on a MISMATCH, and truncated: the recorded output already lives in the
			// proof artifact, and copying it wholesale into the event would put a second
			// copy of the same bytes on the record. What the reader needs is what CHANGED.
			body.RecordedOutput = proto.String(truncateOutput(want))
			body.ObservedOutput = proto.String(truncateOutput(got))
		}
		note, err := seat.Reason(cmd)
		if err != nil {
			return nil, err
		}
		body.Note = proto.String(note)
		if note == "" {
			return nil, feov.Errorf(feov.MissingField, "requires --reason: say what the script ACTUALLY COMPUTES, in your words. A soundness verdict with no reading behind it is the assertion this verb exists to replace — this call re-runs the proof and records your verdict in one act, so read the script FIRST — `show evidence --anchor <the proof's anchor id>` prints it — then call this once with the verdict and what the reading showed; nothing is recorded by this call")
		}
		if _, err := record.Append(s.Identity(), body); err != nil {
			return nil, err
		}
		return reproduceResult{SHA: sha, Reproduced: ok, Got: got, Recorded: want}, nil
	}))
	// A SHA, SHAPED. Registered as a plain string the page printed `string` and a gap id typed here was
	// refused by the proof lookup rather than by its shape — the refusal named no kind.
	c.Flags().Var(flags.SHA(), flags.ID, "the sha256 of the recorded proof to re-run")
	enumhelp.Flag(c, flags.As, record.MustEnum("reproduce", "soundness"), "having READ the script: does it actually establish the claim it is anchored to?")
	seat.Require(c, flags.ID, flags.As)
	// The handler refuses a verdict with no reading behind it; the marker says so where the seat reads.
	return seat.SaysRequired(c, flags.Reason)
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
	SHA string `json:"sha256"`
	// Reproduced is whether the re-run printed the recorded output — the record's own word for it
	// (Reproduce.reproduced, and `reproduced` on the proof's row in `show evidence`), on every
	// result, true or false. It was `matches`: a lens on universe-m20 filtered the envelope for
	// `ok`, `reproduced` and `sound`, read `ok: true` beside the nulls of two keys nothing
	// carried, and took a failed re-run for a recorded success. The word a seat reaches for is the
	// field now, the help page says where it sits, and a `sound` verdict over a failed re-run of a
	// reproducible proof never reaches an `ok: true` at all.
	Reproduced bool   `json:"reproduced"`
	Got        string `json:"got"`
	Recorded   string `json:"recorded"`
}

func (r reproduceResult) Human() string {
	if r.Reproduced {
		return "proof " + r.SHA[:12] + " REPRODUCES — re-running it gives the recorded output byte for byte"
	}
	return "proof " + r.SHA[:12] + " DOES NOT REPRODUCE.\n  recorded: " + r.Recorded + "\n  now:      " + r.Got +
		"\n  Either the computation depends on something that moved, or the recorded output is not what the script produces. Both are findings; neither is recorded for you."
}
