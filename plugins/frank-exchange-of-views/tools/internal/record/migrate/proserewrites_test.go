package migrate_test

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/goldentest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migrate"
)

// EVERY ARCHIVED ID REWRITTEN INSIDE A SEAT'S WORDING IS LISTED, WITH ITS CONTEXT. Migration
// rewrites an archived id wherever a seat's prose names it at word boundaries, and a subject word
// that spells the same — "Q1 revenue", "the G2 summit" — is rewritten as if it were a reference.
// Nothing at the write can tell the two apart, so the over-match is made visible instead: the
// manifest counts the rewrites per field, and this golden lists every one in the archive with the
// text around it, so a reviewer reads each and a change to the set is a diff.
func TestEveryProseRewriteInTheArchiveIsListed(t *testing.T) {
	tarballs := migrate.ArchivedRuns(t)
	sort.Strings(tarballs)
	var b strings.Builder
	for _, tb := range tarballs {
		name := strings.TrimSuffix(tb, ".tar.gz")
		m, _ := migrateArchive(t, tb)
		counted := map[string]int{}
		for _, r := range m.Rewritten {
			counted[r.Field]++
			fmt.Fprintf(&b, "%s\t%s\t%s -> %s\t%s\n", name, r.Field, r.Old, r.New, oneLine(r.Context))
		}
		// The manifest's count and the list are one fact: a count the list does not carry is a
		// rewrite no reviewer saw.
		if fmt.Sprint(counted) != fmt.Sprint(m.ProseRewrites) {
			t.Errorf("%s: the manifest counts %v prose rewrites per field and lists %v", name, m.ProseRewrites, counted)
		}
	}
	goldentest.Assert(t, "prose-rewrites", b.String())
}

func oneLine(s string) string { return strings.Join(strings.Fields(s), " ") }
