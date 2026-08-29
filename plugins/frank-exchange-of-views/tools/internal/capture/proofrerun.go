package capture

import (
	"fmt"
	"sort"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/proof"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// AN UNAUDITED PROOF LOOKS EXACTLY LIKE A CLEAN ONE UNTIL SOMEBODY RUNS IT.
//
// The engine already holds every piece of this. `prove` runs a script twice and grades it
// reproducible or observed; `lens reproduce` lets red re-run one and records the comparison; the
// seat's own affordance list surfaces "proofs nobody has re-run yet" as the place its next pass is
// worth most. What has never existed is anything that re-runs one WITHOUT a seat choosing to.
//
// Across the 2026-08-23 programme that gap cost the run three defective artifacts. Red recorded
// four archive spot-checks and none of them re-ran a proof — the spot-check reconciles COUNTS, and
// a count cannot see that a recorded output is wrong. `pattern_ephemeral_instrument` was staged in
// that run's own gap-pattern memory, describing this exact failure, and was not applied to it.
//
// So this is the mechanical half, and it is deliberately not a duty on a seat: a duty discharged
// by the party whose work is under audit is the shape W1.8 already had to be rescued from.
//
// # What a re-run can and cannot catch
//
// It catches a proof whose recorded output NO LONGER REPRODUCES — the ephemeral instrument (a
// measurement of state that has since moved, recorded as if it were a computation), a script
// edited after it was recorded, and a proof whose stored artifact is missing or unrunnable.
//
// It does NOT catch the wrong-working-directory proof, and claiming otherwise would be the same
// overreach this audit exists to punish. MEASURED, against the six real proofs in
// run-archive/2026-08-23_sleeper-service-plan.tar.gz: all six reproduce byte-exact — INCLUDING the
// two the retrospective calls defective. `lane3_buildstate.sh` says in its own header "run from
// repo root", takes `root="${1:-.}"`, and resolves `./plugins/sleeper-service` against the RUN
// directory instead; the `find` fails the same way on every execution, so the output carrying that
// failure reproduces perfectly. A re-run cannot see it. That class is caught at the write, by the
// error signature `blue prove` refuses on (#591 ask 1), and this audit would report those two
// proofs green all day.
//
// # Why an `observed` proof diverging is not a finding
//
// The basis is recorded BY RUNNING TWICE, and `observed` means the output already moved once. A
// live network call is the common case and a legitimate one. Failing it here would punish the
// evidence class the engine deliberately admits; it is reported and not counted against the run.
// Only a proof that CLAIMED reproducibility and no longer reproduces is a defect.
//
// # Re-running is not new reach
//
// `prove` executes these scripts twice at the write, so nothing here runs code the run has not
// already run. It is bounded twice over: a sample rather than the whole set, and proof.Timeout per
// execution.

// proofRerunSample is how many proofs one capture re-runs. Small on purpose — the audit's job is
// to make "nobody looked" impossible, not to re-execute the run.
const proofRerunSample = 3

type recordedProof struct {
	sha, basis, script string
	rerunBySeat        bool
}

// recordedProofs lists the run's proofs from the RECORD, not from the proofs/ directory.
//
// The record is where a proof's claims live — its basis, its script, its id — and it is the thing
// a reader can refuse. Enumerating the directory instead would make a proof the record never
// recorded invisible to this audit, and a proof recorded with no artifact on disk look like no
// proof at all: both of those are exactly the states worth reporting.
func recordedProofs(b *record.Board) []recordedProof {
	rerun := map[string]bool{}
	for i := range b.Events {
		if body, ok := recordpb.Body(b.Events[i]); ok {
			if r, is := body.(*recordpb.Reproduce); is && r.GetProofSha() != "" {
				rerun[r.GetProofSha()] = true
			}
		}
	}
	seen := map[string]bool{}
	var out []recordedProof
	for i := range b.Events {
		body, ok := recordpb.Body(b.Events[i])
		if !ok {
			continue
		}
		p, is := body.(*recordpb.Proof)
		if !is || p.GetProofSha() == "" || seen[p.GetProofSha()] {
			continue
		}
		seen[p.GetProofSha()] = true
		out = append(out, recordedProof{
			sha: p.GetProofSha(), basis: p.GetProofBasis(), script: p.GetScript(),
			rerunBySeat: rerun[p.GetProofSha()],
		})
	}
	// THE SAMPLE IS DETERMINISTIC AND IT PREFERS THE UNAUDITED. Proofs no seat re-ran come first
	// — that is where a re-run is worth most, and it is the same ordering the seat's own
	// affordance list uses — then by sha, so capturing a run twice checks the same proofs and this
	// audit's own result is reproducible.
	sort.Slice(out, func(i, j int) bool {
		if out[i].rerunBySeat != out[j].rerunBySeat {
			return !out[i].rerunBySeat
		}
		return out[i].sha < out[j].sha
	})
	return out
}

// ProofRerunAudit re-runs a bounded sample of the run's recorded proofs and compares each against
// the output it recorded.
func ProofRerunAudit(run record.Run, sample int) Audit {
	board, err := record.BoardState(run)
	if err != nil || board == nil {
		return Audit{Check: "proof-rerun", Verdict: "SKIP", Detail: "the record could not be read, so no proof could be re-run — which is NOT a run whose proofs reproduce"}
	}
	proofs := recordedProofs(board)
	if len(proofs) == 0 {
		return Audit{Check: "proof-rerun", Verdict: "SKIP", Detail: "this run recorded no proofs"}
	}
	if sample > len(proofs) {
		sample = len(proofs)
	}

	var fails, moved, ran int
	var notes []string
	for _, p := range proofs[:sample] {
		ran++
		matches, got, want, err := proof.Reproduce(run.Dir(), p.sha)
		switch {
		case err != nil:
			// A PROOF THAT CANNOT BE RE-RUN IS THE ONE THING A PROOF IS FOR. The recorded output
			// or the stored script is gone, so nothing in this run can check it any more.
			fails++
			notes = append(notes, fmt.Sprintf("FAIL %s could not be re-run: %v", short(p.sha), err))
		case matches:
			// Reproduced. Nothing to say per proof; the count says it.
		case p.basis == proof.Observed:
			// Recorded AS a measurement that moves. Divergence is its nature, not a defect.
			moved++
			notes = append(notes, fmt.Sprintf("moved %s diverged, as its recorded `%s` basis says it may (%s)",
				short(p.sha), proof.Observed, firstDifference(want, got)))
		default:
			fails++
			notes = append(notes, fmt.Sprintf("FAIL %s recorded basis `%s` and NO LONGER REPRODUCES — %s",
				short(p.sha), orQ(p.basis), firstDifference(want, got)))
		}
	}

	// HOW MANY WERE ACTUALLY RE-RUN IS PART OF THE ANSWER, always. "every sampled proof
	// reproduced" over a sample of zero is the plausible zero this whole audit tier exists to
	// stop printing.
	detail := fmt.Sprintf("re-ran %d of %d recorded proof(s)", ran, len(proofs))
	if unaudited := countUnaudited(proofs); unaudited > 0 {
		detail += fmt.Sprintf("; %d had never been re-run by any seat", unaudited)
	}
	if len(notes) > 0 {
		detail += "; " + strings.Join(notes, "; ")
	} else {
		detail += "; every one reproduced its recorded output"
	}
	v := "PASS"
	if fails > 0 {
		v = "FAIL"
	} else if moved > 0 {
		v = "WARN"
	}
	return Audit{Check: "proof-rerun", Verdict: v, Detail: detail}
}

func countUnaudited(ps []recordedProof) int {
	n := 0
	for _, p := range ps {
		if !p.rerunBySeat {
			n++
		}
	}
	return n
}

func short(sha string) string {
	if len(sha) > 8 {
		return sha[:8]
	}
	return sha
}

// firstDifference names WHERE the two outputs part company, not merely that they did.
//
// A diff nobody can act on is a finding nobody acts on: "output differs" sends a reader to two
// files, while a line number and the two lines send them to the cause. Bounded because an audit
// line is not a diff viewer.
func firstDifference(want, got string) string {
	w, g := strings.Split(want, "\n"), strings.Split(got, "\n")
	for i := 0; i < len(w) || i < len(g); i++ {
		lw, lg := "", ""
		if i < len(w) {
			lw = w[i]
		}
		if i < len(g) {
			lg = g[i]
		}
		if lw != lg {
			return fmt.Sprintf("first divergence at line %d: recorded %s, now %s", i+1, clip(lw), clip(lg))
		}
	}
	// Reached only when the two are equal, which the caller has already ruled out — so say that
	// rather than return an empty string a reader would read as "no difference".
	return "the outputs are byte-identical, which contradicts the comparison that produced this line"
}

func clip(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "(nothing)"
	}
	if len(s) > 80 {
		return fmt.Sprintf("%q…", s[:80])
	}
	return fmt.Sprintf("%q", s)
}
