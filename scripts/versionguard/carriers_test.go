package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// The guard's claim is narrow: a KNOWN class of stale per-binary version constants cannot
// grow. These tests hold the two things that rests on — the sweep sees the declaration forms
// that actually occur, and a new one must be classified by a person rather than defaulted into
// invisibility.
//
// The carrier-agreement tests that used to live here went with the machinery they covered:
// guarding two carriers that already agreed policed a thing that was not broken. Generation is
// the fix for that (#405), not a guard.

func writeAt(t *testing.T, root, rel, body string) {
	t.Helper()
	p := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// scratchTree builds a tree with the two real carriers of the tool-version fact.
func scratchTree(t *testing.T, goVersion, reqVersion string) string {
	t.Helper()
	dir := t.TempDir()
	writeAt(t, dir, "plugins/frank-exchange-of-views/tools/internal/cli/root.go",
		"package cli\n\nconst Version = \""+goVersion+"\"\n")
	writeAt(t, dir, "plugins/frank-exchange-of-views/requirements.json",
		"{\n  \"recordToolVersion\": \""+reqVersion+"\"\n}\n")
	return dir
}

// ---- the sweep ----

func TestSweepDemandsAnUnclassifiedDeclarationBeClassified(t *testing.T) {
	dir := scratchTree(t, "0.68.0", "0.68.0")
	writeAt(t, dir, "plugins/frank-exchange-of-views/tools/internal/thing/thing.go",
		"package thing\n\nconst ShippedVersion = \"9.9.9\"\n")
	p := sweepUnclassified(dir)
	if len(p) != 1 {
		t.Fatalf("got %d, want 1: %v", len(p), p)
	}
	if !strings.Contains(p[0], "ShippedVersion") || !strings.Contains(p[0], "CLASSIFIED NOWHERE") {
		t.Errorf("message:\n%s", p[0])
	}
}

// THE BLIND SPOT THAT SHIPPED IN THE FIRST DRAFT. Requiring the const/var keyword made every
// name inside a grouped block invisible, and grouped blocks are the idiomatic form.
func TestSweepSeesDeclarationsInsideGroupedBlocks(t *testing.T) {
	dir := scratchTree(t, "0.68.0", "0.68.0")
	writeAt(t, dir, "plugins/frank-exchange-of-views/tools/internal/thing/thing.go",
		"package thing\n\nconst (\n\tsomethingElse  = \"x\"\n\tgroupedVersion = \"1.2.3\"\n)\n")
	p := sweepUnclassified(dir)
	if len(p) != 1 {
		t.Fatalf("a declaration inside a grouped const block must be seen; got %d: %v", len(p), p)
	}
	if !strings.Contains(p[0], "groupedVersion") {
		t.Errorf("message:\n%s", p[0])
	}
}

func TestSweepHonoursAnExplicitDenial(t *testing.T) {
	dir := scratchTree(t, "0.68.0", "0.68.0")
	writeAt(t, dir, "plugins/frank-exchange-of-views/tools/internal/thing/thing.go",
		"package thing\n\nconst MassMappingVersion = \"v1\"\n")
	if p := sweepUnclassified(dir); len(p) != 0 {
		t.Errorf("an explicitly denied name must not be flagged: %v", p)
	}
}

func TestSweepIgnoresTestFiles(t *testing.T) {
	dir := scratchTree(t, "0.68.0", "0.68.0")
	writeAt(t, dir, "plugins/frank-exchange-of-views/tools/internal/thing/thing_test.go",
		"package thing\n\nconst FixtureVersion = \"1.0\"\n")
	if p := sweepUnclassified(dir); len(p) != 0 {
		t.Errorf("a fixture in a test file is not a shipped version: %v", p)
	}
}

// A qualified assignment is not a declaration; flagging it would train people to pad the
// denial list, which is how an allowlist stops meaning anything.
func TestSweepIgnoresQualifiedAssignment(t *testing.T) {
	dir := scratchTree(t, "0.68.0", "0.68.0")
	writeAt(t, dir, "plugins/frank-exchange-of-views/tools/internal/thing/thing.go",
		"package thing\n\nfunc init() { record.ToolVersion = \"x\" }\n")
	for _, p := range sweepUnclassified(dir) {
		if strings.Contains(p, "record.ToolVersion") {
			t.Errorf("a qualified assignment must not be treated as a declaration: %s", p)
		}
	}
}

// feov-record's own const Version is not a stale-at-birth constant — it moves, and
// requirements.json restates it. Derivation is the fix (#405); the sweep must not drag it into
// the frozen-binary class.
func TestSweepLeavesTheFeovToolVersionAlone(t *testing.T) {
	if p := sweepUnclassified(scratchTree(t, "0.68.0", "0.68.0")); len(p) != 0 {
		t.Errorf("feov-record's own version must not be swept into the stale class: %v", p)
	}
}

// The pattern is the reach of the whole sweep, so pin what it must catch.
func TestVersionShapedPatternCatchesTheFormsThatOccur(t *testing.T) {
	for _, src := range []string{
		`const Version = "1"`,
		`var ToolVersion = "1"`,
		"const (\n\tfetchUAVersion = \"1.0\"\n)",
		`	schemaVersion = "3"`,
	} {
		if !regexp.MustCompile(reVersionShaped.String()).MatchString(src) {
			t.Errorf("pattern missed a real declaration form: %q", src)
		}
	}
}
