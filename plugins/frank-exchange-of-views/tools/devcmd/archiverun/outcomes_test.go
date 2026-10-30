package main

import (
	"strings"
	"testing"
)

var (
	pinned = map[string]string{"python3": "Python 3.12.16", "node": "v24.19.0"}
	rows   = []Row{{"a.tar.gz", "0b15", Reproduced}, {"a.tar.gz", "62d0", Moved}, {"b.tar.gz", "ab6a", Differs}}
)

// WHAT -write WRITES, THE CHECK ACCEPTS, and only under the interpreters it names.
func TestTheWrittenOutcomesAreTheCheckedOnes(t *testing.T) {
	if p := Compare(Format(pinned, rows), pinned, rows); len(p) != 0 {
		t.Fatalf("the written file disagrees with its own measurement: %q", p)
	}
}

// EVERY DISAGREEMENT IS NAMED, AND A FILE THAT MEASURED NOTHING FAILS: an empty or header-only file
// is NOT MEASURED, never a pass over zero rows.
func TestEveryDisagreementIsNamed(t *testing.T) {
	held := Format(pinned, rows)
	for _, c := range []struct {
		name, held string
		versions   map[string]string
		rows       []Row
		want       string
	}{
		{"an empty file", "", pinned, rows, "NOT MEASURED"},
		{"a header and no rows", Format(pinned, nil), pinned, rows, "NOT MEASURED"},
		{"another interpreter", held, map[string]string{"python3": "Python 3.12.3", "node": "v24.19.0"}, rows, `interpreter python3 is "Python 3.12.3"`},
		{"an absent interpreter", held, map[string]string{"python3": "absent", "node": "v24.19.0"}, rows, `interpreter python3 is "absent"`},
		{"an interpreter the file lacks", held, map[string]string{"python3": "Python 3.12.16", "node": "v24.19.0", "bash": "x"}, rows, "interpreter bash"},
		{"an interpreter the check lacks", held, map[string]string{"python3": "Python 3.12.16"}, rows, "names interpreter node"},
		{"a moved outcome", held, pinned, []Row{rows[0], {"a.tar.gz", "62d0", Differs}, rows[2]}, "re-ran differs, the file holds moved"},
		{"a row the file lacks", held, pinned, append(append([]Row{}, rows...), Row{"c.tar.gz", "ffff", Reproduced}), "holds no row"},
		{"a row no archive re-ran", held, pinned, rows[:2], "b.tar.gz ab6a: the file holds a row no archive re-ran"},
	} {
		t.Run(c.name, func(t *testing.T) {
			p := Compare(c.held, c.versions, c.rows)
			if !strings.Contains(strings.Join(p, "\n"), c.want) {
				t.Errorf("problems %q, want one naming %q", p, c.want)
			}
		})
	}
}
