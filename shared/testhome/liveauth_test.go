package testhome_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/shared/testhome"
)

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

	// Missing file fails loudly.
	if _, err := testhome.ClaudeOAuthEnvFrom(t.TempDir(), 15*time.Minute); err == nil {
		t.Fatal("expected error for missing .credentials.json")
	}

	// Missing accessToken fails loudly.
	emptyHost := t.TempDir()
	_ = os.MkdirAll(filepath.Join(emptyHost, ".claude"), 0o700)
	_ = os.WriteFile(filepath.Join(emptyHost, ".claude", ".credentials.json"), []byte(`{"claudeAiOauth":{"refreshToken":"only-refresh","expiresAt":9999999999999}}`), 0o600)
	if _, err := testhome.ClaudeOAuthEnvFrom(emptyHost, 15*time.Minute); err == nil {
		t.Fatal("expected error when accessToken is absent")
	}

	// Missing/zero expiresAt fails closed.
	noExpHost := t.TempDir()
	_ = os.MkdirAll(filepath.Join(noExpHost, ".claude"), 0o700)
	_ = os.WriteFile(filepath.Join(noExpHost, ".claude", ".credentials.json"), []byte(`{"claudeAiOauth":{"accessToken":"sk-ant-oat01-noexp"}}`), 0o600)
	if _, err := testhome.ClaudeOAuthEnvFrom(noExpHost, 15*time.Minute); err == nil {
		t.Fatal("expected error when expiresAt is absent")
	}

	// Expired (1970) or below-floor expiresAt fails closed.
	shortHost := t.TempDir()
	_ = os.MkdirAll(filepath.Join(shortHost, ".claude"), 0o700)
	shortExp := time.Now().Add(30 * time.Second).UnixMilli()
	_ = os.WriteFile(filepath.Join(shortHost, ".claude", ".credentials.json"),
		[]byte(fmt.Sprintf(`{"claudeAiOauth":{"accessToken":"sk-ant-oat01-short","expiresAt":%d}}`, shortExp)), 0o600)
	if _, err := testhome.ClaudeOAuthEnvFrom(shortHost, 15*time.Minute); err == nil || !strings.Contains(err.Error(), "expires in") {
		t.Fatalf("expected short-lifetime error for 30s Claude token, got %v", err)
	}
	_ = os.WriteFile(filepath.Join(shortHost, ".claude", ".credentials.json"),
		[]byte(`{"claudeAiOauth":{"accessToken":"sk-ant-oat01-1970","expiresAt":1000}}`), 0o600)
	if _, err := testhome.ClaudeOAuthEnvFrom(shortHost, 15*time.Minute); err == nil {
		t.Fatal("expected error for 1970 expiresAt")
	}
}

func TestStageAgyAccessTokenFrom(t *testing.T) {
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

	// Failure cases (all fail closed).
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
	if err := testhome.StageAgyAccessTokenFrom(writeTok(`{"token":{"access_token":"ya29.expired","expiry":"2020-01-01T00:00:00Z"}}`), t.TempDir(), 15*time.Minute); err == nil {
		t.Fatal("expected error for expired token")
	}
}

func TestLiveEnvStripsHarnessAndCredentialVars(t *testing.T) {
	t.Setenv("CLAUDE_CODE_OAUTH_TOKEN", "leaked-claude-token")
	t.Setenv("CLAUDE_CONFIG_DIR", "/host/claude")
	t.Setenv("ANTHROPIC_API_KEY", "leaked-anthropic-key")
	t.Setenv("ANTIGRAVITY_CSRF_TOKEN", "leaked-agy-csrf")
	t.Setenv("FEOV_RUN", "/host/feov/run")
	t.Setenv("PLUGIN_ROOT", "/host/plugin")
	t.Setenv("PLUGIN_DATA", "/host/plugin-data")
	t.Setenv("XDG_CONFIG_HOME", "/host/xdg")

	fakeHome := t.TempDir()
	env := testhome.LiveEnv(fakeHome, "CLAUDE_CODE_OAUTH_TOKEN=lent-access-token")
	kv := map[string]string{}
	for _, e := range env {
		k, v, _ := strings.Cut(e, "=")
		kv[k] = v
	}
	if kv["HOME"] != fakeHome || kv["USERPROFILE"] != fakeHome {
		t.Fatalf("HOME/USERPROFILE = %q/%q, want %q", kv["HOME"], kv["USERPROFILE"], fakeHome)
	}
	if kv["CLAUDE_CODE_OAUTH_TOKEN"] != "lent-access-token" {
		t.Fatalf("lent extra token lost or overwritten: %q", kv["CLAUDE_CODE_OAUTH_TOKEN"])
	}
	for _, stripped := range []string{"CLAUDE_CONFIG_DIR", "ANTHROPIC_API_KEY", "ANTIGRAVITY_CSRF_TOKEN", "FEOV_RUN", "PLUGIN_ROOT", "PLUGIN_DATA", "XDG_CONFIG_HOME"} {
		if v, ok := kv[stripped]; ok {
			t.Errorf("LiveEnv did not strip %s (got %q)", stripped, v)
		}
	}
	if kv["AGY_CLI_DISABLE_AUTO_UPDATE"] != "true" || kv["DISABLE_AUTOUPDATER"] != "1" {
		t.Errorf("LiveEnv did not disable auto-updaters: agy=%q claude=%q", kv["AGY_CLI_DISABLE_AUTO_UPDATE"], kv["DISABLE_AUTOUPDATER"])
	}
}

// TestHostRefreshGuardCatchesEveryLeakAndRotation verifies the outcome-based #1890 gate catches
// raw copies, symlinks (E5), directory copies (E7), and host refresh-token rotation without ever
// printing secret bytes in error messages.
func TestHostRefreshGuardCatchesEveryLeakAndRotation(t *testing.T) {
	const claudeSecret = "sk-ant-ort01-NEVER-PRINT-CLAUDE-REFRESH"
	const agySecret = "1//NEVER-PRINT-AGY-REFRESH"

	host := t.TempDir()
	claudeFile := filepath.Join(host, ".claude", ".credentials.json")
	agyFile := filepath.Join(host, ".gemini", "antigravity-cli", "antigravity-oauth-token")
	_ = os.MkdirAll(filepath.Dir(claudeFile), 0o700)
	_ = os.MkdirAll(filepath.Dir(agyFile), 0o700)
	validExp := time.Now().Add(45 * time.Minute).UTC().Format(time.RFC3339)
	_ = os.WriteFile(claudeFile, []byte(fmt.Sprintf(`{"claudeAiOauth":{"accessToken":"acc-c","refreshToken":%q,"expiresAt":%d}}`, claudeSecret, time.Now().Add(time.Hour).UnixMilli())), 0o600)
	_ = os.WriteFile(agyFile, []byte(fmt.Sprintf(`{"token":{"access_token":"acc-a","refresh_token":%q,"expiry":%q}}`, agySecret, validExp)), 0o600)

	guard := testhome.NewHostRefreshGuardFrom(host)

	t.Run("clean staged temp home passes", func(t *testing.T) {
		fake := t.TempDir()
		if err := testhome.StageAgyAccessTokenFrom(host, fake, 15*time.Minute); err != nil {
			t.Fatal(err)
		}
		if err := guard.CheckTempHome(fake); err != nil {
			t.Fatalf("clean staged temp home rejected: %v", err)
		}
		if err := guard.VerifyHostUnchanged(); err != nil {
			t.Fatalf("unchanged host rejected: %v", err)
		}
	})

	t.Run("catches raw copy of Claude credentials without printing secret", func(t *testing.T) {
		fake := t.TempDir()
		dst := filepath.Join(fake, ".claude", ".credentials.json")
		_ = os.MkdirAll(filepath.Dir(dst), 0o700)
		b, _ := os.ReadFile(claudeFile)
		_ = os.WriteFile(dst, b, 0o600)
		err := guard.CheckTempHome(fake)
		if err == nil || !strings.Contains(err.Error(), "host Claude refresh token") {
			t.Fatalf("expected Claude refresh leak detection, got %v", err)
		}
		if strings.Contains(err.Error(), claudeSecret) {
			t.Fatalf("error message leaked secret bytes: %v", err)
		}
	})

	t.Run("catches os.Symlink to host credential file (E5)", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("symlinks require elevated privileges on Windows")
		}
		fake := t.TempDir()
		dst := filepath.Join(fake, ".claude", ".credentials.json")
		_ = os.MkdirAll(filepath.Dir(dst), 0o700)
		if err := os.Symlink(claudeFile, dst); err != nil {
			t.Fatal(err)
		}
		err := guard.CheckTempHome(fake)
		if err == nil || !strings.Contains(err.Error(), "host Claude refresh token") {
			t.Fatalf("expected symlink leak detection, got %v", err)
		}
	})

	t.Run("catches arbitrary copied agy token file under temp home (E7)", func(t *testing.T) {
		fake := t.TempDir()
		dst := filepath.Join(fake, "nested", "renamed-copy")
		_ = os.MkdirAll(filepath.Dir(dst), 0o700)
		b, _ := os.ReadFile(agyFile)
		_ = os.WriteFile(dst, b, 0o600)
		err := guard.CheckTempHome(fake)
		if err == nil || !strings.Contains(err.Error(), "host Antigravity refresh token") {
			t.Fatalf("expected agy refresh leak detection, got %v", err)
		}
		if strings.Contains(err.Error(), agySecret) {
			t.Fatalf("error message leaked secret bytes: %v", err)
		}
	})

	t.Run("catches host refresh token rotation", func(t *testing.T) {
		_ = os.WriteFile(claudeFile, []byte(`{"claudeAiOauth":{"accessToken":"acc-c2","refreshToken":"sk-ant-ort01-ROTATED","expiresAt":9999999999999}}`), 0o600)
		err := guard.VerifyHostUnchanged()
		if err == nil || !strings.Contains(err.Error(), "host Claude refresh token changed") {
			t.Fatalf("expected host rotation detection, got %v", err)
		}
		if strings.Contains(err.Error(), claudeSecret) || strings.Contains(err.Error(), "ROTATED") {
			t.Fatalf("rotation error leaked secret bytes: %v", err)
		}
	})
}
