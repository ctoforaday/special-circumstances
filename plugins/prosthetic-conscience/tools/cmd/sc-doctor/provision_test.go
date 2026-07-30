package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// newestVersionDir decides WHICH installed version of a sibling plugin the doctor
// reports and provisions. It had no direct test: a mutation sweep flipped `best == ""`
// to `best != ""` with the suite still green, because semverLess("", "0.7.0") happens to
// be true, so the wrong code selects a version anyway — just not necessarily the right
// one.
//
// The case that matters is 0.9 vs 0.10, where lexical ordering disagrees with semver.
// Getting it wrong means `--fix` fetches a release tag for a version the operator does
// not have installed.
func TestNewestVersionDir(t *testing.T) {
	cases := []struct {
		name string
		dirs []string
		want string
	}{
		{"single", []string{"0.7.0"}, "0.7.0"},
		{"ordered", []string{"0.7.0", "0.8.0"}, "0.8.0"},
		{"reverse order on disk", []string{"0.8.0", "0.7.0"}, "0.8.0"},
		// Lexically "0.10.0" < "0.9.0"; by semver it is newer. This is the whole
		// reason semverLess parses integers instead of comparing strings.
		{"double-digit minor beats single", []string{"0.9.0", "0.10.0"}, "0.10.0"},
		{"double-digit patch", []string{"0.9.9", "0.9.10"}, "0.9.10"},
		{"major dominates", []string{"0.99.0", "1.0.0"}, "1.0.0"},
		{"three-way", []string{"0.8.0", "0.10.1", "0.10.0"}, "0.10.1"},
		{"none", nil, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dir := t.TempDir()
			for _, d := range c.dirs {
				if err := os.MkdirAll(filepath.Join(dir, d), 0o755); err != nil {
					t.Fatal(err)
				}
			}
			if got := newestVersionDir(dir); got != c.want {
				t.Errorf("newestVersionDir(%v) = %q, want %q", c.dirs, got, c.want)
			}
		})
	}

	// A FILE beside the version directories is not a version. The plugin cache holds
	// marketplace metadata next to the installs, and treating a file as a version
	// would have --fix build into a path that is not a plugin root.
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "0.8.0"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "0.9.0"), []byte("not a directory"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := newestVersionDir(dir); got != "0.8.0" {
		t.Errorf("newestVersionDir picked %q; a file named like a version is not an install", got)
	}

	// An unreadable/absent plugin directory yields "", which siblingBinaries treats
	// as "skip this plugin" rather than as version zero.
	if got := newestVersionDir(filepath.Join(t.TempDir(), "does-not-exist")); got != "" {
		t.Errorf("a missing plugin directory must yield no version, got %q", got)
	}
}

// siblingRequirements aggregates every OTHER plugin's manifest, and the existing
// fixture's own plugin has no requirements.json — so the self-exclusion was never
// actually exercised, and a mutation sweep turned `!e.IsDir() || e.Name() == selfPlugin`
// into an `&&` with the suite green. Under that, the doctor reports its own manifest
// twice: once as itself and once as its own sibling, which reads to the operator as a
// second plugin needing the same tools.
func TestSiblingRequirementsExcludesSelf(t *testing.T) {
	market := t.TempDir()
	manifest := `{"tools":[{"name":"node","purpose":"scripts","tier":"recommended","check_cmd":"node --version","install":{"windows":"winget","darwin":"brew","linux":"apt"}}]}`
	for _, plugin := range []string{"prosthetic-conscience", "gray-area"} {
		dir := filepath.Join(market, plugin, "0.3.0")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "requirements.json"), []byte(manifest), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	// A loose FILE in the cache root is not a plugin either.
	if err := os.WriteFile(filepath.Join(market, "README.md"), []byte("not a plugin"), 0o644); err != nil {
		t.Fatal(err)
	}

	got := siblingRequirements(filepath.Join(market, "prosthetic-conscience", "0.3.0"))
	if len(got) != 1 {
		t.Fatalf("want only the sibling, got %d: %+v", len(got), got)
	}
	if got[0].Plugin != "gray-area" {
		t.Errorf("aggregated %q — the doctor's own manifest must not come back as a sibling", got[0].Plugin)
	}
}

// --fix's REPORT, which is what an operator with a degraded session actually reads.
// Every branch of it was unreachable from a test until the two process boundaries
// became parameters: the fetch path shells out to gh, and the fallback shells out to
// the Go toolchain.
func TestFixWithReportsEveryOutcome(t *testing.T) {
	missing := binStatus{Name: "sc-doctor", Plugin: "prosthetic-conscience", Version: "0.9.0", Root: "/plug"}

	ok := func(binStatus) error { return nil }
	boom := func(binStatus) error { return fmt.Errorf("gh not on PATH") }
	built := func(binStatus) (string, error) { return "", nil }
	brokenBuild := func(binStatus) (string, error) { return "cmd/x: undefined: Foo", fmt.Errorf("exit 1") }

	cases := []struct {
		name      string
		fetch     func(binStatus) error
		build     func(binStatus) (string, error)
		goPresent bool
		want      []string // substrings, all required
		absent    []string
	}{
		{
			name:  "fetch succeeds — the build is never attempted",
			fetch: ok, build: func(binStatus) (string, error) {
				t.Error("built from source after a successful fetch — the CI-built asset is the primary path, not a race")
				return "", nil
			},
			goPresent: true,
			want:      []string{"sc-doctor (prosthetic-conscience)", "checksum verified"},
			absent:    []string{"fetch failed"},
		},
		{
			name:  "fetch fails, source build succeeds",
			fetch: boom, build: built, goPresent: true,
			want:   []string{"fetch failed", "gh not on PATH", "built from source", "dev fallback"},
			absent: []string{"BUILD FAILED"},
		},
		{
			name:  "fetch fails and the build fails — both causes must survive",
			fetch: boom, build: brokenBuild, goPresent: true,
			// The fetch error is the ROOT cause and the build error is what the
			// operator can act on; a report carrying only one of them sends them
			// to the wrong problem.
			want: []string{"fetch failed", "gh not on PATH", "BUILD FAILED", "undefined: Foo"},
		},
		{
			name:  "fetch fails and there is no Go — the manual route is spelled out",
			fetch: boom, build: brokenBuild, goPresent: false,
			// Naming the exact asset and destination matters: this is the only
			// path where the operator has to do it by hand.
			want:   []string{"fetch failed", "install gh or Go", "sc-doctor_", filepath.Join("/plug", "bin")},
			absent: []string{"BUILD FAILED", "built from source"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := fixWith([]binStatus{missing}, c.fetch, c.build, c.goPresent)
			if len(got) != 1 {
				t.Fatalf("want one report line per missing binary, got %d: %v", len(got), got)
			}
			for _, sub := range c.want {
				if !strings.Contains(got[0], sub) {
					t.Errorf("report %q is missing %q", got[0], sub)
				}
			}
			for _, sub := range c.absent {
				if strings.Contains(got[0], sub) {
					t.Errorf("report %q should not mention %q", got[0], sub)
				}
			}
		})
	}

	// Built binaries are skipped without touching either boundary, and the report
	// stays per-binary rather than collapsing.
	never := func(binStatus) error { t.Error("attempted to provision an already-built binary"); return nil }
	report := fixWith([]binStatus{
		{Name: "a", Built: true},
		missing,
		{Name: "b", Plugin: "gray-area", Version: "0.2.0", Root: "/g"},
	}, func(b binStatus) error {
		if b.Built {
			return never(b)
		}
		return boom(b)
	}, built, true)
	if len(report) != 2 {
		t.Fatalf("want 2 lines for 2 missing binaries, got %d: %v", len(report), report)
	}
	if !strings.Contains(report[0], "sc-doctor") || !strings.Contains(report[1], "b") {
		t.Errorf("report lost per-binary identity or order: %v", report)
	}
}
