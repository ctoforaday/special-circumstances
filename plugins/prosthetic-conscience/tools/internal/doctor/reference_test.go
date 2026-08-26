package doctor

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// writeInstallRecord builds the client's own layout: a plugins dir holding
// installed_plugins.json, with the plugin copied under cache/<marketplace>/<plugin>/<version>.
func writeInstallRecord(t *testing.T, sha string) (pluginsDir, root string) {
	t.Helper()
	pluginsDir = t.TempDir()
	root = filepath.Join(pluginsDir, "cache", "special-circumstances", "prosthetic-conscience", "0.41.0")
	if err := os.MkdirAll(filepath.Join(root, "bin"), 0o755); err != nil {
		t.Fatal(err)
	}
	rec := map[string]any{"plugins": map[string]any{
		"prosthetic-conscience@special-circumstances": []any{
			map[string]any{"installPath": root, "gitCommitSha": sha, "version": "0.41.0"},
		},
	}}
	b, err := json.Marshal(rec)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pluginsDir, "installed_plugins.json"), b, 0o644); err != nil {
		t.Fatal(err)
	}
	return pluginsDir, root
}

// AN INSTALLED PLUGIN HAS A REFERENCE, and until this it did not.
//
// The staleness arm ran `git rev-parse HEAD` against the install directory, which is a copied tree
// with no .git, so the reference was "" for every binary in every real installation — the one place
// this check runs in production. Measured on a live machine: 17 of 17 binaries reported as carrying
// no build stamp while every one of them carried a correct vcs.revision.
func TestAnInstalledPluginTakesItsReferenceFromTheInstallRecord(t *testing.T) {
	const sha = "2bb1d7b817c5611032652e83a9e254fc46b57aab"
	_, root := writeInstallRecord(t, sha)

	if got := ReferenceCommit(binStatus{Name: "sc-stop", Root: root, Built: true}); got != sha {
		t.Errorf("ReferenceCommit = %q, want the recorded install commit %q", got, sha)
	}
}

// A CHECKOUT STILL WINS. Development is the case where HEAD is the truth and the install record is
// either absent or describes a different copy entirely.
func TestACheckoutPrefersItsOwnHead(t *testing.T) {
	dir := t.TempDir()
	if out, err := runGit(t, dir, "init"); err != nil {
		t.Skipf("no git here: %v\n%s", err, out)
	}
	runGit(t, dir, "config", "user.email", "t@example.com")
	runGit(t, dir, "config", "user.name", "t")
	if err := os.WriteFile(filepath.Join(dir, "f"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, dir, "add", "-A")
	if out, err := runGit(t, dir, "commit", "-m", "c"); err != nil {
		t.Skipf("cannot commit here: %v\n%s", err, out)
	}
	head, _ := runGit(t, dir, "rev-parse", "HEAD")

	if got := ReferenceCommit(binStatus{Name: "sc-stop", Root: dir, Built: true}); got != head {
		t.Errorf("ReferenceCommit = %q, want the checkout's HEAD %q", got, head)
	}
}

// NO RECORD MEANS NO ANSWER — never a guess, and never a value that would make an old binary look
// current. Compare turns "" into NoReference, which is reported apart from Unstamped.
func TestAnUnknownInstallYieldsNoReferenceRatherThanAGuess(t *testing.T) {
	root := filepath.Join(t.TempDir(), "cache", "m", "p", "1.0.0")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	if got := ReferenceCommit(binStatus{Name: "x", Root: root, Built: true}); got != "" {
		t.Errorf("ReferenceCommit = %q, want empty for an install nothing recorded", got)
	}
	if got := Compare(BuildStamp{Revision: "abc"}, ""); got != NoReference {
		t.Errorf("Compare with no reference = %q, want %q", got, NoReference)
	}
}

// A record naming a DIFFERENT install must not answer for this one. installPath is the match key
// precisely so a second copy at another version cannot lend its commit to this one.
func TestARecordForAnotherInstallDoesNotAnswer(t *testing.T) {
	pluginsDir, root := writeInstallRecord(t, "2bb1d7b817c5611032652e83a9e254fc46b57aab")
	other := filepath.Join(pluginsDir, "cache", "special-circumstances", "prosthetic-conscience", "0.32.0")
	if err := os.MkdirAll(other, 0o755); err != nil {
		t.Fatal(err)
	}
	_ = root
	if got := ReferenceCommit(binStatus{Name: "sc-stop", Root: other, Built: true}); got != "" {
		t.Errorf("ReferenceCommit = %q for an install the record does not describe; want empty", got)
	}
}

func runGit(t *testing.T, dir string, args ...string) (string, error) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	b, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(b)), err
}
