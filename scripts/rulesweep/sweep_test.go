package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var testClasses = []string{"adjacent-seat-omission", "unobservable-duty"}

// The gate's decision table. These were `--selftest` cases in the .mjs original, which
// could only print ok/FAIL and had to be invoked as a separate CI step; as Go tests they
// run under `go test ./...` with the rest of the module and name their own failures.
func TestEvaluate(t *testing.T) {
	cases := []struct {
		name string
		in   input
		want bool
	}{
		{
			"code-only change is not gated",
			input{files: []string{"plugins/x/tools/internal/record/render.go"}, messages: []string{"fix: thing"}, classes: testClasses},
			true,
		},
		{
			"protocol change with no trailers fails",
			input{files: []string{"plugins/x/skills/y/SKILL.md"}, messages: []string{"fix: thing"}, classes: testClasses},
			false,
		},
		{
			"protocol change with both trailers passes",
			input{
				files:    []string{"plugins/x/agents/red-auditor.md"},
				messages: []string{"fix: thing\n\nRule-Class: adjacent-seat-omission\nSibling-Sweep: checked blue and bench, neither carries this duty"},
				classes:  testClasses,
			},
			true,
		},
		{
			"an invented class fails",
			input{
				files:    []string{"law/README.md"},
				messages: []string{"x\n\nRule-Class: made-up-thing\nSibling-Sweep: checked everything everywhere"},
				classes:  testClasses,
			},
			false,
		},
		{
			"a token sweep fails",
			input{
				files:    []string{"plugins/x/skills/y/SKILL.md"},
				messages: []string{"x\n\nRule-Class: unobservable-duty\nSibling-Sweep: done"},
				classes:  testClasses,
			},
			false,
		},
		{
			// The defect this gate shipped with: prose EXPLAINING the mechanism satisfied
			// the mechanism. A check that passes on a description of the act rather than
			// the act is the very class it exists to prevent.
			"prose DESCRIBING the trailers does not satisfy them",
			input{
				files:    []string{"plugins/x/skills/y/SKILL.md"},
				messages: []string{"a change\n\nthe gate needs Rule-Class: adjacent-seat-omission and\nSibling-Sweep: a statement of what was checked\n\nbut this closing paragraph is prose, not a trailer block"},
				classes:  testClasses,
			},
			false,
		},
		{
			"seat prompts count as protocol",
			input{
				files:    []string{"plugins/frank-exchange-of-views/skills/research-protocol/scripts/debate.js"},
				messages: []string{"x"},
				classes:  testClasses,
			},
			false,
		},
		{
			// Attribution trailers conventionally sit last, so the trailer block cannot
			// be "the final paragraph only" — a rule that fights the convention it rides
			// on gets worked around rather than obeyed.
			"trailers in a non-final paragraph still count",
			input{
				files: []string{"plugins/x/skills/y/SKILL.md"},
				messages: []string{"fix: thing\n\nRule-Class: unobservable-duty\nSibling-Sweep: checked the adjacent seats, none carries this duty\n\n" +
					"Co-Authored-By: Someone <s@example.com>"},
				classes: testClasses,
			},
			true,
		},
		{
			"a comma-separated class list is accepted when every slug is known",
			input{
				files:    []string{"plugins/x/skills/y/SKILL.md"},
				messages: []string{"x\n\nRule-Class: adjacent-seat-omission, unobservable-duty\nSibling-Sweep: checked both neighbours and the bench seat"},
				classes:  testClasses,
			},
			true,
		},
		{
			"one unknown slug among known ones still fails",
			input{
				files:    []string{"plugins/x/skills/y/SKILL.md"},
				messages: []string{"x\n\nRule-Class: adjacent-seat-omission, invented-thing\nSibling-Sweep: checked both neighbours and the bench seat"},
				classes:  testClasses,
			},
			false,
		},
		{
			// Trailers may land on ANY commit in the branch, not only the one that
			// touched the surface.
			"trailers on a later commit in the range count",
			input{
				files:    []string{"plugins/x/skills/y/SKILL.md"},
				messages: []string{"chore: sweep\n\nRule-Class: unobservable-duty\nSibling-Sweep: checked the adjacent seats, none carries this duty", "fix: the protocol change itself"},
				classes:  testClasses,
			},
			true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := evaluate(c.in)
			if got.ok != c.want {
				t.Errorf("evaluate().ok = %v, want %v (problems: %v)", got.ok, c.want, got.problems)
			}
		})
	}
}

func TestTouchesProtocol(t *testing.T) {
	protocol := []string{
		"plugins/x/skills/y/SKILL.md",
		"plugins/x/agents/red.md",
		"plugins/x/skills/y/references/CHEATSHEET.md",
		"plugins/x/commands/doctor.md",
		"plugins/frank-exchange-of-views/skills/research-protocol/scripts/debate.js",
		"law/proposed/rule.md",
	}
	for _, f := range protocol {
		if got := touchesProtocol([]string{f}); len(got) != 1 {
			t.Errorf("%s must be a protocol surface", f)
		}
	}
	// Engineering, not protocol. A change to debate.js's dispatch arithmetic is the
	// first kind; a change to what a seat is TOLD is the second.
	code := []string{
		"plugins/x/tools/internal/record/render.go",
		"plugins/x/tools/cmd/feov-record/main.go",
		"scripts/rulesweep/sweep.go",
		"README.md",
		"plugins/x/skills/y/SKILL.md.bak",
		"outlaw/thing.md", // must not match the ^law/ anchor
	}
	for _, f := range code {
		if got := touchesProtocol([]string{f}); len(got) != 0 {
			t.Errorf("%s must NOT be a protocol surface, got %v", f, got)
		}
	}
}

func TestKnownClasses(t *testing.T) {
	registry := "# Registry\n\n## Section\n\n### adjacent-seat-omission\n\nprose\n\n### unobservable-duty\n\n" +
		"#### not-a-class-too-deep\n\n## not-a-class-too-shallow\n\n### Capitalised-Not-A-Slug\n"
	got := knownClasses(registry)
	want := []string{"adjacent-seat-omission", "unobservable-duty"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("knownClasses = %v, want %v", got, want)
	}
}

func TestRangesAreNotTheSameRange(t *testing.T) {
	diff, log := ranges("origin/main")
	if diff != "origin/main...HEAD" {
		t.Errorf("diff range = %q; a pull request is the three-dot DIFF", diff)
	}
	if log != "origin/main..HEAD" {
		t.Errorf("log range = %q; the three-dot LOG is the symmetric difference and collects the base side too (#146)", log)
	}
}

// The range test against a REAL repository.
//
// evaluate() takes messages already extracted, so every table case above is blind to where
// they came from — which is how the wrong log range (#146) survived a self-test the repo
// described as "genuinely good". A pure core is only as honest as the impure edge feeding
// it. The scenario is the one that actually happened: a branch behind its base, a
// trailer-less commit on the branch, and a commit carrying trailers on the base side only.
func TestCollectUsesTheBranchSideOnly(t *testing.T) {
	dir := t.TempDir()
	// -c rather than `git config`: a scratch repo must not read, or need, identity from
	// the developer's global config.
	git := func(args ...string) {
		t.Helper()
		full := append([]string{
			"-c", "user.email=t@example.com", "-c", "user.name=t", "-c", "commit.gpgsign=false",
		}, args...)
		cmd := exec.Command("git", full...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	write := func(rel, body string) {
		t.Helper()
		p := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	git("init", "-q", "-b", "main")
	write("seed.txt", "seed\n")
	git("add", "-A")
	git("commit", "-q", "-m", "seed")

	// The branch: touches a protocol surface, carries NO trailers.
	git("checkout", "-q", "-b", "feature")
	write("plugins/x/skills/y/SKILL.md", "# y\n")
	git("add", "-A")
	git("commit", "-q", "-m", "feat: a protocol change with no sweep")

	// The base moves on, and the commit landing there DOES carry trailers.
	git("checkout", "-q", "main")
	write("other.txt", "other\n")
	git("add", "-A")
	git("commit", "-q", "-m", "fix: unrelated\n\nRule-Class: adjacent-seat-omission\nSibling-Sweep: checked every sibling seat, none carries this duty")
	git("checkout", "-q", "feature")

	files, messages, err := collect("main", dir)
	if err != nil {
		t.Fatal(err)
	}
	if !slicesContains(files, "plugins/x/skills/y/SKILL.md") {
		t.Errorf("the branch-side protocol file was not seen: %v", files)
	}
	for _, m := range messages {
		if strings.Contains(m, "Rule-Class:") {
			t.Errorf("a base-side-only commit message was collected — the log range is the symmetric difference (#146): %q", m)
		}
	}
	if got := evaluate(input{files: files, messages: messages, classes: testClasses}); got.ok {
		t.Error("a branch behind its base must FAIL on its own missing trailers (#146)")
	}

	// And it must still pass when the trailers really are on the branch.
	git("commit", "-q", "--allow-empty", "-m", "chore: sweep\n\nRule-Class: adjacent-seat-omission\nSibling-Sweep: checked the adjacent seats, none carries this duty")
	files, messages, err = collect("main", dir)
	if err != nil {
		t.Fatal(err)
	}
	if got := evaluate(input{files: files, messages: messages, classes: testClasses}); !got.ok {
		t.Errorf("trailers ON THE BRANCH must pass: %v", got.problems)
	}
}

// A base ref that does not exist must be an ERROR, not an empty range.
//
// The .mjs original returned "" for any git failure, so `--base does-not-exist` produced
// zero changed files and reported "no protocol surface touched" — a PASS. A gate that
// passes when it cannot read its input fails in the direction nobody investigates.
func TestCollectFailsOnAnUnknownBase(t *testing.T) {
	dir := t.TempDir()
	cmd := exec.Command("git", "init", "-q", "-b", "main")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}
	if _, _, err := collect("no-such-ref", dir); err == nil {
		t.Error("an unresolvable base ref must surface as an error, not as an empty (passing) range")
	}
}

func slicesContains(haystack []string, needle string) bool {
	for _, h := range haystack {
		if h == needle {
			return true
		}
	}
	return false
}
