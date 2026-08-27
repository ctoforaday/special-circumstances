package setup

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatenv"
)

// fakeBin writes an executable that prints the wrapper variable it was handed, so the test can
// assert what the SHELL actually delivered rather than what the generator intended to write.
func fakeBin(t *testing.T, dir, name string) string {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(dir, name)
	body := "#!/bin/sh\nprintf '%s|%s' \"$" + seatenv.VarWrapper + "\" \"$*\"\n"
	if err := os.WriteFile(p, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
	return p
}

// THE WRAPPER IS EXECUTED, not read. A generator test that greps its own output proves the
// string was written and nothing about whether a shell would honour it — and the value here is
// a path that must survive quoting, which is exactly where a written-string assertion passes
// and the real thing breaks.
func TestTheWrapperDeliversTheRunToTheRealBinary(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the POSIX wrapper needs /bin/sh; the .cmd twin is asserted separately")
	}
	base := t.TempDir()
	target := fakeBin(t, filepath.Join(base, "runbin"), "feov-record")
	// A run directory with a space in it: the shape that broke a live seat on 2026-08-05, and
	// the reason the value is single-quoted rather than interpolated bare.
	run := filepath.Join(base, "special circumstances", "research", "2026-08-23_x")
	if err := os.MkdirAll(run, 0o755); err != nil {
		t.Fatal(err)
	}

	res := WriteRunWrapper(runOf(t, run), target)
	if res.Skipped != "" {
		t.Fatalf("WriteRunWrapper skipped: %s", res.Skipped)
	}
	if want := filepath.Join(run, WrapperDir); res.BinDir != want {
		t.Errorf("BinDir = %q, want %q", res.BinDir, want)
	}

	// Invoked exactly as the dispatch prompt does: binDir + "/feov-record", plus a seat's args.
	out, err := exec.Command(filepath.Join(res.BinDir, WrapperName), "show", "board").CombinedOutput()
	if err != nil {
		t.Fatalf("running the wrapper: %v (%s)", err, out)
	}
	gotRun, gotArgs, _ := strings.Cut(string(out), "|")
	if gotRun != run {
		t.Errorf("the wrapper delivered %s=%q, want %q", seatenv.VarWrapper, gotRun, run)
	}
	if gotArgs != "show board" {
		t.Errorf("the wrapper delivered args %q, want %q — a seat's own arguments must pass through untouched", gotArgs, "show board")
	}
}

// AND IT NEVER WRAPS A WRAPPER. Pointing --bin-dir at a previous run's .bin would produce a
// shim execing a shim that sets a DIFFERENT run's directory last — so the inner one wins, and
// the seat files against the wrong run through the very mechanism built to stop that.
func TestTheWrapperRefusesToWrapAWrapper(t *testing.T) {
	base := t.TempDir()
	target := fakeBin(t, filepath.Join(base, "runbin"), "feov-record")
	first := filepath.Join(base, "run-a")
	if err := os.MkdirAll(first, 0o755); err != nil {
		t.Fatal(err)
	}
	res := WriteRunWrapper(runOf(t, first), target)
	if res.Skipped != "" {
		t.Fatalf("first wrapper skipped: %s", res.Skipped)
	}

	second := filepath.Join(base, "run-b")
	if err := os.MkdirAll(second, 0o755); err != nil {
		t.Fatal(err)
	}
	again := WriteRunWrapper(runOf(t, second), filepath.Join(res.BinDir, WrapperName))
	if again.Skipped == "" {
		t.Fatalf("wrapping run-a's wrapper for run-b was accepted (BinDir %q) — run-b's seats would file against run-a", again.BinDir)
	}
	if !strings.Contains(again.Skipped, "real binary") {
		t.Errorf("the refusal does not say what to point --bin-dir at: %s", again.Skipped)
	}
}

// A MISS IS STATED, never returned as an empty path the summary can print as success.
func TestAnUnresolvableBinaryIsReportedNotSwallowed(t *testing.T) {
	run := t.TempDir()
	for _, tc := range []struct{ name, bin string }{
		{"empty", ""},
		{"a directory, not a file", run},
		{"a path that does not exist", filepath.Join(run, "nope", "feov-record")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			res := WriteRunWrapper(runOf(t, run), tc.bin)
			if res.Skipped == "" {
				t.Errorf("WriteRunWrapper(runOf(t, %q)) returned BinDir %q with no reason", tc.bin, res.BinDir)
			}
			if res.BinDir != "" {
				t.Errorf("a skipped wrapper still reported BinDir %q, which the summary would print as usable", res.BinDir)
			}
		})
	}
}

// THE WINDOWS TWIN EXISTS AND CARRIES THE SAME RUN. The CI matrix runs windows-latest; a
// platform with no wrapper is a platform that silently falls back to seats typing --run.
func TestTheWindowsTwinIsWrittenWithTheSameRun(t *testing.T) {
	base := t.TempDir()
	target := fakeBin(t, filepath.Join(base, "runbin"), "feov-record")
	run := filepath.Join(base, "run")
	if err := os.MkdirAll(run, 0o755); err != nil {
		t.Fatal(err)
	}
	res := WriteRunWrapper(runOf(t, run), target)
	if res.Skipped != "" {
		t.Fatalf("WriteRunWrapper skipped: %s", res.Skipped)
	}
	b, err := os.ReadFile(filepath.Join(res.BinDir, WrapperName+".cmd"))
	if err != nil {
		t.Fatalf("no .cmd twin: %v", err)
	}
	body := string(b)
	for _, want := range []string{seatenv.VarWrapper + "=" + run, "%*"} {
		if !strings.Contains(body, want) {
			t.Errorf(".cmd does not carry %q:\n%s", want, body)
		}
	}
	if !strings.Contains(body, "\r\n") {
		t.Errorf(".cmd is not CRLF-terminated, which cmd.exe mis-parses:\n%q", body)
	}
}
