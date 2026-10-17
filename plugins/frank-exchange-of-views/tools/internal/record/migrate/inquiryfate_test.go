package migrate_test

import (
	"database/sql"
	"maps"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// archivedFate is the current word for each fate an archive may spell another way; every other
// fate arrives as it was recorded.
var archivedFate = map[string]string{"pursued": "underway", "concluded": "confirmed"}

// AN ARCHIVED FATE ARRIVES UNDER ITS CURRENT WORD, and as the seat recorded it. An archived
// `abandoned` may have been a refutation, and an archived open fate may be where an approach came to
// rest: the record holds the seat's word and nothing that tells them apart, so migration carries
// every inquiry act forward under the fate it was filed with — respelled where the vocabulary
// respelled it — and re-reads none of them. Two archives, because one holds no act of every fate:
// m20 has abandoned beside confirmed and declined, and b9 — a record older than the stated epoch —
// has underway, each archive spelling the open and the confirmed fate with an earlier word.
func TestAnArchivedInquiryFateMigratesAsRecorded(t *testing.T) {
	for _, c := range []struct {
		archive, holds, as string
		n                  int
	}{
		{"2026-10-11_is-91-prime-m20.tar.gz", "concluded", "confirmed", 7},
		{"2026-09-11_is-91-prime-b9.tar.gz", "pursued", "underway", 11},
	} {
		t.Run(c.as, func(t *testing.T) { archivedFatesMigrateAsRecorded(t, c.archive, c.holds, c.as, c.n) })
	}
}

// fatesIn counts the acts of one table by status, as a migration writes them: an act an archive's
// correction chain replaced is written once, as the chain's head, in its root's place.
func fatesIn(t *testing.T, db *sql.DB, table string) map[string]int {
	t.Helper()
	out := map[string]int{}
	for _, a := range writtenActs(t, db, table, "status") {
		out[a.text.String]++
	}
	return out
}

// writtenActs is one table's acts in record order, each column read as a migration writes the act:
// a row a correction replaced is passed over, and a corrected root carries its chain head's columns.
// A record that holds no correction table holds no chain.
func writtenActs(t *testing.T, db *sql.DB, table, text string) []inquiryAct {
	t.Helper()
	rows, err := db.Query(`SELECT e."key", x."` + text + `", x."reason", x."supersedes_status" FROM "` + table + `" x
	  JOIN "events" e ON e."id" = x."event_id" ORDER BY x."event_id"`)
	if err != nil {
		t.Fatal(err)
	}
	var keys []string
	byKey := map[string]inquiryAct{}
	for rows.Next() {
		var k sql.NullString
		var a inquiryAct
		if err := rows.Scan(&k, &a.text, &a.reason, &a.supersedes); err != nil {
			t.Fatal(err)
		}
		keys = append(keys, k.String)
		byKey[k.String] = a
	}
	rows.Close()
	next := map[string]string{}
	var has int
	if err := db.QueryRow(`SELECT count(*) FROM sqlite_master WHERE "name" = 'correction'`).Scan(&has); err != nil {
		t.Fatal(err)
	}
	if has > 0 {
		cr, err := db.Query(`SELECT "corrects", "replacement" FROM "correction"`)
		if err != nil {
			t.Fatal(err)
		}
		for cr.Next() {
			var c, r string
			if err := cr.Scan(&c, &r); err != nil {
				t.Fatal(err)
			}
			next[c] = r
		}
		cr.Close()
	}
	replaced := map[string]bool{}
	for _, r := range next {
		replaced[r] = true
	}
	var out []inquiryAct
	for _, k := range keys {
		if replaced[k] {
			continue
		}
		head := k
		for r, ok := next[head]; ok; r, ok = next[head] {
			head = r
		}
		out = append(out, byKey[head])
	}
	return out
}

func archivedFatesMigrateAsRecorded(t *testing.T, archive, holds, as string, n int) {
	src := recordtest.RequireArchive(t, archive)
	epoch, stated := statedEpoch(t, src)
	if stated && epoch >= record.EventSchema {
		t.Fatalf("%s states epoch %d: this test needs a record older than the binary's %d", archive, epoch, record.EventSchema)
	}
	archived := fatesIn(t, openRO(t, src), "avenue")
	if archived[holds] != n {
		t.Fatalf("%s holds %d %s act(s), want %d: %v", archive, archived[holds], holds, n, archived)
	}
	want := map[string]int{}
	for fate, k := range archived {
		if to, ok := archivedFate[fate]; ok {
			fate = to
		}
		want[fate] += k
	}

	res, dst := migrateArchive(t, archive)
	if len(res.Refusals) != 0 {
		t.Fatalf("refusals: %+v", res.Refusals)
	}
	if got, _ := statedEpoch(t, dst.Dir()); got != record.EventSchema {
		t.Errorf("the migrated record states epoch %d, this binary writes %d", got, record.EventSchema)
	}
	got := fatesIn(t, openRO(t, dst.Dir()), "inquiry")
	if !maps.Equal(got, want) {
		t.Errorf("inquiry acts by fate: the archive holds %v, so the migrated record should hold %v; it holds %v", archived, want, got)
	}
	if got[as] != n {
		t.Errorf("the %d archived %s act(s) arrived as %d %s", n, holds, got[as], as)
	}
	for old := range archivedFate {
		if got[old] != 0 {
			t.Errorf("the migrated record holds %d act(s) under the archived word %q", got[old], old)
		}
	}
	if got["refuted"] != 0 {
		t.Errorf("the migrated record holds %d refuted inquiry act(s): no archived seat filed that fate", got["refuted"])
	}
	// The fold reads the same fates the acts carry: an inquiry the archive left at the fate is at it.
	fam, err := record.FamilyOf(dst)
	if err != nil {
		t.Fatal(err)
	}
	held := 0
	for _, a := range record.InquiriesOf(fam.Events, fam.At) {
		if a.Status == "refuted" {
			t.Errorf("inquiry %s reads refuted after migration", a.ID)
		}
		if a.Status == as {
			held++
		}
	}
	if held == 0 {
		t.Errorf("no inquiry of the migrated record reads %s; the archive ships one under that fate", as)
	}
}

// inquiryAct is one row of an inquiry table, in record order.
type inquiryAct struct {
	text, reason, supersedes sql.NullString
}

func inquiryActs(t *testing.T, db *sql.DB, table, text string) []inquiryAct {
	t.Helper()
	rows, err := db.Query(`SELECT "` + text + `", "reason", "supersedes_status" FROM "` + table + `" ORDER BY "event_id"`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var out []inquiryAct
	for rows.Next() {
		var a inquiryAct
		if err := rows.Scan(&a.text, &a.reason, &a.supersedes); err != nil {
			t.Fatal(err)
		}
		out = append(out, a)
	}
	return out
}

// AN ARCHIVED PROPOSAL ARRIVES AS AN INQUIRY HOLDING A QUESTION AND NO REASON. The archive's table
// and its text column carry earlier names, and the archive's propose wrote the same text into its
// reason.
// Every act arrives, in order, under the current table — a corrected act once, as its chain's head; each proposal's question is its archived
// text byte for byte and carries no reason, and every later act's reason is the archive's, byte for
// byte.
func TestAnArchivedProposalArrivesAsAQuestionWithNoReason(t *testing.T) {
	const archive = "2026-09-11_is-91-prime-b9.tar.gz"
	src := recordtest.RequireArchive(t, archive)
	before := writtenActs(t, openRO(t, src), "avenue", "line")
	if len(before) != 25 {
		t.Fatalf("%s holds %d archived inquiry act(s) once its correction chain is written as its head, want 25", archive, len(before))
	}
	res, dst := migrateArchive(t, archive)
	if len(res.Refusals) != 0 {
		t.Fatalf("refusals: %+v", res.Refusals)
	}
	after := inquiryActs(t, openRO(t, dst.Dir()), "inquiry", "question")
	if len(after) != len(before) {
		t.Fatalf("%d archived act(s) arrived as %d inquiry act(s)", len(before), len(after))
	}
	proposals := 0
	for i, b := range before {
		a := after[i]
		if b.supersedes.String == "" {
			proposals++
			if b.reason != b.text {
				t.Fatalf("archived proposal %d's reason differs from its text, which this test assumes it repeats", i)
			}
			if !a.text.Valid || a.text.String != b.text.String {
				t.Errorf("proposal %d: question %q, archived text %q", i, a.text.String, b.text.String)
			}
			if a.reason.Valid {
				t.Errorf("proposal %d arrived holding a reason (%q): a proposal holds a question and no reason", i, a.reason.String)
			}
			continue
		}
		if want := respellIDs(b.reason.String, res.IDs); a.reason.Valid != b.reason.Valid || a.reason.String != want {
			t.Errorf("act %d: reason %q, archived %q — only a proposal's copy of its question is dropped", i, a.reason.String, want)
		}
	}
	if proposals == 0 {
		t.Fatal("no archived proposal was compared")
	}
}

// AN ARCHIVED REVIEW ARRIVES AS AN INQUIRY REVIEW, under either table an archive wrote it to: m20
// spelled the table with an earlier word, b9 with the current one.
func TestAnArchivedReviewArrivesAsAnInquiryReview(t *testing.T) {
	for _, c := range []struct {
		archive, table string
		n              int
	}{
		{"2026-10-11_is-91-prime-m20.tar.gz", "avenue_review", 5},
		{"2026-09-11_is-91-prime-b9.tar.gz", "inquiry_review", 6},
	} {
		t.Run(c.table, func(t *testing.T) {
			src := recordtest.RequireArchive(t, c.archive)
			var archived int
			if err := openRO(t, src).QueryRow(`SELECT count(*) FROM "` + c.table + `"`).Scan(&archived); err != nil {
				t.Fatal(err)
			}
			if archived != c.n {
				t.Fatalf("%s holds %d %s row(s), want %d", c.archive, archived, c.table, c.n)
			}
			res, dst := migrateArchive(t, c.archive)
			if len(res.Refusals) != 0 {
				t.Fatalf("refusals: %+v", res.Refusals)
			}
			fam, err := record.FamilyOf(dst)
			if err != nil {
				t.Fatal(err)
			}
			got := 0
			for _, e := range fam.Events {
				if _, ok := recordpb.BodyAs[*recordpb.InquiryReview](e); ok {
					got++
				}
			}
			if got != c.n {
				t.Errorf("%d archived review(s) arrived as %d inquiry review(s)", c.n, got)
			}
		})
	}
}

// AN ARCHIVED INQUIRY RULING ARRIVES UNDER THE SUBJECT `inquiry`, with the ruling in the `inquiry`
// arm, whichever words the archive used: quadratic-formula and m20 spell the subject and the
// ruling's column with two different earlier words, and the shard era's run2 ruled on subject
// `inquiry` with a flat ruling word.
func TestAnArchivedInquiryRulingArrivesUnderSubjectInquiry(t *testing.T) {
	for _, c := range []struct {
		archive string
		n       int
	}{
		{"2026-09-02_quadratic-formula.tar.gz", 5},
		{"2026-10-11_is-91-prime-m20.tar.gz", 10},
		{"2026-08-22_is-7-prime-run2.tar.gz", 4},
	} {
		t.Run(c.archive, func(t *testing.T) {
			res, dst := migrateArchive(t, c.archive)
			if len(res.Refusals) != 0 {
				t.Fatalf("refusals: %+v", res.Refusals)
			}
			fam, err := record.FamilyOf(dst)
			if err != nil {
				t.Fatal(err)
			}
			got := 0
			for _, e := range fam.Events {
				r, ok := recordpb.BodyAs[*recordpb.MotionRule](e)
				if !ok || r.Subject() != recordpb.MotionSubject_MOTION_SUBJECT_INQUIRY {
					continue
				}
				if r.GetInquiry() == recordpb.InquiryRuling_INQUIRY_RULING_UNSPECIFIED {
					t.Errorf("a ruling under subject inquiry arrived with no word in the inquiry arm: %v", r.GetRuling())
					continue
				}
				got++
			}
			if got != c.n {
				t.Errorf("%d inquiry ruling(s) arrived, want %d", got, c.n)
			}
		})
	}
}

// AN ARCHIVED ANCHOR KIND ARRIVES AS `inquiry`. b spelled a finding about an inquiry with
// the current word, wave-b with an earlier one, on findings and mints alike.
func TestAnArchivedAboutKindArrivesAsInquiry(t *testing.T) {
	for _, c := range []struct {
		archive         string
		findings, mints int
	}{
		{"2026-09-09_is-91-prime-b.tar.gz", 2, 0},
		{"2026-10-12_is-91-prime-wave-b.tar.gz", 7, 3},
	} {
		t.Run(c.archive, func(t *testing.T) {
			res, dst := migrateArchive(t, c.archive)
			if len(res.Refusals) != 0 {
				t.Fatalf("refusals: %+v", res.Refusals)
			}
			fam, err := record.FamilyOf(dst)
			if err != nil {
				t.Fatal(err)
			}
			findings, mints := 0, 0
			for _, e := range fam.Events {
				if f, ok := recordpb.BodyAs[*recordpb.Finding](e); ok && f.GetAboutKind() == recordpb.AboutKind_ABOUT_KIND_INQUIRY {
					findings++
				}
				if m, ok := recordpb.BodyAs[*recordpb.Mint](e); ok && m.GetAboutKind() == recordpb.AboutKind_ABOUT_KIND_INQUIRY {
					mints++
				}
			}
			if findings != c.findings || mints != c.mints {
				t.Errorf("about an inquiry: %d finding(s) and %d mint(s), want %d and %d", findings, mints, c.findings, c.mints)
			}
		})
	}
}

// respellIDs is an archived seat's wording with each archived id at word boundaries spelled as the
// migrated record spells it — the one change migration makes to a seat's argument.
func respellIDs(s string, ids map[string]string) string {
	old := make([]string, 0, len(ids))
	for k := range ids {
		old = append(old, regexp.QuoteMeta(k))
	}
	if len(old) == 0 {
		return s
	}
	sort.Slice(old, func(i, j int) bool { return len(old[i]) > len(old[j]) })
	re := regexp.MustCompile(`\b(` + strings.Join(old, "|") + `)\b`)
	return re.ReplaceAllStringFunc(s, func(id string) string { return ids[id] })
}
