package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// stageRoot lays out a repository root holding only a registry and the two output directories.
func stageRoot(t *testing.T, registryJSON string) string {
	t.Helper()
	root := t.TempDir()
	for _, rel := range []string{registryPath, classesPath, shippedPath} {
		if err := os.MkdirAll(filepath.Dir(filepath.Join(root, rel)), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, registryPath), []byte(registryJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestClassgenRefusesRowWithoutMaterialDefault(t *testing.T) {
	_, err := generate([]byte(`{"classes":[{"slug":"a","material_default":"always"},{"slug":"hand-added"}]}`))
	if err == nil {
		t.Fatal("a row with no material_default generated — a hand-added row without it would reach setup, which refuses it on the operator's machine")
	}
	if !strings.Contains(err.Error(), `"hand-added"`) || !strings.Contains(err.Error(), "material_default") {
		t.Errorf("the refusal must name the row and the field: %v", err)
	}
}

func TestClassgenRefusesUnknownMaterialDefault(t *testing.T) {
	_, err := generate([]byte(`{"classes":[{"slug":"a","material_default":"sometimes"}]}`))
	if err == nil {
		t.Fatal("an unknown material_default generated")
	}
	if !strings.Contains(err.Error(), `"a"`) || !strings.Contains(err.Error(), `"sometimes"`) {
		t.Errorf("the refusal must name the row and the value: %v", err)
	}
}

// -check has to cover the shipped materiality table as well as the fixture vocabulary: a
// hand-edited value in shippedclasses_gen.go would otherwise change how every archive migrates
// with the gate green.
func TestClassgenCheckCoversShippedTable(t *testing.T) {
	root := stageRoot(t, `{"classes":[{"slug":"a","material_default":"always"},{"slug":"b","material_default":"by_grade"}]}`)
	if _, err := run(root, false); err != nil {
		t.Fatal(err)
	}
	if _, err := run(root, true); err != nil {
		t.Fatalf("a freshly generated tree must pass -check: %v", err)
	}
	p := filepath.Join(root, shippedPath)
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	edited := strings.Replace(string(b), "CLASS_MATERIAL_ALWAYS", "CLASS_MATERIAL_NEVER", 1)
	if edited == string(b) {
		t.Fatal("fixture: the shipped table carries no `always` value to edit")
	}
	if err := os.WriteFile(p, []byte(edited), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err = run(root, true)
	if err == nil || !strings.Contains(err.Error(), filepath.Base(shippedPath)) {
		t.Fatalf("-check passed a hand-edited shipped table: %v", err)
	}
}
