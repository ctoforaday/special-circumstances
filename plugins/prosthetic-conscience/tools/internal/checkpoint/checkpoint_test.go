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

// The frontmatter lines that are NOT keys. Parse's job is to degrade rather than fail
// — the note is agent-written under time pressure — but degrading must not mean
// inventing keys, because the restore hook states this map back to the session as
// recovered fact. A mutation sweep flipped the `k != "" && !HasPrefix(k, "#")` guard
// with the suite still green: neither half was asserted by anything.
func TestFrontmatterSkipsNonKeys(t *testing.T) {
	n := Parse("---\n" +
		"# objective: this is a comment, not a key\n" +
		"  # indented: also a comment\n" +
		": value with no key\n" +
		"   : whitespace-only key\n" +
		"no colon at all\n" +
		"real: kept\n" +
		"---\nbody\n")

	if got := n.Get("real"); got != "kept" {
		t.Errorf("a well-formed key must survive its malformed neighbours: %q", got)
	}
	for _, k := range []string{"# objective", "#objective", "objective", "", "#", "no colon at all"} {
		if got := n.Get(k); got != "" {
			t.Errorf("Parse invented key %q = %q — a commented-out or keyless line is not a fact the note recorded", k, got)
		}
	}
	if len(n.Front) != 1 {
		t.Errorf("frontmatter has %d keys, want exactly 1 (%v)", len(n.Front), n.Front)
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

// What OPENS a check, at the character level. The loop above is well-formed, so the
// numbered() predicate was only ever asked easy questions: a mutation sweep flipped
// every comparison in it — the digit range and the `i > 0 && i < len(t)` guard — with
// the suite still green. This is the note's schema boundary, and the note is
// hand-written under time pressure, which is precisely when it is malformed.
func TestValidationLoopNumbering(t *testing.T) {
	cases := []struct {
		name  string
		line  string
		opens bool
	}{
		{"single digit", "1. `go test`", true},
		{"multi digit", "12. `go test`", true},
		// '9' is the top of the digit range and '0' the bottom; a range check that
		// excludes either end reads a two-digit loop as prose. Both ends get a case.
		{"nine", "9. `go test`", true},
		{"nineteen", "19. `go test`", true},
		{"indented", "   3. `go test`", true},
		{"zero", "0. `go test`", true},
		{"no dot", "1 go test", false},
		{"dot with no digits", ". go test", false},
		{"letter before the dot", "a. go test", false},
		{"digit-adjacent but not a dot", "1x go test", false},
		{"prose that merely contains a number", "we ran 3 checks. all passed", false},
		{"bare digits, nothing after", "12", false},
		{"digits then end-of-line dot only", "12.", true},
		{"empty", "", false},
		{"whitespace only", "   ", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := ParseValidationLoop(c.line)
			if opened := len(got) == 1; opened != c.opens {
				t.Errorf("ParseValidationLoop(%q) opened %d check(s); want opens=%v", c.line, len(got), c.opens)
			}
		})
	}

	// A re-arm marker BEFORE any numbered line has no check to belong to, and must
	// not be silently attached to the first one that follows.
	got := ParseValidationLoop("re-armed by: tools/\n1. `go test`\n")
	if len(got) != 1 {
		t.Fatalf("checks = %d, want 1", len(got))
	}
	if got[0].ReArmedBy != "" {
		t.Errorf("a marker preceding every check was adopted by check 1: %q", got[0].ReArmedBy)
	}

	// Case-insensitive by contract — the schema writes "re-armed by:" but a
	// hand-written note capitalises.
	if got := ParseValidationLoop("1. `go test` · Re-Armed By: tools/\n"); got[0].ReArmedBy != "tools/" {
		t.Errorf("marker must match case-insensitively, got %q", got[0].ReArmedBy)
	}

	// The marker at column ZERO of a continuation line. Every other case here has it
	// mid-line or indented, so `i >= 0` was never distinguished from `i > 0` — under
	// which an unindented continuation loses its surface silently, and a surface that
	// is silently lost is the failure the whole re-arm path exists to prevent.
	unindented := "1. `go test`\nre-armed by: tools/\n"
	if got := ParseValidationLoop(unindented)[0].ReArmedBy; got != "tools/" {
		t.Errorf("a marker at column 0 of a continuation line was dropped: %q", got)
	}

	// A "·" at position zero of the remainder — the clause separator with nothing
	// before it. There is no surface here, and the answer must be "none recorded"
	// rather than the separator and the following prose read as a path.
	if got := ParseValidationLoop("1. `go test` · re-armed by: · a human deciding to ship\n")[0].ReArmedBy; got != "" {
		t.Errorf("an empty surface before the · clause must record nothing, got %q", got)
	}

	// First marker wins: a check whose surface is stated twice keeps the first,
	// rather than letting a later continuation line overwrite it.
	twice := "1. `go test` · re-armed by: tools/\n   re-armed by: docs/\n"
	if got := ParseValidationLoop(twice)[0].ReArmedBy; got != "tools/" {
		t.Errorf("second marker overwrote the first: %q", got)
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

// An ABSOLUTE path in the prose. The relative escapes above are covered; this one was
// not, and it takes a different route through the code: watchableDir drops the empty
// leading segment, so "/etc/passwd" arrives at the containment check already looking
// relative. The IsAbs pre-filter therefore never fires for it and containment is the
// only thing standing there — which is fine, and is exactly why it has to be asserted
// rather than assumed. The contract is about the OUTPUT: no absolute path, and nothing
// the watcher could aim outside the project.
func TestWatchTargetsRefusesAbsolutePaths(t *testing.T) {
	dir := realProject(t)
	for _, prose := range []string{
		"/etc/passwd",
		"any edit under /var/log/",
		"C:/Windows/System32",
		"//server/share/x",
	} {
		got, reason := targets(t, dir, prose)
		for _, p := range got {
			if filepath.IsAbs(p) {
				t.Errorf("targets(%q) = %v — watchPaths must never be handed an absolute path", prose, got)
			}
			if _, err := os.Stat(filepath.Join(dir, p)); err != nil {
				t.Errorf("targets(%q) = %v, which does not exist inside the project", prose, got)
			}
		}
		if len(got) == 0 && reason == "" {
			t.Errorf("targets(%q) resolved nothing and said nothing — an unresolved surface must be REPORTED, or it is a silent miss", prose)
		}
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
	s, err := LoadRearm(func(string) ([]byte, error) { return raw, nil }, "ignored")
	if err != nil {
		t.Fatalf("LoadRearm: %v", err)
	}
	got := s.Sorted()
	if len(got) != 2 || got[0].Index != 1 || got[1].Index != 2 {
		t.Fatalf("sorted = %+v, want ascending by index", got)
	}
}

// A MISSING file is an empty state and no error — the first run has no history.
// This is the only case that may degrade silently.
func TestAMissingStateFileIsNotAnError(t *testing.T) {
	s, err := LoadRearm(func(string) ([]byte, error) { return nil, os.ErrNotExist }, "p")
	if err != nil {
		t.Fatalf("a missing file must not error: %v", err)
	}
	if s.Rearmed == nil {
		t.Error("nil map — callers would panic on write")
	}
}

// A file that EXISTS but cannot be read is an error, and the reason is #165.
//
// The old contract returned an empty state here, and the caller's next save
// wrote a single fresh key. So any damage to this file silently reset coverage
// to one lone record beside surfaces demonstrably being edited — indistinguishable
// from the bug the file exists to report on. It cost three sessions.
func TestAnUnreadableStateFileIsAnErrorSoNothingOverwritesIt(t *testing.T) {
	for _, c := range []struct {
		name string
		read func(string) ([]byte, error)
	}{
		{"corrupt json", func(string) ([]byte, error) { return []byte("{not json"), nil }},
		{"null map", func(string) ([]byte, error) { return []byte(`{"rearmed":null}`), nil }},
		{"permission denied", func(string) ([]byte, error) { return nil, os.ErrPermission }},
	} {
		t.Run(c.name, func(t *testing.T) {
			s, err := LoadRearm(c.read, "p")
			if err == nil {
				t.Fatal("returned an empty state instead of an error; the next save would silently reset coverage")
			}
			if !strings.Contains(err.Error(), "refusing to overwrite") {
				t.Errorf("the error must tell the reader nothing was destroyed: %v", err)
			}
			if s.Rearmed == nil {
				t.Error("nil map — a caller that ignores the error would panic")
			}
		})
	}
}

// Keyed by identity, not position. A renumbered note used to orphan every record.
func TestRecordsSurviveARenumberBecauseTheyAreKeyedByCheckText(t *testing.T) {
	// Written when the check was numbered 2.
	raw := []byte(`{"rearmed":{"2":{"index":2,"check":"2. ` + "`go test ./b`" + ` · re-armed by: b/","by":"b/x.go","event":"change","at":"t"}}}`)
	s, err := LoadRearm(func(string) ([]byte, error) { return raw, nil }, "p")
	if err != nil {
		t.Fatalf("LoadRearm: %v", err)
	}
	// The same check, now numbered 5 after entries were inserted above it.
	key := CheckKey("5. `go test ./b` · re-armed by: b/")
	if _, ok := s.Rearmed[key]; !ok {
		t.Fatalf("the record did not survive a renumber; keys are %v, wanted %q", keysOf(s), key)
	}
}

func TestCheckKeyIgnoresTheLabelAndOnlyTheLabel(t *testing.T) {
	a := CheckKey("1. `go test ./a`  · re-armed by: a/")
	b := CheckKey("11. `go test ./a`  · re-armed by: a/")
	if a != b {
		t.Errorf("label affected identity: %q vs %q", a, b)
	}
	if a == "" {
		t.Fatal("identity is empty — every check would collide on one key")
	}
	if CheckKey("1. `go test ./a`") == CheckKey("1. `go test ./b`") {
		t.Error("two different checks share an identity")
	}
}

func keysOf(s RearmState) []string {
	out := make([]string, 0, len(s.Rearmed))
	for k := range s.Rearmed {
		out = append(out, k)
	}
	return out
}

// #215: release must only remove a lock we still hold. The stale-breaker cannot tell an
// abandoned lock from a slow one, so a displaced holder used to delete its SUCCESSOR's lock
// — after which a third hook could acquire while the second was mid-update.
func TestReleaseOnlyRemovesItsOwnLock(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rearmed.json")

	a, err := lockRearm(path)
	if err != nil {
		t.Fatal(err)
	}
	// Another hook breaks A's lock as stale and takes its own.
	if err := os.Remove(a.path); err != nil {
		t.Fatal(err)
	}
	b, err := lockRearm(path)
	if err != nil {
		t.Fatal(err)
	}

	a.release() // A finishes, unaware it was displaced

	if !b.owned() {
		t.Error("A's release deleted B's lock — that is the mutual-exclusion violation #215 reported")
	}
	b.release()
	if _, err := os.Stat(b.path); !os.IsNotExist(err) {
		t.Error("B's own release must remove its lock")
	}
}

// A displaced holder must ABANDON its write rather than overwrite the successor's. Losing
// an event is the documented cheap failure; a lost update is the bug the lock exists for.
func TestUpdateAbandonsWhenItsLockWasBroken(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rearmed.json")

	// Seed a record so there is something to lose.
	if err := UpdateRearm(path, func(s *RearmState) {
		s.Rearmed[CheckKey("1. `go build` → ok")] = Rearm{Check: "1. `go build` → ok"}
	}); err != nil {
		t.Fatal(err)
	}

	err := UpdateRearm(path, func(s *RearmState) {
		// Mid-update, another hook breaks this lock as stale and takes its own.
		_ = os.Remove(path + ".lock")
		if _, err := lockRearm(path); err != nil {
			t.Fatal(err)
		}
		s.Rearmed["clobber"] = Rearm{Check: "should never be written"}
	})
	if err == nil {
		t.Fatal("a displaced holder must refuse to commit, not write over the successor")
	}
	if !strings.Contains(err.Error(), "nothing is damaged") {
		t.Errorf("the error must say the cost is one lost event: %v", err)
	}

	s, err := LoadRearm(os.ReadFile, path)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := s.Rearmed["clobber"]; ok {
		t.Error("the abandoned update was written anyway")
	}
	if _, ok := s.Rearmed[CheckKey("1. `go build` → ok")]; !ok {
		t.Error("the pre-existing record was lost")
	}
}

// #217: a record whose check no longer exists is dropped — but ONLY when the caller
// actually parsed a loop. An empty slice means "nothing parsed", never "prune everything".
func TestPruneOrphansNeedsAParsedLoop(t *testing.T) {
	// The orphan is a check whose COMMAND is gone. Two WORDINGS of one command are no longer
	// two records at all — see TestTwoWordingsOfOneCommandAreOneRecord, which is the case this
	// fixture used to carry and which #432 turned from an orphan into a match.
	seed := func() *RearmState {
		return &RearmState{Rearmed: map[string]Rearm{
			CheckKey("1. `go test ./...` → ok"): {Check: "1. `go test ./...` → ok"},
			CheckKey("2. `go vet ./...` → ok"):  {Check: "2. `go vet ./...` → ok"},
		}}
	}
	live := []Check{{Line: "1. `go test ./...` → ok"}}

	s := seed()
	dropped := s.PruneOrphans(live)
	if len(dropped) != 1 {
		t.Fatalf("dropped = %v; want exactly the orphaned wording", dropped)
	}
	if len(s.Rearmed) != 1 {
		t.Errorf("state = %v; the current check's record must survive", s.Rearmed)
	}
	if _, ok := s.Rearmed[CheckKey("1. `go test ./...` → ok")]; !ok {
		t.Error("the wrong record was dropped")
	}

	// THE GUARD. A note that failed to parse yields no checks; treating that as "no checks
	// exist" would wipe the history rather than tidy it.
	s = seed()
	if dropped := s.PruneOrphans(nil); dropped != nil || len(s.Rearmed) != 2 {
		t.Errorf("an unparsed loop must delete nothing: dropped=%v left=%d", dropped, len(s.Rearmed))
	}
	s = seed()
	if dropped := s.PruneOrphans([]Check{}); dropped != nil || len(s.Rearmed) != 2 {
		t.Errorf("an empty slice must delete nothing: dropped=%v left=%d", dropped, len(s.Rearmed))
	}

	// Renumbering must NOT orphan: identity strips the ordinal, which is why #213 keyed
	// records this way in the first place.
	s = seed()
	if dropped := s.PruneOrphans([]Check{{Line: "7. `go test ./...` → ok"}}); len(dropped) != 1 {
		t.Errorf("a renumbered check must still match its record: dropped=%v", dropped)
	}
}

// NoteLoopProblems exists to make one mistake impossible: LoopProblems takes the SECTION,
// and the note's OTHER sections are numbered too. `## Next intended steps` is a numbered
// list in the schema, so handing the whole note to LoopProblems reports every step after
// the first as a label disagreeing with its ordinal — confidently, about the wrong section.
func TestNoteLoopProblemsReadsOnlyTheLoopSection(t *testing.T) {
	note := "---\nschema: 2\n---\n" +
		"## Validation loop\n1. `go test ./...` → ok · re-armed by: tools/\n" +
		"## Next intended steps\n1. file the issue\n2. take the issue\n3. close the issue\n" +
		"## Open threads\n1. the other thing\n"

	if got := NoteLoopProblems(note); len(got) != 0 {
		t.Errorf("NoteLoopProblems = %v; numbered lines outside the loop are not loop entries", got)
	}
	// Proof the guard is load-bearing rather than vacuous: the raw call DOES misfire on
	// this note, which is the reason the wrapper exists.
	if got := LoopProblems(note); len(got) == 0 {
		t.Error("the whole-note call no longer misfires — if that is now safe, this wrapper's reason is gone and the test is asserting nothing")
	}

	// And a real problem inside the loop still surfaces through the wrapper.
	bad := "---\nschema: 2\n---\n" +
		"## Validation loop\n1. `a` · re-armed by: tools/\n7b. `b` · re-armed by: tools/\n" +
		"## Next intended steps\n1. only one step\n"
	got := NoteLoopProblems(bad)
	if len(got) != 1 || !strings.Contains(got[0], "7b.") {
		t.Errorf("NoteLoopProblems = %v; want the lettered entry reported", got)
	}

	// No loop section is not a loop problem — an absent loop is a different complaint,
	// raised where something can be done about it.
	if got := NoteLoopProblems("---\nschema: 2\n---\n## Open threads\n1. a\n2. b\n"); got != nil {
		t.Errorf("NoteLoopProblems = %v; a note with no loop has no loop problems to report", got)
	}
}

// #219 was not a missing feature — it was a sweep that stopped at the parsers. LoopProblems
// was written for #192/#193 and wired into ONE of the three packages that parse the loop, so
// a malformed note was reported at session start and nowhere else: not by the seal that
// wrote it, not by the hook whose re-arms it was silently costing.
//
// Detection and surfacing are different sets, and the fix for a class is not three edits —
// it is the thing that fails when a fourth site appears. This walks the real source.
//
// Following hookenv's guard, which passed VACUOUSLY for a while after a refactor moved the
// files it globbed: the count assertion is what makes that survivable.
func TestEveryPackageThatParsesTheLoopAlsoReportsItsProblems(t *testing.T) {
	sources, err := filepath.Glob(filepath.Join("..", "*", "main.go"))
	if err != nil {
		t.Fatalf("walk failed: %v", err)
	}
	if len(sources) < 8 {
		t.Fatalf("walked only %d hook sources (%v) — a broken walk passes this test silently forever", len(sources), sources)
	}

	var parsers, offenders []string
	for _, path := range sources {
		src, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		body := string(src)
		if !strings.Contains(body, "ParseValidationLoop(") {
			continue
		}
		parsers = append(parsers, path)
		if !strings.Contains(body, "LoopProblems(") {
			offenders = append(offenders, path)
		}
	}

	// checkpointseal, checkpointrestore, filechangedrearm. Fewer means the walk broke.
	if len(parsers) < 3 {
		t.Fatalf("found only %d packages parsing the validation loop (%v) — there are three, so the walk is not seeing the source", len(parsers), parsers)
	}
	if len(offenders) > 0 {
		t.Errorf("packages that parse the validation loop but never report what the parse DROPPED:\n  %s\n\nParseValidationLoop silently declines entries a human reader counts — a lettered label opens no check, and a mislabelled ordinal makes state keyed by position disagree with the labels in the note. A site that parses without reporting turns those into a check that is never watched, never re-armed, and never mentioned. Call checkpoint.NoteLoopProblems(note) — or LoopProblems(loop) if you already hold the section — and write the result to stderr. Never refuse: these are hooks, and a numbering slip must not cost a seal or block a tool call.",
			strings.Join(offenders, "\n  "))
	}
}

// #432 friction 2 — THE CASE THAT DROVE THE KEY CHANGE.
//
// The live state held FOURTEEN records for TWELVE checks. In four of the five duplicated
// pairs the command was byte-identical and only the annotation had moved:
//
//	→ 17 pkgs ok  vs  → 18 ok        → 15 pkgs ok  vs  → 21 ok        → 16 files valid  vs  → 16 valid
//
// The old whole-line key called each of those a different check. It also taught the note's
// authors to stop correcting stale counts, because a correction orphaned the record — which
// is how the loop came to document "18 ok" for a thirty-package module ON PURPOSE.
func TestTwoWordingsOfOneCommandAreOneRecord(t *testing.T) {
	a := CheckKey("3. `go test -C plugins/frank-exchange-of-views/tools ./...` → 17 pkgs ok · re-armed by: x/")
	b := CheckKey("3. `go test -C plugins/frank-exchange-of-views/tools ./...` → 18 ok · re-armed by: x/")
	if a != b {
		t.Fatalf("correcting a stale count changed the identity:\n  %q\n  %q", a, b)
	}
	if a != "go test -C plugins/frank-exchange-of-views/tools ./..." {
		t.Errorf("key = %q; the identity is the command, not the sentence around it", a)
	}
}

// THE OTHER HALF, and it must not be lost in the trade: changing what the check RUNS is a
// different check and must still orphan. Measured pair — the goldens entry was corrected from
// the REGENERATE command to the review command, and that record SHOULD not carry over.
func TestChangingTheCommandStillChangesTheIdentity(t *testing.T) {
	old := CheckKey("11. Goldens: `UPDATE_GOLDENS=1 go test ./internal/difftest` — never by hand")
	cur := CheckKey("11. Goldens: `go run ./golden -review` (from scripts/) · re-armed by: x/")
	if old == cur {
		t.Fatal("a check that runs a different command must not inherit the old record")
	}
}

// A check with no command span still gets an identity rather than an empty key — every
// record collapsing onto "" would be the silent merge this whole design is against.
func TestAProseCheckFallsBackToItsWholeLine(t *testing.T) {
	k := CheckKey("10. last_event in rearmed.json advances while work continues")
	if k == "" {
		t.Fatal("a prose check keyed to the empty string would merge with every other prose check")
	}
	if k != "last_event in rearmed.json advances while work continues" {
		t.Errorf("key = %q; want the whole line with the ordinal stripped", k)
	}
	// An unterminated backtick is prose, not a command — it must not swallow the rest of the
	// line as an identity.
	if got := CheckKey("1. a stray ` backtick and then words"); got != "a stray ` backtick and then words" {
		t.Errorf("unterminated span: key = %q, want the whole line", got)
	}
}

// THE HAZARD THE CHANGE INTRODUCES, held explicitly. Keying on the command means two checks
// running one command are one identity. That must be LOUD and must stop the prune, because a
// record standing for two checks makes "orphaned" meaningless.
func TestASharedCommandIsReportedAndStopsThePrune(t *testing.T) {
	live := []Check{
		{Line: "1. `go test ./...` → the fast leg"},
		{Line: "2. `go test ./...` → the slow leg"},
		{Line: "3. `go vet ./...` → ok"},
	}
	dup := CheckKeyCollisions(live)
	if len(dup) != 1 || dup[0] != "go test ./..." {
		t.Fatalf("collisions = %v; want exactly the shared command", dup)
	}

	s := &RearmState{Rearmed: map[string]Rearm{
		"go test ./...": {Check: "1. `go test ./...` → the fast leg"},
		"an old orphan": {Check: "9. `gone` → removed"},
		"go vet ./...":  {Check: "3. `go vet ./...` → ok"},
	}}
	if dropped := s.PruneOrphans(live); dropped != nil {
		t.Errorf("dropped %v while an identity was ambiguous; deletion must wait for the ambiguity to be resolved", dropped)
	}
	if len(s.Rearmed) != 3 {
		t.Errorf("state = %v; nothing may be deleted under a collision", s.Rearmed)
	}

	// No collision, no excuse: the orphan goes.
	clean := []Check{{Line: "3. `go vet ./...` → ok"}}
	s2 := &RearmState{Rearmed: map[string]Rearm{
		"go vet ./...":  {Check: "3. `go vet ./...` → ok"},
		"an old orphan": {Check: "9. `gone` → removed"},
	}}
	if dropped := s2.PruneOrphans(clean); len(dropped) != 1 {
		t.Errorf("dropped = %v; an unambiguous orphan must still be pruned", dropped)
	}
}

// MIGRATION IS NOT A RESET. The live file already held both wordings; on load they collapse,
// and WHICH ONE SURVIVES cannot be Go's map seed. The newest wins, which is what RearmState
// has always claimed to hold.
func TestCollapsingTwoWordingsKeepsTheNewestRun(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rearmed.json")
	body := `{"rearmed":{
	  "` + "`go test ./...` → 17 pkgs ok" + `":{"index":3,"check":"3. ` + "`go test ./...`" + ` → 17 pkgs ok","at":"2026-08-01T00:00:00Z"},
	  "` + "`go test ./...` → 18 ok" + `":{"index":3,"check":"3. ` + "`go test ./...`" + ` → 18 ok","at":"2026-08-15T00:00:00Z"}
	}}`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	s, err := LoadRearm(os.ReadFile, path)
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Rearmed) != 1 {
		t.Fatalf("state = %v; two wordings of one command must collapse to one record", s.Rearmed)
	}
	got := s.Rearmed["go test ./..."]
	if got.At != "2026-08-15T00:00:00Z" {
		t.Errorf("survivor.At = %q; the NEWEST run must survive, not whichever the map yielded last", got.At)
	}
}
