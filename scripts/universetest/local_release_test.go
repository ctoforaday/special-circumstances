//go:build linux

package universetest

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"
)

// fixtureRepo creates an isolated Git repository mimicking the Special Circumstances repo
// with scripts/universe.sh, scripts/universe-manifest.py, .claude-plugin/marketplace.json,
// and plugins/prosthetic-conscience with tools/cmd/alpha and tools/cmd/beta.
type testFixture struct {
	repoDir string
	binDir  string
	stubDir string
	scratch string
	gocache string
	gomod   string
	baseEnv []string
}

const stubClaudeScript = `#!/usr/bin/env bash
set -e

mkdir -p "$STUB_DIR"
n=$(ls "$STUB_DIR" 2>/dev/null | grep -c '^argv\.' || true)
n=$((n + 1))
printf '%s\n' "$@" > "$STUB_DIR/argv.$n"
printf '%s\n' "$$" > "$STUB_DIR/pid.$n"

case "${1:-}" in
  plugin)
    case "${2:-}" in
      marketplace)
        case "${3:-}" in
          list)
            if [ -f "$STUB_DIR/marketplace_added" ]; then
              mkt_path="$(cat "$STUB_DIR/marketplace_added")"
              printf '[{"name":"special-circumstances","source":{"type":"directory","path":"%s"}}]\n' "$mkt_path"
            else
              printf '[]\n'
            fi
            exit 0
            ;;
          remove)
            rm -f "$STUB_DIR/marketplace_added"
            if [ -n "${CLAUDE_CONFIG_DIR:-}" ] && [ -f "$CLAUDE_CONFIG_DIR/settings.json" ]; then
              python3 -c "import json,sys; p=sys.argv[1]; d=json.load(open(p)); d.pop('enabledPlugins', None); json.dump(d, open(p, 'w'), indent=2)" "$CLAUDE_CONFIG_DIR/settings.json" 2>/dev/null || true
            fi
            exit 0
            ;;
          add)
            if [ "${STUB_FAIL_MARKETPLACE_ADD:-}" = "1" ]; then
              echo "marketplace add failed" >&2
              exit 1
            fi
            printf '%s\n' "$4" > "$STUB_DIR/marketplace_added"
            if [ -n "${CLAUDE_CONFIG_DIR:-}" ]; then
              python3 -c "import json,os,sys; p,m=sys.argv[1],sys.argv[2]; d=json.load(open(p)) if os.path.exists(p) else {}; d.setdefault('extraKnownMarketplaces', {})['special-circumstances']={'source':{'path':m}}; json.dump(d, open(p, 'w'), indent=2)" "$CLAUDE_CONFIG_DIR/settings.json" "$4" 2>/dev/null || true
            fi
            exit 0
            ;;
        esac
        ;;
      install)
        pname="${3%@*}"
        echo "$pname" >> "$STUB_DIR/installed.log"
        if [ "${STUB_INSTALL_WRITE_STALE:-}" = "1" ]; then
          mkt_root="${PWD}.marketplace/plugins/$pname"
          if [ -d "$mkt_root" ]; then
            mkdir -p "$mkt_root/bin" "$mkt_root/.fetch"
            printf '#!/bin/sh\necho stale\n' > "$mkt_root/bin/alpha"
            chmod +x "$mkt_root/bin/alpha"
            echo "prosthetic-conscience--v1.0.0" > "$mkt_root/.fetch/installed"
            echo "$mkt_root/bin/alpha" >> "$STUB_DIR/stale_written"
          fi
        fi
        exit 0
        ;;
    esac
    ;;
esac

if [ "$#" -eq 0 ]; then
  printenv > "$STUB_DIR/shell_env"
  exit 0
fi

exit 0
`

const stubCurlScript = `#!/usr/bin/env bash
set -e

mkdir -p "$STUB_DIR"
out=""
url=""
while [ $# -gt 0 ]; do
  case "$1" in
    -o) out="$2"; shift 2 ;;
    http://*|https://*|file://*) url="$1"; shift ;;
    *) shift ;;
  esac
done

if [ -n "$url" ]; then
  echo "$url" >> "$STUB_DIR/curl.log"
  case "$url" in
    file://*)
      path="${url#file://}"
      if [ -n "$out" ]; then
        mkdir -p "$(dirname "$out")"
        cp "$path" "$out"
      else
        cat "$path"
      fi
      ;;
    *)
      echo "curl: unsupported URL scheme: $url" >&2
      exit 1
      ;;
  esac
fi
`

func newTestFixture(t *testing.T) *testFixture {
	t.Helper()
	for _, tool := range []string{"git", "go", "python3", "bash"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skipf("%s not on PATH", tool)
		}
	}

	scratch := t.TempDir()
	repoDir := filepath.Join(scratch, "repo")
	binDir := filepath.Join(scratch, "bin")
	stubDir := filepath.Join(scratch, "stub")

	for _, d := range []string{
		filepath.Join(repoDir, "scripts"),
		filepath.Join(repoDir, ".claude-plugin"),
		filepath.Join(repoDir, "plugins", "prosthetic-conscience", ".claude-plugin"),
		filepath.Join(repoDir, "plugins", "prosthetic-conscience", "hooks"),
		filepath.Join(repoDir, "plugins", "prosthetic-conscience", "tools", "cmd", "alpha"),
		filepath.Join(repoDir, "plugins", "prosthetic-conscience", "tools", "cmd", "beta"),
		binDir,
		stubDir,
	} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}

	// Read host GOCACHE and GOMODCACHE before HOME moves
	gcOut, err := exec.Command("go", "env", "GOCACHE").Output()
	if err != nil {
		t.Fatal(err)
	}
	gocache := strings.TrimSpace(string(gcOut))

	gmcOut, err := exec.Command("go", "env", "GOMODCACHE").Output()
	if err != nil {
		t.Fatal(err)
	}
	gomod := strings.TrimSpace(string(gmcOut))

	// Copy scripts/universe.sh and scripts/universe-manifest.py from real repo
	realRepo, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	uScript, err := os.ReadFile(filepath.Join(realRepo, "scripts", "universe.sh"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repoDir, "scripts", "universe.sh"), uScript, 0o755); err != nil {
		t.Fatal(err)
	}

	uManifest, err := os.ReadFile(filepath.Join(realRepo, "scripts", "universe-manifest.py"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repoDir, "scripts", "universe-manifest.py"), uManifest, 0o755); err != nil {
		t.Fatal(err)
	}

	// Copy fetch-bin.sh
	fetchBin, err := os.ReadFile(filepath.Join(realRepo, "plugins", "prosthetic-conscience", "hooks", "fetch-bin.sh"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repoDir, "plugins", "prosthetic-conscience", "hooks", "fetch-bin.sh"), fetchBin, 0o755); err != nil {
		t.Fatal(err)
	}

	// Plugin metadata
	pluginJSON := "{\n  \"name\": \"prosthetic-conscience\",\n  \"version\": \"1.0.0\"\n}\n"
	if err := os.WriteFile(filepath.Join(repoDir, "plugins", "prosthetic-conscience", ".claude-plugin", "plugin.json"), []byte(pluginJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repoDir, "plugins", "prosthetic-conscience", "plugin.json"), []byte(pluginJSON), 0o644); err != nil {
		t.Fatal(err)
	}

	// The plugin's .claude-plugin/hooks.json, in the real one's shape: its PreToolUse guard is the
	// generated command of prosthetic-conscience's own, guarding the fixture's alpha instead of
	// sc-pretooluse.
	realHooks, err := os.ReadFile(filepath.Join(realRepo, "plugins", "prosthetic-conscience", ".claude-plugin", "hooks.json"))
	if err != nil {
		t.Fatal(err)
	}
	var rh hooksFile
	if err := json.Unmarshal(realHooks, &rh); err != nil || len(rh.Hooks["PreToolUse"]) == 0 || len(rh.Hooks["PreToolUse"][0].Hooks) == 0 {
		t.Fatalf("prosthetic-conscience's hooks.json has no PreToolUse guard to copy (%v)", err)
	}
	guardCmd := rh.Hooks["PreToolUse"][0].Hooks[0].Command
	if !strings.Contains(guardCmd, "/bin/sc-pretooluse") || !strings.Contains(guardCmd, "hook PreToolUse") {
		t.Fatalf("the real PreToolUse guard changed shape: %s", guardCmd)
	}
	var fh hooksFile
	fh.Hooks = map[string][]hookMatcher{"PreToolUse": {{Matcher: "Bash", Hooks: []hookCommand{{Type: "command",
		Command: strings.Replace(guardCmd, "/bin/sc-pretooluse", "/bin/alpha", 1)}}}}}
	fixtureHooks, err := json.MarshalIndent(fh, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repoDir, "plugins", "prosthetic-conscience", ".claude-plugin", "hooks.json"), fixtureHooks, 0o644); err != nil {
		t.Fatal(err)
	}

	// Go modules and commands
	goMod := "module pc\n\ngo 1.27\n"
	if err := os.WriteFile(filepath.Join(repoDir, "plugins", "prosthetic-conscience", "tools", "go.mod"), []byte(goMod), 0o644); err != nil {
		t.Fatal(err)
	}

	mainAlpha := "package main\n\nimport \"fmt\"\n\nfunc main() { fmt.Println(\"alpha\") }\n"
	if err := os.WriteFile(filepath.Join(repoDir, "plugins", "prosthetic-conscience", "tools", "cmd", "alpha", "main.go"), []byte(mainAlpha), 0o644); err != nil {
		t.Fatal(err)
	}

	mainBeta := "package main\n\nimport \"fmt\"\n\nfunc main() { fmt.Println(\"beta\") }\n"
	if err := os.WriteFile(filepath.Join(repoDir, "plugins", "prosthetic-conscience", "tools", "cmd", "beta", "main.go"), []byte(mainBeta), 0o644); err != nil {
		t.Fatal(err)
	}

	// Marketplace manifest
	mktJSON := `{"name":"special-circumstances","plugins":[{"name":"prosthetic-conscience","source":{"type":"git-subdir","url":"https://github.com/ctoforaday/special-circumstances","path":"plugins/prosthetic-conscience","ref":"v1.0.0"}}]}` + "\n"
	if err := os.WriteFile(filepath.Join(repoDir, ".claude-plugin", "marketplace.json"), []byte(mktJSON), 0o644); err != nil {
		t.Fatal(err)
	}

	// Commit initial state in repoDir
	runCmd(t, repoDir, nil, "git", "init", "-b", "main")
	runCmd(t, repoDir, nil, "git", "config", "user.email", "test@example.com")
	runCmd(t, repoDir, nil, "git", "config", "user.name", "Test")
	runCmd(t, repoDir, nil, "git", "add", ".")
	runCmd(t, repoDir, nil, "git", "commit", "-m", "initial")

	// Write stub claude and stub curl
	if err := os.WriteFile(filepath.Join(binDir, "claude"), []byte(stubClaudeScript), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(binDir, "curl"), []byte(stubCurlScript), 0o755); err != nil {
		t.Fatal(err)
	}

	// Build base environment
	var env []string
	for _, kv := range os.Environ() {
		name, _, _ := strings.Cut(kv, "=")
		if strings.HasPrefix(name, "FEOV_") || name == "HOME" || name == "PATH" || name == "WORKDIR" ||
			name == "CLAUDE_CONFIG_DIR" || name == "CLAUDE_CODE_PLUGIN_CACHE_DIR" || name == "XDG_CACHE_HOME" {
			continue
		}
		env = append(env, kv)
	}
	env = append(env,
		"HOME="+scratch,
		"PATH="+binDir+string(os.PathListSeparator)+hostToolsWithout(t),
		"STUB_DIR="+stubDir,
		"GOCACHE="+gocache,
		"GOMODCACHE="+gomod,
	)

	return &testFixture{
		repoDir: repoDir,
		binDir:  binDir,
		stubDir: stubDir,
		scratch: scratch,
		gocache: gocache,
		gomod:   gomod,
		baseEnv: env,
	}
}

// hostToolsWithout is a directory of links to every executable on the host's PATH, the first of
// each name as PATH finds it, less drop. The fixture's PATH is its stubs and this directory, so a
// test takes a tool away whatever the host carries.
func hostToolsWithout(t *testing.T, drop ...string) string {
	t.Helper()
	dir := t.TempDir()
	skip := map[string]bool{}
	for _, d := range drop {
		skip[d] = true
	}
	for _, d := range filepath.SplitList(os.Getenv("PATH")) {
		entries, err := os.ReadDir(d)
		if err != nil {
			continue
		}
		for _, e := range entries {
			name := e.Name()
			dst := filepath.Join(dir, name)
			if skip[name] {
				continue
			}
			if _, err := os.Lstat(dst); err == nil {
				continue
			}
			src := filepath.Join(d, name)
			fi, err := os.Stat(src)
			if err != nil || fi.IsDir() || fi.Mode().Perm()&0o111 == 0 {
				continue
			}
			if err := os.Symlink(src, dst); err != nil {
				t.Fatal(err)
			}
		}
	}
	return dir
}

// addFakeHelper gives the fixture's plugin a helper native/fake/ with an empty ASSETS, the given
// stub native/build.sh, and a stub nativedigest on PATH printing digest; it commits them.
func (f *testFixture) addFakeHelper(t *testing.T, buildSh, digest string) {
	t.Helper()
	native := filepath.Join(f.repoDir, "plugins", "prosthetic-conscience", "native")
	if err := os.MkdirAll(filepath.Join(native, "fake"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(native, "fake", "ASSETS"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(native, "build.sh"), []byte(buildSh), 0o755); err != nil {
		t.Fatal(err)
	}
	f.writeDigest(t, digest)
	f.commit(t, "add native helper fixture")
}

// writeDigest makes the stub nativedigest print d.
func (f *testFixture) writeDigest(t *testing.T, d string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(f.binDir, "nativedigest"), []byte(fmt.Sprintf("#!/bin/sh\necho %q\n", d)), 0o755); err != nil {
		t.Fatal(err)
	}
}

// pathWithout is the fixture's PATH less the host's drop tools, as extra environment for a case
// that takes a tool off PATH.
func (f *testFixture) pathWithout(t *testing.T, drop ...string) string {
	t.Helper()
	return "PATH=" + f.binDir + string(os.PathListSeparator) + hostToolsWithout(t, drop...)
}

// hooksFile is the shape of a plugin's .claude-plugin/hooks.json.
type hooksFile struct {
	Hooks map[string][]hookMatcher `json:"hooks"`
}

type hookMatcher struct {
	Matcher string        `json:"matcher"`
	Hooks   []hookCommand `json:"hooks"`
}

type hookCommand struct {
	Type    string `json:"type"`
	Command string `json:"command"`
}

// guard runs a staged root's PreToolUse guard as hooks.json's exec runs it: its command, read from
// the root's .claude-plugin/hooks.json, by `sh -c`, under env (the environment a session in this
// universe records) and that root's CLAUDE_PLUGIN_ROOT, stdin empty.
func guard(t *testing.T, root string, env []string) {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(root, ".claude-plugin", "hooks.json"))
	if err != nil {
		t.Fatal(err)
	}
	var hf hooksFile
	if err := json.Unmarshal(raw, &hf); err != nil || len(hf.Hooks["PreToolUse"]) == 0 || len(hf.Hooks["PreToolUse"][0].Hooks) == 0 {
		t.Fatalf("%s/.claude-plugin/hooks.json has no PreToolUse guard (%v)", root, err)
	}
	cmd := exec.Command("sh", "-c", hf.Hooks["PreToolUse"][0].Hooks[0].Command)
	cmd.Env = append(append([]string{}, env...), "CLAUDE_PLUGIN_ROOT="+root)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("PreToolUse guard failed: %v\n%s", err, out)
	}
}

// waitFor polls for path for up to 60 s.
func waitFor(path string) bool {
	for deadline := time.Now().Add(60 * time.Second); time.Now().Before(deadline); time.Sleep(100 * time.Millisecond) {
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}
	return false
}

// assertGuardInstallsNothing runs the guard of a root a dead build left, and requires the fetch it
// detaches to fail, leaving .fetch/failed and no bin/alpha: no older release is installed.
func assertGuardInstallsNothing(t *testing.T, root string, env []string) {
	t.Helper()
	guard(t, root, env)
	failed := filepath.Join(root, ".fetch", "failed")
	installed := filepath.Join(root, ".fetch", "installed")
	for deadline := time.Now().Add(60 * time.Second); time.Now().Before(deadline); time.Sleep(100 * time.Millisecond) {
		if _, err := os.Stat(installed); err == nil {
			tag, _ := os.ReadFile(installed)
			t.Fatalf("after the build died, the guard installed a release (.fetch/installed: %s)", strings.TrimSpace(string(tag)))
		}
		if _, err := os.Stat(failed); err == nil {
			break
		}
	}
	if _, err := os.Stat(failed); err != nil {
		t.Fatalf("the guard's fetch left no .fetch/failed within 60 s")
	}
	// The fetch's EXIT trap removes its lock after failed is written; wait for that, so nothing is
	// still writing into the root when bin/ is read.
	lock := filepath.Join(root, ".fetch", "lock")
	for deadline := time.Now().Add(60 * time.Second); time.Now().Before(deadline); time.Sleep(100 * time.Millisecond) {
		if _, err := os.Stat(lock); err != nil {
			break
		}
	}
	if _, err := os.Stat(filepath.Join(root, "bin", "alpha")); err == nil {
		cause, _ := os.ReadFile(failed)
		t.Fatalf("the guard installed bin/alpha after a build died (.fetch/failed: %s)", cause)
	}
}

func runCmd(t *testing.T, dir string, env []string, name string, args ...string) (string, error) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if env != nil {
		cmd.Env = env
	}
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func (f *testFixture) runUniverse(t *testing.T, stdin io.Reader, extraEnv []string, args ...string) (string, int) {
	t.Helper()
	cmd := exec.Command("bash", append([]string{filepath.Join(f.repoDir, "scripts", "universe.sh")}, args...)...)
	cmd.Dir = f.scratch
	envMap := make(map[string]string)
	for _, kv := range f.baseEnv {
		k, v, _ := strings.Cut(kv, "=")
		envMap[k] = v
	}
	for _, kv := range extraEnv {
		k, v, _ := strings.Cut(kv, "=")
		envMap[k] = v
	}
	var cmdEnv []string
	for k, v := range envMap {
		cmdEnv = append(cmdEnv, k+"="+v)
	}
	cmd.Env = cmdEnv
	if stdin != nil {
		cmd.Stdin = stdin
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return string(out), exitErr.ExitCode()
		}
		t.Fatalf("universe.sh %v: %v\n%s", args, err, out)
	}
	return string(out), 0
}

func (f *testFixture) commit(t *testing.T, msg string) {
	t.Helper()
	runCmd(t, f.repoDir, nil, "git", "add", ".")
	runCmd(t, f.repoDir, nil, "git", "commit", "-m", msg)
}

func TestBuildInstallsFromTheLocalRelease(t *testing.T) {
	f := newTestFixture(t)

	// Relative --dir ending in /
	relDir := "rel_wd/"
	absWorkDir := filepath.Join(f.scratch, "rel_wd")

	// stdin a pipe the test holds open
	pr, pw, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer pr.Close()
	defer pw.Close()

	out, code := f.runUniverse(t, pr, nil, "build", "--dir="+relDir)
	if code != 0 {
		t.Fatalf("universe.sh build failed (exit %d):\n%s", code, out)
	}

	// $WORKDIR.release is a sibling of $WORKDIR
	expectedRelease := absWorkDir + ".release"
	if fi, err := os.Stat(expectedRelease); err != nil || !fi.IsDir() {
		t.Fatalf("expected release dir at %s (err: %v)", expectedRelease, err)
	}

	// Verify release did not land inside rel_wd
	if _, err := os.Stat(filepath.Join(absWorkDir, ".release")); err == nil {
		t.Fatalf("release landed at %s/.release instead of sibling", absWorkDir)
	}

	// Each staged root's bin/<n> equals $WORKDIR.release/<p>--v<v>/<n>_linux_<arch>
	hostArch := runtime.GOARCH
	tagDir := filepath.Join(expectedRelease, "prosthetic-conscience--v1.0.0")
	stagedBin := filepath.Join(absWorkDir+".marketplace", "plugins", "prosthetic-conscience", "bin")

	for _, name := range []string{"alpha", "beta"} {
		relAsset := filepath.Join(tagDir, fmt.Sprintf("%s_linux_%s", name, hostArch))
		stagedAsset := filepath.Join(stagedBin, name)

		relBytes, err := os.ReadFile(relAsset)
		if err != nil {
			t.Fatalf("reading release asset %s: %v", relAsset, err)
		}
		stagedBytes, err := os.ReadFile(stagedAsset)
		if err != nil {
			t.Fatalf("reading staged binary %s: %v", stagedAsset, err)
		}
		if !bytes.Equal(relBytes, stagedBytes) {
			t.Errorf("staged binary %s does not equal release asset %s", name, relAsset)
		}
	}

	// .fetch/installed names the tag
	installedPath := filepath.Join(absWorkDir+".marketplace", "plugins", "prosthetic-conscience", ".fetch", "installed")
	instData, err := os.ReadFile(installedPath)
	if err != nil {
		t.Fatalf("reading .fetch/installed: %v", err)
	}
	if got := strings.TrimSpace(string(instData)); got != "prosthetic-conscience--v1.0.0" {
		t.Errorf(".fetch/installed = %q, want prosthetic-conscience--v1.0.0", got)
	}

	// .fetch/log holds fetch-bin.sh's installed ... from <tag>
	logPath := filepath.Join(absWorkDir+".marketplace", "plugins", "prosthetic-conscience", ".fetch", "log")
	logData, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("reading .fetch/log: %v", err)
	}
	if !strings.Contains(string(logData), "installed") || !strings.Contains(string(logData), "alpha") ||
		!strings.Contains(string(logData), "beta") || !strings.Contains(string(logData), "from prosthetic-conscience--v1.0.0") {
		t.Errorf(".fetch/log missing expected log:\n%s", string(logData))
	}

	// SHA256SUMS lists every asset
	sumPath := filepath.Join(tagDir, "SHA256SUMS")
	sumData, err := os.ReadFile(sumPath)
	if err != nil {
		t.Fatalf("reading SHA256SUMS: %v", err)
	}
	for _, name := range []string{"alpha_linux_" + hostArch, "beta_linux_" + hostArch} {
		if !strings.Contains(string(sumData), name) {
			t.Errorf("SHA256SUMS missing %s:\n%s", name, string(sumData))
		}
	}

	// Every URL the curl shim logged is under the release (file://)
	curlLogData, err := os.ReadFile(filepath.Join(f.stubDir, "curl.log"))
	if err != nil {
		t.Fatalf("reading curl.log: %v", err)
	}
	curlLines := strings.Split(strings.TrimSpace(string(curlLogData)), "\n")
	if len(curlLines) == 0 || curlLines[0] == "" {
		t.Fatal("curl.log is empty; curl shim was never invoked")
	}
	for _, line := range curlLines {
		if !strings.HasPrefix(line, "file://") || !strings.Contains(line, "rel_wd.release") {
			t.Errorf("curl logged unexpected URL: %s", line)
		}
	}

	// A second build with sha256sum off PATH writes the same SHA256SUMS
	shadowBin := filepath.Join(f.scratch, "shadow_bin")
	if err := os.MkdirAll(shadowBin, 0o755); err != nil {
		t.Fatal(err)
	}
	if entries, err := os.ReadDir("/usr/bin"); err == nil {
		for _, e := range entries {
			if e.Name() != "sha256sum" && e.Name() != "go" && e.Name() != "gofmt" {
				_ = os.Symlink(filepath.Join("/usr/bin", e.Name()), filepath.Join(shadowBin, e.Name()))
			}
		}
	}
	if goPath, err := exec.LookPath("go"); err == nil {
		_ = os.Symlink(goPath, filepath.Join(shadowBin, "go"))
	}
	if gofmtPath, err := exec.LookPath("gofmt"); err == nil {
		_ = os.Symlink(gofmtPath, filepath.Join(shadowBin, "gofmt"))
	}
	for _, tool := range []string{"claude", "curl"} {
		b, _ := os.ReadFile(filepath.Join(f.binDir, tool))
		_ = os.WriteFile(filepath.Join(shadowBin, tool), b, 0o755)
	}

	var cleanPath []string
	cleanPath = append(cleanPath, shadowBin)
	for _, p := range strings.Split(os.Getenv("PATH"), string(os.PathListSeparator)) {
		if p != "/usr/bin" && p != "/bin" {
			cleanPath = append(cleanPath, p)
		}
	}
	shadowEnv := []string{
		"PATH=" + strings.Join(cleanPath, string(os.PathListSeparator)),
	}

	out2, code2 := f.runUniverse(t, nil, shadowEnv, "build", "--dir="+relDir)
	if code2 != 0 {
		t.Fatalf("second build failed with sha256sum absent (exit %d):\n%s", code2, out2)
	}

	sumData2, err := os.ReadFile(sumPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(sumData) != string(sumData2) {
		t.Errorf("second build SHA256SUMS differs:\ngot:\n%s\nwant:\n%s", string(sumData2), string(sumData))
	}
}

func TestBuildFailsNamingABinaryThatDidNotBuild(t *testing.T) {
	f := newTestFixture(t)

	// Make beta fail to compile
	betaMain := filepath.Join(f.repoDir, "plugins", "prosthetic-conscience", "tools", "cmd", "beta", "main.go")
	if err := os.WriteFile(betaMain, []byte("package main\n broken syntax\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	f.commit(t, "break beta")

	relDir := "wd/"
	absWorkDir := filepath.Join(f.scratch, "wd")

	out, code := f.runUniverse(t, nil, nil, "build", "--dir="+relDir)
	if code == 0 {
		t.Fatalf("build succeeded but beta was broken:\n%s", out)
	}

	// Must name the plugin and quote .fetch/failed's beta_linux_<arch> and the tag
	hostArch := runtime.GOARCH
	wantAsset := fmt.Sprintf("beta_linux_%s", hostArch)
	wantTag := "prosthetic-conscience--v1.0.0"

	if !strings.Contains(out, "prosthetic-conscience") {
		t.Errorf("output missing plugin name prosthetic-conscience:\n%s", out)
	}
	if !strings.Contains(out, wantAsset) {
		t.Errorf("output missing failed asset %s:\n%s", wantAsset, out)
	}
	if !strings.Contains(out, wantTag) {
		t.Errorf("output missing tag %s:\n%s", wantTag, out)
	}

	// Root's bin/ holds no alpha (all or none)
	stagedAlpha := filepath.Join(absWorkDir+".marketplace", "plugins", "prosthetic-conscience", "bin", "alpha")
	if _, err := os.Stat(stagedAlpha); err == nil {
		t.Errorf("bin/alpha exists despite beta build failure (all-or-none violated)")
	}

	// Deletion check: must fail at fetch step, not at ensure assertion
	if strings.Contains(out, "the self-heal still fires in") {
		t.Errorf("failed at ensure assertion instead of fetch step:\n%s", out)
	}
}

func TestEverySessionFetchesOnlyTheLocalRelease(t *testing.T) {
	f := newTestFixture(t)

	relDir := "wd/"
	absWorkDir := filepath.Join(f.scratch, "wd")

	// First build
	if out, code := f.runUniverse(t, nil, nil, "build", "--dir="+relDir); code != 0 {
		t.Fatalf("first build failed (exit %d):\n%s", code, out)
	}
	// Second build
	if out, code := f.runUniverse(t, nil, nil, "build", "--dir="+relDir); code != 0 {
		t.Fatalf("second build failed (exit %d):\n%s", code, out)
	}

	// settings.json holds env.SC_RELEASE_BASE_URL and all four enabledPlugins
	settingsFile := filepath.Join(absWorkDir, "config", "settings.json")
	sData, err := os.ReadFile(settingsFile)
	if err != nil {
		t.Fatalf("reading settings.json: %v", err)
	}
	sStr := string(sData)
	expectedURL := "file://" + absWorkDir + ".release"
	if !strings.Contains(sStr, expectedURL) {
		t.Errorf("settings.json missing %s:\n%s", expectedURL, sStr)
	}
	for _, p := range []string{"prosthetic-conscience", "frank-exchange-of-views", "sleeper-service", "gray-area"} {
		wantPlugin := fmt.Sprintf("%s@special-circumstances", p)
		if !strings.Contains(sStr, wantPlugin) {
			t.Errorf("settings.json missing %s:\n%s", wantPlugin, sStr)
		}
	}

	// Run stub universe.sh shell: records shell_env
	if out, code := f.runUniverse(t, nil, nil, "shell", "--dir="+relDir); code != 0 {
		t.Fatalf("shell failed (exit %d):\n%s", code, out)
	}
	shellEnvData, err := os.ReadFile(filepath.Join(f.stubDir, "shell_env"))
	if err != nil {
		t.Fatalf("reading shell_env: %v", err)
	}
	if !strings.Contains(string(shellEnvData), "SC_RELEASE_BASE_URL="+expectedURL) {
		t.Errorf("shell_env missing SC_RELEASE_BASE_URL=%s:\n%s", expectedURL, string(shellEnvData))
	}

	// With a staged bin/alpha deleted, PreToolUse restores it
	stagedRoot := filepath.Join(absWorkDir+".marketplace", "plugins", "prosthetic-conscience")
	stagedAlpha := filepath.Join(stagedRoot, "bin", "alpha")
	if err := os.Remove(stagedAlpha); err != nil {
		t.Fatalf("removing %s: %v", stagedAlpha, err)
	}

	shellEnvLines := strings.Split(strings.TrimSpace(string(shellEnvData)), "\n")
	guard(t, stagedRoot, shellEnvLines)
	if !waitFor(stagedAlpha) {
		t.Fatalf("staged bin/alpha was not restored by the PreToolUse guard within 60 s")
	}
	if !waitFor(filepath.Join(stagedRoot, ".fetch", "installed")) {
		t.Fatalf("the guard's fetch did not finish within 60 s")
	}

	// With marketplace add failing, env already names the release — written by this build, from a
	// CLAUDE_CONFIG_DIR with no settings.json, so nothing an earlier build wrote can stand in for it.
	if err := os.RemoveAll(filepath.Dir(settingsFile)); err != nil {
		t.Fatal(err)
	}
	failEnv := []string{"STUB_FAIL_MARKETPLACE_ADD=1"}
	out3, code3 := f.runUniverse(t, nil, failEnv, "build", "--dir="+relDir)
	if code3 == 0 {
		t.Fatalf("expected build to fail when marketplace add fails:\n%s", out3)
	}
	sData3, err := os.ReadFile(settingsFile)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(sData3), expectedURL) {
		t.Errorf("settings.json missing %s after marketplace add failure:\n%s", expectedURL, string(sData3))
	}
	// The build died after staging: the roots it staged find no release, never the last build's.
	assertGuardInstallsNothing(t, stagedRoot, shellEnvLines)

	// r10 gap: live $WORKDIR/shell.pid refuses build
	shellPIDFile := filepath.Join(absWorkDir, "shell.pid")
	if err := os.WriteFile(shellPIDFile, []byte(fmt.Sprintf("%d\n", os.Getpid())), 0o644); err != nil {
		t.Fatal(err)
	}
	out4, code4 := f.runUniverse(t, nil, nil, "build", "--dir="+relDir)
	if code4 != 3 || !strings.Contains(out4, "interactive shell is alive") {
		t.Fatalf("expected exit 3 naming shell pid; got %d:\n%s", code4, out4)
	}
	_ = os.Remove(shellPIDFile)

	// r10 gap: with stub claude plugin install writing stale bin/alpha, build ends with bin/alpha equal to release asset
	staleEnv := []string{"STUB_INSTALL_WRITE_STALE=1"}
	out5, code5 := f.runUniverse(t, nil, staleEnv, "build", "--dir="+relDir)
	if code5 != 0 {
		t.Fatalf("build with stale install failed (exit %d):\n%s", code5, out5)
	}
	if written, err := os.ReadFile(filepath.Join(f.stubDir, "stale_written")); err != nil || !strings.Contains(string(written), stagedAlpha) {
		t.Fatalf("the stub install wrote no stale %s, so the case tests nothing (%v)", stagedAlpha, err)
	}
	alphaData, err := os.ReadFile(stagedAlpha)
	if err != nil {
		t.Fatal(err)
	}
	releaseAlpha := filepath.Join(absWorkDir+".release", "prosthetic-conscience--v1.0.0", "alpha_linux_"+runtime.GOARCH)
	want, err := os.ReadFile(releaseAlpha)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(alphaData, want) {
		t.Fatalf("staged bin/alpha (%d bytes) is not the release asset %s (%d bytes) after a stale install", len(alphaData), releaseAlpha, len(want))
	}

	// r10 gap: after a build that carries a helper, a build that dies after staging — flock, a tool
	// native/build.sh runs, off PATH,
	// the helper cache empty — leaves the staged roots fetching only its own unfinished release:
	// the guard installs nothing, not the complete release the first build wrote.
	f.addFakeHelper(t, `#!/usr/bin/env bash
set -e
target="$2"
echo "fake-helper-binary" > "$3/${1}_${target%%-*}_${target##*-}"
`, "digest-r10")
	if out, code := f.runUniverse(t, nil, nil, "build", "--dir="+relDir); code != 0 {
		t.Fatalf("the build carrying the helper failed (exit %d):\n%s", code, out)
	}
	_ = os.RemoveAll(filepath.Join(f.scratch, ".cache", "sc-universe"))
	out6, code6 := f.runUniverse(t, nil, []string{f.pathWithout(t, "flock")}, "build", "--dir="+relDir)
	if code6 == 0 || !strings.Contains(out6, "fake: native/build.sh needs flock on PATH") {
		t.Fatalf("expected the build to die naming flock; exit %d:\n%s", code6, out6)
	}
	if _, err := os.Stat(filepath.Join(stagedRoot, "hooks", "fetch-bin.sh")); err != nil {
		t.Fatalf("the build died before staging, so the case tests nothing: %v\n%s", err, out6)
	}
	assertGuardInstallsNothing(t, stagedRoot, shellEnvLines)
}

func TestBuildAlwaysBuildsTheHelpers(t *testing.T) {
	f := newTestFixture(t)

	// The stub build.sh logs each call and writes the helper asset the release job would name.
	f.addFakeHelper(t, `#!/usr/bin/env bash
set -e
helper="$1"
target="$2"
outdir="$3"
echo "build.sh $helper $target" >> "$STUB_DIR/helper_build.log"
host_os="${target%%-*}"
host_arch="${target##*-}"
ext=""
[ "$host_os" = "windows" ] && ext=".exe"
echo "fake-helper-binary" > "$outdir/${helper}_${host_os}_${host_arch}${ext}"
`, "digest-111122223333")

	relDir := "wd/"
	absWorkDir := filepath.Join(f.scratch, "wd")

	out, code := f.runUniverse(t, nil, nil, "build", "--dir="+relDir)
	if code != 0 {
		t.Fatalf("build failed (exit %d):\n%s", code, out)
	}

	helperAsset := filepath.Join(absWorkDir+".release", "prosthetic-conscience--v1.0.0", "fake_linux_"+runtime.GOARCH)
	relData, err := os.ReadFile(helperAsset)
	if err != nil {
		t.Fatalf("local release missing helper asset %s: %v", helperAsset, err)
	}
	stagedData, err := os.ReadFile(filepath.Join(absWorkDir+".marketplace", "plugins", "prosthetic-conscience", "bin", "fake"))
	if err != nil {
		t.Fatalf("staged bin/fake missing: %v", err)
	}
	if !bytes.Equal(relData, stagedData) {
		t.Errorf("staged bin/fake does not equal the release's helper")
	}

	// With a tool native/build.sh runs off PATH (flock; it fetches its own zig and cmake) and the
	// cache empty, build exits non-zero naming it, before running build.sh.
	_ = os.RemoveAll(filepath.Join(f.scratch, ".cache", "sc-universe"))
	calls := countCalls(t, filepath.Join(f.stubDir, "helper_build.log"))
	out2, code2 := f.runUniverse(t, nil, []string{f.pathWithout(t, "flock")}, "build", "--dir="+relDir)
	if code2 == 0 {
		t.Fatalf("build succeeded with flock missing:\n%s", out2)
	}
	if !strings.Contains(out2, "fake: native/build.sh needs flock on PATH") {
		t.Errorf("the build's error does not name the missing flock:\n%s", out2)
	}
	if n := countCalls(t, filepath.Join(f.stubDir, "helper_build.log")); n != calls {
		t.Errorf("build.sh ran without its toolchain (%d calls, was %d)", n, calls)
	}
}

func TestBuildReusesTheCachedHelper(t *testing.T) {
	f := newTestFixture(t)
	buildLog := filepath.Join(f.stubDir, "helper_build.log")

	// The stub build.sh logs each call, then sleeps when STUB_BUILD_SLEEP is set, so a build can be
	// killed while it runs.
	f.addFakeHelper(t, `#!/usr/bin/env bash
set -e
helper="$1"
target="$2"
outdir="$3"
echo "call" >> "$STUB_DIR/helper_build.log"
if [ "${STUB_BUILD_SLEEP:-}" = "1" ]; then
  sleep 60
fi
host_os="${target%%-*}"
host_arch="${target##*-}"
ext=""
[ "$host_os" = "windows" ] && ext=".exe"
echo "fake-helper-binary" > "$outdir/${helper}_${host_os}_${host_arch}${ext}"
`, "digest-1")

	relDir := "wd/"
	absWorkDir := filepath.Join(f.scratch, "wd")
	cacheHome := filepath.Join(f.scratch, "cache")
	cacheEnv := []string{"XDG_CACHE_HOME=" + cacheHome}
	entry := func(d string) string { return filepath.Join(cacheHome, "sc-universe", "native", "fake", d) }

	// 1st build: compiles the helper (one call).
	if out, code := f.runUniverse(t, nil, cacheEnv, "build", "--dir="+relDir); code != 0 {
		t.Fatalf("1st build failed (exit %d):\n%s", code, out)
	}
	if n := countCalls(t, buildLog); n != 1 {
		t.Fatalf("expected 1 build.sh call, got %d", n)
	}

	// 2nd build, after the three paths are deleted, with the digest unchanged: reuses the cache.
	for _, p := range []string{absWorkDir, absWorkDir + ".marketplace", absWorkDir + ".release"} {
		_ = os.RemoveAll(p)
	}
	if out, code := f.runUniverse(t, nil, cacheEnv, "build", "--dir="+relDir); code != 0 {
		t.Fatalf("2nd build failed (exit %d):\n%s", code, out)
	}
	if n := countCalls(t, buildLog); n != 1 {
		t.Fatalf("expected the cached helper to be reused, but build.sh ran %d times", n)
	}

	// 3rd build, a new digest: a second call, written under the printed digest.
	f.writeDigest(t, "digest-2")
	if out, code := f.runUniverse(t, nil, cacheEnv, "build", "--dir="+relDir); code != 0 {
		t.Fatalf("3rd build failed (exit %d):\n%s", code, out)
	}
	if n := countCalls(t, buildLog); n != 2 {
		t.Fatalf("expected a 2nd build.sh call on the digest change, got %d", n)
	}
	if fi, err := os.Stat(entry("digest-2")); err != nil || !fi.IsDir() {
		t.Fatalf("expected a cache entry at %s", entry("digest-2"))
	}

	// 4th build, a third digest, killed while the stub build.sh sleeps: the kill lands only once
	// the stub has logged its call, so the build is mid-helper when it dies.
	f.writeDigest(t, "digest-3")
	cmd4 := exec.Command("bash", filepath.Join(f.repoDir, "scripts", "universe.sh"), "build", "--dir="+relDir)
	cmd4.Dir = f.scratch
	cmd4.Env = append(append([]string{}, f.baseEnv...), append(cacheEnv, "STUB_BUILD_SLEEP=1")...)
	cmd4.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	var out4 bytes.Buffer
	cmd4.Stdout, cmd4.Stderr = &out4, &out4
	if err := cmd4.Start(); err != nil {
		t.Fatal(err)
	}
	inBuild := false
	for deadline := time.Now().Add(120 * time.Second); time.Now().Before(deadline); time.Sleep(50 * time.Millisecond) {
		if countCalls(t, buildLog) == 3 {
			inBuild = true
			break
		}
	}
	// The whole group: universe.sh, the stub build.sh and its sleep.
	_ = syscall.Kill(-cmd4.Process.Pid, syscall.SIGKILL)
	_ = cmd4.Wait()
	if !inBuild {
		t.Fatalf("the 4th build never reached build.sh within 120 s:\n%s", out4.String())
	}
	if _, err := os.Stat(entry("digest-3")); err == nil {
		t.Fatalf("the killed build left a cache entry %s", entry("digest-3"))
	}

	// 5th build: builds the helper again and finishes.
	if out, code := f.runUniverse(t, nil, cacheEnv, "build", "--dir="+relDir); code != 0 {
		t.Fatalf("5th build failed (exit %d):\n%s", code, out)
	}
	if n := countCalls(t, buildLog); n != 4 {
		t.Fatalf("expected the 5th build to call build.sh (4 calls in all), got %d", n)
	}
	if fi, err := os.Stat(entry("digest-3")); err != nil || !fi.IsDir() {
		t.Fatalf("expected a cache entry at %s after the 5th build", entry("digest-3"))
	}
}

func TestBuildFetchesTheHelpersPinnedExport(t *testing.T) {
	f := newTestFixture(t)
	f.addFakeHelper(t, `#!/usr/bin/env bash
set -e
target="$2"
echo "call" >> "$STUB_DIR/helper_build.log"
echo "fake-helper-binary" > "$3/${1}_${target%%-*}_${target##*-}"
`, "digest-export-1")

	// The export's source, served by the curl shim from a file:// URL, and its pin in the helper's
	// PINS.txt as native/build.sh reads it: <sha256> <file> <url>.
	arch := runtime.GOARCH
	export := "model." + arch + ".bin"
	src := filepath.Join(f.scratch, "upstream", export)
	body := "model weights for " + arch + "\n"
	if err := os.MkdirAll(filepath.Dir(src), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(src, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	native := filepath.Join(f.repoDir, "plugins", "prosthetic-conscience", "native", "fake")
	if err := os.WriteFile(filepath.Join(native, "ASSETS"), []byte("model.bin amd64=model.amd64.bin arm64=model.arm64.bin\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	pins := fmt.Sprintf("# <sha256>  <file>  <url>\n%x  %s  file://%s\n", sha256.Sum256([]byte(body)), export, src)
	if err := os.WriteFile(filepath.Join(native, "PINS.txt"), []byte(pins), 0o644); err != nil {
		t.Fatal(err)
	}
	f.commit(t, "the fake helper ships an export")

	relDir := "wd/"
	absWorkDir := filepath.Join(f.scratch, "wd")
	cacheHome := filepath.Join(f.scratch, "cache")
	cacheEnv := []string{"XDG_CACHE_HOME=" + cacheHome}
	entry := func(d string) string { return filepath.Join(cacheHome, "sc-universe", "native", "fake", d) }

	if out, code := f.runUniverse(t, nil, cacheEnv, "build", "--dir="+relDir); code != 0 {
		t.Fatalf("build failed (exit %d):\n%s", code, out)
	}
	for _, p := range []string{
		filepath.Join(absWorkDir+".release", "prosthetic-conscience--v1.0.0", export),
		filepath.Join(absWorkDir+".marketplace", "plugins", "prosthetic-conscience", "bin", "model.bin"),
		filepath.Join(entry("digest-export-1"), export),
	} {
		if got, err := os.ReadFile(p); err != nil || string(got) != body {
			t.Errorf("%s = %q (%v), want the pinned export's bytes", p, got, err)
		}
	}

	// An export with no pin fails the build naming it, and no cache entry is written without it.
	if err := os.WriteFile(filepath.Join(native, "PINS.txt"), []byte("# <sha256>  <file>  <url>\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	f.writeDigest(t, "digest-export-2")
	f.commit(t, "the export loses its pin")
	out, code := f.runUniverse(t, nil, cacheEnv, "build", "--dir="+relDir)
	if code == 0 || !strings.Contains(out, "the export "+export+" has no pin") {
		t.Fatalf("expected the build to fail naming the unpinned %s; exit %d:\n%s", export, code, out)
	}
	if _, err := os.Stat(entry("digest-export-2")); err == nil {
		t.Fatalf("a cache entry was written without its export: %s", entry("digest-export-2"))
	}
}

func countCalls(t *testing.T, logPath string) int {
	t.Helper()
	data, err := os.ReadFile(logPath)
	if err != nil {
		return 0
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	count := 0
	for _, l := range lines {
		if strings.TrimSpace(l) != "" {
			count++
		}
	}
	return count
}
