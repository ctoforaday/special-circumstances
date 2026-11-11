package doctor

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/toolchain"
	"github.com/ctoforaday/special-circumstances/shared/buildid"
)

// These tests drive the doctor over a plugin that ships native helpers and an ASSETS export, through
// the shipped fetchRelease and, where the plan says so, through run as Main calls it.

// embedExports is the fixture ASSETS line's per-arch release names. Every expectation reads the
// host's entry from it at runtime.GOARCH, so the tests hold on arm64, where the export is not named
// after the arch, as they do on amd64.
var embedExports = map[string]string{"amd64": "gte-small.int8.amd64.ort", "arm64": "gte-small.int8.arm.ort"}

func embedAssetsLine() string {
	return "gte-small.int8.ort amd64=" + embedExports["amd64"] + " arm64=" + embedExports["arm64"]
}

// hostExport is the release name the fixture line gives runtime.GOARCH.
func hostExport(t *testing.T) string {
	t.Helper()
	e, ok := embedExports[runtime.GOARCH]
	if !ok {
		t.Skipf("the fixture ASSETS line names no export for %s", runtime.GOARCH)
	}
	return e
}

// fakeGHSource is a gh that serves `gh release download <tag> --pattern … --dir …` from
// $FAKE_GH_RELEASE/<tag>/, logs each argv to $FAKE_GH_LOG, and exits 1 when $FAKE_GH_FAIL is set.
// It is Go, not a shell stub, so the shipped fetchRelease runs it on Windows as on Linux.
const fakeGHSource = `package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if p := os.Getenv("FAKE_GH_LOG"); p != "" {
		if f, err := os.OpenFile(p, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644); err == nil {
			fmt.Fprintln(f, strings.Join(os.Args[1:], " "))
			f.Close()
		}
	}
	if os.Getenv("FAKE_GH_FAIL") != "" {
		fmt.Fprintln(os.Stderr, "release not found")
		os.Exit(1)
	}
	args := os.Args[1:]
	if len(args) < 3 || args[0] != "release" || args[1] != "download" {
		fmt.Fprintln(os.Stderr, "fake gh serves only release download, got", args)
		os.Exit(2)
	}
	tag, dir := args[2], "."
	var patterns []string
	for i := 3; i+1 < len(args); i++ {
		switch args[i] {
		case "--pattern":
			i++
			patterns = append(patterns, args[i])
		case "--dir":
			i++
			dir = args[i]
		case "--repo":
			i++
		}
	}
	for _, p := range patterns {
		data, err := os.ReadFile(filepath.Join(os.Getenv("FAKE_GH_RELEASE"), tag, p))
		if err != nil {
			fmt.Fprintln(os.Stderr, "no assets match the file pattern", p)
			os.Exit(1)
		}
		if err := os.WriteFile(filepath.Join(dir, p), data, 0o644); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}
`

// fakeGH builds fakeGHSource and returns the directory holding it. The callers make it the only
// directory on PATH: no go, no real gh.
func fakeGH(t *testing.T) string {
	t.Helper()
	src := t.TempDir()
	mustWrite(t, filepath.Join(src, "main.go"), fakeGHSource)
	dir := t.TempDir()
	cmd := exec.Command("go", "build", "-o", filepath.Join(dir, buildid.ExeName("gh")), "main.go")
	cmd.Dir = src
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("building the fake gh: %v\n%s", err, out)
	}
	return dir
}

// writeRelease lays out one tag of a release, every asset listed in its SHA256SUMS.
func writeRelease(t *testing.T, dir string, assets map[string]string) {
	t.Helper()
	var sums strings.Builder
	for name, body := range assets {
		mustWrite(t, filepath.Join(dir, name), body)
		fmt.Fprintf(&sums, "%x  %s\n", sha256.Sum256([]byte(body)), name)
	}
	mustWrite(t, filepath.Join(dir, "SHA256SUMS"), sums.String())
}

// neverBuild fails the test if called: a helper or an export is fetched, never built.
func neverBuild(t *testing.T) func(binStatus) (string, error) {
	return func(b binStatus) (string, error) {
		t.Errorf("build called for %s (kind %q): only a Go command is built from source", b.Name, b.Kind)
		return "", errors.New("never built")
	}
}

// reportLine is the report's line about name, or "".
func reportLine(report []string, name string) string {
	for _, l := range report {
		if strings.HasPrefix(l, name+":") || strings.HasPrefix(l, name+" (") {
			return l
		}
	}
	return ""
}

// stubBoxWarnings removes the box's own settings from a run: what this box's ~/.claude says about
// thinking capture is not what these tests judge.
func stubBoxWarnings(t *testing.T) {
	orig := boxWarnings
	t.Cleanup(func() { boxWarnings = orig })
	boxWarnings = func() []string { return nil }
}

// embedPlugin is a frank-exchange-of-views install with an feov-embed helper and its export, and,
// with cmds, Go commands. It sits at <cache>/<marketplace>/frank-exchange-of-views/1.0.0/.
func embedPlugin(t *testing.T, cache string, ocr bool, cmds ...string) string {
	t.Helper()
	root := filepath.Join(cache, "special-circumstances", "frank-exchange-of-views", "1.0.0")
	mustWrite(t, filepath.Join(root, ".claude-plugin", "plugin.json"), `{"name":"frank-exchange-of-views","version":"1.0.0"}`)
	mustWrite(t, filepath.Join(root, "requirements.json"), `{"tools":[]}`)
	mustWrite(t, filepath.Join(root, "native", "feov-embed", "ASSETS"), embedAssetsLine()+"\n")
	if ocr {
		mustWrite(t, filepath.Join(root, "native", "feov-ocr", "ASSETS"), "")
	}
	for _, c := range cmds {
		mustWrite(t, filepath.Join(root, "tools", "cmd", c, "main.go"), "package main\n\nfunc main() {}\n")
	}
	if len(cmds) > 0 {
		mustWrite(t, filepath.Join(root, "tools", "go.mod"), "module fixture\n\ngo 1.25\n")
	}
	return root
}

func TestEveryEntryHasAKind(t *testing.T) {
	root := embedPlugin(t, t.TempDir(), false, "alpha")
	want := map[string]string{"alpha": "go", "feov-embed": "helper", "gte-small.int8.ort": "asset"}
	bins := binariesOf(root, "frank-exchange-of-views", "1.0.0")
	if len(bins) != len(want) {
		t.Fatalf("binariesOf = %+v, want one entry each for %v", bins, want)
	}
	for _, b := range bins {
		if b.Kind != want[b.Name] {
			t.Errorf("%s: Kind %q, want %q", b.Name, b.Kind, want[b.Name])
		}
	}
}

func TestStampWaitsForEveryHelper(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "plugin.json"), `{"name":"fixture","version":"1.0.0"}`)
	mustWrite(t, filepath.Join(root, "tools", "cmd", "alpha", "main.go"), "package main\nfunc main(){}")
	mustWrite(t, filepath.Join(root, "native", "fake-helper", "ASSETS"), "model.bin amd64=model.amd64.bin arm64=model.arm64.bin\n")
	stamp := filepath.Join(root, ".fetch", "installed")
	if bins := binariesOf(root, "fixture", "1.0.0"); len(bins) != 3 {
		t.Fatalf("want alpha, fake-helper and model.bin, got %+v", bins)
	}
	for _, step := range []struct{ place, missing string }{
		{"", "every entry"},
		{buildid.ExeName("alpha"), "the helper and the ASSETS file"},
		{buildid.ExeName("fake-helper"), "the ASSETS file model.bin"},
	} {
		if step.place != "" {
			mustWrite(t, filepath.Join(root, "bin", step.place), step.place)
		}
		stampInstalled(binariesOf(root, "fixture", "1.0.0"))
		if _, err := os.Stat(stamp); err == nil {
			t.Fatalf(".fetch/installed was written while %s is missing", step.missing)
		}
	}
	mustWrite(t, filepath.Join(root, "bin", "model.bin"), "model")
	stampInstalled(binariesOf(root, "fixture", "1.0.0"))
	if _, err := os.Stat(stamp); err != nil {
		t.Fatalf(".fetch/installed was not written once every entry is present: %v", err)
	}
}

// committedAlpha builds cmd/alpha from a git repository the test creates and commits, so its build
// stamp names that commit and is clean. It returns the binary and the commit.
func committedAlpha(t *testing.T) (bin, commit string) {
	t.Helper()
	repo := t.TempDir()
	mustWrite(t, filepath.Join(repo, "tools", "go.mod"), "module fixture\n\ngo 1.25\n")
	mustWrite(t, filepath.Join(repo, "tools", "cmd", "alpha", "main.go"), "package main\n\nfunc main() {}\n")
	git := func(args ...string) string {
		t.Helper()
		out, err := exec.Command("git", append([]string{"-C", repo, "-c", "user.name=test", "-c", "user.email=test@example.com", "-c", "commit.gpgsign=false"}, args...)...).CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}
	git("init", "-q")
	git("add", ".")
	git("commit", "-q", "-m", "fixture")
	bin = filepath.Join(t.TempDir(), buildid.ExeName("alpha"))
	build := exec.Command("go", "build", "-o", bin, "./cmd/alpha")
	build.Dir = filepath.Join(repo, "tools")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build: %v\n%s", err, out)
	}
	return bin, git("rev-parse", "HEAD")
}

func TestStampRecordsTheCommitWithHelpersPresent(t *testing.T) {
	alpha, commit := committedAlpha(t)
	// An install outside any git tree, where the Go binary's stamp is the only source of the commit,
	// and whose last entries are a helper and an export.
	root := embedPlugin(t, t.TempDir(), true, "alpha")
	data, err := os.ReadFile(alpha)
	if err != nil {
		t.Fatal(err)
	}
	mustWrite(t, filepath.Join(root, "bin", buildid.ExeName("alpha")), string(data))
	mustWrite(t, filepath.Join(root, "bin", buildid.ExeName("feov-ocr")), "helper")
	mustWrite(t, filepath.Join(root, "bin", buildid.ExeName("feov-embed")), "helper")
	mustWrite(t, filepath.Join(root, "bin", "gte-small.int8.ort"), "export")
	if h := HeadCommit(root); h != "" {
		t.Fatalf("the fixture install sits in a git tree (HEAD %s); the case needs none", h)
	}
	stampInstalled(binariesOf(root, "frank-exchange-of-views", "1.0.0"))
	got, err := os.ReadFile(filepath.Join(root, ".fetch", "commit"))
	if err != nil || strings.TrimSpace(string(got)) != commit {
		t.Fatalf(".fetch/commit = %q (%v), want alpha's build commit %s", got, err, commit)
	}
}

func TestFixFetchesTheExportForItsArch(t *testing.T) {
	export := hostExport(t)
	gh := fakeGH(t)
	rel := t.TempDir()
	helperAsset := AssetName("feov-embed", runtime.GOOS, runtime.GOARCH)
	exportBytes := "export for " + runtime.GOARCH + "\n"
	writeRelease(t, filepath.Join(rel, "frank-exchange-of-views--v1.0.0"), map[string]string{
		helperAsset: "helper\n",
		export:      exportBytes,
	})
	ghLog := filepath.Join(t.TempDir(), "gh.log")
	t.Setenv("SC_RELEASE_BASE_URL", "")
	t.Setenv("PATH", gh)
	t.Setenv("FAKE_GH_RELEASE", rel)
	t.Setenv("FAKE_GH_LOG", ghLog)

	t.Run("installs the host arch's export", func(t *testing.T) {
		root := embedPlugin(t, t.TempDir(), false)
		report := fixWith(binariesOf(root, "frank-exchange-of-views", "1.0.0"), fetchRelease, neverBuild(t), true)
		for _, name := range []string{"feov-embed", "gte-small.int8.ort"} {
			if l := reportLine(report, name); !strings.Contains(l, "fetched CI-built release asset") {
				t.Errorf("%s was not fetched; report:\n%s", name, strings.Join(report, "\n"))
			}
		}
		got, err := os.ReadFile(filepath.Join(root, "bin", "gte-small.int8.ort"))
		if err != nil || string(got) != exportBytes {
			t.Fatalf("bin/gte-small.int8.ort = %q (%v), want the %s export's bytes %q", got, err, runtime.GOARCH, exportBytes)
		}
		logged, err := os.ReadFile(ghLog)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(logged), "--pattern "+export+" ") {
			t.Errorf("gh was not asked for %s:\n%s", export, logged)
		}
		if bad := AssetName("gte-small.int8.ort", runtime.GOOS, runtime.GOARCH); strings.Contains(string(logged), bad) {
			t.Errorf("gh was asked for %s, an asset no release carries:\n%s", bad, logged)
		}
	})

	t.Run("a failed fetch is reported, never built", func(t *testing.T) {
		t.Setenv("FAKE_GH_FAIL", "1")
		root := embedPlugin(t, t.TempDir(), false)
		report := fixWith(binariesOf(root, "frank-exchange-of-views", "1.0.0"), fetchRelease, neverBuild(t), true)
		for _, name := range []string{"feov-embed", "gte-small.int8.ort"} {
			if l := reportLine(report, name); !strings.Contains(l, "fetch failed") {
				t.Errorf("the report does not name %s as a fetch failure:\n%s", name, strings.Join(report, "\n"))
			}
		}
	})

	t.Run("a line without the host arch fails naming it", func(t *testing.T) {
		root := embedPlugin(t, t.TempDir(), false)
		mustWrite(t, filepath.Join(root, "native", "feov-embed", "ASSETS"), "gte-small.int8.ort elsewhere=gte-small.int8.elsewhere.ort\n")
		report := fixWith(binariesOf(root, "frank-exchange-of-views", "1.0.0"), fetchRelease, neverBuild(t), true)
		l := reportLine(report, "gte-small.int8.ort")
		if !strings.Contains(l, "fetch failed") || !strings.Contains(l, "has no asset for "+runtime.GOARCH) {
			t.Fatalf("the report does not name the file and %s:\n%s", runtime.GOARCH, strings.Join(report, "\n"))
		}
		if _, err := os.Stat(filepath.Join(root, "bin", "gte-small.int8.ort")); err == nil {
			t.Fatal("an export was installed for a line that names none for this arch")
		}
	})
}

func TestDoctorJudgesHelpersByPresenceOnly(t *testing.T) {
	stubBoxWarnings(t)
	cache := t.TempDir()
	root := embedPlugin(t, cache, true, "alpha")
	git := func(args ...string) string {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-C", cache, "-c", "user.name=test", "-c", "user.email=test@example.com", "-c", "commit.gpgsign=false"}, args...)...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}
	git("init", "-q")
	git("add", ".")
	git("commit", "-q", "-m", "fixture")
	commit := git("rev-parse", "HEAD")

	// Built from the committed tree, so the stamp is clean and names the commit
	// installed_plugins.json records, whatever state the real checkout is in.
	build := exec.Command("go", "build", "-o", filepath.Join(root, "bin", buildid.ExeName("alpha")), "./cmd/alpha")
	build.Dir = filepath.Join(root, "tools")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build: %v\n%s", err, out)
	}
	mustWrite(t, filepath.Join(cache, "installed_plugins.json"),
		fmt.Sprintf(`{"plugins":{"frank-exchange-of-views@special-circumstances":[{"installPath":%q,"version":"1.0.0","gitCommitSha":%q}]}}`, root, commit))
	mustWrite(t, filepath.Join(root, "bin", buildid.ExeName("feov-ocr")), "not a go binary")
	mustWrite(t, filepath.Join(root, "bin", buildid.ExeName("feov-embed")), "not a go binary")
	mustWrite(t, filepath.Join(root, "bin", "gte-small.int8.ort"), "ort bytes")

	var out bytes.Buffer
	if code := run([]string{"-root", root}, &out, "", noExecutable, fix); code != 0 {
		t.Fatalf("exit %d", code)
	}
	lines := strings.Split(out.String(), "\n")
	marks := map[string]string{}
	for _, l := range lines {
		name, mark, ok := strings.Cut(l, " ")
		if ok {
			marks[name] = strings.TrimSpace(mark)
		}
	}
	for name, want := range map[string]string{
		"alpha":              "✓ built",
		"feov-ocr":           "✓ installed",
		"feov-embed":         "✓ installed",
		"gte-small.int8.ort": "✓ installed",
	} {
		if marks[name] != want {
			t.Errorf("%s reads %q, want %q:\n%s", name, marks[name], want, out.String())
		}
	}
	if !strings.Contains(out.String(), "\nVERDICT: READY\n") {
		t.Fatalf("want a bare VERDICT: READY, no qualifier:\n%s", out.String())
	}
}

func TestDoctorNamesTheExportWhenNothingCanProvideIt(t *testing.T) {
	stubBoxWarnings(t)
	export := hostExport(t)
	gh := fakeGH(t)
	root := embedPlugin(t, t.TempDir(), true, "alpha")
	t.Setenv("SC_RELEASE_BASE_URL", "")
	t.Setenv("PATH", gh)
	t.Setenv("FAKE_GH_FAIL", "1")
	if toolchain.Present("go") {
		t.Fatal("go is on the test's PATH; the case is nothing that can provide a file")
	}

	var out bytes.Buffer
	if code := run([]string{"-root", root, "-fix"}, &out, "", noExecutable, fix); code != 0 {
		t.Fatalf("exit %d", code)
	}
	report := out.String()
	for _, want := range []string{AssetName("feov-ocr", runtime.GOOS, runtime.GOARCH), export} {
		if !strings.Contains(report, "place "+want+" into ") {
			t.Errorf("the report does not name %s as the file to place:\n%s", want, report)
		}
	}
	if bad := AssetName("gte-small.int8.ort", runtime.GOOS, runtime.GOARCH); strings.Contains(report, bad) {
		t.Errorf("the report names %s, an asset no release carries:\n%s", bad, report)
	}
}
