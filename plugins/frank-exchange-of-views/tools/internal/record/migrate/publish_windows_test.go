package migrate

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// A junction at --to is refused before anything is created, and its target is untouched: Lstat
// reports it irregular, which the allow-list does not admit.
func TestAJunctionAtToIsRefused(t *testing.T) {
	src := tinySource(t, false)
	base := recordtest.TmpRun(t)
	target := filepath.Join(base, "target")
	if err := os.Mkdir(target, 0o755); err != nil {
		t.Fatal(err)
	}
	to := filepath.Join(base, "out")
	if out, err := exec.Command("cmd", "/c", "mklink", "/J", to, target).CombinedOutput(); err != nil {
		t.Skipf("mklink /J is unavailable: %v %s", err, out)
	}
	t.Cleanup(func() { os.Remove(to) })
	_, err := Migrate(src, to, Entries(), Options{})
	if err == nil {
		t.Fatal("a junction at --to was accepted")
	}
	if !strings.Contains(err.Error(), "junction") {
		t.Errorf("the refusal does not say what it found: %v", err)
	}
	if es, err := os.ReadDir(target); err != nil || len(es) != 0 {
		t.Errorf("the junction's target was written through: %v %v", es, err)
	}
	if left := builds(t, to); len(left) != 0 {
		t.Errorf("a build was created: %v", left)
	}
}
