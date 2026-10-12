package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// tally is a package's verdict counts. Printed for every package, so a test that starts
// skipping is a number that moved.
type tally struct{ pass, fail, skip int }

func (t tally) String() string {
	return fmt.Sprintf("%d pass, %d fail, %d skip", t.pass, t.fail, t.skip)
}

// pkgEvents is what one package's event stream said.
type pkgEvents struct {
	verdict string            // the package-level pass, fail or skip
	tests   map[string]string // top-level test name -> its final verdict
	elapsed time.Duration
	output  string
}

func (p *pkgEvents) tally() tally {
	var t tally
	for _, v := range p.tests {
		switch v {
		case "pass":
			t.pass++
		case "fail":
			t.fail++
		case "skip":
			t.skip++
		}
	}
	return t
}

// outputOr is the package's own output, or the raw bytes when the stream carried none.
func (p *pkgEvents) outputOr(raw []byte) string {
	if strings.TrimSpace(p.output) != "" {
		return p.output
	}
	return string(raw)
}

type event struct {
	Action  string
	Package string
	Test    string
	Output  string
	Elapsed float64
}

// fold reads a `go test -json` stream into per-package events. A VERDICT IS THE EVENT'S ACTION
// FIELD, never a `--- PASS` line scraped from output: a test that prints without a trailing
// newline hides that line, and the event still arrives. stray is every line that was not an
// event, and the compiler's output for a package that did not build.
func fold(stream []byte) (byPkg map[string]*pkgEvents, stray string) {
	byPkg = map[string]*pkgEvents{}
	var other strings.Builder
	var outputs = map[string]*strings.Builder{}
	sc := bufio.NewScanner(bytes.NewReader(stream))
	sc.Buffer(make([]byte, 1<<20), 1<<26)
	for sc.Scan() {
		var ev event
		if json.Unmarshal(sc.Bytes(), &ev) != nil || ev.Action == "" {
			other.Write(sc.Bytes())
			other.WriteByte('\n')
			continue
		}
		// The compiler's output for a package that did not build is an event of its own, keyed
		// by import path rather than by package.
		if ev.Action == "build-output" {
			other.WriteString(ev.Output)
			continue
		}
		if ev.Package == "" {
			continue
		}
		p := byPkg[ev.Package]
		if p == nil {
			p = &pkgEvents{tests: map[string]string{}}
			byPkg[ev.Package] = p
			outputs[ev.Package] = &strings.Builder{}
		}
		switch ev.Action {
		case "output":
			outputs[ev.Package].WriteString(ev.Output)
		case "pass", "fail", "skip":
			switch {
			case ev.Test == "":
				p.verdict = ev.Action
				p.elapsed = time.Duration(ev.Elapsed * float64(time.Second))
			case !strings.Contains(ev.Test, "/"):
				p.tests[ev.Test] = ev.Action
			}
		}
	}
	for pkg, b := range outputs {
		byPkg[pkg].output = b.String()
	}
	return byPkg, other.String()
}
