// Command feov-subagentstart is the SubagentStart hook backend. It records one end of a
// sitting's span and writes NOTHING to stdout.
//
// THE SILENCE IS THE CONTRACT, not an omission. A $N hook that emits
// `additionalContext` re-invokes the seat and fires again — nine times for one seat in the
// measured case (plans/hook-surface-spike.md §10), discarding what it returned every time. This
// binary observes and says nothing, which is the shape that fires once.
//
// A DEDICATED BINARY, NOT A VERB. See internal/hookcmd: a hook has no seat identity, and hanging
// it off an identity-scoped command tree cost an exemption at every layer before the shipped
// invocation stopped exiting 2.
package main

import (
	"os"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/hookcmd"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/sittinghook"
)

func main() { os.Exit(hookcmd.Run(sittinghook.Start, os.Stdin, os.Stdout)) }
