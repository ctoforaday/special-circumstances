package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
)

// armed plants a file and arms a restorer with its original bytes, returning both.
func armed(t *testing.T, body string) (*restorer, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "x.go")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	var r restorer
	r.arm(path, []byte(body))
	return &r, path
}

// THE UNDO IS THE SAFETY PROPERTY, so it is asserted on the file's bytes, not on a return.
func TestRestorePutsTheOriginalBack(t *testing.T) {
	r, path := armed(t, "package p\n\nfunc f() bool { return a == b }\n")
	if err := os.WriteFile(path, []byte("package p\n\nfunc f() bool { return a != b }\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := restore(r)
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	if got != path {
		t.Errorf("restore named %q, want the armed path %q", got, path)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), "!=") {
		t.Errorf("the mutant survived the restore: %s", b)
	}
}

// "NOTHING WAS ARMED" AND "THE RESTORE SUCCEEDED" MUST NOT LOOK THE SAME.
//
// Both are a nil error, so the path is what separates them: the deferred net runs on every
// exit, including sweeps that never mutated anything, and a caller that cannot tell the two
// apart cannot report honestly either.
func TestRestoreWithNothingArmedIsANamedNoOp(t *testing.T) {
	var r restorer
	p, err := restore(&r)
	if err != nil {
		t.Fatalf("an unarmed restorer must not error: %v", err)
	}
	if p != "" {
		t.Errorf("an unarmed restorer named %q; empty means nothing was mutated", p)
	}
}

// A second restore is a no-op, so the deferred net cannot undo a file the loop already put
// back — and cannot clobber a NEWER write to the same path with stale bytes.
func TestRestoreTwiceDoesNotRewrite(t *testing.T) {
	r, path := armed(t, "original\n")
	if _, err := restore(r); err != nil {
		t.Fatal(err)
	}
	// Something else writes the file after the sweep is done with it.
	if err := os.WriteFile(path, []byte("written later by someone else\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	p, err := restore(r)
	if err != nil {
		t.Fatal(err)
	}
	if p != "" {
		t.Errorf("the second restore acted on %q; it must be disarmed", p)
	}
	b, _ := os.ReadFile(path)
	if string(b) != "written later by someone else\n" {
		t.Errorf("the second restore clobbered a later write: %q", b)
	}
}

// A FAILED RESTORE MUST BE LOUD, NAMED, AND STILL ARMED.
//
// This is the defect the file exists for. The error was discarded and the interrupt handler
// printed "interrupted — file restored" regardless, so the single moment a mutant is left in
// tracked source was the single moment the tool said everything was fine.
func TestAFailedRestoreNamesTheFileAndStaysArmed(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("a read-only directory does not block writes the same way on Windows")
	}
	if os.Geteuid() == 0 {
		t.Skip("root writes through read-only permissions, so the failure cannot be staged")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "x.go")
	if err := os.WriteFile(path, []byte("original\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var r restorer
	r.arm(path, []byte("original\n"))
	// Read-only DIRECTORY: os.WriteFile opens with O_TRUNC on a file it cannot replace.
	if err := os.Chmod(path, 0o444); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(path, 0o644) })

	p, err := restore(&r)
	if err == nil {
		t.Fatal("a restore that could not write must NOT report success — that is the whole defect")
	}
	if p != path {
		t.Errorf("the failure named %q; it must name the file that still holds the mutant (%q)", p, path)
	}
	for _, want := range []string{"RESTORE FAILED", path, "git checkout --"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the message must carry %q so it is actionable: %v", want, err)
		}
	}
	// Still armed: a later attempt must know which file is dirty rather than forget.
	if r.path != path {
		t.Errorf("a failed restore disarmed itself (path=%q); the file would be forgotten", r.path)
	}
}

// THE RACE, AS A TEST. Run under -race this fails against the unsynchronised original: the
// mutation loop assigned two package-locals while the signal goroutine read them, so an
// interrupt landing between the assignments could pair one file's path with another file's
// bytes and write the wrong source over a real file.
//
// The scripts module now has a -race CI leg (it was the only module without one, which is
// why this went unseen), so this test is exercised where it can actually fail.
func TestArmAndRestoreAreSafeUnderConcurrency(t *testing.T) {
	dir := t.TempDir()
	paths := make([]string, 8)
	for i := range paths {
		paths[i] = filepath.Join(dir, string(rune('a'+i))+".go")
		if err := os.WriteFile(paths[i], []byte("original "+paths[i]+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	var r restorer
	var wg sync.WaitGroup
	wg.Add(2)
	// The sweep: arms each file in turn, as the mutation loop does.
	go func() {
		defer wg.Done()
		for round := 0; round < 200; round++ {
			for _, p := range paths {
				r.arm(p, []byte("original "+p+"\n"))
			}
		}
	}()
	// The signal handler: restores at an arbitrary moment, repeatedly.
	go func() {
		defer wg.Done()
		for i := 0; i < 200; i++ {
			if _, err := restore(&r); err != nil {
				t.Errorf("concurrent restore: %v", err)
				return
			}
		}
	}()
	wg.Wait()

	// Whatever the interleaving, no file may hold another file's bytes: every restore
	// writes the body that was armed WITH the path it was armed with.
	if _, err := restore(&r); err != nil {
		t.Fatal(err)
	}
	for _, p := range paths {
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		if want := "original " + p + "\n"; string(b) != want {
			t.Errorf("%s holds another file's bytes: %q, want %q", p, b, want)
		}
	}
}
