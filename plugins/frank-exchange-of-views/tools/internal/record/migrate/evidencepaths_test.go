package migrate_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/report"
)

// THE PROOF LAYER SHIPS THE RUN'S OWN COPY OF EACH SCRIPT, NEVER THE PATH A SEAT RAN IT FROM.
//
// A proof's script field is the path blue handed `prove`, and on a real run that is a scratch file
// in the operator's home directory: the archived m19 record holds eleven of them, and universe-m24
// and m25 shipped nine and eleven in evidence.md (m24 dark-side ev548) beside a lead saying every
// script is cached under proofs/<sha256>/. The record holds the sha, and the store keeps the script
// under it; the shipped set names that copy, which a reader of the run directory can open.
//
// The re-run is red's, and the document the report's proof notes point at says so in the party's
// name — as report.md attributes nothing to a seat — rather than in a lens's seat id and sitting
// (m24 dark-side ev547).
func TestTheShippedSetNamesNoScriptPathAndNoSeatInItsProofs(t *testing.T) {
	_, dst := migrateArchive(t, "2026-10-10_is-91-prime-m19.tar.gz")
	proofs, err := record.RecordedProofs(dst)
	if err != nil {
		t.Fatal(err)
	}
	var homed int
	for _, p := range proofs {
		if strings.HasPrefix(p.Script, "/home/") {
			homed++
		}
	}
	if homed == 0 {
		t.Fatal("the archived record holds no proof run from a home-directory path, so this measures nothing")
	}
	docs, err := shippedDocs(dst)
	if err != nil {
		t.Fatal(err)
	}
	seatID := regexp.MustCompile(`\b(red-lens-[a-z-]+|red-chair|blue-[a-z0-9-]+|judge)\b`)
	var evidence string
	for _, d := range docs {
		for _, p := range proofs {
			if strings.Contains(d.Body, p.Script) {
				t.Errorf("%s carries the path proof %s was run from: %s", d.File, p.Label, p.Script)
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
	for _, p := range proofs {
		want := "`proofs/" + p.SHA + "/script.py`"
		if !strings.Contains(evidence, want) {
			t.Errorf("evidence.md does not name proof %s's stored copy %s", p.Label, want)
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
