package toolchain

import (
	"os"
	"path/filepath"
	"testing"
)

// THE SHAPES ARE NOT UNIFORM AND NEVER WILL BE. A parser per tool is a table that goes stale
// silently; these are the real outputs of the tools this repository declares.
func TestExtractVersionReadsTheToolsWeActuallyDeclare(t *testing.T) {
	for _, tc := range []struct{ out, want string }{
		{"go version go1.22.2 linux/amd64", "1.22.2"},
		{"go version go1.25.0 darwin/arm64", "1.25.0"},
		{"git version 2.43.0", "2.43.0"},
		{"gh version 2.97.0 (2026-07-31)", "2.97.0"},
		{"jq-1.7", "1.7"},
		{"Python 3.11.15", "3.11.15"},
		{"qlty 0.577.0", "0.577.0"},
		{"v22.22.2", "22.22.2"},
		// A bare integer is not a version, and skipping it must not strand the real one.
		{"tool 3 build 1.2.3", "1.2.3"},
		// Sentence punctuation is not a component.
		{"ready at 1.2.", "1.2"},
		// NOT MEASURED is a real answer and must be distinguishable from a low version.
		{"some tool with no numbers", ""},
		{"", ""},
	} {
		if got := ExtractVersion(tc.out); got != tc.want {
			t.Errorf("ExtractVersion(%q) = %q, want %q", tc.out, got, tc.want)
		}
	}
}

// A GO DIRECTIVE IS WRITTEN `1.25` AND NO RELEASE IS EVER CALLED THAT — releases are 1.25.N.
// A comparison that treated the missing component as anything but zero would report every
// 1.25.x as below a 1.25 minimum, which is the check firing on the machines it should pass.
func TestCompareVersionsTreatsAMissingComponentAsZero(t *testing.T) {
	for _, tc := range []struct {
		a, b string
		want int
	}{
		{"1.25", "1.25.0", 0},
		{"1.25.0", "1.25", 0},
		{"1.25.3", "1.25", 1},
		{"1.22.2", "1.25", -1},
		{"2.0", "1.99.99", 1},
		{"1.7", "1.7", 0},
		{"0.577.0", "0.500", 1},
		// Pre-release text is ignored rather than ranked — a coarse honest answer over a
		// confident wrong one.
		{"1.25.0-rc1", "1.25.0", 0},
	} {
		if got := CompareVersions(tc.a, tc.b); got != tc.want {
			t.Errorf("CompareVersions(%q, %q) = %d, want %d", tc.a, tc.b, got, tc.want)
		}
	}
}

// PRESENT BUT UNUSABLE IS ITS OWN STATE, and until now it was ✓ with the disqualifying number
// printed helpfully beside it (#521).
func TestATooOldToolIsNeitherFoundNorMissing(t *testing.T) {
	restore := versionOutput
	defer func() { versionOutput = restore }()
	versionOutput = func(string, string) (string, error) { return "go version go1.22.2 linux/amd64", nil }

	got := ProbeIn([]Tool{{Name: "go", Tier: "required", CheckCmd: "go version", MinVersion: "1.25"}}, EnvLocal)[0]
	if !got.TooOld {
		t.Errorf("go1.22.2 against a 1.25 minimum reported TooOld=false — the number was read and ignored")
	}
	if got.Version != "1.22.2" {
		t.Errorf("Version = %q, want the parsed 1.22.2 so the row can say what is actually installed", got.Version)
	}
	if got.VersionUnmeasured != "" {
		t.Errorf("a version that WAS read reports unmeasured: %q", got.VersionUnmeasured)
	}
}

// AND AN UNREADABLE VERSION IS NOT A PASS. facts-are-fields clause 3: if the miss renders as
// the healthy string, the read is not a read.
func TestAnUnreadableVersionIsStatedNotAssumed(t *testing.T) {
	restore := versionOutput
	defer func() { versionOutput = restore }()

	for _, tc := range []struct {
		name string
		out  string
		err  error
	}{
		{"output with no number in it", "go: unknown subcommand", nil},
		{"the command did not run at all", "", errNotRun{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			versionOutput = func(string, string) (string, error) { return tc.out, tc.err }
			got := ProbeIn([]Tool{{Name: "go", Tier: "required", CheckCmd: "go version", MinVersion: "1.25"}}, EnvLocal)[0]
			if got.VersionUnmeasured == "" {
				t.Fatalf("an unreadable version reported no reason (TooOld=%v, Version=%q)", got.TooOld, got.Version)
			}
			if got.TooOld {
				t.Errorf("an unread version was ALSO reported too old — that is a verdict on a number nobody has")
			}
		})
	}
}

type errNotRun struct{}

func (errNotRun) Error() string { return "executable file not found in $PATH" }

// NO MINIMUM MEANS NO COST. Declaring one is what runs the command; every other tool stays a
// single LookPath, which is what it was before this existed.
func TestAToolWithNoMinimumIsNeverInterrogated(t *testing.T) {
	restore := versionOutput
	defer func() { versionOutput = restore }()
	ran := 0
	versionOutput = func(string, string) (string, error) { ran++; return "git version 2.43.0", nil }

	ProbeIn([]Tool{{Name: "git", Tier: "required", CheckCmd: "git --version"}}, EnvLocal)
	if ran != 0 {
		t.Errorf("the check command ran %d time(s) for a tool declaring no minimum", ran)
	}
}

// A TOOL THAT DOES NOT APPLY HERE IS NOT INTERROGATED EITHER. The manifest has already said its
// absence is not a defect; a minimum it cannot meet is not one either.
func TestANotApplicableToolIsNotVersionChecked(t *testing.T) {
	restore := versionOutput
	defer func() { versionOutput = restore }()
	ran := 0
	versionOutput = func(string, string) (string, error) { ran++; return "gh version 1.0.0", nil }

	got := ProbeIn([]Tool{{
		Name: "gh", Tier: "recommended", CheckCmd: "gh --version",
		MinVersion: "2.0", NotApplicableIn: []string{EnvRemote},
	}}, EnvRemote)[0]
	if ran != 0 {
		t.Errorf("a not-applicable tool was version-checked %d time(s)", ran)
	}
	if got.TooOld {
		t.Error("a not-applicable tool was reported too old")
	}
}

// THE ANSWER DEPENDS ON WHERE YOU ASK IT, and the first real run of this check proved it the
// hard way: sc-doctor reported BLOCKED on a machine where every gate passes.
//
// GOTOOLCHAIN=auto selects a toolchain PER MODULE from that module's own `go` directive, so on
// one container `go version` says go1.24.7 at the repository root and go1.25.0 inside `tools/`.
// The number was real both times; the location made one of them meaningless. CheckDir is what
// moves the question to the place whose answer matters.
func TestAVersionQuestionIsAskedWhereItsAnswerMatters(t *testing.T) {
	restore := versionOutput
	defer func() { versionOutput = restore }()
	var askedIn string
	versionOutput = func(_, dir string) (string, error) {
		askedIn = dir
		if dir == "" {
			return "go version go1.24.7 linux/amd64", nil // the root, where the directive does not apply
		}
		return "go version go1.25.0 linux/amd64", nil // inside the module
	}

	base := t.TempDir()
	if err := os.MkdirAll(filepath.Join(base, "tools"), 0o755); err != nil {
		t.Fatal(err)
	}
	tool := Tool{Name: "go", Tier: "required", CheckCmd: "go version", MinVersion: "1.25", CheckDir: "tools"}

	got := ProbeInDir([]Tool{tool}, EnvLocal, base)[0]
	if askedIn != filepath.Join(base, "tools") {
		t.Errorf("the check ran in %q, want the declared %q", askedIn, filepath.Join(base, "tools"))
	}
	if got.TooOld {
		t.Errorf("go1.25.0 read inside the module was reported too old — this is the false BLOCKED that prompted the field")
	}

	// And with no base directory the old behaviour stands: the caller's own cwd.
	askedIn = "sentinel"
	if got = ProbeInDir([]Tool{tool}, EnvLocal, "")[0]; askedIn != "" {
		t.Errorf("with no base directory the check ran in %q, want the caller's cwd", askedIn)
	}
}

// A DECLARED DIRECTORY THAT IS NOT THERE MAKES THE ANSWER UNKNOWABLE, not old. Falling back to
// the cwd would produce a confident number about the wrong place, which is the defect the field
// exists to fix, arrived at through the field itself.
func TestAMissingCheckDirIsUnmeasuredNotAssumed(t *testing.T) {
	restore := versionOutput
	defer func() { versionOutput = restore }()
	ran := 0
	versionOutput = func(_, _ string) (string, error) { ran++; return "go version go1.24.7 linux/amd64", nil }

	got := ProbeInDir([]Tool{{
		Name: "go", Tier: "required", CheckCmd: "go version", MinVersion: "1.25", CheckDir: "tools",
	}}, EnvLocal, t.TempDir())[0]

	if ran != 0 {
		t.Errorf("the check ran %d time(s) from the wrong directory", ran)
	}
	if got.TooOld {
		t.Error("an unreachable check directory was reported as a version below the minimum")
	}
	if got.VersionUnmeasured == "" {
		t.Error("an unreachable check directory reported no reason — it renders as a pass")
	}
}
