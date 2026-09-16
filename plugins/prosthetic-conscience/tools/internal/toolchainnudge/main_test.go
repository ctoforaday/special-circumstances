package toolchainnudge

import (
	"fmt"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/hookfailures"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/toolchain"
)

func status(name, tier string, found bool) toolchain.Status {
	return toolchain.Status{Tool: toolchain.Tool{Name: name, Tier: tier}, Found: found}
}

func TestNudge(t *testing.T) {
	cases := []struct {
		name     string
		statuses []toolchain.Status
		want     []string // substrings; empty slice = expect silence
	}{
		{"all present is silent", []toolchain.Status{status("git", "required", true), status("qlty", "recommended", true)}, nil},
		{"missing recommended named", []toolchain.Status{status("git", "required", true), status("qlty", "recommended", false)}, []string{"qlty", "doctor"}},
		{"missing optional is silent", []toolchain.Status{status("jq", "optional", false)}, nil},
		{"multiple missing listed", []toolchain.Status{status("gh", "recommended", false), status("qlty", "recommended", false)}, []string{"gh", "qlty"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := nudge(c.statuses)
			if len(c.want) == 0 {
				if got != "" {
					t.Fatalf("want silence, got %q", got)
				}
				return
			}
			for _, w := range c.want {
				if !strings.Contains(got, w) {
					t.Fatalf("nudge() = %q; missing %q", got, w)
				}
			}
			if strings.Count(got, "\n") > 0 {
				t.Fatalf("nudge must be one line: %q", got)
			}
		})
	}
}

// SessionStart is the most expensive place in the suite for a warning nobody can act
// on: it is the first thing every session shows, every time. A cloud session ships no
// gh by design, so nudging about it there is pure noise — and noise at line one is how
// an operator learns to skip the line that eventually matters.
func TestNudgeSkipsNotApplicableTools(t *testing.T) {
	na := func(name, tier string) toolchain.Status {
		s := status(name, tier, false)
		s.NotApplicable = true
		return s
	}
	if got := nudge([]toolchain.Status{status("git", "required", true), na("gh", "recommended")}); got != "" {
		t.Fatalf("want silence for a by-design absence, got %q", got)
	}
	got := nudge([]toolchain.Status{na("gh", "recommended"), status("qlty", "recommended", false)})
	if !strings.Contains(got, "qlty") {
		t.Fatalf("a genuinely missing tool must still be named: %q", got)
	}
	if strings.Contains(got, "gh") {
		t.Fatalf("the exempt tool leaked into the nudge: %q", got)
	}
}

// "quality hooks degrade" told an agent nothing it could weigh (#212). The line must name
// WHICH capability is absent, and say whether to carry on — otherwise the agent either
// stops unnecessarily or treats a clean run as proof the checks passed.
func TestNudgeNamesTheCapabilityAndWhetherToContinue(t *testing.T) {
	got := nudge([]toolchain.Status{
		{Tool: toolchain.Tool{Name: "qlty", Purpose: "format + lint quality gate", Tier: "recommended"}, Found: false},
		{Tool: toolchain.Tool{Name: "git", Purpose: "version control", Tier: "required"}, Found: true},
	})
	if got == "" {
		t.Fatal("a missing recommended tool must be announced")
	}
	for _, want := range []string{
		"qlty",                       // which tool
		"format + lint quality gate", // which CAPABILITY is absent
		"SKIPPED",                    // what did not happen
		"nothing is blocked",         // whether to carry on
		"do not read a clean run",    // and what the silence does NOT prove
		"doctor --fix",               // the exit that exists
	} {
		if !strings.Contains(got, want) {
			t.Errorf("nudge is missing %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "git") {
		t.Errorf("a tool that IS present must not be named: %s", got)
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
