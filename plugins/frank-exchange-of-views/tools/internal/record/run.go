package record

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
)

// Run is a run directory that RESOLVED — the handle 28 functions currently take as a string.
//
// # What a string admits that this does not
//
// Three incidents, one shape. A `runDir string` can hold a path nobody dispatched, a path that
// resolves to the same record as another spelling of itself, and the empty string — which
// already means "nobody supplied one" and gets pressed into service as "you were refused" as
// well. seat.Context says it plainly about that last one: two states, one byte, and the
// healthy one wins by default.
//
//   - #526: `show board --run "<typo>"` printed a board for a directory nobody had dispatched.
//     A typo resolves like any other path — RecordsDir's ordinary arm returns
//     <runDir>/records whether or not anything is there — so the read succeeded and reported
//     an empty board, which is the one answer that is never true.
//   - #358: two unequal strings can name one record, so the handle cache had to be keyed on
//     the RESOLVED directory rather than the argument.
//   - seat.Context carries RunErr beside RunDir because a refusal had nowhere else to live.
//
// A Run cannot be any of those. It is constructible only by resolution, its zero value is not
// a run, and it is not a string — so it cannot be handed to filepath.Join or os.ReadFile
// without saying Dir() out loud, which is the moment a reader notices what is being done.
//
// # Why the resolved records directory is carried
//
// Resolution is not free and it is not pure: it reads a pointer file, consults an environment
// variable, and can adopt a root. Doing it once per handle rather than once per call is what
// makes "two spellings, one record" a property of the type instead of a rule every caller has
// to remember.
type Run struct {
	dir     string // absolute run directory
	records string // resolved records directory, from RecordsDir
	// separated is true when the record lives OUTSIDE the run directory. It decides what a
	// missing record directory MEANS at read time, and the two answers are opposites: a local
	// run that has not been written to yet is an honest empty board, while a separated run whose
	// root has gone is the lie this package exists to refuse.
	separated bool
}

// OpenRun resolves a run that MUST ALREADY EXIST, and refuses a path that names no run.
//
// The existence check is the point: RecordsDir answers where a run's events WOULD live, which
// is a different question from whether the run is real. Without it a typo'd --run resolves
// like any other path and the read reports an empty board for a directory nobody dispatched.
//
// IT CHECKS THE RUN DIRECTORY, NOT THE RECORD DIRECTORY, and the difference is not cosmetic.
// A first write CREATES the record directory (store.go's MkdirAll), so a run that exists and
// has simply not been written to yet has no records/ — requiring one would refuse every
// opening act of every run. Measured the moment this was written the other way: five tests
// across fetch, friction and the operator read failed on runs that were entirely real.
//
// The run directory is what setup makes and what a typo gets wrong, which is exactly the
// distinction wanted.
func OpenRun(dir string) (Run, error) {
	r, err := resolveRun(dir)
	if err != nil {
		return Run{}, err
	}
	if _, err := os.Stat(r.dir); err != nil {
		// CODED, because the --json edge reads CodeOf and an uncoded error flattens to the
		// generic "error". These refusals sit where feov.MissingField and the seat's coded
		// Conflict used to be returned, and a caller switching on the code would have seen that
		// distinction quietly disappear.
		return Run{}, feov.Wrap(feov.NotFound, err,
			"record: %s names no run — the directory is not there (%v). "+
				"A mistyped --run resolves like any other path, so this is refused rather than "+
				"reported as an empty board, which is the one answer that is never true", dir, err)
	}
	return r, nil
}

// NewRun resolves a run that is being CREATED, where the directory need not exist yet.
//
// Separate from OpenRun deliberately: `setup` is the one caller that holds a run before there
// is anything on disk to open, and folding the two would mean dropping the existence check
// every reader depends on to accommodate the single writer that cannot meet it.
func NewRun(dir string) (Run, error) { return resolveRun(dir) }

func resolveRun(dir string) (Run, error) {
	if strings.TrimSpace(dir) == "" {
		// The empty string is the state this type exists to stop carrying. It means "nobody
		// supplied a run", and it must not resolve to anything.
		return Run{}, feov.Errorf(feov.MissingField,
			"record: no run directory supplied — pass --run <runDir>, or run inside a dispatch that injects it")
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return Run{}, err
	}
	records, err := RecordsDir(abs)
	if err != nil {
		return Run{}, err
	}
	return Run{dir: abs, records: records, separated: !strings.HasPrefix(records, abs+string(filepath.Separator))}, nil
}

// Dir is the run directory, absolute.
func (r Run) Dir() string { return r.dir }

// Records is where this run's events live, resolved once when the handle was made.
func (r Run) Records() string { return r.records }

// Separated reports whether the record lives outside the run directory.
//
// Read-time meaning, not trivia: RESOLUTION now happens once, when the handle is made, so a
// handle that outlives a change to the record root would otherwise read the old path and find
// nothing — and "nothing" is the honest answer for a fresh local run and a lie for a separated
// one whose root was deleted. `dashboard --watch` holds a handle across regenerations, so the
// window is real rather than theoretical.
func (r Run) Separated() bool { return r.separated }

// Valid reports whether this is a resolved run rather than the zero value.
//
// A caller should rarely need it: a Run obtained from OpenRun or NewRun without an error IS
// valid, and the zero value only reaches somewhere that skipped the constructor. It exists so
// that code holding a Run in a struct can assert the field was filled.
func (r Run) Valid() bool { return r.dir != "" && r.records != "" }

// String makes a Run printable without exposing it as a path. Deliberately not the directory
// alone: a Run that reaches a %s should read as a run, not be mistaken for something a caller
// could have joined onto.
func (r Run) String() string {
	if !r.Valid() {
		return "<unresolved run>"
	}
	return r.dir
}
