// Package hookmain owns the ten lines every hook binary opens with.
//
// THE POINT IS NOT THE LINE COUNT. The duplicated preamble carried the binary's name
// TWICE — once to name the flag set, once to print the version — and that pair has already
// shipped a defect: when internal/checkpointseal began serving three shims from one run(),
// the literal could not be right for all three, so sc-precompact, sc-sessionend and
// sc-subagentstop all answered -version with "sc-checkpoint-seal", the name #201 step 3
// retired. sc-doctor lists hook binaries by name and prints that line, so its table showed
// three rows with one name — and "which of these is stale" is the only question the line
// exists to answer.
//
// So the name is supplied ONCE, and it is supplied as a function, because the one package
// that got this wrong is also the one that cannot know its name until it has parsed a flag.
// An API that only serves the easy nine would leave the tenth writing the code that broke.
package hookmain

import (
	"flag"
	"fmt"
	"io"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/buildid"
)

// Named is the common case: a binary that knows its own name at compile time.
func Named(name string) func() string { return func() string { return name } }

// Preamble parses the flags every hook shares and reports whether the invocation is
// already FINISHED — either because -version was asked for, or because the flags did not
// parse.
//
// A parse failure returns "finished" rather than an error, deliberately. These are hooks:
// a bad flag is never worth failing the event over, and every caller's response to one was
// already `return 0`. Making that the shared behaviour removes the last place a caller
// could accidentally decide otherwise.
//
// resolve is called only when the version is actually requested, so a caller whose name
// depends on a flag can read that flag from the set it declared.
func Preamble(args []string, stdout, stderr io.Writer, resolve func() string, declare ...func(*flag.FlagSet)) (finished bool) {
	// The flag set is named for the resolved binary where that is knowable without
	// parsing; usage text is a human surface and a wrong name there is a small lie.
	fs := flag.NewFlagSet(resolve(), flag.ContinueOnError)
	fs.SetOutput(stderr)
	showVersion := fs.Bool("version", false, "print version and exit")
	for _, d := range declare {
		d(fs)
	}
	if err := fs.Parse(args); err != nil {
		return true
	}
	if *showVersion {
		fmt.Fprintln(stdout, buildid.Line(resolve()))
		return true
	}
	return false
}
