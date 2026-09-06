package tessocr

import (
	"os"
	"strings"
	"testing"
)

// A WINDOWS CHECKOUT IS A VALID INPUT. The band fixture arrives with \r\n there, and the
// un-normalized parser collapsed eleven headers into one — CI's Windows leg caught it. The
// same text must parse identically under both endings.
func TestRotatedBandParsesUnderCRLF(t *testing.T) {
	b, err := os.ReadFile("testdata/p0054.headers.300.txt")
	if err != nil {
		t.Fatal(err)
	}
	lf := ParseRotatedBandHeaders(string(b))
	crlf := ParseRotatedBandHeaders(strings.ReplaceAll(strings.ReplaceAll(string(b), "\r\n", "\n"), "\n", "\r\n"))
	if len(lf) != len(crlf) {
		t.Fatalf("endings changed the parse: %d headers under LF, %d under CRLF", len(lf), len(crlf))
	}
	if len(lf) != 11 {
		t.Fatalf("fixture parsed to %d headers, want 11: %q", len(lf), lf)
	}
	for i := range lf {
		if lf[i] != crlf[i] {
			t.Errorf("header %d differs by line ending: %q vs %q", i, lf[i], crlf[i])
		}
	}
}
