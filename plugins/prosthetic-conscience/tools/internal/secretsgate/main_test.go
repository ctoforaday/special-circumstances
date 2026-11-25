package secretsgate

import (
	"github.com/ctoforaday/special-circumstances/shared/testhome"
	"io"
	"os"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/shared/hookfailures"
)

func TestDenyShape(t *testing.T) {
	d := deny("why")
	if d.HookSpecificOutput.PermissionDecision != "deny" ||
		d.HookSpecificOutput.HookEventName != "PreToolUse" ||
		d.HookSpecificOutput.PermissionDecisionReason != "why" {
		t.Fatalf("deny() malformed: %+v", d)
	}
}

// testRecorder is a recorder for a unit under test: its lines go nowhere, and TestMain points the
// record at a temporary state directory so no test writes the developer's own.
func testRecorder() *hookfailures.Recorder {
	return hookfailures.New("prosthetic-conscience", "test", "PreToolUse", time.Time{}, io.Discard)
}

// No test in this package may write the real failure record.
func TestMain(m *testing.M) { os.Exit(testhome.Run(m, "pc-hookfailures-test-")) }
