package fuzz

import (
	"sort"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// DRIVING A VERB IS NOT DRIVING ITS SURFACE.
//
// The command-path gate asks "did every verb run", and it passes: 52 driven, 4 exempt. That
// says nothing about the flags those verbs take. A verb exercised on every sweep can have half
// its flags never passed — and an unpassed flag is code no run has executed, sitting behind a
// green sweep, which is exactly the shape #276 found at the verb level.
//
// Same for a closed set: an enum whose values are not all driven has members nothing has ever
// recorded, and a downstream consumer that mishandles one would not be caught.

// flagExemptions are flags the fuzz deliberately does not pass, each with its reason. Stated
// rather than omitted: an absence with no reason is indistinguishable from an oversight.
var flagExemptions = map[string]string{
	"json":           "driven on the paths whose JSON envelope the harness parses; a global on every verb, and passing it everywhere would change what the oracles read",
	"help":           "cobra's own",
	"version":        "cobra's own",
	"run":            "the harness passes it on every call through r.exec's prefix, before argv reaches the tally",
	"seat-id":        "same — supplied by r.exec for every seat call",
	"reason-file":    "the file/stdin twin of --reason; the prose channel is driven through --reason and the file path has its own tests in internal/cli",
	"bin-dir":        "setup's flag, and setup is exempt from the sweep (it CREATES a run; the harness builds its own)",
	"memory-dir":     "setup's flag, as above",
	"topic":          "setup's flag, as above",
	"cite":           "setup's flag, as above",
	"max-rounds":     "setup's flag, as above",
	"lanes":          "setup's flag, as above",
	"chair":          "scorecard's chair selector; the harness drives scorecard unscoped, which reads every chair",
	"format":         "graph's renderer selector; both values are driven by the format oracle, not the random sweep",
	"watch":          "dashboard's follow mode blocks until the run-live marker lifts — a sweep driving it would never return",
	"serve":          "dashboard's HTTP mode binds a port and blocks, as above",
	"now":            "dashboard's clock injection, for its own deterministic tests",
	"model":          "dashboard's tier labels, read from run-config in a real run",
	"judgment-model": "as above",
}

// enumExemptions are enum values no sweep can reach, each with its reason.
var enumExemptions = map[string]string{
	"outcome --as HALTED": "a halt reshapes every downstream oracle, so it is driven by TestFuzzHaltPath rather than the random sweep (the same reason `bench halt` is exempt from the command-path gate)",
	// `outcome --as UNVERIFIED` and `outcome --ended deadlock` were exempted here on the claim
	// that "the engine has no path to" them. It was a claim about THIS FILE'S OWN FAKE — the
	// fuzz bench returned a literal `deadlock: false` — dressed as a claim about debate.js, and
	// a real run falsified it on 2026-08-22. Both are driven now; if either stops being
	// reachable the sweep should say so rather than an exemption explaining why it never was.
}

func unreachedFlags() []string {
	execMu.Lock()
	driven := map[string]map[string]bool{}
	for k, v := range execFlags {
		driven[k] = v
	}
	ran := map[string]int{}
	for k, v := range execRuns {
		ran[k] = v
	}
	execMu.Unlock()

	var missing []string
	for path, flags := range cli.CommandFlags() {
		if ran[path] == 0 {
			continue // undriven verbs are the command-path gate's business, not this one
		}
		for _, f := range flags {
			if flagExemptions[f] != "" || driven[path][f] {
				continue
			}
			missing = append(missing, path+" --"+f)
		}
	}
	sort.Strings(missing)
	return missing
}

func unreachedEnumValues() []string {
	execMu.Lock()
	seen := map[string]map[string]bool{}
	for k, v := range execValues {
		seen[k] = v
	}
	execMu.Unlock()

	var missing []string
	for typ, fields := range record.EnumFields {
		for _, e := range fields {
			for _, v := range record.Names(e.Values) {
				if seen[e.Flag][v] {
					continue
				}
				k := typ + " --" + e.Flag + " " + v
				if enumExemptions[k] != "" {
					continue
				}
				missing = append(missing, k)
			}
		}
	}
	sort.Strings(missing)
	return missing
}

// closedGapIDs returns the gaps a spot-check can legitimately sample.
func (r *runner) closedGapIDs() []string { return r.closedInARoundBefore(0) }

// closedInARoundBefore is the same list restricted to gaps closed BEFORE the given round, which
// is what a CARRY needs. Round 0 means no restriction.
//
// A close keys on `gap_id` (keyFields puts it first), so a seat carrying a gap it closed in this
// same sitting writes a duplicate key and is refused: "already recorded a close this sitting".
// That is right — a carry restates an EARLIER round's closure, and a seat restating its own act
// from ten seconds ago is recording the same thing twice. The drive picked from every closed gap
// including its own, and in a single-round run those are the only ones there are, so `merge
// carry` ran twice across 60 runs and was refused both times. A verb whose every invocation is
// refused reports as DRIVEN in the tally, which is how the carry's own bug stayed invisible.
func (r *runner) closedInARoundBefore(round int) []string {
	b, err := record.BoardState(r.runDir)
	if err != nil {
		return nil
	}
	var out []string
	for _, id := range b.GapOrder {
		g := b.Gaps[id]
		if g == nil || g.Open {
			continue
		}
		if round > 0 && g.ClosedRound >= round {
			continue
		}
		out = append(out, id)
	}
	return out
}
