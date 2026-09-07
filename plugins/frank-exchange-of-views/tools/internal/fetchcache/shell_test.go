package fetchcache

import (
	"fmt"
	"strings"
	"testing"
)

// A REACT SKELETON IS NOT A DOCUMENT, and this is the case #668 is about: bytes arrive, a 200 is
// returned, a sha is stored, and a seat verifying a citation against it is verifying the mount
// point. The miss reads exactly like an honest check.
func TestAnEmptyMountPointIsNotRenderable(t *testing.T) {
	body := []byte(`<html><head><title>App</title></head><body><div id="root"></div><script src="/a.js"></script></body></html>`)
	got := ShellReason("text/html", body)
	if got == "" {
		t.Fatal("an empty mount point with no prose was reported as a readable document")
	}
	if !strings.Contains(got, "mount point") {
		t.Errorf("the reason must say what was seen: %q", got)
	}
}

// A SHELL USUALLY YIELDS A LITTLE TEXT — the nav, a cookie banner — which is why the flag cannot
// key on "no text at all". The signal is the RELATIONSHIP: what survived against what was sent.
func TestALittleTextUnderALotOfScriptIsNotRenderable(t *testing.T) {
	body := []byte("<html><body><nav>Home About Contact</nav>" + strings.Repeat("<script>var x=1;</script>", 4000) + "</body></html>")
	got := ShellReason("text/html", body)
	if got == "" {
		t.Fatalf("%d characters of nav under %d bytes of script was reported as a document", len("Home About Contact"), len(body))
	}
	if !strings.Contains(got, "%") {
		t.Errorf("the reason must carry the measurement, not just a verdict: %q", got)
	}
}

// A SHORT PAGE IS NOT A SHELL, and this is the assertion that keeps the flag from firing on every
// small honest document — the failure that would teach a seat to ignore it.
func TestAShortRealArticleIsRenderable(t *testing.T) {
	text := strings.Repeat("This is a real sentence of prose. ", 12) // ~400 chars
	body := []byte("<html><body><article><p>" + text + "</p></article></body></html>")
	if got := ShellReason("text/html", body); got != "" {
		t.Errorf("a short article was flagged as a skeleton: %q", got)
	}
}

// A LONG DOCUMENT IS NEVER CONSULTED ON THE RATIO. Above the floor there is enough prose to read
// whatever the markup weighs, and a heavily-scripted news page with a full article in it must not
// be called unreadable.
func TestALongArticleUnderHeavyMarkupIsRenderable(t *testing.T) {
	text := strings.Repeat("Substantial reporting prose, sentence after sentence. ", 60) // > floor
	body := []byte("<html><body>" + strings.Repeat("<script>x</script>", 9000) + "<article>" + text + "</article></body></html>")
	if got := ShellReason("text/html", body); got != "" {
		t.Errorf("a long article was flagged because its page carried scripts: %q", got)
	}
}

// THE QUESTION IS ABOUT HTML. A PDF that extracts badly is a different fact with its own field,
// and folding them would make "this scan is unreadable" and "this page needs a browser" the same
// sentence to a seat that has to act differently on each.
func TestANonHTMLResponseIsNotAskedThisQuestion(t *testing.T) {
	if got := ShellReason("application/pdf", []byte(strings.Repeat("%PDF binary ", 5000))); got != "" {
		t.Errorf("a PDF was judged by the HTML shell rule: %q", got)
	}
}

// The floor and the ratio are stated rather than magic, so a reader can see what would change.
func TestTheThresholdsAreTheOnesDocumented(t *testing.T) {
	if shellTextFloor != 2000 {
		t.Errorf("floor moved to %d — the doc comment above it must move too", shellTextFloor)
	}
	if fmt.Sprintf("%.2f", shellTextRatio) != "0.02" {
		t.Errorf("ratio moved to %v — the doc comment above it must move too", shellTextRatio)
	}
}
