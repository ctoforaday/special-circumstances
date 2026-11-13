package hookcmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/runlive"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatenv"
	"github.com/ctoforaday/special-circumstances/shared/hookfailures"
)

// hookRun is a live run, with the process environment cleared of identity so only the payload
// speaks. It returns the cwd a payload carries and the run directory.
func hookRun(t *testing.T) (cwd, runDir string) {
	t.Helper()
	for _, v := range []string{seatenv.Var, seatenv.AgentVar, seatenv.TypeVar} {
		t.Setenv(v, "")
	}
	cwd = t.TempDir()
	runDir = filepath.Join(cwd, "run")
	if err := os.MkdirAll(filepath.Join(runDir, "inputs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := runlive.WriteRunLiveMarker(cwd, runDir, nil, time.Now(), "run_test", "", testLead); err != nil {
		t.Fatal(err)
	}
	return cwd, runDir
}

// testRecorder records into the package's isolated state directory (TestMain) and writes its lines
// nowhere.
func testRecorder() *hookfailures.Recorder {
	return hookfailures.New("frank-exchange-of-views", "test", "PreToolUse", time.Now(), io.Discard)
}

// No test in this package may write the developer's own failure record: every entry point here now
// settles into ~/.local/state unless something stops it.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "feov-hookfailures-test-")
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
