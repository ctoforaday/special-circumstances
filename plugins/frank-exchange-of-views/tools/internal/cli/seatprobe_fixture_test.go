package cli

import (
	"os"
	"path/filepath"
	"testing"
)

// A BOARD FOR PROBING WHETHER A SEAT USES THE TOOL AT ALL.
//
// # The question this exists to ask
//
// Every other harness here asks whether the tool ACCEPTS what a seat sends. This one is built for
// the opposite question, and it is the more expensive failure: **does a seat reach for the verb,
// or does it route around into prose and a hand-written file?**
//
// Nothing in the suite can answer that. The fuzz IS a driver — it calls the verb by construction,
// so it measures the tool and says nothing about whether a real seat would find it. The measured
// shape is in the friction logs of real runs: a seat that cannot find the verb it wants "logs
// friction and works around it, losing the capability for the run", and a seat that finds prose
// easier than a flag does the same thing without even noticing.
//
// So: build a board a seat would actually MEET, hand it to a weak model under its real
// constitution, and read back which verbs it used and which it talked its way around. The output
// is not pass/fail — it is a friction corpus and a list of verbs nobody chose. That belongs to
// the operator, not to CI (#363).
//
// # Why a weak model
//
// A strong seat compensates for a bad help string by inferring what was meant. That hides exactly
// the defect being hunted: the constitution and the `--help` text are the only things standing
// between a seat and a hand-written artifact, and their quality is only visible when the reader
// is not clever enough to paper over it. Haiku is the instrument, not the subject.
//
// # Why it is built through the CLI rather than the record API
//
// The fixture must be a board the real write paths would produce. Writing events directly would
// let this file record a state no seat could ever have reached — a gap with no acceptance check,
// a closure with no anchor — and a probe run against an impossible board teaches nothing about
// the possible one.
//
// # Running it
//
//	FEOV_SEAT_PROBE_DIR=<path> go test ./internal/cli -run TestWriteSeatProbeFixture -count=1
//
// Env-guarded because it writes outside the test's temp directory, which is the one thing a test
// must not do by default.
func TestWriteSeatProbeFixture(t *testing.T) {
	dest := os.Getenv("FEOV_SEAT_PROBE_DIR")
	if dest == "" {
		t.Skip("set FEOV_SEAT_PROBE_DIR to write the seat-probe fixture")
	}
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())

	runDir := dest
	if err := os.MkdirAll(filepath.Join(runDir, "blue"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(runDir, "blue", "report.md"), []byte(probeReport), 0o644); err != nil {
		t.Fatal(err)
	}

	for _, s := range []struct{ role, id string }{
		{"lens", "red-lens-r1-L1"},
		{"lens", "red-lens-r1-L5"},
		{"merge", "red-merge-r1"},
		{"blue", "blue-respond-r1"},
		{"bench", "judge-r1"},
	} {
		if _, err := run(t, s.role, "register", "--run", runDir, "--seat-id", s.id); err != nil {
			t.Fatalf("register %s: %v", s.id, err)
		}
	}

	// THE BOARD IS BUILT SO THAT EACH GAP BAITS A DIFFERENT VERB, and one baits none — the
	// control. What a seat does with each is the measurement.
	//
	// Every gap here is minted with grades a reasonable seat might contest, because a board
	// nobody could argue with measures nothing about the arguing channel.
	gaps := []struct {
		key, class, location, problem, fix, check, kind string
		sev, lik, imp, cx                               string
		baits                                           string
	}{
		{
			key: "figure", class: "figure-recount-fails",
			location: `## Findings — "The corpus holds 340 records across 12 sources."`,
			problem:  "The stated total does not match the per-source figures, which sum to 331.",
			fix:      "Recompute the total from the per-source table and correct whichever is wrong.",
			check:    "The stated total equals the sum of the per-source rows.", kind: "computation",
			sev: "high", lik: "certain", imp: "high", cx: "low",
			baits: "blue prove — a computation check CANNOT be closed by prose, so the seat must run something",
		},
		{
			key: "universal", class: "false-universal",
			location: `## Method — "No reader outside the plugin directory touches the event log."`,
			problem:  "Stated as a universal over a tree nobody enumerated in the report.",
			fix:      "Bound the claim to what was actually swept, or show the sweep.",
			check:    "The claim names the scope it was verified over.", kind: "document",
			sev: "high", lik: "high", imp: "high", cx: "medium",
			baits: "motion grade file — `certain` likelihood on a claim the report itself hedges is contestable",
		},
		{
			key: "attestation", class: "self-attestation",
			location: `## Method — "Each figure was independently checked."`,
			problem:  "The checking is asserted and nothing records that it happened.",
			fix:      "Record what was checked and how, or drop the claim.",
			check:    "The report names the artifact each check ran against.", kind: "document",
			sev: "medium", lik: "medium", imp: "medium", cx: "low",
			baits: "blue edit + --answers — the ordinary repair path, and the control for the others",
		},
		{
			key: "scope", class: "verification-scope-blindspot",
			location: `## Limits — "The remaining cases are out of scope."`,
			problem:  "The out-of-scope set is named and never enumerated, so its size is unknown.",
			fix:      "Enumerate the excluded cases or state why the count cannot be had.",
			check:    "The excluded set has a stated size or a stated reason it cannot be counted.", kind: "document",
			sev: "low", lik: "low", imp: "medium", cx: "high",
			baits: "risk_accepted closure, or an avenue — the fix costs more than the defect",
		},
	}
	for _, g := range gaps {
		if _, err := run(t, "merge", "mint", "--run", runDir, "--seat-id", "red-merge-r1",
			"--key", g.key, "--class", g.class,
			"--location", g.location, "--problem", g.problem, "--fix", g.fix,
			"--check", g.check, "--check-kind", g.kind,
			"--severity", g.sev, "--likelihood", g.lik, "--impact", g.imp, "--cx", g.cx,
			"--existence", "verified",
			"--reason", g.problem+" (baits: "+g.baits+")"); err != nil {
			t.Fatalf("mint %s: %v", g.key, err)
		}
	}

	// A ruled direction and an unruled one: the first invites an appeal or a move, the second
	// invites the merge to rule. Both are verbs that measured ZERO real use before they had a
	// projection to render into.
	for _, a := range []struct{ line, hyp string }{
		{"re-derive the per-source counts from the raw corpus", "the 331 figure is the correct one"},
		{"survey how comparable projects state their scope limits", "there is a standard form we are ignoring"},
	} {
		if _, err := run(t, "blue", "avenue", "--run", runDir, "--seat-id", "blue-respond-r1",
			"--line", a.line, "--hypothesis", a.hyp); err != nil {
			t.Fatalf("avenue: %v", err)
		}
	}
	if _, err := run(t, "motion", "direction", "rule", "--run", runDir, "--seat-id", "red-merge-r1",
		"--id", "A2", "--as", "too-thin",
		"--reason", "a survey of how others phrase it does not settle whether OUR claim is bounded"); err != nil {
		t.Fatalf("direction rule: %v", err)
	}

	t.Logf("seat-probe fixture written to %s: 4 gaps, 2 avenues (A2 ruled too-thin), 5 seats registered", runDir)
}

// probeReport is the artifact under audit. It carries the exact sentences the gaps' --location
// fields quote, because a finding is refused unless its quote is present — and a fixture whose
// gaps point at absent text is the defect a real seat caught us shipping once already.
const probeReport = `# Does the event log have readers outside its own tree? — research report

## TL;DR

The corpus was swept and no reader outside the plugin directory touches the event log. The
figures below are drawn from that sweep.

## Findings

The corpus holds 340 records across 12 sources. Broken down by source: 41, 38, 12, 55, 9, 22,
17, 31, 28, 14, 36, 28.

Each source was read at its pinned revision.

## Method

No reader outside the plugin directory touches the event log. Each figure was independently
checked.

## Limits

The remaining cases are out of scope.
`
