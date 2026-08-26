package record

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
}

// OpenRun resolves a run that MUST ALREADY EXIST, and refuses a path that names no run.
//
// The existence check is the point. RecordsDir answers where a run's events WOULD live, which
// is a different question from whether they do, and every read verb wants the second. Without
// it a typo'd --run is indistinguishable from an empty run, and an empty board is reported for
// a directory nobody ever dispatched.
func OpenRun(dir string) (Run, error) {
	r, err := resolveRun(dir)
	if err != nil {
		return Run{}, err
	}
	if _, err := os.Stat(r.records); err != nil {
		return Run{}, fmt.Errorf("record: %s names no run — its record directory %s is not there (%v). "+
			"A mistyped path resolves like any other, so this is refused rather than reported as an "+
			"empty board, which is the one answer that is never true", dir, r.records, err)
	}
	return r, nil
}

// NewRun resolves a run that is being CREATED, where the records directory need not exist yet.
//
// Separate from OpenRun deliberately: `setup` is the one caller that legitimately holds a run
// before it has a record, and folding the two would mean weakening the check every reader
// depends on to accommodate the single writer that does not.
func NewRun(dir string) (Run, error) { return resolveRun(dir) }

func resolveRun(dir string) (Run, error) {
	if strings.TrimSpace(dir) == "" {
		// The empty string is the state this type exists to stop carrying. It means "nobody
		// supplied a run", and it must not resolve to anything.
		return Run{}, fmt.Errorf("record: no run directory supplied")
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return Run{}, err
	}
	records, err := RecordsDir(abs)
	if err != nil {
		return Run{}, err
	}
	return Run{dir: abs, records: records}, nil
}

// Dir is the run directory, absolute.
func (r Run) Dir() string { return r.dir }

// Records is where this run's events live, resolved once when the handle was made.
func (r Run) Records() string { return r.records }

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
