package report

import (
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

// stampedRun seeds a run with the registers and returns the run.md the run assembles to.
func stampedRun(t *testing.T, regs ...stamped) string {
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
	docs, err := shippedDocs(runtest.Open(t, dir))
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

// A UNIFORM RUN IS ONE LINE. Every seat on one build, served by a hook of that same build, by the
// hook's route: nothing to read past the sentence that says so, and nothing about what was installed.
func TestAUniformRunStatesItsBuildsAndHooksInOneLine(t *testing.T) {
	got := buildsLine(t, stampedRun(t,
		stamped{"blue-lane-1", "abc1234", "abc1234", "injected"},
		stamped{"red-lens-evidence", "abc1234", "abc1234", "injected"},
		stamped{"red-chair", "abc1234", "abc1234", "injected"},
	))
	want := "**Builds and hooks.** All 3 registrations, from 3 seats, carry one stamp: feov-record build `abc1234`, " +
		"hook build `abc1234`, run directory from the hook."
	if got != want {
		t.Errorf("the uniform run's line is\n%s\nwant\n%s", got, want)
	}
}

// SEATS THAT DISAGREE ARE NAMED, with what differs in bold — the diagnosis #751 had to make by
// tabulating the database by hand. A mixed binary, a seat no hook reached that typed its --run, and a
// seat that found the run by inference are each a line.
func TestSeatsThatDisagreeOnTheirStampAreNamed(t *testing.T) {
	got := buildsLine(t, stampedRun(t,
		stamped{"blue-lane-1", "abc1234", "abc1234", "injected"},
		stamped{"red-chair", "abc1234", "abc1234", "injected"},
		stamped{"red-lens-evidence", "9999999", "abc1234", "injected"},
		stamped{"blue-lane-2", "abc1234", "", "flag"},
		stamped{"red-lens-logic", "abc1234", "abc1234", "inferred"},
		stamped{"red-lens-logic", "abc1234", "abc1234", "inferred"},
	))
	for _, want := range []string{
		"**Builds and hooks — NOT UNIFORM, MIXED BUILDS.** 2 of 6 registrations carry the run's most common stamp: feov-record build `abc1234`, hook build `abc1234`, run directory from the hook.",
		"- `red-lens-evidence`: **feov-record build `9999999`**, hook build `abc1234`, run directory from the hook",
		"- `blue-lane-2`: feov-record build `abc1234`, **no hook stamp — no hook served the seat, or one too old to stamp it**, **run directory from a `--run` the seat typed**",
		"- `red-lens-logic` (2 registrations): feov-record build `abc1234`, hook build `abc1234`, **run directory found by inference — the hook did not supply it**",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("the disagreement does not state %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "- `red-chair`") || strings.Contains(got, "- `blue-lane-1`") {
		t.Errorf("a seat on the run's common stamp is listed as an exception:\n%s", got)
	}
}

// A REGISTRATION WHOSE HOOK BUILD IS NOT ITS FEOV-RECORD BUILD IS NAMED, by seat and registration
// (#751): the record binary and the hooks came from different builds. The registration that agrees
// with itself is not named, and a registration no hook stamped has no hook build to compare.
func TestARegistrationWhoseHookBuildDiffersFromItsRecordBuildIsNamed(t *testing.T) {
	got := buildsLine(t, stampedRun(t,
		stamped{"blue-lane-1", "abc1234", "abc1234", "injected"},
		stamped{"red-chair", "abc1234", "abc1234", "injected"},
		stamped{"red-lens-evidence", "9999999", "abc1234", "injected"},
		stamped{"red-lens-evidence", "9999999", "abc1234", "injected"},
		stamped{"blue-lane-2", "abc1234", "", "flag"},
	))
	want := "2 of 5 registrations ran a hook build other than their own feov-record build, so the hooks and the record binary came from different builds:\n\n" +
		"- `red-lens-evidence` (2 registrations): feov-record build `9999999`, hook build `abc1234`"
	if !strings.HasSuffix(got, want) {
		t.Errorf("the section does not end naming the one seat whose hook build differs from its record build:\n%s\nwant it to end\n%s", got, want)
	}
	if !strings.HasPrefix(got, "**Builds and hooks — NOT UNIFORM, MIXED BUILDS.**") {
		t.Errorf("the mixed registration does not head the section as MIXED BUILDS:\n%s", got)
	}
}

// THE MIXED-ROOTS RUN #751 FOUND AGREES WITH ITSELF: every seat ran one record binary and one hook,
// from two builds. No registration differs from another, so only the comparison of each
// registration's two builds says so.
func TestARunWhoseEveryRegistrationMixesBuildsIsNamed(t *testing.T) {
	got := buildsLine(t, stampedRun(t,
		stamped{"blue-lane-1", "abc1234+dirty", "def5678", "injected"},
		stamped{"red-chair", "abc1234+dirty", "def5678", "injected"},
	))
	want := "**Builds and hooks — MIXED BUILDS.** All 2 registrations, from 2 seats, carry one stamp: feov-record build `abc1234+dirty`, hook build `def5678`, run directory from the hook.\n\n" +
		"2 of 2 registrations ran a hook build other than their own feov-record build, so the hooks and the record binary came from different builds:\n\n" +
		"- `blue-lane-1`: feov-record build `abc1234+dirty`, hook build `def5678`\n" +
		"- `red-chair`: feov-record build `abc1234+dirty`, hook build `def5678`"
	if got != want {
		t.Errorf("the mixed-roots run reads\n%s\nwant\n%s", got, want)
	}
}

// AN UNSTAMPED RECORD BINARY UNDER A STAMPED HOOK IS A MIX. `unknown` is a build with no commit
// stamp, and a hook that carries one is another build: the 09-10 b3 run's evidence lens ran a
// record binary built outside a git tree under the run's hook.
func TestAnUnstampedRecordBuildUnderAStampedHookIsNamed(t *testing.T) {
	got := buildsLine(t, stampedRun(t,
		stamped{"red-chair", "5679a6e+dirty", "5679a6e+dirty", "injected"},
		stamped{"red-lens-evidence", "unknown", "5679a6e+dirty", "injected"},
	))
	want := "- `red-lens-evidence`: feov-record build `unknown`, hook build `5679a6e+dirty`"
	if !strings.HasSuffix(got, want) || !strings.Contains(got, "MIXED BUILDS") {
		t.Errorf("an unknown record build under a stamped hook is not named as a mix:\n%s\nwant it to end\n%s", got, want)
	}
}

// A RUN NO HOOK STAMPED COMPARES NO BUILDS. Its registrations say they carry no hook stamp, and the
// section claims neither agreement nor a mix.
func TestARunNoHookStampedComparesNoBuilds(t *testing.T) {
	got := buildsLine(t, stampedRun(t,
		stamped{"blue-lane-1", "abc1234", "", "flag"},
		stamped{"red-chair", "abc1234", "", "flag"},
	))
	want := "**Builds and hooks.** All 2 registrations, from 2 seats, carry one stamp: feov-record build `abc1234`, " +
		"no hook stamp — no hook served the seat, or one too old to stamp it, run directory from a `--run` the seat typed."
	if got != want {
		t.Errorf("the unstamped run reads\n%s\nwant\n%s", got, want)
	}
}

// AN UNSTAMPED BUILD IS NOT AGREEMENT. Two binaries built outside a git tree both register
// `unknown`, so a run where every seat says `unknown` has shown nothing about whether its binaries
// matched, and the line says so.
func TestAnUnstampedBuildIsSaidToProveNothingAboutAgreement(t *testing.T) {
	got := buildsLine(t, stampedRun(t,
		stamped{"blue-lane-1", "unknown", "unknown", "injected"},
		stamped{"red-chair", "unknown", "unknown", "injected"},
	))
	if !strings.Contains(got, "`unknown` is a build without a commit stamp, so registrations that agree on it may still have run different binaries.") {
		t.Errorf("a run of unstamped builds reads as agreement:\n%s", got)
	}
}
