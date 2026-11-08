//go:build linux

package universetest

import (
	"bytes"
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
          mkt_root="${WORKDIR}.marketplace/plugins/$pname"
          if [ -d "$mkt_root" ]; then
            mkdir -p "$mkt_root/bin" "$mkt_root/.fetch"
            echo "stale" > "$mkt_root/bin/alpha"
            echo "prosthetic-conscience--v1.0.0" > "$mkt_root/.fetch/installed"
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

	// Plugin hooks.json
	hooksJSON := `{"hooks":[{"event":"PreToolUse","command":"hooks/fetch-bin.sh ensure PreToolUse"}]}` + "\n"
	if err := os.WriteFile(filepath.Join(repoDir, "plugins", "prosthetic-conscience", "hooks", "hooks.json"), []byte(hooksJSON), 0o644); err != nil {
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
		"PATH="+binDir+string(os.PathListSeparator)+os.Getenv("PATH"),
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
	shellEnvLines = append(shellEnvLines, "CLAUDE_PLUGIN_ROOT="+stagedRoot)

	// Run guard: sh -c hooks/fetch-bin.sh ensure PreToolUse
	guardCmd := exec.Command("sh", "-c", filepath.Join(stagedRoot, "hooks", "fetch-bin.sh")+" ensure PreToolUse </dev/null")
	guardCmd.Env = shellEnvLines
	if out, err := guardCmd.CombinedOutput(); err != nil {
		t.Fatalf("PreToolUse guard failed: %v\n%s", err, string(out))
	}

	// Wait up to 5 seconds for restoration
	restored := false
	for i := 0; i < 50; i++ {
		if _, err := os.Stat(stagedAlpha); err == nil {
			restored = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if !restored {
		t.Fatalf("staged bin/alpha was not restored by PreToolUse guard")
	}

	// With marketplace add failing, env already names the release
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
	alphaData, _ := os.ReadFile(stagedAlpha)
	if string(alphaData) == "stale\n" || string(alphaData) == "stale" {
		t.Fatalf("staged bin/alpha remained stale after build")
	}
}

func TestBuildAlwaysBuildsTheHelpers(t *testing.T) {
	f := newTestFixture(t)

	// Add helper native/fake/ with ASSETS and build.sh
	nativeDir := filepath.Join(f.repoDir, "plugins", "prosthetic-conscience", "native", "fake")
	if err := os.MkdirAll(nativeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nativeDir, "ASSETS"), []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}

	// Stub build.sh in native/
	buildSh := `#!/usr/bin/env bash
set -e
helper="$1"
target="$2"
outdir="$3"
echo "build.sh $helper $target" >> "$STUB_DIR/helper_build.log"
host_os="${target%%-*}"
host_arch="${target##*-}"
ext=""
[ "$host_os" = "windows" ] && ext=".exe"
out="${helper}_${host_os}_${host_arch}${ext}"
echo "fake-helper-binary" > "$outdir/$out"
`
	nativeBuildSh := filepath.Join(f.repoDir, "plugins", "prosthetic-conscience", "native", "build.sh")
	if err := os.WriteFile(nativeBuildSh, []byte(buildSh), 0o755); err != nil {
		t.Fatal(err)
	}

	// Stub nativedigest on PATH
	nativedigestSh := `#!/usr/bin/env bash
echo "digest-111122223333"
`
	if err := os.WriteFile(filepath.Join(f.binDir, "nativedigest"), []byte(nativedigestSh), 0o755); err != nil {
		t.Fatal(err)
	}

	// Stub zig and cmake on PATH
	stubZig := filepath.Join(f.binDir, "zig")
	stubCmake := filepath.Join(f.binDir, "cmake")
	_ = os.WriteFile(stubZig, []byte("#!/bin/sh\nexit 0\n"), 0o755)
	_ = os.WriteFile(stubCmake, []byte("#!/bin/sh\nexit 0\n"), 0o755)

	f.commit(t, "add native helper fixture")

	relDir := "wd/"
	absWorkDir := filepath.Join(f.scratch, "wd")

	out, code := f.runUniverse(t, nil, nil, "build", "--dir="+relDir)
	if code != 0 {
		t.Fatalf("build failed (exit %d):\n%s", code, out)
	}

	hostArch := runtime.GOARCH
	helperAsset := filepath.Join(absWorkDir+".release", "prosthetic-conscience--v1.0.0", fmt.Sprintf("fake_linux_%s", hostArch))
	if _, err := os.Stat(helperAsset); err != nil {
		t.Fatalf("local release missing helper asset %s: %v", helperAsset, err)
	}

	stagedFake := filepath.Join(absWorkDir+".marketplace", "plugins", "prosthetic-conscience", "bin", "fake")
	relData, err := os.ReadFile(helperAsset)
	if err != nil {
		t.Fatal(err)
	}
	stagedData, err := os.ReadFile(stagedFake)
	if err != nil {
		t.Fatalf("staged bin/fake missing: %v", err)
	}
	if !bytes.Equal(relData, stagedData) {
		t.Errorf("staged bin/fake does not equal release helper")
	}

	// With zig off PATH and cache empty, build exits non-zero naming zig
	_ = os.Remove(stubZig)
	// Clear cache and release
	cacheDir := filepath.Join(f.scratch, ".cache", "sc-universe")
	_ = os.RemoveAll(cacheDir)
	_ = os.RemoveAll(absWorkDir + ".release")

	out2, code2 := f.runUniverse(t, nil, nil, "build", "--dir="+relDir)
	if code2 == 0 {
		t.Fatalf("build succeeded with zig missing:\n%s", out2)
	}
	if !strings.Contains(out2, "zig") {
		t.Errorf("error output did not name missing tool zig:\n%s", out2)
	}
}

func TestBuildReusesTheCachedHelper(t *testing.T) {
	f := newTestFixture(t)

	nativeDir := filepath.Join(f.repoDir, "plugins", "prosthetic-conscience", "native", "fake")
	if err := os.MkdirAll(nativeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nativeDir, "ASSETS"), []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}

	buildSh := `#!/usr/bin/env bash
set -e
helper="$1"
target="$2"
outdir="$3"
echo "call" >> "$STUB_DIR/helper_build.log"
if [ "${STUB_BUILD_SLEEP:-}" = "1" ]; then
  sleep 10
fi
host_os="${target%%-*}"
host_arch="${target##*-}"
ext=""
[ "$host_os" = "windows" ] && ext=".exe"
out="${helper}_${host_os}_${host_arch}${ext}"
echo "fake-helper-binary" > "$outdir/$out"
`
	nativeBuildSh := filepath.Join(f.repoDir, "plugins", "prosthetic-conscience", "native", "build.sh")
	if err := os.WriteFile(nativeBuildSh, []byte(buildSh), 0o755); err != nil {
		t.Fatal(err)
	}

	digestScript := filepath.Join(f.binDir, "nativedigest")
	writeDigest := func(d string) {
		content := fmt.Sprintf("#!/bin/sh\necho %q\n", d)
		_ = os.WriteFile(digestScript, []byte(content), 0o755)
	}
	writeDigest("digest-1")

	_ = os.WriteFile(filepath.Join(f.binDir, "zig"), []byte("#!/bin/sh\nexit 0\n"), 0o755)
	_ = os.WriteFile(filepath.Join(f.binDir, "cmake"), []byte("#!/bin/sh\nexit 0\n"), 0o755)

	f.commit(t, "native helper cache fixture")

	relDir := "wd/"
	absWorkDir := filepath.Join(f.scratch, "wd")
	mktDir := absWorkDir + ".marketplace"
	releaseDir := absWorkDir + ".release"
	cacheHome := filepath.Join(f.scratch, "cache")
	cacheEnv := []string{"XDG_CACHE_HOME=" + cacheHome}

	// 1st build: compiles helper (logs 1 call)
	if out, code := f.runUniverse(t, nil, cacheEnv, "build", "--dir="+relDir); code != 0 {
		t.Fatalf("1st build failed (exit %d):\n%s", code, out)
	}
	calls1 := countCalls(t, filepath.Join(f.stubDir, "helper_build.log"))
	if calls1 != 1 {
		t.Fatalf("expected 1 build.sh call, got %d", calls1)
	}

	// Delete the three paths: WORKDIR, WORKDIR.marketplace, WORKDIR.release
	_ = os.RemoveAll(absWorkDir)
	_ = os.RemoveAll(mktDir)
	_ = os.RemoveAll(releaseDir)

	// 2nd build: with digest unchanged, reuses cache (build.sh logged once)
	if out, code := f.runUniverse(t, nil, cacheEnv, "build", "--dir="+relDir); code != 0 {
		t.Fatalf("2nd build failed (exit %d):\n%s", code, out)
	}
	calls2 := countCalls(t, filepath.Join(f.stubDir, "helper_build.log"))
	if calls2 != 1 {
		t.Fatalf("expected cached build.sh not to be called again, but got %d calls", calls2)
	}

	// 3rd build: stub nativedigest prints new digest, logs second call and writes cache
	writeDigest("digest-2")
	if out, code := f.runUniverse(t, nil, cacheEnv, "build", "--dir="+relDir); code != 0 {
		t.Fatalf("3rd build failed (exit %d):\n%s", code, out)
	}
	calls3 := countCalls(t, filepath.Join(f.stubDir, "helper_build.log"))
	if calls3 != 2 {
		t.Fatalf("expected 2nd build.sh call on digest change, got %d calls", calls3)
	}
	expectedCache2 := filepath.Join(cacheHome, "sc-universe", "native", "fake", "digest-2")
	if fi, err := os.Stat(expectedCache2); err != nil || !fi.IsDir() {
		t.Fatalf("expected cache dir at %s", expectedCache2)
	}

	// 4th build: killed while stub build.sh sleeps on digest-3
	writeDigest("digest-3")
	cmd4 := exec.Command("bash", filepath.Join(f.repoDir, "scripts", "universe.sh"), "build", "--dir="+relDir)
	cmd4.Dir = f.scratch
	cmd4.Env = append(append([]string{}, f.baseEnv...), append(cacheEnv, "STUB_BUILD_SLEEP=1")...)
	if err := cmd4.Start(); err != nil {
		t.Fatal(err)
	}
	time.Sleep(300 * time.Millisecond)
	_ = syscall.Kill(cmd4.Process.Pid, syscall.SIGKILL)
	_ = cmd4.Wait()

	// Leaves no digest-3 directory in cache
	expectedCache3 := filepath.Join(cacheHome, "sc-universe", "native", "fake", "digest-3")
	if _, err := os.Stat(expectedCache3); err == nil {
		t.Fatalf("partial cache entry %s was left by killed build", expectedCache3)
	}

	// 5th build: logs a call and finishes
	if out, code := f.runUniverse(t, nil, cacheEnv, "build", "--dir="+relDir); code != 0 {
		t.Fatalf("5th build failed (exit %d):\n%s", code, out)
	}
	calls5 := countCalls(t, filepath.Join(f.stubDir, "helper_build.log"))
	if calls5 != 3 {
		t.Fatalf("expected 3rd call on 5th build, got %d", calls5)
	}
	if fi, err := os.Stat(expectedCache3); err != nil || !fi.IsDir() {
		t.Fatalf("expected cache dir at %s after 5th build", expectedCache3)
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
