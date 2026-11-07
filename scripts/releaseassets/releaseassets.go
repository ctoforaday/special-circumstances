package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/ctoforaday/special-circumstances/scripts/internal/staticproof"
	"gopkg.in/yaml.v3"
)

// Runner mappings per pair. A published target with no runner is refused.
var targetRunners = map[string]string{
	"linux/amd64":   "ubuntu-24.04",
	"linux/arm64":   "ubuntu-24.04-arm",
	"darwin/arm64":  "macos-15",
	"windows/amd64": "windows-2025",
	"windows/arm64": "windows-11-arm",
}

// Target holds the facts for one published OS/arch pair.
type Target struct {
	Pair   string `json:"pair"`
	Name   string `json:"name"`
	OS     string `json:"os"`
	Arch   string `json:"arch"`
	Ext    string `json:"ext"`
	Runner string `json:"runner"`
}

// BuildLeg is one entry in native-build's matrix.
type BuildLeg struct {
	Helper  string `json:"helper"`
	Target  string `json:"target"`
	Builder string `json:"builder"`
	OS      string `json:"os"`
	Arch    string `json:"arch"`
	Ext     string `json:"ext"`
}

// RunLeg is one entry in native-run's matrix.
type RunLeg struct {
	Target string `json:"target"`
	Runner string `json:"runner"`
	OS     string `json:"os"`
	Arch   string `json:"arch"`
	Ext    string `json:"ext"`
}

// Matrix is a GitHub Actions matrix payload with an include list.
type Matrix[T any] struct {
	Include []T `json:"include"`
}

// ScopeResult is the decision returned by releaseassets -scope.
type ScopeResult struct {
	Cache     string           `json:"cache"`
	Full      bool             `json:"full"`
	Matrix    Matrix[BuildLeg] `json:"matrix"`
	RunMatrix Matrix[RunLeg]   `json:"run_matrix"`
}

var reForPair = regexp.MustCompile(`(?m)^\s*for\s+pair\s+in\s+([^;]+);`)

// ReadReleasePairs parses the single "for pair in ...;" loop in the release job of hooks.yml.
func ReadReleasePairs(workflowPath string) ([]string, error) {
	b, err := os.ReadFile(workflowPath)
	if err != nil {
		return nil, err
	}
	content := string(b)
	// Find release: job span
	relIdx := strings.Index(content, "\n  release:\n")
	if relIdx < 0 {
		return nil, fmt.Errorf("%s: no release job found", workflowPath)
	}
	relSpan := content[relIdx:]
	matches := reForPair.FindAllStringSubmatch(relSpan, -1)
	if len(matches) != 1 {
		return nil, fmt.Errorf("%s: release job must have exactly one 'for pair in ...;' line, found %d", workflowPath, len(matches))
	}
	pairs := strings.Fields(matches[0][1])
	if len(pairs) == 0 {
		return nil, fmt.Errorf("%s: 'for pair in' line has no targets", workflowPath)
	}
	return pairs, nil
}

// ParseTargets resolves the published pairs to Targets, refusing any pair without a runner.
func ParseTargets(pairs []string) ([]Target, error) {
	var targets []Target
	for _, pair := range pairs {
		runner, ok := targetRunners[pair]
		if !ok {
			return nil, fmt.Errorf("target %q has no runner", pair)
		}
		parts := strings.Split(pair, "/")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid pair format: %q", pair)
		}
		osName, archName := parts[0], parts[1]
		ext := ""
		if osName == "windows" {
			ext = ".exe"
		}
		targets = append(targets, Target{
			Pair:   pair,
			Name:   osName + "-" + archName,
			OS:     osName,
			Arch:   archName,
			Ext:    ext,
			Runner: runner,
		})
	}
	return targets, nil
}

// FindHelpers finds all helper directories under nativeRoot (directories containing an ASSETS file).
func FindHelpers(nativeRoot string) ([]string, error) {
	entries, err := os.ReadDir(nativeRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var helpers []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		assetsPath := filepath.Join(nativeRoot, e.Name(), "ASSETS")
		if _, err := os.Stat(assetsPath); err == nil {
			helpers = append(helpers, e.Name())
		}
	}
	slices.Sort(helpers)
	return helpers, nil
}

// BuilderFor returns the runner that builds helper for target.
func BuilderFor(helper string, target Target) string {
	if helper == "feov-ocr" {
		return "ubuntu-24.04"
	}
	if helper == "feov-embed" {
		if target.OS == "windows" {
			return target.Runner
		}
		return "ubuntu-24.04"
	}
	return "ubuntu-24.04"
}

// DeriveConsumerTargets finds which targets are downloaded by jobs in hooks.yml that need native-build
// (excluding native-run and release).
func DeriveConsumerTargets(workflowPath string) ([]string, error) {
	b, err := os.ReadFile(workflowPath)
	if err != nil {
		return nil, err
	}
	var wf struct {
		Jobs map[string]struct {
			Needs any         `yaml:"needs"`
			Steps []yaml.Node `yaml:"steps"`
		} `yaml:"jobs"`
	}
	if err := yaml.Unmarshal(b, &wf); err != nil {
		return nil, fmt.Errorf("%s: parsing workflow YAML: %w", workflowPath, err)
	}

	targetSet := map[string]bool{}
	// Look for native-*-<target> artifact downloads in consumer jobs
	reDownload := regexp.MustCompile(`native-[a-zA-Z0-9_-]+-([a-z0-9]+-[a-z0-9]+)`)

	for jobName, job := range wf.Jobs {
		if jobName == "native-run" || jobName == "release" || jobName == "native-scope" || jobName == "native-build" {
			continue
		}
		needsNativeBuild := false
		switch n := job.Needs.(type) {
		case string:
			needsNativeBuild = (n == "native-build")
		case []any:
			for _, item := range n {
				if s, ok := item.(string); ok && s == "native-build" {
					needsNativeBuild = true
					break
				}
			}
		}
		if !needsNativeBuild {
			continue
		}

		// Search steps for downloaded artifact patterns
		for _, step := range job.Steps {
			var stepMap map[string]any
			if err := step.Decode(&stepMap); err == nil {
				if with, ok := stepMap["with"].(map[string]any); ok {
					if name, ok := with["name"].(string); ok {
						for _, m := range reDownload.FindAllStringSubmatch(name, -1) {
							targetSet[m[1]] = true
						}
					}
					if pattern, ok := with["pattern"].(string); ok {
						for _, m := range reDownload.FindAllStringSubmatch(pattern, -1) {
							targetSet[m[1]] = true
						}
					}
				}
				// Also check env or run commands if applicable
				if runCmd, ok := stepMap["run"].(string); ok {
					for _, m := range reDownload.FindAllStringSubmatch(runCmd, -1) {
						targetSet[m[1]] = true
					}
				}
			}
		}
	}

	var targets []string
	for t := range targetSet {
		targets = append(targets, t)
	}
	slices.Sort(targets)
	return targets, nil
}

// IsNativeSetChange reports whether any changed file belongs to the native set.
func IsNativeSetChange(changedFiles []string) bool {
	for _, f := range changedFiles {
		clean := filepath.ToSlash(f)
		if strings.HasPrefix(clean, "plugins/frank-exchange-of-views/native/") ||
			strings.HasPrefix(clean, "plugins/frank-exchange-of-views/tools/devcmd/nativedigest/") ||
			strings.HasPrefix(clean, "plugins/frank-exchange-of-views/tools/internal/nativehelper/") ||
			strings.HasPrefix(clean, "plugins/frank-exchange-of-views/tools/internal/tessocr/") ||
			clean == "plugins/frank-exchange-of-views/tools/internal/fetchcache/pagerender.go" ||
			strings.HasPrefix(clean, "scripts/releaseassets/") ||
			strings.HasPrefix(clean, "scripts/internal/staticproof/") ||
			strings.HasPrefix(clean, "scripts/staticproof/") ||
			clean == ".github/workflows/hooks.yml" {
			return true
		}
	}
	return false
}

// DecideScope implements the decision table from plan §III.1.5.
func DecideScope(event, ref string, changedFiles []string, allTargets []Target, testTargetNames []string, helpers []string) (*ScopeResult, error) {
	// Validate event against known workflow events
	validEvents := map[string]bool{
		"pull_request":      true,
		"push":              true,
		"schedule":          true,
		"workflow_dispatch": true,
	}
	if !validEvents[event] {
		return nil, fmt.Errorf("unknown event: %q", event)
	}

	// 1. cache decision: from ref first
	var cache string
	if strings.HasPrefix(ref, "refs/tags/") {
		cache = "none"
	} else if event == "schedule" {
		cache = "save"
	} else if event == "workflow_dispatch" && ref == "refs/heads/main" {
		cache = "save"
	} else {
		cache = "restore"
	}

	// Filter test targets from all targets
	var testTargets []Target
	for _, t := range allTargets {
		if slices.Contains(testTargetNames, t.Name) {
			testTargets = append(testTargets, t)
		}
	}
	if len(testTargets) == 0 && len(testTargetNames) > 0 {
		return nil, fmt.Errorf("none of the consumer targets %v matched published targets", testTargetNames)
	}

	// 2. build legs & full
	var selectedTargets []Target
	var full bool

	isFEOVTag := strings.HasPrefix(ref, "refs/tags/") && strings.Contains(ref, "frank-exchange-of-views--v")
	isOtherTag := strings.HasPrefix(ref, "refs/tags/") && !isFEOVTag

	switch event {
	case "pull_request":
		if IsNativeSetChange(changedFiles) {
			selectedTargets = allTargets
			full = true
		} else {
			selectedTargets = testTargets
			full = false
		}
	case "push":
		if ref == "refs/heads/main" {
			selectedTargets = testTargets
			full = false
		} else if isFEOVTag {
			selectedTargets = allTargets
			full = true
		} else if isOtherTag {
			selectedTargets = testTargets
			full = false
		} else {
			selectedTargets = allTargets
			full = true
		}
	case "schedule":
		selectedTargets = allTargets
		full = true
	case "workflow_dispatch":
		if isFEOVTag {
			selectedTargets = allTargets
			full = true
		} else if isOtherTag {
			selectedTargets = testTargets
			full = false
		} else {
			selectedTargets = allTargets
			full = true
		}
	}

	var buildLegs []BuildLeg
	for _, h := range helpers {
		for _, t := range selectedTargets {
			buildLegs = append(buildLegs, BuildLeg{
				Helper:  h,
				Target:  t.Name,
				Builder: BuilderFor(h, t),
				OS:      t.OS,
				Arch:    t.Arch,
				Ext:     t.Ext,
			})
		}
	}
	if len(buildLegs) == 0 {
		return nil, fmt.Errorf("empty native-build matrix")
	}

	var runLegs []RunLeg
	for _, t := range allTargets {
		runLegs = append(runLegs, RunLeg{
			Target: t.Name,
			Runner: t.Runner,
			OS:     t.OS,
			Arch:   t.Arch,
			Ext:    t.Ext,
		})
	}

	return &ScopeResult{
		Cache:     cache,
		Full:      full,
		Matrix:    Matrix[BuildLeg]{Include: buildLegs},
		RunMatrix: Matrix[RunLeg]{Include: runLegs},
	}, nil
}

// CheckDist verifies that distDir contains valid static assets.
func CheckDist(distDir string, ref string, targets []Target, helpers []string) error {
	entries, err := os.ReadDir(distDir)
	if err != nil {
		return fmt.Errorf("reading dist dir %s: %w", distDir, err)
	}

	present := map[string]os.DirEntry{}
	for _, e := range entries {
		if !e.IsDir() {
			present[e.Name()] = e
		}
	}

	isFEOV := strings.Contains(ref, "frank-exchange-of-views--v")

	if isFEOV {
		// Must have every helper for every target
		for _, h := range helpers {
			for _, t := range targets {
				assetName := fmt.Sprintf("%s_%s_%s%s", h, t.OS, t.Arch, t.Ext)
				if _, ok := present[assetName]; !ok {
					return fmt.Errorf("dist is missing helper asset: %s", assetName)
				}
			}
		}
	}

	// Run staticproof on every executable file in dist
	provenCount := 0
	for name := range present {
		if strings.HasSuffix(name, ".txt") || strings.HasSuffix(name, ".sha256") || name == "SHA256SUMS" {
			continue
		}
		path := filepath.Join(distDir, name)
		artifact := name
		if idx := strings.Index(name, "_"); idx >= 0 {
			artifact = name[:idx]
		}
		if err := staticproof.Check(artifact, path); err != nil {
			return err
		}
		provenCount++
		fmt.Printf("proven static: %s (%s)\n", name, artifact)
	}
	if len(present) > 0 && provenCount == 0 {
		return fmt.Errorf("no executable in dist was proven static")
	}
	return nil
}
