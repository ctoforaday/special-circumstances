// Command feov-sitting-write appends one end of a sitting's span to a run's record.
//
// NOT A HOOK AND NOT A SEAT VERB. It is spawned by the SubagentStart/SubagentStop hooks once they
// have established there is a live run and a real seat — see internal/sittingwrite for why the
// write is a separate process, which is a frequency argument rather than a structural one.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/sittingwrite"
)

func main() {
	run := flag.String("run", "", "the run directory whose record to append to")
	phase := flag.String("phase", "", "which end of the span: open or close")
	agentID := flag.String("agent-id", "", "the harness handle for the subagent")
	agentType := flag.String("agent-type", "", "the agent configuration it was dispatched as")
	transcript := flag.String("transcript", "", "the finished seat's transcript, whose turns are ingested (SubagentStop only)")
	flag.Parse()

	if err := sittingwrite.Write(*run, sittingwrite.Phase(*phase), *agentID, *agentType, *transcript); err != nil {
		// Stderr and a non-zero exit, which the HOOK deliberately ignores: a seat is not blocked
		// because the bookkeeping failed, and the hook's own contract is to stay silent.
		fmt.Fprintln(os.Stderr, "feov-sitting-write:", err)
		os.Exit(1)
	}
}
