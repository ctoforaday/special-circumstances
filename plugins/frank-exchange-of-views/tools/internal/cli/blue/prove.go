package blue

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/lens"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/proof"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
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
	c := seat.Prose(seat.New("prove",
		`settle a claim by COMPUTING it: --location "<the exact sentence>" --script <path under the run dir> [--cites <the method's citation label>] — the tool runs it twice, caches the script and its output, and splices an invisible proof anchor at that sentence`,
		func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
			location, script := seat.Str(cmd, flags.Location), seat.Str(cmd, flags.Script)
			if strings.TrimSpace(location) == "" {
				return nil, fmt.Errorf("blue prove requires --location: the EXACT sentence in blue/report.md this computation backs — a proof anchored to nothing is a script nobody can connect to a claim")
			}
			if strings.TrimSpace(script) == "" {
				return nil, fmt.Errorf("blue prove requires --script: the path (under the run directory) of the program that settles it")
			}

			// Crash-retry: a committed proof for this key is already recorded.
			if prior, err := record.ExistingProofByKey(s.RunDir, s.SeatID, seat.Str(cmd, flags.Key)); err != nil {
				return nil, err
			} else if prior != "" {
				return proveResult{SHA: prior, Idempotent: true}, nil
			}

			res, err := proof.Run(s.RunDir, script)
			if err != nil {
				// A proof that will not run is not evidence, and the failure is a capability
				// signal — the same treatment `blue cite` gives an unreachable source.
				msg := err.Error()
				if _, ferr := record.Append(s.Identity(), "friction", record.NewPayload().Set("reason", msg)); ferr != nil {
					return nil, ferr
				}
				return nil, err
			}

			label := record.NewProofID()
			marker := "<!--proof:" + label + "-->"
			// Spliced under the report lock at the quoted sentence, by the SAME machinery a
			// citation anchor uses: one immortal-anchor mechanism, three classes.
			if err := record.MutateBlueReport(s.RunDir, func(old []byte) ([]byte, error) {
				next, aerr := lens.InsertAnchor(old, location, marker)
				if aerr != nil {
					return nil, aerr
				}
				return next, nil
			}); err != nil {
				return nil, err
			}

			p := record.NewPayload().
				Set("proof_id", label).
				Set("sha256", res.SHA).
				Set("proof_basis", res.Basis).
				Set("script", res.Script).
				Set("exit", res.Exit).
				Set("location", location).
				Set("output", truncateOutput(res.Output))
			if res.Drift != "" {
				p.Set("drift", res.Drift)
			}
			seat.Set(cmd, p, "proof_key", flags.Key)
			seat.Set(cmd, p, "answers", flags.Answers)
			seat.Set(cmd, p, "cites", flags.Cites)
			if err := seat.SetReason(cmd, p, "reason"); err != nil {
				return nil, err
			}
			if _, err := record.Append(s.Identity(), "proof", p); err != nil {
				return nil, err
			}
			return proveResult{Label: label, SHA: res.SHA, Basis: res.Basis, Exit: res.Exit, Drift: res.Drift}, nil
		}))

	c.Flags().String(flags.Location, "", "REQUIRED — the EXACT sentence in blue/report.md this computation backs (it must appear verbatim; the anchor is spliced there)")
	c.Flags().String(flags.Script, "", "REQUIRED — path under the run directory of the program that settles it (.py, .js, .mjs, .sh or .go)")
	c.Flags().Var(flags.CitationAnchor().WithCheck(record.CitationExists), flags.Cites, "the citation label of the METHOD this applies — the source that says trial division or Miller-Rabin decides primality. The method is cited; the instance is computed")
	c.Flags().String(flags.Key, "", "a stable local handle (P1, P2 …) making a retried prove idempotent")
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
