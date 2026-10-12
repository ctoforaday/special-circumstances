// Package testsplit holds what `scripts/testsplit` and `scripts/check` must agree on: which
// packages are run as several processes, and the arguments each path passes.
//
// Dev tooling for this repository only. Nothing here ships to an installing project.
package testsplit

import (
	"strconv"
	"strings"
)

// Row is one package that is run as several test processes instead of one.
//
// A package earns a row when its suite is long and SERIAL: `go test` gives a package one
// process, and tests that do not call t.Parallel run one at a time in it, so the package uses
// one core however many the box has. Dealing its tests across processes gives each process its
// own environment, working directory and package-level state, which is what kept those tests
// from being parallel in the first place.
//
// THIS IS A TUNING TABLE, NOT A DERIVED CARRIER. A missing row costs time and never coverage:
// every package without one runs under plain `go test`, in the same invocation of the tool.
type Row struct {
	Dir     string // the module's directory, relative to the repository root
	Module  string // module path
	Package string // import path
	// SerialSeconds is the package's measured wall clock as one process, and the reason the
	// row exists. It is not read by the tool.
	SerialSeconds int
	Why           string
}

const (
	feovDir = "plugins/frank-exchange-of-views/tools"
	feov    = "github.com/ctoforaday/special-circumstances/" + feovDir
)

// Table is every split package.
//
// internal/record/migrate has NO row, deliberately: its archive store migrates each archived
// run once per test PROCESS, so dealing its sweeps across N processes would migrate every
// archive N times and pay back what the store saved.
var Table = []Row{
	{feovDir, feov, feov + "/internal/cli", 61,
		"478 tests, 20 of them parallel; the rest are serialised by t.Setenv and a package-level fetcher swap"},
	{feovDir, feov, feov + "/internal/record", 48,
		"467 tests, almost all serial, each opening its own database"},
}

// For returns the rows of one module.
func For(module string) []Row {
	var rows []Row
	for _, r := range Table {
		if r.Module == module {
			rows = append(rows, r)
		}
	}
	return rows
}

// Splits reports whether the module in dir (relative to the repository root) has a row, which is
// what decides that its suite is run by the tool rather than by plain `go test`.
func Splits(dir string) bool {
	for _, r := range Table {
		if r.Dir == dir {
			return true
		}
	}
	return false
}

// SuiteArgs is the `go` argument list, run in the scripts module, that runs the whole suite of
// the module in moduleDir. `check`, `golden` and the workflow all run this one command, so the
// three cannot disagree about what "the suite" is.
func SuiteArgs(moduleDir string) []string {
	return []string{"run", "./testsplit", "-module", moduleDir}
}

// NoCache is on BOTH paths, and it lives here because the gate's own arguments no longer show
// it: `check` runs the tool, the tool runs the tests. A cached PASS is byte-identical to a real
// one, and the contract tests shell out to a built binary whose inputs the cache cannot see.
const NoCache = "-count=1"

// UnsplitArgs is the `go` argument list for every package that has no row.
func UnsplitArgs(timeout string, parallel int, pkgs []string) []string {
	args := []string{"test", NoCache, "-json", "-timeout=" + timeout, "-p", strconv.Itoa(parallel)}
	return append(args, pkgs...)
}

// CompileArgs builds a split package's test binary without running it.
func CompileArgs(bin, pkg string) []string {
	return []string{"test", "-c", "-o", bin, pkg}
}

// SplitArgs is the argument list one process of a split package's test binary runs with: the
// flags `go test` itself would pass, and exactly the names dealt to it.
func SplitArgs(timeout string, names []string) []string {
	return []string{"-test.count=1", "-test.v=test2json", "-test.timeout=" + timeout,
		"-test.run", "^(" + strings.Join(names, "|") + ")$"}
}
