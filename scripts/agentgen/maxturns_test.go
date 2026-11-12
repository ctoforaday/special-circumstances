package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/scripts/internal/gitx"
)

// frontmatterValue is the value of a top-level key in a definition's frontmatter, and whether the
// block states it exactly once.
func frontmatterValue(t *testing.T, def []byte, key string) (string, bool) {
	t.Helper()
	if !bytes.HasPrefix(def, []byte("---\n")) {
		t.Fatal("the definition opens with no frontmatter block")
	}
	block, _, closed := strings.Cut(string(def[len("---\n"):]), "\n---\n")
	if !closed {
		t.Fatal("the definition's frontmatter block never closes")
	}
	var values []string
	for _, line := range strings.Split(block, "\n") {
		if v, ok := strings.CutPrefix(line, key+": "); ok {
			values = append(values, v)
		}
	}
	if len(values) != 1 {
		return strings.Join(values, ","), false
	}
	return values[0], true
}

// EVERY DEFINITION THE GENERATOR RENDERS CARRIES maxTurns = THE CONSTANT, once, in its frontmatter.
// Rendered from the shipped sources through the same function main writes with, so a generator
// that stops writing it fails here.
func TestEveryRenderedDefinitionCarriesTheTurnCap(t *testing.T) {
	for name := range surfaceOf {
		out, err := renderDefinition("src", name, []byte("surface\n"))
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if v, once := frontmatterValue(t, out, "maxTurns"); !once || v != strconv.Itoa(maxTurns) {
			t.Errorf("%s renders maxTurns %q (once: %v), want %d", name, v, once, maxTurns)
		}
	}
}

// EVERY COMMITTED DEFINITION CARRIES IT TOO: the files the client loads are the ones that bound a
// sitting, and the plugin's agents directory holds exactly the definitions agentgen writes.
func TestEveryCommittedDefinitionCarriesTheTurnCap(t *testing.T) {
	root, err := gitx.Root()
	if err != nil {
		t.Fatal(err)
	}
	found, err := filepath.Glob(filepath.Join(root, agentsDir, "*.md"))
	if err != nil {
		t.Fatal(err)
	}
	if len(found) != len(surfaceOf) {
		t.Fatalf("%s holds %d definition(s) and agentgen writes %d", agentsDir, len(found), len(surfaceOf))
	}
	for _, p := range found {
		def, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		if v, once := frontmatterValue(t, def, "maxTurns"); !once || v != strconv.Itoa(maxTurns) {
			t.Errorf("%s carries maxTurns %q (once: %v), want %d — regenerate: go -C scripts run ./agentgen", filepath.Base(p), v, once, maxTurns)
		}
	}
}

// A SOURCE THAT STATES ITS OWN maxTurns IS REFUSED: the constant is the one place the value is set.
func TestASourceStatingItsOwnTurnCapIsRefused(t *testing.T) {
	root := tree(t, map[string]string{"seat.md": "---\nname: seat\nmaxTurns: 9\n---\n\nbody\n"})
	if _, err := renderDefinition(root, "seat", []byte("surface\n")); err == nil || !strings.Contains(err.Error(), "maxTurns") {
		t.Errorf("a source stating maxTurns rendered: %v", err)
	}
}
