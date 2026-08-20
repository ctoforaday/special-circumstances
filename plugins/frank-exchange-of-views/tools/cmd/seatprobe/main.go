// Command seatprobe builds a board, dispatches a real seat at it, and reports what the seat chose.
//
// # The question
//
// Every other harness in this repo asks whether the tool ACCEPTS what a seat sends. This one asks
// the opposite, and it is the more expensive failure: GIVEN THE VERBS IT HAS, WHICH DOES A SEAT
// CHOOSE? A seat that had the right verb, did not use it, and wrote something plausible instead
// leaves no trace anywhere — the run simply does without a capability and nothing reports it.
//
// # Why the CLI and not the Agent SDK
//
// The SDK would work. It is also a dependency, and this repo has none — no package.json, no
// pyproject.toml, nothing outside Go modules. The `claude` CLI gives everything the harness needs
// and is already authenticated:
//
//	--print --output-format stream-json --verbose   the full trajectory, as events
//	--model haiku                                   the instrument (see below)
//	--system-prompt-file <constitution>             the seat's real constitution
//	--allowedTools "Bash Read Write Edit Grep Glob" the tools a seat actually has
//	--add-dir <runDir>                              access to the board
//
// AND IT CARRIES THINKING. Verified rather than assumed: a stream-json run emits `thinking` blocks
// alongside `text` and `tool_use`, so the harness can log what the seat was reasoning about when
// it chose a verb — or when it chose prose instead. That was the one capability I could not get by
// dispatching seats by hand, and it is the reason this exists rather than a shell loop.
//
// # Why a weak model
//
// A strong seat compensates for a bad help string by inferring what was meant, which hides the
// defect being hunted: the constitution and the `--help` text are the only things standing between
// a seat and a hand-written artifact, and their quality is only visible when the reader is not
// clever enough to paper over it. Haiku is the instrument, not the subject.
//
// # The record leaves the run directory
//
// By default each board's event record is written OUTSIDE the run, so the only way to the board
// is `show --view <name>`. Without that the measurement is uninterpretable: the first dispatch
// used 5 of 18 verbs, and the trajectory showed why — the seat never called `show` at all, it
// ran `ls` and parsed records/*.jsonl itself. "The surface failed to teach" and "the seat found
// a shortcut and never needed teaching" produce the same number and want opposite fixes.
//
// It is not a sandbox (see internal/record/recordroot.go) — Bash reaches anything. It moves the
// record off the CHEAP path, which is sized to the measured failure: satisficing, not evasion.
// `-records-in-run` restores the old layout as a control arm.
//
// # It reports; it does not gate
//
// Agent behaviour is not deterministic and a flaky gate is one the next person turns off. The
// output is a choice report and a friction corpus addressed to a human (#363). What IS a gate
// lives in internal/seatprobe's tests: that every verb has a board demanding it, and that every
// board carries the state its expectations need.
//
// # Usage
//
//	go run ./cmd/seatprobe -bin <feov-record> -board arithmetic -dir <scratch>
//	go run ./cmd/seatprobe -bin <feov-record> -board all -dir <scratch> -parallel 3
//	go run ./cmd/seatprobe -board arithmetic -dir <scratch> -report-only
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatenv"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatprobe"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/setup"
)

func main() {
	var (
		bin        = flag.String("bin", "", "path to the feov-record binary the board is built with")
		board      = flag.String("board", "arithmetic", "board name, or `all`")
		dir        = flag.String("dir", "", "directory to build boards under (REQUIRED)")
		constDir   = flag.String("constitutions", "", "directory holding the seat constitutions (default: the plugin's agents/)")
		model      = flag.String("model", "haiku", "model for the dispatched seat")
		parallel   = flag.Int("parallel", 2, "how many boards to run at once")
		reportOnly = flag.Bool("report-only", false, "skip build and dispatch; report on what is already there")
		keep       = flag.Bool("keep", false, "keep an existing board directory instead of rebuilding it")
		ask        = flag.Bool("ask", false, "do not dispatch a seat to ACT — ask it to ENUMERATE and ASSESS its options instead. A verb used zero times cannot say whether the seat never perceived it, weighed it and declined, or wanted it and could not reach it; this asks")
		inRun      = flag.Bool("records-in-run", false, "leave the event record under the run directory, where the seat can read it without the tool — the CONTROL arm, for measuring what the separation changes")
		patterns   = flag.String("patterns", "none", "red's gap-pattern memory: `none`, `file` (staged at inputs/red-gap-patterns.md — the MOUNTED FILE form), or `duty` (staged AND named in the prompt, selected by the classes of this board's gaps — the DUTY form)")
	)
	flag.Parse()

	// A TYPO HERE WOULD RUN THE SHIPPED ARM AND REPORT THE ONE YOU ASKED FOR. record's own
	// resolver falls back to shipped on purpose (an unrecognised value must not empty a real
	// seat's work list), so the probe validates the spelling itself rather than inheriting a
	// fallback that is correct for production and silent for an experiment.
	switch *patterns {
	case "none", "file", "duty":
	default:
		fail("no patterns arm %q — one of none, file, duty", *patterns)
	}

	if *dir == "" {
		fail("-dir is required: seatprobe writes real run directories and will not guess where")
	}
	boards := seatprobe.Boards()
	var names []string
	if *board == "all" {
		for n := range boards {
			names = append(names, n)
		}
		sort.Strings(names)
	} else {
		if _, ok := boards[*board]; !ok {
			var have []string
			for n := range boards {
				have = append(have, n)
			}
			sort.Strings(have)
			fail("no board %q — one of %s, or `all`", *board, strings.Join(have, ", "))
		}
		names = []string{*board}
	}
	if !*reportOnly && *bin == "" {
		fail("-bin is required unless -report-only: the board is built through the REAL binary, because a board built any other way can hold a state no seat could reach")
	}

	surface := seatprobe.NewSurface(cli.CommandPaths())
	results := make([]string, len(names))
	// A BOARD THAT NEVER DISPATCHED IS NOT A RESULT, and this used to be invisible: every board
	// could fail — a wrong -constitutions path fails all of them at once — and the command still
	// printed a report and exited 0. A caller scoring that run reads an empty report as a quiet
	// one, which is the plausible zero this repository keeps finding, arriving in the instrument.
	failed := make([]bool, len(names))
	sem := make(chan struct{}, max(1, *parallel))
	var wg sync.WaitGroup
	for i, name := range names {
		wg.Add(1)
		go func(i int, name string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			out, err := probe(boards[name], filepath.Join(*dir, name), *bin, *constDir, *model, *reportOnly, *keep, *inRun, *ask, surface, *patterns)
			if err != nil {
				results[i] = fmt.Sprintf("## %s — FAILED\n\n%v\n", name, err)
				failed[i] = true
				return
			}
			results[i] = out
		}(i, name)
	}
	wg.Wait()

	fmt.Println("# Seat probe")
	fmt.Println()
	fmt.Printf("%d board(s), model %s, the SHIPPED constitution. What each seat CHOSE, of what its role offers.\n\n", len(names), *model)
	for _, r := range results {
		fmt.Println(r)
	}

	n := 0
	for _, f := range failed {
		if f {
			n++
		}
	}
	if n > 0 {
		fmt.Fprintf(os.Stderr, "\nseatprobe: %d of %d board(s) FAILED TO DISPATCH — this run is not a result.\n"+
			"Scoring it would read an empty report as a quiet one. Fix the dispatch and re-run.\n", n, len(names))
		os.Exit(1)
	}
}

// namingTreatment reports how many distinct verb names the arm actually moved, and says so in the
// report beside the arm that claims to have moved them.
//
// It answers a question the arm name cannot: `-naming none` is a LABEL, and the treatment behind
// it is text manipulation that the constitutions are free to drift away from. Reporting the arm
// without the count is reporting the intention.
//
// THE ARMS DO NOT ALL POINT THE SAME WAY, and this function used to assume they did. When the
// constitutions named verbs, every arm was SUBTRACTIVE — `none` removed all, `partial` left a
// few — so `before == 0` could only mean the fixture had nothing to work on, and saying so first,
// before looking at the arm, was right. Once the constitutions stopped naming verbs (the strip),
// `before == 0` became true of EVERY run, while `partial` and `complete` turned ADDITIVE: they
// append names that were never there. The guard then fired on every arm and told the reader that
// a null result in `complete` was a fixture artifact — on the one arm whose treatment was real,
// and in place of the count that would have shown it.
//
// So the direction is taken from the arm, and the count is printed either way: a sentence about
// what the number means is worth nothing if it can be right about the wrong number.
//
// A read failure is NOT MEASURED rather than zero: "0 names survived" and "I could not open the
// constitution" are different facts, and only one of them means the treatment worked.
func namingTreatment(role, constDir string, sf seatprobe.Surface) string {
	src, err := constitutionFor(role, constDir)
	if err != nil {
		return "NOT MEASURED (no constitution for " + role + ": " + err.Error() + ")"
	}
	b, err := os.ReadFile(src)
	if err != nil {
		return "NOT MEASURED (cannot read " + src + ": " + err.Error() + ")"
	}
	named := seatprobe.NamesSurviving(string(b), sf)
	if len(named) == 0 {
		return fmt.Sprintf("the shipped %s constitution names 0 verbs — which is the claim the strip rests on, checked here rather than assumed", role)
	}
	// A SURVIVING NAME IS A DEFECT, NOT AN ARM. There were three arms that varied this on purpose;
	// there are none now, so a constitution that names a verb is a regression in the shipped
	// bytes and the run says so rather than averaging it into a finding.
	var names []string
	for v := range named {
		names = append(names, v)
	}
	sort.Strings(names)
	return fmt.Sprintf("THE SHIPPED %s CONSTITUTION NAMES %d VERB(S): %s — the strip has regressed, and a seat handed a partial list stops looking",
		strings.ToUpper(role), len(names), strings.Join(names, ", "))
}

// trajectoryPath keeps the capture OUT of the run directory.
//
// It used to be written to <runDir>/probe-trajectory.jsonl, where the seat can read it — and one
// did: the docket seat found the harness's own recording of the docket seat and tried to execute
// it, producing 149 "command not found" lines in a single turn. A probe that leaves its
// instrument inside the thing it measures is measuring itself.
func trajectoryPath(runDir string) string {
	return filepath.Join(filepath.Dir(runDir), ".probe", filepath.Base(runDir)+".jsonl")
}

func probe(b seatprobe.Board, runDir, bin, constDir, model string, reportOnly, keep, recordsInRun, ask bool, surface seatprobe.Surface, patterns string) (string, error) {
	recordRoot := ""
	if !reportOnly {
		if !keep {
			if err := os.RemoveAll(runDir); err != nil {
				return "", err
			}
		}
		if err := os.MkdirAll(runDir, 0o755); err != nil {
			return "", err
		}
		// THE RECORD LEAVES THE RUN DIRECTORY, and that is what makes the probe's number mean
		// something. On the first dispatch the seat never called `show` — it ran `ls` and parsed
		// records/*.jsonl directly, so "used 5 of 18 verbs" was uninterpretable: it could mean the
		// surface failed to teach, or that the seat found a shortcut and never needed teaching.
		// Those want opposite fixes. With the record elsewhere, a seat that cannot find a verb has
		// two exits — find it, or file friction — and both of those are signal.
		//
		// A TEMP ROOT, NOT A SIBLING OF THE RUN. `ls ..` is one keystroke.
		if !recordsInRun {
			r, err := os.MkdirTemp("", "feov-records-")
			if err != nil {
				return "", err
			}
			recordRoot = r
			// THE RECORD ROOT IS NOT DELETED. It was, and the first separated run paid for it:
			// every post-hoc question ("did any seat file friction?") hit the resolver's own
			// dangling-pointer refusal, because the evidence had been removed while the pointer
			// binding the run to it survived. An instrument that destroys its own measurement on
			// the way out is one you can only ever read once. It is a temp directory; the OS
			// reclaims it, and the report prints where it went.
		}
		run := func(args ...string) (string, error) {
			cmd := exec.Command(bin, args...)
			// Declared on the BUILD calls only. By the time the seat runs, the run is bound to
			// its root by a pointer, so the seat's own process carries nothing — `env` in its
			// Bash session names no record path.
			if recordRoot != "" {
				cmd.Env = append(os.Environ(), record.RecordRootEnv+"="+recordRoot)
			}
			out, err := cmd.CombinedOutput()
			if err != nil {
				return string(out), fmt.Errorf("%s: %s", strings.Join(args[:min(3, len(args))], " "), strings.TrimSpace(string(out)))
			}
			return string(out), nil
		}
		if err := seatprobe.Build(runDir, b, run); err != nil {
			return "", fmt.Errorf("build: %w", err)
		}
		if patterns != "none" {
			// THE MOUNTED FILE, exactly as run-setup stages it for a real run.
			if r := setup.MirrorGapPatterns(memoryDirs(), runDir); !r.Written {
				return "", fmt.Errorf("patterns arm %q: the corpus did not stage (%s) — a run that reports on memory it never delivered is the defect this arm exists to test", patterns, r.Reason)
			}
		}
		if err := dispatch(b, runDir, bin, constDir, model, ask, surface, patterns, b); err != nil {
			return "", fmt.Errorf("dispatch: %w", err)
		}
	}

	// IN THE ELICITATION ARM THE ANSWER IS THE DELIVERABLE. A choice report over a sitting that
	// deliberately recorded nothing would print "reached for 0 of 18" and mean nothing at all —
	// the plausible zero again, manufactured by the instrument.
	if ask {
		out := "# " + b.Name + " — what the " + b.Seat + " seat thinks its options are\n\n" +
			readAnswer(trajectoryPath(runDir))
		if t := readThinking(trajectoryPath(runDir)); t != "" {
			out += "\n\n### What it was reasoning about\n\n" + t + "\n"
		}
		return out, nil
	}

	// The record answers what LANDED; the trajectory answers what the seat REACHED FOR. Reads
	// and refusals live only in the second, and counting from the record alone reported both as
	// verbs the seat never touched.
	attempts := map[string]map[string]int{}
	role := ""
	for _, s := range seatprobe.Seats {
		if s.ID == b.Seat {
			role = s.Role
		}
	}
	if a, err := seatprobe.Attempted(trajectoryPath(runDir), filepath.Base(bin), surface, role); err == nil {
		attempts[b.Seat] = a
	} else {
		attempts = nil // NOT MEASURED, and the report says so rather than counting zero
	}
	report, err := seatprobe.Report(surface, runDir, []string{b.Seat}, b.Expect, attempts)
	if err != nil {
		return "", err
	}
	// THE ARM, ON THE RESULT. A choice report that does not say which treatment produced it is a
	// number waiting to be compared against a number from a different condition — which is how the
	// "4 of 14" figure came to be cited as a fact about seats rather than about one arm.
	// THE TREATMENT, MEASURED RATHER THAN ASSUMED. naming.go's type doc says this count "is
	// printed with the result" and nothing printed it, so an arm whose redactor had stopped
	// matching would produce a `none` run byte-identical to `partial`, both arms would report the
	// same behaviour, and the experiment would conclude "naming does not matter" — a null result
	// manufactured by the instrument and wearing the clothes of a finding.
	report += "**naming treatment**: " + namingTreatment(role, constDir, surface) + "\n"
	if hu, err := seatprobe.ReadHelpUse(trajectoryPath(runDir), filepath.Base(bin)); err == nil {
		report += fmt.Sprintf("**help use**: %s\n", hu.Line())
	} else {
		report += "**help use**: NOT MEASURED (no trajectory)\n"
	}
	// THE THIRD CHANNEL, BESIDE THE NAMING AND THE HELP. record.SittingOf hands a seat the
	// situation AND the verb that discharges it, and it rides on `show` alone. Measured across
	// the first naming matrix, work-list reads varied THREEFOLD with the arm — so a naming effect
	// reported without this number has a rival explanation sitting underneath it, and the first
	// version of that report was published before anyone asked.
	if vr, err := seatprobe.ReadViewReads(trajectoryPath(runDir), filepath.Base(bin)); err == nil {
		report += fmt.Sprintf("**duty delivery**: %s\n", vr.Line())
	} else {
		report += "**duty delivery**: NOT MEASURED (no trajectory)\n"
	}
	// The reasoning, where the seat produced any. It is the half a record cannot hold: the record
	// says which verb was taken, and this says what the seat was weighing when it chose.
	if recordRoot != "" {
		report += fmt.Sprintf("\n_record kept at %s — query the board with `show --run %s --view <name>`_\n", recordRoot, runDir)
	}
	if t := readThinking(trajectoryPath(runDir)); t != "" {
		report += "\n### What the seat was reasoning about\n\n" + t + "\n"
	}
	return report, nil
}

// dispatch runs one seat at the board through the `claude` CLI.
func dispatch(b seatprobe.Board, runDir, bin, constDir, model string, ask bool, sf seatprobe.Surface, patterns string, board seatprobe.Board) error {
	role := ""
	for _, s := range seatprobe.Seats {
		if s.ID == b.Seat {
			role = s.Role
		}
	}
	if role == "" {
		return fmt.Errorf("board %q names seat %q, which is not one this harness registers", b.Name, b.Seat)
	}
	constitution, err := constitutionFor(role, constDir)
	if err != nil {
		return err
	}
	// THE ARM IS APPLIED TO A COPY, NEVER TO THE SHIPPED FILE. An experiment that edits the
	// artifact it measures cannot be run a second time, and the second run is the one that says
	// whether the first was variance.
	constitution, err = armConstitution(constitution, runDir, sf, role)
	if err != nil {
		return err
	}

	prompt := fmt.Sprintf(`You are the %s seat in a frank-exchange-of-views run. Your seat id is %s.

Run directory: %s — use this ABSOLUTE path when you read files under it.
The record tool is %s. Your RUN is injected — never pass --run. Your seat id is %s and you state it
ONCE: "register" is your first act, it binds that id to you on the record, and every call after it
resolves your seat for you, so you pass --seat-id at register and never again.

Read the board and the artifact under audit, then do your sitting's work. Decide and act; do not ask me what to do.

%s`, role, b.Seat, runDir, bin, b.Seat, b.Sitting())

	// THE ELICITATION ARM. Same board, same constitution, same identity — the seat is asked what
	// it thinks its options are instead of being watched to see which it takes.
	//
	// The tool set drops Write and Edit, and the prompt says to record nothing. Not because a
	// seat would cheat, but because an answer about judgement should not leave board state
	// behind: the run directory has to stay comparable to the acting arm's, or the two probes
	// are measuring different boards.
	tools := "Bash Read Write Edit Grep Glob"
	if ask {
		prompt = seatprobe.ElicitPrompt(role, b.Seat, runDir, bin, b)
		if patterns == "duty" {
			prompt += patternDuty(board)
		}
		tools = "Bash Read Grep Glob"
	}

	args := []string{
		"-p", prompt,
		"--model", model,
		"--output-format", "stream-json",
		"--verbose",
		"--system-prompt-file", constitution,
		"--add-dir", runDir,
		"--allowedTools", tools,
		"--max-turns", "60",
	}
	cmd := exec.Command("claude", args...)
	cmd.Dir = runDir
	// THE RUN IS INJECTED BECAUSE PRODUCTION INJECTS IT. The PreToolUse hook prefixes a seat's
	// feov-record calls with FEOV_RUN (#281), and the probe does not load the plugin's hooks —
	// so without this it told seats to type an absolute path on every call: a HARDER surface
	// than any real run presents, and every mistyped path it measured was friction production
	// had already designed away.
	//
	// THE IDENTITY IS INJECTED AS AN AGENT HANDLE, which is what production injects — and it
	// arrives UNBOUND. Build registers the fixture's seats as the harness, without this handle, so
	// nothing on the record ties it to a seat until the seat itself calls `register`.
	//
	// THAT IS THE POINT, AND IT USED TO BE THE DIVERGENCE. Build bound the handle in advance, so a
	// seat arrived already registered and `register` stopped being its first act. The 2026-08-20
	// run measured the cost: one seat in nine never called it, made 22 tool calls, and recorded
	// events anyway — a first write production would have refused. An instrument that satisfies
	// the guard it is measuring cannot tell a compliant seat from an untested guard.
	//
	// What remains uncontrolled is smaller and is stated here rather than left to be discovered:
	// production's dispatcher never learns an agent handle at all (Workflow's agent() returns a
	// result, not one), so the handle a production seat carries is minted by the hook rather than
	// by the caller. The BINDING path is now identical; only the handle's provenance differs.
	cmd.Env = append(os.Environ(),
		seatenv.Var+"="+runDir,
		seatenv.AgentVar+"="+seatprobe.ProbeAgentID(b.Seat),
		// THE ROUND IS NO LONGER INJECTED, because it is no longer a guess. Every board is a
		// single round-1 sitting and every probe seat id carries `-r1` (see seatprobe.Seats), so
		// the derivation answers 1 for all four. FEOV_ROUND existed because the old derivation
		// could not tell "round 0" from "no round in this name"; it can now, and the variable is
		// gone rather than left set to a value the tool would compute anyway.
	)
	// The directory is created HERE, by the function that owns the path, rather than by the
	// caller. Moving the trajectory out of the run directory and leaving its mkdir behind in
	// probe() failed all nine boards at once — loudly and in seconds, which is the good version
	// of this mistake, but a path and the directory it needs should not be two people's job.
	if err := os.MkdirAll(filepath.Dir(trajectoryPath(runDir)), 0o755); err != nil {
		return err
	}
	// STDIN IS CLOSED EXPLICITLY. Without it the CLI waits three seconds for piped input and
	// warns, on every dispatch, which in a parallel run is noise that reads like a fault.
	devNull, err := os.Open(os.DevNull)
	if err != nil {
		return err
	}
	defer devNull.Close()
	cmd.Stdin = devNull

	out, err := os.Create(trajectoryPath(runDir))
	if err != nil {
		return err
	}
	defer out.Close()
	cmd.Stdout = out
	cmd.Stderr = out
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("claude: %w (trajectory in %s)", err, out.Name())
	}
	return nil
}

// armConstitution writes the arm's constitution beside the trajectory and returns its path.
//
// It lands OUTSIDE the run directory, for the same reason the trajectory does: a seat that can read
// the treatment applied to it is a seat reading the experiment rather than the board.
//
// EVERY ARM IS RENDERED. There was a short-circuit here — `partial` with no directive returned the
// shipped file untouched — and it was correct exactly as long as `partial` MEANT the shipped file.
// It stopped meaning that when the constitutions stopped naming verbs: `partial` became a
// constructed treatment, and the short-circuit went on returning the shipped bytes, so the arm
// dispatched was byte-identical to `none`. Two arms, one treatment, and a difference of zero
// between them that reads as a finding about naming.
//
// Caught mid-experiment, by reading the path rather than by any test: TestTheThreeArmsDiffer
// exercises Constitution(), and the bug lived in the caller that decides whether to call it. The
// test below now covers this function, which is the one the probe actually takes.
func armConstitution(src, runDir string, sf seatprobe.Surface, role string) (string, error) {
	b, err := os.ReadFile(src)
	if err != nil {
		return "", err
	}
	out := seatprobe.Constitution(b)
	dir := filepath.Dir(trajectoryPath(runDir))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	p := filepath.Join(dir, filepath.Base(runDir)+"-constitution.md")
	if err := os.WriteFile(p, out, 0o644); err != nil {
		return "", err
	}
	return p, nil
}

// constitutionFor finds the agent definition a seat actually runs under.
//
// THE REAL ONE, NOT A PARAPHRASE. The probe exists to test whether the constitution teaches the
// surface; handing the seat a summary written for the harness would test the summary.
func constitutionFor(role, dir string) (string, error) {
	if dir == "" {
		// tools/cmd/seatprobe -> tools -> the plugin root.
		wd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		dir = filepath.Join(wd, "..", "agents")
	}
	name := map[string]string{
		"lens":  "red-auditor.md",
		"merge": "red-auditor.md",
		"blue":  "blue-researcher.md",
		"bench": "lead-judge.md",
	}[role]
	p := filepath.Join(dir, name)
	if _, err := os.Stat(p); err != nil {
		return "", fmt.Errorf("no constitution for the %s seat at %s: %w — pass -constitutions", role, p, err)
	}
	return p, nil
}

// readThinking pulls the reasoning blocks out of a captured trajectory.
//
// Thinking is ADAPTIVE on current models: the seat decides per turn whether to reason visibly, so
// an empty result means it did not think out loud, NOT that the capture failed. Said here because
// the two look identical from the outside and this whole package exists to stop that confusion.
func readThinking(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()

	var out []string
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 1<<20), 1<<24)
	for sc.Scan() {
		var ev struct {
			Type    string `json:"type"`
			Message struct {
				Content []struct {
					Type     string `json:"type"`
					Thinking string `json:"thinking"`
				} `json:"content"`
			} `json:"message"`
		}
		if json.Unmarshal(sc.Bytes(), &ev) != nil || ev.Type != "assistant" {
			continue
		}
		for _, b := range ev.Message.Content {
			if b.Type == "thinking" && strings.TrimSpace(b.Thinking) != "" {
				out = append(out, "- "+strings.TrimSpace(b.Thinking))
			}
		}
	}
	if len(out) == 0 {
		return ""
	}
	return strings.Join(out, "\n")
}

func fail(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "seatprobe: "+format+"\n", a...)
	os.Exit(2)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// readAnswer pulls the seat's own prose out of a captured trajectory — the assistant text blocks,
// in order, which in the elicitation arm ARE the result.
//
// Tool calls and thinking are excluded: thinking is reported separately (it is what the seat was
// weighing, not what it decided to tell us), and a tool call is the reading it did to answer.
func readAnswer(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return "_(no trajectory captured)_"
	}
	defer f.Close()

	var out []string
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 1<<20), 1<<24)
	for sc.Scan() {
		var ev struct {
			Type    string `json:"type"`
			Message struct {
				Content []struct {
					Type string `json:"type"`
					Text string `json:"text"`
				} `json:"content"`
			} `json:"message"`
		}
		if json.Unmarshal(sc.Bytes(), &ev) != nil || ev.Type != "assistant" {
			continue
		}
		for _, b := range ev.Message.Content {
			if b.Type == "text" && strings.TrimSpace(b.Text) != "" {
				out = append(out, strings.TrimSpace(b.Text))
			}
		}
	}
	if len(out) == 0 {
		return "_(the seat produced no prose — an answer of silence, which is itself a result)_"
	}
	return strings.Join(out, "\n\n")
}

// memoryDirs is where red's accumulated gap patterns live, as run-setup reads them.
func memoryDirs() []string {
	wd, err := os.Getwd()
	if err != nil {
		return nil
	}
	// tools/ -> plugin -> plugins -> repo root
	return []string{filepath.Join(wd, "..", "..", "..", "feov-memory", "red-gap-patterns")}
}

// patternDuty renders the DUTY form: the entries whose class matches a gap on THIS board, tied to
// the gap in front of the seat rather than mounted as a file to read.
//
// It mirrors debate.js's patternDutyClause deliberately. The claim under test — stated in red's
// constitution and in blue's, both as "measured" — is that this form changes behaviour where the
// mounted file does not: "duty-embedded patterns caught both warned classes in round 1; the mounted
// file prevented nothing" against "lanes verifiably read the gap-pattern file and committed both
// warned patterns anyway". Both sentences rest on one run apiece.
func patternDuty(b seatprobe.Board) string {
	idx := setup.BuildPatternIndex(memoryDirs())
	var lines []string
	seen := map[string]bool{}
	for _, g := range b.Gaps {
		for _, e := range idx.ByClass[g.Class] {
			if seen[e.File] {
				continue
			}
			seen[e.File] = true
			lines = append(lines, fmt.Sprintf("  [%s] %s — %s", g.Class, e.Title, e.Hook))
		}
	}
	if len(lines) == 0 {
		return "\n\nPATTERN DUTY: none of this board's gap classes has an entry in red's memory."
	}
	return "\n\nPATTERN DUTY (red's accumulated memory, selected BY THE CLASS of the gaps in front of you — not the whole corpus).\n" +
		"These are defects red has already caught in THIS class of gap:\n\n" + strings.Join(lines, "\n") +
		"\n\nCheck any repair you would propose against each one before you would claim the gap closed."
}
