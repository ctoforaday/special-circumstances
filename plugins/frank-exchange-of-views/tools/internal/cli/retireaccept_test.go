package cli

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// T-B49 (a retire, per gap kind): THE AUTHOR TAKES THE RAISER'S STATED RETIRE. On a gap about a
// citation and on a gap about a computation, docketed at impasse, the raiser's closing states that
// anchor's retire; blue's `retire --accept G` writes the retire of exactly that anchor, answers = G and
// accepted, keeps the sentence, ends the docketing unruled, and places no estoppel row — the estoppel
// is a pair's alone. Before any retire is stated, the accept is refused naming the gap.
func TestTheAuthorTakesTheRaisersStatedRetire(t *testing.T) {
	for _, kind := range []string{"citation", "computation"} {
		t.Run(kind, func(t *testing.T) {
			runDir := acceptRun(t)
			var id string
			if kind == "citation" {
				if out, err := run(t, "cite", "--run", runDir, "--seat-id", blueSeat, "--quote", acceptS+".", "--url", "https://costs/1", "--title", "Source"); err != nil {
					t.Fatalf("blue's cite: %v %s", err, out)
				}
				id = lastBody(t, runDir, &recordpb.Cite{}).GetLabel()
			} else {
				id = computeSentence(t, runDir, acceptS+".")
			}
			out, err := runMint(t, "--run", runDir, "--seat-id", lensSeat, "--key", "G", "--about-kind", "anchor", "--about", id,
				"--class", "spec_underspecification", "--problem", "the defect", "--fix", "the fix", "--check-kind", "document",
				"--check", "the acceptance check", "--severity", "medium", "--likelihood", "medium", "--impact", "medium", "--complexity", "low")
			gap := gapID(out)
			if err != nil || gap == "" {
				t.Fatalf("the lens's gap about %s: %v %s", id, err, out)
			}
			d := docketNow(t, runDir, gap, recordpb.DocketingCause_DOCKETING_CAUSE_IMPASSE, "")
			accept := func() (string, error) {
				return runAt(t, "retire", "--run", runDir, "--seat-id", blueSeat, "--accept", gap, "--reason", "red is right", "--json")
			}
			out, err = accept()
			heldRefusal(t, "retire --accept before any retire is stated", out, err, "states no retire")
			if out, err := redStates(t, runDir, gap, "--retire"); err != nil {
				t.Fatalf("the raiser's closing stating the retire: %v %s", err, out)
			}
			estoppels := estoppelRows(t, runDir)
			if out, err := accept(); err != nil {
				t.Fatalf("retire --accept %s: %v %s", gap, err, out)
			}
			r := lastBody(t, runDir, &recordpb.Retire{})
			if len(r.GetAnchors()) != 1 || r.GetAnchors()[0] != id || r.GetAnswers() != handleText(t, runDir, gap) || !r.GetAccepted() {
				t.Errorf("the retire = %+v, want %s's retire answering %s, accepted", r, id, gap)
			}
			if !strings.Contains(readReport(t, runDir), acceptS) {
				t.Errorf("the retire took its sentence: %s", readReport(t, runDir))
			}
			if got := estoppelRows(t, runDir); got != estoppels {
				t.Errorf("estoppel rows %d after the retire accept, want %d: only a pair places one", got, estoppels)
			}
			_, ds, _ := partyRecord(t, runDir)
			if k := ds[d]; k == nil || k.Standing() || k.Ruled() {
				t.Errorf("docketing %s reads %+v after the author's accept; want ended, unruled", d, k)
			}
		})
	}
}

// estoppelRows is how many estoppel tombstones the record holds.
func estoppelRows(t *testing.T, runDir string) int {
	t.Helper()
	var n int
	if err := openRunDB(t, runDir).QueryRow(`SELECT count(*) FROM "tombstone" WHERE "kind" = ?`, recordpb.Word(recordpb.TombstoneKind_TOMBSTONE_KIND_ESTOPPEL)).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}
