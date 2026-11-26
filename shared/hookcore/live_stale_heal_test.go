package hookcore_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/shared/testhome"
)

// TestLiveAntigravityStaleInstallSelfHeal pins #1933 end-to-end under real `agy -p`:
//   - installs the PR-built binary-shipping plugins (prosthetic-conscience, gray-area,
//     frank-exchange-of-views) with hooks/fetch-bin.sh into an isolated testhome;
//   - points SC_RELEASE_BASE_URL at a local httptest server so ensure runs without network downloads;
//   - (c) verifies a current .fetch/installed stamp stays completely silent (zero fetch hits,
//     no .fetch/notified.installing, no "installing its hook binaries" in transcript);
//   - (b) makes .fetch/installed stale on the SAME conversation and resumes it (--conversation,
//     where SessionStart does not fire and only PreInvocation invocationNum 0 fires), verifying
//     all three plugins' ensure path runs (local server receives each tag's SHA256SUMS request,
//     .fetch/notified.installing or .fetch/log is created, and transcript carries the installing notice);
//   - (a) verifies a fresh conversation with stale .fetch/installed triggers ensure on Turn 1.
func TestLiveAntigravityStaleInstallSelfHeal(t *testing.T) {
	if os.Getenv("HOOKCORE_LIVE") != "1" {
		t.Skip("skipping live test without HOOKCORE_LIVE=1")
	}
	agyBin, err := exec.LookPath("agy")
	if err != nil {
		t.Fatalf("agy CLI not found on PATH under HOOKCORE_LIVE=1: %v", err)
	}
	repoRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatalf("abs repo root: %v", err)
	}

	plugins := []string{"prosthetic-conscience", "gray-area", "frank-exchange-of-views"}

	stageAllWithFetchBin := func(t *testing.T, fakeHome string) map[string]string {
		t.Helper()
		tags := make(map[string]string, len(plugins))
		for _, name := range plugins {
			stageShippedPlugin(t, fakeHome, repoRoot, name)
			srcFetch := filepath.Join(repoRoot, "plugins", name, "hooks", "fetch-bin.sh")
			destPlugin := filepath.Join(fakeHome, ".gemini", "config", "plugins", name)
			if err := os.MkdirAll(filepath.Join(destPlugin, "hooks"), 0o755); err != nil {
				t.Fatalf("mkdir hooks for %s: %v", name, err)
			}
			data, err := os.ReadFile(srcFetch)
			if err != nil {
				t.Fatalf("read %s: %v", srcFetch, err)
			}
			if err := os.WriteFile(filepath.Join(destPlugin, "hooks", "fetch-bin.sh"), data, 0o755); err != nil {
				t.Fatalf("write fetch-bin.sh for %s: %v", name, err)
			}
			// Mirror a real plugin install's tools/cmd/<bin>/ directories so fetch-bin.sh discovers
			// the plugin's binary list when healing a stale stamp.
			cmdEntries, err := os.ReadDir(filepath.Join(repoRoot, "plugins", name, "tools", "cmd"))
			if err != nil {
				t.Fatalf("read tools/cmd for %s: %v", name, err)
			}
			for _, ce := range cmdEntries {
				if !ce.IsDir() || strings.HasPrefix(ce.Name(), ".") {
					continue
				}
				if err := os.MkdirAll(filepath.Join(destPlugin, "tools", "cmd", ce.Name()), 0o755); err != nil {
					t.Fatalf("mkdir tools/cmd/%s for %s: %v", ce.Name(), name, err)
				}
			}
			var manifest struct{ Name, Version string }
			pData, err := os.ReadFile(filepath.Join(destPlugin, ".antigravity-plugin", "plugin.json"))
			if err != nil {
				t.Fatalf("read manifest for %s: %v", name, err)
			}
			if err := json.Unmarshal(pData, &manifest); err != nil || manifest.Name == "" || manifest.Version == "" {
				t.Fatalf("parse manifest for %s: err=%v manifest=%+v", name, err, manifest)
			}
			tags[name] = manifest.Name + "--v" + manifest.Version
		}
		return tags
	}

	newRecordingServer := func(t *testing.T) (*httptest.Server, func() []string) {
		t.Helper()
		var mu sync.Mutex
		var paths []string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mu.Lock()
			paths = append(paths, r.URL.Path)
			mu.Unlock()
			http.NotFound(w, r)
		}))
		t.Cleanup(srv.Close)
		return srv, func() []string {
			mu.Lock()
			defer mu.Unlock()
			return append([]string(nil), paths...)
		}
	}

	waitForPluginFetches := func(t *testing.T, label string, getPaths func() []string, tags map[string]string) {
		t.Helper()
		deadline := time.Now().Add(20 * time.Second)
		for {
			paths := getPaths()
			var missing []string
			for _, name := range plugins {
				want := "/" + tags[name] + "/SHA256SUMS"
				found := false
				for _, p := range paths {
					if p == want {
						found = true
						break
					}
				}
				if !found {
					missing = append(missing, want)
				}
			}
			if len(missing) == 0 {
				return
			}
			if time.Now().After(deadline) {
				t.Fatalf("%s: timed out waiting for fetch-bin.sh ensure requests %v; got paths=%v", label, missing, paths)
			}
			time.Sleep(250 * time.Millisecond)
		}
	}

	makeStale := func(t *testing.T, fakeHome string) {
		t.Helper()
		for _, name := range plugins {
			fetchDir := filepath.Join(fakeHome, ".gemini", "config", "plugins", name, ".fetch")
			if err := os.WriteFile(filepath.Join(fetchDir, "installed"), []byte(name+"--v0.0.1\n"), 0o644); err != nil {
				t.Fatalf("write stale stamp for %s: %v", name, err)
			}
		}
	}

	readTranscript := func(t *testing.T, fakeHome, conv string) string {
		t.Helper()
		p := filepath.Join(fakeHome, ".gemini", "antigravity-cli", "brain", conv, ".system_generated", "logs", "transcript.jsonl")
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("read transcript %s: %v", p, err)
		}
		return string(b)
	}

	runTurn := func(t *testing.T, workDir string, env []string, conv string) {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		defer cancel()
		args := []string{"-p", "Reply with the single word OK.", "--model", "gemini-3.8-flash-low", "--output-format", "text"}
		if conv != "" {
			args = append(args, "--conversation", conv)
		}
		cmd := exec.CommandContext(ctx, agyBin, args...)
		cmd.Dir = workDir
		cmd.Env = env
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("agy -p (conv=%q) failed: %v\n%s", conv, err, string(out))
		}
	}

	t.Run("current stamp stays silent then resumed stale turn heals via PreInvocation", func(t *testing.T) {
		fakeHome := testhome.NewLiveHome(t)
		copyAgyCredentials(t, fakeHome)
		tags := stageAllWithFetchBin(t, fakeHome)
		srv, getPaths := newRecordingServer(t)
		env := liveTestEnv(t, fakeHome, "SC_RELEASE_BASE_URL="+srv.URL)
		workDir := t.TempDir()

		// (c) Current stamp: Turn 1 must stay completely silent and issue zero fetch requests.
		runTurn(t, workDir, env, "")
		time.Sleep(1 * time.Second)
		if paths := getPaths(); len(paths) != 0 {
			t.Fatalf("current stamp triggered unexpected fetch requests: %v", paths)
		}
		for _, name := range plugins {
			fetchDir := filepath.Join(fakeHome, ".gemini", "config", "plugins", name, ".fetch")
			for _, artifact := range []string{"notified.installing", "log"} {
				if _, err := os.Stat(filepath.Join(fetchDir, artifact)); err == nil {
					t.Errorf("current stamp unexpectedly created %s/.fetch/%s", name, artifact)
				}
			}
		}
		convs := listAgyBrainConvs(fakeHome)
		if len(convs) != 1 {
			t.Fatalf("want 1 conversation after Turn 1, got %v", convs)
		}
		var conv string
		for id := range convs {
			conv = id
		}
		tr1 := readTranscript(t, fakeHome, conv)
		if strings.Contains(tr1, "installing its hook binaries") {
			t.Errorf("current stamp leaked 'installing its hook binaries' into Turn 1 transcript")
		}

		// (b) Resumed conversation with stale stamp: SessionStart does not fire on --conversation,
		// so only PreInvocation (invocationNum == 0) triggers ensure across all three plugins.
		makeStale(t, fakeHome)
		runTurn(t, workDir, env, conv)
		waitForPluginFetches(t, "resumed stale turn", getPaths, tags)

		for _, name := range plugins {
			fetchDir := filepath.Join(fakeHome, ".gemini", "config", "plugins", name, ".fetch")
			_, errNotif := os.Stat(filepath.Join(fetchDir, "notified.installing"))
			_, errLog := os.Stat(filepath.Join(fetchDir, "log"))
			if errNotif != nil && errLog != nil {
				t.Errorf("resumed stale turn left neither .fetch/notified.installing nor .fetch/log for %s", name)
			}
		}
		tr2 := readTranscript(t, fakeHome, conv)
		if !strings.Contains(tr2, "installing its hook binaries") {
			t.Errorf("resumed stale turn transcript missing 'installing its hook binaries' notice")
		}
		assertNoAgyPanics(t, fakeHome)
	})

	t.Run("fresh conversation with stale stamp heals on Turn 1", func(t *testing.T) {
		fakeHome := testhome.NewLiveHome(t)
		copyAgyCredentials(t, fakeHome)
		tags := stageAllWithFetchBin(t, fakeHome)
		makeStale(t, fakeHome)
		srv, getPaths := newRecordingServer(t)
		env := liveTestEnv(t, fakeHome, "SC_RELEASE_BASE_URL="+srv.URL)
		workDir := t.TempDir()

		// (a) Fresh conversation with stale stamp triggers ensure on Turn 1.
		runTurn(t, workDir, env, "")
		waitForPluginFetches(t, "fresh stale turn", getPaths, tags)

		convs := listAgyBrainConvs(fakeHome)
		if len(convs) != 1 {
			t.Fatalf("want 1 conversation after fresh stale Turn 1, got %v", convs)
		}
		var conv string
		for id := range convs {
			conv = id
		}
		tr := readTranscript(t, fakeHome, conv)
		if !strings.Contains(tr, "installing its hook binaries") {
			t.Errorf("fresh stale turn transcript missing 'installing its hook binaries' notice")
		}
		assertNoAgyPanics(t, fakeHome)
	})
}
