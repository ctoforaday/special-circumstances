package testhome

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gofrs/flock"
)

// DefaultMinTokenLifetime is the minimum remaining validity required before lending a host OAuth
// access token into an isolated test HOME (matches scripts/universe.sh AGY_TOKEN_MIN_MINUTES=15,
// exceeding any single HOOKCORE_LIVE test timeout so a lent token never expires mid-run).
const DefaultMinTokenLifetime = 15 * time.Minute

// DefaultMaxScanBytes is the per-file ceiling for HostRefreshGuard.CheckTempHome (64 MiB — large
// enough to scan staged Go hook binaries under fakeHome/bin/ while still failing closed "not
// scanned" on arbitrarily huge files).
const DefaultMaxScanBytes int64 = 64 << 20

var (
	liveGoCacheOnce sync.Once
	liveGoCache     string
	liveGoModCache  string
)

func hostGoCaches() (string, string) {
	liveGoCacheOnce.Do(func() {
		goBin, err := exec.LookPath("go")
		if err != nil {
			goBin = "/usr/local/go/bin/go"
		}
		out, err := exec.Command(goBin, "env", "GOCACHE", "GOMODCACHE").Output()
		if err == nil {
			lines := strings.Split(strings.TrimSpace(string(out)), "\n")
			if len(lines) >= 1 {
				liveGoCache = strings.TrimSpace(lines[0])
			}
			if len(lines) >= 2 {
				liveGoModCache = strings.TrimSpace(lines[1])
			}
		}
	})
	return liveGoCache, liveGoModCache
}

// LiveTestingT is the subset of *testing.T needed by NewLiveHome, LiveEnv, LiveClaudeEnv, and
// GuardLiveTempHome so package testhome does not import testing.
type LiveTestingT interface {
	Helper()
	TempDir() string
	Cleanup(func())
	Errorf(format string, args ...any)
	Fatalf(format string, args ...any)
}

// NewLiveHome creates an isolated temporary HOME for a live harness test and registers the #1890
// structural refresh-token guard on it by construction.
func NewLiveHome(t LiveTestingT) string {
	t.Helper()
	dir := t.TempDir()
	GuardLiveTempHome(t, dir)
	return dir
}

// LiveEnv builds an isolated child process environment anchored at fakeHome, asserts that the
// resulting environment carries no path into the host home and no host refresh-token bytes, and
// registers the #1890 structural guard on fakeHome and any external CLAUDE_CONFIG_DIR in extra.
//
// It strips HOME, USERPROFILE, XDG_*, CLAUDE_*, ANTHROPIC_*, ANTIGRAVITY_*, FEOV_*, PLUGIN_ROOT,
// PLUGIN_DATA, GOCACHE, GOMODCACHE, and GOFLAGS so neither harness credentials nor parent session
// attribution leak into the child, disables both clients' auto-updaters, pins host Go caches with
// GOFLAGS=-modcacherw, and appends extra last so lent credentials take effect after stripping.
func LiveEnv(t LiveTestingT, fakeHome string, extra ...string) []string {
	t.Helper()
	hostHome, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("testhome: resolve user home for LiveEnv: %v", err)
	}
	env, extraDirs, err := BuildLiveEnv(os.Environ(), fakeHome, hostHome, extra...)
	if err != nil {
		t.Fatalf("%v", err)
	}
	GuardLiveTempHome(t, fakeHome, extraDirs...)
	return env
}

// BuildLiveEnv is the pure builder behind LiveEnv, exposed for deterministic unit testing.
// It returns the isolated environment, any external config directories (such as a CLAUDE_CONFIG_DIR
// outside fakeHome) that must also be scanned by CheckTempHome, or an error if the child env points
// into hostHome or contains host refresh-token bytes.
func BuildLiveEnv(base []string, fakeHome, hostHome string, extra ...string) ([]string, []string, error) {
	env := stripAndBuildLiveEnv(base, fakeHome, extra...)
	g := NewHostRefreshGuardFrom(hostHome)
	extraDirs, err := g.CheckLiveEnv(env, fakeHome)
	if err != nil {
		return nil, nil, err
	}
	return env, extraDirs, nil
}

func stripAndBuildLiveEnv(base []string, fakeHome string, extra ...string) []string {
	var env []string
	var baseGoFlags string
	for _, e := range base {
		k, v, _ := strings.Cut(e, "=")
		switch {
		case k == "HOME", k == "USERPROFILE",
			k == "PLUGIN_ROOT", k == "PLUGIN_DATA",
			k == "GOCACHE", k == "GOMODCACHE",
			strings.HasPrefix(k, "XDG_"),
			strings.HasPrefix(k, "CLAUDE_"),
			strings.HasPrefix(k, "ANTHROPIC_"),
			strings.HasPrefix(k, "ANTIGRAVITY_"),
			strings.HasPrefix(k, "FEOV_"):
			continue
		case k == "GOFLAGS":
			baseGoFlags = v
			continue
		}
		env = append(env, e)
	}
	env = append(env,
		"HOME="+fakeHome,
		"USERPROFILE="+fakeHome,
		"AGY_CLI_DISABLE_AUTO_UPDATE=true",
		"DISABLE_AUTOUPDATER=1",
	)
	gc, gmc := hostGoCaches()
	if gc != "" {
		env = append(env, "GOCACHE="+gc)
	}
	if gmc != "" {
		env = append(env, "GOMODCACHE="+gmc)
	}
	goflags := baseGoFlags
	if goflags == "" {
		goflags = "-modcacherw"
	} else if !strings.Contains(goflags, "-modcacherw") {
		goflags += " -modcacherw"
	}
	env = append(env, "GOFLAGS="+goflags)
	env = append(env, extra...)
	return env
}

// LiveClaudeEnv builds a guarded isolated child environment anchored at fakeHome and authenticated
// ONLY via a lent CLAUDE_CODE_OAUTH_TOKEN access token read from the host login (never copying
// .credentials.json or its refreshToken, #1890).
func LiveClaudeEnv(t LiveTestingT, fakeHome string, extra ...string) ([]string, error) {
	t.Helper()
	tokEnv, err := ClaudeOAuthEnv()
	if err != nil {
		return nil, err
	}
	args := make([]string, 0, len(extra)+1)
	args = append(args, tokEnv)
	args = append(args, extra...)
	return LiveEnv(t, fakeHome, args...), nil
}

// ClaudeOAuthEnv returns "CLAUDE_CODE_OAUTH_TOKEN=<accessToken>" read from the host user's
// ~/.claude/.credentials.json, lending ONLY the short-lived access token and enforcing at least
// DefaultMinTokenLifetime remaining validity (#1890).
func ClaudeOAuthEnv() (string, error) {
	hostHome, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("testhome: resolve user home for Claude credentials: %w", err)
	}
	return ClaudeOAuthEnvFrom(hostHome, DefaultMinTokenLifetime)
}

// ClaudeOAuthEnvFrom is ClaudeOAuthEnv scoped to an explicit hostHome and minimum remaining
// lifetime (for tests).
func ClaudeOAuthEnvFrom(hostHome string, minLifetime time.Duration) (string, error) {
	credPath := filepath.Join(hostHome, ".claude", ".credentials.json")
	raw, err := os.ReadFile(credPath)
	if err != nil {
		return "", fmt.Errorf("testhome: read Claude credentials %s (sign in with `claude auth login`): %w", credPath, err)
	}
	var creds struct {
		OAuth struct {
			AccessToken string `json:"accessToken"`
			ExpiresAt   int64  `json:"expiresAt"`
		} `json:"claudeAiOauth"`
	}
	if err := json.Unmarshal(raw, &creds); err != nil {
		return "", fmt.Errorf("testhome: parse Claude credentials %s: %w", credPath, err)
	}
	if creds.OAuth.AccessToken == "" {
		return "", fmt.Errorf("testhome: %s has no claudeAiOauth.accessToken (sign in with `claude auth login`)", credPath)
	}
	if creds.OAuth.ExpiresAt <= 0 {
		return "", fmt.Errorf("testhome: %s missing or invalid claudeAiOauth.expiresAt (refusing unverified expiry)", credPath)
	}
	exp := time.UnixMilli(creds.OAuth.ExpiresAt)
	left := time.Until(exp)
	if NeedsRenewal(left, minLifetime) {
		return "", fmt.Errorf("testhome: Claude access token in %s expires in %s (need >= %s; sign in again with `claude auth login` or renew via `scripts/universe.sh`)",
			credPath, left.Round(time.Second), minLifetime)
	}
	return "CLAUDE_CODE_OAUTH_TOKEN=" + creds.OAuth.AccessToken, nil
}

// NeedsRenewal is the pure lifetime check shared by Claude and Antigravity token lending: returns
// true whenever the remaining lifetime is strictly below minLifetime.
func NeedsRenewal(left, minLifetime time.Duration) bool {
	return left < minLifetime
}

// StageAgyAccessToken stages a minimal, refresh-stripped Antigravity OAuth access token into
// fakeHome (plus non-secret antigravity_state.pbtxt and installation_id companion files, #1890).
//
// If the host access token has less than DefaultMinTokenLifetime remaining, it acquires an
// exclusive cross-process file lock in the user cache directory and renews through the host login
// once (mirroring scripts/universe.sh lend_agy_access_token) before staging.
func StageAgyAccessToken(fakeHome string) error {
	hostHome, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("testhome: resolve user home for Antigravity credentials: %w", err)
	}
	lockPath := agyRenewLockPath(hostHome)
	return StageAgyAccessTokenWithRenew(hostHome, fakeHome, DefaultMinTokenLifetime, lockPath, renewHostAgyToken)
}

// StageAgyAccessTokenFrom is StageAgyAccessToken scoped to an explicit hostHome and minimum
// remaining lifetime without attempting host renewal (for unit tests).
func StageAgyAccessTokenFrom(hostHome, fakeHome string, minLifetime time.Duration) error {
	return StageAgyAccessTokenWithRenew(hostHome, fakeHome, minLifetime, "", nil)
}

// StageAgyAccessTokenWithRenew is the injectable renewal entry point used by StageAgyAccessToken
// and unit tests (pinning both the renewal trigger and the post-lock re-check).
func StageAgyAccessTokenWithRenew(hostHome, fakeHome string, minLifetime time.Duration, lockPath string, renew func(hostHome string) error) error {
	err := stageAgyAccessTokenOnce(hostHome, fakeHome, minLifetime)
	if err == nil || renew == nil || !errors.Is(err, errAgyTokenShort) {
		return err
	}
	runLocked := func() error {
		// Re-check inside the cross-process lock in case another test process or universe.sh
		// already renewed the host token while we waited for the lock.
		if err2 := stageAgyAccessTokenOnce(hostHome, fakeHome, minLifetime); err2 == nil {
			return nil
		}
		if rerr := renew(hostHome); rerr != nil {
			return fmt.Errorf("testhome: host Antigravity access token needed renewal (%v), and automatic renewal failed (run `agy -p ok --model gemini-3.8-flash-low` in your real HOME): %w", err, rerr)
		}
		return stageAgyAccessTokenOnce(hostHome, fakeHome, minLifetime)
	}
	if lockPath == "" {
		return runLocked()
	}
	return withProcessLock(lockPath, runLocked)
}

func agyRenewLockPath(hostHome string) string {
	cacheDir, err := os.UserCacheDir()
	if err != nil || cacheDir == "" {
		cacheDir = filepath.Join(hostHome, ".cache")
	}
	return filepath.Join(cacheDir, "special-circumstances", "agy-renew.lock")
}

func withProcessLock(lockPath string, fn func() error) error {
	if err := os.MkdirAll(filepath.Dir(lockPath), 0o700); err != nil {
		return fmt.Errorf("testhome: mkdir lock dir: %w", err)
	}
	fl := flock.New(lockPath)
	if err := fl.Lock(); err != nil {
		return fmt.Errorf("testhome: flock %s: %w", lockPath, err)
	}
	defer func() { _ = fl.Unlock() }()
	return fn()
}

var errAgyTokenShort = errors.New("antigravity access token below minimum lifetime")

// renewHostAgyToken runs one short `agy -p` turn in hostHome to refresh the host's access token.
// Like scripts/universe.sh lend_agy_access_token, it intentionally preserves host XDG_RUNTIME_DIR
// and XDG_STATE_HOME (while stripping FEOV_*/CLAUDE_*/ANTIGRAVITY_*) so the host-home turn behaves
// identically to a normal host invocation and never disturbs the running remote-control daemon.
func renewHostAgyToken(hostHome string) error {
	agyBin, err := exec.LookPath("agy")
	if err != nil {
		return fmt.Errorf("agy not found on PATH: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, agyBin, "-p", "Reply with ok.", "--model", "gemini-3.8-flash-low", "--output-format", "text", "--print-timeout", "30s")
	cmd.Dir = hostHome
	extra := []string{
		"XDG_CONFIG_HOME=" + filepath.Join(hostHome, ".config"),
		"XDG_CACHE_HOME=" + filepath.Join(hostHome, ".cache"),
		"XDG_DATA_HOME=" + filepath.Join(hostHome, ".local", "share"),
	}
	if rt := os.Getenv("XDG_RUNTIME_DIR"); rt != "" {
		extra = append(extra, "XDG_RUNTIME_DIR="+rt)
	}
	if st := os.Getenv("XDG_STATE_HOME"); st != "" {
		extra = append(extra, "XDG_STATE_HOME="+st)
	}
	cmd.Env = stripAndBuildLiveEnv(os.Environ(), hostHome, extra...)
	if out, err := cmd.CombinedOutput(); err != nil {
		firstLine, _, _ := strings.Cut(strings.TrimSpace(string(out)), "\n")
		if len(firstLine) > 160 {
			firstLine = firstLine[:160] + "..."
		}
		return fmt.Errorf("host `agy -p` renewal failed (%v): %s", err, firstLine)
	}
	return nil
}

func stageAgyAccessTokenOnce(hostHome, fakeHome string, minLifetime time.Duration) error {
	srcDir := filepath.Join(hostHome, ".gemini", "antigravity-cli")
	srcTok := filepath.Join(srcDir, "antigravity-oauth-token")
	raw, err := os.ReadFile(srcTok)
	if err != nil {
		return fmt.Errorf("testhome: read Antigravity credentials %s (sign in with `agy`): %w", srcTok, err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		return fmt.Errorf("testhome: parse Antigravity credentials %s: %w", srcTok, err)
	}
	tokObj, ok := doc["token"].(map[string]any)
	if !ok {
		return fmt.Errorf("testhome: %s missing token object", srcTok)
	}
	accessTok, _ := tokObj["access_token"].(string)
	if accessTok == "" {
		return fmt.Errorf("testhome: %s missing token.access_token", srcTok)
	}
	expStr, _ := tokObj["expiry"].(string)
	if expStr == "" {
		return fmt.Errorf("testhome: %s missing token.expiry (refusing unverified expiry)", srcTok)
	}
	exp, perr := parseExpiry(expStr)
	if perr != nil {
		return fmt.Errorf("testhome: %s has unparseable token.expiry %q (refusing fail-closed): %w", srcTok, expStr, perr)
	}
	left := time.Until(exp)
	if NeedsRenewal(left, minLifetime) {
		return fmt.Errorf("%w: %s expires in %s (need >= %s; renew with `agy -p ok --model gemini-3.8-flash-low`)",
			errAgyTokenShort, srcTok, left.Round(time.Second), minLifetime)
	}

	// Build a minimal staged token carrying only what agy needs to authenticate with the access
	// token — never copying refresh_token, id_token_email, writer_id, or expires_in.
	minimalTok := map[string]any{
		"access_token": accessTok,
		"expiry":       expStr,
	}
	if tt, _ := tokObj["token_type"].(string); tt != "" {
		minimalTok["token_type"] = tt
	}
	minimalDoc := map[string]any{
		"token": minimalTok,
	}
	if am, ok := doc["auth_method"]; ok {
		minimalDoc["auth_method"] = am
	}

	stripped, err := json.Marshal(minimalDoc)
	if err != nil {
		return fmt.Errorf("testhome: marshal refresh-stripped Antigravity token: %w", err)
	}
	if bytes.Contains(stripped, []byte("refresh_token")) {
		return errors.New("testhome: refresh_token substring survived strip in Antigravity token JSON")
	}

	dstDir := filepath.Join(fakeHome, ".gemini", "antigravity-cli")
	if err := os.MkdirAll(dstDir, 0o700); err != nil {
		return fmt.Errorf("testhome: mkdir %s: %w", dstDir, err)
	}
	dstTok := filepath.Join(dstDir, "antigravity-oauth-token")
	if err := os.WriteFile(dstTok, stripped, 0o600); err != nil {
		return fmt.Errorf("testhome: write %s: %w", dstTok, err)
	}
	for _, extra := range []string{"antigravity_state.pbtxt", "installation_id"} {
		data, err := os.ReadFile(filepath.Join(srcDir, extra))
		if err != nil {
			continue
		}
		if err := os.WriteFile(filepath.Join(dstDir, extra), data, 0o600); err != nil {
			return fmt.Errorf("testhome: write %s: %w", extra, err)
		}
	}
	return nil
}

func parseExpiry(s string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
		return t, nil
	}
	return time.Parse(time.RFC3339, s)
}

// HostRefreshGuard enforces the #1890 invariant structurally: no refresh token ever reaches the
// child's world. It checks:
//  1. Every symlink in fakeHome (file or directory) resolves inside fakeHome and never into the
//     host's credential paths.
//  2. Every credential file (.credentials.json, antigravity-oauth-token, or JSON under .claude/
//     or .gemini/, including any external CLAUDE_CONFIG_DIR) contains no non-empty refreshToken
//     or refresh_token field regardless of value (catches copy-then-rotate).
//  3. No regular file in fakeHome (or child env value) contains the host's refresh-token bytes
//     (reporting only SHA-256 prefixes, never secret bytes).
//  4. Missing roots, unreadable files, oversized files, and walk errors fail closed as "not scanned".
type HostRefreshGuard struct {
	hostHome         string
	claudeRefresh    []byte
	claudeRefreshSum string
	agyRefresh       []byte
	agyRefreshSum    string
	MaxScanBytes     int64
}

// NewHostRefreshGuard creates a structural guard anchored at the current user's home directory.
func NewHostRefreshGuard() (*HostRefreshGuard, error) {
	hostHome, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("testhome: resolve user home for refresh guard: %w", err)
	}
	return NewHostRefreshGuardFrom(hostHome), nil
}

// NewHostRefreshGuardFrom creates a structural guard anchored at an explicit hostHome (for tests).
func NewHostRefreshGuardFrom(hostHome string) *HostRefreshGuard {
	g := &HostRefreshGuard{
		hostHome:     hostHome,
		MaxScanBytes: DefaultMaxScanBytes,
	}
	if tok := readHostClaudeRefreshToken(hostHome); len(tok) > 0 {
		g.claudeRefresh = tok
		g.claudeRefreshSum = sha256Prefix(tok)
	}
	if tok := readHostAgyRefreshToken(hostHome); len(tok) > 0 {
		g.agyRefresh = tok
		g.agyRefreshSum = sha256Prefix(tok)
	}
	return g
}

// CheckLiveEnv verifies that no environment entry carries host refresh-token bytes and that no
// home/config path variable points into hostHome (outside fakeHome and outside an isolated TMPDIR).
// It returns any external CLAUDE_CONFIG_DIR outside fakeHome so the caller can pass it to CheckTempHome.
func (g *HostRefreshGuard) CheckLiveEnv(env []string, fakeHome string) ([]string, error) {
	if g == nil {
		return nil, nil
	}
	canonHost := cleanOrCanon(g.hostHome)
	canonFake := cleanOrCanon(fakeHome)
	hostIsFake := canonHost != "" && canonHost == canonFake
	hostClaudeDir := cleanOrCanon(filepath.Join(g.hostHome, ".claude"))
	hostAgyDir := cleanOrCanon(filepath.Join(g.hostHome, ".gemini", "antigravity-cli"))
	canonTmp := cleanOrCanon(os.TempDir())
	tmpSafe := canonTmp != "" && canonTmp != canonHost && canonTmp != hostClaudeDir && canonTmp != hostAgyDir &&
		!isInside(hostClaudeDir, canonTmp) && !isInside(hostAgyDir, canonTmp)

	var extraDirs []string
	for _, entry := range env {
		k, v, _ := strings.Cut(entry, "=")
		if len(g.claudeRefresh) > 0 && bytes.Contains([]byte(v), g.claudeRefresh) {
			return nil, fmt.Errorf("testhome(#1890): child env %s contains host Claude refresh token (sha256:%s)", k, g.claudeRefreshSum)
		}
		if len(g.agyRefresh) > 0 && bytes.Contains([]byte(v), g.agyRefresh) {
			return nil, fmt.Errorf("testhome(#1890): child env %s contains host Antigravity refresh token (sha256:%s)", k, g.agyRefreshSum)
		}
		switch k {
		case "HOME", "USERPROFILE", "CLAUDE_CONFIG_DIR", "XDG_CONFIG_HOME", "XDG_DATA_HOME", "XDG_STATE_HOME", "XDG_CACHE_HOME":
			if v == "" {
				continue
			}
			canonV := cleanOrCanon(v)
			inFake := canonFake != "" && isInside(canonV, canonFake)
			inTmp := tmpSafe && isInside(canonV, canonTmp)
			if !hostIsFake && !inFake && !inTmp && canonHost != "" && pathOverlaps(canonV, canonHost) {
				return nil, fmt.Errorf("testhome(#1890): child env %s=%q points into host home %q", k, v, g.hostHome)
			}
			if k == "CLAUDE_CONFIG_DIR" && !inFake {
				extraDirs = append(extraDirs, v)
			}
		}
	}
	return extraDirs, nil
}

// CheckTempHome structurally verifies fakeHome (and any extraDirs such as an external
// CLAUDE_CONFIG_DIR) contains no symlinks escaping fakeHome, no non-empty refreshToken or
// refresh_token fields in credential JSON files, and no bytes matching the host's refresh tokens.
// Missing roots, walk errors, unreadable files, and oversized files fail closed as "not scanned".
func (g *HostRefreshGuard) CheckTempHome(fakeHome string, extraDirs ...string) error {
	if g == nil {
		return errors.New("testhome(#1890): not scanned: nil guard")
	}
	maxBytes := g.MaxScanBytes
	if maxBytes <= 0 {
		maxBytes = DefaultMaxScanBytes
	}
	roots := make([]string, 0, 1+len(extraDirs))
	roots = append(roots, fakeHome)
	for _, d := range extraDirs {
		if d != "" {
			roots = append(roots, d)
		}
	}
	for i, root := range roots {
		// External CLAUDE_CONFIG_DIR may legitimately not be created yet if the test only ran agy;
		// the primary fakeHome root MUST exist (missing fakeHome is a "not scanned" failure).
		info, err := os.Lstat(root)
		if err != nil {
			if i > 0 && errors.Is(err, fs.ErrNotExist) {
				continue
			}
			return fmt.Errorf("testhome(#1890): not scanned: root %s: %w", root, err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("testhome(#1890): not scanned: root %s is a symlink", root)
		}
		if !info.IsDir() {
			return fmt.Errorf("testhome(#1890): not scanned: root %s is not a directory", root)
		}
		canonRoot, err := filepath.EvalSymlinks(root)
		if err != nil {
			return fmt.Errorf("testhome(#1890): not scanned: resolve root %s: %w", root, err)
		}
		if err := g.scanRoot(root, canonRoot, maxBytes); err != nil {
			return err
		}
	}
	return nil
}

func (g *HostRefreshGuard) scanRoot(root, canonRoot string, maxBytes int64) error {
	if g.hostHome != "" {
		canonHost := cleanOrCanon(g.hostHome)
		hostCredDirs := []string{
			cleanOrCanon(filepath.Join(g.hostHome, ".claude")),
			cleanOrCanon(filepath.Join(g.hostHome, ".gemini", "antigravity-cli")),
		}
		if canonRoot == canonHost {
			return fmt.Errorf("testhome(#1890): temp home root %s resolves to host home %s", root, g.hostHome)
		}
		for _, hdir := range hostCredDirs {
			if hdir != "" && isInside(hdir, canonRoot) {
				return fmt.Errorf("testhome(#1890): temp home root %s contains host credential path %s", root, hdir)
			}
		}
	}
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("testhome(#1890): not scanned: walk %s: %w", path, walkErr)
		}
		info, lerr := os.Lstat(path)
		if lerr != nil {
			return fmt.Errorf("testhome(#1890): not scanned: lstat %s: %w", path, lerr)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			target, rerr := filepath.EvalSymlinks(path)
			if rerr != nil {
				link, rlerr := os.Readlink(path)
				if rlerr != nil {
					return fmt.Errorf("testhome(#1890): not scanned: readlink %s: %w", path, rlerr)
				}
				if !filepath.IsAbs(link) {
					link = filepath.Join(filepath.Dir(path), link)
				}
				target = filepath.Clean(link)
			}
			if !isInside(target, canonRoot) {
				return fmt.Errorf("testhome(#1890): symlink %s resolves outside temp home (%s -> %s)", path, path, target)
			}
			return nil
		}
		if info.IsDir() || !info.Mode().IsRegular() {
			return nil
		}
		if info.Size() > maxBytes {
			return fmt.Errorf("testhome(#1890): not scanned: file %s is %d bytes (> %d limit)", path, info.Size(), maxBytes)
		}
		data, rerr := os.ReadFile(path)
		if rerr != nil {
			return fmt.Errorf("testhome(#1890): not scanned: read %s: %w", path, rerr)
		}
		if isCredentialJSONPath(path) {
			if field := findNonEmptyRefreshField(data); field != "" {
				return fmt.Errorf("testhome(#1890): temp home credential file %s contains non-empty %s", path, field)
			}
		}
		if len(g.claudeRefresh) > 0 && bytes.Contains(data, g.claudeRefresh) {
			return fmt.Errorf("testhome(#1890): temp home file %s contains host Claude refresh token (sha256:%s)", path, g.claudeRefreshSum)
		}
		if len(g.agyRefresh) > 0 && bytes.Contains(data, g.agyRefresh) {
			return fmt.Errorf("testhome(#1890): temp home file %s contains host Antigravity refresh token (sha256:%s)", path, g.agyRefreshSum)
		}
		return nil
	})
}

func isCredentialJSONPath(path string) bool {
	base := filepath.Base(path)
	if base == ".credentials.json" || base == "antigravity-oauth-token" {
		return true
	}
	if strings.HasSuffix(base, ".json") {
		slashed := filepath.ToSlash(path)
		if strings.Contains(slashed, "/.claude/") || strings.Contains(slashed, "/.gemini/") {
			return true
		}
	}
	return false
}

// findNonEmptyRefreshField parses JSON and returns "refreshToken" or "refresh_token" if any object
// at any depth carries a non-empty string value under either key.
func findNonEmptyRefreshField(raw []byte) string {
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return ""
	}
	return walkRefreshField(v)
}

func walkRefreshField(v any) string {
	switch x := v.(type) {
	case map[string]any:
		for k, val := range x {
			if k == "refreshToken" || k == "refresh_token" {
				if s, ok := val.(string); ok && strings.TrimSpace(s) != "" {
					return k
				}
			}
			if found := walkRefreshField(val); found != "" {
				return found
			}
		}
	case []any:
		for _, item := range x {
			if found := walkRefreshField(item); found != "" {
				return found
			}
		}
	}
	return ""
}

// GuardLiveTempHome registers a t.Cleanup hook verifying structurally that fakeHome (and any
// extraDirs) never holds a symlink escaping fakeHome, any non-empty refreshToken/refresh_token
// field, or any host refresh-token bytes.
func GuardLiveTempHome(t LiveTestingT, fakeHome string, extraDirs ...string) {
	t.Helper()
	g, err := NewHostRefreshGuard()
	if err != nil {
		t.Errorf("testhome: %v", err)
		return
	}
	guardWith(t, g, fakeHome, extraDirs...)
}

// GuardLiveTempHomeWith is GuardLiveTempHome with an explicit HostRefreshGuard (for unit tests).
func GuardLiveTempHomeWith(t LiveTestingT, g *HostRefreshGuard, fakeHome string, extraDirs ...string) {
	t.Helper()
	guardWith(t, g, fakeHome, extraDirs...)
}

func guardWith(t LiveTestingT, g *HostRefreshGuard, fakeHome string, extraDirs ...string) {
	dirs := append([]string(nil), extraDirs...)
	t.Cleanup(func() {
		if err := g.CheckTempHome(fakeHome, dirs...); err != nil {
			t.Errorf("%v", err)
		}
	})
}

func cleanOrCanon(p string) string {
	if p == "" {
		return ""
	}
	if canon, err := filepath.EvalSymlinks(p); err == nil {
		return canon
	}
	if abs, err := filepath.Abs(p); err == nil {
		return abs
	}
	return filepath.Clean(p)
}

func isInside(path, root string) bool {
	if path == "" || root == "" {
		return false
	}
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	return rel == "." || filepath.IsLocal(rel)
}

func pathOverlaps(a, b string) bool {
	return isInside(a, b) || isInside(b, a)
}

func readHostClaudeRefreshToken(hostHome string) []byte {
	raw, err := os.ReadFile(filepath.Join(hostHome, ".claude", ".credentials.json"))
	if err != nil {
		return nil
	}
	var creds struct {
		OAuth struct {
			RefreshToken string `json:"refreshToken"`
		} `json:"claudeAiOauth"`
	}
	if json.Unmarshal(raw, &creds) != nil || creds.OAuth.RefreshToken == "" {
		return nil
	}
	return []byte(creds.OAuth.RefreshToken)
}

func readHostAgyRefreshToken(hostHome string) []byte {
	raw, err := os.ReadFile(filepath.Join(hostHome, ".gemini", "antigravity-cli", "antigravity-oauth-token"))
	if err != nil {
		return nil
	}
	var doc struct {
		Token struct {
			RefreshToken string `json:"refresh_token"`
		} `json:"token"`
	}
	if json.Unmarshal(raw, &doc) != nil || doc.Token.RefreshToken == "" {
		return nil
	}
	return []byte(doc.Token.RefreshToken)
}

func sha256Prefix(b []byte) string {
	if len(b) == 0 {
		return "missing"
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:6])
}
