package blue

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchortext"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/proof"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/reportproj"
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
// refused it — "no evidence shown" — and the round was spent. Three lines of trial division
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
	c := seat.Prose(seat.New("prove", func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
		run, err := s.Run()
		if err != nil {
			return nil, err
		}
		location, script := seat.Str(cmd, flags.Quote), seat.Str(cmd, flags.Script)
		if strings.TrimSpace(location) == "" {
			return nil, fmt.Errorf("blue prove requires --quote: the EXACT sentence in blue/report.md this computation backs — a proof anchored to nothing is a script nobody can connect to a claim")
		}
		if strings.TrimSpace(script) == "" {
			return nil, fmt.Errorf("blue prove requires --script: the path (under the run directory) of the program that settles it")
		}

		// Crash-retry: a committed proof for this key is already recorded.
		if prior, err := record.ExistingProofByKey(run, s.SeatID, seat.Str(cmd, flags.Key)); err != nil {
			return nil, err
		} else if prior != "" {
			return proveResult{SHA: prior, Idempotent: true}, nil
		}

		res, err := proof.Run(run.Dir(), script)
		if err != nil {
			// A proof that will not run is not evidence, and the failure is a capability
			// signal — the same treatment `blue cite` gives an unreachable source.
			msg := err.Error()
			if _, ferr := record.Append(s.Identity(), &recordpb.Friction{Text: proto.String(msg)}); ferr != nil {
				return nil, ferr
			}
			return nil, err
		}

		// THE ENVIRONMENT FAILING IS NOT THE COMPUTATION ANSWERING, and exit 0 does not tell the
		// two apart. `proof.Run` works in the RUN directory; a script authored against the repo
		// root prints "No such file or directory", carries on, and ends clean — which is how a
		// report came to cite an enumeration that enumerated nothing as re-runnable evidence.
		// Refused rather than graded, because the artifact is otherwise indistinguishable from a
		// real negative result, and the seat is one edit away from a correct one.
		if res.Failed != "" && !seat.Given(cmd, flags.ExpectError) {
			return nil, fmt.Errorf("blue prove: %s ran to exit %d, but its output carries an ENVIRONMENT error, not a result:\n    %s\n"+
				"A proof runs with the RUN DIRECTORY as its working directory (%s), not the repository root — a script that reaches for repo paths must resolve them itself. "+
				"Fix the script and re-run; if capturing this failure IS the point (proving a path is absent, a command missing), say so with --expect-error and it is recorded as the result it is",
				script, res.Exit, res.Failed, run.Dir())
		}

		// THE PROOF EVENT IS THE ANCHOR: it carries the quote in Location, and reportproj.Render
		// re-places the <!--proof:p-…--> marker at replay. No file is spliced, so there is no
		// torn-splice window; a --key retry is idempotent (handled above). Mint the id and VALIDATE
		// the placement against the current render — a mis-quote or in-fence quote is refused now.
		label := record.NewProofID()
		marker := "<!--proof:" + label + "-->"
		current, err := reportproj.RenderFromRecord(run)
		if err != nil {
			return nil, err
		}
		if _, aerr := anchortext.InsertAnchor([]byte(current), location, marker); aerr != nil {
			return nil, aerr
		}

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
			Script:     proto.String(res.Script),
			Exit:       proto.Int32(int32(res.Exit)),
			ProofKey:   proto.String(seat.Str(cmd, flags.Key)),
			Answers:    proto.String(seat.Str(cmd, flags.Answers)),
			Cites:      proto.String(seat.Str(cmd, flags.Cites)),
			Location:   proto.String(location),
		}
		if res.Drift != "" {
			body.Drift = proto.String(res.Drift)
		}
		why, err := seat.Reason(cmd)
		if err != nil {
			return nil, err
		}
		body.Text = proto.String(why)
		if _, err := record.Append(s.Identity(), body); err != nil {
			return nil, err
		}
		return proveResult{Label: label, SHA: res.SHA, Basis: res.Basis, Exit: res.Exit, Drift: res.Drift}, nil
	}))

	c.Flags().String(flags.Quote, "", flags.DescQuote+". The proof anchor is spliced there")
	c.Flags().String(flags.Script, "", "REQUIRED — path under the run directory of the program that settles it (.py, .js, .mjs, .sh or .go)")
	c.Flags().Var(flags.CitationAnchor().WithCheck(record.CitationExists), flags.Cites, "the citation label of the METHOD this applies — the source that says trial division or Miller-Rabin decides primality. The method is cited; the instance is computed")
	c.Flags().Bool(flags.ExpectError, false, "this proof's POINT is a failing command (a path that must be absent, a tool that must be missing) — record the environment error as the result instead of refusing it")
	c.Flags().String(flags.Key, "", flags.DescKey)
	c.Flags().Var(flags.GapID().WithCheck(record.GapExists), flags.Answers, "the gap id this computation settles (R1-4) — REQUIRED to close a gap red minted with --check-kind computation, which prose cannot answer")
	return c
}

// truncateOutput bounds what the record carries. The FULL output is always on disk under
// <run>/proofs/<sha>/output — this is the excerpt a projection shows, and the cap exists so
// one chatty script cannot swamp every reader of the record.
func truncateOutput(s string) string {
	const cap = 2000
	if len(s) <= cap {
		return s
	}
	return s[:cap] + "\n… truncated; the full output is in <run>/proofs/<sha256>/output"
}

type proveResult struct {
	Label      string `json:"proof_id,omitempty"`
	SHA        string `json:"sha256"`
	Basis      string `json:"proof_basis,omitempty"`
	Exit       int    `json:"exit"`
	Drift      string `json:"drift,omitempty"`
	Idempotent bool   `json:"idempotent,omitempty"`
}

func (r proveResult) Human() string {
	if r.Idempotent {
		return "blue prove (idempotent retry — already recorded as " + r.SHA[:12] + ")"
	}
	out := fmt.Sprintf("proof %s recorded (%s, exit %d) — anchored, script and output cached", r.Label, r.Basis, r.Exit)
	if r.Drift != "" {
		out += "\n  NOT reproducible: " + r.Drift + " — recorded as a measurement, not a proof"
	}
	return out
}
