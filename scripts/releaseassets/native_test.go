package main

import (
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
)

var (
	nativeSetOnce sync.Once
	nativeSetVal  NativeSet
	nativeSetErr  error
)

// repoNativeSet is the native set of this repository, derived once per test binary.
func repoNativeSet(t *testing.T) NativeSet {
	t.Helper()
	nativeSetOnce.Do(func() {
		root, err := filepath.Abs(filepath.Join("..", ".."))
		if err != nil {
			nativeSetErr = err
			return
		}
		nativeSetVal, nativeSetErr = DeriveNativeSet(root)
	})
	if nativeSetErr != nil {
		t.Fatal(nativeSetErr)
	}
	return nativeSetVal
}

// TestTheNativeSetIsDerived: every package that imports internal/nativehelper is in the set
// because it imports it, and a path outside it is not.
func TestTheNativeSetIsDerived(t *testing.T) {
	s := repoNativeSet(t)
	for _, dir := range []string{
		feovTools + "/internal/tessocr",
		feovTools + "/internal/fetchcache",
		feovTools + "/internal/cli",
		feovTools + "/internal/seatprobe",
		feovTools + "/internal/setup",
	} {
		if !slices.Contains(s.Dirs, dir) {
			t.Errorf("%s imports internal/nativehelper and is not in the derived set %v", dir, s.Dirs)
		}
	}
	for p, want := range map[string]bool{
		feovTools + "/internal/cli/fetch.go":                              true,
		feovTools + "/internal/fetchcache/pagerender.go":                  true,
		"plugins/frank-exchange-of-views/native/frame/frame.cpp":          true,
		".github/actions/install-helpers/action.yml":                      true,
		".github/workflows/hooks.yml":                                     true,
		"scripts/staticproof/main.go":                                     true,
		feovTools + "/internal/record/record.go":                          false,
		"README.md":                                                       false,
		".github/actions/go-cache/action.yml":                             false,
		feovTools + "/internal/tessocr/testdata/corpus/nbs602-form/x.txt": false,
	} {
		if got := s.Contains(p); got != want {
			t.Errorf("the native set holds %s: %v, want %v", p, got, want)
		}
	}
}

// mustPassNames is native-run's NATIVE_RUN_MUST_PASS and NATIVE_RUN_MUST_PASS_LINUX, read from the
// job's env in hooks.yml.
func mustPassNames(t *testing.T) []string {
	t.Helper()
	jobs, err := jobsOf(readHooks(t))
	if err != nil {
		t.Fatal(err)
	}
	env := mapValue(mapValue(jobs, "native-run"), "env")
	var names []string
	for _, key := range []string{"NATIVE_RUN_MUST_PASS", "NATIVE_RUN_MUST_PASS_LINUX"} {
		v := mapValue(env, key)
		if v == nil {
			t.Fatalf("native-run's env has no %s", key)
		}
		names = append(names, strings.Fields(v.Value)...)
	}
	if len(names) == 0 {
		t.Fatal("native-run's must-pass list is empty")
	}
	return names
}

// TestEveryMustPassNameIsATest: a name on native-run's list that no package native-run executes
// declares as `func Name(t *testing.T)` can never report PASS, so every full native-run would fail
// on it. The packages are RequirePackages', the set the job runs.
func TestEveryMustPassNameIsATest(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	pkgs, err := RequirePackages(root)
	if err != nil {
		t.Fatal(err)
	}
	funcs, err := TestFuncs(root, pkgs)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range mustPassNames(t) {
		if _, ok := funcs[name]; !ok {
			t.Errorf("native-run must pass %s, and no package it runs (%v) declares func %s(t *testing.T)", name, pkgs, name)
		}
	}
	// The control: a name nothing declares is refused by the same lookup.
	if _, ok := funcs["TestNoSuchTestAnywhere"]; ok {
		t.Fatal("the lookup finds a test nobody wrote")
	}
}

// TestNativeRunRunsEveryRequirePackage: the packages calling nativehelper.Require are found by
// parsing their calls, and native-run and its servercore container take the list from
// -testpackages rather than naming packages.
func TestNativeRunRunsEveryRequirePackage(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	pkgs, err := RequirePackages(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"./internal/cli", "./internal/fetchcache", "./internal/nativehelper", "./internal/seatprobe", "./internal/tessocr"} {
		if !slices.Contains(pkgs, want) {
			t.Errorf("RequirePackages = %v, missing %s", pkgs, want)
		}
	}
	jobs, err := jobsOf(readHooks(t))
	if err != nil {
		t.Fatal(err)
	}
	steps := mapValue(mapValue(jobs, "native-run"), "steps")
	uses := 0
	for _, step := range steps.Content {
		run := mapValue(step, "run")
		if run == nil {
			continue
		}
		if strings.Contains(run.Value, "go test") && strings.Contains(run.Value, "./internal/") {
			t.Errorf("a native-run step names its packages instead of reading -testpackages:\n%s", run.Value)
		}
		if strings.Contains(run.Value, "./releaseassets -testpackages") {
			uses++
		}
	}
	if uses != 2 {
		t.Errorf("%d native-run steps read -testpackages, want 2: the host run and the servercore container", uses)
	}
}
