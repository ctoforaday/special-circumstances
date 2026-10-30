package migrate_test

import (
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migrate"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/report"
)

// THE SHIPPED SET NAMES EACH COMPUTATION BY ITS STORED COMMAND, NEVER THE PATH A SEAT RAN IT FROM.
//
// An archived proof's script field is the path blue handed the tool, and on a real run that is a
// scratch file in the operator's home directory: the archived m19 record holds eleven of them, and
// universe-m24 and m25 shipped nine and eleven in evidence.md (m24 dark-side ev548). The record
// holds the bundle and the command that names the stored file; the author's path is not carried,
// and nothing a reader is shipped names it.
//
// The re-run is red's, and the document the report's computation notes point at says so in the
// party's name — as report.md attributes nothing to a seat — rather than in a lens's seat id and
// sitting (m24 dark-side ev547).
func TestTheShippedSetNamesNoScriptPathAndNoSeatInItsComputations(t *testing.T) {
	const name = "2026-10-10_is-91-prime-m19.tar.gz"
	_, dst := migrateArchive(t, name)
	computations, err := record.RecordedComputations(dst)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range computations {
		if strings.Contains(strings.Join(p.Argv, " "), "/home/") {
			t.Errorf("computation %s's command names a home-directory path: %q", p.Label, p.Argv)
		}
	}
	homed := archivedScriptPaths(t, name)
	if len(homed) == 0 {
		t.Fatal("the archived record holds no proof run from a home-directory path, so this measures nothing")
	}
	docs, err := shippedDocs(dst)
	if err != nil {
		t.Fatal(err)
	}
	seatID := regexp.MustCompile(`\b(red-lens-[a-z-]+|red-chair|blue-[a-z0-9-]+|judge)\b`)
	var evidence string
	for _, d := range docs {
		for _, path := range homed {
			if strings.Contains(d.Body, path) {
				t.Errorf("%s carries the path an archived proof was run from: %s", d.File, path)
				break
			}
		}
		if d.File == report.FileEvidence {
			evidence = d.Body
		}
	}
	if evidence == "" {
		t.Fatal("the shipped set holds no evidence document")
	}
	for _, p := range computations {
		want := report.Literal(strings.Join(p.Argv, " ")) + " (bundle sha256 `" + p.SHA[:12] + "`)"
		if !strings.Contains(evidence, want) {
			t.Errorf("evidence.md does not name computation %s's stored command %s", p.Label, want)
		}
	}
	for _, d := range docs {
		if d.File != report.FileEvidence && d.File != report.FileReport {
			continue
		}
		// A script and its output ship as the exact bytes that ran, in fences, and m19's reads two
		// cache files by absolute path; every line outside a fence is the tool's or the seat's prose.
		fenced := false
		for _, line := range strings.Split(d.Body, "\n") {
			if strings.HasPrefix(line, "```") {
				fenced = !fenced
				continue
			}
			if fenced {
				continue
			}
			if strings.Contains(line, "/home/") {
				t.Errorf("%s carries a home-directory path: %s", d.File, line)
			}
			// The re-run's note is the seat's text; what is checked is the attribution the tool writes.
			if strings.HasPrefix(line, "- **audited by") || strings.HasPrefix(line, "- ~~audited by") {
				if head, _, _ := strings.Cut(line, ":"); seatID.MatchString(head) {
					t.Errorf("%s attributes a re-run to a seat: %s", d.File, head)
				}
			}
		}
	}
}

// archivedScriptPaths is every home-directory script path the archive's proofs were run from.
func archivedScriptPaths(t *testing.T, name string) []string {
	t.Helper()
	sq, err := migrate.OpenSQLite(filepath.Join(recordtest.ExtractArchive(t, name), "records"), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer sq.Close()
	evs, err := sq.Events()
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, e := range evs {
		if s, _ := e.Fields["script"].(string); e.Word == "proof" && strings.HasPrefix(s, "/home/") {
			out = append(out, s)
		}
	}
	return out
}
