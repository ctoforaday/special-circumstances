package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// write stages a one-package module for the tool to read.
func write(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, body := range files {
		p := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

const goMod = "module example.test/m\n\ngo 1.27\n"

// THE GATE SEES A DEAD FUNCTION, AND ONLY A DEAD ONE. A gate that lists nothing passes for ever,
// so the pinned tool is driven here over a module holding exactly one function nothing reaches,
// one a test alone reaches, and one only a second `main` package reaches.
func TestTheGateListsWhatNothingReachesAndNothingElse(t *testing.T) {
	tool, err := pinnedTool("..")
	if err != nil {
		t.Fatal(err)
	}
	dir := write(t, map[string]string{
		"go.mod":              goMod,
		"cmd/tool/main.go":    "package main\n\nimport \"example.test/m/lib\"\n\nfunc main() { lib.Live() }\n",
		"devcmd/gen/main.go":  "package main\n\nimport \"example.test/m/lib\"\n\nfunc main() { lib.OnlyTheGenerator() }\n",
		"lib/lib.go":          "package lib\n\nfunc Live() {}\n\nfunc OnlyTheGenerator() {}\n\nfunc OnlyATest() {}\n\nfunc dead() {}\n",
		"lib/lib_test.go":     "package lib\n\nimport \"testing\"\n\nfunc TestIt(t *testing.T) { OnlyATest() }\n",
		"lib/lib_windows.go":  "package lib\n\nfunc init() { onlyWindows() }\n",
		"lib/lib_platform.go": "package lib\n\nfunc onlyWindows() {}\n",
	})
	for goos, want := range map[string][]string{
		"linux":   {"dead", "onlyWindows"},
		"windows": {"dead"},
	} {
		lines, err := unreachable(tool, dir, goos)
		if err != nil {
			t.Fatalf("GOOS=%s: %v", goos, err)
		}
		if len(lines) != len(want) {
			t.Fatalf("GOOS=%s lists %d function(s), want %d (%v):\n%s", goos, len(lines), len(want), want, strings.Join(lines, "\n"))
		}
		for i, name := range want {
			if !strings.HasSuffix(lines[i], "unreachable func: "+name) {
				t.Errorf("GOOS=%s line %d = %q, want it to name %s", goos, i, lines[i], name)
			}
		}
	}
}

// A MODULE THE TOOL CANNOT LOAD IS NOT A CLEAN MODULE: the analysis prints no findings for it
// either, so the failure has to arrive as an error and never as an empty list.
func TestAModuleThatDoesNotLoadIsAnErrorNotAnEmptyList(t *testing.T) {
	tool, err := pinnedTool("..")
	if err != nil {
		t.Fatal(err)
	}
	dir := write(t, map[string]string{
		"go.mod":           goMod,
		"cmd/tool/main.go": "package main\n\nfunc main() { undefined() }\n",
	})
	if lines, err := unreachable(tool, dir, "linux"); err == nil {
		t.Fatalf("a module that does not compile was read as %d finding(s) and no error", len(lines))
	}
	if lines, err := unreachable(tool, write(t, map[string]string{"go.mod": goMod, "lib/lib.go": "package lib\n\nfunc F() {}\n"}), "linux"); err == nil {
		t.Fatalf("a module with no main package was read as %d finding(s) and no error — nothing roots the analysis", len(lines))
	}
}
