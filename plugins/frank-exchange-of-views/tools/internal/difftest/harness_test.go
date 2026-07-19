// Package difftest carries the golden harness for feov-record.
//
// It was born as the R2g DIFFERENTIAL GATE: identical command sequences driven
// through the frozen mjs oracle and the Go port, compared on stdout, stderr,
// exit code, event structure, and byte-identical renders. The gate did its job —
// it caught three real port bugs (JS `${undefined}` interpolation, UTF-16 vs
// byte slicing, and HTML escaping in encoding/json) — and then the oracle was
// retired: it existed only to validate the port, and it had never been used in
// a run.
//
// What remains is the harness that drove it, now serving the GOLDEN suites:
// scenario replay (golden_test.go), the help and error contracts
// (contract_test.go), and replay determinism (fuzz_test.go). Every golden in
// testdata/ was recorded while the gate was green, so the oracle's validation is
// preserved in the files even though the oracle is gone.
//
// The normalization rules below outlived the gate for the same reason they
// existed: nonces are random per register, and mtimes drive winner selection, so
// a run is only comparable to another run after both are canonicalized.
package difftest

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"
)

// cmd is one CLI invocation. Role selects the seat contract; the Go side takes it
// as a subcommand, the mjs side as a distinct script file.
type cmd struct {
	role string
	args []string
	// mtimes, when set, are applied to matching shard files after the command —
	// policy 4, for the winner-selection fallback.
	mtimes map[string]time.Time
}

type scenario struct {
	name string
	cmds []cmd
	// files written into the run dir before the scenario (registry, prose payloads).
	seed map[string]string
}

func repoRoot(t *testing.T) string {
	t.Helper()
	// .../plugins/frank-exchange-of-views/tools/internal/difftest -> up 5
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Clean(filepath.Join(wd, "..", "..", "..", "..", ".."))
}

func buildBinary(t *testing.T) string {
	t.Helper()
	out := filepath.Join(t.TempDir(), "feov-record")
	if runtime.GOOS == "windows" {
		out += ".exe"
	}
	root := repoRoot(t)
	c := exec.Command("go", "build", "-o", out, "./cmd/feov-record")
	c.Dir = filepath.Join(root, "plugins", "frank-exchange-of-views", "tools")
	if b, err := c.CombinedOutput(); err != nil {
		t.Fatalf("building feov-record: %v\n%s", err, b)
	}
	return out
}

type invocation struct {
	stdout, stderr string
	code           int
}

func runGo(bin, runDir string, c cmd) invocation {
	args := append([]string{c.role}, substitute(c.args, runDir)...)
	return capture(exec.Command(bin, args...))
}

func substitute(args []string, runDir string) []string {
	out := make([]string, len(args))
	for i, a := range args {
		out[i] = strings.ReplaceAll(a, "{RUN}", runDir)
	}
	return out
}

func capture(c *exec.Cmd) invocation {
	var so, se strings.Builder
	c.Stdout, c.Stderr = &so, &se
	// HERMETIC. --run is inferred from .claude/run-live.json when the flag is absent,
	// walking up from CLAUDE_PROJECT_DIR or cwd. The harness runs inside this repo, so
	// a live research run on the developer's machine would satisfy the flag and change
	// what the binary prints — the `missing_required_flags` case passed on CI and
	// failed locally, which is the same machine-dependence that once put a developer's
	// home directory into the goldens. Point the binary at a directory with no marker
	// so the recorded behaviour is a property of the tool, not of the machine.
	c.Env = append(os.Environ(), "CLAUDE_PROJECT_DIR="+os.TempDir())
	err := c.Run()
	code := 0
	if ee, ok := err.(*exec.ExitError); ok {
		code = ee.ExitCode()
	} else if err != nil {
		code = -1
	}
	return invocation{stdout: so.String(), stderr: se.String(), code: code}
}

var shardRe = regexp.MustCompile(`^events-(.+)-([0-9a-f]{8})\.jsonl$`)

// The verdict verb checkpoints into the USER'S HOME cache, so the path carries
// both the home directory and a hash of the run dir. Normalizing only the hash
// left the developer's home path baked into two goldens — files that could never
// pass on another machine, and did not: CI on Linux was the first environment to
// run them. A golden is a contract about BEHAVIOUR, so anything naming the
// machine has to go, and the whole path is replaced rather than its tail.
// placeholderPathRe matches a path rooted at one of our placeholders, which is
// the only place a separator is ours to normalize.
var placeholderPathRe = regexp.MustCompile(`\{(?:RUN|MIRROR)\}[^\s"']*`)

var mirrorRe = regexp.MustCompile(`\S*[\\/]run-mirror[\\/][0-9a-f]{12}`)

// nonceMapper assigns canonical placeholders in shard DISCOVERY order.
type nonceMapper struct {
	seen map[string]string
	n    int
}

func newMapper() *nonceMapper { return &nonceMapper{seen: map[string]string{}} }

// observe scans for shards that appeared since the last call and numbers them.
func (m *nonceMapper) observe(recordsDir string) {
	entries, err := os.ReadDir(recordsDir)
	if err != nil {
		return
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	sort.Strings(names)
	for _, n := range names {
		g := shardRe.FindStringSubmatch(n)
		if g == nil {
			continue
		}
		if _, ok := m.seen[g[2]]; !ok {
			m.n++
			m.seen[g[2]] = fmt.Sprintf("NONCE%03d", m.n)
		}
	}
}

func (m *nonceMapper) normalize(s string) string {
	for raw, placeholder := range m.seen {
		s = strings.ReplaceAll(s, raw, placeholder)
	}
	return sortNonceLists(s)
}

// nonceListRe matches the anomaly footer's dispatch list: "(NONCE001, NONCE002)".
var nonceListRe = regexp.MustCompile(`\((NONCE\d{3}(?:, NONCE\d{3})+)\)`)

// sortNonceLists canonicalizes the ORDER of that list, which is incomparable
// across the two sides by construction.
//
// Both implementations list a seat's shards in filename order, and the filename
// carries a RANDOM nonce — so the same run yields "(a1b2, 0c9f)" on one side and
// "(3d4e, 7f10)" on the other, sorting differently even though placeholders are
// assigned by discovery. The fuzzer surfaced this immediately. The alternative
// (assigning placeholders by sorted hex instead of discovery order) would make
// this line agree while making the WINNER line incomparable, since the winner
// tracks creation/mtime rather than hex — so the list order is normalized and
// the winner, which is the semantic claim, stays compared verbatim.
func sortNonceLists(s string) string {
	return nonceListRe.ReplaceAllStringFunc(s, func(match string) string {
		inner := strings.TrimSuffix(strings.TrimPrefix(match, "("), ")")
		parts := strings.Split(inner, ", ")
		sort.Strings(parts)
		return "(" + strings.Join(parts, ", ") + ")"
	})
}

// collect reads the comparable state of a run dir: shard events (parsed) keyed by
// normalized filename, plus every projection file's normalized bytes.
type state struct {
	events  map[string][]map[string]any
	renders map[string]string
	pointer map[string]string
}

func collect(t *testing.T, runDir string, m *nonceMapper) state {
	t.Helper()
	recs := filepath.Join(runDir, "records")
	st := state{events: map[string][]map[string]any{}, renders: map[string]string{}, pointer: map[string]string{}}
	entries, err := os.ReadDir(recs)
	if err != nil {
		return st
	}
	for _, e := range entries {
		name := e.Name()
		switch {
		case strings.HasPrefix(name, ".lock-"), strings.Contains(name, ".tmp-"):
			continue // policy 5
		case strings.HasPrefix(name, ".active-"):
			b, _ := os.ReadFile(filepath.Join(recs, name))
			st.pointer[name] = m.normalize(string(b))
		case shardRe.MatchString(name):
			b, _ := os.ReadFile(filepath.Join(recs, name))
			var evs []map[string]any
			for _, line := range strings.Split(m.normalize(string(b)), "\n") {
				if strings.TrimSpace(line) == "" {
					continue
				}
				var ev map[string]any
				if err := json.Unmarshal([]byte(line), &ev); err != nil {
					// A deliberately torn line is part of the fixture, not a failure:
					// record it verbatim so both sides must tear identically.
					evs = append(evs, map[string]any{"__unparseable": line})
					continue
				}
				evs = append(evs, ev)
			}
			st.events[m.normalize(name)] = evs
		}
	}
	shadow := filepath.Join(recs, "render-shadow")
	if fs, err := os.ReadDir(shadow); err == nil {
		for _, f := range fs {
			if strings.Contains(f.Name(), ".tmp-") {
				continue
			}
			b, _ := os.ReadFile(filepath.Join(shadow, f.Name()))
			st.renders[f.Name()] = m.normalize(string(b))
		}
	}
	return st
}

// normalizeOutput strips the run-dir path (which differs by construction) and
// canonicalizes nonces so CLI text is comparable.
func normalizeOutput(inv invocation, runDir string, m *nonceMapper) invocation {
	clean := func(s string) string {
		s = strings.ReplaceAll(s, runDir, "{RUN}")
		s = strings.ReplaceAll(s, filepath.ToSlash(runDir), "{RUN}")
		s = m.normalize(s)
		// The verdict verb's recovery mirror is keyed by sha1 of the run dir, and
		// the two sides run in different temp dirs by construction — the hash
		// difference is harness noise, not a port divergence.
		s = mirrorRe.ReplaceAllString(s, "{MIRROR}")
		// Separators inside OUR OWN paths are OS-specific: Windows renders
		// "{RUN}\records\render-shadow" where Linux renders "{RUN}/records/...".
		// Normalized only within placeholder-rooted tokens, never globally — seat
		// prose legitimately contains backslashes (the hostile-prose fixture carries
		// "backslash \ and \n literal"), and flattening those would corrupt the very
		// payloads that golden is there to protect.
		s = placeholderPathRe.ReplaceAllStringFunc(s, func(tok string) string {
			return strings.ReplaceAll(tok, "\\", "/")
		})
		// The tool name differs by construction: "feov-record merge" vs "red-merge.mjs".
		for _, prefix := range []string{"feov-record lens", "feov-record merge", "feov-record blue", "feov-record bench",
			"red-lens.mjs", "red-merge.mjs", "blue.mjs", "bench.mjs"} {
			s = strings.ReplaceAll(s, prefix, "{TOOL}")
		}
		return strings.ReplaceAll(s, "\r\n", "\n")
	}
	return invocation{stdout: clean(inv.stdout), stderr: clean(inv.stderr), code: inv.code}
}

func applyMtimes(t *testing.T, runDir string, m *nonceMapper, c cmd) {
	t.Helper()
	if len(c.mtimes) == 0 {
		return
	}
	recs := filepath.Join(runDir, "records")
	entries, err := os.ReadDir(recs)
	if err != nil {
		return
	}
	for _, e := range entries {
		norm := m.normalize(e.Name())
		if when, ok := c.mtimes[norm]; ok {
			p := filepath.Join(recs, e.Name())
			if err := os.Chtimes(p, when, when); err != nil {
				t.Fatalf("chtimes %s: %v", p, err)
			}
		}
	}
}

func seed(t *testing.T, runDir string, files map[string]string) {
	t.Helper()
	for rel, content := range files {
		p := filepath.Join(runDir, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// TestGoldensAreMachineIndependent is the guard for the failure that shipped:
// two goldens carried the developer's home directory, so they could only ever
// pass on one machine — and nothing noticed until CI ran on Linux. A golden is a
// contract about behaviour; a path naming a user, a home directory, or a temp
// root is a contract about a laptop.
func TestGoldensAreMachineIndependent(t *testing.T) {
	entries, err := os.ReadDir("testdata")
	if err != nil {
		t.Skip("no goldens yet")
	}
	// Patterns that can only be true on the machine that recorded them.
	machine := []string{`[A-Za-z]:\\Users\\`, `/home/[a-z]`, `/Users/[a-z]`, `AppData`, `/tmp/`, `\\Temp\\`}
	res := make([]*regexp.Regexp, len(machine))
	for i, p := range machine {
		res[i] = regexp.MustCompile(p)
	}
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".golden") {
			continue
		}
		b, err := os.ReadFile(filepath.Join("testdata", e.Name()))
		if err != nil {
			t.Fatal(err)
		}
		for i, re := range res {
			if loc := re.FindIndex(b); loc != nil {
				line := 1 + strings.Count(string(b[:loc[0]]), "\n")
				t.Errorf("%s:%d carries a machine-specific path (%s): %q — normalize it in the harness, do not re-record",
					e.Name(), line, machine[i], strings.TrimSpace(string(b[max(0, loc[0]-40):loc[1]+20])))
			}
		}
	}
}
