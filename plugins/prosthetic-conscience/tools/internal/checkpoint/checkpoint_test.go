package checkpoint

import (
	"os"
	"path/filepath"
	"strings"
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

func TestParseValidationLoop(t *testing.T) {
	loop := "1. `go test ./...`  → all ok  · re-armed by: any .go edit under tools/\n" +
		"   last run: pass\n" +
		"2. `qlty check`  → no issues\n" +
		"   re-armed by: .qlty/qlty.toml · and any new plugin\n" +
		"3. `make docs`  → builds\n"
	got := ParseValidationLoop(loop)
	if len(got) != 3 {
		t.Fatalf("checks = %d, want 3: %+v", len(got), got)
	}
	if got[0].ReArmedBy != "any .go edit under tools/" {
		t.Errorf("check 1 surface = %q", got[0].ReArmedBy)
	}
	// The marker on a CONTINUATION line still belongs to its check — an agent
	// under pressure wraps, and losing the surface to a line break is silent.
	if got[1].ReArmedBy != ".qlty/qlty.toml" {
		t.Errorf("check 2 surface = %q — the · clause should be trimmed", got[1].ReArmedBy)
	}
	// Absent is distinct from empty: check 3 recorded no surface at all.
	if got[2].ReArmedBy != "" {
		t.Errorf("check 3 surface = %q, want empty", got[2].ReArmedBy)
	}
	if got[0].Index != 1 || got[2].Index != 3 {
		t.Errorf("indices wrong: %+v", got)
	}
	if !strings.Contains(got[0].Line, "go test") {
		t.Errorf("check line not captured verbatim: %q", got[0].Line)
	}
}

// realProject builds an on-disk tree, because containment is now enforced by
// os.Root and a map-backed fake cannot exercise a symlink escape.
func realProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for _, d := range []string{"tools", "manifest", ".qlty"} {
		if err := os.MkdirAll(filepath.Join(dir, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	for _, f := range []string{filepath.Join(".qlty", "qlty.toml"), "go.mod"} {
		if err := os.WriteFile(filepath.Join(dir, f), []byte("x\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func targets(t *testing.T, dir, prose string) ([]string, string) {
	t.Helper()
	within, closeRoot := RootedWithin(dir)
	defer func() { _ = closeRoot() }()
	return WatchTargets(prose, within)
}

// THE measured constraint (hook-surface-spike.md §6): watchPaths takes paths,
// never patterns. A glob must be reduced to the directory that contains it, and
// the result must be something the watcher will actually accept.
func TestWatchTargetsReducesPatternsToDirectories(t *testing.T) {
	dir := realProject(t)
	cases := []struct {
		name, prose string
		want        []string
	}{
		{"glob reduces to its directory", "any edit to manifest/*.yml", []string{"manifest"}},
		{"deep glob reduces too", "any change under tools/**/*.go", []string{"tools"}},
		{"a literal file is watchable as itself", "the file .qlty/qlty.toml", []string{".qlty/qlty.toml"}},
		{"a bare directory passes through", "anything under tools/", []string{"tools"}},
		{"two surfaces both resolve", "tools/ or manifest/*.yml", []string{"manifest", "tools"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, why := targets(t, dir, c.prose)
			if len(got) != len(c.want) {
				t.Fatalf("targets = %v (reason %q), want %v", got, why, c.want)
			}
			for i := range got {
				if got[i] != c.want[i] {
					t.Fatalf("targets = %v, want %v", got, c.want)
				}
			}
			if why != "" {
				t.Errorf("resolved but still gave a reason: %q", why)
			}
		})
	}
}

// A root watch was MEASURED to catch the hooks' own writes and re-trigger them
// ten times over. Refusing it is a rule, not a preference, and it must say so.
func TestWatchTargetsRefusesTheProjectRoot(t *testing.T) {
	dir := realProject(t)
	for _, prose := range []string{"any .go edit", "**/*.go", "./"} {
		got, why := targets(t, dir, prose)
		for _, p := range got {
			if p == "." || p == "" {
				t.Fatalf("%q registered the project root: %v", prose, got)
			}
		}
		if len(got) == 0 && why == "" {
			t.Errorf("%q resolved to nothing with no reason given", prose)
		}
	}
}

// Unresolvable is DATA. A surface expressed as prose has no path, and a session
// that silently watched nothing must not look like one that watched everything.
func TestWatchTargetsReportsWhyItFoundNothing(t *testing.T) {
	dir := realProject(t)
	got, why := targets(t, dir, "a new seal")
	if len(got) != 0 {
		t.Fatalf("targets = %v, want none", got)
	}
	if why == "" {
		t.Error("silent failure — the gap must be visible")
	}
	if _, why := targets(t, dir, ""); why == "" {
		t.Error("an absent surface must also report why")
	}
}

// A checkpoint must not be able to point the watcher outside the project.
func TestWatchTargetsCannotEscapeTheProject(t *testing.T) {
	dir := realProject(t)
	got, _ := targets(t, dir, "../../etc or ../secrets/")
	if len(got) != 0 {
		t.Errorf("escaped the project dir: %v", got)
	}
}

// THE reason containment is os.Root and not path arithmetic.
//
// A lexical check — filepath.Rel, or the strings.HasPrefix it replaced — is
// blind to symlinks by construction. Measured directly: with proj/watched a
// symlink to an outside tree, Rel reports "contained: true" while os.Root
// reports "path escapes from parent". The note is agent-written, so a surface
// that resolves through a symlink is exactly the case that matters.
func TestWatchTargetsRefusesASymlinkEscape(t *testing.T) {
	dir := realProject(t)
	outside := t.TempDir()
	if err := os.MkdirAll(filepath.Join(outside, "secrets"), 0o755); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "watched")
	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("symlinks unavailable here: %v", err)
	}
	// Sanity: the escape is real, and a LEXICAL check would wave it through.
	if _, err := os.Stat(link); err != nil {
		t.Fatalf("fixture symlink is not traversable: %v", err)
	}
	rel, err := filepath.Rel(dir, filepath.Join(dir, "watched"))
	if err != nil || strings.HasPrefix(rel, "..") {
		t.Fatalf("fixture does not exercise the lexical blind spot (rel=%q)", rel)
	}

	got, why := targets(t, dir, "anything under watched/")
	if len(got) != 0 {
		t.Errorf("registered a watch that escapes via a symlink: %v", got)
	}
	if why == "" {
		t.Error("refused the surface without saying so")
	}
}

// Separator handling: targets come back in slash form so the re-arm hook can
// compare them against a slash-normalised file_path on every platform.
func TestWatchTargetsReturnSlashSeparatedPaths(t *testing.T) {
	dir := realProject(t)
	if err := os.MkdirAll(filepath.Join(dir, "a", "b"), 0o755); err != nil {
		t.Fatal(err)
	}
	got, _ := targets(t, dir, "a/b/*.go")
	if len(got) != 1 || got[0] != "a/b" {
		t.Fatalf("targets = %v, want [a/b] in slash form", got)
	}
	if strings.Contains(got[0], "\\") {
		t.Errorf("target %q carries a backslash; the re-arm hook compares slash form", got[0])
	}
}

// A project directory that cannot be opened must admit NOTHING. Refusing to
// watch is the safe failure; watching something unverified is not.
func TestUnopenableProjectAdmitsNothing(t *testing.T) {
	got, why := targets(t, filepath.Join(t.TempDir(), "does-not-exist"), "tools/")
	if len(got) != 0 {
		t.Errorf("registered %v against a project dir that cannot be opened", got)
	}
	if why == "" {
		t.Error("failed silently")
	}
}

func TestRearmStateRoundTripsAndSorts(t *testing.T) {
	raw := []byte(`{"rearmed":{"2":{"index":2,"check":"b","by":"x.go","event":"change","at":"t"},` +
		`"1":{"index":1,"check":"a","by":"y.go","event":"add","at":"t"}}}`)
	s := LoadRearm(func(string) ([]byte, error) { return raw, nil }, "ignored")
	got := s.Sorted()
	if len(got) != 2 || got[0].Index != 1 || got[1].Index != 2 {
		t.Fatalf("sorted = %+v, want ascending by index", got)
	}
}

// Losing re-arm history must never cost a session start.
func TestRearmStateDegradesToEmpty(t *testing.T) {
	for _, c := range []struct {
		name string
		read func(string) ([]byte, error)
	}{
		{"missing file", func(string) ([]byte, error) { return nil, os.ErrNotExist }},
		{"corrupt json", func(string) ([]byte, error) { return []byte("{not json"), nil }},
		{"null map", func(string) ([]byte, error) { return []byte(`{"rearmed":null}`), nil }},
	} {
		t.Run(c.name, func(t *testing.T) {
			s := LoadRearm(c.read, "p")
			if s.Rearmed == nil {
				t.Error("nil map — callers would panic on write")
			}
			if len(s.Sorted()) != 0 {
				t.Error("invented state from unreadable input")
			}
		})
	}
}
