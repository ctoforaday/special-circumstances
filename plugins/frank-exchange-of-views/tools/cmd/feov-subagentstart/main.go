// Command feov-subagentstart is the SubagentStart hook backend. It records one end of a sitting's
// span and hands the dispatched seat its work list.
//
// THE ONE EVENT IN THIS PLUGIN THAT SPEAKS TO A SEAT. A document returning `additionalContext` here
// reaches the dispatched subagent's context and nothing else — one firing, verified three times
// (plans/hook-surface-spike.md §5 and §10, and again 2026-09-25). The nine-firing loop that makes
// `SubagentStop` useless as a channel is that event's and does not apply to this one; §10 measured
// both ends in the same run and they are opposite results.
//
// The work list is rendered by the spawned writer, which already carries the record — see
// internal/sittinghook for why this process links nothing that can read one.
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
