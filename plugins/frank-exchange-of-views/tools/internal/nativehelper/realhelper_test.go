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
