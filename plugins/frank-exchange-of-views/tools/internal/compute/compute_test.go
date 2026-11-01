package compute

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatenv"
)

// fxPath is the fixture program every test here runs (testdata/fx), built once for the package: a
// binary, so no arm depends on an interpreter being installed, and every arm runs on Windows too.
var fxPath string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "compute-fx-*")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fxPath = filepath.Join(dir, "fx")
	if runtime.GOOS == "windows" {
		fxPath += ".exe"
	}
	build := exec.Command("go", "build", "-o", fxPath, "./testdata/fx")
	if out, err := build.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "building the fixture program: %v\n%s", err, out)
		os.Exit(1)
	}
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

// bundle writes files into a fresh bundle directory and returns it.
func bundle(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, body := range files {
		p := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func noRefusal(string) error { return nil }

// program reads a bundle for the fixture run with args.
func program(t *testing.T, dir string, args ...string) *Program {
	t.Helper()
	p, err := ReadProgram(dir, append([]string{fxPath}, args...), noRefusal)
	if err != nil {
		t.Fatalf("ReadProgram: %v", err)
	}
	return p
}

// run executes a program twice, failing the test on a tool error.
func run(t *testing.T, p *Program) *Result {
	t.Helper()
	res, err := Run(t.TempDir(), p)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	return res
}

// isolateTemp points the system temporary directory at a fresh one the test can inspect.
func isolateTemp(t *testing.T, dir string) {
	t.Helper()
	for _, k := range []string{"TMPDIR", "TMP", "TEMP"} {
		t.Setenv(k, dir)
	}
}

// THE POINT OF THE VERB: a computation settles the question and leaves an artifact, and two runs that
// agree are `reproducible`.
func TestAgreeingRunsAreReproducible(t *testing.T) {
	p := program(t, bundle(t, map[string]string{"data.txt": "9\n"}), "print", "9 is composite")
	res := run(t, p)
	if res.Drift != "" {
		t.Errorf("two runs of a deterministic program differed: %s", res.Drift)
	}
	if res.Output != "9 is composite\n" || res.Exit != 0 {
		t.Errorf("recorded %q exit %d", res.Output, res.Exit)
	}
	if res.Bundle != BundleSha(p.Argv, Sums(p.Files)) {
		t.Errorf("bundle %s is not the identity of the program's argv and files", res.Bundle)
	}
	// A program named by its absolute path is a toolchain: its path and bytes are named, never part
	// of the bundle's identity.
	want, err := fileSha(fxPath)
	if err != nil {
		t.Fatal(err)
	}
	if res.ToolchainPath != fxPath || res.ToolchainSha256 != want {
		t.Errorf("toolchain = %s %s, want %s %s", res.ToolchainPath, res.ToolchainSha256, fxPath, want)
	}
}

// UNDECLARED DRIFT IS REPORTED, each kind by name: the caller refuses it until the seat states what
// holds across the runs.
func TestEachKindOfDriftIsNamed(t *testing.T) {
	counter := filepath.Join(t.TempDir(), "count")
	for name, c := range map[string]struct {
		args []string
		says string
	}{
		"output": {[]string{"random"}, "output"},
		"exit":   {[]string{"exitcount", counter}, "exit code moved 0 -> 1"},
		"out/":   {[]string{"outrandom"}, "out/r.txt differs"},
	} {
		t.Run(name, func(t *testing.T) {
			res := run(t, program(t, bundle(t, map[string]string{"a": "a"}), c.args...))
			if !strings.Contains(res.Drift, c.says) {
				t.Errorf("drift %q does not name %q", res.Drift, c.says)
			}
		})
	}
}

// out/ IS PART OF THE RESULT: the first run's files come back, and two runs that wrote the same
// out/ agree.
func TestOutIsCaptured(t *testing.T) {
	res := run(t, program(t, bundle(t, map[string]string{"a": "a"}), "out", "t.csv", "n,7\n"))
	if res.Drift != "" {
		t.Fatalf("drift %q", res.Drift)
	}
	if len(res.Out) != 1 || res.Out[0].Path != "t.csv" || string(res.Out[0].Body) != "n,7\n" {
		t.Errorf("out = %+v", res.Out)
	}
}

// A PROGRAM THAT REACHES OUT OF ITS COPY BY A RELATIVE PATH reaches a different place from each copy,
// because the copies sit at different depths: a file present beside one only moves the output.
func TestReachingOutOfTheCopyMoves(t *testing.T) {
	tmp := t.TempDir()
	isolateTemp(t, tmp)
	// The first copy is <tmp>/feov-compute-*/a/bundle, so ../../../x is <tmp>/x; from the second,
	// one level deeper, it is inside the scratch directory, where nothing is.
	if err := os.WriteFile(filepath.Join(tmp, "x"), []byte("present\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	p := program(t, bundle(t, map[string]string{"a": "a"}), "read", "../../../x")
	res, err := Run(t.TempDir(), p)
	if err != nil {
		t.Fatal(err)
	}
	if res.Drift == "" {
		t.Errorf("a program reading a file present at one depth only agreed with itself: %q", res.Output)
	}
}

// A MISSING PATH SAID IN LOWERCASE is no signature the tool knows, and two runs that agree on it are
// recorded with the exit they share — the residue the computation lens judges.
func TestAMissingPathSaidInLowercaseIsRecorded(t *testing.T) {
	res := run(t, program(t, bundle(t, map[string]string{"a": "a"}), "read", "../nothing-here"))
	if res.Drift != "" || res.Exit != 1 || res.Failed != "" {
		t.Errorf("drift %q exit %d failed %q, want an agreeing exit 1 with no signature: %q", res.Drift, res.Exit, res.Failed, res.Output)
	}
	if !strings.Contains(res.Output, "no such file or directory") && !strings.Contains(res.Output, "cannot find") {
		t.Errorf("output %q", res.Output)
	}
}

// "No such file or directory" on stderr at exit 0 is the environment failing, not the program
// answering: Failed names it, and the caller refuses it without --expect-error.
func TestAnEnvironmentErrorOnStderrAtExitZeroIsNamed(t *testing.T) {
	res := run(t, program(t, bundle(t, map[string]string{"a": "a"}), "stderr", "find: ./plugins: No such file or directory"))
	if res.Exit != 0 || res.Failed == "" {
		t.Errorf("exit %d failed %q: output %q carried an environment error", res.Exit, res.Failed, res.Output)
	}
}

// A PROGRAM THAT CALLS THE TOOL IS NAMED, whatever it printed: the marker is the signal.
func TestAProgramThatCallsTheToolIsNamed(t *testing.T) {
	res := run(t, program(t, bundle(t, map[string]string{"a": "a"}), "call"))
	if len(res.Calls) != 1 || res.Calls[0] != "feov-record show board --json" {
		t.Errorf("calls = %q, want the one call", res.Calls)
	}
	if os.Getenv(seatenv.VarComputeRerun) != "" {
		t.Errorf("%s leaked into the caller's environment", seatenv.VarComputeRerun)
	}
}

// EVERY PIN IS IN THE PROGRAM'S ENVIRONMENT, and the copy's own directory is its PWD.
func TestTheProgramRunsUnderThePins(t *testing.T) {
	res := run(t, program(t, bundle(t, map[string]string{"a": "a"}), "env"))
	for _, want := range []string{
		"SOURCE_DATE_EPOCH=0\n", "PYTHONHASHSEED=0\n", "TZ=UTC\n", "LC_ALL=C\n", "PYTHONUTF8=1\n",
		"PWD=" + BundleToken + "\n", `MATPLOTLIBRC holds "svg.hashsalt: feov\n"`, "checker marker set: true",
	} {
		if !strings.Contains(res.Output, want) {
			t.Errorf("the program's environment lacks %q:\n%s", want, res.Output)
		}
	}
}

// THE BUNDLE'S REFUSALS: each is something a copy could not honour, and each is named.
func TestTheBundleRefusals(t *testing.T) {
	big := strings.Repeat("x", MaxBundle+1)
	cases := map[string]struct {
		files map[string]string
		argv  []string
		says  string
	}{
		"over the limit":            {map[string]string{"big.txt": big}, nil, "over"},
		"an out entry":              {map[string]string{"out": "x"}, nil, "out"},
		"a sources entry":           {map[string]string{"sources/C-1": "x"}, nil, "sources"},
		"an interpreter cache":      {map[string]string{"__pycache__/m.cpython-312.pyc": "x"}, nil, "__pycache__"},
		"a compiled file":           {map[string]string{"m.pyc": "x"}, nil, "m.pyc"},
		"no command":                {map[string]string{"a": "a"}, []string{}, "command after `--`"},
		"a relative command absent": {map[string]string{"a": "a"}, []string{"./run.sh"}, "the bundle holds no run.sh"},
		"a command outside":         {map[string]string{"a": "a"}, []string{"../x/run.sh"}, "relative"},
		"no file":                   {map[string]string{}, nil, "holds no file"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			argv := c.argv
			if argv == nil {
				argv = []string{fxPath, "print", "x"}
			}
			_, err := ReadProgram(bundle(t, c.files), argv, noRefusal)
			if err == nil || !strings.Contains(err.Error(), c.says) {
				t.Errorf("err = %v, want one saying %q", err, c.says)
			}
		})
	}
	t.Run("a symlink", func(t *testing.T) {
		dir := bundle(t, map[string]string{"a": "a"})
		if err := os.Symlink(filepath.Join(dir, "a"), filepath.Join(dir, "b")); err != nil {
			t.Skipf("no symlinks here: %v", err)
		}
		if _, err := ReadProgram(dir, []string{fxPath, "print"}, noRefusal); err == nil || !strings.Contains(err.Error(), "symlink") {
			t.Errorf("err = %v", err)
		}
	})
	t.Run("the caller's refusal comes before any read", func(t *testing.T) {
		dir := bundle(t, map[string]string{"a": "a", "record.db": "r"})
		var seen []string
		_, err := ReadProgram(dir, []string{fxPath}, func(p string) error {
			seen = append(seen, filepath.Base(p))
			if filepath.Base(p) == "record.db" {
				return fmt.Errorf("this run's record")
			}
			return nil
		})
		if err == nil || !strings.Contains(err.Error(), "this run's record") {
			t.Errorf("err = %v, seen %q", err, seen)
		}
	})
}

// A BARE COMMAND NOT ON PATH is named, so the seat knows which program this environment lacks.
func TestABareCommandNotOnPathIsNamed(t *testing.T) {
	p, err := ReadProgram(bundle(t, map[string]string{"a": "a"}), []string{"feov-no-such-program-xyz", "a"}, noRefusal)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Run(t.TempDir(), p); err == nil || !strings.Contains(err.Error(), "feov-no-such-program-xyz is not on PATH") {
		t.Errorf("err = %v", err)
	}
}

// THE IDENTITY IS FRAMED: every field is length-prefixed, so an argument holding a newline or a space
// and one splitting it differently are different bundles; and file bytes are in it.
func TestTheBundleIdentityIsFramed(t *testing.T) {
	dir := bundle(t, map[string]string{"a": "a"})
	sha := func(argv ...string) string {
		p, err := ReadProgram(dir, argv, noRefusal)
		if err != nil {
			t.Fatal(err)
		}
		return p.Sha
	}
	if sha("python3", "a\nb") == sha("python3", "a", "b") {
		t.Error("an argument with a newline and two arguments hash alike")
	}
	if sha("python3", "a b") == sha("python3", "a", "b") {
		t.Error("an argument with a space and two arguments hash alike")
	}
	before := sha("python3", "a")
	if err := os.WriteFile(filepath.Join(dir, "a"), []byte("b"), 0o644); err != nil {
		t.Fatal(err)
	}
	if sha("python3", "a") == before {
		t.Error("two bundles with different file bytes hash alike")
	}
}

// NO COPY REMAINS after an act — not even one the program left read-only, which neither Windows nor
// a directory without its write bit lets RemoveAll take.
func TestNoCopyRemains(t *testing.T) {
	tmp := t.TempDir()
	isolateTemp(t, tmp)
	p := program(t, bundle(t, map[string]string{"a": "a"}), "readonly")
	if _, err := Run(t.TempDir(), p); err != nil {
		t.Fatal(err)
	}
	left, err := os.ReadDir(tmp)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range left {
		t.Errorf("the act left %s in the temporary directory", e.Name())
	}
}

// A COPY'S ROOT READS AS <bundle> in what the program writes — spelled as the program sees it, which
// is the canonical path: under a symlinked temporary directory getcwd names the real one.
func TestTheCopyRootReadsAsTheBundle(t *testing.T) {
	real := t.TempDir()
	tmp := real
	if runtime.GOOS != "windows" {
		tmp = filepath.Join(t.TempDir(), "link")
		if err := os.Symlink(real, tmp); err != nil {
			t.Fatal(err)
		}
	}
	isolateTemp(t, tmp)
	res := run(t, program(t, bundle(t, map[string]string{"data.txt": "x"}), "where"))
	if res.Drift != "" {
		t.Fatalf("a program printing its own location moved between copies: %s\n%s", res.Drift, res.Output)
	}
	if !strings.Contains(res.Output, "cwd: "+BundleToken+"\n") || !strings.Contains(res.Output, "file: "+filepath.Join(BundleToken, "data.txt")) {
		t.Errorf("the output does not name the copy as %s:\n%s", BundleToken, res.Output)
	}
	if strings.Contains(res.Output, real) || strings.Contains(res.Output, tmp) {
		t.Errorf("the output holds a copy's path:\n%s", res.Output)
	}
	if res.Exit != 1 {
		t.Errorf("exit %d, want the program's 1", res.Exit)
	}
}

// The same, through an interpreter: a Python program printing __file__ and os.getcwd() and raising.
func TestAPythonProgramsOwnPathReadsAsTheBundle(t *testing.T) {
	py, err := exec.LookPath("python3")
	if err != nil {
		t.Skip("python3 is not on PATH; the fixture arm above covers the mechanism")
	}
	real := t.TempDir()
	tmp := real
	if runtime.GOOS != "windows" {
		tmp = filepath.Join(t.TempDir(), "link")
		if err := os.Symlink(real, tmp); err != nil {
			t.Fatal(err)
		}
	}
	isolateTemp(t, tmp)
	dir := bundle(t, map[string]string{"where.py": "import os\nprint(os.path.abspath(__file__))\nprint(os.getcwd())\nraise SystemExit('the claim does not hold')\n"})
	p, err := ReadProgram(dir, []string{filepath.Base(py), "where.py"}, noRefusal)
	if err != nil {
		t.Fatal(err)
	}
	res := run(t, p)
	if res.Drift != "" || !strings.Contains(res.Output, filepath.Join(BundleToken, "where.py")) || strings.Contains(res.Output, real) {
		t.Errorf("drift %q, output:\n%s", res.Drift, res.Output)
	}
}
