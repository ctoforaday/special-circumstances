package migrate_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/goldentest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migrate"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// AN ARCHIVED PROPOSAL'S METHOD MOVES TO ITS FATES. Before epoch 30 the method was written on the
// proposal and the inquiry's row printed it whatever became of the inquiry; from epoch 30 it is a
// fate's account and a proposal holds none. So every inquiry the archive gave a fate holds, migrated,
// the method the archive's row printed — the latest fate's, carried from the proposal — and no
// migrated proposal holds one.
//
// THE RESIDUE IS STATED, NOT HIDDEN: an inquiry the archive left at its proposal has no fate to carry
// the method to, and its row prints none. The golden line per archive counts both: `carried` is the
// inquiries whose method arrived on their fates, `dropped` the proposal-only ones whose method went.
func TestTheProposalsMethodMovesToItsFates(t *testing.T) {
	dir := filepath.Dir(archivePath(t, "2026-09-11_is-91-prime-b9.tar.gz"))
	tarballs, err := filepath.Glob(filepath.Join(dir, "*.tar.gz"))
	if err != nil || len(tarballs) == 0 {
		t.Fatalf("no archived run found under %s: %v", dir, err)
	}
	var lines []string
	carriedAll := 0
	for _, tb := range tarballs {
		name := filepath.Base(tb)
		src := recordtest.RequireArchive(t, name)
		printed, fated := archivedInquiryMethods(t, src)

		man, dst := migrateArchive(t, name)
		if len(man.Refusals) != 0 {
			t.Fatalf("%s: %d event(s) refused migrating; the first: %+v", name, len(man.Refusals), man.Refusals[0])
		}
		fam, err := record.FamilyOf(dst)
		if err != nil {
			t.Fatal(err)
		}
		for _, e := range fam.Events {
			if q, ok := recordpb.BodyAs[*recordpb.Inquiry](e); ok && q.SupersedesStatus == nil && q.GetMethod() != "" {
				t.Errorf("%s: the migrated proposal of %s holds the method %q; a proposal holds none", name, q.GetInquiryId(), q.GetMethod())
			}
		}
		migratedID := func(old string) string {
			if to, ok := man.IDs[old]; ok {
				return to
			}
			return old
		}
		got := map[string]string{}
		for _, a := range record.InquiriesOf(fam.Events, fam.At) {
			got[a.ID] = a.Method
		}
		carried, dropped := 0, 0
		for old, method := range printed {
			want := ""
			if fated[old] {
				want, carried = method, carried+1
			} else {
				dropped++
			}
			if g := got[migratedID(old)]; g != want {
				t.Errorf("%s: inquiry %s (archived %s) holds the method %q after migration; the archive's row printed %q and it has a fate: %v — want %q",
					name, migratedID(old), old, g, method, fated[old], want)
			}
		}
		carriedAll += carried
		lines = append(lines, fmt.Sprintf("%s carried=%d dropped=%d", strings.TrimSuffix(name, ".tar.gz"), carried, dropped))
	}
	// AN ARCHIVE THAT EXERCISES THE TRANSLATION IS WHAT MAKES THIS A TEST: with no carried method
	// anywhere, every comparison above is "" against "".
	if carriedAll == 0 {
		t.Fatal("no archived inquiry carried a method onto a fate — the translation is measured on nothing")
	}
	goldentest.Assert(t, "archived_inquiry_methods", strings.Join(lines, "\n")+"\n")
}

// archivedInquiryMethods reads the archive's own inquiry acts, translated by the registry as replay
// translates them and before replay's carry: the method each inquiry's row printed under the fold
// the archive was written for (the last method any of its acts carried — the proposal's, for every
// archive older than epoch 30), and which inquiries hold a fate beyond the proposal.
func archivedInquiryMethods(t *testing.T, runDir string) (printed map[string]string, fated map[string]bool) {
	t.Helper()
	recs := filepath.Join(runDir, "records")
	var src migrate.Source
	if _, err := os.Stat(filepath.Join(recs, "record.db")); err == nil {
		sq, err := migrate.OpenSQLite(recs, t.TempDir())
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = sq.Close() })
		src = sq
	} else {
		js, err := migrate.OpenJSONL(recs)
		if err != nil {
			t.Fatal(err)
		}
		src = js
	}
	evs, err := src.Events()
	if err != nil {
		t.Fatal(err)
	}
	reg := migrate.Entries()
	printed, fated = map[string]string{}, map[string]bool{}
	for _, old := range evs {
		bodies, err := reg.Translate(old)
		if err != nil {
			continue
		}
		for _, b := range bodies {
			q, ok := b.(*recordpb.Inquiry)
			if !ok || q.GetInquiryId() == "" {
				continue
			}
			if m := q.GetMethod(); m != "" {
				printed[q.GetInquiryId()] = m
			}
			if q.SupersedesStatus != nil {
				fated[q.GetInquiryId()] = true
			}
		}
	}
	return printed, fated
}

// THE RESIDUE, AS DATA. No archived run leaves an inquiry with a method at its proposal, so the
// residue the golden counts as `dropped` is stated here on a fixture: Q1's proposal method arrives on
// both its fates, Q2's own fate keeps the method it states, and Q3, never given a fate, holds none —
// the method a proposal carried is not invented onto a fate that does not exist.
func TestAProposalOnlyInquiryLosesItsMethod(t *testing.T) {
	old := newOldRecord(t)
	if _, err := old.db.Exec(`CREATE TABLE "inquiry" ("event_id" INTEGER PRIMARY KEY, "inquiry_id" TEXT, "question" TEXT, "hypothesis" TEXT, "method" TEXT, "status" TEXT, "supersedes_status" TEXT, "reason" TEXT)`); err != nil {
		t.Fatal(err)
	}
	if _, err := old.db.Exec(`INSERT INTO "enum_event_type" ("value", "means") VALUES ('inquiry', 'former vocabulary')`); err != nil {
		t.Fatal(err)
	}
	id := old.event(t, "blue-respond", 0, ts1, "register")
	old.row(t, "register", `"event_id", "tool_version"`, id, "abc1234")
	propose := func(ts, q, question, method string) {
		id := old.event(t, "blue-respond", 0, ts, "inquiry")
		old.row(t, "inquiry", `"event_id", "inquiry_id", "question", "method", "status"`, id, q, question, method, "proposed")
	}
	fate := func(ts, q, status, method string) {
		id := old.event(t, "blue-respond", 0, ts, "inquiry")
		var m any
		if method != "" {
			m = method
		}
		old.row(t, "inquiry", `"event_id", "inquiry_id", "method", "status", "supersedes_status", "reason"`, id, q, m, status, "1", "7 divides 91")
	}
	propose(ts2, "Q-00000001", "whether trial division settles it", "trial division")
	propose(ts2, "Q-00000002", "whether a sieve settles it", "a sieve")
	propose(ts2, "Q-00000003", "whether a lookup table settles it", "a lookup table")
	fate(ts3, "Q-00000001", "underway", "")
	fate(ts4, "Q-00000001", "confirmed", "")
	fate(ts4, "Q-00000002", "declined", "a sieve to 100")

	src, err := migrate.OpenSQLite(old.records(), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer src.Close()
	dst := runtest.New(t, recordtest.TmpRun(t))
	res, err := migrate.Replay(src, migrate.Entries(), dst, migrate.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Refusals) != 0 {
		t.Fatalf("refusals: %+v", res.Refusals)
	}
	fam, err := record.FamilyOf(dst)
	if err != nil {
		t.Fatal(err)
	}
	ids := map[string]string{}
	for k, v := range res.IDs {
		ids[v] = k
	}
	back := func(id string) string {
		if k, ok := ids[id]; ok {
			return k
		}
		return id
	}
	var fates []string
	for _, e := range fam.Events {
		if q, ok := recordpb.BodyAs[*recordpb.Inquiry](e); ok {
			if q.SupersedesStatus == nil && q.GetMethod() != "" {
				t.Errorf("the migrated proposal of %s holds the method %q", q.GetInquiryId(), q.GetMethod())
			}
			if q.SupersedesStatus != nil {
				fates = append(fates, fmt.Sprintf("%s %s %q", back(q.GetInquiryId()), recordpb.Word(q.GetStatus()), q.GetMethod()))
			}
		}
	}
	want := []string{`Q-00000001 underway "trial division"`, `Q-00000001 confirmed "trial division"`, `Q-00000002 declined "a sieve to 100"`}
	if strings.Join(fates, "\n") != strings.Join(want, "\n") {
		t.Errorf("the migrated fates:\n%s\nwant\n%s", strings.Join(fates, "\n"), strings.Join(want, "\n"))
	}
	for _, a := range record.InquiriesOf(fam.Events, fam.At) {
		if back(a.ID) == "Q-00000003" && a.Method != "" {
			t.Errorf("Q3 has no fate to carry its proposal's method to, and holds %q", a.Method)
		}
	}
}
