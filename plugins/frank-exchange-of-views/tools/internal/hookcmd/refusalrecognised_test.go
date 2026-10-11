package hookcmd_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/hookcmd"
)

// THE HOOK RECOGNISES THE REFUSAL THE RECORD TOOL ACTUALLY WRITES.
//
// The record tool logs each refusal it gives, so the failure hook must not log the same call
// again as a `failure`. It tells the tool's own refusal by its shape — and the test of that used a
// hand-typed literal no binary emitted, while the hook matched a byte prefix the tool's handler
// refusals did not start with: every --json refusal a handler gave was logged twice.
//
// So the bytes come from the tool. Each call below is refused by the real tree through cli.Invoke —
// the binary's own sequence — and the hook is handed what that wrote, the way the harness hands it
// a failed Bash call.
func TestTheHookRecognisesTheRecordToolsOwnRefusal(t *testing.T) {
	for name, argv := range map[string][]string{
		// With no seat named nothing reads a record, so each is refused wherever it runs.
		"refused before a command resolved":   {"finding", "--json"},
		"and the same call in its human form": {"finding"},
		"a flag the root does not have":       {"--nonesuch", "--json"},
	} {
		t.Run(name, func(t *testing.T) {
			root := cli.NewRootFor("")
			var stdout, stderr bytes.Buffer
			root.SetOut(&stdout)
			root.SetErr(&stderr)
			if _, err := cli.Invoke(root, argv, ""); err == nil {
				t.Fatal("the call was not refused, so there is nothing for the hook to recognise")
			}
			seen := stdout.String() + stderr.String()
			// The harness names the tool as the seat invoked it; a test binary is named otherwise.
			seen = strings.Replace(seen, root.Name()+": ", "feov-record: ", 1)
			input, _ := json.Marshal(map[string]string{"command": "feov-record " + strings.Join(argv, " ")})
			if entry, logs := hookcmd.FailureEntry("Bash", input, "Exit code 2\n"+seen); logs {
				t.Errorf("the record tool's own refusal is logged a second time, as a failure: %s\n\nthe tool wrote:\n%.300s", entry, seen)
			}
		})
	}
}
