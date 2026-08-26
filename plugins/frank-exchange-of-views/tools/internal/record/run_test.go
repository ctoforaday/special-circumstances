package record

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
)

// existingRun makes a run directory that has actually been dispatched. The DIRECTORY is what
// distinguishes a real run from any other path; the records inside it are a separate question,
// and one a run legitimately answers with "none yet".
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

// A DISPATCHED RUN WITH NO RECORDS YET IS A REAL RUN, AND ITS EMPTY BOARD IS TRUE.
//
// This is the line #526 is actually about, and the first draft of OpenRun drew it in the wrong
// place. Requiring a records/ directory conflates two states that could not be less alike:
//
//   - never dispatched — a typo. Nothing made this path. An empty board is a fiction.
//   - dispatched, nothing filed yet — round 0, before the first claim. An empty board is the
//     honest answer, and refusing it would refuse every run its opening move.
//
// A write CREATES records/ (store.go's MkdirAll), so the second state has no records directory
// through no fault of its own. Measured when this was checked the other way: five tests across
// fetch, friction and the operator read failed on runs that were entirely real.
func TestARunWithNoRecordsYetOpens(t *testing.T) {
	dispatched := filepath.Join(t.TempDir(), "research", "2026-01-01_fresh")
	if err := os.MkdirAll(dispatched, 0o755); err != nil {
		t.Fatal(err)
	}
	r, err := OpenRun(dispatched)
	if err != nil {
		t.Fatalf("a dispatched run must open before its first write — that is round 0: %v", err)
	}
	if !r.Valid() {
		t.Error("a dispatched run reports itself invalid")
	}
}

// NewRun is for the one caller that holds a run before there is anything on disk to open. It
// must NOT accept the empty string either — creating a run is still not a reason to resolve
// nothing.
func TestNewRunAllowsARunNotYetOnDisk(t *testing.T) {
	unmade := filepath.Join(t.TempDir(), "research", "2026-01-01_new")
	r, err := NewRun(unmade)
	if err != nil {
		t.Fatalf("setup must be able to resolve a run it is about to create: %v", err)
	}
	if !r.Valid() {
		t.Error("a run being created is still a resolved run")
	}
	if _, err := NewRun(""); err == nil {
		t.Error("NewRun accepted the empty string; creation is not a reason to resolve nothing")
	}
	// And the same path is refused by OpenRun, which is the distinction the two constructors
	// exist to draw.
	if _, err := OpenRun(unmade); err == nil {
		t.Error("OpenRun accepted a path nobody made — that is the check every reader depends on")
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

// THE REFUSALS CARRY A CODE, because the --json edge reads one.
//
// feov.CodeOf returns the literal "error" for any error with no *feov.Error in its chain, so an
// uncoded refusal does not fail loudly — it reports a generic code that is indistinguishable
// from every other uncoded failure. The sites these two refusals replaced returned
// feov.MissingField and a coded Conflict, so leaving them uncoded would have retired a
// distinction the structured surface exposes, without retiring anything that says so.
func TestTheRefusalsAreCoded(t *testing.T) {
	if _, err := OpenRun(""); feov.CodeOf(err) != string(feov.MissingField) {
		t.Errorf("an unsupplied run codes as %q, want %q", feov.CodeOf(err), feov.MissingField)
	}
	unmade := filepath.Join(t.TempDir(), "research", "2026-01-01_nothing-here")
	if _, err := OpenRun(unmade); feov.CodeOf(err) != string(feov.NotFound) {
		t.Errorf("a path nobody dispatched codes as %q, want %q", feov.CodeOf(err), feov.NotFound)
	}
}
