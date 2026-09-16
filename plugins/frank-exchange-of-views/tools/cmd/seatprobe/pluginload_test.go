package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// WHICH COPY OF THE PLUGIN THE SEAT LOADED IS READ FROM A FIELD, NOT INFERRED.
//
// `--agent frank-exchange-of-views:<agent>` resolves against the plugins the session loaded, and
// with nothing said that is the installed release rather than the tree under test. The failure is
// silent by construction: the seat sits, answers, and is scored — under other text than the one
// the probe was pointed at. The session's `init` event names each loaded plugin and its path, so
// the check is a comparison of paths, and a trajectory that carries no such event is NOT MEASURED
// rather than a pass.

func writeTrajectory(t *testing.T, lines ...string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "trajectory.jsonl")
	if err := os.WriteFile(p, []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

// THE FIXTURE MARSHALS ITS JSON RATHER THAN SPLICING IT. A Windows temporary directory is
// `X:\tmp\...`, and a path concatenated into a JSON string loses its separators to escapes: the
// event then names a directory nobody has, so the check refuses a session that was correct. The
// product read the field properly; only the fixture wrote it by hand, and it failed on the one
// platform its author does not run.
func initEvent(t *testing.T, paths ...string) string {
	t.Helper()
	type plugin struct {
		Name    string `json:"name"`
		Path    string `json:"path"`
		Version string `json:"version"`
	}
	ps := make([]plugin, 0, len(paths)+1)
	for _, p := range paths {
		ps = append(ps, plugin{Name: "frank-exchange-of-views", Path: p, Version: "1.70.0"})
	}
	ps = append(ps, plugin{Name: "prosthetic-conscience", Path: "/elsewhere/prosthetic-conscience", Version: "0.47.0"})
	b, err := json.Marshal(map[string]any{"type": "system", "subtype": "init", "plugins": ps})
	if err != nil {
		t.Fatalf("the fixture could not marshal its own init event: %v", err)
	}
	return string(b)
}

// A path with backslashes must survive the trajectory on every platform, not only where the
// separator is harmless. This is the regression the Windows leg caught.
func TestPluginLoadedFromReadsAPathWithBackslashes(t *testing.T) {
	const win = `X:\tmp\TestPluginLoadedFromReadsAPathWithBackslashes\001`
	p := writeTrajectory(t, initEvent(t, win))
	if err := pluginLoadedFrom(p, win); err != nil {
		t.Errorf("a plugin path carrying backslashes was refused: %v", err)
	}
}

func TestPluginLoadedFromAcceptsTheDirectoryItWasPointedAt(t *testing.T) {
	want := t.TempDir()
	p := writeTrajectory(t,
		`{"type":"system","subtype":"hook","hook_name":"SessionStart"}`,
		initEvent(t, want),
		`{"type":"assistant","message":{"content":[{"type":"text","text":"hello"}]}}`)
	if err := pluginLoadedFrom(p, want); err != nil {
		t.Errorf("a session that loaded the plugin from the directory it was given was refused: %v", err)
	}
}

func TestPluginLoadedFromRefusesTheInstalledRelease(t *testing.T) {
	// The whole reason this exists: the branch's tree was asked for and the release answered.
	installed := "/home/someone/.claude/plugins/cache/special-circumstances/frank-exchange-of-views/1.70.0"
	want := t.TempDir()
	p := writeTrajectory(t, initEvent(t, installed))
	err := pluginLoadedFrom(p, want)
	if err == nil {
		t.Fatal("a seat that loaded the INSTALLED plugin was accepted — the sitting measured other text than the tree under test")
	}
	if !strings.Contains(err.Error(), installed) || !strings.Contains(err.Error(), want) {
		t.Errorf("the refusal must name both paths, so the reader can see which text the seat actually got: %v", err)
	}
}

func TestPluginLoadedFromSaysNotMeasuredWhenNoInitEvent(t *testing.T) {
	// A trajectory with no init event is a session that never said what it loaded. Folding that
	// into the healthy case is the plausible zero, arriving in the instrument.
	p := writeTrajectory(t, `{"type":"assistant","message":{"content":[{"type":"text","text":"hello"}]}}`)
	err := pluginLoadedFrom(p, t.TempDir())
	if err == nil || !strings.Contains(err.Error(), "NOT MEASURED") {
		t.Errorf("a trajectory with no init event must report NOT MEASURED, not a pass: %v", err)
	}
}

func TestPluginLoadedFromSaysSoWhenTheSessionLoadedNoSuchPlugin(t *testing.T) {
	p := writeTrajectory(t,
		`{"type":"system","subtype":"init","plugins":[{"name":"prosthetic-conscience","path":"/elsewhere","version":"0.47.0"}]}`)
	err := pluginLoadedFrom(p, t.TempDir())
	if err == nil {
		t.Fatal("a session carrying no frank-exchange-of-views plugin at all was accepted")
	}
}

func TestPluginLoadedFromRefusesAMissingTrajectory(t *testing.T) {
	if err := pluginLoadedFrom(filepath.Join(t.TempDir(), "absent.jsonl"), t.TempDir()); err == nil {
		t.Error("a dispatch that left no trajectory cannot have its loaded plugin checked, and must say so")
	}
}

// A RELATIVE OR SYMLINKED -plugin-dir IS THE SAME TREE. The session reports the resolved absolute
// path, so comparing the strings as given would refuse a run that was entirely correct — and the
// next person would delete the check rather than the flag.
func TestPluginLoadedFromResolvesTheDirectoryBeforeComparing(t *testing.T) {
	real := t.TempDir()
	link := filepath.Join(t.TempDir(), "link")
	if err := os.Symlink(real, link); err != nil {
		t.Skipf("symlinks unavailable here: %v", err)
	}
	p := writeTrajectory(t, initEvent(t, real))
	if err := pluginLoadedFrom(p, link); err != nil {
		t.Errorf("a symlink to the loaded directory was refused: %v", err)
	}
}
