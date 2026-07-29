package checkpoint

import (
	"path/filepath"
	"testing"
)

func TestParseFrontmatterAndSections(t *testing.T) {
	n := Parse("---\nschema: 2\nobjective: \"ship the restore hook\"\nagent_id: null\n---\n" +
		"## Validation loop\n1. go test\n2. go vet\n## Open threads\n- none\n")

	if got := n.Get("objective"); got != "ship the restore hook" {
		t.Errorf("objective = %q — quotes must be stripped", got)
	}
	if got := n.Get("schema"); got != "2" {
		t.Errorf("schema = %q", got)
	}
	if got := n.Get("missing"); got != "" {
		t.Errorf("absent key = %q, want empty", got)
	}
	body, ok := n.Section("Validation loop")
	if !ok || body != "1. go test\n2. go vet" {
		t.Errorf("validation loop = %q (present=%v)", body, ok)
	}
	if len(n.Sections) != 2 {
		t.Errorf("sections = %d, want 2", len(n.Sections))
	}
}

// Absent and empty are different facts: an empty validation loop is a note that
// recorded no checks; a missing one is a note written before they existed.
func TestAbsentAndEmptyAreDistinct(t *testing.T) {
	n := Parse("## Validation loop\n\n")
	if body, ok := n.Section("Validation loop"); !ok || body != "" {
		t.Errorf("empty section: body=%q present=%v — want present and empty", body, ok)
	}
	if _, ok := n.Section("Next intended steps"); ok {
		t.Error("absent section reported present")
	}
	if _, ok := n.NonEmptySection("Validation loop"); ok {
		t.Error("an empty section is not content")
	}
}

// The note is written by an agent under time pressure. A parser that refuses a
// slightly malformed note is one that fails exactly when it is needed.
func TestMalformedNotesDegradeRatherThanFail(t *testing.T) {
	for _, raw := range []string{
		"",
		"no frontmatter at all\n## Validation loop\n1. go test\n",
		"---\nnot: closed\n## Validation loop\n1. go test\n",
		"---\n---\n## Validation loop\n1. go test\n",
		"## Validation loop\n1. go test",
	} {
		n := Parse(raw)
		if raw == "" {
			if len(n.Sections) != 0 {
				t.Errorf("empty input produced %d sections", len(n.Sections))
			}
			continue
		}
		if _, ok := n.NonEmptySection("Validation loop"); !ok {
			t.Errorf("lost the validation loop from:\n%q", raw)
		}
	}
}

// Template scaffolding is not content — restoring it would introduce rather than
// reinforce, which is the failure that gets a digest read as prompt injection.
func TestScaffoldingIsNotContent(t *testing.T) {
	n := Parse("## Validation loop\n← load-bearing\n<!-- sealed: trigger=auto -->\n")
	if body, ok := n.NonEmptySection("Validation loop"); ok {
		t.Errorf("scaffolding counted as content: %q", body)
	}
}

func TestHeadingsInFileOrder(t *testing.T) {
	got := Headings("## B\nx\n## A\ny\n### not a section\n")
	if len(got) != 2 || got[0] != "B" || got[1] != "A" {
		t.Errorf("headings = %v, want [B A]", got)
	}
}

// The search order is the contract the seal and the restore share: a workspace
// note wins over the session-local fallback, and projects/ wins over research/.
func TestNotePathSearchOrder(t *testing.T) {
	root := filepath.Join("/w")
	present := map[string]bool{}
	exists := func(p string) bool { return present[p] }
	glob := func(pat string) ([]string, error) {
		var out []string
		for p := range present {
			if ok, _ := filepath.Match(pat, p); ok {
				out = append(out, p)
			}
		}
		return out, nil
	}

	proj := filepath.Join(root, "projects", "alpha", "CHECKPOINT.md")
	res := filepath.Join(root, "research", "beta", "CHECKPOINT.md")
	fallback := filepath.Join(root, ".claude", "checkpoints", "CHECKPOINT.md")

	if got := NotePath(root, exists, glob); got != "" {
		t.Errorf("nothing present, got %q", got)
	}
	present[fallback] = true
	if got := NotePath(root, exists, glob); got != fallback {
		t.Errorf("fallback: got %q", got)
	}
	present[res] = true
	if got := NotePath(root, exists, glob); got != res {
		t.Errorf("research should beat the fallback: got %q", got)
	}
	present[proj] = true
	if got := NotePath(root, exists, glob); got != proj {
		t.Errorf("projects should beat research: got %q", got)
	}
	if got := NotePath("", exists, glob); got != "" {
		t.Errorf("no project dir, got %q", got)
	}
}

// Two workspace notes is ambiguous; the choice must at least be STABLE, or the
// seal and the restore can pick different files on the same tree.
func TestNotePathIsDeterministicUnderAmbiguity(t *testing.T) {
	root := "/w"
	present := map[string]bool{
		filepath.Join(root, "projects", "zeta", "CHECKPOINT.md"):  true,
		filepath.Join(root, "projects", "alpha", "CHECKPOINT.md"): true,
	}
	exists := func(p string) bool { return present[p] }
	glob := func(pat string) ([]string, error) {
		var out []string
		for p := range present {
			if ok, _ := filepath.Match(pat, p); ok {
				out = append(out, p)
			}
		}
		return out, nil
	}
	first := NotePath(root, exists, glob)
	for range 20 {
		if got := NotePath(root, exists, glob); got != first {
			t.Fatalf("unstable choice: %q then %q — the seal and the restore would diverge", first, got)
		}
	}
	if want := filepath.Join(root, "projects", "alpha", "CHECKPOINT.md"); first != want {
		t.Errorf("got %q, want the lexicographically first: %q", first, want)
	}
}
