package debatejs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// EVERY LABEL NAMES THE RUN BY ITS DIRECTORY, without the date the directory's name opens with.
// The run directory is `<date>_<slug>`; a trailing separator, or a name with no date before its
// first `_`, keeps what is there rather than losing a part of it.
func TestEveryLabelNamesTheRunByItsDirectory(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	backend := func(seatID, label, prompt string) Envelope {
		e := Envelope{"synopsis": "t", "rulings": []any{}, "dispositions": []any{},
			"manifest": []any{}, "saturation_reached": false, "open_gaps": 0}
		if seatID == "red-chair" {
			e["plan"] = planWith([]any{}, []any{})
		}
		return e
	}
	for name, want := range map[string]string{
		"2026-10-19_is-91-prime/": "is-91-prime",
		"2026-10-19_a_b":          "a_b",
		"draft_one":               "draft_one",
		"plain":                   "plain",
	} {
		dir := filepath.ToSlash(t.TempDir()) + "/" + name
		ds, err := Capture(ScriptPath(filepath.Join(wd, "..", "..")), Config{Topic: "t", RunDir: dir, BinDir: t.TempDir(),
			Lanes: 1, Model: "haiku", JudgmentModel: "haiku", Backend: backend})
		if err != nil {
			t.Fatalf("driving debate.js in %s: %v", dir, err)
		}
		if len(ds) == 0 {
			t.Fatalf("%s: no seat was dispatched, so no label was read", dir)
		}
		for _, d := range ds {
			if !strings.HasSuffix(d.Label, " · "+want) {
				t.Errorf("run directory %q: %s is labelled %q, want it to end %q", dir, d.SeatID, d.Label, " · "+want)
			}
		}
	}
}
