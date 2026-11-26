package main

import (
	"context"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/shared/testhome"
)

// TestHarnessCLIFailsClosedUnderHookcoreLive pins S1b: under HOOKCORE_LIVE=1 (even with CI unset),
// a missing harness binary fails TestLiveBothCLIsInstantiateAgentsWithValidTools closed rather
// than skipping.
func TestHarnessCLIFailsClosedUnderHookcoreLive(t *testing.T) {
	if os.Getenv("PLUGINPARITY_FAILCLOSED_SUBPROC") == "1" {
		t.Skip("inside fail-closed subprocess")
	}
	emptyHome := testhome.NewLiveHome(t)
	emptyPath := t.TempDir()
	base := testhome.LiveEnv(t, emptyHome,
		"HOOKCORE_LIVE=1",
		"PLUGINPARITY_FAILCLOSED_SUBPROC=1",
	)
	env := make([]string, 0, len(base)+2)
	for _, e := range base {
		k, _, _ := strings.Cut(e, "=")
		if k == "CI" || k == "PATH" || (runtime.GOOS == "windows" && strings.EqualFold(k, "PATH")) {
			continue
		}
		env = append(env, e)
	}
	env = append(env, "CI=", "PATH="+emptyPath)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.v", "-test.run", "^TestLiveBothCLIsInstantiateAgentsWithValidTools$")
	cmd.Env = env
	outBytes, err := cmd.CombinedOutput()
	out := string(outBytes)
	if err == nil {
		t.Fatalf("expected non-zero exit when harness CLI is missing under HOOKCORE_LIVE=1:\n%s", out)
	}
	if strings.Contains(out, "--- SKIP:") {
		t.Fatalf("expected fail-closed FAIL, not SKIP:\n%s", out)
	}
	if !strings.Contains(out, "--- FAIL: TestLiveBothCLIsInstantiateAgentsWithValidTools") {
		t.Fatalf("output missing FAIL line:\n%s", out)
	}
}
