package main

import (
	"fmt"
	"strings"

	"github.com/ctoforaday/special-circumstances/scripts/internal/testjson"
)

// tally is a package's verdict counts. Printed for every package, so a test that starts
// skipping is a number that moved.
type tally struct{ pass, fail, skip int }

func (t tally) String() string {
	return fmt.Sprintf("%d pass, %d fail, %d skip", t.pass, t.fail, t.skip)
}

// tallyOf counts every top-level verdict a package's events gave.
func tallyOf(p *testjson.Package) tally {
	var t tally
	for _, v := range p.Tests {
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
func outputOr(p *testjson.Package, raw []byte) string {
	if p != nil && strings.TrimSpace(p.Output) != "" {
		return p.Output
	}
	return string(raw)
}
