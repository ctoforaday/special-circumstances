package record

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"google.golang.org/protobuf/proto"
	"os"
	"path/filepath"
	"regexp"
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

// THE MARKER MUST NOT COMPOSE A COMMAND IT CANNOT MAKE RUNNABLE.
//
// It used to open with `feov-record <role> show board --run <this directory>`. The tool is not
// on PATH — every dispatch hands the seat an absolute path — so that line was a command that
// cannot execute, written by the one file whose job is to unblock a seat that cannot see a board.
//
// Measured 2026-08-16, six blue seats at one board: three copied this form, two recovered by
// hunting for the binary, and one took `command not found` as "that tool isn't available" and
// answered the whole sitting against a board it had never read. That output is indistinguishable
// from a sitting that read the board — same structure, same confidence, gaps invented from the
// pattern corpus instead of quoted. A read that misses and a read that succeeds produced the
// same bytes, which is the class this repository keeps finding.
//
// The fix is not to write the path (the separation exists to withhold it) but to stop composing
// the invocation at all: name the shape, point at the handle the dispatch gave, and say out loud
// what `command not found` does and does not mean. This test holds that line — it fails on any
// bare occurrence of the binary name that is followed by something that reads as arguments, and
// passes the two places the text must still SAY the name: the quoted error string it is teaching
// the seat to interpret, and the `find` command for recovering the path.
func TestTheMarkerComposesNoInvocationTheSeatCannotRun(t *testing.T) {
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
	text := string(b)

	// The two legitimate mentions, removed before the scan so the check below is about
	// invocations only. Both are ABOUT the name rather than uses of it.
	allowed := []string{
		`"feov-record: command not found"`,   // the error the seat is being taught to read
		"`find / -name feov-record -type f`", // how to recover the handle it lost
	}
	scan := text
	for _, a := range allowed {
		if !strings.Contains(scan, a) {
			t.Fatalf("the marker no longer contains %q — if it was reworded, reword this gate with it,\n"+
				"but do not delete the clause: a seat that reads `command not found` as a missing\n"+
				"capability writes a report against a board it never saw.\nmarker:\n%s", a, text)
		}
		scan = strings.ReplaceAll(scan, a, "")
	}

	// Anything left that names the binary and is followed by a word is a command form.
	bare := regexp.MustCompile(`\bfeov-record\b[ \t]+\S`)
	if m := bare.FindString(scan); m != "" {
		t.Fatalf("the marker composes a bare invocation (%q) — the tool is not on PATH, so this is a\n"+
			"command the seat cannot run, offered by the file that exists to unblock it.\n"+
			"Name the shape and point at the dispatch's path instead.\nmarker:\n%s", m, text)
	}

	// And the clause that makes the miss loud must still be there in substance.
	if !strings.Contains(text, "does NOT mean the") {
		t.Fatalf("the marker no longer tells the seat that a failed lookup is not a missing record —\n"+
			"that sentence is the whole fix; without it the marker is silent on the case that\n"+
			"actually happened.\nmarker:\n%s", text)
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
	if _, _, err := RegisterSeat(Identity{RunDir: run, SeatID: "blue-respond-r1", Round: RoundIn(run)("blue-respond-r1")}, ""); err != nil {
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

	if _, _, err := RegisterSeat(Identity{RunDir: run, SeatID: "blue-respond-r1", Round: RoundIn(run)("blue-respond-r1")}, ""); err != nil {
		t.Fatalf("register: %v", err)
	}
	if _, err := Append(Identity{RunDir: run, SeatID: "blue-respond-r1", Round: RoundIn(run)("blue-respond-r1")}, &recordpb.Friction{Text: proto.String("the verb I wanted was not there")}); err != nil {
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
	if _, _, err := RegisterSeat(Identity{RunDir: run, SeatID: "blue-respond-r1", Round: RoundIn(run)("blue-respond-r1")}, ""); err != nil {
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

// A RUN DIRECTORY DELETED AND REBUILT AT THE SAME PATH IS A NEW RUN.
//
// The binding has two halves — the pointer in the user cache and the marker in the run — and
// deleting the directory removes one while leaving the other. Measured: the seat probe rebuilds
// nine boards at fixed paths every invocation, and its second run refused all nine with a
// conflict naming two temp directories the operator never chose.
func TestARebuiltRunDirectoryDoesNotInheritTheOldRoot(t *testing.T) {
	isolate(t)
	run := t.TempDir()
	first := filepath.Join(t.TempDir(), "first")
	t.Setenv(RecordRootEnv, first)
	if _, err := RecordsDir(run); err != nil {
		t.Fatalf("adopt: %v", err)
	}

	// What every probe invocation does: remove the run directory and build a fresh one there.
	if err := os.RemoveAll(run); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(run, 0o755); err != nil {
		t.Fatal(err)
	}

	second := filepath.Join(t.TempDir(), "second")
	t.Setenv(RecordRootEnv, second)
	got, err := RecordsDir(run)
	if err != nil {
		t.Fatalf("a rebuilt run refused its own fresh root: %v", err)
	}
	if !samePath(got, second) {
		t.Fatalf("resolved %s, want the fresh root %s — the new run inherited its predecessor's record", got, second)
	}
}

// The marker is load-bearing in BOTH directions now, so a failure to write it must be an error:
// a silently missing marker would make every later resolve treat a live separated run as a
// deleted one and adopt a fresh empty root in place of its record.
func TestAdoptionFailsIfTheMarkerCannotBeWritten(t *testing.T) {
	isolate(t)
	// A path that exists as a FILE cannot hold the marker inside it.
	f := filepath.Join(t.TempDir(), "not-a-directory")
	if err := os.WriteFile(f, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv(RecordRootEnv, filepath.Join(t.TempDir(), "elsewhere"))
	if _, err := RecordsDir(f); err == nil {
		t.Fatal("adoption succeeded without writing the marker — a later resolve would then treat " +
			"this run as deleted and silently hand it a fresh empty root")
	}
}

// A SEAT RECORDS INTO A RUN THAT EXISTS. IT NEVER CREATES ONE.
//
// Measured (#358). The run directory reaches a seat as a string it resolves against its OWN
// working directory, and this MkdirAll obligingly built a second blackboard wherever that landed:
// a full duplicate tree with the lane's 13.7 KB draft, its own shards, clock and locks, while the
// real run's candidates directory stayed empty for the whole run. The seat was told "registered".
//
// Work landing outside the run is indistinguishable from a seat that produced nothing — the
// plausible zero, built by a helpful mkdir.
func TestRegisterRefusesToCreateARunDirectory(t *testing.T) {
	parent := t.TempDir()
	missing := filepath.Join(parent, "research", "no-such-run")

	_, _, err := RegisterSeat(Identity{RunDir: missing, SeatID: "red-merge-r1", Round: RoundIn(missing)("red-merge-r1")}, "")
	if err == nil {
		t.Fatal("a seat created a run directory from nothing and reported success — the exact failure that produced a second blackboard beside a live run")
	}
	// The message must name the RESOLVED path: a seat that passed a relative --run cannot see
	// where it landed, and that is the whole defect.
	if !strings.Contains(err.Error(), missing) {
		t.Errorf("the refusal does not name the path it resolved to: %v", err)
	}
	if _, statErr := os.Stat(missing); statErr == nil {
		t.Error("the run directory was created despite the refusal")
	}

	// AND A RUN THAT EXISTS STILL WORKS. The check is existence, not a guess about what a run
	// should contain, so a legitimately sparse run directory is never rejected.
	real := filepath.Join(parent, "a-real-run")
	if err := os.MkdirAll(real, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, _, err := RegisterSeat(Identity{RunDir: real, SeatID: "red-merge-r1", Round: RoundIn(real)("red-merge-r1")}, ""); err != nil {
		t.Errorf("an existing run directory was refused: %v", err)
	}
}
