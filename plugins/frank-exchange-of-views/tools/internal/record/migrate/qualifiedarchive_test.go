package migrate_test

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// AN ARCHIVED OUTCOME IS CARRIED AS HISTORY, AND QUALIFIED TRANSLATES NOTHING (ruled by gblock
// 2026-11-12, F3 and F10). QUALIFIED is a new outcome, not a new name for an old one: no archived run
// split its question, so none can say a part went unanswered, and migrate has no translation to make.
// An archived VERIFIED migrates as VERIFIED, and the record derives VERIFIED over it — a PASS over no
// part derives VERIFIED — while the live gate holds every new run's PASS until its question is split.
// The residue is stated: an archived record reads `question_unsplit` on its board, and its stamp stands.
func TestAnArchivedVerifiedOutcomeIsCarriedAsVerified(t *testing.T) {
	verified := 0
	for _, name := range []string{"2026-10-10_is-91-prime-m19.tar.gz", "2026-10-19_is-91-prime-m30.tar.gz", "2026-10-24_is-91-prime-1565.tar.gz"} {
		t.Run(name, func(t *testing.T) {
			src := sourceOutcomes(t, name)
			if len(src) == 0 {
				t.Skipf("%s holds no outcome", name)
			}
			want, _ := src[len(src)-1].Fields["verdict"].(string)
			_, run := migrateArchive(t, name)
			got, err := record.RecordedOutcome(run)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.EqualFold(got, want) {
				t.Fatalf("the source stamps %q and the migrated record %q — an archived outcome is carried as written", want, got)
			}
			parts, err := record.QuestionParts(run)
			if err != nil {
				t.Fatal(err)
			}
			if len(parts) != 0 {
				t.Errorf("the migrated record holds %d question part(s), and no archive split its question", len(parts))
			}
			if !strings.EqualFold(got, "VERIFIED") {
				return
			}
			verified++
			derived, why, ok, err := record.DeriveVerdict(run)
			if err != nil || !ok || derived != "VERIFIED" {
				t.Errorf("the record derives %q (ok=%v, err=%v) over an archived VERIFIED: %s", derived, ok, err, why)
			}
		})
	}
	if verified == 0 {
		t.Error("no archive named here stamps VERIFIED, so the carried VERIFIED is not measured — name one that does")
	}
}
