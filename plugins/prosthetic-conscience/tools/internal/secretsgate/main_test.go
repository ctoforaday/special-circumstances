package secretsgate

import (
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/hookfailures"
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
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "pc-hookfailures-test-")
	if err != nil {
		fmt.Fprintln(os.Stderr, "TestMain:", err)
		os.Exit(1)
	}
	defer os.RemoveAll(dir)
	for _, k := range []string{"HOME", "USERPROFILE", "XDG_STATE_HOME"} {
		if err := os.Setenv(k, dir); err != nil {
			fmt.Fprintln(os.Stderr, "TestMain:", err)
			os.Exit(1)
		}
	}
	os.Exit(m.Run())
}
