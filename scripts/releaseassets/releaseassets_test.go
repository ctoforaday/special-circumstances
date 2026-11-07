package main

import (
	"bytes"
	"debug/elf"
	"encoding/binary"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// makeDynamicELF writes a minimal 64-bit little-endian ELF with a PT_DYNAMIC header.
func makeDynamicELF(t *testing.T, path string) {
	t.Helper()
	var b bytes.Buffer
	// ELF Header (64 bytes)
	hdr := elf.Header64{
		Ident:     [16]byte{0x7f, 'E', 'L', 'F', 2, 1, 1, 0},
		Type:      2,  // ET_EXEC
		Machine:   62, // EM_X86_64
		Version:   1,
		Entry:     0x400000,
		Phoff:     64, // Program header offset right after ELF header
		Shoff:     0,
		Flags:     0,
		Ehsize:    64,
		Phentsize: 56, // sizeof(Elf64_Phdr)
		Phnum:     2,
		Shentsize: 64,
		Shnum:     0,
		Shstrndx:  0,
	}
	if err := binary.Write(&b, binary.LittleEndian, &hdr); err != nil {
		t.Fatal(err)
	}
	// Phdr 0: PT_LOAD
	pLoad := elf.Prog64{
		Type:   1, // PT_LOAD
		Flags:  5,
		Off:    0,
		Vaddr:  0x400000,
		Paddr:  0x400000,
		Filesz: 0x1000,
		Memsz:  0x1000,
		Align:  0x1000,
	}
	if err := binary.Write(&b, binary.LittleEndian, &pLoad); err != nil {
		t.Fatal(err)
	}
	// Phdr 1: PT_DYNAMIC
	pDyn := elf.Prog64{
		Type:   2, // PT_DYNAMIC
		Flags:  6,
		Off:    0x1000,
		Vaddr:  0x401000,
		Paddr:  0x401000,
		Filesz: 0x100,
		Memsz:  0x100,
		Align:  8,
	}
	if err := binary.Write(&b, binary.LittleEndian, &pDyn); err != nil {
		t.Fatal(err)
	}
	// Pad to 0x1000
	pad := make([]byte, 0x1000-b.Len())
	b.Write(pad)
	// Dynamic section entries (DT_NULL)
	b.Write(make([]byte, 0x100))
	if err := os.WriteFile(path, b.Bytes(), 0o755); err != nil {
		t.Fatal(err)
	}
}

// makeStaticELF writes a minimal static 64-bit ELF with only PT_LOAD.
func makeStaticELF(t *testing.T, path string) {
	t.Helper()
	var b bytes.Buffer
	hdr := elf.Header64{
		Ident:     [16]byte{0x7f, 'E', 'L', 'F', 2, 1, 1, 0},
		Type:      2,
		Machine:   62,
		Version:   1,
		Entry:     0x400000,
		Phoff:     64,
		Ehsize:    64,
		Phentsize: 56,
		Phnum:     1,
	}
	if err := binary.Write(&b, binary.LittleEndian, &hdr); err != nil {
		t.Fatal(err)
	}
	pLoad := elf.Prog64{
		Type:   1, // PT_LOAD
		Flags:  5,
		Off:    0,
		Vaddr:  0x400000,
		Paddr:  0x400000,
		Filesz: 0x100,
		Memsz:  0x100,
		Align:  0x1000,
	}
	if err := binary.Write(&b, binary.LittleEndian, &pLoad); err != nil {
		t.Fatal(err)
	}
	pad := make([]byte, 0x100-b.Len())
	b.Write(pad)
	if err := os.WriteFile(path, b.Bytes(), 0o755); err != nil {
		t.Fatal(err)
	}
}

func TestADynamicELFIsRefused(t *testing.T) {
	dir := t.TempDir()
	badELF := filepath.Join(dir, "feov-record_linux_amd64")
	makeDynamicELF(t, badELF)

	err := CheckDist(dir, "refs/tags/frank-exchange-of-views--v0.1.0", nil, nil)
	if err == nil {
		t.Fatal("CheckDist passed a dynamic ELF, want refusal")
	}
	if !strings.Contains(err.Error(), "not statically linked") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEveryExecutableIsProven(t *testing.T) {
	dir := t.TempDir()
	goodELF := filepath.Join(dir, "feov-record_linux_amd64")
	makeStaticELF(t, goodELF)

	targets := []Target{
		{OS: "linux", Arch: "amd64", Name: "linux-amd64"},
	}
	err := CheckDist(dir, "refs/tags/gray-area--v0.1.0", targets, nil)
	if err != nil {
		t.Fatalf("CheckDist failed on proven static binary: %v", err)
	}
}

func TestAFEOVDistMissingOneHelperIsRefused(t *testing.T) {
	dir := t.TempDir()
	makeStaticELF(t, filepath.Join(dir, "feov-record_linux_amd64"))
	makeStaticELF(t, filepath.Join(dir, "feov-ocr_linux_amd64"))
	// Intentionally missing feov-ocr_windows_arm64.exe

	targets := []Target{
		{OS: "linux", Arch: "amd64", Name: "linux-amd64"},
		{OS: "windows", Arch: "arm64", Name: "windows-arm64", Ext: ".exe"},
	}
	helpers := []string{"feov-ocr"}

	err := CheckDist(dir, "refs/tags/frank-exchange-of-views--v0.1.0", targets, helpers)
	if err == nil {
		t.Fatal("CheckDist passed a FEOV dist missing a helper asset")
	}
	if !strings.Contains(err.Error(), "feov-ocr_windows_arm64.exe") {
		t.Fatalf("error %q does not name the missing helper asset", err)
	}
}

func TestADistWithoutHelpersNeedsNone(t *testing.T) {
	dir := t.TempDir()
	makeStaticELF(t, filepath.Join(dir, "gray-area_linux_amd64"))

	targets := []Target{
		{OS: "linux", Arch: "amd64", Name: "linux-amd64"},
		{OS: "windows", Arch: "arm64", Name: "windows-arm64", Ext: ".exe"},
	}
	helpers := []string{"feov-ocr"}

	// gray-area tag has no helpers
	err := CheckDist(dir, "refs/tags/gray-area--v0.1.0", targets, helpers)
	if err != nil {
		t.Fatalf("CheckDist demanded helpers of non-FEOV plugin: %v", err)
	}
}

func TestATargetWithoutARunnerIsRefused(t *testing.T) {
	tmpWF := filepath.Join(t.TempDir(), "hooks.yml")
	content := `
jobs:
  release:
    steps:
      - run: |
          for pair in windows/amd64 darwin/amd64 linux/amd64; do
            echo $pair
          done
`
	if err := os.WriteFile(tmpWF, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	pairs, err := ReadReleasePairs(tmpWF)
	if err != nil {
		t.Fatalf("ReadReleasePairs: %v", err)
	}
	_, err = ParseTargets(pairs)
	if err == nil {
		t.Fatal("ParseTargets passed darwin/amd64 which has no runner")
	}
	if !strings.Contains(err.Error(), "darwin/amd64") {
		t.Fatalf("error %q did not name darwin/amd64", err)
	}
}

func TestScopeBuildsEveryConsumersTarget(t *testing.T) {
	wfPath := filepath.Join("..", "..", ".github", "workflows", "hooks.yml")
	if _, err := os.Stat(wfPath); err != nil {
		t.Skipf("workflow file %s not found: %v", wfPath, err)
	}
	b, err := os.ReadFile(wfPath)
	if err != nil {
		t.Fatal(err)
	}

	var wfDoc struct {
		On map[string]any `yaml:"on"`
	}
	if err := yaml.Unmarshal(b, &wfDoc); err != nil {
		t.Fatalf("unmarshaling workflow: %v", err)
	}

	pairs, err := ReadReleasePairs(wfPath)
	if err != nil {
		t.Fatal(err)
	}
	targets, err := ParseTargets(pairs)
	if err != nil {
		t.Fatal(err)
	}
	consumers, err := DeriveConsumerTargets(wfPath)
	if err != nil {
		t.Fatal(err)
	}
	helpers := []string{"feov-ocr"}

	// Must derive consumers dynamically
	if len(consumers) == 0 {
		t.Fatal("no consumer targets derived from workflow")
	}

	testCases := []struct {
		name       string
		event      string
		ref        string
		changed    []string
		wantCache  string
		wantFull   bool
		needTarget string
	}{
		{
			name:       "pull request with native change",
			event:      "pull_request",
			ref:        "refs/pull/123/merge",
			changed:    []string{"plugins/frank-exchange-of-views/native/feov-ocr/main.cpp"},
			wantCache:  "restore",
			wantFull:   true,
			needTarget: "linux-amd64",
		},
		{
			name:       "pull request without native change",
			event:      "pull_request",
			ref:        "refs/pull/123/merge",
			changed:    []string{"README.md"},
			wantCache:  "restore",
			wantFull:   false,
			needTarget: "linux-amd64",
		},
		{
			name:       "push to main",
			event:      "push",
			ref:        "refs/heads/main",
			wantCache:  "restore",
			wantFull:   false,
			needTarget: "linux-amd64",
		},
		{
			name:       "push to FEOV tag",
			event:      "push",
			ref:        "refs/tags/frank-exchange-of-views--v0.2.0",
			wantCache:  "none",
			wantFull:   true,
			needTarget: "linux-amd64",
		},
		{
			name:       "push to other tag",
			event:      "push",
			ref:        "refs/tags/gray-area--v0.2.0",
			wantCache:  "none",
			wantFull:   false,
			needTarget: "linux-amd64",
		},
		{
			name:       "schedule",
			event:      "schedule",
			ref:        "refs/heads/main",
			wantCache:  "save",
			wantFull:   true,
			needTarget: "linux-amd64",
		},
		{
			name:       "dispatch on main",
			event:      "workflow_dispatch",
			ref:        "refs/heads/main",
			wantCache:  "save",
			wantFull:   true,
			needTarget: "linux-amd64",
		},
		{
			name:       "dispatch on other branch",
			event:      "workflow_dispatch",
			ref:        "refs/heads/feature",
			wantCache:  "restore",
			wantFull:   true,
			needTarget: "linux-amd64",
		},
		{
			name:       "dispatch on FEOV tag",
			event:      "workflow_dispatch",
			ref:        "refs/tags/frank-exchange-of-views--v0.2.0",
			wantCache:  "none",
			wantFull:   true,
			needTarget: "linux-amd64",
		},
		{
			name:       "dispatch on other tag",
			event:      "workflow_dispatch",
			ref:        "refs/tags/gray-area--v0.2.0",
			wantCache:  "none",
			wantFull:   false,
			needTarget: "linux-amd64",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			scope, err := DecideScope(tc.event, tc.ref, tc.changed, targets, consumers, helpers)
			if err != nil {
				t.Fatalf("DecideScope failed: %v", err)
			}
			if scope.Cache != tc.wantCache {
				t.Errorf("cache = %s, want %s", scope.Cache, tc.wantCache)
			}
			if scope.Full != tc.wantFull {
				t.Errorf("full = %v, want %v", scope.Full, tc.wantFull)
			}
			if len(scope.Matrix.Include) == 0 {
				t.Fatal("empty native-build matrix")
			}
			// Verify every consumer target is in the matrix
			matrixTargets := map[string]bool{}
			for _, leg := range scope.Matrix.Include {
				matrixTargets[leg.Target] = true
			}
			for _, ct := range consumers {
				if !matrixTargets[ct] {
					t.Errorf("matrix missing consumer target %s", ct)
				}
			}
		})
	}

	// Deletion check 1: unknown event in on: fails
	_, err = DecideScope("unknown_event", "refs/heads/main", nil, targets, consumers, helpers)
	if err == nil {
		t.Fatal("DecideScope accepted unknown event")
	}

	// Deletion check 2: fixture workflow where feov-record downloads windows-arm64
	// must derive windows-arm64 as a consumer
	fixtureWF := filepath.Join(t.TempDir(), "fixture_hooks.yml")
	fContent := `
jobs:
  release:
    steps:
      - run: |
          for pair in windows/amd64 windows/arm64 darwin/arm64 linux/amd64 linux/arm64; do
            echo $pair
          done
  feov-record:
    needs: [native-build]
    steps:
      - uses: actions/download-artifact@v4
        with:
          name: native-feov-ocr-windows-arm64
`
	if err := os.WriteFile(fixtureWF, []byte(fContent), 0o644); err != nil {
		t.Fatal(err)
	}
	derived, err := DeriveConsumerTargets(fixtureWF)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(derived, "windows-arm64") {
		t.Fatalf("DeriveConsumerTargets did not find windows-arm64 in fixture: %v", derived)
	}
}

func TestTheHelperCacheStepsObeyTheScope(t *testing.T) {
	wfPath := filepath.Join("..", "..", ".github", "workflows", "hooks.yml")
	if _, err := os.Stat(wfPath); err != nil {
		t.Skipf("workflow file %s not found: %v", wfPath, err)
	}
	b, err := os.ReadFile(wfPath)
	if err != nil {
		t.Fatal(err)
	}

	var root yaml.Node
	if err := yaml.Unmarshal(b, &root); err != nil {
		t.Fatal(err)
	}

	// Walk every node looking for cache actions
	checkCacheSteps(t, &root, "hooks.yml")
}

func checkCacheSteps(t *testing.T, node *yaml.Node, path string) {
	t.Helper()
	if node.Kind == yaml.MappingNode {
		for i := 0; i < len(node.Content); i += 2 {
			k := node.Content[i].Value
			if k == "jobs" {
				checkJobs(t, node.Content[i+1])
			}
		}
	}
}

func checkJobs(t *testing.T, jobsNode *yaml.Node) {
	t.Helper()
	for i := 0; i < len(jobsNode.Content); i += 2 {
		jobName := jobsNode.Content[i].Value
		jobBody := jobsNode.Content[i+1]
		for j := 0; j < len(jobBody.Content); j += 2 {
			if jobBody.Content[j].Value == "steps" {
				stepsNode := jobBody.Content[j+1]
				for _, step := range stepsNode.Content {
					checkStep(t, jobName, step)
				}
			}
		}
	}
}

func checkStep(t *testing.T, jobName string, stepNode *yaml.Node) {
	t.Helper()
	var uses, ifCond string
	var hasRestoreKeys bool
	for i := 0; i < len(stepNode.Content); i += 2 {
		k := stepNode.Content[i].Value
		v := stepNode.Content[i+1].Value
		if k == "uses" {
			uses = v
		} else if k == "if" {
			ifCond = v
		} else if k == "with" {
			withNode := stepNode.Content[i+1]
			for w := 0; w < len(withNode.Content); w += 2 {
				if withNode.Content[w].Value == "restore-keys" {
					hasRestoreKeys = true
				}
			}
		}
	}

	if strings.HasPrefix(uses, "actions/cache/restore") {
		if jobName != "native-build" {
			t.Errorf("job %s uses actions/cache/restore, only native-build may use it", jobName)
		}
		wantIf := "needs.native-scope.outputs.cache != 'none'"
		if strings.TrimSpace(ifCond) != wantIf {
			t.Errorf("restore step if: got %q, want %q", ifCond, wantIf)
		}
		if hasRestoreKeys {
			t.Errorf("restore step has restore-keys, must have none")
		}
	} else if strings.HasPrefix(uses, "actions/cache/save") {
		if jobName != "native-build" {
			t.Errorf("job %s uses actions/cache/save, only native-build may use it", jobName)
		}
		wantIf := "needs.native-scope.outputs.cache == 'save' && steps.helper-cache.outputs.cache-hit != 'true'"
		if strings.TrimSpace(ifCond) != wantIf {
			t.Errorf("save step if: got %q, want %q", ifCond, wantIf)
		}
	} else if strings.HasPrefix(uses, "actions/cache@") {
		t.Errorf("job %s uses combined actions/cache: must use actions/cache/restore or actions/cache/save", jobName)
	}
}

func TestNoNativeJobSkipsAtTheJobLevel(t *testing.T) {
	wfPath := filepath.Join("..", "..", ".github", "workflows", "hooks.yml")
	if _, err := os.Stat(wfPath); err != nil {
		t.Skipf("workflow file %s not found: %v", wfPath, err)
	}
	b, err := os.ReadFile(wfPath)
	if err != nil {
		t.Fatal(err)
	}

	var root yaml.Node
	if err := yaml.Unmarshal(b, &root); err != nil {
		t.Fatal(err)
	}

	var jobsNode *yaml.Node
	if root.Kind == yaml.DocumentNode && len(root.Content) > 0 {
		doc := root.Content[0]
		for i := 0; i < len(doc.Content); i += 2 {
			if doc.Content[i].Value == "jobs" {
				jobsNode = doc.Content[i+1]
				break
			}
		}
	}
	if jobsNode == nil {
		t.Fatal("no jobs node in workflow")
	}

	checkedJobs := map[string]bool{
		"native-scope": true,
		"native-build": true,
		"native-run":   true,
		"native-fuzz":  true,
		"ocr-corpus":   true,
		"feov-record":  true,
	}

	for i := 0; i < len(jobsNode.Content); i += 2 {
		name := jobsNode.Content[i].Value
		body := jobsNode.Content[i+1]
		if !checkedJobs[name] {
			continue
		}
		for j := 0; j < len(body.Content); j += 2 {
			if body.Content[j].Value == "if" {
				t.Errorf("job %s has a job-level if: %q — native jobs must have no job-level if", name, body.Content[j+1].Value)
			}
		}
	}
}
