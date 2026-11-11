package migrate_test

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migrate"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/report"
)

// eiffel is the one archived run that holds a reading's in-act mint (EventSchema 56): four readings of
// a citation that minted their gap in their own act (source events 165/166/167, 789/790/791,
// 792/793/794, 795/796/797), two of a source the lens found (180/181/182, 540/541/542), and one held
// reading of a citation naming another act's gap (798).
const eiffel = "2026-11-04_eiffel-tower-height.tar.gz"

// readLens is the Eiffel lens whose readings these are.
const readLens = "red-lens-evidence"

// T13. THE EIFFEL READING ACTS COME FORWARD, each member its own event (plans/gap-is-the-one-holder.md
// §III.7, §III.11): zero refusals at this binary's epoch; the four cited in-act Mints come forward
// ABOUT their citation — about_kind anchor, about_ref the reading's anchor, no quote — with their
// Anchor rows still placing the gap at the citation's sentence; the two found-source Mints keep their
// quote; no Verify names a gap and the held reading is a Verify alone; the six class-32 gaps and the 25
// are on the gap view with their closures.
//
// THE READING BEHIND EACH GAP, read off the migrated record cut at a source event (migrateEiffelThrough):
// through 222, 167 (`absent`) behind G-8dbfce13 and 182 (`refutes`) behind G-3b2b7d93; through 237,
// 182 still; through 691, 542 behind G-1d97d860; through 845, 794 and 797 behind G-dd7b403e and
// G-de476176. At head none stands behind any of the six: C-2dc96eab is retired (223); the found
// readings' claims were taken out by edits 238 and 692 while each gap's token stayed in the sentence;
// the three repaired 32s' citations were re-read `supports` (952/953/955), which is no reading a mint
// of the class stands on (record.readingBehind asks the class's own predicate of the standing reading).
// G-f66af91d's reading 791 is of a sentence edit 835 changed, so through 845 none stands behind it.
//
// NO READING DEBT STANDS at head or through any cut: 182 and 542 are LOST (edits 238, 692) and each
// joins its gap as of the report before the loss, both minted in the reading's sitting; the held reading
// 798 of C-3094c824 is discharged through 845 by G-1d97d860, the open class-32 gap whose token the
// citation's sentence holds (the F9 cross-arm), until edit 846 names the citation. docket.md's "read
// for the sentence" line names the citation's url for each open gap a reading stands behind.
func TestTheEiffelReadingActsComeForward(t *testing.T) {
	res, dir := migrate.MigratedArchive(t, eiffel)
	run := runtest.Open(t, dir)
	if len(res.Refusals) != 0 {
		t.Fatalf("%d event(s) refused; the first: %+v", len(res.Refusals), res.Refusals[0])
	}
	if res.EventSchema != recordpb.EventSchema {
		t.Fatalf("migrated at epoch %d, this binary writes %d", res.EventSchema, recordpb.EventSchema)
	}
	db := openRO(t, dir)
	src := openRO(t, recordtest.RequireArchive(t, eiffel))

	// THE FOUR CITED IN-ACT MINTS ARE ABOUT THEIR CITATION, and their Anchor rows still place them.
	cited := []struct{ gap, citation string }{
		{"G-8dbfce13", "C-2dc96eab"},
		{"G-f66af91d", "C-a51c49f1"},
		{"G-dd7b403e", "C-a1560eb7"},
		{"G-de476176", "C-486ed6b5"},
	}
	for _, c := range cited {
		var kind, ref, loc sql.NullString
		if err := db.QueryRow(`SELECT "about_kind", "about_ref", "location" FROM "mint" WHERE "gap_id" = ?`, c.gap).Scan(&kind, &ref, &loc); err != nil {
			t.Fatalf("mint %s: %v", c.gap, err)
		}
		if kind.String != "anchor" || ref.String != c.citation || loc.Valid {
			t.Errorf("mint %s is about_kind %q about_ref %q location %v; want anchor %s and no quote", c.gap, kind.String, ref.String, loc, c.citation)
		}
		var sentenceOf sql.NullString
		if err := db.QueryRow(`SELECT "sentence_of" FROM "anchor" WHERE "id" = ?`, c.gap).Scan(&sentenceOf); err != nil {
			t.Fatalf("anchor %s: %v", c.gap, err)
		}
		if sentenceOf.String != c.citation {
			t.Errorf("anchor %s places at sentence_of %q, want %s: the in-act Anchor is a report op and stays", c.gap, sentenceOf.String, c.citation)
		}
	}
	// THE TWO FOUND-SOURCE MINTS KEEP THEIR QUOTE: already the live mint of that subject.
	for _, gap := range []string{"G-3b2b7d93", "G-1d97d860"} {
		var kind, loc sql.NullString
		if err := db.QueryRow(`SELECT "about_kind", "location" FROM "mint" WHERE "gap_id" = ?`, gap).Scan(&kind, &loc); err != nil {
			t.Fatalf("mint %s: %v", gap, err)
		}
		if kind.Valid || !loc.Valid || loc.String == "" {
			t.Errorf("found-source mint %s is about_kind %v location %v; want no about and its quote", gap, kind, loc)
		}
	}
	// NO VERIFY NAMES A GAP: the columns are not in the schema, and the held reading is a Verify alone.
	for _, col := range verifyColumns(t, db) {
		if col == "carried_by" || col == "carrier_minted" {
			t.Errorf("the migrated verify table carries %q", col)
		}
	}
	var held int
	if err := db.QueryRow(`SELECT count(*) FROM "verify" v JOIN "events" e ON e."id" = v."event_id" WHERE v."anchor" = 'C-3094c824' AND v."outcome" = 'refutes' AND e."seat_id" = ?`, readLens).Scan(&held); err != nil {
		t.Fatal(err)
	}
	if held != 1 {
		t.Errorf("the held reading of C-3094c824 came forward %d time(s), want once", held)
	}
	if got, want := res.Out["verify"], res.In["verify"]; got != want {
		t.Errorf("%d readings in, %d out", want, got)
	}

	// THE SEVEN GAPS ARE ON THE GAP VIEW WITH THEIR CLOSURES.
	fam, err := record.FamilyOf(run)
	if err != nil {
		t.Fatal(err)
	}
	for _, gap := range []string{"G-8dbfce13", "G-3b2b7d93", "G-1d97d860", "G-f66af91d", "G-dd7b403e", "G-de476176", "G-b2ba7d66"} {
		g := fam.Gap(gap)
		if g == nil || g.Open || g.Closure == nil || g.Closure.GetClosureClass() != recordpb.Disposition_DISPOSITION_REPAIRED {
			t.Errorf("gap %s on the migrated gap view: %+v; want closed `repaired`", gap, g)
		}
	}

	// THE READING BEHIND EACH GAP, read off the migrated record with the report as it stands.
	behind := map[string]string{}
	for _, gap := range []string{"G-8dbfce13", "G-3b2b7d93", "G-1d97d860", "G-f66af91d", "G-dd7b403e", "G-de476176"} {
		v, id, has, err := record.ReadingBehind(run, gap)
		if err != nil {
			t.Fatal(err)
		}
		if has {
			behind[gap] = fmt.Sprintf("%s %s", recordpb.Word(v.GetOutcome()), sourceID(t, src, db, id))
		}
	}
	for gap, r := range behind {
		t.Errorf("ReadingBehind(%s) at head = %q, want none: its citation is retired, its claim no longer locates, or its citation was re-read `supports`", gap, r)
	}

	// NO DEBT AT HEAD, and no failed check on a refuted citation.
	for _, b := range passBlockers(t, run) {
		if b.Kind == record.BlockerReadingDebt || (b.Kind == record.BlockerFailedCheck && strings.HasPrefix(b.Subject, "C-")) {
			t.Errorf("at head the PASS is held by %s on %s (owner %s); want no reading debt and no failed check on a citation", b.Kind, b.Subject, b.Owner)
		}
	}

	// THE RECORD CUT AT A SOURCE EVENT: the reading behind each gap as it stood, and the debts.
	gaps := []string{"G-8dbfce13", "G-3b2b7d93", "G-1d97d860", "G-f66af91d", "G-dd7b403e", "G-de476176"}
	for _, cut := range []struct {
		through int64
		behind  map[string]string
	}{
		{222, map[string]string{"G-8dbfce13": "absent 167", "G-3b2b7d93": "refutes 182"}},
		{237, map[string]string{"G-3b2b7d93": "refutes 182"}},
		{691, map[string]string{"G-1d97d860": "refutes 542"}},
		{845, map[string]string{"G-dd7b403e": "refutes 794", "G-de476176": "refutes 797"}},
	} {
		c := migrateEiffelThrough(t, cut.through)
		cdb := openRO(t, c.Dir())
		got := map[string]string{}
		for _, gap := range gaps {
			v, id, has, err := record.ReadingBehind(c, gap)
			if err != nil {
				t.Fatal(err)
			}
			if has {
				got[gap] = fmt.Sprintf("%s %s", recordpb.Word(v.GetOutcome()), sourceID(t, src, cdb, id))
			}
		}
		if fmt.Sprint(got) != fmt.Sprint(cut.behind) {
			t.Errorf("through %d ReadingBehind = %v, want %v", cut.through, got, cut.behind)
		}
		for _, b := range passBlockers(t, c) {
			if b.Kind == record.BlockerReadingDebt {
				t.Errorf("through %d a reading debt stands: %s on %q (owner %s, since source event %s, `%s`)", cut.through, b.Kind, b.Subject, b.Owner, sourceID(t, src, cdb, b.Since()), recordpb.Word(b.Outcome()))
			}
		}
		if cut.through != 845 {
			continue
		}
		// THE HELD READING'S DISCHARGE IS THE OPEN GAP IN ITS CITATION'S SENTENCE: the sentence that
		// holds C-3094c824 holds G-1d97d860's token through 845.
		md, err := record.RenderRun(c)
		if err != nil {
			t.Fatal(err)
		}
		if i := strings.Index(md, "<!--cite:C-3094c824-->"); i < 0 || !strings.Contains(sentenceAround(md, i), "<!--gap:G-1d97d860-->") {
			t.Errorf("through 845 the sentence holding C-3094c824 does not hold G-1d97d860, which is what discharges reading 798's debt")
		}
		// docket.md names the source read for each open gap a reading stands behind.
		if _, err := report.Assemble(c); err != nil {
			t.Fatal(err)
		}
		docket, err := os.ReadFile(filepath.Join(c.Dir(), report.FileDocket))
		if err != nil {
			t.Fatal(err)
		}
		for gap, want := range map[string]string{
			"G-dd7b403e": "read for the sentence: https://en.wikipedia.org/wiki/Eiffel_Tower — " + readLens + "'s `refutes` reading, event ",
			"G-de476176": "read for the sentence: https://en.wikipedia.org/wiki/Eiffel_Tower — " + readLens + "'s `refutes` reading, event ",
		} {
			if block := gapBlockOf(string(docket), gap); !strings.Contains(block, want) {
				t.Errorf("docket.md through 845, block for %s, does not say %q:\n%s", gap, want, block)
			}
		}
		if block := gapBlockOf(string(docket), "G-f66af91d"); strings.Contains(block, "read for the sentence") {
			t.Errorf("docket.md through 845 names a reading behind G-f66af91d, whose reading 791 is of a sentence edit 835 changed:\n%s", block)
		}
	}
}

// sentenceAround is the sentence of md that holds offset i: from the period or line break before it
// to the one after.
func sentenceAround(md string, i int) string {
	start := strings.LastIndexAny(md[:i], ".\n") + 1
	end := strings.IndexAny(md[i:], ".\n")
	if end < 0 {
		return md[start:]
	}
	return md[start : i+end]
}

// T14. A SYNTHETIC IN-ACT RECORD COMES FORWARD in both shapes, so the translation does not depend on
// Eiffel's particulars: a cited reading's act (its Mint about the citation), a found source's act (its
// Mint quoting), a held reading (a Verify alone) and a plain class-32 mint — zero refusals.
func TestASyntheticInActRecordComesForward(t *testing.T) {
	const lens, blue = "red-lens-evidence", "blue-respond"
	const s1, s2, s3 = "The tower is 330 metres tall.", "The flag was removed in 1957.", "The antenna was added in 2022."
	var o oldStream
	o.add(lens, "register", lens+":register:#1", map[string]any{"agent_id": "l1"})
	o.add(blue, "report_part", blue+":report_part:#1", map[string]any{"text": "# T — research report\n\n" + s1 + " " + s2 + " " + s3 + "\n"})
	o.add(blue, "cite", blue+":cite:C-aaaaaaa1", map[string]any{"label": "C-aaaaaaa1", "url": "https://ex/a", "title": "A",
		"location": s1, "access_date": "2026-11-01", "text": "the page", "source_text_origin": "embedded",
		"work_status": "not_checked", "source_completeness": "not_asked", "copy_substitution": "none"})
	mint := func(gap, quote string) map[string]any {
		return map[string]any{"gap_id": gap, "mint_key": "", "class": "claim_contradicts_source", "location": quote,
			"problem": "p", "required_fix": "f", "fix_basis": "proposed", "acceptance_check": "a", "check_kind": "source",
			"severity": "medium", "likelihood": "medium", "impact": "medium", "complexity_cost": "low", "mint_reason": "r", "class_material": "always"}
	}
	reading := func(claim, anchor, url, carried string, minted int64) map[string]any {
		f := map[string]any{"claim": claim, "outcome": "refutes", "confidence": "high", "text": "read it",
			"source_completeness": "not_asked", "copy_substitution": "none", "carried_by": carried, "carrier_minted": minted}
		if anchor != "" {
			f["anchor"] = anchor
		} else {
			f["url"], f["title"] = url, "F"
		}
		return f
	}
	// A cited reading's act.
	o.add(lens, "mint", lens+":mint:G-aaaaaaa1", mint("G-aaaaaaa1", s1))
	o.add(lens, "anchor", lens+":anchor:G-aaaaaaa1", map[string]any{"id": "G-aaaaaaa1", "sentence_of": "C-aaaaaaa1"})
	o.add(lens, "verify", lens+":verify:#1:C-aaaaaaa1", reading(s1, "C-aaaaaaa1", "", "G-aaaaaaa1", 1))
	// A found source's act.
	o.add(lens, "mint", lens+":mint:G-aaaaaaa2", mint("G-aaaaaaa2", s2))
	o.add(lens, "anchor", lens+":anchor:G-aaaaaaa2", map[string]any{"id": "G-aaaaaaa2", "location": s2})
	o.add(lens, "verify", lens+":verify:#2", reading(s2, "", "https://ex/f", "G-aaaaaaa2", 1))
	// A held reading by another lens, naming the cited act's gap.
	o.add("red-lens-logic", "verify", "red-lens-logic:verify:#1:C-aaaaaaa1", reading(s1, "C-aaaaaaa1", "", "G-aaaaaaa1", 0))
	// A plain mint, no reading.
	o.add(lens, "mint", lens+":mint:G-aaaaaaa3", mint("G-aaaaaaa3", s3))

	res, dst := replaySlice(t, o.evs, migrate.Options{})
	if len(res.Refusals) != 0 {
		t.Fatalf("refusals: %+v", res.Refusals)
	}
	if res.Out["verify"] != 3 || res.Out["mint"] != 3 {
		t.Errorf("out %v, want 3 readings and 3 mints", res.Out)
	}
	m, err := record.MergedEvents(dst)
	if err != nil {
		t.Fatal(err)
	}
	var mints []string
	for _, e := range m.Events {
		if mt, ok := recordpb.BodyAs[*recordpb.Mint](e); ok {
			mints = append(mints, fmt.Sprintf("%s about=%s/%s quote=%q", mt.GetGapId(), recordpb.Word(mt.GetAboutKind()), mt.GetAboutRef(), mt.GetLocation()))
		}
	}
	want := []string{
		`G-aaaaaaa1 about=anchor/C-aaaaaaa1 quote=""`,
		`G-aaaaaaa2 about=/ quote="` + s2 + `"`,
		`G-aaaaaaa3 about=/ quote="` + s3 + `"`,
	}
	if strings.Join(mints, "\n") != strings.Join(want, "\n") {
		t.Errorf("the migrated mints are\n%s\nwant\n%s", strings.Join(mints, "\n"), strings.Join(want, "\n"))
	}
}

// AN IN-ACT MINT THAT ALREADY NAMES A SUBJECT IS REFUSED, never overwritten: the archive then names two
// subjects for one gap — the reading's citation and the mint's own about — and the translation writes
// neither.
func TestAnInActMintAlreadyAboutSomethingIsRefused(t *testing.T) {
	const lens, blue = "red-lens-evidence", "blue-respond"
	const s1 = "The tower is 330 metres tall."
	var o oldStream
	o.add(lens, "register", lens+":register:#1", map[string]any{"agent_id": "l1"})
	o.add(blue, "report_part", blue+":report_part:#1", map[string]any{"text": "# T — research report\n\n## Findings\n\n" + s1 + "\n"})
	o.add(blue, "cite", blue+":cite:C-aaaaaaa1", map[string]any{"label": "C-aaaaaaa1", "url": "https://ex/a", "title": "A",
		"location": s1, "access_date": "2026-11-01", "text": "the page", "source_text_origin": "embedded",
		"work_status": "not_checked", "source_completeness": "not_asked", "copy_substitution": "none"})
	o.add(lens, "mint", lens+":mint:G-aaaaaaa1", map[string]any{"gap_id": "G-aaaaaaa1", "mint_key": "", "class": "claim_contradicts_source",
		"about_kind": "section", "about_ref": "Findings",
		"problem": "p", "required_fix": "f", "fix_basis": "proposed", "acceptance_check": "a", "check_kind": "source",
		"severity": "medium", "likelihood": "medium", "impact": "medium", "complexity_cost": "low", "mint_reason": "r", "class_material": "always"})
	o.add(lens, "verify", lens+":verify:#1:C-aaaaaaa1", map[string]any{"claim": s1, "anchor": "C-aaaaaaa1", "outcome": "refutes", "confidence": "high", "text": "read it",
		"source_completeness": "not_asked", "copy_substitution": "none", "carried_by": "G-aaaaaaa1", "carrier_minted": int64(1)})
	res, _ := replaySlice(t, o.evs, migrate.Options{})
	var got []string
	for _, r := range res.Refusals {
		got = append(got, r.Word+": "+r.Err)
	}
	want := "mint: gap G-aaaaaaa1 was minted in the act of a reading of citation C-aaaaaaa1 and is about section Findings already: the archive names two subjects for one gap, and the translation writes neither"
	if len(got) != 1 || got[0] != want {
		t.Errorf("refusals = %q, want the one naming both subjects", got)
	}
}

// migrateEiffelThrough migrates the Eiffel archive cut at source events.id through: a private copy
// of the extracted run with every row of a later event deleted, so the migrated record is the
// archive as it stood at that event and every fold over it reads that board.
func migrateEiffelThrough(t *testing.T, through int64) record.Run {
	t.Helper()
	from := filepath.Join(t.TempDir(), "from")
	if err := os.CopyFS(from, os.DirFS(recordtest.RequireArchive(t, eiffel))); err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite", filepath.Join(from, "records", "record.db"))
	if err != nil {
		t.Fatal(err)
	}
	// The archive's own append-only triggers refuse the cut; the copy is not the record.
	for _, trg := range namesOf(t, db, "trigger") {
		if _, err := db.Exec(fmt.Sprintf(`DROP TRIGGER %q`, trg)); err != nil {
			t.Fatal(err)
		}
	}
	tables := namesOf(t, db, "table")
	for _, tbl := range tables {
		col := ""
		for _, c := range columnsOf(t, db, tbl) {
			if (tbl == "events" && c == "id") || c == "event_id" {
				col = c
			}
		}
		if col == "" {
			continue
		}
		if _, err := db.Exec(fmt.Sprintf(`DELETE FROM %q WHERE %q > ?`, tbl, col), through); err != nil {
			t.Fatalf("cutting %s: %v", tbl, err)
		}
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	to := recordtest.TmpRun(t)
	res, err := migrate.Migrate(from, to, migrate.Entries(), migrate.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Refusals) != 0 {
		t.Fatalf("cut through %d: %d event(s) refused; the first: %+v", through, len(res.Refusals), res.Refusals[0])
	}
	return runtest.Open(t, to)
}

// namesOf is the names of the schema objects of one kind in db.
func namesOf(t *testing.T, db *sql.DB, kind string) []string {
	t.Helper()
	rows, err := db.Query(`SELECT "name" FROM "sqlite_master" WHERE "type" = ?`, kind)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var names []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			t.Fatal(err)
		}
		names = append(names, n)
	}
	return names
}

func passBlockers(t *testing.T, run record.Run) []record.Blocker {
	t.Helper()
	bs, err := record.PassBlockers(run)
	if err != nil {
		t.Fatal(err)
	}
	return bs
}

func verifyColumns(t *testing.T, db *sql.DB) []string {
	t.Helper()
	return columnsOf(t, db, "verify")
}

func columnsOf(t *testing.T, db *sql.DB, table string) []string {
	t.Helper()
	rows, err := db.Query(fmt.Sprintf(`PRAGMA table_info(%q)`, table))
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var cols []string
	for rows.Next() {
		var cid, notnull, pk int
		var name, typ string
		var dflt sql.NullString
		if err := rows.Scan(&cid, &name, &typ, &notnull, &dflt, &pk); err != nil {
			t.Fatal(err)
		}
		cols = append(cols, name)
	}
	return cols
}

// sourceID is the archive's events.id of the event the migrated record db wrote as id: the migrated
// event's key is the source's, and the source's events.id is what the plan's rows name.
func sourceID(t *testing.T, src, db *sql.DB, id int64) string {
	t.Helper()
	var key string
	if err := db.QueryRow(`SELECT "key" FROM "events" WHERE "id" = ?`, id).Scan(&key); err != nil {
		t.Fatalf("event %d: %v", id, err)
	}
	var sid int64
	if err := src.QueryRow(`SELECT "id" FROM "events" WHERE "key" = ?`, key).Scan(&sid); err != nil {
		t.Fatalf("source event with key %q: %v", key, err)
	}
	return fmt.Sprint(sid)
}

// gapBlockOf is docket.md's block for gap g: from its heading to the blank line that ends it.
func gapBlockOf(docket, g string) string {
	i := strings.Index(docket, "### "+g)
	if i < 0 {
		return ""
	}
	rest := docket[i:]
	if j := strings.Index(rest, "\n\n"); j >= 0 {
		return rest[:j]
	}
	return rest
}
