package setup_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/setup"
)

// SETUP MUST NOT STUB A FILE THE TOOL RENDERS ON READ.
//
// This is an external test package because it reads the view list from internal/cli, which
// imports setup. The list is DERIVED — cli.ViewNames() walks the real command tree — so a view
// added tomorrow is covered without anyone remembering to extend a list here. A literal list of
// forbidden filenames is the thing that went stale in the first place.
//
// WHAT IT COSTS WHEN THIS BREAKS, measured rather than imagined. `debate.md` and
// `red/citation-ledger.md` became projections and setup went on writing a one-line husk for each;
// every run since carried two files that nothing filled and nothing read. `red/ledger.md` and
// `red/archive.md` went further: they stopped being stubbed AND stopped being written, and the
// capture audit that read them returned SKIP — "no ledger/archive (pre-sharding run)" — on every
// modern run for months. That reading is benign and available, which is exactly why nobody caught
// it: an audit whose input has vanished looks like an audit declining to judge an old run.
//
// An absent file is honest. A husk is a promise that something will fill it.
func TestSetupStubsNoFileTheToolRenders(t *testing.T) {
	runDir := t.TempDir()
	setup.BuildSkeleton(runOf(t, runDir))

	rendered := map[string]bool{}
	for _, v := range cli.ViewNames() {
		rendered[strings.ToLower(v)+".md"] = true
	}
	if len(rendered) == 0 {
		t.Fatal("cli.ViewNames() is empty, so this check would pass by comparing against nothing — " +
			"the view list moved and took the guard with it")
	}

	var offenders []string
	err := filepath.WalkDir(runDir, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		// Case-sensitivity mattered while blue/CHANGELOG.md existed: it shared a name with the
		// retired `changelog` VIEW and only one of the two had a writer. Both are gone now — the
		// file with #251, the view with the board collapse — and the rule is kept because the
		// next artifact to share a name with a projection will need it.
		rel, _ := filepath.Rel(runDir, p)
		rel = filepath.ToSlash(rel)
		if base := filepath.Base(p); rendered[base] && base == strings.ToLower(base) {
			offenders = append(offenders, rel)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(offenders) > 0 {
		t.Fatalf("setup stubs %v, which the tool renders on read (`show --view <name>`). Nothing will "+
			"ever fill them, so every run carries a husk that reads as an empty artifact rather than an "+
			"absent one — and any audit that reads one reports the empty case forever", offenders)
	}
}
