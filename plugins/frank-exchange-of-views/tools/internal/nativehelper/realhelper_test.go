package nativehelper

import (
	"bytes"
	"testing"
	"time"
)

// These run against the built feov-ocr through Require, on every native-run target, and build no
// fake helper: native-run's servercore container runs this package's test binary with no Go
// present.

// TestFrameChannelSurvivesLibraryStdout: a library's writes to stdout, through each of the four
// routes C and C++ code takes, land in the request's captured stderr and never in a frame, and a
// payload of every byte value crosses unchanged. It fails on Linux without the fd move or the
// paired fd 1 capture, and on Windows without binary mode or with frames written to a stdout
// handle read before the redirect.
func TestFrameChannelSurvivesLibraryStdout(t *testing.T) {
	c := Require(t, "feov-ocr")
	r, err := c.Do(OpSelftestStdout)
	if err != nil {
		t.Fatalf("selftest-stdout: %v", err)
	}
	if len(r.Result) != 1 || len(r.Result[0]) != 256 {
		t.Fatalf("selftest-stdout returned %d fields, the first %d bytes; want one field of 256", len(r.Result), len(firstOr(r.Result)))
	}
	for i, b := range r.Result[0] {
		if int(b) != i {
			t.Fatalf("byte %d crossed the frame as %#x", i, b)
		}
	}
	for _, line := range []string{"printf", "cout", "fwrite", "write"} {
		want := []byte("selftest-stdout: " + line + "\n")
		if !bytes.Contains(r.Stderr, want) {
			t.Errorf("the request's captured stderr lacks the %s write %q; it holds:\n%s", line, want, r.Stderr)
		}
	}
	// The helper is still in step: a second request on the same process answers.
	if _, err := c.Do(OpSelftestSleep, le32(1)); err != nil {
		t.Fatalf("the request after selftest-stdout: %v", err)
	}
}

// TestStdoutOutsideARequestReachesStderr: the built helper's stdout writes before its first
// request and between requests, where no request's capture covers fd 1, land on its stderr pipe
// and never in a frame. Without claim()'s fd 1 move the first write corrupts the handshake's
// frame; without Capture::end's, the between-requests writes land in the capture file and are
// truncated away.
func TestStdoutOutsideARequestReachesStderr(t *testing.T) {
	t.Setenv(EnvSelftestStdout, "1")
	c := Require(t, "feov-ocr")
	for i := 0; i < 2; i++ {
		if _, err := c.Do(OpSelftestSleep, le32(1)); err != nil {
			t.Fatalf("request %d after stdout writes outside a request: %v", i+1, err)
		}
	}
	c.mu.Lock()
	p := c.proc
	c.mu.Unlock()
	// The client clears the capture before each request while the helper holds it open; on every
	// target that clear succeeds, so a crash reports its own capture.
	if p.captureStale != nil {
		t.Errorf("clearing the capture the helper holds open failed: %v", p.captureStale)
	}
	if err := c.Close(); err != nil {
		t.Fatalf("feov-ocr on stdin EOF: %v", err)
	}
	// On Windows fd 2 is in text mode, so the stderr pipe carries CRLF; only the capture file and
	// the frame descriptor are binary.
	stderr := bytes.ReplaceAll(p.stderr.Bytes(), []byte("\r\n"), []byte("\n"))
	for _, who := range []string{"selftest-stdout before the first request", "selftest-stdout between requests"} {
		for _, route := range []string{"printf", "cout", "fwrite", "write"} {
			if want := who + ": " + route + "\n"; !bytes.Contains(stderr, []byte(want)) {
				t.Errorf("the helper's stderr lacks %q; it holds:\n%s", want, stderr)
			}
		}
	}
}

// TestHelperExitsOnStdinEOF: the built helper exits 0 within Close's two-second grace when its
// stdin closes, relying on no signal.
func TestHelperExitsOnStdinEOF(t *testing.T) {
	c := Require(t, "feov-ocr")
	start := time.Now()
	if err := c.Close(); err != nil {
		t.Fatalf("feov-ocr on stdin EOF: %v", err)
	}
	if d := time.Since(start); d > closeGrace {
		t.Errorf("feov-ocr took %s to exit on stdin EOF, over %s", d, closeGrace)
	}
}

func firstOr(fields [][]byte) []byte {
	if len(fields) == 0 {
		return nil
	}
	return fields[0]
}
