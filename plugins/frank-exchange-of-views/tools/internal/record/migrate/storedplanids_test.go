package migrate_test

import (
	"database/sql"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migrate"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migration"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// This file is the oracle for stored plans across a renumbering. It reads both databases and never
// migrate's own table of id phrases or its newID map: an event is the same event on both records
// when its type, seat, key and stamp agree, and a number in a plan is read as a numeral by the
// token rule alone.

// eventRow is what identifies an event across the two records.
type eventRow struct{ typ, seat, key, ts string }

// eventsByID is every event of the record db holds, by events.id.
func eventsByID(t *testing.T, db *sql.DB) map[int64]eventRow {
	t.Helper()
	rows, err := db.Query(`SELECT "id", "type", "seat_id", "key", "ts" FROM "events"`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	out := map[int64]eventRow{}
	for rows.Next() {
		var id int64
		var r eventRow
		if err := rows.Scan(&id, &r.typ, &r.seat, &r.key, &r.ts); err != nil {
			t.Fatal(err)
		}
		out[id] = r
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return out
}

// storedPlan is one row of a record's plan table.
type storedPlan struct {
	at      int64
	printed string
}

// storedPlansOf is the record's plan table in stream order, or nil when the record has none (an
// archive older than epoch 31).
func storedPlansOf(t *testing.T, db *sql.DB) []storedPlan {
	t.Helper()
	var n int
	if err := db.QueryRow(`SELECT count(*) FROM sqlite_master WHERE type = 'table' AND name = 'plan'`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n == 0 {
		return nil
	}
	rows, err := db.Query(`SELECT "event_id", "printed" FROM "plan" ORDER BY "event_id"`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var out []storedPlan
	for rows.Next() {
		var p storedPlan
		if err := rows.Scan(&p.at, &p.printed); err != nil {
			t.Fatal(err)
		}
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return out
}

var planKey = regexp.MustCompile(`"key":"[0-9a-f]{16}"`)

// numeral is one all-digit token of a printing, and where it sits.
type numeral struct {
	start, end int
	n          int64
}

func wordByte(c byte) bool {
	return c >= '0' && c <= '9' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c == '_' || c == '-'
}

// numeralsOf is every numeral of s: a maximal [A-Za-z0-9_-] token that is all digits. The key is
// masked first, so its hex is never one.
func numeralsOf(s string) []numeral {
	s = planKey.ReplaceAllStringFunc(s, func(k string) string { return strings.Repeat("k", len(k)) })
	var out []numeral
	for i := 0; i < len(s); {
		if !wordByte(s[i]) {
			i++
			continue
		}
		j := i
		for j < len(s) && wordByte(s[j]) {
			j++
		}
		if n, err := strconv.ParseInt(s[i:j], 10, 64); err == nil && !strings.ContainsAny(s[i:j], "-_") {
			out = append(out, numeral{i, j, n})
		}
		i = j
	}
	return out
}

// maskedPlan is printed with the key and every numeral masked: two printings mask alike exactly
// when only numbers and the key tell them apart.
func maskedPlan(printed string) string {
	var b strings.Builder
	last := 0
	for _, n := range numeralsOf(printed) {
		b.WriteString(printed[last:n.start])
		b.WriteString("#")
		last = n.end
	}
	b.WriteString(printed[last:])
	return planKey.ReplaceAllString(b.String(), `"key":"#"`)
}

// moved is whether source event n sits at another number on the migrated record — or nowhere.
func moved(n int64, src, dst map[int64]eventRow) bool {
	e, ok := src[n]
	return ok && dst[n] != e
}

// planPairsOf migrates the named archive and returns the source's stored plans beside the migrated
// record's, and both records' events.
func planPairsOf(t *testing.T, name string) (srcPlans, dstPlans []storedPlan, src, dst map[int64]eventRow, res *migration.Manifest) {
	t.Helper()
	res, dir := migrate.MigratedArchive(t, name)
	ddb := openRO(t, dir)
	dstPlans, dst = storedPlansOf(t, ddb), eventsByID(t, ddb)
	// A jsonl-era source has no database, and stored no plan.
	srcDir := recordtest.ExtractArchive(t, name)
	if _, err := os.Stat(filepath.Join(srcDir, "records", "record.db")); err != nil {
		return nil, dstPlans, nil, dst, res
	}
	sdb := openRO(t, srcDir)
	return storedPlansOf(t, sdb), dstPlans, eventsByID(t, sdb), dst, res
}

// seatSpans is the seat words a plan's lines carry, in order: each Go-quoted span, each remand
// line's tail after the prefix, and each remand's direction. It is the oracle's own scan; a quote
// that does not close ends the line's spans with the rest of it.
func seatSpans(p record.Plan) []string {
	var out []string
	for _, line := range p.Why {
		for i := 0; i < len(line); {
			switch {
			case line[i] == '"':
				q, err := strconv.QuotedPrefix(line[i:])
				if err != nil {
					out = append(out, line[i:])
					i = len(line)
					continue
				}
				out = append(out, q)
				i += len(q)
			case strings.HasPrefix(line[i:], record.RemandDirectionPrefix):
				out = append(out, line[i:])
				i = len(line)
			default:
				i++
			}
		}
	}
	for _, r := range p.RemandOwed {
		out = append(out, r.Direction)
	}
	return out
}

// lineNumerals is the numerals of a plan's lines outside every seat span, line by line.
func lineNumerals(p record.Plan) [][]int64 {
	var out [][]int64
	for _, line := range p.Why {
		var ns []int64
		cut := line
		for _, s := range seatSpans(record.Plan{Why: []string{line}}) {
			cut = strings.Replace(cut, s, strings.Repeat(" ", len(s)), 1)
		}
		for _, n := range numeralsOf(cut) {
			ns = append(ns, n.n)
		}
		out = append(out, ns)
	}
	return out
}

// A STORED PLAN CROSSES A RENUMBERING NAMING THE EVENTS ITS SOURCE NAMED (F8). m28 holds two
// class_new events, a chair's closing and eight positions no chair filed, which migrate writes as
// none, so every later event lands up to eleven numbers lower; sixteen of its twenty stored plans
// name such an event — as a head, a pin, or a number a line
// prints after `head`, `pin`, `edit` or `event`. Driven on the real archive through Migrate, and
// checked against both databases rather than migrate's own map:
//
//	(i)   only numbers and the key changed;
//	(ii)  every number that changed, and the head and every pin, name on the migrated record the
//	      event the source number named on the source;
//	(iii) no line keeps the number of a source event the renumbering moved;
//	(iv)  the key is the decoded plan's own;
//	(v)   the seat words are byte for byte the source's (vacuous on m28; planids_internal_test
//	      carries it).
func TestAStoredPlanCrossesARenumbering(t *testing.T) {
	const name = "2026-10-18_is-91-prime-m28.tar.gz"
	recordtest.RequireArchive(t, name)
	srcPlans, dstPlans, src, dst, res := planPairsOf(t, name)
	if len(res.Refusals) != 0 {
		t.Fatalf("%d refusal(s); the first: %+v", len(res.Refusals), res.Refusals[0])
	}
	if len(srcPlans) != 20 || len(dstPlans) != len(srcPlans) {
		t.Fatalf("the source stores %d plan(s) and the migrated record %d; want m28's 20 on both", len(srcPlans), len(dstPlans))
	}
	// firstDropped is the first source event that writes none on the migrated record.
	written := map[eventRow]bool{}
	for _, e := range dst {
		written[e] = true
	}
	firstDropped, dropped := int64(0), map[string]int{}
	for id, e := range src {
		if written[e] {
			continue
		}
		dropped[e.typ]++
		if firstDropped == 0 || id < firstDropped {
			firstDropped = id
		}
	}
	if firstDropped == 0 {
		t.Fatal("fixture: every m28 event writes one on the migrated record, so nothing renumbers")
	}
	if want := map[string]int{"class_new": 2, "closing": 1, "position": 8}; !maps.Equal(dropped, want) {
		t.Errorf("m28's events that write none: %v, want %v", dropped, want)
	}
	same := func(a, b int64) bool { e, ok := src[a]; return ok && dst[b] == e }
	translated := 0
	for i := range srcPlans {
		s, d := srcPlans[i], dstPlans[i]
		s.printed = kindsTranslated(t, s.printed)
		if !same(s.at, d.at) {
			t.Errorf("plan %d: the migrated plan event %d is not the source's %d", i, d.at, s.at)
		}
		namesMoved := false
		for _, n := range numeralsOf(s.printed) {
			if n.n >= firstDropped && moved(n.n, src, dst) {
				namesMoved = true
			}
		}
		if !namesMoved {
			if s.printed != d.printed {
				t.Errorf("plan at source event %d names no moved event and is not byte-identical:\nsource   %s\nmigrated %s", s.at, s.printed, d.printed)
			}
			continue
		}
		translated++
		// (i)
		if maskedPlan(s.printed) != maskedPlan(d.printed) {
			t.Errorf("plan at source event %d: more than numbers and the key changed:\nsource   %s\nmigrated %s", s.at, s.printed, d.printed)
			continue
		}
		// (ii) every numeral position that differs
		sn, dn := numeralsOf(s.printed), numeralsOf(d.printed)
		for k := range sn {
			if sn[k].n != dn[k].n && !same(sn[k].n, dn[k].n) {
				t.Errorf("plan at source event %d: %d became %d, which is not the event the source's %d named", s.at, sn[k].n, dn[k].n, sn[k].n)
			}
		}
		sp, _, err := record.ReadPlan([]byte(s.printed))
		if err != nil {
			t.Fatal(err)
		}
		dp, _, err := record.ReadPlan([]byte(d.printed))
		if err != nil {
			t.Fatal(err)
		}
		// (ii) the head and each pin
		if sp.Head != 0 && !same(sp.Head, dp.Head) {
			t.Errorf("plan at source event %d: head %d migrated as %d, another event", s.at, sp.Head, dp.Head)
		}
		for k := range sp.StaleAreas {
			if a, b := sp.StaleAreas[k].Pin, dp.StaleAreas[k].Pin; a != 0 && !same(a, b) {
				t.Errorf("plan at source event %d: stale area %s's pin %d migrated as %d, another event", s.at, sp.StaleAreas[k].SeatID, a, b)
			}
		}
		// (iii)
		sl, dl := lineNumerals(sp), lineNumerals(dp)
		for k := range sl {
			for j, n := range sl[k] {
				if n >= firstDropped && moved(n, src, dst) && dl[k][j] == n {
					t.Errorf("plan at source event %d: line %q keeps %d, a moved event's number", s.at, dp.Why[k], n)
				}
			}
		}
		// (iv)
		if want := dp.WithKey().Key; dp.Key != want {
			t.Errorf("plan at source event %d: key %s, the plan's own is %s", s.at, dp.Key, want)
		}
		// (v)
		if a, b := seatSpans(sp), seatSpans(dp); strings.Join(a, "\x00") != strings.Join(b, "\x00") {
			t.Errorf("plan at source event %d: seat words changed: %q -> %q", s.at, a, b)
		}
	}
	if translated != 16 {
		t.Errorf("%d of m28's stored plans name a moved event; want its 16", translated)
	}
}
