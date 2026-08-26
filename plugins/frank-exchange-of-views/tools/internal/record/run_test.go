package record

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// existingRun makes a run directory that has actually been dispatched — the records directory
// is what distinguishes one from any other path.
func existingRun(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "records"), 0o755); err != nil {
		t.Fatal(err)
	}
	return dir
}

// A MISTYPED --run IS REFUSED, NOT REPORTED AS AN EMPTY RUN.
//
// This is #526 as a test. RecordsDir answers where a run's events WOULD live, so a path nobody
// dispatched resolves exactly like one that was — and the read then succeeds and prints a board
// of zeros for a directory that never existed. An empty board is the one answer that is never
// true, so the handle refuses to be constructed at all.
func TestOpenRunRefusesAPathThatNamesNoRun(t *testing.T) {
	typo := filepath.Join(t.TempDir(), "reserach", "2026-01-01_typo")
	r, err := OpenRun(typo)
	if err == nil {
		t.Fatalf("a path nobody dispatched was accepted as a run: %+v", r)
	}
	if r.Valid() {
		t.Error("the refused handle is valid, so a caller ignoring the error would still read")
	}
	for _, want := range []string{typo, "names no run", "never true"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal must carry %q so an operator can see which path: %v", want, err)
		}
	}
}

// The empty string is the state this type exists to stop carrying: it means "nobody supplied a
// run" and must not resolve to anything, least of all the working directory.
func TestOpenRunRefusesTheEmptyString(t *testing.T) {
	for _, in := range []string{"", "   ", "\t"} {
		if r, err := OpenRun(in); err == nil {
			t.Errorf("%q was accepted as a run: %+v", in, r)
		}
	}
}

// A real run opens, and carries its resolved record directory rather than re-deriving it.
func TestOpenRunResolvesOnce(t *testing.T) {
	dir := existingRun(t)
	r, err := OpenRun(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !r.Valid() {
		t.Fatal("a resolved run reports itself invalid")
	}
	if !filepath.IsAbs(r.Dir()) {
		t.Errorf("Dir() = %q, want an absolute path — two spellings of one run must not be two runs", r.Dir())
	}
	if got, want := r.Records(), filepath.Join(r.Dir(), "records"); got != want {
		t.Errorf("Records() = %q, want %q", got, want)
	}
}

// TWO SPELLINGS OF ONE RUN ARE ONE RUN. #358 keyed a cache on the resolved directory for
// exactly this reason; carrying the resolution in the handle makes it a property of the type
// rather than a rule each caller has to remember.
func TestTwoSpellingsResolveToTheSameRun(t *testing.T) {
	dir := existingRun(t)
	a, err := OpenRun(dir)
	if err != nil {
		t.Fatal(err)
	}
	// The same run, spelled with a redundant segment.
	b, err := OpenRun(filepath.Join(dir, "records", ".."))
	if err != nil {
		t.Fatal(err)
	}
	if a.Dir() != b.Dir() || a.Records() != b.Records() {
		t.Errorf("two spellings gave two runs:\n  %s -> %s\n  %s -> %s", dir, a.Dir(), filepath.Join(dir, "records", ".."), b.Dir())
	}
}

// NewRun is for the one caller that legitimately holds a run before it has a record. It must
// NOT accept the empty string either — creating a run is still not a reason to resolve nothing.
func TestNewRunAllowsARunWithNoRecordYet(t *testing.T) {
	fresh := filepath.Join(t.TempDir(), "research", "2026-01-01_new")
	if err := os.MkdirAll(fresh, 0o755); err != nil {
		t.Fatal(err)
	}
	r, err := NewRun(fresh)
	if err != nil {
		t.Fatalf("setup must be able to hold a run before its record exists: %v", err)
	}
	if !r.Valid() {
		t.Error("a run being created is still a resolved run")
	}
	if _, err := NewRun(""); err == nil {
		t.Error("NewRun accepted the empty string; creation is not a reason to resolve nothing")
	}
	// And the same path is refused by OpenRun, which is the distinction the two constructors
	// exist to draw.
	if _, err := OpenRun(fresh); err == nil {
		t.Error("OpenRun accepted a run with no record — that is the check every reader depends on")
	}
}

// The zero value is not a run, and says so when printed. A Run that reaches a message without
// having been resolved must not read as a plausible path.
func TestTheZeroValueIsNotARun(t *testing.T) {
	var r Run
	if r.Valid() {
		t.Error("the zero Run reports itself valid")
	}
	if r.Dir() != "" || r.Records() != "" {
		t.Errorf("the zero Run carries paths: %q %q", r.Dir(), r.Records())
	}
	if got := r.String(); got != "<unresolved run>" {
		t.Errorf("String() = %q; an unresolved run must not print as a path", got)
	}
}
