package compute

import (
	"errors"
	"io/fs"
	"path/filepath"
	"testing"
	"time"
)

// errShared stands in for Windows' ERROR_SHARING_VIOLATION on a platform that never returns it.
var errShared = errors.New("sharing violation")

// refuseWalks has Bundle's walk refused with errShared — or with refusal, when given — for the
// first n attempts and then walk the real store, and counts the attempts.
func refuseWalks(t *testing.T, n int, refusal error) *int {
	t.Helper()
	prevWalk, prevRefused, prevBudget := walkBundle, sharingRefused, sharingBudget
	t.Cleanup(func() { walkBundle, sharingRefused, sharingBudget = prevWalk, prevRefused, prevBudget })
	if refusal == nil {
		refusal = errShared
	}
	calls := 0
	sharingRefused = func(err error) bool { return errors.Is(err, errShared) }
	walkBundle = func(root string, fn fs.WalkDirFunc) error {
		calls++
		if calls <= n {
			return &fs.PathError{Op: "open", Path: root, Err: refusal}
		}
		return filepath.WalkDir(root, fn)
	}
	return &calls
}

// THE LOSING LANE READS dir/ WHILE THE WINNER'S RENAME STILL HOLDS IT (V1a.5 (e)): the read is
// refused with a sharing violation for as long as that handle is open, and Bundle reads it once it
// closes rather than failing the act.
func TestBundleWaitsOutASharingViolation(t *testing.T) {
	runDir := t.TempDir()
	p := program(t, bundle(t, map[string]string{"a": "a"}), "print", "7")
	if err := WriteBundle(runDir, p.Sha, p.Argv, p.Files); err != nil {
		t.Fatal(err)
	}
	calls := refuseWalks(t, 2, nil)
	files, err := Bundle(runDir, p.Sha)
	if err != nil {
		t.Fatalf("a sharing violation that clears failed the read: %v", err)
	}
	if BundleSha(p.Argv, Sums(files)) != p.Sha {
		t.Errorf("the read after the violation cleared does not hash to %s", Short(p.Sha))
	}
	if *calls != 3 {
		t.Errorf("Bundle walked %d times, want 3: two refused, one read", *calls)
	}
	// The loser's own path: WriteBundle finds dir/ present and checks it through the same read.
	*calls = 0
	if err := WriteBundle(runDir, p.Sha, p.Argv, p.Files); err != nil {
		t.Errorf("the second lane's write failed on a sharing violation that clears: %v", err)
	}
}

// Every refusal that is not a sharing violation returns at once: a retry there would only delay
// the error.
func TestBundleRetriesNothingButASharingViolation(t *testing.T) {
	runDir := t.TempDir()
	p := program(t, bundle(t, map[string]string{"a": "a"}), "print", "7")
	if err := WriteBundle(runDir, p.Sha, p.Argv, p.Files); err != nil {
		t.Fatal(err)
	}
	other := errors.New("permission denied")
	calls := refuseWalks(t, 1, other)
	if _, err := Bundle(runDir, p.Sha); !errors.Is(err, other) {
		t.Fatalf("Bundle returned %v, want the refusal %v", err, other)
	}
	if *calls != 1 {
		t.Errorf("Bundle walked %d times on an error that is not a sharing violation, want 1", *calls)
	}
}

// A sharing violation that never clears is returned once the budget runs out.
func TestBundleSharingRetryIsBounded(t *testing.T) {
	runDir := t.TempDir()
	p := program(t, bundle(t, map[string]string{"a": "a"}), "print", "7")
	if err := WriteBundle(runDir, p.Sha, p.Argv, p.Files); err != nil {
		t.Fatal(err)
	}
	refuseWalks(t, 1<<30, nil)
	sharingBudget = 20 * time.Millisecond
	start := time.Now()
	if _, err := Bundle(runDir, p.Sha); !errors.Is(err, errShared) {
		t.Fatalf("Bundle returned %v, want the standing sharing violation", err)
	}
	if took := time.Since(start); took > 5*time.Second {
		t.Errorf("Bundle waited %v on a %v budget", took, sharingBudget)
	}
}
