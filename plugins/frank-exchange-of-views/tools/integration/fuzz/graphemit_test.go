package fuzz

import (
	"fmt"
	"sort"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli"
)

// THE OBSERVED GRAPH (#535 step 1): which SEAT reached which VERB, and what that verb records.
//
// The sweep already answers "was every verb driven" — that is the command-path gate. It cannot
// answer the question #525 hit and could not resolve: a verb at zero across every archived run
// is either unreachable in principle or merely unused, and those are the same zero. Attributing
// each edge to a seat separates them: a recording verb no seat reaches under 40 randomised runs
// is a candidate DEAD TRANSITION, and one reached by seats the live runs happen not to exercise
// is merely unreached.
//
// It is derived on both sides, per #535's own drift rule. The edges come from argv the harness
// actually ran; the verb -> event edge comes from cli.CommandRecords(), which reads the
// annotation off the cobra tree. Nothing here is a hand-kept list, so nothing here can go stale
// while reporting coverage of a surface it never saw.
//
// WHAT THIS IS NOT: the permitted graph. It says what the fuzz DID traverse, never what a run
// MAY traverse, so an absence here is a lead and not a verdict — the fuzz may simply not drive
// the path. Separating those two is #535 step 2 and is deliberately not attempted here.

// graphEdges renders the observed seat -> verb edges, and the recording verbs no seat reached.
func graphEdges() (seats int, edges int, reached map[string][]string, unreachedRecording []string) {
	execMu.Lock()
	byVerb := map[string]map[string]bool{}
	for seat, verbs := range execEdges {
		seats++
		for v := range verbs {
			edges++
			if byVerb[v] == nil {
				byVerb[v] = map[string]bool{}
			}
			byVerb[v][seat] = true
		}
	}
	execMu.Unlock()

	reached = map[string][]string{}
	for v, ss := range byVerb {
		for s := range ss {
			reached[v] = append(reached[v], s)
		}
		sort.Strings(reached[v])
	}
	for verb := range cli.CommandRecords() {
		if len(byVerb[verb]) == 0 {
			unreachedRecording = append(unreachedRecording, verb)
		}
	}
	sort.Strings(unreachedRecording)
	return seats, edges, reached, unreachedRecording
}

// graphReport is the human-facing rendering, logged beside execReport so the two tallies of one
// sweep are read together. Every recording verb appears with the event it writes, so a reader
// never has to hold the annotation in their head to know what an edge costs the record.
func graphReport() string {
	records := cli.CommandRecords()
	seats, edges, reached, unreachedRecording := graphEdges()

	var b strings.Builder
	fmt.Fprintf(&b, "observed graph (#535 step 1): %d seats · %d seat->verb edges · %d of %d recording verbs reached",
		seats, edges, len(records)-len(unreachedRecording), len(records))
	if len(unreachedRecording) > 0 {
		// NAMED, WITH THE EVENT EACH WOULD HAVE WRITTEN. A count of unreached verbs sends a
		// reader hunting; the names with their events say what the record is missing.
		var rows []string
		for _, v := range unreachedRecording {
			rows = append(rows, v+" ["+records[v]+"]")
		}
		fmt.Fprintf(&b, "\n  recording verbs NO seat reached: %s", strings.Join(rows, ", "))
	}
	verbs := make([]string, 0, len(records))
	for v := range records {
		verbs = append(verbs, v)
	}
	sort.Strings(verbs)
	for _, v := range verbs {
		if len(reached[v]) == 0 {
			continue
		}
		fmt.Fprintf(&b, "\n  %s [%s] <- %s", v, records[v], strings.Join(reached[v], ", "))
	}
	return b.String()
}
