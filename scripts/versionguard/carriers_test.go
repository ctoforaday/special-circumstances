package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// The guard's own claim is that a version fact cannot disagree with itself unnoticed. These
// tests hold the three things that claim rests on: carriers are compared, a carrier that stops
// matching is an ERROR rather than a pass, and a new version-shaped declaration must be
// classified by a person.

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

func TestCarriersAgreeIsClean(t *testing.T) {
	if p := checkCarriers(scratchTree(t, "0.68.0", "0.68.0")); len(p) != 0 {
		t.Errorf("agreeing carriers reported problems: %v", p)
	}
}

// THE POINT OF THE FILE.
func TestDisagreeingCarriersAreReported(t *testing.T) {
	p := checkCarriers(scratchTree(t, "0.68.0", "0.67.0"))
	if len(p) != 1 {
		t.Fatalf("got %d problems, want 1: %v", len(p), p)
	}
	for _, want := range []string{"DISAGREE", "0.68.0", "0.67.0", "root.go", "requirements.json"} {
		if !strings.Contains(p[0], want) {
			t.Errorf("the message must name %q so it can be acted on:\n%s", want, p[0])
		}
	}
}

// A carrier whose declaration moved is NOT "no problem found". This is the failure mode the
// whole guard exists to prevent, reappearing inside the guard.
func TestACarrierThatStopsMatchingIsAnError(t *testing.T) {
	dir := scratchTree(t, "0.68.0", "0.68.0")
	// The constant is renamed — the file still exists and still parses.
	writeAt(t, dir, "plugins/frank-exchange-of-views/tools/internal/cli/root.go",
		"package cli\n\nconst ToolRelease = \"0.68.0\"\n")
	p := checkCarriers(dir)
	if len(p) == 0 {
		t.Fatal("a carrier that no longer matches must be an ERROR — silently checking one " +
			"carrier against itself is a guard that passes because it stopped looking")
	}
	if !strings.Contains(p[0], "no longer being checked") {
		t.Errorf("the message must say the carrier is unchecked, not merely absent:\n%s", p[0])
	}
}

func TestAMissingCarrierFileIsAnError(t *testing.T) {
	dir := scratchTree(t, "0.68.0", "0.68.0")
	if err := os.Remove(filepath.Join(dir, "plugins/frank-exchange-of-views/requirements.json")); err != nil {
		t.Fatal(err)
	}
	if p := checkCarriers(dir); len(p) == 0 {
		t.Error("a deleted carrier must be reported, not treated as agreement")
	}
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

// The declared carrier itself must not be flagged by the sweep — it is already checked.
func TestSweepDoesNotDoubleReportADeclaredCarrier(t *testing.T) {
	if p := sweepUnclassified(scratchTree(t, "0.68.0", "0.68.0")); len(p) != 0 {
		t.Errorf("the declared carrier is checked by checkCarriers and must not also be swept: %v", p)
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
