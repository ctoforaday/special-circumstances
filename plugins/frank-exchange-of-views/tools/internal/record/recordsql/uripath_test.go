package recordsql

import (
	"os"
	"path/filepath"
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
	for _, name := range []string{"plain", "hash#01", "query?a=1", "pct%20"} {
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

// AND THE ESCAPING IS ORDER-DEPENDENT: `%` has to go first, or it corrupts the escapes that
// follow it and the path is wrong in a second, quieter way.
func TestURIPathEscapesPercentBeforeTheRest(t *testing.T) {
	for _, c := range []struct{ in, want string }{
		{"/runs/plain", "/runs/plain"},
		{"/runs/hash#01", "/runs/hash%2301"},
		{"/runs/query?a=1", "/runs/query%3fa=1"},
		{"/runs/100%done", "/runs/100%25done"},
		{"/runs/%23", "/runs/%2523"}, // an already-literal `%23` stays a literal, not a `#`
	} {
		if got := uriPath(c.in); got != c.want {
			t.Errorf("uriPath(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
