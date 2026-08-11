package record

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// recordsDirT resolves for tests that lay a record directory out by hand. It panics rather than
// taking a *testing.T because several call sites are inside closures that have none.
func recordsDirT(runDir string) string {
	d, err := RecordsDir(runDir)
	if err != nil {
		panic(err)
	}
	return d
}

// isolate points the pointer cache at a temp directory and clears the declaring variable, so a
// test neither reads the developer's real cache nor inherits a root from the surrounding shell.
//
// NOT PARALLEL, AND THAT IS STRUCTURAL: both the cache seam and the environment are process-wide.
func isolate(t *testing.T) {
	t.Helper()
	cache := t.TempDir()
	prev := cacheDirFn
	cacheDirFn = func() (string, error) { return cache, nil }
	t.Cleanup(func() { cacheDirFn = prev })
	t.Setenv(RecordRootEnv, "")
}

func TestRecordsDirDefaultsUnderTheRun(t *testing.T) {
	isolate(t)
	run := t.TempDir()
	got, err := RecordsDir(run)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if want := filepath.Join(run, "records"); got != want {
		t.Fatalf("default root = %s, want %s", got, want)
	}
	// The separated-run machinery must stay entirely inert on an ordinary run: a marker here
	// would arm the lost-pointer refusal on a run whose record never moved.
	if _, err := os.Stat(filepath.Join(run, separatedMarker)); err == nil {
		t.Fatal("an unseparated run was given a " + separatedMarker + " marker")
	}
}

func TestDeclaringARootAdoptsItThenResolvesWithoutTheEnvironment(t *testing.T) {
	isolate(t)
	run, root := t.TempDir(), filepath.Join(t.TempDir(), "elsewhere")

	t.Setenv(RecordRootEnv, root)
	got, err := RecordsDir(run)
	if err != nil {
		t.Fatalf("adopt: %v", err)
	}
	if !samePath(got, root) {
		t.Fatalf("adopted %s, want %s", got, root)
	}

	// THE POINT OF THE POINTER. The harness runs boards concurrently in one process, so
	// resolution cannot depend on a process-global variable — clear it and the run must still
	// find its own root.
	t.Setenv(RecordRootEnv, "")
	got, err = RecordsDir(run)
	if err != nil {
		t.Fatalf("resolve after adopt: %v", err)
	}
	if !samePath(got, root) {
		t.Fatalf("resolved %s without the environment, want %s", got, root)
	}
}

func TestTheMarkerNamesNoPathBecauseTheSeatReadsIt(t *testing.T) {
	isolate(t)
	run, root := t.TempDir(), filepath.Join(t.TempDir(), "elsewhere")
	t.Setenv(RecordRootEnv, root)
	if _, err := RecordsDir(run); err != nil {
		t.Fatalf("adopt: %v", err)
	}

	b, err := os.ReadFile(filepath.Join(run, separatedMarker))
	if err != nil {
		t.Fatalf("no marker in the run directory: %v", err)
	}
	// The whole separation is defeated by a marker that helpfully includes the path — a seat
	// reads this directory, which is the entire reason the record left it.
	if strings.Contains(string(b), root) {
		t.Fatalf("the marker hands the record root back to anyone reading the run directory:\n%s", b)
	}
}

func TestALostPointerRefusesInsteadOfReportingAnEmptyBoard(t *testing.T) {
	isolate(t)
	run, root := t.TempDir(), filepath.Join(t.TempDir(), "elsewhere")
	t.Setenv(RecordRootEnv, root)
	if _, err := RecordsDir(run); err != nil {
		t.Fatalf("adopt: %v", err)
	}

	// A cleaned cache, or a run directory copied to another machine.
	ptr, err := rootPointerPath(mustAbs(t, run))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(ptr); err != nil {
		t.Fatal(err)
	}
	t.Setenv(RecordRootEnv, "")

	if _, err := RecordsDir(run); err == nil {
		t.Fatal("a separated run with no pointer resolved anyway — every projection would then " +
			"read an empty directory and render a board of zeros for a run that has a full record")
	}

	// And the refusal has to survive the whole way up: MergedEvents flattens IsNotExist into an
	// empty merge, which is the honest zero this must never be confused with.
	if _, err := MergedEvents(run); err == nil {
		t.Fatal("MergedEvents returned a clean empty merge for an unreachable record")
	}
}

func TestConflictingDeclarationsRefuse(t *testing.T) {
	isolate(t)
	run := t.TempDir()
	first := filepath.Join(t.TempDir(), "first")
	t.Setenv(RecordRootEnv, first)
	if _, err := RecordsDir(run); err != nil {
		t.Fatalf("adopt: %v", err)
	}

	// Preferring either silently writes half the run into each root.
	t.Setenv(RecordRootEnv, filepath.Join(t.TempDir(), "second"))
	if _, err := RecordsDir(run); err == nil {
		t.Fatal("a second root was accepted for a run that already has one")
	}
}

func TestTwoRunsCannotShareOneRoot(t *testing.T) {
	isolate(t)
	root := filepath.Join(t.TempDir(), "shared")
	t.Setenv(RecordRootEnv, root)
	if _, err := RecordsDir(t.TempDir()); err != nil {
		t.Fatalf("first run: %v", err)
	}
	// Their shards would merge into one board wearing both runs' history — and the shard
	// filenames carry seat ids, not run ids, so nothing downstream could tell them apart.
	if _, err := RecordsDir(t.TempDir()); err == nil {
		t.Fatal("a second run adopted a root that already belongs to another")
	}
}

func TestARunWithEventsInPlaceRefusesToSeparate(t *testing.T) {
	isolate(t)
	run := t.TempDir()
	if _, _, err := RegisterSeat(run, "blue-respond-r1"); err != nil {
		t.Fatalf("register: %v", err)
	}
	t.Setenv(RecordRootEnv, filepath.Join(t.TempDir(), "elsewhere"))
	if _, err := RecordsDir(run); err == nil {
		t.Fatal("a run with shards already on disk was separated, orphaning them where nothing reads them")
	}
}

func TestDeclaringTheDefaultPathChangesNothing(t *testing.T) {
	isolate(t)
	run := t.TempDir()
	t.Setenv(RecordRootEnv, filepath.Join(run, "records"))
	if _, err := RecordsDir(run); err != nil {
		t.Fatalf("resolve: %v", err)
	}
	// Spelling the default out longhand must not hand an ordinary run the machinery of a
	// separated one — it would start failing the moment the cache was cleaned.
	if _, err := os.Stat(filepath.Join(run, separatedMarker)); err == nil {
		t.Fatal("declaring the default path marked the run as separated")
	}
}

// TestASeparatedRunKeepsNoEventsUnderTheRun is the property the separation exists for. Everything
// above is a guard on the resolver; this is the behaviour a seat meets.
func TestASeparatedRunKeepsNoEventsUnderTheRun(t *testing.T) {
	isolate(t)
	run, root := t.TempDir(), filepath.Join(t.TempDir(), "elsewhere")
	t.Setenv(RecordRootEnv, root)

	if _, _, err := RegisterSeat(run, "blue-respond-r1"); err != nil {
		t.Fatalf("register: %v", err)
	}
	if _, err := Append(run, "blue-respond-r1", "friction", NewPayload().Set("note", "the verb I wanted was not there")); err != nil {
		t.Fatalf("append: %v", err)
	}
	t.Setenv(RecordRootEnv, "")

	// Nothing an `ls` of the run directory turns up is a record.
	ents, err := os.ReadDir(run)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range ents {
		if e.Name() == "records" {
			t.Fatal("a separated run still has a records/ directory — the seat's cheap path is open")
		}
	}

	// And the board still reads, through the tool, with no environment set.
	m, err := MergedEvents(run)
	if err != nil {
		t.Fatalf("merge: %v", err)
	}
	if len(m.Events) != 2 {
		t.Fatalf("merged %d events from the separated root, want 2 (register + friction)", len(m.Events))
	}
}

func mustAbs(t *testing.T, p string) string {
	t.Helper()
	a, err := filepath.Abs(p)
	if err != nil {
		t.Fatal(err)
	}
	return a
}

func TestADeletedRootRefusesInsteadOfReadingAsAnEmptyRun(t *testing.T) {
	isolate(t)
	run, root := t.TempDir(), filepath.Join(t.TempDir(), "elsewhere")
	t.Setenv(RecordRootEnv, root)
	if _, _, err := RegisterSeat(run, "blue-respond-r1"); err != nil {
		t.Fatalf("register: %v", err)
	}
	t.Setenv(RecordRootEnv, "")

	// What the seat probe does to every board it does not -keep.
	if err := os.RemoveAll(root); err != nil {
		t.Fatal(err)
	}
	if _, err := RecordsDir(run); err == nil {
		t.Fatal("a pointer to a deleted root resolved — MergedEvents would then take its IsNotExist " +
			"arm and report a run that had a full record as a board of zeros")
	}
	if _, err := MergedEvents(run); err == nil {
		t.Fatal("MergedEvents reported a clean empty merge for a record whose directory is gone")
	}
}
