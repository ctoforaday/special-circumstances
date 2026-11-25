package testhome

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestStageAgyAccessTokenWiresRenewHostAgyToken pins the production StageAgyAccessToken wiring
// through renewHostAgyTokenFn so passing a nil renewer fails in CI (closes X4b).
func TestStageAgyAccessTokenWiresRenewHostAgyToken(t *testing.T) {
	host := t.TempDir()
	cache := t.TempDir()
	t.Setenv("HOME", host)
	t.Setenv("USERPROFILE", host)
	t.Setenv("XDG_CACHE_HOME", cache)

	tokPath := filepath.Join(host, ".gemini", "antigravity-cli", "antigravity-oauth-token")
	if err := os.MkdirAll(filepath.Dir(tokPath), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(tokPath, []byte(`{"token":{"access_token":"ya29.expired","refresh_token":"1//r","expiry":"2020-01-01T00:00:00Z"}}`), 0o600); err != nil {
		t.Fatal(err)
	}

	orig := renewHostAgyTokenFn
	t.Cleanup(func() { renewHostAgyTokenFn = orig })
	calls := 0
	renewHostAgyTokenFn = func(hh string) error {
		calls++
		fresh := time.Now().Add(50 * time.Minute).UTC().Format(time.RFC3339)
		return os.WriteFile(filepath.Join(hh, ".gemini", "antigravity-cli", "antigravity-oauth-token"),
			[]byte(fmt.Sprintf(`{"auth_method":"oauth","token":{"access_token":"ya29.renewed","refresh_token":"1//r2","expiry":%q}}`, fresh)), 0o600)
	}

	fake := t.TempDir()
	if err := StageAgyAccessToken(fake); err != nil {
		t.Fatalf("StageAgyAccessToken: %v", err)
	}
	if calls != 1 {
		t.Fatalf("renewHostAgyTokenFn called %d times, want 1", calls)
	}
	staged, err := os.ReadFile(filepath.Join(fake, ".gemini", "antigravity-cli", "antigravity-oauth-token"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(staged), "ya29.renewed") || strings.Contains(string(staged), "refresh_token") {
		t.Fatalf("staged token invalid: %s", string(staged))
	}
}
