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
	nativeSet := repoNativeSet(t)

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
			name:       "pull request changing a package that imports nativehelper",
			event:      "pull_request",
			ref:        "refs/pull/123/merge",
			changed:    []string{"plugins/frank-exchange-of-views/tools/internal/cli/fetch.go"},
			wantCache:  "restore",
			wantFull:   true,
			needTarget: "linux-amd64",
		},
		{
			name:       "pull request changing the helper install action",
			event:      "pull_request",
			ref:        "refs/pull/123/merge",
			changed:    []string{".github/actions/install-helpers/action.yml"},
			wantCache:  "restore",
			wantFull:   true,
			needTarget: "linux-amd64",
		},
		{
			name:       "pull request changing a fetchcache file outside the renderer",
			event:      "pull_request",
			ref:        "refs/pull/123/merge",
			changed:    []string{"plugins/frank-exchange-of-views/tools/internal/fetchcache/pdfextract.go"},
			wantCache:  "restore",
			wantFull:   true,
			needTarget: "linux-amd64",
		},
		{
			name:       "pull request changing a subpackage of a nativehelper importer",
			event:      "pull_request",
			ref:        "refs/pull/123/merge",
			changed:    []string{"plugins/frank-exchange-of-views/tools/internal/cli/lens/close.go"},
			wantCache:  "restore",
			wantFull:   false,
			needTarget: "linux-amd64",
		},
		{
			name:       "pull request changing a corpus golden",
			event:      "pull_request",
			ref:        "refs/pull/123/merge",
			changed:    []string{"plugins/frank-exchange-of-views/tools/internal/tessocr/testdata/corpus/nbs602-form/reading.darwin-arm64.golden"},
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
			scope, err := DecideScope(tc.event, tc.ref, tc.changed, nativeSet, targets, consumers, helpers)
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
	_, err = DecideScope("unknown_event", "refs/heads/main", nil, nativeSet, targets, consumers, helpers)
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

// TestAMainPushBuildsWhatFeovRecordInstalls: feov-record runs on Linux and Windows and installs
// every helper for its runner's target, and a push to main builds only the targets a consumer's
// download names. So main's build matrix holds every helper for linux-amd64 and windows-amd64.
func TestAMainPushBuildsWhatFeovRecordInstalls(t *testing.T) {
	wfPath := filepath.Join("..", "..", ".github", "workflows", "hooks.yml")
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
	helpers, err := FindHelpers(filepath.Join("..", "..", "plugins", "frank-exchange-of-views", "native"))
	if err != nil {
		t.Fatal(err)
	}
	scope, err := DecideScope("push", "refs/heads/main", nil, repoNativeSet(t), targets, consumers, helpers)
	if err != nil {
		t.Fatal(err)
	}
	built := map[string]bool{}
	for _, leg := range scope.Matrix.Include {
		built[leg.Helper+" "+leg.Target] = true
	}
	for _, h := range helpers {
		for _, target := range []string{"linux-amd64", "windows-amd64"} {
			if !built[h+" "+target] {
				t.Errorf("a push to main builds no %s for %s, which feov-record installs there; matrix %v", h, target, scope.Matrix.Include)
			}
		}
	}
}

// TestADownloadNamingNoTargetIsRefused: a consumer's native download is read for its target, a
// wildcard helper included; one whose target is an expression names no target and is refused,
// never read as no download at all.
func TestADownloadNamingNoTargetIsRefused(t *testing.T) {
	wf := func(pattern string) string {
		p := filepath.Join(t.TempDir(), "hooks.yml")
		body := `
jobs:
  feov-record:
    needs: [native-build]
    steps:
      - uses: actions/download-artifact@v4
        with:
          pattern: ` + pattern + `
`
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		return p
	}
	got, err := DeriveConsumerTargets(wf("native-*-windows-amd64"))
	if err != nil || !slices.Equal(got, []string{"windows-amd64"}) {
		t.Fatalf("a wildcard helper pattern: got %v, %v; want [windows-amd64]", got, err)
	}
	_, err = DeriveConsumerTargets(wf(`"native-*-${{ runner.os == 'Windows' && 'windows' || 'linux' }}-amd64"`))
	if err == nil || !strings.Contains(err.Error(), "feov-record") {
		t.Fatalf("an expression in place of the target was not refused naming its job: %v", err)
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
