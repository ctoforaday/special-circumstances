package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

var linuxAMD64 = Target{Pair: "linux/amd64", Name: "linux-amd64", OS: "linux", Arch: "amd64"}
var windowsARM64 = Target{Pair: "windows/arm64", Name: "windows-arm64", OS: "windows", Arch: "arm64", Ext: ".exe"}

// fixturePlugin lays out a plugin with the named commands and helpers, and returns its directory.
func fixturePlugin(t *testing.T, cmds []string, helpers map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for _, c := range cmds {
		if err := os.MkdirAll(filepath.Join(dir, "tools", "cmd", c), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	for h, assets := range helpers {
		if err := os.MkdirAll(filepath.Join(dir, "native", h), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "native", h, "ASSETS"), []byte(assets), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func staticDist(t *testing.T, names ...string) string {
	t.Helper()
	dir := t.TempDir()
	for _, n := range names {
		makeStaticELF(t, filepath.Join(dir, n))
	}
	return dir
}

// TestADynamicELFIsRefused builds what a plain host build of feov-record is (audit round 1): a Go
// program importing net with cgo on, linked against the host's libc, and puts it in a dist.
func TestADynamicELFIsRefused(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skipf("SKIPPED, NOT PASSED: the dynamic program is a host ELF, and this host is %s", runtime.GOOS)
	}
	cc := os.Getenv("CC")
	if cc == "" {
		cc = "cc"
	}
	if _, err := exec.LookPath(cc); err != nil {
		if os.Getenv("GITHUB_ACTIONS") != "" {
			t.Fatalf("no C compiler (%s) on this runner, so a dynamic ELF cannot be built: %v", cc, err)
		}
		t.Skipf("SKIPPED, NOT PASSED: no C compiler (%s): %v", cc, err)
	}
	src := t.TempDir()
	prog := "package main\n\nimport (\n\t\"fmt\"\n\t\"net\"\n)\n\nfunc main() { fmt.Println(net.LookupHost(\"localhost\")) }\n"
	if err := os.WriteFile(filepath.Join(src, "main.go"), []byte(prog), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "go.mod"), []byte("module dyn\n\ngo 1.27\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	dist := t.TempDir()
	out := filepath.Join(dist, "feov-record_linux_amd64")
	build := exec.Command("go", "build", "-o", out, ".")
	build.Dir = src
	build.Env = append(os.Environ(), "CGO_ENABLED=1", "GOOS=linux", "GOARCH="+runtime.GOARCH, "GOFLAGS=")
	if b, err := build.CombinedOutput(); err != nil {
		t.Fatalf("building the cgo program: %v\n%s", err, b)
	}
	plugin := fixturePlugin(t, []string{"feov-record"}, nil)
	target := Target{OS: "linux", Arch: runtime.GOARCH}
	if runtime.GOARCH != "amd64" {
		renamed := filepath.Join(dist, "feov-record_linux_"+runtime.GOARCH)
		if err := os.Rename(out, renamed); err != nil {
			t.Fatal(err)
		}
		out = renamed
	}
	err := CheckDist(dist, plugin, []Target{target}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("CheckDist passed a cgo build importing net, a dynamically linked ELF")
	}
	if !strings.Contains(err.Error(), "not statically linked") || !strings.Contains(err.Error(), out) {
		t.Fatalf("the refusal does not name the dynamic asset as not statically linked: %v", err)
	}
}

// TestEveryExecutableIsProven: three Go assets and one helper. Each kind's dynamic member is
// refused on its own, so a check that skips staticproof for either kind fails here.
func TestEveryExecutableIsProven(t *testing.T) {
	plugin := fixturePlugin(t, []string{"feov-record", "feov-hook", "feov-migrate"}, map[string]string{"feov-ocr": ""})
	all := []string{"feov-record_linux_amd64", "feov-hook_linux_amd64", "feov-migrate_linux_amd64", "feov-ocr_linux_amd64"}

	var out bytes.Buffer
	if err := CheckDist(staticDist(t, all...), plugin, []Target{linuxAMD64}, &out); err != nil {
		t.Fatalf("CheckDist refused an all-static dist: %v", err)
	}
	for _, n := range all {
		if !strings.Contains(out.String(), "proven static: "+n+" ") {
			t.Errorf("%s was not proven static:\n%s", n, out.String())
		}
	}
	for _, dynamic := range []string{"feov-migrate_linux_amd64", "feov-ocr_linux_amd64"} {
		t.Run(dynamic, func(t *testing.T) {
			dist := staticDist(t, all...)
			makeDynamicELF(t, filepath.Join(dist, dynamic))
			err := CheckDist(dist, plugin, []Target{linuxAMD64}, &bytes.Buffer{})
			if err == nil || !strings.Contains(err.Error(), dynamic) {
				t.Fatalf("a dynamic %s passed the check: %v", dynamic, err)
			}
		})
	}
}

func TestAFEOVDistMissingOneHelperIsRefused(t *testing.T) {
	plugin := fixturePlugin(t, []string{"feov-record"}, map[string]string{"feov-ocr": ""})
	dist := staticDist(t, "feov-record_linux_amd64", "feov-record_windows_arm64.exe", "feov-ocr_linux_amd64")
	err := CheckDist(dist, plugin, []Target{linuxAMD64, windowsARM64}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "feov-ocr_windows_arm64.exe") {
		t.Fatalf("a dist missing feov-ocr_windows_arm64.exe was not refused naming it: %v", err)
	}
}

// TestAnUnsuffixedHelperIsRefused: native-build's artifacts merge into the release dist, so a
// stray file in one would be published. The dist holds exactly the release's names.
func TestAnUnsuffixedHelperIsRefused(t *testing.T) {
	plugin := fixturePlugin(t, []string{"feov-record"}, map[string]string{"feov-ocr": ""})
	dist := staticDist(t, "feov-record_linux_amd64", "feov-ocr_linux_amd64", "feov-ocr")
	err := CheckDist(dist, plugin, []Target{linuxAMD64}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "unexpected [feov-ocr]") {
		t.Fatalf("a dist holding an unsuffixed feov-ocr was not refused naming it: %v", err)
	}
}

func TestADistWithoutHelpersNeedsNone(t *testing.T) {
	plugin := fixturePlugin(t, []string{"gray-area"}, nil)
	dist := staticDist(t, "gray-area_linux_amd64", "gray-area_windows_arm64.exe", "SHA256SUMS")
	if err := CheckDist(dist, plugin, []Target{linuxAMD64, windowsARM64}, &bytes.Buffer{}); err != nil {
		t.Fatalf("CheckDist demanded something of a plugin with no helpers: %v", err)
	}
}

// TestEveryAssetsExportIsExpected: an ASSETS line's release asset for each published arch is in
// the dist, and a missing one is refused.
func TestEveryAssetsExportIsExpected(t *testing.T) {
	plugin := fixturePlugin(t, []string{"feov-record"}, map[string]string{
		"feov-embed": "gte-small.int8.ort amd64=gte-small.int8.amd64.ort arm64=gte-small.int8.arm.ort\n",
	})
	want, err := ExpectedDist(plugin, []Target{linuxAMD64, windowsARM64})
	if err != nil {
		t.Fatal(err)
	}
	for _, n := range []string{"gte-small.int8.amd64.ort", "gte-small.int8.arm.ort"} {
		if a, ok := want[n]; !ok || a != "" {
			t.Errorf("ExpectedDist holds %s as %q, %v; want it as a non-executable export", n, a, ok)
		}
	}
	dist := staticDist(t, "feov-record_linux_amd64", "feov-record_windows_arm64.exe", "feov-embed_linux_amd64", "feov-embed_windows_arm64.exe", "gte-small.int8.amd64.ort")
	err = CheckDist(dist, plugin, []Target{linuxAMD64, windowsARM64}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "gte-small.int8.arm.ort") {
		t.Fatalf("a dist missing an ASSETS export was not refused naming it: %v", err)
	}
}

func TestPluginFromRef(t *testing.T) {
	if p, err := PluginFromRef("refs/tags/frank-exchange-of-views--v1.2.3"); err != nil || p != "frank-exchange-of-views" {
		t.Errorf("got %q, %v", p, err)
	}
	for _, bad := range []string{"refs/heads/main", "refs/tags/v1.2.3", "refs/tags/--v1", "refs/tags/gray-area--v"} {
		if _, err := PluginFromRef(bad); err == nil {
			t.Errorf("%s names a plugin release", bad)
		}
	}
}

// TestEveryExportMatchesItsPin: an ASSETS export in the dist is held to its sha256 in the helper's
// PINS.txt, never staticproofed; bytes no pin names are refused naming the export, and an export
// with no pin is refused.
func TestEveryExportMatchesItsPin(t *testing.T) {
	plugin := fixturePlugin(t, []string{"feov-record"}, map[string]string{
		"feov-embed": "gte-small.int8.ort amd64=gte-small.int8.amd64.ort arm64=gte-small.int8.arm.ort\n",
	})
	exports := map[string]string{"gte-small.int8.amd64.ort": "amd64 model", "gte-small.int8.arm.ort": "arm model"}
	pins := "# <sha256> <file> <url>\r\n"
	for name, body := range exports {
		pins += fmt.Sprintf("%x %s https://example.invalid/%s\r\n", sha256.Sum256([]byte(body)), name, name)
	}
	pinsPath := filepath.Join(plugin, "native", "feov-embed", "PINS.txt")
	if err := os.WriteFile(pinsPath, []byte(pins), 0o644); err != nil {
		t.Fatal(err)
	}
	dist := staticDist(t, "feov-record_linux_amd64", "feov-record_windows_arm64.exe", "feov-embed_linux_amd64", "feov-embed_windows_arm64.exe")
	for name, body := range exports {
		if err := os.WriteFile(filepath.Join(dist, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	targets := []Target{linuxAMD64, windowsARM64}
	if err := CheckDist(dist, plugin, targets, &bytes.Buffer{}); err != nil {
		t.Fatalf("CheckDist refused exports matching their pins: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dist, "gte-small.int8.arm.ort"), []byte("other bytes"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := CheckDist(dist, plugin, targets, &bytes.Buffer{}); err == nil || !strings.Contains(err.Error(), "gte-small.int8.arm.ort") {
		t.Fatalf("an export whose bytes no pin names was not refused naming it: %v", err)
	}
	if err := os.WriteFile(pinsPath, []byte(strings.Split(pins, "\r\n")[1]+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := CheckDist(dist, plugin, targets, &bytes.Buffer{}); err == nil || !strings.Contains(err.Error(), "want exactly one") {
		t.Fatalf("an export with no pin was not refused: %v", err)
	}
}
