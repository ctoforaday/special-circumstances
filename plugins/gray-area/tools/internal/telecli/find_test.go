package telecli

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/gray-area/tools/internal/catalogue"
)

// peerTurn is a peer's message holding "conduct section", as the client writes one.
func peerTurn(uuid string) string {
	return turnLine(uuid, "S", "/w", 0, map[string]any{"isMeta": true, "origin": map[string]any{"kind": "peer"}},
		"is the conduct section settled?")
}

// matchIn is the ripRecord ripgrep would report for "conduct section" on line n of a file whose
// line n starts at byte lineStart and holds line.
func matchIn(path string, n int, lineStart int64, line string) ripRecord {
	i := strings.Index(line, "conduct section")
	return ripRecord{path: path, line: n, matches: []rawMatch{{abs: lineStart + int64(i), text: "conduct section"}}}
}

// EVERY WAY A RECORD CAN FAIL TO READ BACK IS COUNTED, one at a time and all together, and none of
// them can become a row. A record find could not read as ripgrep matched it was dropped in silence
// — or read as `?` — and under --in that made "none has a hit in C" out of a record that might
// have been C's.
func TestRecordsAtCountsWhatItCannotReadBack(t *testing.T) {
	dir := t.TempDir()
	write := func(name, body string) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		return p
	}
	first, second := peerTurn("p1"), peerTurn("p2")

	// The file changed since the search: the bytes at the offset are not the match.
	changed := write("changed.jsonl", first+"\n")
	changedRec := matchIn(changed, 1, 0, first)
	changedRec.matches[0].abs += 2
	// The line ripgrep matched is past the end of the file.
	short := write("short.jsonl", first+"\n")
	pastEnd := ripRecord{path: short, line: 3, matches: []rawMatch{{abs: 5000, text: "conduct section"}}}
	// The transcript went away.
	gone := ripRecord{path: filepath.Join(dir, "gone.jsonl"), line: 1, matches: []rawMatch{{abs: 10, text: "conduct section"}}}
	// The file grew AHEAD of the line since the search, so its absolute offset now lies before
	// the line's start: a negative Start.
	grown := write("grown.jsonl", first+"\n"+second+"\n")
	grownRec := matchIn(grown, 2, 0, second)
	// Truncated THROUGH its only match: the span runs past the torn line.
	cut := strings.Index(first, "conduct section") + 4
	through := write("through.jsonl", first[:cut])
	throughRec := matchIn(through, 1, 0, first)
	// Truncated AFTER its only match but before the line's end: span bytes intact, line
	// unparseable.
	after := write("after.jsonl", first[:strings.Index(first, "settled")])
	afterRec := matchIn(after, 1, 0, first)

	all := []ripRecord{changedRec, pastEnd, gone, grownRec, throughRec, afterRec}
	names := []string{"bytes differ", "line past the end", "file gone", "negative start", "truncated through the match", "truncated after the match"}
	paths := map[string]catalogue.TranscriptFile{}
	for i, r := range all {
		paths[r.path] = catalogue.TranscriptFile{Path: r.path, SessionID: "S" + names[i]}
		hits, stale := recordsAt(r.path, []ripRecord{r}, false)
		if stale != 1 {
			t.Errorf("%s: stale = %d, want 1", names[i], stale)
		}
		for _, h := range hits {
			if !h.Stale {
				t.Errorf("%s: a record not read back was returned as a current hit (%q)", names[i], h.Channel)
			}
		}
	}

	rows, matched, stale := decodeHits(all, paths, "")
	if stale != len(all) || matched != len(all) || len(rows) != 0 {
		t.Errorf("together: %d rows, %d matched, %d stale — want no row, %d matched, %d stale",
			len(rows), matched, stale, len(all), len(all))
	}

	// BOTH TRUNCATED CASES REFUSE through settle under --in: the silent zero criterion 4 names.
	// The after-the-match case is the one that fails if recordsAt drops DecodeHit's ok.
	for _, r := range []ripRecord{throughRec, afterRec} {
		rows, matched, stale := decodeHits([]ripRecord{r}, paths, catalogue.ChannelPeer)
		var out, errw bytes.Buffer
		err := settle(&out, &errw, rows, matched, stale, "conduct section", catalogue.ChannelPeer, false)
		if !errors.Is(err, errStaleZero) {
			t.Errorf("%s: settle returned %v with stdout %q, want errStaleZero", filepath.Base(r.path), err, out.String())
		}
	}
}

// parseRipgrep HAS NO LINE CAP: a bufio.Scanner with an 8 MB buffer dropped a longer match line
// and every line after it, with its error unchecked.
func TestParseRipgrepHasNoLineCap(t *testing.T) {
	long := strings.Repeat("x", 9<<20)
	raw := "a.jsonl\x001:0:first\n" + "b.jsonl\x002:40:" + long + "\n" + "c.jsonl\x003:70:th:ird\n"
	recs := parseRipgrep(raw)
	if len(recs) != 3 {
		t.Fatalf("%d records, want 3", len(recs))
	}
	if recs[1].path != "b.jsonl" || len(recs[1].matches[0].text) != len(long) {
		t.Errorf("the long match came back as %q with %d bytes", recs[1].path, len(recs[1].matches[0].text))
	}
	if c := recs[2]; c.path != "c.jsonl" || c.line != 3 || c.matches[0].abs != 70 || c.matches[0].text != "th:ird" {
		t.Errorf("the record after the long one read as %+v", c)
	}
}

// settle REFUSES A ZERO IT COULD NOT READ, filtered or not, and says nothing false beside it.
func TestSettleRefusesAZeroItCouldNotRead(t *testing.T) {
	run := func(rows []row, stale int, only catalogue.Channel) (string, string, error) {
		var out, errw bytes.Buffer
		err := settle(&out, &errw, rows, 2, stale, "conduct section", only, false)
		return out.String(), errw.String(), err
	}
	out, _, err := run(nil, 2, catalogue.ChannelPeer)
	if !errors.Is(err, errStaleZero) || strings.Contains(out, "none has a hit") {
		t.Errorf("--in peer, stale 2: err %v, stdout %q — want errStaleZero and no worded zero", err, out)
	}
	out, _, err = run(nil, 2, "")
	if !errors.Is(err, errStaleZero) || strings.Contains(out, "SESSION") {
		t.Errorf("unfiltered, stale 2: err %v, stdout %q — want errStaleZero and no table", err, out)
	}
	out, _, err = run(nil, 0, catalogue.ChannelPeer)
	if err != nil || !strings.Contains(out, `none has a hit in "peer"`) {
		t.Errorf("--in peer, stale 0: err %v, stdout %q — want the worded zero", err, out)
	}
	one := []row{{f: catalogue.TranscriptFile{Path: "p", SessionID: "S"}, hits: 1,
		best: catalogue.Hit{Channel: catalogue.ChannelUser, Snippet: "the conduct section"}}}
	out, errOut, err := run(one, 1, "")
	if err != nil || !strings.Contains(out, "the conduct section") ||
		!strings.Contains(errOut, "1 matching record(s) could not be read back as ripgrep matched them") {
		t.Errorf("rows and stale 1: err %v, stdout %q, stderr %q — want the rows and the stated count", err, out, errOut)
	}
}

// RIPGREP'S OFFSETS ARE THE FILE'S RAW BYTES only under --encoding none: by default it strips a
// UTF-8 byte-order mark and reports every offset after it three bytes short, so every record of
// such a file would read back stale. Run through real ripgrep with exactly find's flags.
func TestRipgrepOffsetsAreRawBytes(t *testing.T) {
	rg, err := exec.LookPath("rg")
	if err != nil {
		t.Skip("ripgrep is not installed, so `find` cannot be exercised here — a skipped assertion, not a passing one")
	}
	path := filepath.Join(t.TempDir(), "bom.jsonl")
	// Line 1 is the byte-order mark and an unrelated record (it never parses, and holds no match).
	body := "\xEF\xBB\xBF" + `{"type":"queue-operation","operation":"enqueue"}` + "\n" +
		turnLine("u1", "S", "/w", 0, origin("human"), []any{map[string]any{"type": "text", "text": "tidy the conduct section"}}) + "\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	raw, err := exec.Command(rg, ripgrepArgs("conduct section", false, []string{path})...).Output()
	if err != nil {
		t.Fatalf("ripgrep: %v", err)
	}
	recs := parseRipgrep(string(raw))
	if len(recs) != 1 {
		t.Fatalf("%d records from %q, want 1", len(recs), raw)
	}
	hits, stale := recordsAt(path, recs, false)
	if stale != 0 || len(hits) != 1 || hits[0].Channel != catalogue.ChannelUser {
		t.Errorf("stale %d, hits %+v — want the human's record read back as ripgrep matched it", stale, hits)
	}
}
