package record

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// THE RULING SEATS' LISTS SAY WHETHER THIS RUN HAS LAW, so no seat probes the directory for it. The
// bench and the chair are told the corpus and where it is when setup staged one, and that there is
// none when it did not — an absent directory and an empty one alike. A lens's list carries no law
// line: no lens prompt sends it to the law.
func TestTheRulingSeatsAreToldWhetherTheRunHasLaw(t *testing.T) {
	for _, tc := range []struct {
		name  string
		stage func(t *testing.T, run Run)
		want  string
	}{
		{"no law directory", func(t *testing.T, run Run) {}, "no law is staged for this run — setup found no law corpus"},
		{"an empty law directory", func(t *testing.T, run Run) {
			if err := os.MkdirAll(LawDir(run), 0o755); err != nil {
				t.Fatal(err)
			}
		}, "no law is staged for this run — " + "{LAW}" + " holds no file"},
		{"two law files", func(t *testing.T, run Run) {
			if err := os.MkdirAll(LawDir(run), 0o755); err != nil {
				t.Fatal(err)
			}
			for _, f := range []string{"statute.md", "precedents.md"} {
				if err := os.WriteFile(filepath.Join(LawDir(run), f), []byte("law\n"), 0o644); err != nil {
					t.Fatal(err)
				}
			}
		}, "this run's law is staged: 2 file(s) in {LAW} — read them before you rule"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			run := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().register("red-chair").seed()
			tc.stage(t, run)
			want := strings.ReplaceAll(tc.want, "{LAW}", LawDir(run))
			for _, seat := range []struct{ role, id string }{{"bench", "judge"}, {"chair", "red-chair"}} {
				if !workMentions(t, run, seat.role, seat.id, want) {
					t.Errorf("%s's list does not say %q", seat.id, want)
				}
			}
			if workMentions(t, run, "lens", evLens, "law is staged") {
				t.Errorf("a lens's list carries a law line")
			}
		})
	}
}
