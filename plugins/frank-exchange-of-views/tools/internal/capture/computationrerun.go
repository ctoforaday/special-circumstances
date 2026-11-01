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

// recordedComputation is one distinct BUNDLE on the record, keyed by its sha, with every act on it.
// Several computations can share one: `compute` of a bundle already recorded records a new
// computation anchor over the same sha. The bundle is executed once and each act compared against
// that execution with its OWN recorded result, graded by its OWN basis: two acts on one bundle can
// have recorded different outputs, and one can be `variant` where another is `reproducible`.
type recordedComputation struct {
	sha         string
	argv        []string
	acts        []*recordpb.Computation
	rerunBySeat bool
}

// bases is every basis the bundle's acts were recorded with, sorted.
func (p recordedComputation) bases() []string {
	seen := map[string]bool{}
	var out []string
	for _, a := range p.acts {
		if w := recordpb.Word(a.GetBasis()); !seen[w] {
			seen[w] = true
			out = append(out, w)
		}
	}
	sort.Strings(out)
	return out
}

// recordedComputations lists the run's computations from the RECORD, not from the computations/ directory.
//
// The record is where a computation's claims live — its basis, its program, its id — and it is the
// thing a reader can refuse. Enumerating the directory instead would make a computation the record
// never recorded invisible to this audit, and a computation recorded with no artifact on disk look
// like no computation at all: both of those are exactly the states worth reporting.
func recordedComputations(evs []*record.Event) []recordedComputation {
	// A BUNDLE A SEAT RE-RAN is one with a reproduce that RAN it; a reproduce that compared against
	// an earlier execution ran nothing.
	rerun := map[string]bool{}
	for i := range evs {
		if r, ok := recordpb.BodyAs[*recordpb.Reproduce](evs[i]); ok && r.GetBundleSha() != "" && r.GetRan() {
			rerun[r.GetBundleSha()] = true
		}
	}
	at := map[string]int{} // sha → its row of out
	var out []recordedComputation
	for i := range evs {
		p, ok := recordpb.BodyAs[*recordpb.Computation](evs[i])
		if !ok || p.GetBundleSha() == "" {
			continue
		}
		k, seen := at[p.GetBundleSha()]
		if !seen {
			k = len(out)
			at[p.GetBundleSha()] = k
			out = append(out, recordedComputation{sha: p.GetBundleSha(), argv: p.GetArgv(), rerunBySeat: rerun[p.GetBundleSha()]})
		}
		out[k].acts = append(out[k].acts, p)
	}
	// THE SAMPLE IS DETERMINISTIC AND IT PREFERS THE UNAUDITED. Bundles no seat re-ran come first
	// — that is where a re-run is worth most, and it is the same ordering the seat's own
	// affordance list uses — then by sha, so capturing a run twice checks the same bundles and this
	// audit's own result is reproducible.
	sort.Slice(out, func(i, j int) bool {
		if out[i].rerunBySeat != out[j].rerunBySeat {
			return !out[i].rerunBySeat
		}
		return out[i].sha < out[j].sha
	})
	return out
}

// ComputationRerunAudit executes a bounded sample of the run's recorded bundles, once each, and
// compares every act on each against that execution.
//
// IT EXECUTES FRESH AND NEITHER READS NOR WRITES THE RE-RUN STORE. A seat's standing re-run under
// rerun/ is what the seats compared against; this audit is the check of the record that does not
// believe it, so it reads only the bundle (dir/) and each act's recorded result (acts/), and
// stores nothing.
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
	var notes, mixed []string
	for _, p := range computations[:sample] {
		ran++
		if b := p.bases(); len(b) > 1 {
			// ONE BUNDLE, ACTS OF DIFFERENT BASES: the same bytes moved between one act's two runs and
			// not between another's, so the bundle's output is not settled whatever this execution
			// shows. Each act is still graded by its own basis below.
			mixed = append(mixed, fmt.Sprintf("bundle %s backs acts recorded %s", compute.Short(p.sha), strings.Join(b, " and ")))
		}
		e, err := compute.Rerun(run.Dir(), p.sha, p.argv)
		switch {
		case err != nil:
			// A COMPUTATION THAT CANNOT BE RE-RUN IS THE ONE THING A COMPUTATION IS FOR. The stored
			// bundle is gone or cannot start, so nothing in this run can check it any more.
			fails++
			notes = append(notes, fmt.Sprintf("FAIL %s could not be re-run: %v", compute.Short(p.sha), err))
			continue
		case len(e.Calls) > 0:
			// A BUNDLE THAT CALLS THE TOOL CANNOT BE CHECKED BY ANYONE, this audit included: the
			// tool refuses every call made under a re-run, whatever the program then prints.
			fails++
			notes = append(notes, fmt.Sprintf("FAIL %s calls feov-record (%s), which refuses every call made under a re-run, so no checker can run it", compute.Short(p.sha), e.Calls[0]))
			continue
		}
		if e.Leftover != "" {
			// The execution finished before its copy's removal failed, so it is graded below; the
			// leftover is this audit's own defect, said beside the grades.
			notes = append(notes, fmt.Sprintf("the re-run of %s left its copy behind: %s", compute.Short(p.sha), e.Leftover))
		}
		for _, a := range p.acts {
			c, err := compute.Compare(run.Dir(), p.sha, a.GetComputationId(), int(a.GetExit()), e)
			basis := recordpb.Word(a.GetBasis())
			if err != nil {
				fails++
				notes = append(notes, fmt.Sprintf("FAIL %s on %s could not be compared: %v", a.GetComputationId(), compute.Short(p.sha), err))
				continue
			}
			switch compute.Grade(c, a.GetBasis() == recordpb.Basis_BASIS_VARIANT) {
			case compute.Reproduced:
				// Nothing to say per act; the count says it.
			case compute.Moved:
				// Recorded AS something that does not hold still: a `variant` computation's two runs
				// already differed, and its invariant says what holds across them. Divergence here is
				// the recorded fact arriving on time, not a defect — and this audit becoming a
				// REGRESSION detector rather than a discoverer of what compute already knew is the
				// whole point of grading from the store.
				moved++
				notes = append(notes, fmt.Sprintf("moved %s on %s diverged, as its recorded `%s` basis says it may (%s)",
					a.GetComputationId(), compute.Short(p.sha), orQ(basis), difference(c, e)))
			case compute.Differs:
				fails++
				notes = append(notes, fmt.Sprintf("FAIL %s on %s recorded basis `%s` and NO LONGER REPRODUCES — %s",
					a.GetComputationId(), compute.Short(p.sha), orQ(basis), difference(c, e)))
			}
		}
	}

	// HOW MANY WERE ACTUALLY RE-RUN IS PART OF THE ANSWER, always. "every sampled computation
	// reproduced" over a sample of zero is the plausible zero this whole audit tier exists to
	// stop printing.
	// BUNDLES, NOT COMPUTATIONS, ARE WHAT IS RE-RUN, and the count names both: smoke 07d's record held 9
	// computations over 6 bundles, and "3 of 6 recorded proof(s)" undercounted the record by a third.
	covered, total := 0, 0
	for i, p := range computations {
		total += len(p.acts)
		if i < sample {
			covered += len(p.acts)
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
	if len(mixed) > 0 {
		detail += "; " + strings.Join(mixed, "; ")
	}
	v := "PASS"
	if fails > 0 {
		v = "FAIL"
	} else if moved > 0 {
		v = "WARN"
	}
	return Audit{Check: "computation-rerun", Verdict: v, Detail: detail}
}

// difference names where an act's recorded result and the execution part: the first line of output
// that differs, or what else moved where the output agrees.
func difference(c compute.Comparison, e compute.Execution) string {
	if c.Recorded != e.Output {
		return firstDifference(c.Recorded, e.Output)
	}
	return c.Drift
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
