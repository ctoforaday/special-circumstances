package installshape

// THE CLIENTS RUN WITH NO CREDENTIAL THEY COULD BILL, AND THAT IS ENFORCED, NOT HOPED.
//
// Three layers, each sufficient for the environment variables and the first two for files:
//  1. The test refuses to start while the process holds any ANTHROPIC_* variable or a Claude Code
//     token, so a developer's shell cannot lend one even by accident (refuseCredentialEnv).
//  2. Every client process gets an environment built from an ALLOWLIST (box.env), never a copy of
//     os.Environ(): nothing the allowlist does not name can reach it.
//  3. Every config, cache and data root points into a fresh temp HOME, which the test checks holds
//     no credential file before and after each client call (box.refuseCredentialFiles), and the
//     client itself must report `loggedIn: false` with no auth method (box.assertLoggedOut), which
//     is the check that catches a machine-wide fallback the first two cannot see.
//
// A no-login `claude -p` exits non-zero after its SessionStart hooks fire; one that exits 0 ran a
// model turn, and the test fails on it.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"
)

// pins is clients.json: the client versions the scripts job installs and this test holds them to.
type pins struct {
	Claude struct {
		Version string `json:"version"`
	} `json:"claude"`
	Agy struct {
		Version string `json:"version"`
		URL     string `json:"url"`
		SHA512  string `json:"sha512"`
	} `json:"agy"`
}

func readPins(t *testing.T) pins {
	t.Helper()
	b, err := os.ReadFile("clients.json")
	if err != nil {
		t.Fatalf("reading the client pins: %v", err)
	}
	var p pins
	if err := json.Unmarshal(b, &p); err != nil {
		t.Fatalf("clients.json: %v", err)
	}
	if p.Claude.Version == "" || p.Agy.Version == "" || p.Agy.URL == "" || p.Agy.SHA512 == "" {
		t.Fatalf("clients.json pins no claude version, or no agy version/url/sha512: %+v", p)
	}
	return p
}

// credentialEnv reports whether an environment variable could carry a credential to a client.
func credentialEnv(key string) bool {
	switch key {
	case "CLAUDE_CODE_OAUTH_TOKEN", "CLAUDE_CODE_OAUTH_REFRESH_TOKEN", "CLAUDE_CODE_USE_BEDROCK",
		"CLAUDE_CODE_USE_VERTEX", "AWS_BEARER_TOKEN_BEDROCK", "GEMINI_API_KEY", "GOOGLE_API_KEY":
		return true
	}
	return strings.HasPrefix(key, "ANTHROPIC_")
}

// refuseCredentialEnv fails the test, naming each variable (never its value), while the test
// process holds anything that could authenticate a client.
func refuseCredentialEnv(t *testing.T) {
	t.Helper()
	var held []string
	for _, kv := range os.Environ() {
		k, _, _ := strings.Cut(kv, "=")
		if credentialEnv(k) {
			held = append(held, k)
		}
	}
	if len(held) > 0 {
		sort.Strings(held)
		t.Fatalf("REFUSING TO RUN: the test process holds %s, which could authenticate a client and bill it. Unset them and re-run.", strings.Join(held, ", "))
	}
}

// client is one resolved client binary and the version it reports.
type client struct {
	name, path, version string
}

// resolveClient finds a client on PATH or in ~/.local/bin, where both installers put it. On CI a
// missing client fails; locally it skips, naming the client.
func resolveClient(t *testing.T, name string) client {
	t.Helper()
	p, err := exec.LookPath(name)
	if err != nil {
		if home, herr := os.UserHomeDir(); herr == nil {
			cand := filepath.Join(home, ".local", "bin", name)
			if fi, serr := os.Stat(cand); serr == nil && !fi.IsDir() {
				p, err = cand, nil
			}
		}
	}
	if err != nil {
		if os.Getenv("CI") == "true" {
			t.Fatalf("%s is not installed on CI: the install-shape contract must run on every PR", name)
		}
		t.Skipf("%s is not installed (PATH or ~/.local/bin): the install-shape contract did not run", name)
	}
	out, err := exec.Command(p, "--version").Output()
	if err != nil {
		t.Fatalf("%s --version: %v", p, err)
	}
	fields := strings.Fields(string(out))
	if len(fields) == 0 {
		t.Fatalf("%s --version printed nothing", p)
	}
	return client{name: name, path: p, version: fields[0]}
}

// box is one isolated client environment: its own HOME, Claude config dir and agy config, and a
// PATH holding only the two clients and the system directories.
type box struct {
	t       *testing.T
	label   string // the shape, prefixed to every failure
	home    string
	config  string // CLAUDE_CONFIG_DIR: deliberately NOT $HOME/.claude, which holds the host-cache decoy
	work    string // the session's working directory
	bindir  string // PATH entry holding the clients
	release string // SC_RELEASE_BASE_URL
	claude  client
	agy     client
}

func newBox(t *testing.T, label, root, releaseURL string, claude, agy client) *box {
	t.Helper()
	b := &box{
		t:       t,
		label:   label,
		home:    filepath.Join(root, "home"),
		bindir:  filepath.Join(root, "clients"),
		work:    filepath.Join(root, "work"),
		release: releaseURL,
		claude:  claude,
		agy:     agy,
	}
	b.config = filepath.Join(b.home, ".claude-config")
	for _, d := range []string{b.home, b.config, b.work, b.bindir, filepath.Join(b.home, "tmp")} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	for _, c := range []client{claude, agy} {
		if err := os.Symlink(c.path, filepath.Join(b.bindir, c.name)); err != nil {
			t.Fatal(err)
		}
	}
	b.refuseCredentialFiles()
	return b
}

// env is the client environment, built from an allowlist. extra entries are appended verbatim.
func (b *box) env(extra ...string) []string {
	env := []string{
		"PATH=" + b.bindir + string(os.PathListSeparator) + "/usr/local/bin:/usr/bin:/bin",
		"HOME=" + b.home,
		"USERPROFILE=" + b.home,
		"CLAUDE_CONFIG_DIR=" + b.config,
		"XDG_CONFIG_HOME=" + filepath.Join(b.home, ".config"),
		"XDG_CACHE_HOME=" + filepath.Join(b.home, ".cache"),
		"XDG_DATA_HOME=" + filepath.Join(b.home, ".local", "share"),
		"XDG_STATE_HOME=" + filepath.Join(b.home, ".local", "state"),
		"TMPDIR=" + filepath.Join(b.home, "tmp"),
		"LANG=C.UTF-8",
		"CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1",
		// Each client's own updater switch: an update replaces the binary under test in place.
		// agy honours its variable only at the exact value "true".
		"DISABLE_AUTOUPDATER=1",
		"AGY_CLI_DISABLE_AUTO_UPDATE=true",
		// A hook or the doctor that runs `go` must not fetch a toolchain into the temp HOME.
		"GOTOOLCHAIN=local",
		"GOFLAGS=-modcacherw",
		"SC_RELEASE_BASE_URL=" + b.release,
	}
	return append(env, extra...)
}

// credentialFiles are where a client could find a credential under this box's roots.
func (b *box) credentialFiles() []string {
	return []string{
		filepath.Join(b.config, ".credentials.json"),
		filepath.Join(b.home, ".claude", ".credentials.json"),
		filepath.Join(b.home, ".config", "anthropic"),
		filepath.Join(b.home, ".anthropic"),
		filepath.Join(b.home, ".gemini", "oauth_creds.json"),
		filepath.Join(b.home, ".gemini", "google_accounts.json"),
		filepath.Join(b.home, ".aws"),
		filepath.Join(b.home, ".config", "gcloud"),
	}
}

func (b *box) refuseCredentialFiles() {
	b.t.Helper()
	for _, f := range b.credentialFiles() {
		if _, err := os.Lstat(f); err == nil {
			b.t.Fatalf("%s: REFUSING TO CONTINUE: %s exists in the isolated HOME and could authenticate a client", b.label, f)
		}
	}
}

// result is one finished client or hook process.
type result struct {
	stdout, stderr string
	code           int
}

// exec runs a process in this box and returns how it ended; a failure to start is fatal.
func (b *box) exec(dir string, env []string, stdin string, timeout time.Duration, name string, args ...string) result {
	b.t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Env = env
	cmd.Dir = dir
	cmd.Stdin = strings.NewReader(stdin)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	r := result{stdout: stdout.String(), stderr: stderr.String()}
	var ee *exec.ExitError
	switch {
	case err == nil:
	case errors.As(err, &ee):
		r.code = ee.ExitCode()
	default:
		b.t.Fatalf("%s: %s %s: %v", b.label, filepath.Base(name), strings.Join(args, " "), err)
	}
	if ctx.Err() != nil {
		b.t.Fatalf("%s: %s %s timed out after %s (check `uptime` before blaming the code)\nstdout:\n%s\nstderr:\n%s", b.label, filepath.Base(name), strings.Join(args, " "), timeout, r.stdout, r.stderr)
	}
	return r
}

// must runs a client subcommand that has to succeed, and returns its stdout.
func (b *box) must(c client, args ...string) string {
	b.t.Helper()
	r := b.exec(b.work, b.env(), "", 3*time.Minute, c.path, args...)
	b.refuseCredentialFiles()
	if r.code != 0 {
		b.t.Fatalf("%s: %s %s exited %d\nstdout:\n%s\nstderr:\n%s", b.label, c.name, strings.Join(args, " "), r.code, r.stdout, r.stderr)
	}
	return r.stdout
}

// assertLoggedOut holds the client to its own answer: no login, no auth method. An isolated config
// dir is not proof of that: a machine-wide token resolves from outside it.
func (b *box) assertLoggedOut() {
	b.t.Helper()
	r := b.exec(b.work, b.env(), "", time.Minute, b.claude.path, "auth", "status", "--json")
	var st struct {
		LoggedIn   *bool  `json:"loggedIn"`
		AuthMethod string `json:"authMethod"`
	}
	if err := json.Unmarshal([]byte(r.stdout), &st); err != nil || st.LoggedIn == nil {
		b.t.Fatalf("%s: REFUSING TO CONTINUE: `claude auth status --json` did not say whether it is logged in (%v):\n%s%s", b.label, err, r.stdout, r.stderr)
	}
	if *st.LoggedIn || st.AuthMethod != "none" {
		b.t.Fatalf("%s: REFUSING TO CONTINUE: the isolated client reports loggedIn=%v authMethod=%q, so a session could bill", b.label, *st.LoggedIn, st.AuthMethod)
	}
}

// session runs one no-login `claude -p`. Its SessionStart, UserPromptSubmit and SessionEnd hooks
// fire and it then exits non-zero; exit 0 means a model turn ran, which this test must never cause.
func (b *box) session(args ...string) {
	b.t.Helper()
	b.assertLoggedOut()
	full := append(append([]string{}, args...), "-p", "installshape: no model turn may run")
	r := b.exec(b.work, b.env(), "", 3*time.Minute, b.claude.path, full...)
	b.refuseCredentialFiles()
	if r.code == 0 {
		b.t.Fatalf("%s: BILLING GUARD: `claude -p` exited 0, so a model turn ran\nstdout:\n%s\nstderr:\n%s", b.label, r.stdout, r.stderr)
	}
	if !strings.Contains(r.stdout+r.stderr, "Not logged in") {
		b.t.Fatalf("%s: `claude -p` exited %d without reporting \"Not logged in\"; this test only knows the no-login session\nstdout:\n%s\nstderr:\n%s", b.label, r.code, r.stdout, r.stderr)
	}
}

// supported says why this host cannot run the contract, or "".
func supported() string {
	switch {
	case runtime.GOOS != "linux":
		// Windows: neither client's install on a runner is measured, and the hook commands are
		// sh. macOS: the client's credential store is the login keychain, which no temp HOME
		// isolates, so the no-billing guarantee would not hold there.
		return "the install-shape contract runs on Linux only (" + runtime.GOOS + ")"
	case raceBuild:
		return "-race: the contract drives subprocesses only; the plain test run covers it"
	}
	return ""
}
