package proof

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
)

// AN INTERPRETER IS RESOLVED, NEVER ASSUMED — and the name to assume does not exist.
//
// Every extension here used to name ONE binary. `.py` named `python`, which a stock modern
// Linux or macOS box does not have: Debian-derived distributions stopped shipping the
// unversioned name years ago and macOS removed it in 12.3. `python3` is not the fix either —
// on Windows the python.org installer provides `python.exe` while `python3` is typically the
// Microsoft Store alias stub, which opens the Store rather than running the script. There is
// no single correct name, so the tool asks instead of guessing (#522).
//
// THE OTHER THREE HAD THE SAME DEFECT and only `.py` had been noticed. `bash` is absent on
// Alpine and on a Windows box without git-bash; `node` and `go` are ordinary PATH lookups that
// can miss. Fixing the instance would have left three siblings — see [[refactoring-safety]],
// fix the class.
//
// WHY THIS CANNOT BORROW THE SUITE'S OWN VOCABULARY. `toolchain.Present` / `NotApplicableIn`
// express exactly this, and live in `plugins/prosthetic-conscience/tools`, a SEPARATE Go
// module. Importing across that boundary is a bigger decision than this fix; the shape is
// duplicated here deliberately and narrowly, and this comment is the record of why.
//
// AND PROOFS ARE NOT DECLARED not-applicable IN A CLOUD SESSION, which is the tempting move
// and the wrong one: `blue prove` anchors a claim to a computation, so excusing proofs where
// the interpreter is missing would silently thin every such run's evidence base while every
// gate stayed green. That is the defect one level up. They stay required, and an absent
// interpreter is LOUD.
var candidates = map[string][]string{
	".py":  {"python3", "python"},
	".js":  {"node"},
	".mjs": {"node"},
	".sh":  {"bash", "sh"},
	".go":  {"go"},
}

// argvFor is the extra argv an interpreter needs after its own name.
var argvFor = map[string][]string{"go": {"run"}}

// windowsExtra are candidates that only make sense on Windows — `py -3` is the launcher the
// python.org installer ships, and it is the reliable way to reach 3.x there.
var windowsExtra = map[string][]string{".py": {"py"}}

// resolver caches PATH lookups for the life of the process. A run executes every proof script
// twice by design, and a debate can carry dozens, so re-probing PATH per invocation would turn
// one answer into hundreds of identical syscalls.
type resolver struct {
	mu   sync.Mutex
	seen map[string]string // candidate name -> resolved path, "" for a confirmed miss
	look func(string) (string, error)
	// calls is test scaffolding: the fake `look` counts its own probes into it, so a test can
	// assert the cache actually PREVENTS a second lookup rather than assert it exists. Nothing
	// in this file writes it, and it is nil in production.
	calls map[string]int
}

var defaultResolver = &resolver{seen: map[string]string{}, look: exec.LookPath}

func (r *resolver) resolve(name string) string {
	r.mu.Lock()
	defer r.mu.Unlock()
	if p, ok := r.seen[name]; ok {
		return p
	}
	p, err := r.look(name)
	if err != nil {
		p = "" // a confirmed miss is CACHED too: the answer is stable and the cost is not
	}
	r.seen[name] = p
	return p
}

// available reports every interpreter name this environment can actually run, across all
// extensions. It is what turns an unusable environment into a routable one: the seat chooses
// the extension, and until it is told which ones work here it can only guess.
func (r *resolver) available() []string {
	seen := map[string]bool{}
	var out []string
	for ext := range candidates {
		for _, name := range candidatesFor(ext) {
			if r.resolve(name) != "" && !seen[name] {
				seen[name] = true
				out = append(out, name)
			}
		}
	}
	sort.Strings(out)
	return out
}

func candidatesFor(ext string) []string {
	c := append([]string{}, candidates[ext]...)
	if runtime.GOOS == "windows" {
		c = append(c, windowsExtra[ext]...)
	}
	return c
}

// interpreterFor maps a script to how it runs.
//
// BY EXTENSION, deliberately: a shebang does not work on Windows and this suite runs there.
// An unknown extension is REFUSED rather than guessed, because running a file under the wrong
// interpreter produces a confident error that reads exactly like a failed proof — and the same
// sentence is why an UNRESOLVABLE one is refused here rather than attempted.
func interpreterFor(path string) ([]string, error) {
	return defaultResolver.interpreterFor(path)
}

func (r *resolver) interpreterFor(path string) ([]string, error) {
	ext := strings.ToLower(filepath.Ext(path))
	tried := candidatesFor(ext)
	if len(tried) == 0 {
		return nil, fmt.Errorf("proof: %s has no known interpreter — a proof script must be .py, .js, .mjs, .sh or .go so the tool runs it the way you meant rather than guessing", filepath.Base(path))
	}
	for _, name := range tried {
		if p := r.resolve(name); p != "" {
			argv := append([]string{p}, argvFor[name]...)
			// `py` takes the version selector before the script, and without it the launcher
			// may pick a 2.x it finds first.
			if name == "py" {
				argv = append(argv, "-3")
			}
			return argv, nil
		}
	}
	// THE MISS IS LOUD, AND IT SAYS WHAT WOULD WORK. A proof that cannot RUN must never
	// produce something a reader takes for a proof that FAILED, and a seat told only "no" will
	// rewrite the same script in the same language. The environment is often not the seat's to
	// change — it is a fresh container — but the EXTENSION is, so the way out is named here.
	msg := fmt.Sprintf("proof: cannot run %s — no interpreter for %s in this environment (tried %s)",
		filepath.Base(path), ext, strings.Join(tried, ", "))
	if avail := r.available(); len(avail) > 0 {
		msg += fmt.Sprintf(". This environment DOES have %s — write the proof in a language one of those runs, and it will execute here",
			strings.Join(avail, ", "))
	} else {
		msg += ". This environment has NONE of the supported interpreters, so no proof can run in it at all — record that with the friction verb rather than dropping the proof silently"
	}
	return nil, fmt.Errorf("%s", msg)
}
