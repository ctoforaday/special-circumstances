// Command sc-subagentstop is a thin shim for the SubagentStop seal.
//
// One binary per EVENT (#201 step 3). sc-checkpoint-seal used to serve all three seal events
// from one binary, dispatched by an -event flag — the inverse of the rule, and the flag
// existed only to work around it. Each shim now knows its own event by construction, so the
// flag, its hook_event_name fallback, and the version-skew path they guarded are all gone.
//
// The seal itself did not fork: all three run the SAME unit, so the drift the merged binary
// was written to avoid is prevented by the library rather than by the merge.
package main

import (
	"os"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/checkpointseal"
)

func main() { os.Exit(checkpointseal.MainFor(checkpointseal.EvSubagentStop)) }
