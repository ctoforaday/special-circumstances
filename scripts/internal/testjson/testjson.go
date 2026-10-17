// Package testjson reads a `go test -json` stream into per-package, per-test verdicts.
//
// Dev tooling for this repository only; racescope and testsplit both read their runs through it.
//
// A VERDICT IS THE EVENT'S ACTION FIELD, never a `--- PASS` line scraped from output: a test
// that prints without a trailing newline glues that line to its own text, and the event still
// arrives. A subtest's verdict is not its parent's. Everything that is not a verdict and is
// worth printing is kept: a package's output events as its Output, and every line that was not
// an event, the compiler's output for a package that did not build, and output no package
// claims, as Stray.
package testjson

import (
	"bufio"
	"bytes"
	"encoding/json"
	"strings"
	"time"
)

// Package is what one package's events said.
type Package struct {
	Verdict string            // the package-level pass, fail or skip; "" when none came
	Tests   map[string]string // top-level test name -> its final pass, fail or skip
	Elapsed time.Duration     // from the package-level verdict
	Output  string            // the package's output events, in order
}

// Stream is a whole `go test -json` stream.
type Stream struct {
	Packages map[string]*Package
	Stray    string
}

type event struct {
	Action  string
	Package string
	Test    string
	Output  string
	Elapsed float64
}

// Read folds a `go test -json` stream, which may hold any number of packages.
func Read(stream []byte) Stream {
	s := Stream{Packages: map[string]*Package{}}
	var stray strings.Builder
	outputs := map[string]*strings.Builder{}
	sc := bufio.NewScanner(bytes.NewReader(stream))
	sc.Buffer(make([]byte, 1<<20), 1<<26)
	for sc.Scan() {
		var ev event
		if json.Unmarshal(sc.Bytes(), &ev) != nil || ev.Action == "" {
			// A `go` command error, or a build failure from a toolchain that prints it as text.
			stray.Write(sc.Bytes())
			stray.WriteByte('\n')
			continue
		}
		// The compiler's output for a package that did not build is an event of its own, keyed
		// by import path rather than by package.
		if ev.Action == "build-output" || (ev.Package == "" && ev.Action == "output") {
			stray.WriteString(ev.Output)
			continue
		}
		if ev.Package == "" {
			continue
		}
		p := s.Packages[ev.Package]
		if p == nil {
			p = &Package{Tests: map[string]string{}}
			s.Packages[ev.Package] = p
			outputs[ev.Package] = &strings.Builder{}
		}
		switch ev.Action {
		case "output":
			outputs[ev.Package].WriteString(ev.Output)
		case "pass", "fail", "skip":
			switch {
			case ev.Test == "":
				p.Verdict = ev.Action
				p.Elapsed = time.Duration(ev.Elapsed * float64(time.Second))
			case !strings.Contains(ev.Test, "/"):
				p.Tests[ev.Test] = ev.Action
			}
		}
	}
	for pkg, b := range outputs {
		s.Packages[pkg].Output = b.String()
	}
	s.Stray = stray.String()
	return s
}

// Answers sorts a list of asked-for names by what came back for each.
type Answers struct {
	Pass, Fail, Skip []string
	// Missing are names with no pass, fail or skip event — an Example with no `// Output:` is
	// one, and so is every name of a process that died first. Asked for and not answered: NOT
	// MEASURED. `go test -run` exits 0 when its pattern matches nothing, so the exit code cannot
	// say the names ran; only their events can.
	Missing []string
}

// Answer reports, for each name, the verdict pkg's events gave it. A package that sent no
// events answers nothing, so every name is Missing.
func (s Stream) Answer(pkg string, names []string) Answers {
	var tests map[string]string
	if p := s.Packages[pkg]; p != nil {
		tests = p.Tests
	}
	var a Answers
	for _, n := range names {
		switch tests[n] {
		case "pass":
			a.Pass = append(a.Pass, n)
		case "fail":
			a.Fail = append(a.Fail, n)
		case "skip":
			a.Skip = append(a.Skip, n)
		default:
			a.Missing = append(a.Missing, n)
		}
	}
	return a
}
