package hookcore_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/shared/testhome"
)

// TestLiveFailClosedOnMissingBinaryAndCredentials drives the real HOOKCORE_LIVE test functions via
// the compiled test binary (os.Args[0]) in an isolated empty HOME and controlled PATH, pinning
// Fatalf (never Skip) on missing binaries and missing credentials (closes S1, S2, S3).
func TestLiveFailClosedOnMissingBinaryAndCredentials(t *testing.T) {
	if os.Getenv("HOOKCORE_LIVE_SUBPROC") == "1" {
		t.Skip("inside fail-closed subprocess")
	}
	stubBinDir := func(t *testing.T, names ...string) string {
		t.Helper()
		dir := t.TempDir()
		for _, name := range names {
			if runtime.GOOS == "windows" {
				name += ".exe"
			}
			p := filepath.Join(dir, name)
			if err := os.WriteFile(p, []byte("#!/bin/sh\nexit 1\n"), 0o755); err != nil {
				t.Fatal(err)
			}
		}
		return dir
	}

	cases := []struct {
		name       string
		runPattern string
		stubBins   []string
		wantFail   string
		wantSubstr string
	}{
		{
			name:       "TestLiveClaudeContract missing binary fails closed (S1)",
			runPattern: "^TestLiveClaudeContract$",
			wantFail:   "--- FAIL: TestLiveClaudeContract",
			wantSubstr: "claude binary not found",
		},
		{
			name:       "TestLiveClaudeContract missing credentials fails closed (S2)",
			runPattern: "^TestLiveClaudeContract$",
			stubBins:   []string{"claude"},
			wantFail:   "--- FAIL: TestLiveClaudeContract",
			wantSubstr: ".credentials.json",
		},
		{
			name:       "TestLiveCrossClientSubagentLifecycleAndOverwrite/Claude missing binary fails closed (S3)",
			runPattern: "^TestLiveCrossClientSubagentLifecycleAndOverwrite$/^Claude$",
			wantFail:   "--- FAIL: TestLiveCrossClientSubagentLifecycleAndOverwrite/Claude",
			wantSubstr: "claude CLI not found",
		},
		{
			name:       "TestLiveAntigravityContract missing binary fails closed",
			runPattern: "^TestLiveAntigravityContract$",
			wantFail:   "--- FAIL: TestLiveAntigravityContract",
			wantSubstr: "agy binary not found",
		},
		{
			name:       "TestLiveAntigravityContract missing credentials fails closed",
			runPattern: "^TestLiveAntigravityContract$",
			stubBins:   []string{"agy"},
			wantFail:   "--- FAIL: TestLiveAntigravityContract",
			wantSubstr: "antigravity-oauth-token",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			emptyHome := testhome.NewLiveHome(t)
			pathDir := stubBinDir(t, tc.stubBins...)
			env := replaceEnvKey(testhome.LiveEnv(t, emptyHome,
				"HOOKCORE_LIVE=1",
				"HOOKCORE_LIVE_SUBPROC=1",
			), "PATH", pathDir)

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, os.Args[0], "-test.v", "-test.run", tc.runPattern)
			cmd.Env = env
			outBytes, err := cmd.CombinedOutput()
			out := string(outBytes)
			if err == nil {
				t.Fatalf("expected non-zero exit under HOOKCORE_LIVE=1, got exit 0:\n%s", out)
			}
			if strings.Contains(out, "--- SKIP:") {
				t.Fatalf("expected fail-closed FAIL, not SKIP:\n%s", out)
			}
			if !strings.Contains(out, tc.wantFail) {
				t.Fatalf("output missing %q:\n%s", tc.wantFail, out)
			}
			if !strings.Contains(out, tc.wantSubstr) {
				t.Fatalf("output missing %q:\n%s", tc.wantSubstr, out)
			}
		})
	}
}

func replaceEnvKey(env []string, key, val string) []string {
	prefix := key + "="
	out := make([]string, 0, len(env)+1)
	for _, e := range env {
		if strings.HasPrefix(e, prefix) || (runtime.GOOS == "windows" && strings.EqualFold(e[:min(len(e), len(prefix))], prefix)) {
			continue
		}
		out = append(out, e)
	}
	return append(out, prefix+val)
}
