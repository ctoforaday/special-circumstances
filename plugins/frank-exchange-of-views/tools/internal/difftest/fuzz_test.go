package difftest

import (
	"fmt"
	"math/rand"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
)

// TestDifferentialFuzz drives RANDOM valid verb sequences through both
// implementations. The R2g plan names this "the highest-value test class for a
// port": hand-written scenarios encode what the author already thought of, while
// a generator reaches the interleavings nobody enumerated — a close before a
// mint, a regrade of a closed gap, a dispose naming a key that never existed, a
// seat that mints in two rounds without re-registering.
//
// The sequences are VALID-ISH by construction rather than valid: rejected
// commands are as interesting as accepted ones, because an error message is part
// of the contract a seat reads. Both sides must reject identically.
//
// The seed is fixed so a failure is reproducible; widen iterations locally when
// hunting.
func TestDifferentialFuzz(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node unavailable — the oracle cannot be driven; gate not run")
	}
	if testing.Short() {
		t.Skip("fuzz differential is slow (two processes per command); skipped under -short")
	}
	root := repoRoot(t)
	bin := buildBinary(t)

	const sequences = 12
	const maxLen = 14
	rng := rand.New(rand.NewSource(0x5EED))

	for i := 0; i < sequences; i++ {
		t.Run(fmt.Sprintf("seq%02d", i), func(t *testing.T) {
			cmds := generate(rng, maxLen)
			goDir, mjsDir := t.TempDir(), t.TempDir()
			goMap, mjsMap := newMapper(), newMapper()

			for j, c := range cmds {
				gi := runGo(bin, goDir, c)
				mi := runMjs(t, root, mjsDir, c)
				goMap.observe(filepath.Join(goDir, "records"))
				mjsMap.observe(filepath.Join(mjsDir, "records"))
				got := normalizeOutput(gi, goDir, goMap)
				want := normalizeOutput(mi, mjsDir, mjsMap)
				if got.code != want.code || got.stdout != want.stdout || got.stderr != want.stderr {
					t.Fatalf("cmd %d %v %v diverged\n go:  code=%d out=%q err=%q\n mjs: code=%d out=%q err=%q\nsequence so far: %s",
						j, c.role, c.args, got.code, got.stdout, got.stderr, want.code, want.stdout, want.stderr, dumpSeq(cmds[:j+1]))
				}
			}

			g, w := collect(t, goDir, goMap), collect(t, mjsDir, mjsMap)
			if !reflect.DeepEqual(g.events, w.events) {
				t.Errorf("event logs diverged\nsequence: %s\n go:  %s\n mjs: %s", dumpSeq(cmds), jsonDump(g.events), jsonDump(w.events))
			}
			for name, want := range w.renders {
				if g.renders[name] != want {
					t.Errorf("render %s diverged\nsequence: %s\n--- go ---\n%s\n--- oracle ---\n%s",
						name, dumpSeq(cmds), g.renders[name], want)
				}
			}
		})
	}
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
	seat := func(role string) (string, string) {
		var s string
		switch role {
		case "merge":
			s = pick(rng, []string{"red-merge-r1", "red-merge-r2"})
		case "lens":
			s = pick(rng, []string{"red-lens-r1-L1", "red-lens-r1-L5", "red-lens-r2-L6"})
		case "blue":
			s = pick(rng, []string{"blue-respond-r1", "blue-synthesize"})
		default:
			s = "judge-r1"
		}
		return role, s
	}
	for len(out) < n {
		role := pick(rng, []string{"merge", "merge", "merge", "lens", "lens", "blue", "bench"})
		_, s := seat(role)
		args := []string{}
		switch role {
		case "merge":
			switch rng.Intn(7) {
			case 0:
				args = []string{"register", "--run", "{RUN}", "--seat-id", s}
			case 1, 2:
				args = []string{"mint", "--run", "{RUN}", "--seat-id", s,
					"--class", pick(rng, fuzzClasses), "--check", "acceptance check",
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
					args = append(args, "--basis", "movement reason")
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
				args = []string{"finding", "--run", "{RUN}", "--seat-id", s,
					"--label", fmt.Sprintf("L%d-F%d", 1+rng.Intn(6), 1+rng.Intn(4)),
					"--severity", pick(rng, fuzzGrades), "--likelihood", pick(rng, fuzzGrades),
					"--impact", pick(rng, fuzzGrades), "--text", "a finding"}
			case 3:
				args = []string{"observe", "--run", "{RUN}", "--seat-id", s, "--text", "an observation"}
				if rng.Intn(2) == 0 {
					args = append(args, "--label", fmt.Sprintf("N%d", rng.Intn(3)))
				}
			}
		case "blue":
			switch rng.Intn(3) {
			case 0:
				args = []string{"revision", "--run", "{RUN}", "--seat-id", s, "--text", "revised", "--claim-count", fmt.Sprint(rng.Intn(300))}
			case 1:
				args = []string{"manifest-row", "--run", "{RUN}", "--seat-id", s, "--id", pick(rng, fuzzIDs), "--row", "checked"}
			case 2:
				args = []string{"dispute", "--run", "{RUN}", "--seat-id", s, "--id", pick(rng, fuzzIDs),
					"--dimension", "likelihood", "--proposed", pick(rng, fuzzGrades), "--evidence", "why"}
			}
		case "bench":
			args = []string{"opinion", "--run", "{RUN}", "--seat-id", s, "--gap-id", pick(rng, fuzzIDs),
				"--disposition", "carried", "--principle", "p", "--tension", "t", "--review-flag", "r"}
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
