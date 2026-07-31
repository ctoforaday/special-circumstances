package recallindex

import (
	"strings"
	"testing"
	"time"
)

func TestDecide(t *testing.T) {
	cases := []struct {
		name    string
		qmd     bool
		file    string
		run     bool
		mention string
	}{
		{"absent qmd skips and points to doctor", false, "report.md", false, "doctor --fix"},
		{"non-markdown never indexes", true, "main.go", false, "not markdown"},
		{"markdown write triggers update", true, "blue/report.md", true, "qmd update"},
		{"extension match is case-insensitive", true, "NOTES.MD", true, "qmd update"},
		{"unknown file (no path in payload) skips", true, "unknown", false, "not markdown"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			run, reason := decide(c.qmd, c.file)
			if run != c.run {
				t.Fatalf("decide(%v,%q) run = %v; want %v", c.qmd, c.file, run, c.run)
			}
			if !strings.Contains(reason, c.mention) {
				t.Fatalf("decide(%v,%q) reason = %q; missing %q", c.qmd, c.file, reason, c.mention)
			}
			if !strings.Contains(reason, c.file) {
				t.Fatalf("reason must carry the file for the hook log: %q", reason)
			}
		})
	}
}

func TestFileFrom(t *testing.T) {
	var in hookInput
	if got := fileFrom(in); got != "unknown" {
		t.Fatalf("empty payload should yield unknown: %q", got)
	}
	in.ToolInput.Path = "p.md"
	if got := fileFrom(in); got != "p.md" {
		t.Fatalf("fileFrom should use path: %q", got)
	}
	in.ToolInput.FilePath = "f.md"
	if got := fileFrom(in); got != "f.md" {
		t.Fatalf("file_path should win over path: %q", got)
	}
}

func TestShouldUpdate(t *testing.T) {
	now := time.Now()
	if !shouldUpdate(now, time.Time{}, time.Minute) {
		t.Fatal("zero stamp (first write ever) must update")
	}
	if shouldUpdate(now, now.Add(-30*time.Second), time.Minute) {
		t.Fatal("stamp inside the window must debounce")
	}
	if !shouldUpdate(now, now.Add(-90*time.Second), time.Minute) {
		t.Fatal("stamp past the window must update")
	}
	// The BOUNDARY itself. `>=` and `>` differ on exactly one input, and a mutation
	// sweep confirmed that changing it broke nothing: the cases above straddle the
	// window at 30s and 90s and never land on it. An inequality boundary is the
	// classic off-by-one that statement coverage always reports as covered.
	if !shouldUpdate(now, now.Add(-time.Minute), time.Minute) {
		t.Error("a stamp exactly one interval old must update — the window is closed at its end, not open")
	}
	if shouldUpdate(now, now.Add(-time.Minute+time.Nanosecond), time.Minute) {
		t.Error("one nanosecond short of the interval must still debounce")
	}
	// A stamp in the FUTURE (clock skew, a restored file, a hand-edited stamp) must
	// not read as "long overdue" via a negative duration.
	if shouldUpdate(now, now.Add(time.Hour), time.Minute) {
		t.Error("a future stamp must debounce, not update")
	}
}
