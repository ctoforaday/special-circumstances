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
)

// DefaultMinTokenLifetime is the minimum remaining validity required before lending a host OAuth
// access token into an isolated test HOME (matches scripts/universe.sh AGY_TOKEN_MIN_MINUTES=15,
// exceeding any single HOOKCORE_LIVE test timeout so a lent token never expires mid-run).
const DefaultMinTokenLifetime = 15 * time.Minute

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

// LiveEnv builds an isolated child process environment anchored at fakeHome.
//
// It strips HOME, USERPROFILE, XDG_*, CLAUDE_*, ANTHROPIC_*, ANTIGRAVITY_*, FEOV_*, PLUGIN_ROOT,
// PLUGIN_DATA, GOCACHE, GOMODCACHE, and GOFLAGS so neither harness credentials nor parent session
// attribution leak into the child, disables both clients' auto-updaters, pins host Go caches with
// GOFLAGS=-modcacherw so temp-home builds do not create read-only module caches, and appends extra
// last so lent credentials (such as CLAUDE_CODE_OAUTH_TOKEN) take effect after stripping.
func LiveEnv(fakeHome string, extra ...string) []string {
	return liveEnvFrom(os.Environ(), fakeHome, extra...)
}

func liveEnvFrom(base []string, fakeHome string, extra ...string) []string {
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

// LiveClaudeEnv builds an isolated child environment anchored at fakeHome and authenticated ONLY
// via a lent CLAUDE_CODE_OAUTH_TOKEN access token read from the host login (never copying
// .credentials.json or its refreshToken, #1890).
func LiveClaudeEnv(fakeHome string, extra ...string) ([]string, error) {
	tokEnv, err := ClaudeOAuthEnv()
	if err != nil {
		return nil, err
	}
	args := make([]string, 0, len(extra)+1)
	args = append(args, tokEnv)
	args = append(args, extra...)
	return LiveEnv(fakeHome, args...), nil
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
	if left < minLifetime {
		return "", fmt.Errorf("testhome: Claude access token in %s expires in %s (need >= %s; renew with a host turn `claude -p ok --model haiku --max-turns 1` or `claude auth login`)",
			credPath, left.Round(time.Second), minLifetime)
	}
	return "CLAUDE_CODE_OAUTH_TOKEN=" + creds.OAuth.AccessToken, nil
}

// agyHostRenewMu serializes automatic host Antigravity token renewals across parallel tests.
var agyHostRenewMu sync.Mutex

// StageAgyAccessToken stages a minimal, refresh-stripped Antigravity OAuth access token into
// fakeHome (plus non-secret antigravity_state.pbtxt and installation_id companion files, #1890).
//
// If the host access token has less than DefaultMinTokenLifetime remaining, it renews through the
// host login once (mirroring scripts/universe.sh lend_agy_access_token) before staging.
func StageAgyAccessToken(fakeHome string) error {
	hostHome, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("testhome: resolve user home for Antigravity credentials: %w", err)
	}
	return stageAgyAccessTokenWithRenew(hostHome, fakeHome, DefaultMinTokenLifetime, renewHostAgyToken)
}

// StageAgyAccessTokenFrom is StageAgyAccessToken scoped to an explicit hostHome and minimum
// remaining lifetime without attempting host renewal (for unit tests).
func StageAgyAccessTokenFrom(hostHome, fakeHome string, minLifetime time.Duration) error {
	return stageAgyAccessTokenWithRenew(hostHome, fakeHome, minLifetime, nil)
}

func stageAgyAccessTokenWithRenew(hostHome, fakeHome string, minLifetime time.Duration, renew func(hostHome string) error) error {
	err := stageAgyAccessTokenOnce(hostHome, fakeHome, minLifetime)
	if err == nil || renew == nil || !errors.Is(err, errAgyTokenShort) {
		return err
	}
	agyHostRenewMu.Lock()
	defer agyHostRenewMu.Unlock()
	// Re-check inside the lock in case another test already renewed the host token.
	if err2 := stageAgyAccessTokenOnce(hostHome, fakeHome, minLifetime); err2 == nil {
		return nil
	}
	if rerr := renew(hostHome); rerr != nil {
		return fmt.Errorf("testhome: host Antigravity access token needed renewal (%v), and automatic renewal failed (run `agy -p ok --model gemini-3.8-flash-low` in your real HOME): %w", err, rerr)
	}
	return stageAgyAccessTokenOnce(hostHome, fakeHome, minLifetime)
}

var errAgyTokenShort = errors.New("antigravity access token below minimum lifetime")

func renewHostAgyToken(hostHome string) error {
	agyBin, err := exec.LookPath("agy")
	if err != nil {
		return fmt.Errorf("agy not found on PATH: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, agyBin, "-p", "Reply with ok.", "--model", "gemini-3.8-flash-low", "--output-format", "text", "--print-timeout", "30s")
	cmd.Dir = hostHome
	cmd.Env = LiveEnv(hostHome,
		"XDG_CONFIG_HOME="+filepath.Join(hostHome, ".config"),
		"XDG_CACHE_HOME="+filepath.Join(hostHome, ".cache"),
		"XDG_DATA_HOME="+filepath.Join(hostHome, ".local", "share"),
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("host `agy -p` renewal failed: %w\n%s", err, strings.TrimSpace(string(out)))
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
	if left < minLifetime {
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

// HostRefreshGuard enforces the #1890 invariant against the OUTCOME rather than source spelling:
// it fingerprints the host Claude refreshToken and Antigravity refresh_token (holding only hashes
// for messages, never printing secret bytes), verifies no file or symlink under any temp HOME
// contains either refresh token's bytes, and verifies neither host refresh token changed during
// the test.
type HostRefreshGuard struct {
	hostHome         string
	claudeRefresh    []byte
	claudeRefreshSum string
	agyRefresh       []byte
	agyRefreshSum    string
}

// NewHostRefreshGuard snapshots the current user's host refresh tokens before a live test runs.
func NewHostRefreshGuard() (*HostRefreshGuard, error) {
	hostHome, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("testhome: resolve user home for refresh guard: %w", err)
	}
	return NewHostRefreshGuardFrom(hostHome), nil
}

// NewHostRefreshGuardFrom snapshots refresh tokens under an explicit hostHome (for tests).
func NewHostRefreshGuardFrom(hostHome string) *HostRefreshGuard {
	g := &HostRefreshGuard{hostHome: hostHome}
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

// CheckTempHome scans every file and symlink target under tempHome and fails if any holds the host
// Claude or Antigravity refresh token bytes. It catches raw copies, os.Symlink, cp -r, and exec cp
// regardless of how the test constructed tempHome.
func (g *HostRefreshGuard) CheckTempHome(tempHome string) error {
	if g == nil || (len(g.claudeRefresh) == 0 && len(g.agyRefresh) == 0) {
		return nil
	}
	var foundErr error
	_ = filepath.WalkDir(tempHome, func(path string, d fs.DirEntry, err error) error {
		if err != nil || foundErr != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		// Follow symlinks via os.Stat + os.ReadFile so an os.Symlink to ~/.claude/.credentials.json
		// or ~/.gemini/antigravity-cli/antigravity-oauth-token is caught immediately.
		info, serr := os.Stat(path)
		if serr != nil || info.IsDir() || info.Size() > 8<<20 {
			return nil
		}
		data, rerr := os.ReadFile(path)
		if rerr != nil {
			return nil
		}
		if len(g.claudeRefresh) > 0 && bytes.Contains(data, g.claudeRefresh) {
			foundErr = fmt.Errorf("testhome(#1890): temp home file %s contains host Claude refresh token (sha256:%s)", path, g.claudeRefreshSum)
			return filepath.SkipAll
		}
		if len(g.agyRefresh) > 0 && bytes.Contains(data, g.agyRefresh) {
			foundErr = fmt.Errorf("testhome(#1890): temp home file %s contains host Antigravity refresh token (sha256:%s)", path, g.agyRefreshSum)
			return filepath.SkipAll
		}
		return nil
	})
	return foundErr
}

// VerifyHostUnchanged confirms that neither the host Claude refreshToken nor the host Antigravity
// refresh_token changed since the guard was created.
func (g *HostRefreshGuard) VerifyHostUnchanged() error {
	if g == nil {
		return nil
	}
	if g.claudeRefreshSum != "" {
		nowSum := sha256Prefix(readHostClaudeRefreshToken(g.hostHome))
		if nowSum != g.claudeRefreshSum {
			return fmt.Errorf("testhome(#1890): host Claude refresh token changed during test (before sha256:%s, after sha256:%s)", g.claudeRefreshSum, nowSum)
		}
	}
	if g.agyRefreshSum != "" {
		nowSum := sha256Prefix(readHostAgyRefreshToken(g.hostHome))
		if nowSum != g.agyRefreshSum {
			return fmt.Errorf("testhome(#1890): host Antigravity refresh token changed during test (before sha256:%s, after sha256:%s)", g.agyRefreshSum, nowSum)
		}
	}
	return nil
}

// liveTestingT is the subset of *testing.T needed by GuardLiveTempHome so package testhome does
// not import testing.
type liveTestingT interface {
	Helper()
	Cleanup(func())
	Errorf(format string, args ...any)
}

// GuardLiveTempHome snapshots the host's refresh tokens now and registers a t.Cleanup hook that
// verifies (1) no file or symlink under fakeHome contains either host refresh token's bytes, and
// (2) neither host refresh token changed while the test ran.
func GuardLiveTempHome(t liveTestingT, fakeHome string) {
	t.Helper()
	g, err := NewHostRefreshGuard()
	if err != nil {
		t.Errorf("testhome: %v", err)
		return
	}
	t.Cleanup(func() {
		if err := g.CheckTempHome(fakeHome); err != nil {
			t.Errorf("%v", err)
		}
		if err := g.VerifyHostUnchanged(); err != nil {
			t.Errorf("%v", err)
		}
	})
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
