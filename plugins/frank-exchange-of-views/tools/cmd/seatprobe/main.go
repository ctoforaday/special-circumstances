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
//	--agent frank-exchange-of-views:<agent>         the definition production dispatches, with
//	                                                its skills and its declared tools
//	--allowedTools <the agent's own tools: line>    PERMISSION, not availability — nobody is at
//	                                                the keyboard to answer an approval prompt
//	--add-dir <runDir>                              access to the board
//
// The PROMPT is not written here at all: internal/debatejs runs the shipped debate.js and the
// probe dispatches what it hands the board's seat, verbatim.
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
		buildOnly  = flag.Bool("build-only", false, "stage the board and STOP — no seat is dispatched and nothing is scored. For a harness that drives its own dispatch (the interview) and needs the same fixture. NOT -report-only, which skips the build and reports on whatever is already there")
		printSit   = flag.Bool("print-sitting", false, "print the named board's SITTING text and exit — the situation a seat is dispatched into. For a harness that drives the dispatch itself (the interview) and must hand the seat the same situation this probe would")
		ask        = flag.Bool("ask", false, "do not dispatch a seat to ACT — ask it to ENUMERATE and ASSESS its options instead. A verb used zero times cannot say whether the seat never perceived it, weighed it and declined, or wanted it and could not reach it; this asks")
		inRun      = flag.Bool("records-in-run", false, "leave the event record under the run directory, where the seat can read it without the tool — the CONTROL arm, for measuring what the separation changes")
		memoryDir  = flag.String("memory", "", "directory holding red's accumulated gap patterns, staged into inputs/red-gap-patterns.md as run-setup stages it (default: the repo's feov-memory/red-gap-patterns, resolved from the working directory — pass this when running from anywhere but the tools module)")
		debatePath = flag.String("debate", "", "path to the shipped debate.js the probe takes its prompts from (default: the plugin's skills/research-protocol/scripts/debate.js)")
	)
	flag.Parse()

	// THE SITTING TEXT, FOR A CALLER THAT DRIVES ITS OWN DISPATCH. The interview harness holds a
	// session open across turns, which this binary does not do — but it must put the seat in the
	// SAME situation, and the situation lives here. Printing it is how the two stay one fixture
	// rather than two that drift.
	if *printSit {
		b, ok := seatprobe.Boards()[*board]
		if !ok {
			fail("no board %q", *board)
		}
		// THE SITTING IS debate.js's, NOT THIS BINARY'S. It embeds the run directory and the
		// tool path, so both are required here — a prompt rendered against a placeholder would
		// tell the seat to read a directory that does not exist, and the interview would be
		// measuring a seat that cannot see its own board.
		if *dir == "" || *bin == "" {
			fail("-print-sitting renders the prompt debate.js hands the seat, which names the run directory and the tool: pass -dir and -bin (the same values the -build-only staging used)")
		}
		d, err := seatprobe.ProductionPrompt(debateScript(*debatePath), b, filepath.Join(*dir, b.Name), filepath.Dir(*bin), *model, *model)
		if err != nil {
			fail("%v", err)
		}
		fmt.Print(d.Prompt)
		return
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
			out, err := probe(boards[name], filepath.Join(*dir, name), *bin, *constDir, *model, debateScript(*debatePath), *memoryDir, *reportOnly, *keep, *inRun, *ask, *buildOnly, surface)
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
// agentFor maps a seat role to the agent definition production dispatches for it — the same
// mapping debate.js makes with agentType.
// NO DEFAULT. A role that fell through would dispatch under ANOTHER SEAT'S constitution and the run
// would report it as its own role — a seat measured against duties it was never given, reported as
// a finding about that role. An unknown role is a programming error and says so.
func agentFor(role string) string {
	switch role {
	case "lens":
		return "red-auditor"
	case "merge":
		return "red-auditor"
	case "blue":
		return "blue-researcher"
	case "bench":
		return "lead-judge"
	}
	fail("no agent definition for role %q — the probe will not dispatch a seat under another seat's constitution", role)
	return ""
}

func trajectoryPath(runDir string) string {
	return filepath.Join(filepath.Dir(runDir), ".probe", filepath.Base(runDir)+".jsonl")
}

func probe(b seatprobe.Board, runDir, bin, constDir, model, debatePath, memoryDir string, reportOnly, keep, recordsInRun, ask, buildOnly bool, surface seatprobe.Surface) (string, error) {
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
				// THE err GOES IN THE MESSAGE, and it is not decoration. Reporting only the tool's
				// output describes a tool that RAN and refused. A process that never started —
				// wrong path, missing execute bit, an extensionless binary on Windows — produces
				// no output at all, so the board fails with a blank reason and the operator has
				// nothing to chase. Measured on the first Windows run of the sibling gate: nine
				// boards, nine empty failures.
				return string(out), fmt.Errorf("%s: %v: %s", strings.Join(args[:min(3, len(args))], " "), err, strings.TrimSpace(string(out)))
			}
			return string(out), nil
		}
		if err := seatprobe.Build(runDir, b, run); err != nil {
			return "", fmt.Errorf("build: %w", err)
		}
		// RED'S MEMORY, STAGED AS run-setup STAGES IT. This used to be an arm — `none` mounted
		// nothing — and an arm is no longer available: debate.js's prompt names
		// inputs/red-gap-patterns.md in blue's very first batched read, unconditionally, because
		// every real run has the file. A probe that withheld it would hand the seat a prompt whose
		// opening instruction fails, and score what it did next.
		if r := setup.MirrorGapPatterns(memoryDirs(memoryDir), runDir); !r.Written {
			return "", fmt.Errorf("red's gap-pattern corpus did not stage (%s) — the dispatched prompt names the file in its first instruction, so a run without it is measuring a broken read", r.Reason)
		}
		// THE FIXTURE, AND NOTHING ELSE. A caller driving its own dispatch — the interview, which
		// holds a session open across turns — needs the board this probe would have built, staged
		// the same way, and then needs this binary to stop. Scoring a sitting that never happened
		// would print "reached for 0" and mean nothing: the plausible zero, manufactured.
		//
		// It returns HERE rather than through a second staging function, so there is one build
		// path. A copy would drift the first time a board changed, and then the interview and the
		// probe would be asking about different fixtures while reporting on one.
		if buildOnly {
			return "", nil
		}
		if err := dispatch(b, runDir, bin, constDir, model, debateScript(debatePath), ask); err != nil {
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
func dispatch(b seatprobe.Board, runDir, bin, constDir, model, debatePath string, ask bool) error {
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
	// THE PROMPT IS debate.js's, RENDERED BY RUNNING IT.
	//
	// What stood here was a paraphrase written in this file: ~950 characters against production's
	// 12,800–24,000. It said the same THINGS in different words, which is the worst case — the
	// probe reported help-reading counts, verb-reach counts and a friction rate for a seat that
	// was never given production's instructions, and every one of those numbers was published as
	// a finding about seats. There is no fallback and no substitute: a board whose seat debate.js
	// does not dispatch fails here rather than being handed something written locally.
	d, err := seatprobe.ProductionPrompt(debatePath, b, runDir, filepath.Dir(bin), model, model)
	if err != nil {
		return err
	}
	if d.AgentType != "frank-exchange-of-views:"+agentFor(role) {
		return fmt.Errorf("board %q: debate.js dispatches %s under %q, and this harness maps its %s role to %q — the probe would run the seat under the wrong constitution",
			b.Name, b.Seat, d.AgentType, role, "frank-exchange-of-views:"+agentFor(role))
	}

	// THE PROMPT MUST NAME THE BINARY THIS PROBE WAS GIVEN, and this is checked against the
	// RENDERED prompt rather than against a second copy of debate.js's spelling kept here.
	//
	// debate.js builds the tool path itself, from the DIRECTORY it is handed: `${binDir}/feov-record`.
	// The probe passes filepath.Dir(bin), so a -bin whose basename is anything else tells the seat
	// to run a DIFFERENT FILE in that directory — and says nothing, because neither side is wrong
	// on its own terms.
	//
	// MEASURED 2026-08-21, and it is the reason this exists. A run was driven with
	// `-bin <scratch>/feov-record8`; the seats were told to run `<scratch>/feov-record`, which
	// existed, because a build from four days earlier was still lying in that directory. Nine
	// boards dispatched, every seat worked, the record filled up, and the report was a measurement
	// of a binary the operator had not built — one still carrying the role-prefixed surface the
	// tool had since retired. The only symptom was a denominator: `help×0 of 0 tool calls`, because
	// the trajectory reader counts invocations of the binary it was TOLD about.
	//
	// A stale artifact is what made it silent. With nothing at that path the seats would have
	// failed loudly on their first call; with something there, the miss returned a plausible run.
	prompt := d.Prompt
	if !strings.Contains(prompt, bin) {
		return fmt.Errorf("board %q: the dispatched prompt does not name %s.\n\n"+
			"debate.js builds the tool path from the DIRECTORY it is given, so the seat will run a\n"+
			"different file in %s — silently, if one happens to be there. Name the binary the way\n"+
			"production names it, or the run measures whatever else is in that directory.",
			b.Name, bin, filepath.Dir(bin))
	}

	// THE ELICITATION ARM. Same board, same constitution, same identity — the seat is asked what
	// it thinks its options are instead of being watched to see which it takes.
	//
	// The tool set drops Write and Edit, and the prompt says to record nothing. Not because a
	// seat would cheat, but because an answer about judgement should not leave board state
	// behind: the run directory has to stay comparable to the acting arm's, or the two probes
	// are measuring different boards.
	// SUBTRACTIVE, NOT A REPLACEMENT LIST. The elicitation arm must not leave board state behind —
	// an answer about judgement that records events is measuring a different sitting — but stating
	// the tool set here is what removed WebSearch and ToolSearch from every probed seat. The agent
	// definition supplies the set; this only takes the two write verbs off it.
	var deny []string
	if ask {
		prompt = seatprobe.ElicitPrompt(role, b.Seat, runDir, bin, d.Prompt)
		deny = []string{"Write", "Edit"}
	}
	// THE BOARD'S OWN WITHHOLDING, on both arms. `blocked` is a sitting about an unreachable
	// source, and it is only that sitting if the source is actually unreachable.
	deny = append(deny, b.Deny...)

	// DISPATCH THE PRODUCTION AGENT, NOT A CONSTITUTION FILE.
	//
	// This passed --system-prompt-file <constitution> and an --allowedTools list written here. Both
	// were wrong, and wrong in the direction that flatters the instrument:
	//
	//   SKILLS  the constitution declares `skills: [research-protocol, critical-stance,
	//           terse-communication]`, and a raw system-prompt file loads NONE of them. The probed
	//           seat did not have the protocol it operates under.
	//   TOOLS   the constitution declares WebSearch and ToolSearch. The hand-written list had
	//           neither — so `corroborate`, which is the verb for a source the seat goes and FINDS,
	//           was unmeetable, and the board scored the seat UNMET on it across every run.
	//
	// --agent resolves the same definition production resolves, with its skills and its declared
	// tools. It requires the plugin to be INSTALLED, which is also what production requires.
	args := []string{
		"-p", prompt,
		"--model", model,
		"--output-format", "stream-json",
		"--verbose",
		"--agent", "frank-exchange-of-views:" + agentFor(role),
		"--add-dir", runDir,
		"--max-turns", "60",
	}
	// THE PERMISSION GRANT, TAKEN FROM THE AGENT'S OWN `tools:` LINE. Nobody is at the keyboard,
	// and an ungranted Bash call comes back "This command requires approval" — measured: the seat's
	// first two calls, both `--seat-id … --help` and exactly the act under test, were refused; it
	// concluded the record tool was not responding and spent the sitting reasoning about a blocker
	// the harness had created. `--permission-mode auto` and `dontAsk` do not help; both still refuse
	// the record binary.
	//
	// Passing this alongside --agent does NOT narrow the seat's tools — measured, an agent
	// dispatched with `--allowedTools Bash` still lists WebSearch and ToolSearch as its own. The
	// old defect was a hand-written list passed with --system-prompt-file, where the list was the
	// only tool source. Reading the declaration is what keeps the grant from becoming that list
	// again.
	granted, err := seatprobe.GrantedTools(constitution)
	if err != nil {
		return fmt.Errorf("board %q: %w", b.Name, err)
	}
	args = append(args, "--allowedTools")
	args = append(args, granted...)
	if len(deny) > 0 {
		args = append(args, "--disallowedTools")
		args = append(args, deny...)
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
		// THE ROUND IS NO LONGER INJECTED, because it is no longer a guess. Every probe seat id
		// carries its round (see seatprobe.Seats — three sit round 1, and the bench sits round 2,
		// which is the first round a judge can sit at all), so the derivation answers it. FEOV_ROUND
		// existed because the old derivation could not tell "round 0" from "no round in this name";
		// it can now, and the variable is gone rather than left set to a value the tool would
		// compute anyway.
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
//
// THE DEFAULT IS RELATIVE TO THE WORKING DIRECTORY, AND THAT IS A REAL CONSTRAINT, not a detail.
// The probe is a development instrument run from the tools module, and this walks up from there.
// Run it from anywhere else and the corpus does not stage — which fails the board loudly (the
// dispatched prompt names inputs/red-gap-patterns.md in blue's first instruction, so a run without
// it is measuring a broken read) but tells the caller nothing about how to fix it. The interview
// harness, which runs from its own scratch directory, hit exactly that. Hence the flag.
func memoryDirs(flagValue string) []string {
	if flagValue != "" {
		return []string{flagValue}
	}
	wd, err := os.Getwd()
	if err != nil {
		return nil
	}
	// tools/ -> plugin -> plugins -> repo root
	return []string{filepath.Join(wd, "..", "..", "..", "feov-memory", "red-gap-patterns")}
}

// debateScript resolves the orchestrator the probe takes its prompts from.
//
// ONE FILE, NOT A COPY OF ONE. The prompt a seat is dispatched with is rendered by executing this
// script, so an edit to a clause reaches the probe on the next run. The default mirrors
// constitutionFor's: the probe is a development instrument, run from the tools module, and the
// plugin tree sits above it. A missing file is a hard failure at the point of use — a probe that
// fell back to a written prompt would be the defect this whole route removed.
func debateScript(flagValue string) string {
	if flagValue != "" {
		return flagValue
	}
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	// tools/ -> the plugin root.
	return filepath.Join(wd, "..", "skills", "research-protocol", "scripts", "debate.js")
}
