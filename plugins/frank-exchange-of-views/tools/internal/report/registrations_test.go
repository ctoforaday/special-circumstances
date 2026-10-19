package report

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"google.golang.org/protobuf/proto"
)

// stamped is one register as a seat's first act writes it: the seat, then tool_version,
// hook_version and run_via, each "" for a field the register does not carry.
type stamped [4]string

// stampedRun seeds a run with the registers and, when config is not "", writes it as the run's
// inputs/run-config.json. It returns the run.md the run assembles to.
func stampedRun(t *testing.T, config string, regs ...stamped) string {
	t.Helper()
	dir := recordtest.TmpRun(t)
	var evs []*record.Event
	for _, r := range regs {
		reg := &recordpb.Register{}
		if r[1] != "" {
			reg.ToolVersion = proto.String(r[1])
		}
		if r[2] != "" {
			reg.HookVersion = proto.String(r[2])
		}
		if r[3] != "" {
			reg.RunVia = proto.String(r[3])
		}
		evs = append(evs, recordtest.Event(t, r[0], reg))
	}
	recordtest.Seed(t, dir, evs...)
	if config != "" {
		if err := os.MkdirAll(filepath.Join(dir, "inputs"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "inputs", "run-config.json"), []byte(config), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	docs, err := AssembleAll(runtest.Open(t, dir))
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range docs {
		if d.File == FileRun {
			return d.Body
		}
	}
	t.Fatal("the set ships no run.md")
	return ""
}

// buildsLine is the paragraph that opens with the builds-and-hooks label, through the end of the
// section's list of exceptions.
func buildsLine(t *testing.T, runMD string) string {
	t.Helper()
	_, after, ok := strings.Cut(runMD, "**Builds and hooks")
	if !ok {
		t.Fatalf("run.md never states the builds and hooks the seats registered with:\n%s", runMD)
	}
	body, _, _ := strings.Cut(after, "\n## ")
	return "**Builds and hooks" + strings.TrimRight(body, "\n ")
}

const installedOne = `{"hooks":{"state":"installed","versions":["1.73.0"],"binaries":{"1.73.0":["feov-pretooluse"]},"missing":{"1.73.0":[]},"why":""}}`

// A UNIFORM RUN IS ONE LINE. Every seat on one build, one hook build and the hook's route, with
// setup's provenance beside it: nothing to read past the sentence that says so.
func TestAUniformRunStatesItsBuildsAndHooksInOneLine(t *testing.T) {
	got := buildsLine(t, stampedRun(t, installedOne,
		stamped{"blue-lane-1", "abc1234", "def5678", "injected"},
		stamped{"red-lens-evidence", "abc1234", "def5678", "injected"},
		stamped{"red-chair", "abc1234", "def5678", "injected"},
	))
	want := "**Builds and hooks.** All 3 registrations, from 3 seats, carry one stamp: feov-record build `abc1234`, " +
		"hook build `def5678`, run directory from the hook. Setup found the hooks installed, cached as version 1.73.0."
	if got != want {
		t.Errorf("the uniform run's line is\n%s\nwant\n%s", got, want)
	}
}

// SEATS THAT DISAGREE ARE NAMED, with what differs in bold — the diagnosis #751 had to make by
// tabulating the database by hand. A mixed binary, a seat no hook reached and a seat that found the
// run by inference are each a line, and setup's "installed" beside an unstamped registration is
// called out as the contradiction it is.
func TestSeatsThatDisagreeOnTheirStampAreNamed(t *testing.T) {
	got := buildsLine(t, stampedRun(t, installedOne,
		stamped{"blue-lane-1", "abc1234", "def5678", "injected"},
		stamped{"red-chair", "abc1234", "def5678", "injected"},
		stamped{"red-lens-evidence", "9999999", "def5678", "injected"},
		stamped{"blue-lane-2", "abc1234", "", "wrapper"},
		stamped{"red-lens-logic", "abc1234", "0000000", "inferred"},
		stamped{"red-lens-logic", "abc1234", "0000000", "inferred"},
	))
	for _, want := range []string{
		"**Builds and hooks — NOT UNIFORM.** 2 of 6 registrations carry the run's most common stamp: feov-record build `abc1234`, hook build `def5678`, run directory from the hook.",
		"- `red-lens-evidence`: **feov-record build `9999999`**, hook build `def5678`, run directory from the hook",
		"- `blue-lane-2`: feov-record build `abc1234`, **no hook stamp — no hook served the seat, or one too old to stamp it**, **run directory from the per-run wrapper — no hook reached the seat**",
		"- `red-lens-logic` (2 registrations): feov-record build `abc1234`, **hook build `0000000`**, **run directory found by inference — neither the hook nor the wrapper supplied it**",
		"**Yet 1 of 6 registrations carry no hook stamp.**",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("the disagreement does not state %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "- `red-chair`") || strings.Contains(got, "- `blue-lane-1`") {
		t.Errorf("a seat on the run's common stamp is listed as an exception:\n%s", got)
	}
}

// SETUP'S PROVENANCE AND THE STAMPS ARE SET SIDE BY SIDE, and each way they can contradict says so.
func TestSetupsHookProvenanceIsComparedWithTheStamps(t *testing.T) {
	for name, c := range map[string]struct {
		config string
		want   string
	}{
		"setup saw no hooks, a hook stamped the seats": {
			`{"hooks":{"state":"absent","versions":[],"binaries":{},"missing":{},"why":"no cached copy of frank-exchange-of-views"}}`,
			"**Setup found the hooks absent:** no cached copy of frank-exchange-of-views. **Yet 2 of 2 registrations were stamped by a hook**",
		},
		"a cached version lacks a binary it registers": {
			`{"hooks":{"state":"installed","versions":["1.72.0","1.73.0"],"binaries":{},"missing":{"1.72.0":[],"1.73.0":["feov-pretooluse"]},"why":""}}`,
			"cached as versions 1.72.0 and 1.73.0; which of them served a seat is not something setup can see. **A cached version registers hook binaries it does not ship:** 1.73.0 registers feov-pretooluse.",
		},
		"no provenance in the configuration": {
			`{"topic":"t"}`,
			"**Setup recorded nothing about the installed hooks**",
		},
		"no configuration": {
			"",
			"**Setup recorded nothing about the installed hooks**",
		},
	} {
		t.Run(name, func(t *testing.T) {
			got := buildsLine(t, stampedRun(t, c.config,
				stamped{"blue-lane-1", "abc1234", "def5678", "injected"},
				stamped{"red-chair", "abc1234", "def5678", "injected"},
			))
			if !strings.Contains(got, c.want) {
				t.Errorf("run.md does not state %q:\n%s", c.want, got)
			}
			if !strings.Contains(got, "NOT UNIFORM") {
				t.Errorf("a contradiction with setup's provenance does not mark the stamps as not uniform:\n%s", got)
			}
		})
	}
}

// AN UNSTAMPED BUILD IS NOT AGREEMENT. Two binaries built outside a git tree both register
// `unknown`, so a run where every seat says `unknown` has shown nothing about whether its binaries
// matched, and the line says so.
func TestAnUnstampedBuildIsSaidToProveNothingAboutAgreement(t *testing.T) {
	got := buildsLine(t, stampedRun(t, installedOne,
		stamped{"blue-lane-1", "unknown", "unknown", "injected"},
		stamped{"red-chair", "unknown", "unknown", "injected"},
	))
	if !strings.Contains(got, "`unknown` is a build without a commit stamp, so registrations that agree on it may still have run different binaries.") {
		t.Errorf("a run of unstamped builds reads as agreement:\n%s", got)
	}
}
