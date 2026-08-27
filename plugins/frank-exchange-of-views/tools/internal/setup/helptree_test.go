package setup

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func rendered(role string) (string, error) { return "# " + role + " surface\n", nil }

// A STAGING STEP THAT CAN DO NOTHING MUST SAY SO.
//
// This is the third repair in this file's family for one defect: MirrorGapPatterns returned
// {Written: true, Files: 55} having written nothing at all, because the directory did not exist
// and the write's error was discarded. The renderer here is INJECTED — internal/cli owns the
// command tree and imports this package, so the dependency can only run one way — which means a
// caller can arrive without one. If that were absorbed, setup would print no help line and the
// run would look exactly like a run whose seats had nothing to learn.
func TestANilRendererIsReportedRatherThanAbsorbed(t *testing.T) {
	dir := t.TempDir()
	got := StageHelpTrees(dir, []string{"merge"}, nil)
	if got.Written {
		t.Fatal("a nil renderer reported Written — the summary would say the surface was staged")
	}
	if !strings.Contains(got.Reason, "renderer") {
		t.Errorf("the reason must name what was missing; got %q", got.Reason)
	}
	if entries, _ := os.ReadDir(filepath.Join(dir, "inputs")); len(entries) != 0 {
		t.Errorf("inputs/ has %d entries after a refused staging", len(entries))
	}
}

// It creates inputs/ ITSELF. Every caller outside run.go arrives without it — a probe, a test, a
// reordering of setup — and that is precisely the condition under which the sibling mirror
// reported fifty-five staged files and wrote none.
func TestStagingCreatesInputsInAFreshRunDirectory(t *testing.T) {
	dir := t.TempDir()
	got := StageHelpTrees(dir, []string{"merge", "blue"}, rendered)
	if !got.Written {
		t.Fatalf("nothing staged into a fresh run directory: %q", got.Reason)
	}
	for _, role := range []string{"merge", "blue"} {
		body, err := os.ReadFile(filepath.Join(dir, "inputs", "help-"+role+".md"))
		if err != nil {
			t.Fatalf("help-%s.md: %v", role, err)
		}
		if !strings.Contains(string(body), role) {
			t.Errorf("help-%s.md does not carry %s's surface: %q", role, role, string(body))
		}
	}
	if got.Bytes == 0 {
		t.Error("Bytes is 0 with two files written — the summary would report a free staging")
	}
}

// A PARTIAL STAGING IS NOT A SUCCESS FOR THE ROLES IT MISSED.
//
// Reporting the set as staged because SOME of it was is the same shape as the count that was
// believable: three seats get their surface, the fourth walks the tree, and the summary says all
// four were served. Roles carries who actually got one and Reason names who did not.
func TestAFailedRoleIsNamedWhileTheOthersStillStage(t *testing.T) {
	dir := t.TempDir()
	got := StageHelpTrees(dir, []string{"merge", "adjudicator", "blue"}, func(role string) (string, error) {
		if role == "adjudicator" {
			return "", errors.New("not a role with a seat namespace")
		}
		return rendered(role)
	})
	if !got.Written {
		t.Fatalf("two renderable roles staged nothing: %q", got.Reason)
	}
	if strings.Join(got.Roles, ",") != "merge,blue" {
		t.Errorf("Roles = %v, want the two that rendered", got.Roles)
	}
	if !strings.Contains(got.Reason, "adjudicator") {
		t.Errorf("the reason must name the role that did NOT stage; got %q", got.Reason)
	}
	if _, err := os.Stat(filepath.Join(dir, "inputs", "help-adjudicator.md")); !os.IsNotExist(err) {
		t.Error("a role whose render failed still got a file — an empty surface reads as a role with no verbs")
	}
}

// AND A RE-RUN REPLACES IT. The other mirrors keep a pre-staged file ("already staged"), because
// an operator may have hand-placed one. This one is GENERATED FROM THE BINARY, so a kept copy is
// a copy of some OTHER binary's surface — the precise drift the generation exists to remove. A
// setup re-run after a plugin update must land the new surface, not preserve the old.
func TestARerunRestagesRatherThanKeepingAStaleSurface(t *testing.T) {
	dir := t.TempDir()
	if got := StageHelpTrees(dir, []string{"merge"}, func(string) (string, error) {
		return "# the OLD binary's surface\n", nil
	}); !got.Written {
		t.Fatalf("seeding failed: %q", got.Reason)
	}
	if got := StageHelpTrees(dir, []string{"merge"}, rendered); !got.Written {
		t.Fatalf("the re-run staged nothing: %q", got.Reason)
	}
	body, err := os.ReadFile(filepath.Join(dir, "inputs", "help-merge.md"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(body), "OLD") {
		t.Error("the re-run kept the previous surface; after a plugin update every seat would read " +
			"help for a binary that is no longer there, which is worse than no help at all")
	}
}
