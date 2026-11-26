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

// TestStageAgyAccessTokenWaitThenRenew exercises the [5m, 6m) wait-then-renew branch end to end
// with an injected clock sleep so it runs deterministically in 0ms and proves it sleeps past
// AgyRefreshThreshold, renews once, and stages the fresh token.
func TestStageAgyAccessTokenWaitThenRenew(t *testing.T) {
	host := t.TempDir()
	cache := t.TempDir()
	t.Setenv("HOME", host)
	t.Setenv("USERPROFILE", host)
	t.Setenv("XDG_CACHE_HOME", cache)

	tokPath := filepath.Join(host, ".gemini", "antigravity-cli", "antigravity-oauth-token")
	if err := os.MkdirAll(filepath.Dir(tokPath), 0o700); err != nil {
		t.Fatal(err)
	}
	inBand := time.Now().Add(5*time.Minute + 20*time.Second).UTC().Format(time.RFC3339)
	if err := os.WriteFile(tokPath, []byte(fmt.Sprintf(`{"token":{"access_token":"ya29.inband","refresh_token":"1//r","expiry":%q}}`, inBand)), 0o600); err != nil {
		t.Fatal(err)
	}

	origRenew := renewHostAgyTokenFn
	origSleep := sleepAgyRenewFn
	t.Cleanup(func() {
		renewHostAgyTokenFn = origRenew
		sleepAgyRenewFn = origSleep
	})

	var slept []time.Duration
	sleepAgyRenewFn = func(d time.Duration) { slept = append(slept, d) }
	calls := 0
	renewHostAgyTokenFn = func(hh string) error {
		calls++
		fresh := time.Now().Add(55 * time.Minute).UTC().Format(time.RFC3339)
		return os.WriteFile(filepath.Join(hh, ".gemini", "antigravity-cli", "antigravity-oauth-token"),
			[]byte(fmt.Sprintf(`{"auth_method":"oauth","token":{"access_token":"ya29.afterwait","refresh_token":"1//r2","expiry":%q}}`, fresh)), 0o600)
	}

	fake := t.TempDir()
	if err := StageAgyAccessToken(fake); err != nil {
		t.Fatalf("StageAgyAccessToken in [5m, 6m) band: %v", err)
	}
	if len(slept) != 1 || slept[0] < 20*time.Second || slept[0] > 30*time.Second {
		t.Fatalf("expected single wait around 25s (20s + 5s margin), got %v", slept)
	}
	if calls != 1 {
		t.Fatalf("renewHostAgyTokenFn called %d times, want 1", calls)
	}
	staged, err := os.ReadFile(filepath.Join(fake, ".gemini", "antigravity-cli", "antigravity-oauth-token"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(staged), "ya29.afterwait") || strings.Contains(string(staged), "refresh_token") {
		t.Fatalf("staged token invalid after wait-then-renew: %s", string(staged))
	}

	// M4: if another writer (e.g. host agy or universe.sh) refreshes the token during sleep,
	// post-sleep re-check stages the refreshed token and skips renewHostAgyTokenFn (0 calls).
	t.Run("skips renew when external writer refreshes token during sleep (M4)", func(t *testing.T) {
		inBand2 := time.Now().Add(5*time.Minute + 15*time.Second).UTC().Format(time.RFC3339)
		if err := os.WriteFile(tokPath, []byte(fmt.Sprintf(`{"token":{"access_token":"ya29.beforewait","refresh_token":"1//r","expiry":%q}}`, inBand2)), 0o600); err != nil {
			t.Fatal(err)
		}
		sleepCalls := 0
		sleepAgyRenewFn = func(time.Duration) {
			sleepCalls++
			fresh := time.Now().Add(55 * time.Minute).UTC().Format(time.RFC3339)
			_ = os.WriteFile(tokPath, []byte(fmt.Sprintf(`{"auth_method":"oauth","token":{"access_token":"ya29.external","refresh_token":"1//r3","expiry":%q}}`, fresh)), 0o600)
		}
		renewCalls := 0
		renewHostAgyTokenFn = func(string) error {
			renewCalls++
			return fmt.Errorf("unexpected renew after external refresh during sleep")
		}
		fake2 := t.TempDir()
		if err := StageAgyAccessToken(fake2); err != nil {
			t.Fatalf("StageAgyAccessToken after external refresh during sleep: %v", err)
		}
		if sleepCalls != 1 {
			t.Fatalf("sleep called %d times, want 1", sleepCalls)
		}
		if renewCalls != 0 {
			t.Fatalf("renew called %d times after external refresh during sleep, want 0", renewCalls)
		}
		staged2, err := os.ReadFile(filepath.Join(fake2, ".gemini", "antigravity-cli", "antigravity-oauth-token"))
		if err != nil || !strings.Contains(string(staged2), "ya29.external") {
			t.Fatalf("expected externally refreshed token staged, got err=%v body=%s", err, string(staged2))
		}
	})
}
