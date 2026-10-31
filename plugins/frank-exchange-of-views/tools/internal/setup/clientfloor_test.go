package setup

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// THE CLIENT FLOOR IS REFUSED BEFORE A RUN EXISTS. Below it the engine would run on a client whose
// resume repeats a finished parallel group, and a client whose version cannot be read refuses the same
// way: nothing is checked by a number nothing read.
func TestPreflightClientHoldsTheFloor(t *testing.T) {
	answer := func(r ExecResult) ExecFunc {
		return func(name string, args []string) ExecResult {
			if name != clientBin || len(args) != 1 || args[0] != "--version" {
				t.Errorf("asked %s %v, want the client's --version", name, args)
			}
			return r
		}
	}
	for _, tc := range []struct {
		name    string
		reply   ExecResult
		ok      bool
		version string
		reason  string
	}{
		{"at the floor", ExecResult{Stdout: "2.1.305 (Claude Code)\n"}, true, "2.1.305", ""},
		{"above it", ExecResult{Stdout: "2.1.309 (Claude Code)\n"}, true, "2.1.309", ""},
		{"above it by a longer number", ExecResult{Stdout: "2.10.0\n"}, true, "2.10.0", ""},
		{"a bare major above it", ExecResult{Stdout: "3\n"}, true, "3", ""},
		{"below it", ExecResult{Stdout: "2.1.304 (Claude Code)\n"}, false, "", "2.1.304 is below the 2.1.305"},
		{"below it by a number string order reads above", ExecResult{Stdout: "2.1.99\n"}, false, "", "2.1.99 is below"},
		{"will not run", ExecResult{Err: errors.New("executable file not found")}, false, "", "did not run (executable file not found)"},
		{"exits non-zero", ExecResult{Status: 3}, false, "", "did not run (exit 3)"},
		{"prints nothing", ExecResult{}, false, "", "names no version"},
		{"prints another thing", ExecResult{Stdout: "usage: claude\n"}, false, "", "names no version"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := PreflightClient("2.1.305", answer(tc.reply))
			if got.OK != tc.ok || got.Version != tc.version || !strings.Contains(got.Reason, tc.reason) {
				t.Errorf("= %+v, want ok %v, version %q, a reason naming %q", got, tc.ok, tc.version, tc.reason)
			}
			if !got.OK && got.Remedy == "" {
				t.Error("a refusal with no remedy")
			}
		})
	}
}

// SETUP REFUSES A CLIENT BELOW THE FLOOR AND CREATES NOTHING, and refuses a plugin that states none:
// a manifest with no floor leaves nothing to check the client against.
func TestSetupRefusesTheClientAndCreatesNothing(t *testing.T) {
	refused := func(t *testing.T, cfg Config, runDir, want string) {
		t.Helper()
		var out, errb bytes.Buffer
		if code := Run(cfg, &out, &errb); code != 2 {
			t.Fatalf("exit %d, want 2:\n%s", code, errb.String())
		}
		if !strings.Contains(errb.String(), "CLIENT PREFLIGHT FAILED") || !strings.Contains(errb.String(), want) {
			t.Errorf("the refusal does not say %q:\n%s", want, errb.String())
		}
		if _, err := os.Stat(filepath.Join(runDir, "inputs")); !os.IsNotExist(err) {
			t.Errorf("the refusal left the run's inputs behind (%v)", err)
		}
	}
	t.Run("below the floor", func(t *testing.T) {
		good := reports(strconv.Itoa(record.EventSchema))
		cfg, runDir := runCfg(t, func(name string, args []string) ExecResult {
			if name == clientBin {
				return ExecResult{Stdout: "2.1.300 (Claude Code)\n"}
			}
			return good(name, args)
		})
		refused(t, cfg, runDir, "2.1.300 is below the 2.1.305")
	})
	t.Run("no floor on the manifest", func(t *testing.T) {
		cfg, runDir := runCfg(t, reports(strconv.Itoa(record.EventSchema)))
		manifest := filepath.Join(filepath.Dir(cfg.BinDir), "requirements.json")
		if err := os.WriteFile(manifest, fmt.Appendf(nil, `{"eventSchema":%d,"tools":[{"name":"node","min_version":"20"}]}`, record.EventSchema), 0o644); err != nil {
			t.Fatal(err)
		}
		refused(t, cfg, runDir, "declares no min_version for claude")
	})
}
