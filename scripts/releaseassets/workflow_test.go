package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func readHooks(t *testing.T) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", ".github", "workflows", "hooks.yml"))
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// jobsOf is the workflow's jobs mapping: the document's root mapping, then its jobs key.
func jobsOf(wf []byte) (*yaml.Node, error) {
	var root yaml.Node
	if err := yaml.Unmarshal(wf, &root); err != nil {
		return nil, err
	}
	if root.Kind != yaml.DocumentNode || len(root.Content) != 1 || root.Content[0].Kind != yaml.MappingNode {
		return nil, fmt.Errorf("the workflow is not one mapping document")
	}
	if jobs := mapValue(root.Content[0], "jobs"); jobs != nil && jobs.Kind == yaml.MappingNode {
		return jobs, nil
	}
	return nil, fmt.Errorf("the workflow has no jobs mapping")
}

func mapValue(m *yaml.Node, key string) *yaml.Node {
	if m == nil || m.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value == key {
			return m.Content[i+1]
		}
	}
	return nil
}

const (
	restoreIf = "needs.native-scope.outputs.cache != 'none'"
	saveIf    = "needs.native-scope.outputs.cache == 'save' && steps.helper-cache.outputs.cache-hit != 'true'"
	helperKey = "native-${{ matrix.helper }}-${{ matrix.target }}-${{ matrix.builder }}-${{ steps.image.outputs.image }}-${{ steps.digest.outputs.digest }}"
)

// cacheStepProblems lists every way the workflow's cache steps break native-scope's cache rule:
// a cache action outside native-build's one restore and one save, either with another condition,
// a key whose image segment is not the image step's output, a restore with restore-keys, or the
// combined actions/cache anywhere.
func cacheStepProblems(wf []byte) ([]string, error) {
	jobs, err := jobsOf(wf)
	if err != nil {
		return nil, err
	}
	var problems []string
	restores, saves, imageSteps := 0, 0, 0
	for i := 0; i+1 < len(jobs.Content); i += 2 {
		job := jobs.Content[i].Value
		steps := mapValue(jobs.Content[i+1], "steps")
		if steps == nil {
			continue
		}
		for _, step := range steps.Content {
			uses, ifCond, id := "", "", ""
			if n := mapValue(step, "uses"); n != nil {
				uses = n.Value
			}
			if n := mapValue(step, "if"); n != nil {
				ifCond = strings.TrimSpace(n.Value)
			}
			if n := mapValue(step, "id"); n != nil {
				id = n.Value
			}
			with := mapValue(step, "with")
			key := ""
			if n := mapValue(with, "key"); n != nil {
				key = n.Value
			}
			if job == "native-build" && id == "image" {
				if run := mapValue(step, "run"); run != nil && strings.Contains(run.Value, `"image=${ImageVersion}" >> "$GITHUB_OUTPUT"`) && strings.Contains(run.Value, "exit 1") {
					imageSteps++
				}
			}
			action, _, _ := strings.Cut(uses, "@")
			switch action {
			case "actions/cache/restore":
				restores++
				if job != "native-build" {
					problems = append(problems, job+" restores a cache; only native-build's helper restore may")
				}
				if ifCond != restoreIf {
					problems = append(problems, fmt.Sprintf("%s's restore runs if %q, want %q", job, ifCond, restoreIf))
				}
				if mapValue(with, "restore-keys") != nil {
					problems = append(problems, job+"'s restore has restore-keys: a prefix hit restores another digest's helper")
				}
				if key != helperKey {
					problems = append(problems, fmt.Sprintf("%s's restore key is %q, want %q", job, key, helperKey))
				}
			case "actions/cache/save":
				saves++
				if job != "native-build" {
					problems = append(problems, job+" saves a cache; only native-build's helper save may")
				}
				if ifCond != saveIf {
					problems = append(problems, fmt.Sprintf("%s's save runs if %q, want %q", job, ifCond, saveIf))
				}
				if key != helperKey {
					problems = append(problems, fmt.Sprintf("%s's save key is %q, want %q", job, key, helperKey))
				}
			case "actions/cache":
				problems = append(problems, job+" uses the combined actions/cache, whose post step saves on any event")
			}
		}
	}
	if restores != 1 || saves != 1 {
		problems = append(problems, fmt.Sprintf("the workflow holds %d cache restores and %d cache saves, want native-build's one of each", restores, saves))
	}
	if imageSteps != 1 {
		problems = append(problems, "native-build has no step id image that writes the runner's ImageVersion to its output and fails when it is empty")
	}
	return problems, nil
}

func TestTheHelperCacheStepsObeyTheScope(t *testing.T) {
	wf := readHooks(t)
	problems, err := cacheStepProblems(wf)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range problems {
		t.Error(p)
	}

	restoreLine := "        if: " + restoreIf + "\n        uses: actions/cache/restore@"
	if strings.Count(string(wf), restoreLine) != 1 {
		t.Fatalf("hooks.yml does not hold the restore step as the fixtures expect:\n%s", restoreLine)
	}
	restoreWith := "        uses: actions/cache/restore@v4\n        # zizmor: ignore[cache-poisoning]\n        with:\n"
	saveLine := "        if: " + saveIf + "\n        uses: actions/cache/save@"
	keyLine := "          key: " + helperKey + "\n"
	for _, tc := range []struct{ name, old, new string }{
		{"the restore runs on github.event_name", restoreLine, "        if: github.event_name != 'push'\n        uses: actions/cache/restore@"},
		{"the restore has no if", restoreLine, "        uses: actions/cache/restore@"},
		{"the save is the combined actions/cache", saveLine, "        if: " + saveIf + "\n        uses: actions/cache@"},
		{"the restore has restore-keys", restoreWith, restoreWith + "          restore-keys: native-feov-ocr-\n"},
		{"the key reads the env context", keyLine, strings.Replace(keyLine, "steps.image.outputs.image", "env.ImageVersion", 1)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if !strings.Contains(string(wf), tc.old) {
				t.Fatalf("hooks.yml does not hold %q", tc.old)
			}
			fixture := strings.Replace(string(wf), tc.old, tc.new, 1)
			if fixture == string(wf) {
				t.Fatal("the fixture changed nothing")
			}
			problems, err := cacheStepProblems([]byte(fixture))
			if err != nil {
				t.Fatal(err)
			}
			if len(problems) == 0 {
				t.Errorf("a workflow where %s passes the check", tc.name)
			}
		})
	}
}

// TestTheNativeMatricesDoNotFailFast: one red leg must not cancel the others, or a darwin failure
// hides whether Windows passed.
func TestTheNativeMatricesDoNotFailFast(t *testing.T) {
	check := func(wf []byte) []string {
		jobs, err := jobsOf(wf)
		if err != nil {
			t.Fatal(err)
		}
		var bad []string
		for _, name := range []string{"native-build", "native-run"} {
			ff := mapValue(mapValue(mapValue(jobs, name), "strategy"), "fail-fast")
			if ff == nil || ff.Value != "false" {
				bad = append(bad, name)
			}
		}
		return bad
	}
	wf := readHooks(t)
	if bad := check(wf); len(bad) > 0 {
		t.Errorf("%v do not set strategy.fail-fast: false", bad)
	}
	fixture := strings.Replace(string(wf), "    runs-on: ${{ matrix.runner }}\n    strategy:\n      fail-fast: false\n", "    runs-on: ${{ matrix.runner }}\n    strategy:\n", 1)
	if fixture == string(wf) {
		t.Fatal("the fixture changed nothing")
	}
	if bad := check([]byte(fixture)); len(bad) != 1 || bad[0] != "native-run" {
		t.Errorf("native-run without fail-fast: false reads as %v", bad)
	}
}
