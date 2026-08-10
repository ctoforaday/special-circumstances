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
			out, err := probe(boards[name], filepath.Join(*dir, name), *bin, *constDir, *model, *reportOnly, *keep, surface)
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

func probe(b seatprobe.Board, runDir, bin, constDir, model string, reportOnly, keep bool, surface seatprobe.Surface) (string, error) {
	if !reportOnly {
		if !keep {
			if err := os.RemoveAll(runDir); err != nil {
				return "", err
			}
		}
		if err := os.MkdirAll(runDir, 0o755); err != nil {
			return "", err
		}
		run := func(args ...string) (string, error) {
			cmd := exec.Command(bin, args...)
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

	report, err := seatprobe.Report(surface, runDir, []string{b.Seat}, b.Expect)
	if err != nil {
		return "", err
	}
	// The reasoning, where the seat produced any. It is the half a record cannot hold: the record
	// says which verb was taken, and this says what the seat was weighing when it chose.
	if t := readThinking(runDir); t != "" {
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

Run directory: %s — use this ABSOLUTE path in every command.
The record tool is %s. Pass --run <the run directory> and --seat-id %s on every invocation.

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
	// STDIN IS CLOSED EXPLICITLY. Without it the CLI waits three seconds for piped input and
	// warns, on every dispatch, which in a parallel run is noise that reads like a fault.
	devNull, err := os.Open(os.DevNull)
	if err != nil {
		return err
	}
	defer devNull.Close()
	cmd.Stdin = devNull

	out, err := os.Create(filepath.Join(runDir, "probe-trajectory.jsonl"))
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
func readThinking(runDir string) string {
	f, err := os.Open(filepath.Join(runDir, "probe-trajectory.jsonl"))
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
