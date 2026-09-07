//go:build tessocr && cgo

package fetchcache

import (
	"os"
	"testing"
)

// THE OTHER HALF OF THE ENGINE SEAM. Its stub twin (enginestub_test.go) pins the refusal
// when the engine is absent; nothing pinned the SUCCESS path, and that arm carries the
// same plausible zero in reverse — invert the error check after a successful read and
// the engine returns an empty page with a nil error, which every caller above records as
// a page that was read and found blank. A blank page and an unread one are different
// facts, and only this build configuration can tell them apart.
//
// The fixture is tessocr's own crop rather than a second copy: one image, one place to
// re-cut it if the corpus moves.
func TestTheEngineBuildReturnsTextItActuallyRead(t *testing.T) {
	png, err := os.ReadFile("../tessocr/testdata/gridcrop.png")
	if err != nil {
		t.Fatal(err)
	}
	eng := &TessocrPageEngine{}
	res, err := eng.ReadPage(png)
	if err != nil {
		t.Fatalf("the engine build failed to read a known-good page: %v", err)
	}
	if res.Text == "" {
		t.Fatal("the engine returned an empty page with no error — a page nobody read is " +
			"indistinguishable from a blank one once it reaches the record")
	}
	if eng.Identity() == "" {
		t.Error("the engine reported no identity; the record's re-derivation claim rests on it")
	}
}
