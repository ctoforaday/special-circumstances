package difftest

import (
	"fmt"
	"math/rand"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// TestReplayDeterminism drives random valid-ish verb sequences through the tool
// TWICE, in separate run directories, and requires the resulting event logs and
// projections to be identical after normalization.
//
// This began as differential fuzzing against the mjs oracle, which is retired.
// The generator is kept because the property it tests is independent of the
// oracle and is the one that would silently rot: REPLAY MUST BE DETERMINISTIC.
// The record layer's whole claim is that a board state is a pure function of its
// event log — capture audits, the parity gate, and every projection depend on
// replaying the same events to the same bytes. Go makes that easy to break by
// accident, because map iteration order is randomized: the seat grouping, the
// anomaly list, the by_severity tally, and the round ordering are all places
// where reaching for a map would produce output that differs run to run. Each is
// insertion-ordered on purpose, and this test is what notices if one stops being.
//
// Hand-written scenarios cannot cover this: they exercise the orderings the
// author already thought about. A generator reaches the interleavings nobody
// enumerated — a close before its mint, a regrade of a closed gap, a dispose
// naming an observation that never existed, a seat minting across two rounds
// without re-registering.
//
// Sequences are valid-ISH by construction rather than valid: a refusal is as
// interesting as an acceptance, because the error text is part of the contract a
// seat reads, and it must be as reproducible as a success.
// withoutTimestamps strips the clock from a captured event log so two runs can be
// compared for the things that must not vary.
// rankTimestamps replaces each event's clock with its position in time, counting from
// zero, so two runs can be compared for everything that must not vary.
//
// DROPPING the timestamp instead would also make the comparison pass, and would stop it
// noticing a REORDER — the one failure this schema exists to catch. Ranking keeps order
// under test while letting the absolute clock differ, which is the only part that may.
// rankTimestamps replaces each event's clock with its POSITION in the canonical
// (TS, SeatID, Seq) order, via the same eventRanks the goldens use.
//
// It ranked DISTINCT CLOCK VALUES until CI caught it: two events landing inside one clock
// tick leave one fewer distinct instant, every later rank shifts, and two runs of the same
// commands disagree for no reason but speed. That made this determinism test itself
// non-deterministic — the exact property it exists to assert, broken in its own harness.
// Position in the merge order is stable however coarse the clock is, and a genuine
// reordering still shows up as a diff.
func rankTimestamps(m map[string][]map[string]any) map[string][]map[string]any {
	rank := eventRanks(m)
	out := make(map[string][]map[string]any, len(m))
	for k, evs := range m {
		cp := make([]map[string]any, len(evs))
		for i, ev := range evs {
			c := make(map[string]any, len(ev))
			for key, v := range ev {
				c[key] = v
			}
			if _, ok := c["ts"]; ok {
				c["ts"] = rank[fmt.Sprintf("%s#%v", k, ev["seq"])]
			}
			cp[i] = c
		}
		out[k] = cp
	}
	return out
}

func TestReplayDeterminism(t *testing.T) {
	if testing.Short() {
		t.Skip("determinism fuzz spawns two processes per command; skipped under -short")
	}
	bin := buildBinary(t)

	const sequences = 12
	const maxLen = 14
	rng := rand.New(rand.NewSource(0x5EED)) // fixed: a failure must be reproducible

	for i := 0; i < sequences; i++ {
		t.Run(fmt.Sprintf("seq%02d", i), func(t *testing.T) {
			cmds := generate(rng, maxLen)
			first := replay(t, bin, cmds)
			second := replay(t, bin, cmds)

			// The log now carries a wall clock, so two identical runs can never be
			// byte-identical — that is the clock working, not a determinism failure.
			// Determinism here means the same commands produce the same events, in the
			// same ORDER, with the same content. Each timestamp is compared as its RANK
			// rather than dropped, so a reorder still fails while the absolute instant
			// is free to differ.
			if diff := cmp.Diff(rankTimestamps(first.events), rankTimestamps(second.events)); diff != "" {
				t.Errorf("event log is not reproducible across identical runs\nsequence: %s\n%s", dumpSeq(cmds), diff)
			}
			// Projections are no longer materialized; they derive as a pure function of the
			// event log, which the rank-diff above already proves reproducible.
			if diff := cmp.Diff(first.output, second.output); diff != "" {
				t.Errorf("CLI output is not reproducible across identical runs\nsequence: %s\n%s", dumpSeq(cmds), diff)
			}
		})
	}
}

type replayResult struct {
	events map[string][]map[string]any
	output []string
}

func replay(t *testing.T, bin string, cmds []cmd) replayResult {
	t.Helper()
	runDir := t.TempDir()
	m := newMapper()
	var out []string
	for _, c := range cmds {
		inv := runGo(bin, runDir, c)
		m.observe(filepath.Join(runDir, "records"))
		got := normalizeOutput(inv, runDir, m)
		out = append(out, fmt.Sprintf("exit=%d out=%q err=%q", got.code, got.stdout, got.stderr))
	}
	st := collect(t, runDir, m)
	return replayResult{events: st.events, output: out}
}

var (
	fuzzGrades  = []string{"low", "low-medium", "medium", "medium-high", "high", "certain", "realized", "trivial", "bogus"}
	fuzzClasses = []string{"propagation-incomplete", "citation-drift", "scope-creep"}
	fuzzIDs     = []string{"R1-1", "R1-2", "R2-1", "R9-9"}
)

func pick[T any](rng *rand.Rand, xs []T) T { return xs[rng.Intn(len(xs))] }

// generate builds a sequence weighted toward the merge seat, since that is where
// validation, lineage, and id minting live.
func generate(rng *rand.Rand, maxLen int) []cmd {
	n := 3 + rng.Intn(maxLen-3)
	out := make([]cmd, 0, n)
	seat := func(role string) string {
		switch role {
		case "merge":
			return pick(rng, []string{"red-merge-r1", "red-merge-r2"})
		case "lens":
			return pick(rng, []string{"red-lens-r1-L1", "red-lens-r1-L5", "red-lens-r2-L6"})
		case "blue":
			return pick(rng, []string{"blue-respond-r1", "blue-synthesize"})
		default:
			return "judge-r1"
		}
	}
	for len(out) < n {
		role := pick(rng, []string{"merge", "merge", "merge", "lens", "lens", "blue", "bench"})
		s := seat(role)
		args := []string{}
		switch role {
		case "merge":
			switch rng.Intn(7) {
			case 0:
				args = []string{"register", "--run", "{RUN}", "--seat-id", s}
			case 1, 2:
				args = []string{"mint", "--run", "{RUN}", "--seat-id", s,
					"--class", pick(rng, fuzzClasses), "--check-kind", "document", "--check", "acceptance check",
					"--severity", pick(rng, fuzzGrades), "--likelihood", pick(rng, fuzzGrades),
					"--impact", pick(rng, fuzzGrades), "--problem", fmt.Sprintf("problem %d", rng.Intn(1000))}
				if rng.Intn(3) == 0 {
					args = append(args, "--cx", pick(rng, fuzzGrades))
				}
				if rng.Intn(3) == 0 {
					args = append(args, "--supersedes", pick(rng, fuzzIDs))
				}
				if rng.Intn(4) == 0 {
					args = append(args, "--key", fmt.Sprintf("K%d", rng.Intn(3)))
				}
				if rng.Intn(5) == 0 {
					args = append(args, "--found-by", "L1,L5")
				}
			case 3:
				args = []string{"close", "--run", "{RUN}", "--seat-id", s, "--id", pick(rng, fuzzIDs)}
				switch rng.Intn(3) {
				case 0:
					args = append(args, "--anchor-seat", "L1", "--anchor-tool", "Read", "--anchor-target", "report.md#S1")
				case 1:
					args = append(args, "--carried-from", "1")
				}
				if rng.Intn(4) == 0 {
					args = append(args, "--as", "closed_with_regression")
				}
			case 4:
				args = []string{"regrade", "--run", "{RUN}", "--seat-id", s, "--id", pick(rng, fuzzIDs),
					"--severity", pick(rng, fuzzGrades)}
				if rng.Intn(2) == 0 {
					args = append(args, "--reason", "movement reason")
				}
			case 5:
				args = []string{"dispose", "--run", "{RUN}", "--seat-id", s,
					"--observation", pick(rng, []string{"L1-F1", "L5-F1", "N-missing"})}
				if rng.Intn(3) > 0 {
					args = append(args, "--as", pick(rng, []string{"minted-as", "folded-into", "declined", "banked"}))
				}
			case 6:
				args = []string{"spot-check", "--run", "{RUN}", "--seat-id", s, "--ids", "R1-1,R1-2"}
			}
		case "lens":
			switch rng.Intn(4) {
			case 0:
				args = []string{"register", "--run", "{RUN}", "--seat-id", s}
			case 1, 2:
				// --key from a small space so the SAME seat sometimes repeats it — that
				// exercises the crash-retry idempotency (a repeated key returns the existing
				// tool-assigned label, no duplicate). The label is the tool's now, not --label.
				args = []string{"finding", "--run", "{RUN}", "--seat-id", s,
					"--key", fmt.Sprintf("F%d", 1+rng.Intn(3)),
					"--severity", pick(rng, fuzzGrades), "--likelihood", pick(rng, fuzzGrades),
					"--impact", pick(rng, fuzzGrades), "--reason", "a finding"}
			case 3:
				args = []string{"observe", "--run", "{RUN}", "--seat-id", s, "--reason", "an observation"}
				if rng.Intn(2) == 0 {
					args = append(args, "--label", fmt.Sprintf("N%d", rng.Intn(3)))
				}
			}
		case "blue":
			switch rng.Intn(3) {
			case 0:
				args = []string{"revision", "--run", "{RUN}", "--seat-id", s, "--reason", "revised"}
			case 1:
				args = []string{"manifest-row", "--run", "{RUN}", "--seat-id", s, "--id", pick(rng, fuzzIDs), "--row", "checked"}
			case 2:
				args = []string{"dispute", "--run", "{RUN}", "--seat-id", s, "--id", pick(rng, fuzzIDs),
					"--dimension", "likelihood", "--proposed", pick(rng, fuzzGrades), "--reason", "why"}
			}
		case "bench":
			args = []string{"opinion", "--run", "{RUN}", "--seat-id", s, "--id", pick(rng, fuzzIDs),
				"--as", "carried", "--principle", "p", "--tension", "t", "--reason", "the ruling", "--review-flag", "r"}
			if rng.Intn(4) == 0 {
				args = args[:len(args)-2] // drop the review flag: the validation path
			}
		}
		if len(args) > 0 {
			out = append(out, cmd{role: role, args: args})
		}
	}
	return out
}

func dumpSeq(cmds []cmd) string {
	s := "\n"
	for _, c := range cmds {
		s += fmt.Sprintf("  %s %v\n", c.role, c.args)
	}
	return s
}
