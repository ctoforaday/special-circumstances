package migrate_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/consistency"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migrate"
)

// THE ARCHIVED L6 RESIDUE IS m37's ONE COMPUTATION. No live write leaves a computation in the report
// citing a citation an act took out of it, but replay asks no write rule, so an archived act migrates
// as it was made: m37's anchor-only retire of C-849a1f53 while K-93f0a1a7 still cites it is carried.
// consistency.Check's L6 fold, run over every migrated committed archive, names exactly that one; its
// other kinds have never been measured on an archive and are not asserted.
func TestConsistencyNamesTheArchivedL6Residue(t *testing.T) {
	var got []string
	for _, name := range migrate.ArchivedRuns(t) {
		_, run := migrateArchive(t, name)
		res, err := consistency.Check(run)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		for _, v := range res.Violations {
			if v.Rule != "l6" {
				continue
			}
			got = append(got, fmt.Sprintf("%s %s", strings.TrimSuffix(name, ".tar.gz"), v.Anchor))
			if !strings.Contains(v.Detail, "cites C-849a1f53,") {
				t.Errorf("%s: %s's violation names another citation: %s", name, v.Anchor, v.Detail)
			}
		}
	}
	if want := []string{"bundle-store/2026-10-31_is-91-prime-m37 K-93f0a1a7"}; fmt.Sprint(got) != fmt.Sprint(want) {
		t.Errorf("the archived L6 violations are %v, want %v", got, want)
	}
}
