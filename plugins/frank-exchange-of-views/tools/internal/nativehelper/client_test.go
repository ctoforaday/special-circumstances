package nativehelper

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func le32(v uint32) []byte { return binary.LittleEndian.AppendUint32(nil, v) }

func fakeClient(t *testing.T, digest string) *Client {
	t.Helper()
	c := newClient(fakeName, Pins{Digest: digest})
	t.Cleanup(func() { _ = c.Close() })
	return c
}

func asError(t *testing.T, err error, kind error) *Error {
	t.Helper()
	if !errors.Is(err, kind) {
		t.Fatalf("got %v, want %v", err, kind)
	}
	var e *Error
	if !errors.As(err, &e) {
		t.Fatalf("%v is not an *Error", err)
	}
	return e
}

func TestASelftestRoundTrips(t *testing.T) {
	t.Setenv(EnvDir, fake(t, "d1", ProtocolVersion, ""))
	starts := startLog(t)
	c := fakeClient(t, "d1")
	r, err := c.Do(OpSelftestStdout)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Result) != 1 || len(r.Result[0]) != 256 || len(r.Stderr) == 0 {
		t.Fatalf("selftest-stdout: %+v", r)
	}
	for i, b := range r.Result[0] {
		if int(b) != i {
			t.Fatalf("byte %d is %#x", i, b)
		}
	}
	_, err = c.Do(Op(99))
	var oe *OpError
	if !errors.As(err, &oe) || oe.Status != StatusUnknownOp {
		t.Fatalf("an unknown op returns %v, want an *OpError with unknown-op", err)
	}
	if _, err := c.Do(OpSelftestSleep, le32(1)); err != nil {
		t.Fatalf("an op refusal took the helper down: %v", err)
	}
	if n := starts(); n != 1 {
		t.Errorf("the helper started %d times, want 1: an op refusal is not a helper failure", n)
	}
}

func TestACrashedRequestIsNotRetried(t *testing.T) {
	t.Setenv(EnvDir, fake(t, "d1", ProtocolVersion, "crash"))
	starts := startLog(t)
	c := fakeClient(t, "d1")
	_, err := c.Do(OpSelftestStdout)
	e := asError(t, err, ErrCrashed)
	// The fake writes "crashing as built to" into the capture file the client names, never to its
	// stderr pipe: the client reads it back from there.
	if e.Op != OpSelftestStdout || !strings.Contains(e.Exit, "3") || !bytes.Contains(e.Stderr, []byte("crashing as built to")) {
		t.Errorf("the crash does not name its op, exit status and captured stderr: %v", err)
	}
	if e.Handshake {
		t.Errorf("a crash after a good handshake is marked as the helper never coming up")
	}
	if n := starts(); n != 1 {
		t.Fatalf("one crashed request started the helper %d times, want 1: the request was resent", n)
	}
	_, err = c.Do(OpSelftestStdout)
	asError(t, err, ErrCrashed)
	if n := starts(); n != 2 {
		t.Errorf("the next request started %d helpers in all, want 2: one fresh helper per request after a crash", n)
	}
}

func TestAFailedHandshakeLatches(t *testing.T) {
	for _, tc := range []struct {
		name    string
		digest  string
		version uint16
		kind    error
	}{
		{"another protocol version", "d1", ProtocolVersion + 1, ErrProtocol},
		{"another source digest", "other", ProtocolVersion, ErrDigestMismatch},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(EnvDir, fake(t, tc.digest, tc.version, ""))
			starts := startLog(t)
			c := fakeClient(t, "d1")
			for i := 0; i < 3; i++ {
				_, err := c.Do(OpSelftestSleep, le32(1))
				e := asError(t, err, tc.kind)
				if !e.Handshake {
					t.Errorf("a failed handshake is not marked as the helper never coming up")
				}
				if tc.kind == ErrDigestMismatch && (!strings.Contains(err.Error(), "other") || !strings.Contains(err.Error(), "d1")) {
					t.Errorf("the mismatch does not name both digests: %v", err)
				}
			}
			if n := starts(); n != 1 {
				t.Errorf("three requests after a failed handshake spawned %d helpers, want 1", n)
			}
		})
	}
}

func TestHelperExitsOnStdinEOF(t *testing.T) {
	t.Setenv(EnvDir, fake(t, "d1", ProtocolVersion, ""))
	c := fakeClient(t, "d1")
	if err := c.Start(); err != nil {
		t.Fatal(err)
	}
	if err := c.Close(); err != nil {
		t.Fatalf("the helper did not exit 0 on stdin EOF: %v", err)
	}
	// The control: a helper that ignores EOF is killed, and Close says so.
	t.Setenv(EnvDir, fake(t, "d1", ProtocolVersion, "ignore-eof"))
	pids := startPIDs(t)
	c = fakeClient(t, "d1")
	if err := c.Start(); err != nil {
		t.Fatal(err)
	}
	start := time.Now()
	if err := c.Close(); err == nil || !strings.Contains(err.Error(), "killed") {
		t.Fatalf("a helper that ignores EOF closed with %v, want a kill", err)
	}
	if d := time.Since(start); d > 10*time.Second {
		t.Errorf("Close took %s to kill a helper that ignores EOF", d)
	}
	if started := pids(); len(started) != 1 {
		t.Errorf("the start log holds %v, want one pid", started)
	} else {
		assertGone(t, started[0])
	}
}

func TestADeadlineKillsAndNames(t *testing.T) {
	t.Setenv(EnvDir, fake(t, "d1", ProtocolVersion, ""))
	starts := startLog(t)
	c := fakeClient(t, "d1")
	if err := c.Start(); err != nil {
		t.Fatal(err)
	}
	c.deadline = 300 * time.Millisecond
	_, err := c.Do(OpSelftestSleep, le32(20000))
	e := asError(t, err, ErrDeadline)
	if e.Op != OpSelftestSleep || e.Elapsed < c.deadline || e.Elapsed > 10*time.Second {
		t.Errorf("the deadline error names op %s after %s: %v", e.Op, e.Elapsed, err)
	}
	if !strings.Contains(err.Error(), "selftest-sleep") {
		t.Errorf("the message does not name the op: %v", err)
	}
	c.deadline = DefaultDeadline
	if _, err := c.Do(OpSelftestSleep, le32(1)); err != nil {
		t.Fatalf("the request after a deadline failed: %v", err)
	}
	if n := starts(); n != 2 {
		t.Errorf("%d helpers started, want 2: the overrun one killed, one fresh", n)
	}
}

// TestAProtocolErrorKillsAndNames: every response the client refuses is helper-protocol naming
// the op and the defect, never a panic or an *OpError, and the helper is gone afterwards.
func TestAProtocolErrorKillsAndNames(t *testing.T) {
	for _, tc := range []struct{ mode, detail string }{
		{"garbage", "magic"},
		{"wrong-id", "response id"},
		{"no-fields", "without its status and stderr"},
		{"one-field", "without its status and stderr"},
		{"empty-status", "without its status and stderr"},
		{"status-7", "status-7"},
	} {
		t.Run(tc.mode, func(t *testing.T) {
			t.Setenv(EnvDir, fake(t, "d1", ProtocolVersion, tc.mode))
			pids := startPIDs(t)
			c := fakeClient(t, "d1")
			_, err := c.Do(OpSelftestStdout)
			e := asError(t, err, ErrProtocol)
			if e.Op != OpSelftestStdout || !strings.Contains(e.Detail, tc.detail) {
				t.Errorf("the protocol error does not name its op and %q: %v", tc.detail, err)
			}
			started := pids()
			if len(started) != 1 {
				t.Fatalf("the start log holds %v, want one pid", started)
			}
			assertGone(t, started[0])
		})
	}
}

// floodTail is the fake's last writes in a flood request: the flood's last line, then the
// selftest's own diagnostics.
const floodTail = "fakehelper: the flood's last line\nprintf\ncout\nfwrite\nwrite1\n"

// TestAStderrFloodIsBoundedAndMarked: a request that writes far past StderrKeep to stderr still
// succeeds, and its stderr field is the last StderrKeep bytes after a line naming the cut.
func TestAStderrFloodIsBoundedAndMarked(t *testing.T) {
	t.Setenv(EnvDir, fake(t, "d1", ProtocolVersion, "flood"))
	c := fakeClient(t, "d1")
	r, err := c.Do(OpSelftestStdout)
	if err != nil {
		t.Fatalf("a request that flooded stderr failed: %v", err)
	}
	marker := fmt.Sprintf("frame: %d earlier bytes of stderr dropped\n", 1<<20+len(floodTail)-StderrKeep)
	if len(r.Stderr) != len(marker)+StderrKeep || !bytes.HasPrefix(r.Stderr, []byte(marker)) {
		t.Errorf("the stderr field is %d bytes starting %.80q; want %d bytes after %q", len(r.Stderr), r.Stderr, len(marker)+StderrKeep, marker)
	}
	if !bytes.HasSuffix(r.Stderr, []byte(floodTail)) {
		t.Errorf("the stderr field lost the request's last writes: ends %q", r.Stderr[max(len(r.Stderr)-80, 0):])
	}
}

// TestACrashAfterAFloodKeepsTheHelpersLastWords: a helper that floods its capture and then dies
// writing to its stderr pipe is helper-crashed carrying both, the pipe's last line intact.
func TestACrashAfterAFloodKeepsTheHelpersLastWords(t *testing.T) {
	t.Setenv(EnvDir, fake(t, "d1", ProtocolVersion, "flood-crash"))
	c := fakeClient(t, "d1")
	_, err := c.Do(OpSelftestStdout)
	e := asError(t, err, ErrCrashed)
	for _, want := range []string{"writing a frame failed, as built to", "fakehelper: the flood's last line", "earlier bytes of stderr dropped"} {
		if !bytes.Contains(e.Stderr, []byte(want)) {
			t.Errorf("the crash's stderr lacks %q (%d bytes)", want, len(e.Stderr))
		}
	}
	if len(e.Stderr) > 2*StderrKeep+200 {
		t.Errorf("the crash carries %d bytes of stderr, over the bound", len(e.Stderr))
	}
}

// TestAWrongAssetHashIsADigestMismatch: the handshake's asset sha256 is checked against this
// binary's pin for its arch, and a mismatch names both.
func TestAWrongAssetHashIsADigestMismatch(t *testing.T) {
	pins := Pins{Digest: "d1", Assets: []AssetPin{{File: "model.ort", SHA256: map[string]string{runtime.GOARCH: "aaaa"}}}}
	t.Setenv(EnvDir, fake(t, "d1", ProtocolVersion, "", "bbbb"))
	c := newClient(fakeName, pins)
	t.Cleanup(func() { _ = c.Close() })
	err := c.Start()
	e := asError(t, err, ErrDigestMismatch)
	if e.Want != "aaaa" || e.Got != "bbbb" || !strings.Contains(err.Error(), "model.ort") || !strings.Contains(err.Error(), "aaaa") || !strings.Contains(err.Error(), "bbbb") {
		t.Errorf("the mismatch does not name the asset and both hashes: %v", err)
	}
	// The control: the pinned hash passes.
	t.Setenv(EnvDir, fake(t, "d1", ProtocolVersion, "", "aaaa"))
	c = newClient(fakeName, pins)
	t.Cleanup(func() { _ = c.Close() })
	if err := c.Start(); err != nil {
		t.Fatalf("the pinned asset hash is refused: %v", err)
	}
}

// TestTheDeadlineExcludesTheWaitForTheLock: a request queued behind another for longer than the
// deadline still has its whole deadline once it holds the lock.
func TestTheDeadlineExcludesTheWaitForTheLock(t *testing.T) {
	t.Setenv(EnvDir, fake(t, "d1", ProtocolVersion, ""))
	c := fakeClient(t, "d1")
	if err := c.Start(); err != nil {
		t.Fatal(err)
	}
	c.deadline = time.Second
	c.mu.Lock() // the first holder: a request in flight for two and a half deadlines
	done := make(chan error, 1)
	go func() {
		_, err := c.Do(OpSelftestSleep, le32(1))
		done <- err
	}()
	time.Sleep(c.deadline * 5 / 2)
	c.mu.Unlock()
	if err := <-done; err != nil {
		t.Fatalf("a request that waited for the lock failed: %v", err)
	}
}

// TestALookedUpFileMustBeExecutable: a regular file without an execute bit at the looked-up path
// is helper-absent, under the override and beside the executable, and it does not latch.
func TestALookedUpFileMustBeExecutable(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("SKIPPED, NOT PASSED: on Windows a file under the .exe name is executable; there is no execute bit to withhold")
	}
	b, err := os.ReadFile(filepath.Join(fake(t, "d1", ProtocolVersion, ""), exe(fakeName)))
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	p := filepath.Join(dir, exe(fakeName))
	if err := os.WriteFile(p, b, 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv(EnvDir, dir)
	c := fakeClient(t, "d1")
	e := asError(t, c.Start(), ErrAbsent)
	if e.Path != p || !strings.Contains(e.Detail, "execute bit") || e.Handshake {
		t.Errorf("the absent helper does not name the path and the missing bit: %v", e)
	}
	os.Unsetenv(EnvDir)
	c.exeDir = func() (string, error) { return dir, nil }
	asError(t, c.Start(), ErrAbsent)
	if err := os.Chmod(p, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := c.Start(); err != nil {
		t.Fatalf("helper-absent latched: with the bit set, the helper still does not start: %v", err)
	}
}

// TestAnEmptyOverrideIsAbsent: NATIVE_HELPER_DIR set to the empty string is set, so it is
// helper-absent naming the variable, never a lookup beside the executable.
func TestAnEmptyOverrideIsAbsent(t *testing.T) {
	beside := fake(t, "d1", ProtocolVersion, "")
	t.Setenv(EnvDir, "")
	c := fakeClient(t, "d1")
	c.exeDir = func() (string, error) { return beside, nil }
	err := c.Start()
	asError(t, err, ErrAbsent)
	if !strings.Contains(err.Error(), EnvDir+" is set to the empty string") {
		t.Errorf("the empty override is not named: %v", err)
	}
	// The control: unset, the helper beside the executable starts.
	os.Unsetenv(EnvDir)
	if err := c.Start(); err != nil {
		t.Fatalf("with the override unset, the helper beside the executable is not found: %v", err)
	}
}

func TestTheOverrideHasNoFallback(t *testing.T) {
	beside := t.TempDir()
	b, err := os.ReadFile(filepath.Join(fake(t, "d1", ProtocolVersion, ""), exe(fakeName)))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(beside, exe(fakeName)), b, 0o755); err != nil {
		t.Fatal(err)
	}
	empty := t.TempDir()
	t.Setenv(EnvDir, empty)
	c := fakeClient(t, "d1")
	c.exeDir = func() (string, error) { return beside, nil }
	err = c.Start()
	e := asError(t, err, ErrAbsent)
	if e.Path != filepath.Join(empty, exe(fakeName)) || !strings.Contains(err.Error(), EnvDir) {
		t.Errorf("the absent helper does not name the override: %v", err)
	}
	// The control: unset, the same client finds the helper beside the executable.
	os.Unsetenv(EnvDir)
	if err := c.Start(); err != nil {
		t.Fatalf("with the override unset, the helper beside the executable is not found: %v", err)
	}
	t.Setenv(EnvDir, "none")
	c = fakeClient(t, "d1")
	c.exeDir = func() (string, error) { return beside, nil }
	asError(t, c.Start(), ErrAbsent)
}

// recordTB is a TB that records a Fatalf instead of ending the test.
type recordTB struct{ fatal string }

func (r *recordTB) Helper()                   {}
func (r *recordTB) Fatalf(f string, a ...any) { r.fatal = fmt.Sprintf(f, a...) }
func (r *recordTB) Skipf(f string, a ...any)  { r.fatal = "skipped: " + fmt.Sprintf(f, a...) }
func (r *recordTB) Cleanup(func())            {}

// TestAHelperWithoutPinsIsRefused: a binary carries pins only for the helpers under native/, and
// asking for any other is an error naming it, at New and at Require.
func TestAHelperWithoutPinsIsRefused(t *testing.T) {
	if _, err := New("feov-nonesuch"); err == nil || !strings.Contains(err.Error(), "feov-nonesuch") {
		t.Errorf("New for a helper without pins returned %v", err)
	}
	r := &recordTB{}
	if c := Require(r, "feov-nonesuch"); c != nil || !strings.Contains(r.fatal, "feov-nonesuch") || strings.HasPrefix(r.fatal, "skipped") {
		t.Errorf("Require for a helper without pins: client %v, %q; want a failure naming it", c, r.fatal)
	}
}

// TestRequireFailsInCIWithoutAHelper runs this test binary again as a child under each setting
// and reads the child's verdict.
func TestRequireFailsInCIWithoutAHelper(t *testing.T) {
	if os.Getenv("NATIVEHELPER_REQUIRE_CHILD") == "1" {
		require(t, fakeName, Pins{Digest: "d1"})
		return
	}
	helper := fake(t, "d1", ProtocolVersion, "")
	for _, tc := range []struct {
		name, ci, dir, want string
		set                 bool
	}{
		{"CI without the variable fails", "true", "", "--- FAIL", false},
		{"CI with none skips", "true", "none", "--- SKIP", false},
		{"CI with a mistyped directory fails", "true", helper + "-typo", "--- FAIL", false},
		{"outside CI without the variable skips", "", "", "--- SKIP", false},
		{"outside CI with the variable empty fails", "", "", "--- FAIL", true},
		{"CI with the helper passes", "true", helper, "--- PASS", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.run=^TestRequireFailsInCIWithoutAHelper$", "-test.v", "-test.count=1")
			env := []string{"NATIVEHELPER_REQUIRE_CHILD=1"}
			for _, kv := range os.Environ() {
				if !strings.HasPrefix(kv, "GITHUB_ACTIONS=") && !strings.HasPrefix(kv, EnvDir+"=") {
					env = append(env, kv)
				}
			}
			if tc.ci != "" {
				env = append(env, "GITHUB_ACTIONS="+tc.ci)
			}
			if tc.dir != "" || tc.set {
				env = append(env, EnvDir+"="+tc.dir)
			}
			cmd.Env = env
			out, err := cmd.CombinedOutput()
			if !strings.Contains(string(out), tc.want) {
				t.Fatalf("the child did not report %s (exit %v):\n%s", tc.want, err, out)
			}
			if (tc.want == "--- FAIL") != (err != nil) {
				t.Errorf("the child's exit status (%v) disagrees with its verdict %s", err, tc.want)
			}
		})
	}
}
