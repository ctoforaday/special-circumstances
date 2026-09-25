// Command feov-subagentstart is the SubagentStart hook backend. It records one end of a
// sitting's span and writes NOTHING to stdout.
//
// THE SILENCE IS A CHOICE, NOT A CONSTRAINT, and this comment said the opposite. It read that a
// hook emitting `additionalContext` "re-invokes the seat and fires again — nine times", citing
// plans/hook-surface-spike.md §10. That nine-firing loop is `SubagentStop`'s, measured in the same
// section and REFUTED as a channel because it delivers nothing anywhere. §10's other half is the
// opposite result: `SubagentStart` fires ONCE and its marker arrives in the SEAT's context —
// "VERIFIED TWICE, INDEPENDENTLY" in §5's status table, by #500 and #507.
//
// So injection here is available and is the correct shape for handing a subagent its own context.
// What this binary does today is observe, which needs no injection; #1122 is the change that would
// use the channel, and a reader who took the old comment at its word would not have looked.
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

func main() {
	os.Exit(hookcmd.Run("feov-subagentstart", "SubagentStart", sittinghook.Start, os.Stdin, os.Stdout))
}
