package doctor

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/buildid"
)

// noFix stands in for fix(): the real one reaches the network (gh release download)
// and the local Go toolchain. No test in this package may call it.
func noFix(calls *[]binStatus) func([]binStatus) []string {
	return func(bins []binStatus) []string {
		*calls = append(*calls, bins...)
		return []string{"fix ran"}
	}
}

func noExecutable() (string, error) { return "", errors.New("no executable") }

// absentTool is a manifest entry whose check command cannot resolve on any machine,
// so the probe is deterministic wherever this test runs.
func manifestJSON(tier string) string {
	return fmt.Sprintf(`{"tools":[{"name":"phantom","tier":%q,"check_cmd":"sc-definitely-not-real-xyz --version",`+
		`"install":{"windows":"winget install phantom","darwin":"brew install phantom","linux":"apt install phantom"}}]}`, tier)
}

// doctorRoot builds <tmp>/<plugin>/<version>/ with a manifest and a tools/cmd/<name>
// layout, i.e. the marketplace-cache shape the doctor is written against. bin/ is
// created non-empty so the fixture is not itself an EMPTY-BIN WINDOW — that warning
// is exercised deliberately in TestRunDanceWarningDowngradesReady.
func doctorRoot(t *testing.T, manifest string, cmds ...string) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), "prosthetic-conscience", "0.9.1")
	if err := os.MkdirAll(filepath.Join(root, "bin"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "bin", ".keep"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if manifest != "" {
		if err := os.WriteFile(filepath.Join(root, "requirements.json"), []byte(manifest), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, c := range cmds {
		if err := os.MkdirAll(filepath.Join(root, "tools", "cmd", c), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func doctor(t *testing.T, root string, args ...string) (string, int) {
	t.Helper()
	var out bytes.Buffer
	var fixed []binStatus
	code := run(append([]string{"-root", root}, args...), &out, "", noExecutable, noFix(&fixed))
	return out.String(), code
}

// The contract: a deterministic table plus exactly one verdict line, always exit 0 —
// a preflight reports, it does not fail the caller.
func TestRunVerdicts(t *testing.T) {
	cases := []struct {
		name     string
		tier     string
		cmds     []string
		want     string
		wantLine []string
	}{
		{"missing required tool blocks", "required", nil, "VERDICT: BLOCKED", []string{"phantom", "✗", "install:"}},
		{"missing recommended tool degrades", "recommended", nil, "VERDICT: DEGRADED", []string{"phantom"}},
		{"missing optional tool stays ready", "optional", nil, "VERDICT: READY", []string{"phantom"}},
		{"an unbuilt binary degrades", "optional", []string{"sc-doctor"}, "VERDICT: DEGRADED", []string{"sc-doctor", "not built"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			out, code := doctor(t, doctorRoot(t, manifestJSON(c.tier), c.cmds...))
			if code != 0 {
				t.Fatalf("exit %d; the preflight must always exit 0", code)
			}
			if !strings.Contains(out, c.want) {
				t.Fatalf("want %q in:\n%s", c.want, out)
			}
			for _, w := range c.wantLine {
				if !strings.Contains(out, w) {
					t.Errorf("output missing %q:\n%s", w, out)
				}
			}
			if n := strings.Count(out, "VERDICT:"); n != 1 {
				t.Errorf("want exactly one verdict line, got %d", n)
			}
		})
	}
}

// A built binary flips the verdict — the fact --fix exists to establish.
func TestRunReadyOnceTheBinaryIsBuilt(t *testing.T) {
	root := doctorRoot(t, manifestJSON("optional"), "sc-doctor")
	if err := os.WriteFile(filepath.Join(root, "bin", "sc-doctor"+exeSuffix()), []byte("x"), 0o755); err != nil {
		t.Fatal(err)
	}
	out, _ := doctor(t, root)
	if !strings.Contains(out, "VERDICT: READY") {
		t.Fatalf("a provisioned tree must read READY:\n%s", out)
	}
	if !strings.Contains(out, "✓ built") {
		t.Errorf("the built binary is not reported as built:\n%s", out)
	}
}

// An unreadable or malformed manifest is reported, not crashed on, and still exits 0.
func TestRunManifestFailures(t *testing.T) {
	cases := []struct {
		name, manifest, want string
	}{
		{"missing manifest", "", "cannot read requirements.json"},
		{"malformed manifest", `{"tools":`, "malformed requirements.json"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			out, code := doctor(t, doctorRoot(t, c.manifest))
			if code != 0 {
				t.Fatalf("exit %d", code)
			}
			if !strings.Contains(out, c.want) {
				t.Fatalf("want %q in %q", c.want, out)
			}
			if strings.Contains(out, "VERDICT:") {
				t.Error("a doctor that could not read its manifest must not publish a verdict")
			}
		})
	}
}

// The dance warnings are cache facts, and a stale cache must pull READY down.
func TestRunDanceWarningDowngradesReady(t *testing.T) {
	root := doctorRoot(t, manifestJSON("optional"))
	// A newer version dir than the one running, with an empty bin/.
	newer := filepath.Join(filepath.Dir(root), "0.9.2", "bin")
	if err := os.MkdirAll(newer, 0o755); err != nil {
		t.Fatal(err)
	}
	out, _ := doctor(t, root)
	if !strings.Contains(out, "DANCE INCOMPLETE") || !strings.Contains(out, "EMPTY-BIN WINDOW") {
		t.Fatalf("both cache warnings expected:\n%s", out)
	}
	if !strings.Contains(out, "VERDICT: DEGRADED") {
		t.Fatalf("warnings must downgrade READY:\n%s", out)
	}
}

// The suite is one preflight: a sibling plugin's manifest is probed and reported here.
func TestRunReportsSiblingPlugins(t *testing.T) {
	root := doctorRoot(t, manifestJSON("optional"))
	sib := filepath.Join(filepath.Dir(filepath.Dir(root)), "frank-exchange-of-views", "0.6.0")
	if err := os.MkdirAll(sib, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sib, "requirements.json"), []byte(manifestJSON("recommended")), 0o644); err != nil {
		t.Fatal(err)
	}
	out, _ := doctor(t, root)
	if !strings.Contains(out, "-- frank-exchange-of-views --") {
		t.Fatalf("sibling section missing:\n%s", out)
	}
}

// -fix must reach the provisioner with the binaries it is meant to provision, and
// must re-probe afterwards rather than reporting the pre-fix state.
func TestFixIsWiredAndReProbes(t *testing.T) {
	root := doctorRoot(t, manifestJSON("optional"), "sc-doctor")
	var fixed []binStatus
	var out bytes.Buffer
	// The stand-in provisioner actually creates the binary, so the re-probe has
	// something new to see.
	fixFn := func(bins []binStatus) []string {
		for _, b := range bins {
			if b.Built {
				continue
			}
			_ = os.MkdirAll(filepath.Join(b.Root, "bin"), 0o755)
			_ = os.WriteFile(filepath.Join(b.Root, "bin", b.Name+exeSuffix()), []byte("x"), 0o755)
		}
		fixed = append(fixed, bins...)
		return []string{"provisioned"}
	}
	if code := run([]string{"-root", root, "-fix"}, &out, "", noExecutable, fixFn); code != 0 {
		t.Fatalf("exit %d", code)
	}
	if len(fixed) != 1 || fixed[0].Name != "sc-doctor" || fixed[0].Built {
		t.Fatalf("fix received the wrong work list: %+v", fixed)
	}
	if !strings.Contains(out.String(), "provisioned") {
		t.Errorf("the fix report is not surfaced:\n%s", out.String())
	}
	if !strings.Contains(out.String(), "✓ built") || !strings.Contains(out.String(), "VERDICT: READY") {
		t.Fatalf("--fix must re-probe; the table still shows the pre-fix state:\n%s", out.String())
	}
}

// Without -fix the doctor is read-only. That is the difference between the two modes,
// so it is asserted rather than assumed.
func TestPlainRunNeverProvisions(t *testing.T) {
	root := doctorRoot(t, manifestJSON("optional"), "sc-doctor")
	var fixed []binStatus
	var out bytes.Buffer
	run([]string{"-root", root}, &out, "", noExecutable, noFix(&fixed))
	if len(fixed) != 0 {
		t.Fatal("a plain run called the provisioner")
	}
	if _, err := os.Stat(filepath.Join(root, "bin", "sc-doctor"+exeSuffix())); !os.IsNotExist(err) {
		t.Fatal("a plain run installed a binary")
	}
}

func TestResolveRoot(t *testing.T) {
	exe := func() (string, error) { return filepath.Join("plugin", "bin", "sc-doctor"), nil }
	cases := []struct {
		name, flag, env string
		exe             func() (string, error)
		want            string
	}{
		{"flag wins", "from-flag", "from-env", exe, "from-flag"},
		{"env is next", "", "from-env", exe, "from-env"},
		{"executable grandparent is the fallback", "", "", exe, "plugin"},
		{"unknown executable yields no root", "", "", noExecutable, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := resolveRoot(c.flag, c.env, c.exe); got != c.want {
				t.Fatalf("resolveRoot = %q; want %q", got, c.want)
			}
		})
	}
}

func TestVersionFlagSkipsTheProbe(t *testing.T) {
	var out bytes.Buffer
	var fixed []binStatus
	if code := run([]string{"-version"}, &out, "", noExecutable, noFix(&fixed)); code != 0 {
		t.Fatalf("exit %d", code)
	}
	if !strings.Contains(out.String(), "sc-doctor") || !strings.Contains(out.String(), buildid.Revision()) {
		t.Fatalf("version output = %q", out.String())
	}
	if strings.Contains(out.String(), "VERDICT:") {
		t.Error("-version must not run the preflight")
	}
}

func TestUnknownFlagExitsCleanly(t *testing.T) {
	var out bytes.Buffer
	var fixed []binStatus
	if code := run([]string{"-nope"}, &out, "", noExecutable, noFix(&fixed)); code != 0 {
		t.Fatalf("exit %d on an unknown flag", code)
	}
}

// hookBinaries derives the OWNING plugin and version from the root path, and --fix
// stamps a release tag out of them. Getting this wrong sends the fetch at another
// plugin's tag, so the derivation is pinned here.
func TestHookBinariesTagsItsOwnPluginAndVersion(t *testing.T) {
	root := doctorRoot(t, manifestJSON("optional"), "sc-doctor", "sc-quality-gate")
	got := hookBinaries(root)
	if len(got) != 2 {
		t.Fatalf("want both commands, got %+v", got)
	}
	for _, b := range got {
		if b.Plugin != "prosthetic-conscience" || b.Version != "0.9.1" {
			t.Errorf("binary mis-tagged: %+v", b)
		}
		if b.Root != root {
			t.Errorf("root = %q, want %q", b.Root, root)
		}
		if b.Built {
			t.Errorf("%s reported built with no bin/", b.Name)
		}
	}
}

// Non-directory entries under tools/cmd are not commands.
func TestBinariesOfIgnoresFiles(t *testing.T) {
	root := doctorRoot(t, manifestJSON("optional"), "sc-doctor")
	if err := os.WriteFile(filepath.Join(root, "tools", "cmd", "README.md"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := binariesOf(root, "p", "1.0.0"); len(got) != 1 || got[0].Name != "sc-doctor" {
		t.Fatalf("a stray file was read as a command: %+v", got)
	}
	if got := binariesOf(filepath.Join(root, "nope"), "p", "1.0.0"); got != nil {
		t.Fatalf("a missing tools/cmd must yield nothing, got %+v", got)
	}
}

func TestSiblingHelpersOnAMissingCache(t *testing.T) {
	orphan := filepath.Join(t.TempDir(), "a", "b", "c")
	if got := siblingBinaries(orphan); got != nil {
		t.Errorf("no cache means no siblings, got %+v", got)
	}
	if got := siblingRequirements(orphan); got != nil {
		t.Errorf("no cache means no sibling requirements, got %+v", got)
	}
	if got := newestVersionDir(orphan); got != "" {
		t.Errorf("newestVersionDir on a missing dir = %q", got)
	}
	if got := danceWarnings(orphan); len(got) != 0 {
		t.Errorf("no cache means no dance warnings, got %v", got)
	}
}

// fix must skip binaries that already exist — otherwise every doctor run would
// re-download the whole suite. This is the one fix() path safe to exercise: it
// reaches neither the network nor the Go toolchain.
func TestFixSkipsBuiltBinaries(t *testing.T) {
	if got := fix([]binStatus{{Name: "sc-doctor", Built: true}, {Name: "feov-record", Built: true}}); len(got) != 0 {
		t.Fatalf("fix reported work for already-built binaries: %v", got)
	}
	if got := fix(nil); len(got) != 0 {
		t.Fatalf("fix on an empty list reported %v", got)
	}
}

// ---- versionOf ----

// TestHelperVersionOutput is not a real test; it is the child process versionOf runs.
func TestHelperVersionOutput(t *testing.T) {
	switch os.Getenv("SC_DOCTOR_HELPER") {
	case "multiline":
		fmt.Println("  phantom 1.2.3  ")
		fmt.Println("second line must be dropped")
		os.Exit(0)
	case "oneline":
		fmt.Print("  phantom 1.2.3  ") // no trailing newline at all
		os.Exit(0)
	case "leadingblank":
		// A banner that opens with a blank line. The newline is then at index 0,
		// which is the one input distinguishing `i >= 0` from `i > 0` — under the
		// latter the whole multi-line banner lands in a single-column table.
		fmt.Print("\nphantom 1.2.3\nsecond line must be dropped\n")
		os.Exit(0)
	default:
		t.Skip("helper process, invoked by TestVersionOf")
	}
}

// versionOf reports the FIRST line of a check command, trimmed — the table has one
// column for it, and a multi-line banner would wreck the report. A command that emits
// no newline at all must be trimmed the same way.
func TestVersionOf(t *testing.T) {
	for _, mode := range []string{"multiline", "oneline", "leadingblank"} {
		t.Run(mode, func(t *testing.T) {
			t.Setenv("SC_DOCTOR_HELPER", mode)
			got := versionOf(os.Args[0] + " -test.run=TestHelperVersionOutput")
			if strings.Contains(got, "\n") {
				t.Fatalf("versionOf = %q; the table has ONE column for this", got)
			}
			if mode == "leadingblank" {
				// An empty first line carries no version, and reporting "" is
				// honest — the table then shows a bare ✓ rather than a banner.
				if got != "" {
					t.Fatalf("versionOf = %q; want the (empty) first line", got)
				}
				return
			}
			if got != "phantom 1.2.3" {
				t.Fatalf("versionOf = %q; want the trimmed first line", got)
			}
		})
	}
}

func TestVersionOfFailures(t *testing.T) {
	cases := []struct{ name, cmd string }{
		{"empty command", ""},
		{"whitespace only", "   "},
		{"binary that does not resolve", "sc-definitely-not-real-xyz --version"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := versionOf(c.cmd); got != "" {
				t.Fatalf("versionOf(%q) = %q; want empty", c.cmd, got)
			}
		})
	}
}
