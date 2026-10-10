package difftest

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
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
// enumerated — a close before its mint, a regrade of a closed gap, a mint whose
// --found-by names a finding that never existed, a bench ruling on a motion
// nobody filed.
//
// Sequences are valid-ISH by construction rather than valid: a refusal is as
// interesting as an acceptance, because the error text is part of the contract a
// seat reads, and it must be as reproducible as a success.
//
// BUT A SEQUENCE OF NOTHING BUT REFUSALS PROVES NOTHING, and that is what this test ran on for
// as long as the tree has been flat. The generator composed `feov-record merge mint …`; the role
// groups were gone, so every command it built exited 2 with `no command named "chair" exists`,
// the record was never written, and the two replays agreed perfectly on zero events. Over the
// fixed seed set every arm landed 0 times — chair register 0/8, a blue arm 0/7, merge
// spot-check 0/4 among them. The "every arm lands" subtest and the event floor below are what
// make that state a failure instead of a pass.
//
// rankTimestamps replaces each event's clock with its POSITION in the record's own order,
// which is the order recordsql returns.
//
// DROPPING the timestamp instead would also make the comparison pass, and would stop it
// noticing a REORDER — the one failure this schema exists to catch. Ranking keeps order
// under test while letting the absolute clock differ, which is the only part that may.
//
// It ranked DISTINCT CLOCK VALUES until CI caught it: two events landing inside one clock
// tick leave one fewer distinct instant, every later rank shifts, and two runs of the same
// commands disagree for no reason but speed. That made this determinism test itself
// non-deterministic — the exact property it exists to assert, broken in its own harness.
// Position in the merge order is stable however coarse the clock is, and a genuine
// reordering still shows up as a diff.
func rankTimestamps(evs []map[string]any) []map[string]any {
	out := make([]map[string]any, len(evs))
	for i, ev := range evs {
		c := make(map[string]any, len(ev))
		for key, v := range ev {
			c[key] = v
		}
		if _, ok := c["ts"]; ok {
			c["ts"] = i
		}
		out[i] = c
	}
	return out
}

const (
	fuzzSeed      = 0x5EED // fixed: a failure must be reproducible
	fuzzSequences = 12
	fuzzMaxLen    = 14
	// fuzzEventFloor is the fewest events the generated commands must write across the seed set,
	// setup and prologue excluded. Measured at 47 when set (it was 0 before the generator spoke
	// the flat tree); the floor sits under that so a reweighting does not trip it, and far enough
	// above zero that "the fuzz wrote nothing and passed" cannot recur.
	fuzzEventFloor = 20
)

func TestReplayDeterminism(t *testing.T) {
	if testing.Short() {
		t.Skip("determinism fuzz spawns two processes per command; skipped under -short")
	}
	bin := buildBinary(t)

	// EVERY sequence is generated before any runs. Generating inside the subtest drew from the
	// shared rng only for the sequences a -run filter let through, so `-run .../seq03` replayed
	// seq00's commands under seq03's name and a reported failure could not be reproduced by
	// filtering to it. Up front, seqNN is the same commands however the test is filtered.
	rng := rand.New(rand.NewSource(fuzzSeed))
	seqs := make([][]cmd, fuzzSequences)
	for i := range seqs {
		seqs[i] = generate(rng, fuzzMaxLen)
	}
	// THE ACT TOUR, after the random sequences so they are drawn exactly as before.
	seqs = append(seqs, actTour(rng))
	drawn, landed := map[string]int{}, map[string]int{}
	repeatRefused := 0             // tour repeats the record refused, each naming the act that answers the first
	refusal := map[string]string{} // the first refusal each arm met, for the guard's message
	written := 0
	ran := 0 // sequences whose body ran in this process; a -run filter skips the rest

	for i, cmds := range seqs {
		t.Run(fmt.Sprintf("seq%02d", i), func(t *testing.T) {
			ran++
			t.Logf("sequence:%s", dumpSeq(cmds))
			first := replay(t, bin, cmds)
			second := replay(t, bin, cmds)

			// Tallied from the FIRST replay: the same runGo invocation the comparison below
			// rests on, so "this arm landed" is a claim about the command the fuzz actually ran.
			for j, c := range cmds {
				drawn[c.arm]++
				if first.codes[j] == 0 {
					landed[c.arm]++
				} else if _, seen := refusal[c.arm]; !seen {
					refusal[c.arm] = firstLine(first.stderrs[j])
				}
			}
			written += first.written
			// A REPEAT THE RECORD REFUSES NAMES THE ACT THAT ANSWERS THE FIRST. The tour files each
			// act twice; where a sitting holds one of it the second is refused, and the refusal a
			// seat reads ends with its row (record's superseders).
			for j, c := range cmds {
				if strings.HasSuffix(c.arm, tourRepeat) && first.codes[j] != 0 && strings.Contains(first.stderrs[j], "has already recorded a") {
					repeatRefused++
					if !strings.Contains(first.stderrs[j], "The first stands. If it was wrong, say so in ") {
						t.Errorf("%s: the refusal of a repeated act names no answering act:\n%s", c.arm, firstLine(first.stderrs[j]))
					}
				}
			}

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

	// THE GUARD. Two replays that both refuse everything agree perfectly, so agreement alone is
	// not evidence the fuzz exercised anything. Every arm must be ACCEPTED at least once over the
	// seed set — exit 0 from the exact command runGo ran — and the accepted commands must have
	// written a non-trivial number of events. An arm that only ever refuses is a verb on the wrong
	// seat, a verb path the tree does not have, or a precondition the generator never builds.
	//
	// It measures the WHOLE seed set or nothing. Under a -run filter only some sequences run, so
	// the tallies are partial: judging them would fail every arm the filtered-out sequences would
	// have drawn (a false failure that teaches people to ignore the guard), and passing them would
	// claim a measurement that was never taken. It skips instead, and says why.
	t.Run("every arm lands", func(t *testing.T) {
		if ran < len(seqs) {
			t.Skipf("not measured: %d of %d sequences ran (filtered run)", ran, len(seqs))
		}
		// The tour's repeats reach the once-per-sitting refusals: a tour in which none was refused
		// has stopped filing the second act, and the check above then holds nothing.
		if repeatRefused < 6 {
			t.Errorf("the act tour's repeats were refused %d time(s) for a second act in the sitting, want at least 6 (position ×2, closing ×2, spot-check, regrade, manifest row)", repeatRefused)
		}
		for _, a := range append(append([]fuzzArm{}, fuzzArms...), tourArms...) {
			t.Logf("%-28s landed %2d/%-2d", a.name, landed[a.name], drawn[a.name])
			switch {
			case drawn[a.name] == 0:
				t.Errorf("arm %q was never drawn over the seed set — raise its weight, or it tests nothing", a.name)
			case landed[a.name] == 0:
				t.Errorf("arm %q never exited 0 in %d tries over the seed set: every command it builds is refused, "+
					"so it exercises one error path and no state. Check the verb path and seat against "+
					"`feov-record --seat-id <seat> <verb> --help`. First refusal: %s", a.name, drawn[a.name], refusal[a.name])
			}
		}
		t.Logf("events written by generated commands (setup excluded): %d", written)
		if written < fuzzEventFloor {
			t.Errorf("the generated commands wrote %d events across the seed set, under the floor of %d — "+
				"two replays of a run that records nothing agree trivially", written, fuzzEventFloor)
		}
	})
}

type replayResult struct {
	events  []map[string]any
	output  []string
	codes   []int    // exit code per command, in order
	stderrs []string // raw stderr per command, in order
	// written counts the events the generated commands added, the round-0 setup excluded.
	written int
	// setup is how many of events the round-0 setup and prologue wrote.
	setup int
}

func replay(t *testing.T, bin string, cmds []cmd) replayResult {
	t.Helper()
	// TmpRun, not t.TempDir: collect opens the record in this process, and the cached handle
	// must be released before cleanup removes the directory — or Windows cannot remove it.
	runDir := recordtest.TmpRun(t)
	m := newMapper()
	prepareRun(t, bin, runDir, m, nil)
	// A BOARD BEFORE THE FUZZ STARTS, as every sitting after the first has: GAP001 open, GAP002
	// before the bench as MOTION001 — the placeholders the prologue's own output is given below,
	// which is how a generated command names an id it cannot predict. Without it, every gap verb — close, regrade, carry, manifest-row, both
	// motions — lands only when a random mint happened to precede it on the right seat with the
	// right id, and a docket ruling only when a random filing did too. Over the fixed seed set
	// several of them never did; the bench's ruling was 0 in 10 with the gap alone.
	mintArgs := func(problem string) []string {
		return []string{"--run", "{RUN}", "--seat-id", "red-lens-evidence", "--class", "scope-creep",
			"--check-kind", "document", "--check", "re-read the section", "--severity", "medium",
			"--likelihood", "medium", "--impact", "medium", "--problem", problem}
	}
	prologue := []cmd{
		{verb: "register", args: []string{"--run", "{RUN}", "--seat-id", "red-lens-evidence"}},
		{verb: "mint", args: mintArgs("the open gap the fuzz starts from")},
		{verb: "mint", args: mintArgs("a contested claim awaiting a bench ruling")},
		{verb: "motion", args: []string{"docket", "file", "--run", "{RUN}", "--seat-id", "red-chair", "--id", "GAP002",
			"--reason", "contested, and not red's to close"}},
	}
	// A RECORDED PROOF, for a sequence with a reproduce in it — and only for one, so the sequences
	// that name none start from the board they always did. Its sha is read off the record.
	vars := map[string]string{}
	if namesVar(cmds, "{PROOF}") {
		if err := os.WriteFile(filepath.Join(runDir, "fuzz-proof.js"), []byte("console.log('fuzz proof');\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		prologue = append(prologue, cmd{verb: "prove", args: []string{"--run", "{RUN}", "--seat-id", "blue-respond",
			"--quote", "A claim sits under S2.", "--script", "{RUN}/fuzz-proof.js", "--reason", "the computation a lens re-runs"}})
	}
	for _, pre := range prologue {
		inv := runGo(bin, runDir, m, pre)
		if inv.code != 0 {
			t.Fatalf("fuzz prologue %s %v: exit %d\nstderr: %s", pre.verb, pre.args, inv.code, inv.stderr)
		}
		m.observe(filepath.Join(runDir, "records"))
		// READ FOR ITS IDS: each id the prologue prints takes its placeholder here, in print order.
		normalizeOutput(inv, runDir, m)
	}
	setup := collect(t, runDir, m).events
	setupEvents := len(setup)
	for _, ev := range setup {
		if p, ok := ev["proof"].(map[string]any); ok {
			vars["{PROOF}"], _ = p["proof_sha"].(string)
		}
	}
	res := replayResult{setup: setupEvents}
	for _, c := range cmds {
		c.args = fillVars(c.args, vars)
		inv := runGo(bin, runDir, m, c)
		envelopeVars(inv.stdout, vars)
		m.observe(filepath.Join(runDir, "records"))
		got := normalizeOutput(inv, runDir, m)
		res.output = append(res.output, fmt.Sprintf("exit=%d out=%q err=%q", got.code, got.stdout, got.stderr))
		res.codes = append(res.codes, inv.code)
		res.stderrs = append(res.stderrs, inv.stderr)
	}
	res.events = collect(t, runDir, m).events
	res.written = len(res.events) - setupEvents
	return res
}

// Each pool REPEATS its accepted values so an acceptance is the common case and a refusal the
// minority: a mint needs three grades and a class to pass, and with a third of the grades
// invalid it landed 3 times in 22 — too rarely to leave anything on the board for the verbs
// downstream of it. The refusal values stay in every pool; they are the error paths.
var (
	fuzzGrades = []string{"low", "low_medium", "medium", "medium_high", "high", "certain", "realized", "trivial",
		"low", "medium", "high", "medium", "low-medium", "bogus"} // the last two are refused
	fuzzClasses = []string{"scope-creep", "citation-drift", "scope-creep", "citation-drift", "propagation-incomplete"} // the last is not staged
	// GAP001 and GAP002 are on the board from the prologue, and MOTION001 before the bench; the
	// others name an id only once the sequence's own output has printed one, and until then reach
	// the tool as the id of their kind that names nothing, refused for not existing.
	fuzzGapIDs  = []string{"GAP001", "GAP001", "GAP001", "GAP002", "GAP003"}
	fuzzMotions = []string{"MOTION001", "MOTION001", "MOTION002", "MOTION003"}
	// Report text the seeded round-0 report carries, plus one line it does not (a mis-quote).
	fuzzQuotes = []string{"A claim sits under S2.", "A claim sits under S4.", "a sentence the report never says"}
	// Finding ids, which a mint's --found-by is checked against at the write, by the placeholder of
	// the sequence's first, second and third finding — so some name a recorded finding and some
	// name nothing.
	fuzzFoundBy = []string{"FINDING001", "FINDING002", "FINDING003", "L1,L5"}

	// The evidence lens twice: it originated GAP001, and only the originator may close or regrade a
	// gap — the other two lenses are where the wrong-originator refusal comes from.
	lensSeats = []string{"red-lens-evidence", "red-lens-evidence", "red-lens-logic", "red-lens-dark-side"}
	blueSeats = []string{"blue-respond", "blue-synthesize"}
)

func pick[T any](rng *rand.Rand, xs []T) T { return xs[rng.Intn(len(xs))] }

// fuzzArm is one kind of command the generator builds: a verb path, the seats that may type it,
// and the flags. The seat list is the TREE the verb lives on — `mint`, `close` and `regrade` are
// the lens's, `carry` and `spot-check` the chair's, the docket ruling the bench's — and a verb put
// on a seat whose tree lacks it is refused before it reaches the record, every time.
type fuzzArm struct {
	name   string
	weight int
	seats  []string
	verb   []string // the verb path, first word first: {"motion", "grade", "file"}
	flags  func(rng *rand.Rand) []string
	// steps, when set, builds SEVERAL commands for the arm's seat — a tour arm's prerequisites,
	// its act and the repeat of it. Only the step labelled with the arm's own name is what "every
	// arm lands" measures.
	steps func(rng *rand.Rand, seat string) []cmd
}

// fuzzArms is weighted toward the lens, since that is where validation, lineage, and id minting
// live now that the lens mints and closes its own gaps.
var fuzzArms = []fuzzArm{
	{name: "chair register", weight: 1, seats: []string{"red-chair"}, verb: []string{"register"},
		flags: func(*rand.Rand) []string { return nil }},
	{name: "lens register", weight: 1, seats: lensSeats, verb: []string{"register"},
		flags: func(*rand.Rand) []string { return nil }},
	{name: "lens finding", weight: 3, seats: lensSeats, verb: []string{"finding"},
		flags: func(rng *rand.Rand) []string {
			// --key from a small space so the SAME seat sometimes repeats it — that exercises
			// the crash-retry idempotency (a repeated key returns the existing tool-assigned
			// id, no duplicate).
			f := []string{"--key", fmt.Sprintf("F%d", 1+rng.Intn(3)),
				"--severity", pick(rng, fuzzGrades), "--likelihood", pick(rng, fuzzGrades),
				"--impact", pick(rng, fuzzGrades), "--reason", "a finding"}
			if rng.Intn(4) > 0 {
				f = append(f, "--quote", pick(rng, fuzzQuotes))
			}
			return f
		}},
	{name: "lens mint", weight: 4, seats: lensSeats, verb: []string{"mint"},
		flags: func(rng *rand.Rand) []string {
			f := []string{"--class", pick(rng, fuzzClasses), "--check-kind", "document", "--check", "acceptance check",
				"--severity", pick(rng, fuzzGrades), "--likelihood", pick(rng, fuzzGrades),
				"--impact", pick(rng, fuzzGrades), "--problem", fmt.Sprintf("problem %d", rng.Intn(1000))}
			// GAP001 is on the board from the prologue, so naming it distinct drives the field the mint's
			// duplicate screen is answered with.
			if rng.Intn(4) == 0 {
				f = append(f, "--distinct-from", "GAP001")
			}
			if rng.Intn(3) == 0 {
				f = append(f, "--complexity", pick(rng, fuzzGrades))
			}
			if rng.Intn(3) == 0 {
				f = append(f, "--supersedes", pick(rng, fuzzGapIDs))
			}
			if rng.Intn(4) == 0 {
				f = append(f, "--key", fmt.Sprintf("K%d", rng.Intn(3)))
			}
			if rng.Intn(5) == 0 {
				f = append(f, "--found-by", pick(rng, fuzzFoundBy))
			}
			return f
		}},
	{name: "lens close", weight: 2, seats: lensSeats, verb: []string{"close"},
		flags: func(rng *rand.Rand) []string {
			f := []string{"--id", pick(rng, fuzzGapIDs),
				"--verified-by", "L1", "--verified-with", "Read", "--verified-against", "report.md#S1"}
			if rng.Intn(4) > 0 {
				f = append(f, "--reason", "re-read the repaired sentence")
			}
			if rng.Intn(4) == 0 {
				f = append(f, "--as", "repaired_with_regression") // no successor: the validation path
			}
			return f
		}},
	{name: "lens regrade", weight: 2, seats: lensSeats, verb: []string{"regrade"},
		flags: func(rng *rand.Rand) []string {
			f := []string{"--id", pick(rng, fuzzGapIDs), "--severity", pick(rng, fuzzGrades)}
			if rng.Intn(4) > 0 {
				f = append(f, "--reason", "movement reason")
			}
			return f
		}},
	// Weight 2: a carry lands only after a closure, and over the fixed seed set a single weight drew it
	// twice, both before any gap had closed.
	{name: "chair carry", weight: 2, seats: []string{"red-chair"}, verb: []string{"carry"},
		flags: func(rng *rand.Rand) []string {
			f := []string{"--id", pick(rng, fuzzGapIDs), "--carried-from", "1"}
			if rng.Intn(4) == 0 {
				f = append(f, "--as", "repaired_with_regression")
			}
			return f
		}},
	{name: "chair spot-check", weight: 1, seats: []string{"red-chair"}, verb: []string{"spot-check"},
		flags: func(rng *rand.Rand) []string {
			if rng.Intn(2) == 0 {
				return []string{"--none", "--reason", "nothing archived yet"}
			}
			return []string{"--ids", "GAP001,GAP002", "--reason", "re-read both closures"}
		}},
	{name: "chair motion docket file", weight: 2, seats: []string{"red-chair"}, verb: []string{"motion", "docket", "file"},
		flags: func(rng *rand.Rand) []string {
			return []string{"--id", pick(rng, fuzzGapIDs), "--reason", "contested, and not red's to close"}
		}},
	{name: "blue manifest-row", weight: 1, seats: blueSeats, verb: []string{"manifest-row"},
		flags: func(rng *rand.Rand) []string { return []string{"--id", pick(rng, fuzzGapIDs), "--reason", "checked"} }},
	// Was `blue dispute`, a verb the motion collapse retired: a grade dispute is a grade motion.
	{name: "blue motion grade file", weight: 1, seats: blueSeats, verb: []string{"motion", "grade", "file"},
		flags: func(rng *rand.Rand) []string {
			return []string{"--id", pick(rng, fuzzGapIDs), "--dimension", "likelihood",
				"--proposed", pick(rng, fuzzGrades), "--reason", "why"}
		}},
	// Was `bench reason`: the bench's disposition of a gap is its ruling on a docket motion.
	{name: "bench motion docket rule", weight: 2, seats: []string{"judge"}, verb: []string{"motion", "docket", "rule"},
		flags: func(rng *rand.Rand) []string {
			f := []string{"--id", pick(rng, fuzzMotions), "--as", "remanded", "--principle", "p", "--tension", "t",
				"--reason", "the ruling", "--review-flag", "r", "--settled", "the proposition this ruling bars"}
			if rng.Intn(4) > 0 {
				// omitted: no --reopens-on, the validation path. A remand states its direction.
				f = append(f, "--reopens-on", "a reproduction on the shipped binary")
			}
			return f
		}},
}

// tourArms are the recording verbs the random sequences rarely land, each with the prerequisites
// it needs: one per verb, on a seat whose tree has it — the lens's close, regrade and reproduce;
// blue's and the chair's position and closing; the bench's declare, certify, halt and outcome.
// Every arm files its act and then files it AGAIN: where a sitting holds one of the act the repeat
// is refused, and the refusal names the act that answers the first.
//
// KEPT OUT OF fuzzArms, so the random sequences are drawn exactly as they were measured. They run
// as ONE extra sequence, the act tour, in which each arm is drawn once, in this order: the acts on
// GAP001 before the closure that ends it, the outcome before the halt that would make it
// derivable.
var tourArms = []fuzzArm{
	tourArm("blue manifest-row", blueSeats, []string{"manifest-row"}, func(t string) []string { return []string{"--id", "GAP001", "--reason", t} }, nil),
	tourArm("blue cite", []string{"blue-respond"}, []string{"cite"},
		func(t string) []string {
			return []string{"--quote", "A claim sits under S2.", "--url", fuzzSourceURL("/corr"), "--title", t}
		}, nil),
	tourArm("blue prove", []string{"blue-respond"}, []string{"prove"},
		func(t string) []string {
			return []string{"--quote", "A claim sits under S2.", "--script", "{RUN}/fuzz-proof.js", "--reason", t}
		}, nil),
	// blue-respond alone: a position is refused for a blue seat that owes none (record.SeatOwesPosition).
	tourArm("blue position", []string{"blue-respond"}, []string{"position"}, reasonOnly, nil),
	tourArm("chair position", []string{"red-chair"}, []string{"position"}, reasonOnly, nil),
	tourArm("blue closing", blueSeats, []string{"closing"}, func(t string) []string { return []string{"--id", "GAP001", "--reason", t} }, nil),
	tourArm("chair closing", []string{"red-chair"}, []string{"closing"}, func(t string) []string { return []string{"--id", "GAP001", "--reason", t} }, nil),
	tourArm("log", []string{"red-lens-evidence", "red-chair", "blue-respond", "judge"}, []string{"log"},
		func(t string) []string { return []string{"--type", "defect", "--reason", t} }, nil),
	tourArm("chair avenue review", []string{"red-chair"}, []string{"avenue", "review"}, reasonOnly, nil),
	tourArm("chair spot-check", []string{"red-chair"}, []string{"spot-check"}, func(t string) []string { return []string{"--none", "--reason", t} }, nil),
	tourArm("lens regrade", []string{"red-lens-evidence"}, []string{"regrade"},
		func(t string) []string { return []string{"--id", "GAP001", "--severity", "high", "--reason", t} }, nil),
	tourArm("chair motion grade rule", []string{"red-chair"}, []string{"motion", "grade", "rule"},
		func(t string) []string { return []string{"--id", "{MOTION}", "--as", "rejected", "--reason", t} },
		func(name string) []cmd { return []cmd{fileGradeMotion(name)} }),
	tourArm("bench motion docket rule", []string{"judge"}, []string{"motion", "docket", "rule"},
		func(t string) []string {
			return []string{"--id", "MOTION001", "--as", "remanded", "--principle", "p", "--tension", "t", "--review-flag", "r",
				"--settled", "the proposition this ruling bars", "--reopens-on", "a reproduction on the shipped binary", "--reason", t}
		}, nil),
	tourArm("blue avenue propose", blueSeats, []string{"avenue", "propose"},
		func(t string) []string { return []string{"--reason", t, "--hypothesis", "it would settle something"} }, nil),
	tourArm("lens reproduce", []string{"red-lens-evidence"}, []string{"reproduce"},
		func(t string) []string { return []string{"--id", "{PROOF}", "--as", "sound", "--reason", t} }, nil),
	tourArm("bench declare", []string{"judge"}, []string{"declare"}, reasonOnly, nil),
	tourArm("bench certify", []string{"judge"}, []string{"certify"}, reasonOnly, nil),
	tourArm("bench outcome", []string{"judge"}, []string{"outcome"},
		func(t string) []string { return []string{"--as", "UNVERIFIED", "--reason", t} }, nil),
	tourArm("lens close", []string{"red-lens-evidence"}, []string{"close"},
		func(t string) []string {
			return []string{"--id", "GAP001", "--verified-by", "L1", "--verified-with", "Read", "--verified-against", "report.md#S1", "--reason", t}
		}, nil),
	tourArm("bench halt", []string{"judge"}, []string{"halt"}, reasonOnly, nil),
}

func reasonOnly(t string) []string { return []string{"--reason", t} }

func fileGradeMotion(name string) cmd {
	return fuzzStep(name+" · setup", []string{"motion", "grade", "file"}, "blue-respond",
		"--id", "GAP001", "--dimension", "severity", "--proposed", "low", "--reason", "the consequence is bounded", "--json")
}

// fuzzStep is one command, spelled as its seat types it.
func fuzzStep(arm string, verb []string, seat string, flags ...string) cmd {
	args := append(append([]string{}, verb[1:]...), "--run", "{RUN}", "--seat-id", seat)
	return cmd{verb: verb[0], args: append(args, flags...), arm: arm}
}

// tourRepeat labels the second filing of a tour arm's act.
const tourRepeat = " · repeat"

// tourArm is a tour arm: `before` builds its prerequisites, then the act, then the act again with
// other wording.
func tourArm(name string, seats, verb []string, flagsFor func(text string) []string, before func(name string) []cmd) fuzzArm {
	return fuzzArm{name: name, weight: 1, seats: seats, verb: verb, steps: func(rng *rand.Rand, seat string) []cmd {
		var out []cmd
		if before != nil {
			out = before(name)
		}
		return append(out,
			fuzzStep(name, verb, seat, flagsFor("as first recorded")...),
			fuzzStep(name+tourRepeat, verb, seat, flagsFor("the first was wrong, and this says what holds")...))
	}}
}

// actTour is one sequence holding every tour arm once.
func actTour(rng *rand.Rand) []cmd {
	var out []cmd
	for _, a := range tourArms {
		out = append(out, a.steps(rng, pick(rng, a.seats))...)
	}
	return out
}

// namesVar reports whether any command in a sequence names a placeholder.
func namesVar(cmds []cmd, v string) bool {
	for _, c := range cmds {
		for _, a := range c.args {
			if strings.Contains(a, v) {
				return true
			}
		}
	}
	return false
}

// fillVars substitutes the placeholders a correction arm names — {KEY}, {MOTION}, {PROOF} — into a

// fillVars substitutes the placeholders a tour arm names — {MOTION}, {PROOF} — into a COPY of the
// arguments, since both replays of a sequence share its commands.
func fillVars(args []string, vars map[string]string) []string {
	out := make([]string, len(args))
	for i, a := range args {
		for k, v := range vars {
			a = strings.ReplaceAll(a, k, v)
		}
		a = strings.ReplaceAll(a, "{MOTION}", "")
		out[i] = a
	}
	return out
}

// envelopeVars reads the placeholders off a --json envelope: a filing's `motion_id`. Output that is
// not an envelope changes nothing.
func envelopeVars(stdout string, vars map[string]string) {
	var env struct {
		OK     *bool `json:"ok"`
		Result struct {
			MotionID string `json:"motion_id"`
		} `json:"result"`
	}
	if json.Unmarshal([]byte(strings.TrimSpace(stdout)), &env) != nil || env.OK == nil {
		return
	}
	if env.Result.MotionID != "" {
		vars["{MOTION}"] = env.Result.MotionID
	}
}

// generate builds one sequence from fuzzArms, each command spelled exactly as its seat types it:
// `feov-record <verb path> --run <run> --seat-id <seat> <flags…>`.
func generate(rng *rand.Rand, maxLen int) []cmd {
	total := 0
	for _, a := range fuzzArms {
		total += a.weight
	}
	n := 3 + rng.Intn(maxLen-3)
	out := make([]cmd, 0, n)
	for len(out) < n {
		r := rng.Intn(total)
		var a fuzzArm
		for _, a = range fuzzArms {
			if r < a.weight {
				break
			}
			r -= a.weight
		}
		args := append(append([]string{}, a.verb[1:]...), "--run", "{RUN}", "--seat-id", pick(rng, a.seats))
		out = append(out, cmd{verb: a.verb[0], args: append(args, a.flags(rng)...), arm: a.name})
	}
	return out
}

func firstLine(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}

func dumpSeq(cmds []cmd) string {
	var b strings.Builder
	b.WriteString("\n")
	for _, c := range cmds {
		fmt.Fprintf(&b, "  [%s] feov-record %s %s\n", c.arm, c.verb, strings.Join(c.args, " "))
	}
	return b.String()
}
