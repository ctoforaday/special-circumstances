// Package gcpolicy sets how a shipped binary's garbage collector runs: never on a schedule for a
// short-lived command, at the runtime's default for a long-running mode.
//
// Every `main` under plugins/*/tools/cmd calls Short or Long as its first statement, so each binary
// states its class where it starts; the authored test in scripts/internal/gcpolicy derives the set
// of mains from the directory tree and fails one that states neither. A binary with a long mode
// behind a flag or a verb calls Short in main and Long when that mode starts.
//
// WHY THE GC IS OFF. A hook or a seat verb runs for milliseconds to a fraction of a second and
// exits; the exit returns its heap whole, so a collection that runs before then is CPU spent on
// memory nobody reuses. Measured 2026-11-03 on an epoch-55 record, before and after interleaved,
// median CPU per call: feov-pretooluse StructuredOutput 234 -> 191 ms, feov-subagentstart
// 216 -> 186 ms, feov-subagentstop 39 -> 32 ms, sc-pretooluse Bash 67 -> 52 ms; feov-record
// `show` for a lens 194 -> 155 ms, for the chair 157 -> 114 ms, `show report` 62 -> 51 ms,
// `scorecard` 103 -> 80 ms; gray-area rework over a 136 MB transcript 1018 -> 916 ms. Hooks that
// finish in under ~5 ms never reach the runtime's first collection and do not move, nor do
// telepathy's reads, which spend their time in SQLite and ripgrep.
//
// WHY A LIMIT. With the GC off the heap grows with every allocation, so an input far larger than
// any measured one would grow without bound. The memory limit turns the GC back on at Cap and only
// there: below it nothing is collected, at it the runtime collects to stay under it. Peak RSS with
// the GC off, the same day: hooks at most 49 MB (feov-pretooluse StructuredOutput); feov-record
// seat and operator reads 34-67 MB (`show` for a lens is the largest); telepathy reads 14-28 MB;
// gray-area over the largest transcript on the box (136 MB) 169 MB. Cap sits above every one of
// them, so none of them collects; three copies of that transcript peak at 246 MB under Cap against
// 506 MB with the GC simply off.
//
// WHAT STAYS ON. A process that runs for minutes or allocates in proportion to a whole corpus keeps
// the default: feov-migrate (743 MB peak with the GC off against 45 MB with it on, migrating the
// largest populated record, and no faster), feov-record dashboard --watch and --serve, and
// telepathy backfill (6.4 GB with the GC off against 163 MB, over 1.7 GB of transcripts).
//
// AN OPERATOR'S SETTING WINS. GOGC or GOMEMLIMIT in the environment is applied by the runtime
// before main runs; when either is set, neither function changes anything.
package gcpolicy

import (
	"math"
	"os"
	"runtime/debug"
)

// Cap is the memory limit a short-lived command runs under with its GC off: 256 MiB.
const Cap = 256 << 20

// Short turns the scheduled GC off and sets the memory limit to Cap, unless the operator set
// GOGC or GOMEMLIMIT.
func Short() {
	if operatorSet() {
		return
	}
	debug.SetGCPercent(-1)
	debug.SetMemoryLimit(Cap)
}

// Long restores the runtime's default GC — GOGC=100 and no memory limit — for a long-running mode,
// unless the operator set GOGC or GOMEMLIMIT.
func Long() {
	if operatorSet() {
		return
	}
	debug.SetGCPercent(100)
	debug.SetMemoryLimit(math.MaxInt64)
}

// operatorSet reports whether the environment carries a GC setting the runtime has already
// applied. An empty value is not one: the runtime reads it as the default, and so does this.
func operatorSet() bool {
	return os.Getenv("GOGC") != "" || os.Getenv("GOMEMLIMIT") != ""
}
