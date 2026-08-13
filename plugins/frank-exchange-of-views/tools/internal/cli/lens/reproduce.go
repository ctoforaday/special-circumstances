package lens

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/enumhelp"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/proof"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
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
// leaves its result to the reader — which held for near-match and claim-index, and did not hold
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
	c := seat.Prose(seat.New("reproduce",
		`re-RUN a proof AND READ IT: --id <sha256> --as sound|unsound --reason "<what the script actually computes>" — the tool records whether it REPRODUCED; you judge whether it PROVES the claim`,
		func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
			sha := seat.Str(cmd, flags.ID)
			if sha == "" {
				// THE WAY THROUGH, NAMED. This message used to say the sha was "listed in the report
				// beside the sentence it backs". It is not: the report carries an opaque
				// `<!--proof:p-…-->` anchor, and the sha lives on the record. A seat reading the
				// document had the token this verb does not take and no path to the one it does.
				return nil, fmt.Errorf("lens reproduce requires --id: the sha256 of the proof to re-run. Reading the report and holding a `<!--proof:p-…-->` anchor, resolve it with `lens show evidence --run <runDir>` — every proof is listed there with its anchor, its sha256, its script, and whether anyone has re-run it yet")
			}
			ok, got, want, err := proof.Reproduce(s.RunDir, sha)
			if err != nil {
				return nil, err
			}
			soundness := seat.Str(cmd, flags.As)
			if soundness == "" {
				return nil, fmt.Errorf("lens reproduce requires --as sound|unsound: re-running proves the script is DETERMINISTIC, not that it establishes the claim — `print(\"7 is prime\")` reproduces forever. Read the script and say whether it computes what it is anchored to")
			}
			p := record.NewPayload().Set("proof_sha", sha).Set("reproduced", ok).Set("soundness", soundness)
			if !ok {
				// Only on a MISMATCH, and truncated: the recorded output already lives in the
				// proof artifact, and copying it wholesale into the event would put a second
				// copy of the same bytes on the record. What the reader needs is what CHANGED.
				p.Set("recorded_output", truncateOutput(want)).Set("observed_output", truncateOutput(got))
			}
			if err := seat.SetReason(cmd, p, "note"); err != nil {
				return nil, err
			}
			if p.Str("note") == "" {
				return nil, fmt.Errorf("lens reproduce requires --reason: say what the script ACTUALLY COMPUTES, in your words. A soundness verdict with no reading behind it is the assertion this verb exists to replace")
			}
			if _, err := record.Append(s.RunDir, s.SeatID, "reproduce", p); err != nil {
				return nil, err
			}
			return reproduceResult{SHA: sha, Matches: ok, Got: got, Recorded: want}, nil
		}))
	c.Flags().String(flags.ID, "", "REQUIRED — the sha256 of the recorded proof to re-run")
	enumhelp.Flag(c, flags.As, record.MustEnum("reproduce", "soundness"), ("REQUIRED — having READ the script: does it actually establish the claim it is anchored to?"))
	return c
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
	SHA      string `json:"sha256"`
	Matches  bool   `json:"matches"`
	Got      string `json:"got"`
	Recorded string `json:"recorded"`
}

func (r reproduceResult) Human() string {
	if r.Matches {
		return "proof " + r.SHA[:12] + " REPRODUCES — re-running it gives the recorded output byte for byte"
	}
	return "proof " + r.SHA[:12] + " DOES NOT REPRODUCE.\n  recorded: " + r.Recorded + "\n  now:      " + r.Got +
		"\n  Either the computation depends on something that moved, or the recorded output is not what the script produces. Both are findings; neither is recorded for you."
}
