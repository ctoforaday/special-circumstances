package migrate_test

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/goldentest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migrate"
)

// archivesAnswer is the questions the archived runs hold an answer to between them. The rest no
// archived run answers — none holds a failed sitting, a stop that lands after its seat's next
// opening, a delivery in no sitting, an upheld position, a ruled docket motion or a stamped
// spot-check change, and a lens whose pin was never moved has its work list asked of no archive —
// and package record's seeded records hold each (TestEverySeededSittingAndDeliveryArmAnswersAsItsGoldenSays).
var archivesAnswer = []string{qNeverClosed, qStillOpen, qReadyLens, qStaleArea, qLastReadPin,
	"last_read unchanged", "last_read served", qWholeOwed, qBlockers, qClosingDuty, qClosingMissing,
	qDelivery, qWholeDelivery, qChangesDelivery, qUnsatDelivery, qExchange}

// EVERY ARCHIVED RUN ANSWERS ITS SITTING AND DELIVERY QUESTIONS AS ITS GOLDEN SAYS. The folds behind
// the plan, the PASS gate, the work lists, the closing duties and the delivery lines read which
// sittings closed, which were sat through, and what each delivery handed its seat; this pins what
// each answers on every committed archive, migrated (sittingDeliveryAnswers), so a fold that reads
// those facts another way answers the same or fails here.
//
// A QUESTION NO ARCHIVE ANSWERS PINS NOTHING, so the totals are asserted and logged: each question in
// archivesAnswer is answered at least once over the sweep, each other question by no archive — one
// an archive comes to answer joins the list — and no call is refused.
func TestEveryArchivedRunAnswersItsSittingAndDeliveryQuestionsAsItsGoldenSays(t *testing.T) {
	names := migrate.ArchivedRuns(t)
	if len(names) == 0 {
		t.Fatal("no archived run found")
	}
	totals := map[string]int{}
	var out strings.Builder
	var obtaining, rendering time.Duration
	for _, name := range names {
		began := time.Now()
		_, run := migrateArchive(t, name)
		obtaining += time.Since(began)
		began = time.Now()
		text, err := sittingDeliveryAnswers(run, totals, false)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		rendering += time.Since(began)
		fmt.Fprintf(&out, "== %s ==\n%s\n", strings.TrimSuffix(name, ".tar.gz"), text)
	}
	t.Logf("%d archives: %s obtaining them migrated, %s rendering", len(names), obtaining.Round(time.Millisecond), rendering.Round(time.Millisecond))
	for _, q := range sittingDeliveryQuestions {
		t.Logf("%-50s %d", q, totals[q])
		switch answered := slices.Contains(archivesAnswer, q); {
		case q == qRefused && totals[q] != 0:
			t.Errorf("%d call(s) were refused on the archived runs: each is a line of the golden that says so", totals[q])
		case answered && totals[q] == 0:
			t.Errorf("no archived run answers %q: the sweep pins nothing about it", q)
		case !answered && q != qRefused && totals[q] != 0:
			t.Errorf("an archived run answers %q, %d time(s): add it to archivesAnswer", q, totals[q])
		}
	}
	goldentest.Assert(t, "archived_sitting_delivery_answers", out.String())
}

// THE RENDERER IS ONE TEXT IN BOTH PACKAGES THAT HOLD IT: this package's copy and package record's
// are the same bytes after the package clause, so the archived runs and the seeded records are asked
// the same questions.
func TestTheSittingAndDeliveryRendererIsOneText(t *testing.T) {
	body := func(path string) string {
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		_, after, ok := strings.Cut(string(b), "\n")
		if !ok || !strings.HasPrefix(string(b), "package ") {
			t.Fatalf("%s does not open with its package clause", path)
		}
		return after
	}
	if body("sittingdeliveryrender_test.go") != body(filepath.Join("..", "sittingdeliveryrender_test.go")) {
		t.Error("sittingdeliveryrender_test.go differs between internal/record and internal/record/migrate: copy the one that changed over the other, keeping each package clause")
	}
}
