// Command agentgen writes each seat's VERB SURFACE into its agent definition, so a seat arrives
// knowing its surface instead of spending tool calls and turns fetching it; with -check it fails
// when a committed definition is stale.
//
// Dev tooling for this repository only. Nothing here ships to an installing project.
//
// # What this is for, measured
//
// Across one completed 69-seat run: 171 tool calls — 10% of every call in the run — went on
// obtaining the seat manual, and 4,280,752 characters of it entered seat context, 92% of all Read
// traffic. A single BARREN lens sitting reads 100,565 characters of which 69,603 (69.2%) is the
// lens reading its own manual, before it looks at the report. The document is identical for every
// seat of a role and unchanged between runs.
//
// The manual mechanism itself is sound and is NOT what this replaces: only 8 `--help` calls
// happened in that whole run, against ~16 per seat under per-verb help. What is wasteful is the
// FETCH — a static document obtained at runtime in 2–7 tool calls, each a turn that re-reads the
// seat's grown context. This moves the same bytes into the agent definition, where they arrive
// once with the system prompt and cost no turns.
//
// # Why four renders and not thirteen
//
// The surface is scoped by ROLE (record/roles.go), so every seat of a role gets a byte-identical
// manual apart from ONE header line naming the seat. Measured: red-lens-evidence and red-lens-voice
// differ by exactly one line; blue-respond and blue-synthesize likewise. So the tool is asked once
// per role and the header is written here, naming the role rather than a sample seat — otherwise
// blue-respond's constitution would open by calling itself frontier.
//
// # The binary's own name is part of the output
//
// `manual` renders "<argv0> manual — the lens surface, seat <id>: N commands…", so a probe built to
// a scratch path emits a tool name that does not exist. The seat would read it and type it. The
// build below therefore writes to a file NAMED feov-record, and checkBinaryName refuses anything
// else before a byte is generated.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ctoforaday/special-circumstances/scripts/internal/gitx"
)

const (
	srcDir    = "scripts/agentgen/src"
	agentsDir = "plugins/frank-exchange-of-views/agents"
	toolsDir  = "plugins/frank-exchange-of-views/tools"
	beginMark = "<!-- BEGIN GENERATED SURFACE — scripts/agentgen writes this. DO NOT EDIT BY HAND. -->"
	endMark   = "<!-- END GENERATED SURFACE -->"
)

// surfaceOf names, per agent definition, the role whose surface it carries and a seat the tool will
// accept to render it. The seat is a RENDERING HANDLE, not the definition's identity: the header
// the tool writes from it is replaced below, because one definition can serve several seats.
//
// Bound to the roster by TestEveryAgentDefinitionNamesADispatchableSeat, so a definition added here
// against a seat no dispatch produces fails rather than generating a surface nobody can use.
var surfaceOf = map[string]struct{ role, sample string }{
	"red-lens-adversary":    {"lens", "red-lens-adversary"},
	"red-lens-architecture": {"lens", "red-lens-architecture"},
	"red-lens-computation":  {"lens", "red-lens-computation"},
	"red-lens-dark-side":    {"lens", "red-lens-dark-side"},
	"red-lens-evidence":     {"lens", "red-lens-evidence"},
	"red-lens-logic":        {"lens", "red-lens-logic"},
	"red-lens-voice":        {"lens", "red-lens-voice"},
	"red-chair":             {"chair", "red-chair"},
	"blue-researcher":       {"blue", "blue-respond"},
	"blue-synthesizer":      {"blue", "blue-synthesize"},
	"lead-judge":            {"bench", "judge"},
}

func main() {
	check := flag.Bool("check", false, "verify the committed definitions carry the surface the tool renders, instead of writing them")
	flag.Parse()

	root, err := gitx.Root()
	if err != nil {
		fail(err)
	}
	bin, cleanup, err := buildTool(root)
	if err != nil {
		fail(err)
	}
	defer cleanup()

	// ONE RENDER PER ROLE, not per definition: the surface is the role's and the tool would
	// otherwise be asked the same question eleven times.
	byRole := map[string][]byte{}
	for name, s := range surfaceOf {
		if _, done := byRole[s.role]; done {
			continue
		}
		out, err := renderSurface(bin, s.sample, s.role)
		if err != nil {
			fail(fmt.Errorf("rendering the %s surface for %s: %w", s.role, name, err))
		}
		byRole[s.role] = out
	}

	stale := []string{}
	for name, s := range surfaceOf {
		path := filepath.Join(root, agentsDir, name+".md")
		authored, err := expand(filepath.Join(root, srcDir), name+".md", 0)
		if err != nil {
			fail(fmt.Errorf("%s: %w", name, err))
		}
		want, err := splice(authored, byRole[s.role])
		if err != nil {
			fail(fmt.Errorf("%s: %w", name, err))
		}
		cur, err := os.ReadFile(path)
		if err != nil && !os.IsNotExist(err) {
			fail(fmt.Errorf("reading %s: %w", name, err))
		}
		if bytes.Equal(cur, want) {
			continue
		}
		if *check {
			stale = append(stale, name)
			continue
		}
		if err := os.WriteFile(path, want, 0o644); err != nil {
			fail(err)
		}
		fmt.Printf("agentgen: wrote %s.md\n", name)
	}

	if !*check {
		fmt.Printf("agentgen: %d definition(s) carry %d role surface(s)\n", len(surfaceOf), len(byRole))
		return
	}
	if len(stale) > 0 {
		fmt.Fprintf(os.Stderr, "agentgen: %d agent definition(s) carry a stale surface: %s\n"+
			"The surface block is generated from the tool's own `manual`; edit the verb, never the block.\n"+
			"Regenerate: (cd scripts && go run ./agentgen)\n", len(stale), strings.Join(stale, ", "))
		os.Exit(1)
	}
	fmt.Printf("agentgen: %d definition(s) match the surface the tool renders\n", len(surfaceOf))
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "agentgen:", err)
	os.Exit(1)
}

// buildTool compiles feov-record TO A FILE OF THAT NAME. `manual` prints argv0 as the command a
// seat should type, so a binary at ./probe renders "probe manual — …" and the generated
// constitution would name a tool that does not exist on any box.
func buildTool(root string) (string, func(), error) {
	dir, err := os.MkdirTemp("", "agentgen")
	if err != nil {
		return "", func() {}, err
	}
	cleanup := func() { os.RemoveAll(dir) }
	bin := filepath.Join(dir, "feov-record")
	cmd := exec.Command("go", "build", "-o", bin, "./cmd/feov-record")
	cmd.Dir = filepath.Join(root, toolsDir)
	if out, err := cmd.CombinedOutput(); err != nil {
		cleanup()
		return "", func() {}, fmt.Errorf("building feov-record: %v\n%s", err, out)
	}
	if filepath.Base(bin) != "feov-record" {
		cleanup()
		return "", func() {}, fmt.Errorf("the tool must be built to a file named feov-record — manual renders argv0 as the command a seat types")
	}
	return bin, cleanup, nil
}

// renderSurface asks the tool for a seat's whole surface and replaces the tool's own header.
//
// THE HEADER NAMES THE SAMPLE SEAT, AND SEVERAL SEATS SHARE ONE DEFINITION. Left alone,
// blue-respond's constitution would open "the blue surface, seat blue-synthesize". The header is
// therefore written here, naming the ROLE, which is the thing that is actually true of every seat
// the block reaches.
func renderSurface(bin, seat, role string) ([]byte, error) {
	cmd := exec.Command(bin, "--seat-id", seat, "manual")
	// A SEAT'S ENVIRONMENT MUST NOT REACH THE GENERATOR. FEOV_RUN and friends make the tool answer
	// as a seat inside a live run; the surface would then be rendered against that run's state.
	cmd.Env = scrubbed(os.Environ())
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("`feov-record --seat-id %s manual` failed: %w", seat, err)
	}
	lines := bytes.SplitN(out, []byte("\n"), 2)
	if len(lines) != 2 {
		return nil, fmt.Errorf("the %s surface rendered no body", role)
	}
	head := fmt.Sprintf("Your surface — the %s surface, rendered from the record tool's own `manual`. "+
		"Every command you may run is below, each under a header naming it and followed by the help it prints. "+
		"A name you did not read here is a guess.", role)
	return append([]byte(head+"\n"), lines[1]...), nil
}

func scrubbed(env []string) []string {
	out := env[:0:0]
	for _, kv := range env {
		if strings.HasPrefix(kv, "FEOV_") {
			continue
		}
		out = append(out, kv)
	}
	return out
}

// expand reads a source and inlines every `@include <path>` line, relative to the source root.
//
// THE OUTPUT MUST BE SELF-CONTAINED, so this is a GENERATION-TIME directive and never survives into
// a committed definition. The harness loads an agent .md as the seat's system prompt and expands
// nothing; a definition that shipped an `@include` would hand the seat a line naming a file it
// cannot read, in the document that tells it what it may do.
//
// The depth bound is not decoration: a fragment that includes itself would otherwise hang the
// generator, and a gate that hangs is a gate nobody runs.
func expand(srcRoot, rel string, depth int) ([]byte, error) {
	if depth > 8 {
		return nil, fmt.Errorf("@include nested more than 8 deep at %s — a fragment probably includes itself", rel)
	}
	raw, err := os.ReadFile(filepath.Join(srcRoot, rel))
	if err != nil {
		return nil, err
	}
	var out bytes.Buffer
	for _, line := range strings.SplitAfter(string(raw), "\n") {
		t := strings.TrimSpace(strings.TrimSuffix(line, "\n"))
		inc, ok := strings.CutPrefix(t, "@include ")
		if !ok {
			out.WriteString(line)
			continue
		}
		body, err := expand(srcRoot, strings.TrimSpace(inc), depth+1)
		if err != nil {
			return nil, fmt.Errorf("%s includes %s: %w", rel, inc, err)
		}
		// EXACTLY ONE TRAILING NEWLINE IS THE @include LINE'S OWN. Trimming all of them eats a
		// fragment that ends in a deliberate blank line, and the blank lines between duties are
		// not decoration — they are what makes the rendered constitution readable. Measured: the
		// first extraction silently dropped four of them across the two blue definitions.
		out.Write(bytes.TrimSuffix(body, []byte("\n")))
		if strings.HasSuffix(line, "\n") {
			out.WriteString("\n")
		}
	}
	return out.Bytes(), nil
}

// splice replaces the generated block, or appends one when the definition has none yet.
func splice(cur, surface []byte) ([]byte, error) {
	block := []byte(beginMark + "\n\n" + string(surface) + "\n" + endMark + "\n")
	i := bytes.Index(cur, []byte(beginMark))
	if i < 0 {
		body := bytes.TrimRight(cur, "\n")
		return append(append(body, []byte("\n\n")...), block...), nil
	}
	j := bytes.Index(cur[i:], []byte(endMark))
	if j < 0 {
		return nil, fmt.Errorf("the generated block opens and never closes — %s is missing", endMark)
	}
	tail := cur[i+j+len(endMark):]
	out := append(append([]byte{}, cur[:i]...), block...)
	return append(out, bytes.TrimLeft(tail, "\n")...), nil
}
