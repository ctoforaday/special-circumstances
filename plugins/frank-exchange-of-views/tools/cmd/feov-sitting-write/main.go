// Command feov-sitting-write appends a harness-observed fact about a sitting to a run's record:
// one end of its span, or that it reached the run's tool-call limit.
//
// NOT A HOOK AND NOT A SEAT VERB. It is spawned by the SubagentStart/SubagentStop hooks once they
// have established there is a live run and a real seat, and by the PreToolUse hook once per sitting
// that reaches the limit — see internal/sittingwrite for why the write is a separate process, which
// is a frequency argument rather than a structural one.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/sittingwrite"
)

func main() {
	run := flag.String("run", "", "the run directory whose record to append to")
	phase := flag.String("phase", "", "open or close (an end of the span), or limit (the sitting reached the tool-call limit)")
	agentID := flag.String("agent-id", "", "the harness handle for the subagent")
	agentType := flag.String("agent-type", "", "the agent configuration it was dispatched as")
	transcript := flag.String("transcript", "", "the finished seat's transcript, whose turns are ingested and whose path is recorded (SubagentStop only)")
	sessionID := flag.String("session-id", "", "the conversation this sitting happened in, from the hook payload")
	promptID := flag.String("prompt-id", "", "the dispatch's own id, from the hook payload — it changes when an agent is re-prompted and its agent id does not")
	sitting := flag.Int("sitting", 0, "the seat's sitting that reached the limit (limit only)")
	limit := flag.Int("limit", 0, "the per-sitting tool-call limit it reached (limit only)")
	flag.Parse()

	var err error
	if sittingwrite.Phase(*phase) == sittingwrite.Limit {
		err = sittingwrite.WriteLimit(*run, *agentID, *agentType, *sitting, *limit)
	} else {
		// STDOUT IS THE SEAT'S CHANNEL, and stderr is the caller's. On the opening end this process
		// prints the seat's work list here and the SubagentStart hook passes it to the dispatched
		// subagent; a diagnostic printed to stdout would arrive in a seat's context as its work.
		err = sittingwrite.Write(sittingwrite.Sitting{
			RunDir: *run, Phase: sittingwrite.Phase(*phase),
			AgentID: *agentID, AgentType: *agentType, TranscriptPath: *transcript,
			SessionID: *sessionID, PromptID: *promptID,
		}, os.Stdout)
	}
	if err != nil {
		// Stderr and a non-zero exit. The span hooks ignore it — a seat is not blocked because the
		// bookkeeping failed; the PreToolUse hook passes it to its own stderr.
		fmt.Fprintln(os.Stderr, "feov-sitting-write:", err)
		os.Exit(1)
	}
}
