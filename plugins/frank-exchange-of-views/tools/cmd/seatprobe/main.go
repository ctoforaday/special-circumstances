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
		inRun      = flag.Bool("records-in-run", false, "leave the event record under the run directory, where the seat can read it without the tool — the CONTROL arm, for measuring what the separation changes")
	)
	flag.Parse()

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
	sem := make(chan struct{}, max(1, *parallel))
	var wg sync.WaitGroup
	for i, name := range names {
		wg.Add(1)
		go func(i int, name string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			out, err := probe(boards[name], filepath.Join(*dir, name), *bin, *constDir, *model, *reportOnly, *keep, *inRun, surface)
			if err != nil {
				results[i] = fmt.Sprintf("## %s — FAILED\n\n%v\n", name, err)
				return
			}
			results[i] = out
		}(i, name)
	}
	wg.Wait()

	fmt.Println("# Seat probe")
	fmt.Println()
	fmt.Printf("%d board(s), model %s. What each seat CHOSE, of what its role offers.\n\n", len(names), *model)
	for _, r := range results {
		fmt.Println(r)
	}
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

func probe(b seatprobe.Board, runDir, bin, constDir, model string, reportOnly, keep, recordsInRun bool, surface seatprobe.Surface) (string, error) {
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
		recordRoot := ""
		if !recordsInRun {
			r, err := os.MkdirTemp("", "feov-records-")
			if err != nil {
				return "", err
			}
			recordRoot = r
			if !keep {
				defer os.RemoveAll(recordRoot)
			}
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
		if err := dispatch(b, runDir, bin, constDir, model); err != nil {
			return "", fmt.Errorf("dispatch: %w", err)
		}
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
	// The reasoning, where the seat produced any. It is the half a record cannot hold: the record
	// says which verb was taken, and this says what the seat was weighing when it chose.
	if t := readThinking(trajectoryPath(runDir)); t != "" {
		report += "\n### What the seat was reasoning about\n\n" + t + "\n"
	}
	return report, nil
}

// dispatch runs one seat at the board through the `claude` CLI.
func dispatch(b seatprobe.Board, runDir, bin, constDir, model string) error {
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

	prompt := fmt.Sprintf(`You are the %s seat in a frank-exchange-of-views run. Your seat id is %s.

Run directory: %s — use this ABSOLUTE path when you read files under it.
The record tool is %s. Your identity and run are INJECTED (%s): do not pass --run or --seat-id.

Read the board and the artifact under audit, then do your sitting's work. Decide and act; do not ask me what to do.

%s`, role, b.Seat, runDir, bin, b.Seat, b.Sitting())

	args := []string{
		"-p", prompt,
		"--model", model,
		"--output-format", "stream-json",
		"--verbose",
		"--system-prompt-file", constitution,
		"--add-dir", runDir,
		"--allowedTools", "Bash Read Write Edit Grep Glob",
		"--max-turns", "60",
	}
	cmd := exec.Command("claude", args...)
	cmd.Dir = runDir
	// THE IDENTITY IS INJECTED, BECAUSE THAT IS WHAT PRODUCTION DOES. The PreToolUse hook
	// prefixes a seat's feov-record calls with FEOV_RUN, and #348 extended the same treatment to
	// identity — ResolveSeat prefers the injected seat over --seat-id and REFUSES a flag that
	// disagrees, because "omitting --seat-id is correct and always right".
	//
	// The probe does not load the plugin's hooks, so without this it was telling seats to type
	// both on every call: a HARDER surface than any real run presents, and every mistyped path
	// or seat id it measured was friction production had already designed away.
	cmd.Env = append(os.Environ(),
		seatenv.Var+"="+runDir,
		seatenv.SeatVar+"="+b.Seat,
		// Every board is a single round-1 sitting. Injected rather than inferred: the regex
		// fallback over a seat id cannot tell "round 0" from "no round in this name", which is
		// the defect #348 put this variable here to end.
		seatenv.RoundVar+"=1",
	)
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
