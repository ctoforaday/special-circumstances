package capture

import (
	"fmt"
	"sort"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/compute"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// AN UNAUDITED COMPUTATION LOOKS EXACTLY LIKE A CLEAN ONE UNTIL SOMEBODY RUNS IT.
//
// The engine already holds every piece of this. `compute` runs a program twice and grades it
// reproducible or variant; `lens reproduce` lets red re-run one and records the comparison; the
// seat's own affordance list surfaces "computations nobody has re-run yet" as the place its next pass is
// worth most. What has never existed is anything that re-runs one WITHOUT a seat choosing to.
//
// Across the 2026-08-23 programme that gap cost the run three defective artifacts. Red recorded
// four archive spot-checks and none of them re-ran a computation — the spot-check reconciles COUNTS, and
// a count cannot see that a recorded output is wrong. The ephemeral-instrument failure was already
// a known one when that run made three of it.
//
// So this is the mechanical half, and it is deliberately not a duty on a seat: a duty discharged
// by the party whose work is under audit is the shape W1.8 already had to be rescued from.
//
// # What a re-run can and cannot catch
//
// It catches a computation whose recorded output NO LONGER REPRODUCES — the ephemeral instrument (a
// measurement of state that has since moved, recorded as if it were a computation), a script
// edited after it was recorded, and a computation whose stored artifact is missing or unrunnable.
//
// It does NOT catch the wrong-working-directory computation, and claiming otherwise would be the same
// overreach this audit exists to punish. MEASURED, against the six real computations in
// run-archive/2026-08-23_sleeper-service-plan.tar.gz: all six reproduce byte-exact — INCLUDING the
// two the retrospective calls defective. `lane3_buildstate.sh` says in its own header "run from
// repo root", takes `root="${1:-.}"`, and resolves `./plugins/sleeper-service` against the RUN
// directory instead; the `find` fails the same way on every execution, so the output carrying that
// failure reproduces perfectly. A re-run cannot see it. That class is caught at the write, by the
// error signature `blue compute` refuses on (#591 ask 1), and this audit would report those two
// computations green all day.
//
// # Why a `variant` computation diverging is not a finding
//
// The basis is recorded BY RUNNING TWICE, and `variant` means the output already moved once. A
// live network call is the common case and a legitimate one. Failing it here would punish the
// evidence class the engine deliberately admits; it is reported and not counted against the run.
// Only a computation that CLAIMED reproducibility and no longer reproduces is a defect.
//
// # Re-running is not new reach
//
// `compute` executes these scripts twice at the write, so nothing here runs code the run has not
// already run. It is bounded twice over: a sample rather than the whole set, and compute.Timeout per
// execution.

// computationRerunSample is how many computations one capture re-runs. Small on purpose — the audit's job is
// to make "nobody looked" impossible, not to re-execute the run.
const computationRerunSample = 3

// recordedComputation is one distinct BUNDLE on the record, keyed by its sha. Several computations can
// share one: `compute` of a bundle already recorded records a new computation anchor over the same
// sha, and a re-run of the bundle answers for every computation that holds it. computations counts
// them; first is the earliest act on the bundle, whose recorded result the re-run is compared with.
type recordedComputation struct {
	sha, basis   string
	first        *recordpb.Computation
	rerunBySeat  bool
	computations int
}

// recordedComputations lists the run's computations from the RECORD, not from the computations/ directory.
//
// The record is where a computation's claims live — its basis, its script, its id — and it is the thing
// a reader can refuse. Enumerating the directory instead would make a computation the record never
// recorded invisible to this audit, and a computation recorded with no artifact on disk look like no
// computation at all: both of those are exactly the states worth reporting.
func recordedComputations(evs []*record.Event) []recordedComputation {
	// A BUNDLE A SEAT RE-RAN is one with a reproduce that RAN it; a reproduce that judged its
	// anchor on an earlier re-run ran nothing.
	rerun := map[string]bool{}
	for i := range evs {
		if body, ok := recordpb.Body(evs[i]); ok {
			if r, is := body.(*recordpb.Reproduce); is && r.GetBundleSha() != "" && r.GetRan() {
				rerun[r.GetBundleSha()] = true
			}
		}
	}
	at := map[string]int{} // sha → its row of out
	var out []recordedComputation
	for i := range evs {
		body, ok := recordpb.Body(evs[i])
		if !ok {
			continue
		}
		p, is := body.(*recordpb.Computation)
		if !is || p.GetBundleSha() == "" {
			continue
		}
		if k, seen := at[p.GetBundleSha()]; seen {
			out[k].computations++
			continue
		}
		at[p.GetBundleSha()] = len(out)
		out = append(out, recordedComputation{
			sha: p.GetBundleSha(), basis: recordpb.Word(p.GetBasis()), first: p,
			rerunBySeat: rerun[p.GetBundleSha()], computations: 1,
		})
	}
	// THE SAMPLE IS DETERMINISTIC AND IT PREFERS THE UNAUDITED. Computations no seat re-ran come first
	// — that is where a re-run is worth most, and it is the same ordering the seat's own
	// affordance list uses — then by sha, so capturing a run twice checks the same computations and this
	// audit's own result is reproducible.
	sort.Slice(out, func(i, j int) bool {
		if out[i].rerunBySeat != out[j].rerunBySeat {
			return !out[i].rerunBySeat
		}
		return out[i].sha < out[j].sha
	})
	return out
}

// ComputationRerunAudit re-runs a bounded sample of the run's recorded computations and compares each against
// the output it recorded.
func ComputationRerunAudit(run record.Run, sample int) Audit {
	fam, err := record.FamilyOf(run)
	if err != nil {
		return Audit{Check: "computation-rerun", Verdict: "SKIP", Detail: "the record could not be read, so no computation could be re-run — which is NOT a run whose computations reproduce"}
	}
	computations := recordedComputations(fam.Events)
	if len(computations) == 0 {
		return Audit{Check: "computation-rerun", Verdict: "SKIP", Detail: "this run recorded no computations"}
	}
	if sample > len(computations) {
		sample = len(computations)
	}

	var fails, moved, ran int
	var notes []string
	for _, p := range computations[:sample] {
		ran++
		rr, err := compute.Reproduce(run.Dir(), p.sha, p.first.GetArgv(), p.first.GetComputationId(), int(p.first.GetExit()))
		got, want := rr.Got, rr.Want
		switch {
		case err != nil:
			// A COMPUTATION THAT CANNOT BE RE-RUN IS THE ONE THING A COMPUTATION IS FOR. The recorded output
			// or the stored bundle is gone, so nothing in this run can check it any more.
			fails++
			notes = append(notes, fmt.Sprintf("FAIL %s could not be re-run: %v", short(p.sha), err))
		case len(rr.ToolCalls) > 0:
			// A BUNDLE THAT CALLS THE TOOL CANNOT BE CHECKED BY ANYONE, this audit included: the
			// tool refuses every call made under a re-run, whatever the program then prints.
			fails++
			notes = append(notes, fmt.Sprintf("FAIL %s calls feov-record (%s), which refuses every call made under a re-run, so no checker can run it", short(p.sha), rr.ToolCalls[0]))
		case rr.Matches:
			// Reproduced. Nothing to say per computation; the count says it.
		case p.basis == recordpb.Word(recordpb.Basis_BASIS_VARIANT):
			// Recorded AS something that does not hold still: a `variant` computation's two runs
			// already differed, and its invariant says what holds across them. Divergence here is
			// the recorded fact arriving on time, not a defect — and this audit becoming a
			// REGRESSION detector rather than a discoverer of what compute already knew is the
			// whole point of grading from the store.
			moved++
			notes = append(notes, fmt.Sprintf("moved %s diverged, as its recorded `%s` basis says it may (%s)",
				short(p.sha), orQ(p.basis), firstDifference(want, got)))
		default:
			fails++
			notes = append(notes, fmt.Sprintf("FAIL %s recorded basis `%s` and NO LONGER REPRODUCES — %s",
				short(p.sha), orQ(p.basis), firstDifference(want, got)))
		}
	}

	// HOW MANY WERE ACTUALLY RE-RUN IS PART OF THE ANSWER, always. "every sampled computation
	// reproduced" over a sample of zero is the plausible zero this whole audit tier exists to
	// stop printing.
	// BUNDLES, NOT COMPUTATIONS, ARE WHAT IS RE-RUN, and the count names both: smoke 07d's record held 9
	// computations over 6 bundles, and "3 of 6 recorded proof(s)" undercounted the record by a third.
	covered, total := 0, 0
	for i, p := range computations {
		total += p.computations
		if i < sample {
			covered += p.computations
		}
	}
	detail := fmt.Sprintf("re-ran %d of %d distinct bundle(s), behind %d of the %d recorded computation(s)", ran, len(computations), covered, total)
	if unaudited := countUnaudited(computations); unaudited > 0 {
		detail += fmt.Sprintf("; %d bundle(s) had never been re-run by any seat", unaudited)
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
	return Audit{Check: "computation-rerun", Verdict: v, Detail: detail}
}

func countUnaudited(ps []recordedComputation) int {
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
