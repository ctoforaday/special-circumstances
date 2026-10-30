package blue

import (
	"fmt"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"strings"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchortext"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/proof"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/reportvoice"
)

// prove: settle a claim by COMPUTING it, and leave the computation as the evidence.
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
// prove the INSTANCE by running it. Several methods on one claim is several anchors on one
// sentence, which is exactly how a mathematician argues.
//
// THE BASIS IS DERIVED, never claimed: the tool runs the script twice. Identical output is
// `reproducible` — the auditor can reach the same bytes independently. Moving output is
// `observed`, still evidence but of a system in motion (a live call measures what an API
// actually does, which beats its documentation), and the report must not read the two alike.
func newProve() *cobra.Command {
	c := seat.Prose(seat.New("prove", proveResult{}, func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
		run, err := s.Run()
		if err != nil {
			return nil, err
		}
		location, script := seat.Location(cmd), seat.Str(cmd, flags.Script)
		if strings.TrimSpace(location) == "" {
			return nil, feov.Errorf(feov.MissingField, "requires --quote: the EXACT sentence in the report (as `show report` serves it) this computation backs — a proof anchored to nothing is a script nobody can connect to a claim")
		}
		if strings.TrimSpace(script) == "" {
			return nil, feov.Errorf(feov.MissingField, "requires --script: the path of the program that settles it — absolute, or relative to the run directory")
		}
		if err := record.RefuseRecordFile(run, proof.ScriptPath(run.Dir(), script), "--script", "a program"); err != nil {
			return nil, err
		}

		// CRASH-RETRY AND KEY COLLISION LOOK IDENTICAL AT THE KEY, AND ARE OPPOSITE FACTS.
		//
		// A retry is the SAME program arriving twice: returning the recorded sha is right, and
		// re-running it would splice a second anchor. A collision is a DIFFERENT program under a
		// key an earlier dispatch of this seat already burned — and this used to take the same
		// path, answering "already recorded as <sha>" and recording NOTHING. A seat that believed
		// the message shipped a report claiming a proof it does not hold, and a resumed dispatch
		// cannot see which keys its earlier self used. Measured in
		// research/2026-09-02_quadratic-formula (blue-respond, --key P1).
		//
		// THE NOTE IS PRINTED IN THE REPORT — as this proof's [^PN] footnote and its evidence.md
		// heading (report/proofs.go) — so it meets the same voice advisory as edit's --new. ADVISORY,
		// COMPUTED BEFORE ANY RETURN so an idempotent retry carries it too; it refuses nothing.
		why, err := seat.Reason(cmd)
		if err != nil {
			return nil, err
		}
		tells := spanVoiceTells(why)

		// The script's own sha settles it, and costs a file read rather than an execution.
		key := seat.Str(cmd, flags.Key)
		if prior, err := record.ExistingProofByKey(run, s.SeatID, key); err != nil {
			return nil, err
		} else if prior != nil {
			now, serr := proof.ScriptSha(run.Dir(), script)
			if serr != nil {
				return nil, serr
			}
			if now != prior.GetProofSha() {
				return nil, feov.Errorf(feov.Conflict, "--key %q already names proof %s on this run, and %s hashes to %s — "+
					"a key is one program's name, not a slot to reuse. This is not a retry: recording nothing here would "+
					"leave you citing a proof of something else. Give this program its own --key",
					key, prior.GetProofSha()[:12], script, now[:12])
			}
			// THE SAME PROGRAM IS NOT YET THE SAME ACT: a retry states the sentence it backs, the
			// method it cites, the gap it answers and its note again, or it is a second proof reusing
			// the name.
			retry := &recordpb.Proof{}
			proofFields(cmd, retry, why)
			if err := seat.RepeatsKeyed(key, "proof", prior.GetProofId(), prior, retry, proveAsserted...); err != nil {
				return nil, err
			}
			// A proof on the record without its anchor — one migrate brought forward from an archive
			// whose run crashed between the two — is finished by the retry, at the location it states.
			if err := seat.PlaceOwed(s, prior.GetProofId(), key); err != nil {
				return nil, err
			}
			return proveResult{SHA: prior.GetProofSha(), Idempotent: true, VoiceTells: tells}, nil
		}

		res, err := proof.Run(run.Dir(), script)
		if err != nil {
			// A proof that will not run is not evidence, and the failure is a capability
			// signal — the same treatment `blue cite` gives an unreachable source.
			msg := err.Error()
			if _, ferr := record.Append(s.Identity(), &recordpb.Log{Text: proto.String(msg), Type: recordpb.LogType_LOG_TYPE_DEFECT.Enum(), Source: recordpb.LogSource_LOG_SOURCE_TOOL.Enum()}); ferr != nil {
				return nil, ferr
			}
			return nil, record.ToolLogged(err)
		}

		// THE ENVIRONMENT FAILING IS NOT THE COMPUTATION ANSWERING, and exit 0 does not tell the
		// two apart. `proof.Run` works in the RUN directory; a script authored against the repo
		// root prints "No such file or directory", carries on, and ends clean — which is how a
		// report came to cite an enumeration that enumerated nothing as re-runnable evidence.
		// Refused rather than graded, because the artifact is otherwise indistinguishable from a
		// real negative result, and the seat is one edit away from a correct one.
		expectError, err := cmd.Flags().GetBool(flags.ExpectError)
		if err != nil {
			return nil, err
		}
		// --expect-error IS A CLAIM ABOUT THE RUN, and a run that failed nothing makes it false: the
		// flag would be accepted and used for nothing.
		if expectError && res.Failed == "" {
			return nil, feov.Errorf(feov.Validation, "--expect-error says this proof's point is a failing command, but %s ran to exit %d with no environment error in its output — drop --expect-error, or fix the script so it captures the failure it is meant to prove", script, res.Exit)
		}
		if res.Failed != "" && !expectError {
			return nil, feov.Errorf(feov.Validation, "%s ran to exit %d, but its output carries an ENVIRONMENT error, not a result:\n    %s\n"+
				"A proof runs with the RUN DIRECTORY as its working directory (%s), not the repository root — a script that reaches for repo paths must resolve them itself. "+
				"Fix the script and re-run; if capturing this failure IS the point (proving a path is absent, a command missing), say so with --expect-error and it is recorded as the result it is",
				script, res.Exit, res.Failed, run.Dir())
		}

		// THE ANCHOR EVENT BELOW IS THE MARKER: it carries the quote, and reportproj.Render re-places
		// the proof anchor at replay. No file is spliced; a --key retry is idempotent (handled
		// above). Mint the id; the placement is validated against the report the Proof and its Anchor
		// land on — a mis-quote or in-fence quote is refused and nothing recorded.
		label := record.NewID("proof")

		// `output` does not survive onto the event, and that is not a silent drop: it stays in the
		// proof cache addressed by proof_sha — content is not a fact about the debate, and the
		// census records that reasoning against the key. `location` DOES survive now (#709): under
		// report-as-record there is no report.md, so the projection re-places this proof marker by
		// re-locating `location` — the anchoring site must be a fact the record holds, not one
		// recoverable only from the spliced marker report/proofs.go never read back off the event.
		body := &recordpb.Proof{
			ProofId:    proto.String(label),
			ProofSha:   proto.String(res.SHA),
			ProofBasis: proto.String(res.Basis),
			Exit:       proto.Int32(int32(res.Exit)),
		}
		if res.Drift != "" {
			body.Drift = proto.String(res.Drift)
		}
		proofFields(cmd, body, why)
		if _, err := seat.AppendPlaced(s, body, label, location, proveRefusal); err != nil {
			return nil, err
		}
		return proveResult{Label: label, SHA: res.SHA, Basis: res.Basis, Exit: res.Exit, Drift: res.Drift, VoiceTells: tells}, nil
	}))

	// DescQuoteAlone: this verb's --reason is the proof note the report prints, no place for a section name.
	flags.Text(c, flags.Quote, flags.DescQuoteAlone)
	c.Flags().String(flags.Script, "", "path of the program that settles it (.py, .js, .mjs, .sh or .go). A relative path is read from the run directory, so a script drafted in the run's scratch/ is passed as scratch/<name>; an absolute path is read where it lies. The tool keeps its own copy with the proof, so the proof does not depend on the draft")
	seat.Require(c, flags.Quote, flags.Script)
	c.Flags().Var(flags.CitationAnchor().WithCheck(record.CitationExists(record.ProveCitesAdvice)), flags.Cites, "the citation id of the METHOD this applies — the source that says trial division or Miller-Rabin decides primality. The method is cited; the instance is computed")
	c.Flags().Bool(flags.ExpectError, false, "this proof's POINT is a failing command (a path that must be absent, a tool that must be missing) — record the environment error as the result instead of refusing it")
	c.Flags().String(flags.Key, "", flags.DescKey+". A repeat under the handle with another program, --quote, --cites, --answers or --reason is refused: a handle names one proof")
	c.Flags().Var(flags.GapID().WithCheck(record.GapExists), flags.Answers, "the gap id this computation settles — REQUIRED to close a gap whose check kind is computation, which prose cannot answer")
	seat.Records(c, "proof")
	return c
}

// proofFields sets every field of a proof that the SEAT types. `script` is the flag's own value:
// proof.Run records the path it was handed.
func proofFields(cmd *cobra.Command, body *recordpb.Proof, why string) {
	body.Location = proto.String(seat.Location(cmd))
	body.Script = proto.String(seat.Str(cmd, flags.Script))
	body.ProofKey = proto.String(seat.Str(cmd, flags.Key))
	body.Answers = proto.String(seat.Str(cmd, flags.Answers))
	body.Cites = proto.String(seat.Str(cmd, flags.Cites))
	body.Text = proto.String(why)
}

// proveAsserted is what a proof states besides its program, and so what a retry under its key must
// state again.
var proveAsserted = []seat.Asserted{
	{Flag: flags.Quote, Field: "location"},
	{Flag: flags.Cites, Field: "cites"},
	{Flag: flags.Answers, Field: "answers"},
	{Flag: flags.Reason, Field: "text"},
}

func proveRefusal(err error) error { return anchortext.Refusal(err) }

type proveResult struct {
	Label      string `json:"proof_id,omitempty"`
	SHA        string `json:"sha256"`
	Basis      string `json:"proof_basis,omitempty"`
	Exit       int    `json:"exit"`
	Drift      string `json:"drift,omitempty"`
	Idempotent bool   `json:"idempotent,omitempty"`
	// VoiceTells is ADVICE on the note, and the proof is already recorded by the time it renders.
	VoiceTells []string `json:"voice_tells,omitempty"`
}

func (r proveResult) Human() string {
	note := reportvoice.Note("the proof's footnote, and its heading in evidence.md", r.VoiceTells)
	if r.Idempotent {
		return seat.Repeated("the proof of program "+r.SHA[:12]) + note
	}
	out := fmt.Sprintf("proof %s recorded (%s, exit %d) — anchored, script and output cached", r.Label, r.Basis, r.Exit)
	if r.Drift != "" {
		out += "\n  NOT reproducible: " + r.Drift + " — recorded as a measurement, not a proof"
	}
	return out + note
}
