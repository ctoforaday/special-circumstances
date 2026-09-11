package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// fixture: plugin "bins" ships binaries at 1.2.0; plugin "prose" ships none.
func fixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "plugins", "bins", ".claude-plugin", "plugin.json"), `{"name":"bins","version":"1.2.0"}`)
	if err := os.MkdirAll(filepath.Join(root, "plugins", "bins", "tools", "cmd", "x"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(root, "plugins", "prose", ".claude-plugin", "plugin.json"), `{"name":"prose","version":"0.1.0"}`)
	return root
}

const manifestIn = `{"name":"m","owner":{"name":"o"},"description":"d — e","plugins":[` +
	`{"name":"bins","source":"./plugins/bins","description":"b"},` +
	`{"name":"prose","source":"./plugins/prose","description":"p"}]}`

func sources(t *testing.T, out []byte) map[string]json.RawMessage {
	t.Helper()
	var m market
	if err := json.Unmarshal(out, &m); err != nil {
		t.Fatal(err)
	}
	got := map[string]json.RawMessage{}
	for _, p := range m.Plugins {
		got[p.Name] = p.Source
	}
	return got
}

func TestAPluginWithBinariesIsPinnedToItsTag(t *testing.T) {
	out, err := render(fixture(t), []byte(manifestIn))
	if err != nil {
		t.Fatal(err)
	}
	var src gitSubdir
	if err := json.Unmarshal(sources(t, out)["bins"], &src); err != nil {
		t.Fatalf("bins should carry a git-subdir object: %v", err)
	}
	want := gitSubdir{Source: "git-subdir", URL: repoURL, Path: "plugins/bins", Ref: "bins--v1.2.0"}
	if src != want {
		t.Fatalf("got %+v, want %+v", src, want)
	}
}

func TestAPluginWithoutBinariesKeepsItsRelativePath(t *testing.T) {
	out, err := render(fixture(t), []byte(manifestIn))
	if err != nil {
		t.Fatal(err)
	}
	if got := string(sources(t, out)["prose"]); got != `"./plugins/prose"` {
		t.Fatalf("prose source = %s", got)
	}
}

func TestRenderIsStableAndKeepsProse(t *testing.T) {
	root := fixture(t)
	once, err := render(root, []byte(manifestIn))
	if err != nil {
		t.Fatal(err)
	}
	twice, err := render(root, once)
	if err != nil {
		t.Fatal(err)
	}
	if string(once) != string(twice) {
		t.Fatalf("render is not a fixed point:\n%s\n---\n%s", once, twice)
	}
	if !strings.Contains(string(once), "d — e") {
		t.Fatalf("the description was escaped or lost:\n%s", once)
	}
}

func TestAVersionBumpMovesTheRef(t *testing.T) {
	root := fixture(t)
	before, _ := render(root, []byte(manifestIn))
	writeFile(t, filepath.Join(root, "plugins", "bins", ".claude-plugin", "plugin.json"), `{"name":"bins","version":"1.3.0"}`)
	after, err := render(root, before)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(after), `"ref": "bins--v1.3.0"`) {
		t.Fatalf("the bump did not reach the ref:\n%s", after)
	}
}

func TestAnUnknownManifestFieldIsRefused(t *testing.T) {
	in := strings.Replace(manifestIn, `"description":"d — e"`, `"description":"d — e","metadata":{}`, 1)
	if _, err := render(fixture(t), []byte(in)); err == nil {
		t.Fatal("a field the generator does not carry would have been dropped silently")
	}
}

func TestAPluginWithBinariesButNoVersionIsRefused(t *testing.T) {
	root := fixture(t)
	writeFile(t, filepath.Join(root, "plugins", "bins", ".claude-plugin", "plugin.json"), `{"name":"bins"}`)
	if _, err := render(root, []byte(manifestIn)); err == nil {
		t.Fatal("a plugin with binaries and no version has no tag to pin to")
	}
}
