package testhome_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/shared/testhome"
)

// fakeLiveT implements testhome.LiveTestingT so unit tests can drive GuardLiveTempHome,
// NewLiveHome, and LiveEnv directly and assert cleanup errors (closing X6).
type fakeLiveT struct {
	t        *testing.T
	cleanups []func()
	errors   []string
	fatals   []string
}

func (f *fakeLiveT) Helper()                   {}
func (f *fakeLiveT) TempDir() string           { return f.t.TempDir() }
func (f *fakeLiveT) Cleanup(fn func())         { f.cleanups = append(f.cleanups, fn) }
func (f *fakeLiveT) Errorf(s string, a ...any) { f.errors = append(f.errors, fmt.Sprintf(s, a...)) }
func (f *fakeLiveT) Fatalf(s string, a ...any) { f.fatals = append(f.fatals, fmt.Sprintf(s, a...)) }

func (f *fakeLiveT) runCleanups() {
	for i := len(f.cleanups) - 1; i >= 0; i-- {
		f.cleanups[i]()
	}
}

func TestClaudeOAuthEnvFrom(t *testing.T) {
	host := t.TempDir()
	claudeDir := filepath.Join(host, ".claude")
	if err := os.MkdirAll(claudeDir, 0o700); err != nil {
		t.Fatal(err)
	}
	validExp := time.Now().Add(60 * time.Minute).UnixMilli()
	raw := fmt.Sprintf(`{"claudeAiOauth":{"accessToken":"sk-ant-oat01-test-access","refreshToken":"sk-ant-ort01-secret-refresh","expiresAt":%d}}`, validExp)
	if err := os.WriteFile(filepath.Join(claudeDir, ".credentials.json"), []byte(raw), 0o600); err != nil {
		t.Fatal(err)
	}

	kv, err := testhome.ClaudeOAuthEnvFrom(host, 15*time.Minute)
	if err != nil {
		t.Fatalf("ClaudeOAuthEnvFrom: %v", err)
	}
	if kv != "CLAUDE_CODE_OAUTH_TOKEN=sk-ant-oat01-test-access" {
		t.Fatalf("got %q, want access token env var only", kv)
	}
	if strings.Contains(kv, "refresh") {
		t.Fatalf("ClaudeOAuthEnvFrom leaked refresh token: %q", kv)
	}

	if _, err := testhome.ClaudeOAuthEnvFrom(t.TempDir(), 15*time.Minute); err == nil {
		t.Fatal("expected error for missing .credentials.json")
	}

	emptyHost := t.TempDir()
	_ = os.MkdirAll(filepath.Join(emptyHost, ".claude"), 0o700)
	_ = os.WriteFile(filepath.Join(emptyHost, ".claude", ".credentials.json"), []byte(`{"claudeAiOauth":{"refreshToken":"only-refresh","expiresAt":9999999999999}}`), 0o600)
	if _, err := testhome.ClaudeOAuthEnvFrom(emptyHost, 15*time.Minute); err == nil {
		t.Fatal("expected error when accessToken is absent")
	}

	noExpHost := t.TempDir()
	_ = os.MkdirAll(filepath.Join(noExpHost, ".claude"), 0o700)
	_ = os.WriteFile(filepath.Join(noExpHost, ".claude", ".credentials.json"), []byte(`{"claudeAiOauth":{"accessToken":"sk-ant-oat01-noexp"}}`), 0o600)
	if _, err := testhome.ClaudeOAuthEnvFrom(noExpHost, 15*time.Minute); err == nil {
		t.Fatal("expected error when expiresAt is absent")
	}

	shortHost := t.TempDir()
	_ = os.MkdirAll(filepath.Join(shortHost, ".claude"), 0o700)
	shortExp := time.Now().Add(30 * time.Second).UnixMilli()
	_ = os.WriteFile(filepath.Join(shortHost, ".claude", ".credentials.json"),
		[]byte(fmt.Sprintf(`{"claudeAiOauth":{"accessToken":"sk-ant-oat01-short","expiresAt":%d}}`, shortExp)), 0o600)
	if _, err := testhome.ClaudeOAuthEnvFrom(shortHost, 15*time.Minute); err == nil || !strings.Contains(err.Error(), "claude auth login") {
		t.Fatalf("expected short-lifetime error naming claude auth login, got %v", err)
	}
	_ = os.WriteFile(filepath.Join(shortHost, ".claude", ".credentials.json"),
		[]byte(`{"claudeAiOauth":{"accessToken":"sk-ant-oat01-1970","expiresAt":1000}}`), 0o600)
	if _, err := testhome.ClaudeOAuthEnvFrom(shortHost, 15*time.Minute); err == nil {
		t.Fatal("expected error for 1970 expiresAt")
	}
}

func TestStageAgyAccessTokenFromAndRenewal(t *testing.T) {
	host := t.TempDir()
	srcDir := filepath.Join(host, ".gemini", "antigravity-cli")
	if err := os.MkdirAll(srcDir, 0o700); err != nil {
		t.Fatal(err)
	}
	validExp := time.Now().Add(45 * time.Minute).UTC().Format(time.RFC3339Nano)
	raw := fmt.Sprintf(`{
  "auth_method": "oauth",
  "id_token_email": "user@example.com",
  "writer_id": "writer-99",
  "refresh_token": "top-level-refresh",
  "token": {
    "access_token": "ya29.test-access",
    "token_type": "Bearer",
    "expires_in": 3599,
    "refresh_token": "1//secret-refresh-token",
    "expiry": %q
  }
}`, validExp)
	if err := os.WriteFile(filepath.Join(srcDir, "antigravity-oauth-token"), []byte(raw), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "antigravity_state.pbtxt"), []byte("state: 1\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "installation_id"), []byte("inst-123"), 0o600); err != nil {
		t.Fatal(err)
	}

	fake := t.TempDir()
	if err := testhome.StageAgyAccessTokenFrom(host, fake, 15*time.Minute); err != nil {
		t.Fatalf("StageAgyAccessTokenFrom: %v", err)
	}

	dstTok := filepath.Join(fake, ".gemini", "antigravity-cli", "antigravity-oauth-token")
	info, err := os.Stat(dstTok)
	if err != nil {
		t.Fatalf("stat staged token: %v", err)
	}
	if runtime.GOOS != "windows" && info.Mode().Perm() != 0o600 {
		t.Errorf("staged token mode = %o, want 0600", info.Mode().Perm())
	}
	data, err := os.ReadFile(dstTok)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"refresh_token", "secret-refresh-token", "id_token_email", "writer_id", "expires_in"} {
		if strings.Contains(string(data), forbidden) {
			t.Fatalf("staged token retained %q: %s", forbidden, data)
		}
	}
	var parsed struct {
		AuthMethod string `json:"auth_method"`
		Token      struct {
			AccessToken string `json:"access_token"`
			TokenType   string `json:"token_type"`
			Expiry      string `json:"expiry"`
		} `json:"token"`
	}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("parse staged token: %v", err)
	}
	if parsed.AuthMethod != "oauth" || parsed.Token.AccessToken != "ya29.test-access" || parsed.Token.TokenType != "Bearer" || parsed.Token.Expiry != validExp {
		t.Fatalf("unexpected staged token payload: %+v", parsed)
	}
	for _, extra := range []string{"antigravity_state.pbtxt", "installation_id"} {
		if _, err := os.Stat(filepath.Join(fake, ".gemini", "antigravity-cli", extra)); err != nil {
			t.Errorf("expected companion file %s to be staged: %v", extra, err)
		}
	}

	writeTok := func(body string) string {
		h := t.TempDir()
		d := filepath.Join(h, ".gemini", "antigravity-cli")
		_ = os.MkdirAll(d, 0o700)
		_ = os.WriteFile(filepath.Join(d, "antigravity-oauth-token"), []byte(body), 0o600)
		return h
	}
	if err := testhome.StageAgyAccessTokenFrom(t.TempDir(), t.TempDir(), 15*time.Minute); err == nil {
		t.Fatal("expected error for missing antigravity-oauth-token")
	}
	if err := testhome.StageAgyAccessTokenFrom(writeTok(`{"token":{"refresh_token":"only"}}`), t.TempDir(), 15*time.Minute); err == nil {
		t.Fatal("expected error when token.access_token is absent")
	}
	if err := testhome.StageAgyAccessTokenFrom(writeTok(`{"token":{"access_token":"ya29.noexp"}}`), t.TempDir(), 15*time.Minute); err == nil {
		t.Fatal("expected fail-closed error when token.expiry is absent")
	}
	if err := testhome.StageAgyAccessTokenFrom(writeTok(`{"token":{"access_token":"ya29.badexp","expiry":"2020-01-01 00:00:00"}}`), t.TempDir(), 15*time.Minute); err == nil || !strings.Contains(err.Error(), "unparseable") {
		t.Fatalf("expected fail-closed error on unparseable expiry, got %v", err)
	}
	shortExp := time.Now().Add(30 * time.Second).UTC().Format(time.RFC3339)
	if err := testhome.StageAgyAccessTokenFrom(writeTok(fmt.Sprintf(`{"token":{"access_token":"ya29.short","expiry":%q}}`, shortExp)), t.TempDir(), 15*time.Minute); err == nil || !strings.Contains(err.Error(), "expires in") {
		t.Fatalf("expected below-floor lifetime error for 30s token, got %v", err)
	}

	// Pure renewal decision + boundary table + StageAgyAccessTokenWithRenew (closes X4, X5, M7, M8).
	if testhome.DefaultAgyMinTokenLifetime != testhome.AgyRefreshThreshold+testhome.AgyTestRuntimeBudget {
		t.Fatalf("DefaultAgyMinTokenLifetime (%s) != AgyRefreshThreshold (%s) + AgyTestRuntimeBudget (%s)",
			testhome.DefaultAgyMinTokenLifetime, testhome.AgyRefreshThreshold, testhome.AgyTestRuntimeBudget)
	}
	if testhome.DefaultClaudeMinTokenLifetime != 15*time.Minute {
		t.Fatalf("DefaultClaudeMinTokenLifetime = %s, want 15m", testhome.DefaultClaudeMinTokenLifetime)
	}
	if !testhome.NeedsRenewal(testhome.DefaultAgyMinTokenLifetime-time.Second, testhome.DefaultAgyMinTokenLifetime) ||
		testhome.NeedsRenewal(testhome.DefaultAgyMinTokenLifetime, testhome.DefaultAgyMinTokenLifetime) {
		t.Fatal("NeedsRenewal boundary check failed")
	}

	t.Run("PlanAgyTokenRenewal decision boundaries", func(t *testing.T) {
		minLife := testhome.DefaultAgyMinTokenLifetime // 15m (5m cliff + 10m budget)
		thresh := testhome.AgyRefreshThreshold         // 5m
		maxWait := testhome.MaxAgyRefreshWait          // 90s
		cases := []struct {
			name       string
			left       time.Duration
			wantAction testhome.AgyRenewalAction
			wantWait   time.Duration
		}{
			{"15m+1s stages immediately", 15*time.Minute + time.Second, testhome.AgyTokenStage, 0},
			{"exact 15m stages immediately", 15 * time.Minute, testhome.AgyTokenStage, 0},
			{"15m-1s fails fast with wait duration", 15*time.Minute - time.Second, testhome.AgyTokenFail, 10*time.Minute + 4*time.Second},
			{"6m26s exceeds 90s maxWait and fails fast", 6*time.Minute + 26*time.Second, testhome.AgyTokenFail, 91 * time.Second},
			{"6m25s hits exact 90s maxWait and waits then renews", 6*time.Minute + 25*time.Second, testhome.AgyTokenWaitThenRenew, 90 * time.Second},
			{"5m+1s waits then renews (6s)", 5*time.Minute + time.Second, testhome.AgyTokenWaitThenRenew, 6 * time.Second},
			{"exact 5m waits margin then renews (5s)", 5 * time.Minute, testhome.AgyTokenWaitThenRenew, 5 * time.Second},
			{"5m-1s renews immediately", 5*time.Minute - time.Second, testhome.AgyTokenRenewNow, 0},
			{"expired (-30s) renews immediately", -30 * time.Second, testhome.AgyTokenRenewNow, 0},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				gotAction, gotWait := testhome.PlanAgyTokenRenewal(tc.left, minLife, thresh, maxWait)
				if gotAction != tc.wantAction || gotWait != tc.wantWait {
					t.Fatalf("PlanAgyTokenRenewal(%s, %s) = (%s, %s), want (%s, %s)",
						tc.left, minLife, gotAction, gotWait, tc.wantAction, tc.wantWait)
				}
			})
		}
	})
	t.Run("auto-renews short token under process lock and succeeds", func(t *testing.T) {
		short := time.Now().Add(2 * time.Minute).UTC().Format(time.RFC3339)
		h := writeTok(fmt.Sprintf(`{"token":{"access_token":"ya29.old","expiry":%q}}`, short))
		lockPath := filepath.Join(t.TempDir(), "renew.lock")
		calls := 0
		renewFn := func(hh string) error {
			calls++
			fresh := time.Now().Add(50 * time.Minute).UTC().Format(time.RFC3339)
			return os.WriteFile(filepath.Join(hh, ".gemini", "antigravity-cli", "antigravity-oauth-token"),
				[]byte(fmt.Sprintf(`{"token":{"access_token":"ya29.fresh","refresh_token":"1//r","expiry":%q}}`, fresh)), 0o600)
		}
		dst := t.TempDir()
		if err := testhome.StageAgyAccessTokenWithRenew(h, dst, testhome.DefaultAgyMinTokenLifetime, lockPath, renewFn); err != nil {
			t.Fatalf("StageAgyAccessTokenWithRenew: %v", err)
		}
		if calls != 1 {
			t.Fatalf("renew called %d times, want 1", calls)
		}
		// Second call with fresh token must NOT renew again.
		if err := testhome.StageAgyAccessTokenWithRenew(h, t.TempDir(), testhome.DefaultAgyMinTokenLifetime, lockPath, renewFn); err != nil {
			t.Fatalf("second StageAgyAccessTokenWithRenew: %v", err)
		}
		if calls != 1 {
			t.Fatalf("renew called %d times after token became fresh, want 1", calls)
		}
	})

	t.Run("fails fast with ErrAgyTokenWaitTooLong without calling renew in dead-zone band (M8)", func(t *testing.T) {
		midBand := time.Now().Add(10 * time.Minute).UTC().Format(time.RFC3339)
		h := writeTok(fmt.Sprintf(`{"token":{"access_token":"ya29.mid","expiry":%q}}`, midBand))
		lockPath := filepath.Join(t.TempDir(), "mid.lock")
		calls := 0
		renewFn := func(string) error {
			calls++
			return nil
		}
		err := testhome.StageAgyAccessTokenWithRenew(h, t.TempDir(), testhome.DefaultAgyMinTokenLifetime, lockPath, renewFn)
		if err == nil {
			t.Fatal("expected fail-fast error for 10m token below 15m floor")
		}
		if !errors.Is(err, testhome.ErrAgyTokenWaitTooLong) {
			t.Fatalf("expected ErrAgyTokenWaitTooLong, got %v", err)
		}
		if errors.Is(err, testhome.ErrAgyTokenNotRefreshed) {
			t.Fatalf("fail-fast branch must not wrap ErrAgyTokenNotRefreshed when agy never ran: %v", err)
		}
		if calls != 0 {
			t.Fatalf("renew called %d times on fail-fast path, want 0", calls)
		}
	})

	t.Run("fails with named ErrAgyTokenNotRefreshed without looping when renewal does not extend expiry", func(t *testing.T) {
		unchanged := time.Now().Add(3 * time.Minute).UTC().Format(time.RFC3339)
		h := writeTok(fmt.Sprintf(`{"token":{"access_token":"ya29.stuck","expiry":%q}}`, unchanged))
		lockPath := filepath.Join(t.TempDir(), "stuck.lock")
		calls := 0
		renewFn := func(string) error {
			calls++
			return nil // simulates agy exiting 0 without rotating the token
		}
		err := testhome.StageAgyAccessTokenWithRenew(h, t.TempDir(), testhome.DefaultAgyMinTokenLifetime, lockPath, renewFn)
		if err == nil {
			t.Fatal("expected named failure when renewal did not advance expiry")
		}
		if !errors.Is(err, testhome.ErrAgyTokenNotRefreshed) {
			t.Fatalf("expected ErrAgyTokenNotRefreshed, got %v", err)
		}
		if !strings.Contains(err.Error(), "agy did not refresh") || !strings.Contains(err.Error(), testhome.AgyRefreshThreshold.String()) {
			t.Fatalf("error must name reason and agy threshold (%s), got %q", testhome.AgyRefreshThreshold, err.Error())
		}
		if calls != 1 {
			t.Fatalf("renew called %d times, want exactly 1 (must never loop)", calls)
		}
	})

	t.Run("fails closed when renewal advances expiry but remains below floor", func(t *testing.T) {
		before := time.Now().Add(1 * time.Minute).UTC().Format(time.RFC3339)
		h := writeTok(fmt.Sprintf(`{"token":{"access_token":"ya29.short1","expiry":%q}}`, before))
		lockPath := filepath.Join(t.TempDir(), "stillshort.lock")
		calls := 0
		renewFn := func(hh string) error {
			calls++
			slightlyNewer := time.Now().Add(2 * time.Minute).UTC().Format(time.RFC3339)
			return os.WriteFile(filepath.Join(hh, ".gemini", "antigravity-cli", "antigravity-oauth-token"),
				[]byte(fmt.Sprintf(`{"token":{"access_token":"ya29.short2","expiry":%q}}`, slightlyNewer)), 0o600)
		}
		err := testhome.StageAgyAccessTokenWithRenew(h, t.TempDir(), testhome.DefaultAgyMinTokenLifetime, lockPath, renewFn)
		if err == nil || errors.Is(err, testhome.ErrAgyTokenNotRefreshed) || !strings.Contains(err.Error(), "expires in") {
			t.Fatalf("expected below-floor error (not ErrAgyTokenNotRefreshed), got %v", err)
		}
		if calls != 1 {
			t.Fatalf("renew called %d times, want 1", calls)
		}
	})

	t.Run("fails closed without renewing when token file becomes corrupt while waiting on lock (M7)", func(t *testing.T) {
		h := writeTok(`{"token":{"access_token":"ya29.stale","expiry":"2020-01-01T00:00:00Z"}}`)
		lockPath := filepath.Join(t.TempDir(), "corrupt.lock")
		locked := make(chan struct{})
		release := make(chan struct{})
		done := make(chan error, 1)
		go func() {
			blockerHome := writeTok(`{"token":{"access_token":"ya29.b","expiry":"2020-01-01T00:00:00Z"}}`)
			done <- testhome.StageAgyAccessTokenWithRenew(blockerHome, t.TempDir(), testhome.DefaultAgyMinTokenLifetime, lockPath, func(bh string) error {
				close(locked)
				<-release
				fresh := time.Now().Add(50 * time.Minute).UTC().Format(time.RFC3339)
				_ = os.WriteFile(filepath.Join(bh, ".gemini", "antigravity-cli", "antigravity-oauth-token"),
					[]byte(fmt.Sprintf(`{"token":{"access_token":"ya29.ok","expiry":%q}}`, fresh)), 0o600)
				// Corrupt the waiter's token file while the waiter is queued on the lock.
				return os.WriteFile(filepath.Join(h, ".gemini", "antigravity-cli", "antigravity-oauth-token"), []byte("{corrupt"), 0o600)
			})
		}()
		<-locked
		waiterDone := make(chan error, 1)
		waiterCalls := 0
		go func() {
			waiterDone <- testhome.StageAgyAccessTokenWithRenew(h, t.TempDir(), testhome.DefaultAgyMinTokenLifetime, lockPath, func(string) error {
				waiterCalls++
				return nil
			})
		}()
		time.Sleep(50 * time.Millisecond)
		close(release)
		if err := <-done; err != nil {
			t.Fatalf("lock holder failed: %v", err)
		}
		err := <-waiterDone
		if err == nil || !strings.Contains(err.Error(), "parse Antigravity credentials") {
			t.Fatalf("expected parse error from post-lock re-check, got %v", err)
		}
		if waiterCalls != 0 {
			t.Fatalf("waiter called renew %d times on corrupt token file, want 0", waiterCalls)
		}
	})

	t.Run("skips renewal when token is refreshed while waiting on process lock (X5)", func(t *testing.T) {
		h := writeTok(`{"token":{"access_token":"ya29.stale","expiry":"2020-01-01T00:00:00Z"}}`)
		// Pre-hold the lock file from a helper goroutine, refresh the host token while the caller
		// is queued on the lock, and assert the post-lock re-check skips renewFn (0 calls).
		lockPath := filepath.Join(t.TempDir(), "contended.lock")
		locked := make(chan struct{})
		release := make(chan struct{})
		done := make(chan error, 1)
		go func() {
			// Use StageAgyAccessTokenWithRenew with a blocking renewFn to hold the exact same lock.
			blockerHome := writeTok(`{"token":{"access_token":"ya29.b","expiry":"2020-01-01T00:00:00Z"}}`)
			done <- testhome.StageAgyAccessTokenWithRenew(blockerHome, t.TempDir(), 15*time.Minute, lockPath, func(bh string) error {
				close(locked)
				<-release
				fresh := time.Now().Add(50 * time.Minute).UTC().Format(time.RFC3339)
				body := []byte(fmt.Sprintf(`{"token":{"access_token":"ya29.fresh2","refresh_token":"1//r2","expiry":%q}}`, fresh))
				_ = os.WriteFile(filepath.Join(bh, ".gemini", "antigravity-cli", "antigravity-oauth-token"), body, 0o600)
				return os.WriteFile(filepath.Join(h, ".gemini", "antigravity-cli", "antigravity-oauth-token"), body, 0o600)
			})
		}()
		<-locked
		waiterDone := make(chan error, 1)
		waiterCalls := 0
		go func() {
			waiterDone <- testhome.StageAgyAccessTokenWithRenew(h, t.TempDir(), 15*time.Minute, lockPath, func(string) error {
				waiterCalls++
				return fmt.Errorf("unexpected second renewal after peer renewed under lock")
			})
		}()
		time.Sleep(50 * time.Millisecond)
		close(release)
		if err := <-done; err != nil {
			t.Fatalf("lock holder failed: %v", err)
		}
		if err := <-waiterDone; err != nil {
			t.Fatalf("waiter failed after peer renewed under lock: %v", err)
		}
		if waiterCalls != 0 {
			t.Fatalf("waiter called renew %d times after post-lock re-check should have succeeded", waiterCalls)
		}
	})
}

func TestBuildLiveEnvStripsAndValidates(t *testing.T) {
	host := t.TempDir()
	const claudeSecret = "sk-ant-ort01-ENV-SECRET"
	_ = os.MkdirAll(filepath.Join(host, ".claude"), 0o700)
	_ = os.WriteFile(filepath.Join(host, ".claude", ".credentials.json"),
		[]byte(fmt.Sprintf(`{"claudeAiOauth":{"accessToken":"acc","refreshToken":%q,"expiresAt":9999999999999}}`, claudeSecret)), 0o600)

	base := []string{
		"PATH=/usr/bin",
		"CLAUDE_CODE_OAUTH_TOKEN=leaked-claude-token",
		"CLAUDE_CONFIG_DIR=" + filepath.Join(host, ".claude"),
		"ANTHROPIC_API_KEY=leaked-anthropic-key",
		"ANTIGRAVITY_CSRF_TOKEN=leaked-agy-csrf",
		"FEOV_RUN=/host/feov/run",
		"PLUGIN_ROOT=/host/plugin",
		"PLUGIN_DATA=/host/plugin-data",
		"XDG_CONFIG_HOME=/host/xdg",
		"APPDATA=/host/appdata",
	}
	fakeHome := t.TempDir()
	env, extraDirs, err := testhome.BuildLiveEnv(base, fakeHome, host, "CLAUDE_CODE_OAUTH_TOKEN=lent-access-token")
	if err != nil {
		t.Fatalf("BuildLiveEnv: %v", err)
	}
	if len(extraDirs) != 0 {
		t.Fatalf("unexpected extraDirs: %v", extraDirs)
	}
	kv := map[string]string{}
	for _, e := range env {
		k, v, _ := strings.Cut(e, "=")
		kv[k] = v
	}
	if kv["HOME"] != fakeHome || kv["USERPROFILE"] != fakeHome {
		t.Fatalf("HOME/USERPROFILE = %q/%q, want %q", kv["HOME"], kv["USERPROFILE"], fakeHome)
	}
	if kv["CLAUDE_CODE_OAUTH_TOKEN"] != "lent-access-token" {
		t.Fatalf("lent extra token lost: %q", kv["CLAUDE_CODE_OAUTH_TOKEN"])
	}
	for _, stripped := range []string{"CLAUDE_CONFIG_DIR", "ANTHROPIC_API_KEY", "ANTIGRAVITY_CSRF_TOKEN", "FEOV_RUN", "PLUGIN_ROOT", "PLUGIN_DATA", "XDG_CONFIG_HOME", "APPDATA"} {
		if v, ok := kv[stripped]; ok {
			t.Errorf("BuildLiveEnv did not strip %s (got %q)", stripped, v)
		}
	}
	if kv["AGY_CLI_DISABLE_AUTO_UPDATE"] != "true" || kv["DISABLE_AUTOUPDATER"] != "1" {
		t.Errorf("BuildLiveEnv did not disable auto-updaters")
	}

	// Rejects fakeHome equal to hostHome or an ancestor of hostHome up front (SHOULD #6).
	if _, _, err := testhome.BuildLiveEnv(base, host, host); err == nil || !strings.Contains(err.Error(), "is or contains host home") {
		t.Fatalf("expected fakeHome==host refusal up front, got %v", err)
	}
	if _, _, err := testhome.BuildLiveEnv(base, filepath.Dir(host), host); err == nil || !strings.Contains(err.Error(), "is or contains host home") {
		t.Fatalf("expected fakeHome ancestor-of-host refusal up front, got %v", err)
	}
	// Rejects any *REFRESH* env key name up front (SHOULD #6).
	if _, _, err := testhome.BuildLiveEnv(base, fakeHome, host, "CLAUDE_CODE_OAUTH_REFRESH_TOKEN=non-host-val"); err == nil || !strings.Contains(err.Error(), "refresh credential") {
		t.Fatalf("expected *REFRESH* env key refusal, got %v", err)
	}
	// Rejects CLAUDE_CONFIG_DIR pointing into hostHome.
	if _, _, err := testhome.BuildLiveEnv(base, fakeHome, host, "CLAUDE_CONFIG_DIR="+filepath.Join(host, ".claude")); err == nil || !strings.Contains(err.Error(), "points into host home") {
		t.Fatalf("expected host-home path refusal, got %v", err)
	}
	// Rejects env var containing host refresh-token bytes (without printing secret).
	if _, _, err := testhome.BuildLiveEnv(base, fakeHome, host, "EXTRA_SECRET="+claudeSecret); err == nil || !strings.Contains(err.Error(), "host Claude refresh token") || strings.Contains(err.Error(), claudeSecret) {
		t.Fatalf("expected refresh-token env refusal without secret leak, got %v", err)
	}
	// External CLAUDE_CONFIG_DIR outside fakeHome (and outside hostHome) is returned for scanning.
	extCfg := t.TempDir()
	_, extDirs, err := testhome.BuildLiveEnv(base, fakeHome, host, "CLAUDE_CONFIG_DIR="+extCfg)
	if err != nil || len(extDirs) != 1 || extDirs[0] != extCfg {
		t.Fatalf("expected external CLAUDE_CONFIG_DIR returned for scan, got dirs=%v err=%v", extDirs, err)
	}

	// Scratch TMPDIR living under hostHome/.claude/scratch/tmp must work cleanly on every OS
	// (setting TMPDIR for Unix and TMP/TEMP for Windows where os.TempDir reads TMP/TEMP) while
	// still refusing hostHome itself and hostHome/.claude.
	scratchTmp := filepath.Join(host, ".claude", "scratch", "tmp")
	scratchFake := filepath.Join(scratchTmp, "fake-home")
	scratchState := filepath.Join(scratchTmp, "state-dir")
	_ = os.MkdirAll(scratchFake, 0o700)
	_ = os.MkdirAll(scratchState, 0o700)
	t.Setenv("TMPDIR", scratchTmp)
	t.Setenv("TMP", scratchTmp)
	t.Setenv("TEMP", scratchTmp)
	if _, _, err := testhome.BuildLiveEnv(base, scratchFake, host, "XDG_STATE_HOME="+scratchState); err != nil {
		t.Fatalf("BuildLiveEnv under scratch TMPDIR failed: %v", err)
	}
	if _, _, err := testhome.BuildLiveEnv(base, scratchFake, host, "CLAUDE_CONFIG_DIR="+filepath.Join(host, ".claude")); err == nil {
		t.Fatalf("expected host .claude refusal even under scratch TMPDIR")
	}
}

// TestLiveEnvGuardsExternalClaudeConfigDir pins M11: calling the public testhome.LiveEnv must
// register a cleanup guard covering both fakeHome and any external CLAUDE_CONFIG_DIR.
func TestLiveEnvGuardsExternalClaudeConfigDir(t *testing.T) {
	host := t.TempDir()
	t.Setenv("HOME", host)
	t.Setenv("USERPROFILE", host)

	ft := &fakeLiveT{t: t}
	fakeHome := t.TempDir()
	extCfg := t.TempDir()
	_ = testhome.LiveEnv(ft, fakeHome, "CLAUDE_CONFIG_DIR="+extCfg)
	if len(ft.cleanups) == 0 {
		t.Fatal("LiveEnv did not register cleanup guard")
	}
	if err := os.WriteFile(filepath.Join(extCfg, ".credentials.json"), []byte(`{"claudeAiOauth":{"refreshToken":"sk-ant-ort01-ext-leak"}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	ft.runCleanups()
	if len(ft.errors) == 0 || !strings.Contains(ft.errors[0], "refreshToken") {
		t.Fatalf("LiveEnv cleanup did not catch leaked refreshToken in external CLAUDE_CONFIG_DIR, got %v", ft.errors)
	}
}

// TestGuardLiveTempHomeStructuralMatrix exercises GuardLiveTempHome via fakeLiveT across every
// structural evasion and fail-closed condition required by #1935 Round 2.
func TestGuardLiveTempHomeStructuralMatrix(t *testing.T) {
	const claudeSecret = "sk-ant-ort01-NEVER-PRINT-CLAUDE-REFRESH"
	const agySecret = "1//NEVER-PRINT-AGY-REFRESH"

	host := t.TempDir()
	claudeDir := filepath.Join(host, ".claude")
	agyDir := filepath.Join(host, ".gemini", "antigravity-cli")
	claudeFile := filepath.Join(claudeDir, ".credentials.json")
	agyFile := filepath.Join(agyDir, "antigravity-oauth-token")
	_ = os.MkdirAll(claudeDir, 0o700)
	_ = os.MkdirAll(agyDir, 0o700)
	validExp := time.Now().Add(45 * time.Minute).UTC().Format(time.RFC3339)
	_ = os.WriteFile(claudeFile, []byte(fmt.Sprintf(`{"claudeAiOauth":{"accessToken":"acc-c","refreshToken":%q,"expiresAt":%d}}`, claudeSecret, time.Now().Add(time.Hour).UnixMilli())), 0o600)
	_ = os.WriteFile(agyFile, []byte(fmt.Sprintf(`{"token":{"access_token":"acc-a","refresh_token":%q,"expiry":%q}}`, agySecret, validExp)), 0o600)

	runGuard := func(t *testing.T, g *testhome.HostRefreshGuard, fakeHome string, extraDirs ...string) []string {
		t.Helper()
		ft := &fakeLiveT{t: t}
		testhome.GuardLiveTempHomeWith(ft, g, fakeHome, extraDirs...)
		ft.runCleanups()
		return ft.errors
	}

	t.Run("clean staged temp home passes", func(t *testing.T) {
		g := testhome.NewHostRefreshGuardFrom(host)
		fake := t.TempDir()
		if err := testhome.StageAgyAccessTokenFrom(host, fake, 15*time.Minute); err != nil {
			t.Fatal(err)
		}
		if errs := runGuard(t, g, fake); len(errs) != 0 {
			t.Fatalf("clean staged temp home failed: %v", errs)
		}
	})

	t.Run("plain copy of Claude credentials fails without printing secret", func(t *testing.T) {
		g := testhome.NewHostRefreshGuardFrom(host)
		fake := t.TempDir()
		dst := filepath.Join(fake, ".claude", ".credentials.json")
		_ = os.MkdirAll(filepath.Dir(dst), 0o700)
		b, _ := os.ReadFile(claudeFile)
		_ = os.WriteFile(dst, b, 0o600)
		errs := runGuard(t, g, fake)
		if len(errs) == 0 || strings.Contains(strings.Join(errs, "\n"), claudeSecret) {
			t.Fatalf("expected failure without secret bytes, got %v", errs)
		}
	})

	t.Run("file symlink (E5) fails", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("symlinks require privileges on Windows")
		}
		g := testhome.NewHostRefreshGuardFrom(host)
		fake := t.TempDir()
		dst := filepath.Join(fake, ".claude", ".credentials.json")
		_ = os.MkdirAll(filepath.Dir(dst), 0o700)
		if err := os.Symlink(claudeFile, dst); err != nil {
			t.Fatal(err)
		}
		if errs := runGuard(t, g, fake); len(errs) == 0 || !strings.Contains(errs[0], "symlink") {
			t.Fatalf("expected symlink failure, got %v", errs)
		}
	})

	t.Run("directory symlink to host .claude (E5b) fails", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("symlinks require privileges on Windows")
		}
		g := testhome.NewHostRefreshGuardFrom(host)
		fake := t.TempDir()
		if err := os.Symlink(claudeDir, filepath.Join(fake, ".claude")); err != nil {
			t.Fatal(err)
		}
		if errs := runGuard(t, g, fake); len(errs) == 0 || !strings.Contains(errs[0], "symlink") {
			t.Fatalf("expected directory symlink failure (E5b), got %v", errs)
		}
	})

	t.Run("directory symlink to host antigravity-cli (E5c) fails", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("symlinks require privileges on Windows")
		}
		g := testhome.NewHostRefreshGuardFrom(host)
		fake := t.TempDir()
		_ = os.MkdirAll(filepath.Join(fake, ".gemini"), 0o700)
		if err := os.Symlink(agyDir, filepath.Join(fake, ".gemini", "antigravity-cli")); err != nil {
			t.Fatal(err)
		}
		if errs := runGuard(t, g, fake); len(errs) == 0 || !strings.Contains(errs[0], "symlink") {
			t.Fatalf("expected directory symlink failure (E5c), got %v", errs)
		}
	})

	t.Run("renamed copied agy token under temp home (E7 cp -r) fails", func(t *testing.T) {
		g := testhome.NewHostRefreshGuardFrom(host)
		fake := t.TempDir()
		dst := filepath.Join(fake, "nested", "renamed-copy")
		_ = os.MkdirAll(filepath.Dir(dst), 0o700)
		b, _ := os.ReadFile(agyFile)
		_ = os.WriteFile(dst, b, 0o600)
		errs := runGuard(t, g, fake)
		if len(errs) == 0 || !strings.Contains(errs[0], "host Antigravity refresh token") || strings.Contains(errs[0], agySecret) {
			t.Fatalf("expected renamed copy failure without secret leak, got %v", errs)
		}
	})

	t.Run("copy-then-rotate with DIFFERENT refreshToken value still fails structurally", func(t *testing.T) {
		g := testhome.NewHostRefreshGuardFrom(host)
		fake := t.TempDir()
		dst := filepath.Join(fake, ".claude", ".credentials.json")
		_ = os.MkdirAll(filepath.Dir(dst), 0o700)
		_ = os.WriteFile(dst, []byte(`{"claudeAiOauth":{"accessToken":"new-acc","refreshToken":"sk-ant-ort01-BRAND-NEW-ROTATED-TOKEN"}}`), 0o600)
		if errs := runGuard(t, g, fake); len(errs) == 0 || !strings.Contains(errs[0], "refreshToken") {
			t.Fatalf("expected structural refreshToken failure on copy-then-rotate, got %v", errs)
		}
	})

	t.Run("copy-then-rotate agy refresh_token with DIFFERENT value still fails structurally", func(t *testing.T) {
		g := testhome.NewHostRefreshGuardFrom(host)
		fake := t.TempDir()
		dst := filepath.Join(fake, ".gemini", "antigravity-cli", "antigravity-oauth-token")
		_ = os.MkdirAll(filepath.Dir(dst), 0o700)
		_ = os.WriteFile(dst, []byte(`{"token":{"access_token":"new-acc","refresh_token":"1//DIFFERENT-ROTATED"}}`), 0o600)
		if errs := runGuard(t, g, fake); len(errs) == 0 || !strings.Contains(errs[0], "refresh_token") {
			t.Fatalf("expected structural refresh_token failure on copy-then-rotate, got %v", errs)
		}
	})

	t.Run("external CLAUDE_CONFIG_DIR holding refreshToken fails", func(t *testing.T) {
		g := testhome.NewHostRefreshGuardFrom(host)
		fake := t.TempDir()
		extCfg := t.TempDir()
		_ = os.WriteFile(filepath.Join(extCfg, ".credentials.json"), []byte(`{"claudeAiOauth":{"refreshToken":"1//ext"}}`), 0o600)
		if errs := runGuard(t, g, fake, extCfg); len(errs) == 0 || !strings.Contains(errs[0], "refreshToken") {
			t.Fatalf("expected external CLAUDE_CONFIG_DIR failure, got %v", errs)
		}
	})

	t.Run("oversized file fails closed as not scanned", func(t *testing.T) {
		g := testhome.NewHostRefreshGuardFrom(host)
		g.MaxScanBytes = 64
		fake := t.TempDir()
		_ = os.WriteFile(filepath.Join(fake, "big.bin"), make([]byte, 128), 0o600)
		if errs := runGuard(t, g, fake); len(errs) == 0 || !strings.Contains(errs[0], "not scanned") {
			t.Fatalf("expected oversized 'not scanned' failure, got %v", errs)
		}
	})

	t.Run("unreadable 0000 file fails closed as not scanned", func(t *testing.T) {
		if runtime.GOOS == "windows" || os.Geteuid() == 0 {
			t.Skip("0000 permissions check requires non-root Unix")
		}
		g := testhome.NewHostRefreshGuardFrom(host)
		fake := t.TempDir()
		p := filepath.Join(fake, "unreadable.dat")
		_ = os.WriteFile(p, []byte("x"), 0o000)
		t.Cleanup(func() { _ = os.Chmod(p, 0o600) })
		if errs := runGuard(t, g, fake); len(errs) == 0 || !strings.Contains(errs[0], "not scanned") {
			t.Fatalf("expected unreadable 'not scanned' failure, got %v", errs)
		}
	})

	t.Run("unreadable 0000 directory fails closed as not scanned (M4d)", func(t *testing.T) {
		if runtime.GOOS == "windows" || os.Geteuid() == 0 {
			t.Skip("0000 permissions check requires non-root Unix")
		}
		g := testhome.NewHostRefreshGuardFrom(host)
		fake := t.TempDir()
		d := filepath.Join(fake, "unreadable-dir")
		_ = os.MkdirAll(d, 0o700)
		_ = os.WriteFile(filepath.Join(d, "inner"), []byte("x"), 0o600)
		if err := os.Chmod(d, 0o000); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = os.Chmod(d, 0o700) })
		if errs := runGuard(t, g, fake); len(errs) == 0 || !strings.Contains(errs[0], "not scanned") {
			t.Fatalf("expected unreadable dir 'not scanned' failure (M4d), got %v", errs)
		}
	})

	t.Run("unparseable credential JSON fails closed as not scanned", func(t *testing.T) {
		g := testhome.NewHostRefreshGuardFrom(host)
		for _, rel := range []string{
			filepath.Join(".claude", ".credentials.json"),
			filepath.Join(".gemini", "antigravity-cli", "antigravity-oauth-token"),
		} {
			fake := t.TempDir()
			dst := filepath.Join(fake, rel)
			_ = os.MkdirAll(filepath.Dir(dst), 0o700)
			_ = os.WriteFile(dst, []byte("{not valid json"), 0o600)
			if errs := runGuard(t, g, fake); len(errs) == 0 || !strings.Contains(errs[0], "not scanned") {
				t.Fatalf("expected unparseable %s 'not scanned' failure, got %v", rel, errs)
			}
		}
	})

	t.Run("host token rotated mid-test then copied under renamed path is caught at cleanup", func(t *testing.T) {
		rotHost := t.TempDir()
		rotClaudeDir := filepath.Join(rotHost, ".claude")
		rotClaudeFile := filepath.Join(rotClaudeDir, ".credentials.json")
		rotAgyDir := filepath.Join(rotHost, ".gemini", "antigravity-cli")
		rotAgyFile := filepath.Join(rotAgyDir, "antigravity-oauth-token")
		_ = os.MkdirAll(rotClaudeDir, 0o700)
		_ = os.MkdirAll(rotAgyDir, 0o700)
		_ = os.WriteFile(rotClaudeFile, []byte(`{"claudeAiOauth":{"refreshToken":"sk-ant-ort01-INITIAL"}}`), 0o600)
		_ = os.WriteFile(rotAgyFile, []byte(`{"token":{"refresh_token":"1//AGY-INITIAL"}}`), 0o600)

		g := testhome.NewHostRefreshGuardFrom(rotHost)
		// Host legitimately rotates both tokens mid-test, then a rogue child copies each NEW token
		// under a renamed path (pins both Claude and Antigravity re-reads, closing N4b).
		const rotatedClaude = "sk-ant-ort01-ROTATED-MID-TEST"
		const rotatedAgy = "1//AGY-ROTATED-MID-TEST"
		_ = os.WriteFile(rotClaudeFile, []byte(fmt.Sprintf(`{"claudeAiOauth":{"refreshToken":%q}}`, rotatedClaude)), 0o600)
		_ = os.WriteFile(rotAgyFile, []byte(fmt.Sprintf(`{"token":{"refresh_token":%q}}`, rotatedAgy)), 0o600)

		fakeClaude := t.TempDir()
		_ = os.WriteFile(filepath.Join(fakeClaude, "sneaky-renamed-copy"), []byte(rotatedClaude), 0o600)
		errs := runGuard(t, g, fakeClaude)
		if len(errs) == 0 || !strings.Contains(errs[0], "host Claude refresh token") || strings.Contains(errs[0], rotatedClaude) {
			t.Fatalf("expected mid-test rotated host Claude token copy to be caught without secret leak, got %v", errs)
		}

		fakeAgy := t.TempDir()
		_ = os.WriteFile(filepath.Join(fakeAgy, "sneaky-agy-renamed-copy"), []byte(rotatedAgy), 0o600)
		errs = runGuard(t, g, fakeAgy)
		if len(errs) == 0 || !strings.Contains(errs[0], "host Antigravity refresh token") || strings.Contains(errs[0], rotatedAgy) {
			t.Fatalf("expected mid-test rotated host Antigravity token copy to be caught without secret leak, got %v", errs)
		}
	})

	t.Run("missing root fails closed as not scanned", func(t *testing.T) {
		g := testhome.NewHostRefreshGuardFrom(host)
		gone := filepath.Join(t.TempDir(), "removed-before-cleanup")
		if errs := runGuard(t, g, gone); len(errs) == 0 || !strings.Contains(errs[0], "not scanned") {
			t.Fatalf("expected missing root 'not scanned' failure, got %v", errs)
		}
	})

	t.Run("NewLiveHome registers cleanup guard by construction", func(t *testing.T) {
		ft := &fakeLiveT{t: t}
		dir := testhome.NewLiveHome(ft)
		if len(ft.cleanups) == 0 {
			t.Fatal("NewLiveHome did not register cleanup guard")
		}
		dst := filepath.Join(dir, ".claude", ".credentials.json")
		_ = os.MkdirAll(filepath.Dir(dst), 0o700)
		_ = os.WriteFile(dst, []byte(`{"claudeAiOauth":{"refreshToken":"sk-ant-ort01-poison"}}`), 0o600)
		ft.runCleanups()
		if len(ft.errors) == 0 {
			t.Fatal("NewLiveHome cleanup did not catch poisoned .credentials.json")
		}
	})

	t.Run("scratch TMPDIR under host .claude permits internal symlinks and rejects escaping ones", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("symlinks require privileges on Windows")
		}
		g := testhome.NewHostRefreshGuardFrom(host)
		scratchFake := filepath.Join(host, ".claude", "scratch", "tmp", "fake-home")
		_ = os.MkdirAll(scratchFake, 0o700)
		_ = os.WriteFile(filepath.Join(scratchFake, "target.txt"), []byte("ok"), 0o600)
		if err := os.Symlink(filepath.Join(scratchFake, "target.txt"), filepath.Join(scratchFake, "internal.link")); err != nil {
			t.Fatal(err)
		}
		if errs := runGuard(t, g, scratchFake); len(errs) != 0 {
			t.Fatalf("internal symlink under scratch TMPDIR failed: %v", errs)
		}
		if err := os.Symlink(claudeDir, filepath.Join(scratchFake, "escape.link")); err != nil {
			t.Fatal(err)
		}
		if errs := runGuard(t, g, scratchFake); len(errs) == 0 || !strings.Contains(errs[0], "symlink") {
			t.Fatalf("escaping symlink under scratch TMPDIR should fail, got %v", errs)
		}
	})
}

// TestLiveTestsUseTesthome is the CI presence gate: in any test package that gates on
// HOOKCORE_LIVE (outside shared/testhome itself), no *_test.go file may set a child HOME= or call
// Setenv("HOME", ...) directly — isolated homes must be constructed via shared/testhome.NewLiveHome
// and LiveEnv/LiveClaudeEnv.
func TestLiveTestsUseTesthome(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("repo root: %v", err)
	}
	type fileInfo struct {
		rel          string
		hasLive      bool
		rawHomeSetup bool
	}
	byPkg := map[string][]fileInfo{}
	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			switch d.Name() {
			case ".git", ".claude", "vendor", "node_modules":
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(d.Name(), "_test.go") {
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		relSlash := filepath.ToSlash(rel)
		if strings.HasPrefix(relSlash, "shared/testhome/") {
			return nil
		}
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return err
		}
		var fi fileInfo
		fi.rel = relSlash
		ast.Inspect(f, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.BasicLit:
				if node.Kind == token.STRING {
					if s, err := strconv.Unquote(node.Value); err == nil {
						if s == "HOOKCORE_LIVE" {
							fi.hasLive = true
						}
						if strings.HasPrefix(s, "HOME=") || strings.HasPrefix(s, "USERPROFILE=") {
							fi.rawHomeSetup = true
						}
					}
				}
			case *ast.CallExpr:
				if sel, ok := node.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == "Setenv" && len(node.Args) >= 1 {
					if lit, ok := node.Args[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
						if s, err := strconv.Unquote(lit.Value); err == nil && (s == "HOME" || s == "USERPROFILE") {
							fi.rawHomeSetup = true
						}
					}
				}
			}
			return true
		})
		pkgDir := filepath.Dir(relSlash)
		byPkg[pkgDir] = append(byPkg[pkgDir], fi)
		return nil
	})
	if err != nil {
		t.Fatalf("walk repo: %v", err)
	}
	var violations []string
	for _, files := range byPkg {
		pkgHasLive := false
		for _, fi := range files {
			if fi.hasLive {
				pkgHasLive = true
				break
			}
		}
		if !pkgHasLive {
			continue
		}
		for _, fi := range files {
			if fi.rawHomeSetup {
				violations = append(violations, fi.rel+": HOOKCORE_LIVE test package sets HOME=/Setenv(\"HOME\") directly instead of using shared/testhome.NewLiveHome/LiveEnv")
			}
		}
	}
	for _, v := range violations {
		t.Error(v)
	}
}
