package runlive

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// A RUN IS ONE RUN HOWEVER ITS PATH IS SPELLED. The project directory arrives from the working
// directory, which the OS reports with its symlinks resolved, and the run directory from an argument
// spelled through a link — macOS's /var → /private/var puts every temp directory there. A take-over,
// a re-setup and a capture of that run must all find its row.
func TestARunNamedThroughASymlinkIsTheSameRun(t *testing.T) {
	real := t.TempDir()
	link := filepath.Join(t.TempDir(), "link")
	if err := os.Symlink(real, link); err != nil {
		t.Skipf("this platform makes no symlink here: %v", err)
	}
	mustMkdir(t, filepath.Join(real, "research", "r"))
	if _, err := WriteRunLiveMarker(real, "research/r", nil, time.Now(), "", "", "old-lead"); err != nil {
		t.Fatal(err)
	}
	viaLink := filepath.Join(link, "research", "r")
	if !SameRun(real, "research/r", viaLink) || !SameRun(link, "research/r", filepath.Join(real, "research", "r")) {
		t.Fatalf("research/r and %s are one run", viaLink)
	}
	// A run whose directory is gone still compares through its project's link.
	if !SameRun(real, "research/gone", filepath.Join(link, "research", "gone")) {
		t.Fatal("a run directory that does not exist, named through the link, is not the same run")
	}
	if SameRun(real, "research/r", filepath.Join(link, "research", "s")) {
		t.Fatal("two runs compared as one")
	}
	if _, err := TakeOver(real, viaLink, "old-lead", "new-lead"); err != nil {
		t.Fatalf("a take-over of the run named through a link: %v", err)
	}
	if m, ok := RowOf(link, "research/r"); !ok || m.LeadSession != "new-lead" {
		t.Fatalf("the row after the take-over: %+v, %v", m, ok)
	}
}
