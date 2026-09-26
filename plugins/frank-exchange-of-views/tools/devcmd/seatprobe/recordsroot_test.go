package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRecordsBaseIsTheHomeScratchArea(t *testing.T) {
	got := recordsBase(filepath.Join("home", "someone"))
	want := filepath.Join("home", "someone", ".claude", "scratch", "seatprobe")
	if got != want {
		t.Errorf("records base = %q, want %q", got, want)
	}
}

// THE ROOT IS KEPT, SO WHERE IT IS KEPT IS THE POINT. A probe's records outlive the run on
// purpose; under the system temp directory they outlive it only until the next sweep or reboot.
func TestNewRecordsRootIsUnderTheScratchAreaAndNotTheSystemTemp(t *testing.T) {
	home := t.TempDir()
	root, err := newRecordsRootIn(home)
	if err != nil {
		t.Fatalf("no record root: %v", err)
	}
	if !strings.HasPrefix(root, recordsBase(home)+string(os.PathSeparator)) {
		t.Errorf("record root %q is not under the scratch base %q", root, recordsBase(home))
	}
	if st, err := os.Stat(root); err != nil || !st.IsDir() {
		t.Errorf("the record root was not created: %v", err)
	}
	// THE SEGMENTS ARE PINNED, NOT THE PREFIX. A fake home is itself inside the system temp
	// directory on most boxes, so "is it under os.TempDir()" cannot fire here and would read as a
	// clean board; what distinguishes the fixed root from the old one is the scratch path itself.
	if want := filepath.Join(".claude", "scratch", "seatprobe"); !strings.Contains(root, want) {
		t.Errorf("the record root %q does not name the scratch area %q", root, want)
	}
}

func TestNewRecordsRootsDoNotCollide(t *testing.T) {
	home := t.TempDir()
	a, err := newRecordsRootIn(home)
	if err != nil {
		t.Fatal(err)
	}
	b, err := newRecordsRootIn(home)
	if err != nil {
		t.Fatal(err)
	}
	if a == b {
		t.Errorf("two runs shared a record root (%s): one run's records would read as the other's", a)
	}
}
