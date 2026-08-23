package recordsql

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A RUN DIRECTORY IS A FILESYSTEM PATH, NOT A URI, AND THE DSN IS BOTH.
//
// `file:` + path makes the DSN a URI, so `#` opens a fragment and `?` opens the query: a path
// containing either was TRUNCATED there, silently, to a path that still opens. Two runs then share
// one database and every write succeeds, which is why nothing reported it.
//
// FOUND BY ACCIDENT, 2026-08-23. A Go subtest whose name repeats gets `#01` appended and t.TempDir
// builds its directory from the test name, so the second of two identically-named cases opened the
// FIRST one's database — surfacing as "this seat has already recorded a mint this sitting" in a run
// that had recorded nothing at all.
//
// It is reachable outside a test: run directories are named from the topic, and "C# concurrency"
// produces exactly this path.
func TestARunDirectoryWithURIPunctuationGetsItsOwnDatabase(t *testing.T) {
	base := t.TempDir()
	// THE RESERVED SET IS EXACTLY THREE, MEASURED RATHER THAN ASSUMED. With the escape removed,
	// `hash#01`, `query?a=1` and `pct%20` are the only names that land elsewhere; the rest pass
	// through untouched because SQLite's URI grammar reserves only `?` and `#`, and `%` is the
	// escape introducer — `pct%20` unescaped decodes to a space and opens a different file. The
	// benign names are here so an over-broad escape shows up as one of THEM breaking.
	for _, name := range []string{
		"plain", "hash#01", "query?a=1", "pct%20",
		"with space", "amp&sand", "semi;colon", "at@sign", "plus+plus", "brack[et]", "uni-日本語",
	} {
		t.Run(name, func(t *testing.T) {
			dir := filepath.Join(base, name)
			if err := os.MkdirAll(dir, 0o755); err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(dir, "record.db")
			db, err := Open(path)
			if err != nil {
				t.Fatalf("Open(%q) = %v", path, err)
			}
			defer db.Close()

			// THE FILE MUST LAND WHERE IT WAS ASKED FOR. A truncated URI opens SOMETHING — that is
			// the whole problem — so asserting the open succeeded proves nothing. Assert the bytes
			// are at the path the caller named.
			if _, err := os.Stat(path); err != nil {
				t.Errorf("Open(%q) succeeded and no database exists there: %v.\n\n"+
					"The DSN is a URI, so the path was truncated at its punctuation and the database "+
					"was created somewhere else — which two runs can then share, with every write "+
					"reporting success.", path, err)
			}
		})
	}
}

// NOT MEASURED HERE, AND SAYING SO RATHER THAN IMPLYING COVERAGE: a path that BEGINS `//` parses
// as a URI authority, and no amount of percent-encoding fixes that — `/` cannot be escaped without
// changing the path. On Linux filepath.Join collapses doubled separators so it is unreachable; on
// WINDOWS, where this suite also runs, a UNC path (`\\server\share`) and a drive letter both put
// characters into the URI that this test cannot exercise on the box it runs on. If the store is
// ever opened against either, that is an unverified path, not a covered one.

// THE DSN IS BUILT BY net/url, AND THIS PINS THE SHAPE IT PRODUCES.
//
// The point is not the exact spelling — it is that the path is ESCAPED BY A URI BUILDER rather
// than concatenated, so a character nobody thought of is handled by the grammar's own rules and
// not by a list somebody has to keep. The `//` case is here because percent-encoding cannot fix
// it at all: `/` is not escapable without changing the path, and url.URL answers it by emitting
// an explicit EMPTY AUTHORITY instead.
func TestTheDSNIsBuiltByAURIBuilderRatherThanConcatenated(t *testing.T) {
	const q = "?_txlock=immediate&_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)"
	for _, c := range []struct{ path, want string }{
		{"/runs/plain/record.db", "file:///runs/plain/record.db" + q},
		{"/runs/hash#01/record.db", "file:///runs/hash%2301/record.db" + q},
		{"/runs/query?a=1/record.db", "file:///runs/query%3Fa=1/record.db" + q},
		{"/runs/100%done/record.db", "file:///runs/100%25done/record.db" + q},
		{"/runs/with space/record.db", "file:///runs/with%20space/record.db" + q},
		// A UNC-shaped path keeps its leading pair, behind an EMPTY authority.
		{"//server/share/record.db", "file:////server/share/record.db" + q},
	} {
		if got := dsnFor(c.path); got != c.want {
			t.Errorf("dsnFor(%q) =\n  %s\nwant\n  %s", c.path, got, c.want)
		}
	}
	// THE QUERY IS VERBATIM. Running it through url.Values would escape the parentheses in
	// `busy_timeout(5000)`, and a pragma the driver cannot parse is a setting silently not applied
	// — _txlock in particular is the difference between working and not.
	if got := dsnFor("/x/record.db"); !strings.Contains(got, "busy_timeout(5000)") {
		t.Errorf("the pragma parentheses were escaped out of the query: %s", got)
	}
}
