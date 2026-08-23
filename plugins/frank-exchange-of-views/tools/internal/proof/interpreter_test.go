package proof

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// fakeLook builds a resolver whose PATH contains exactly the named binaries.
func fakeLook(present ...string) *resolver {
	set := map[string]bool{}
	for _, p := range present {
		set[p] = true
	}
	calls := map[string]int{}
	r := &resolver{seen: map[string]string{}}
	r.look = func(name string) (string, error) {
		calls[name]++
		if set[name] {
			return "/usr/bin/" + name, nil
		}
		return "", fmt.Errorf("not found")
	}
	r.calls = calls
	return r
}

// THE NAME TO ASSUME DOES NOT EXIST, so the tool asks.
//
// `.py` named `python`, which a stock modern Linux or macOS box does not have. `python3` is
// not the fix either: on Windows that is typically the Store alias stub while `python.exe` is
// the real one. Measured on a machine with only python3, the suite reported
// `exec: "python": executable file not found in $PATH` from three tests (#522).
func TestTheInterpreterIsResolvedAcrossCandidates(t *testing.T) {
	for _, tc := range []struct {
		name    string
		present []string
		script  string
		wantBin string
	}{
		{"only python3 — the stock modern box", []string{"python3"}, "p.py", "python3"},
		{"only python — an older box, and Windows", []string{"python"}, "p.py", "python"},
		{"both — the versioned name wins", []string{"python3", "python"}, "p.py", "python3"},
		// The three siblings that had the identical defect and had not been noticed.
		{"bash present", []string{"bash"}, "p.sh", "bash"},
		{"only sh — Alpine, and a Windows box without git-bash", []string{"sh"}, "p.sh", "sh"},
		{"node", []string{"node"}, "p.mjs", "node"},
		{"go", []string{"go"}, "p.go", "go"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			argv, err := fakeLook(tc.present...).interpreterFor(tc.script)
			if err != nil {
				t.Fatalf("interpreterFor(%q) with %v = %v", tc.script, tc.present, err)
			}
			if filepath.Base(argv[0]) != tc.wantBin {
				t.Errorf("interpreterFor(%q) ran %q, want %q", tc.script, argv[0], tc.wantBin)
			}
		})
	}

	// `go` still carries its subcommand — the argv is how it runs, not just what runs it.
	argv, err := fakeLook("go").interpreterFor("p.go")
	if err != nil || len(argv) != 2 || argv[1] != "run" {
		t.Errorf("interpreterFor(p.go) = %v, %v — want the `run` subcommand carried through", argv, err)
	}
}

// A MISS IS LOUD, AND IT NAMES THE WAY OUT.
//
// A proof that cannot RUN must never produce something a reader takes for a proof that FAILED.
// And the environment is often not the seat's to change — a fresh cloud container comes back
// the same way every time — while the EXTENSION is, so the refusal routes rather than just
// refusing.
func TestAnUnresolvableInterpreterRefusesAndRoutes(t *testing.T) {
	// python absent, node and bash present: the seat can reach a working proof by changing
	// language, and nothing else was going to tell it so.
	_, err := fakeLook("node", "bash").interpreterFor("compute.py")
	if err == nil {
		t.Fatal("a .py proof was accepted with no python on PATH")
	}
	for _, want := range []string{"compute.py", "python3", "python", "node", "bash", "write the proof in a language"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal does not carry %q:\n%v", want, err)
		}
	}

	// NOTHING available is a different message, because there is no way out by language and
	// the only honest move left is to say so on the record.
	_, err = fakeLook().interpreterFor("compute.py")
	if err == nil {
		t.Fatal("a .py proof was accepted with an empty PATH")
	}
	if !strings.Contains(err.Error(), "friction") {
		t.Errorf("with no interpreter at all the seat is not pointed at the friction channel:\n%v", err)
	}
	if strings.Contains(err.Error(), "DOES have") {
		t.Errorf("an empty environment is offered alternatives it does not have:\n%v", err)
	}
}

// AN UNKNOWN EXTENSION IS STILL REFUSED, and distinguishably: it is the seat naming a language
// the tool does not run, not the environment lacking one it does.
func TestAnUnknownExtensionIsRefusedOnItsOwnTerms(t *testing.T) {
	_, err := fakeLook("python3", "node", "bash", "go").interpreterFor("compute.rb")
	if err == nil {
		t.Fatal("a .rb proof was accepted")
	}
	if !strings.Contains(err.Error(), "no known interpreter") {
		t.Errorf("an unknown extension is reported as an environment miss:\n%v", err)
	}
}

// PATH IS PROBED ONCE PER NAME. Every proof runs TWICE by design and a debate carries dozens,
// so an unmemoised lookup turns one answer into hundreds of identical syscalls — and a
// confirmed MISS is cached too, because that answer is just as stable.
func TestPathIsProbedOncePerName(t *testing.T) {
	r := fakeLook("node")
	for i := 0; i < 5; i++ {
		if _, err := r.interpreterFor("p.py"); err == nil {
			t.Fatal("python resolved when it is not present")
		}
		if _, err := r.interpreterFor("p.mjs"); err != nil {
			t.Fatalf("node did not resolve: %v", err)
		}
	}
	for _, name := range []string{"python3", "python", "node"} {
		if n := r.calls[name]; n != 1 {
			t.Errorf("PATH was probed %d times for %q, want 1", n, name)
		}
	}
}

// The Windows launcher is offered only on Windows, and carries its version selector — without
// it `py` can pick a 2.x it finds first.
func TestTheWindowsLauncherIsOfferedOnlyThere(t *testing.T) {
	argv, err := fakeLook("py").interpreterFor("p.py")
	if runtime.GOOS != "windows" {
		if err == nil {
			t.Errorf("`py` was used off Windows, where it is not the launcher: %v", argv)
		}
		return
	}
	if err != nil {
		t.Fatalf("`py` did not resolve on Windows: %v", err)
	}
	if len(argv) < 2 || argv[len(argv)-1] != "-3" {
		t.Errorf("interpreterFor = %v, want the -3 selector so the launcher cannot pick a 2.x", argv)
	}
}
