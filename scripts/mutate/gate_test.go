package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func entry(file, mut, text, why string) recordEntry {
	return recordEntry{File: file, Mutation: mut, LineText: text, Why: why}
}

func surv(file, from, to, text string) survivor {
	return survivor{rel: file, line: 1, mutation: mutation{from: from, to: to}, text: text}
}

// The reconciliation is MULTISET, both directions: two identical lines are two mutants and
// need two judgements, and an entry no survivor matches is rot, not slack.
func TestReconcileIsMultisetBothWays(t *testing.T) {
	survived := []survivor{
		surv("a.go", "==", "!=", "if a == b {"),
		surv("a.go", "==", "!=", "if a == b {"), // the same text twice: two mutants
		surv("b.go", "&&", "||", "if x && y {"),
	}
	entries := []recordEntry{
		entry("a.go", "==->!=", "if a == b {", "equivalent"),
		entry("c.go", ">=->>", "if n >= 0 {", "explains a survivor that no longer exists"),
	}
	unexplained, stale := reconcile(survived, entries)
	if len(unexplained) != 2 {
		t.Errorf("unexplained = %v, want the duplicate a.go mutant AND the b.go mutant", unexplained)
	}
	if len(stale) != 1 || stale[0].File != "c.go" {
		t.Errorf("stale = %v, want exactly the c.go entry", stale)
	}
	if u, s := reconcile(nil, nil); len(u)+len(s) != 0 {
		t.Errorf("empty vs empty reconciled to %v / %v", u, s)
	}
}

// An absent record is an empty one — a module with no survivors owes no file. A malformed
// record, or an entry with no why, is an ERROR: "unreadable" reported as "empty" would let
// a corrupt record pass a release.
func TestLoadRecordRefusesWhatItCannotJudge(t *testing.T) {
	dir := t.TempDir()
	if got, err := loadRecord(dir); err != nil || got != nil {
		t.Errorf("absent record = (%v, %v), want (nil, nil)", got, err)
	}
	write := func(body string) {
		if err := os.WriteFile(filepath.Join(dir, RecordName), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("{not json")
	if _, err := loadRecord(dir); err == nil || !strings.Contains(err.Error(), "unreadable") {
		t.Errorf("malformed record = %v, want an unreadable error", err)
	}
	write(`[{"file":"a.go","mutation":"==->!=","line_text":"if a == b {","why":""}]`)
	if _, err := loadRecord(dir); err == nil || !strings.Contains(err.Error(), "no why") {
		t.Errorf("whyless entry = %v, want a refusal naming the missing explanation", err)
	}
	write(`[{"file":"a.go","mutation":"==->!=","line_text":"if a == b {","why":"equivalent"}]`)
	if got, err := loadRecord(dir); err != nil || len(got) != 1 {
		t.Errorf("valid record = (%v, %v), want one entry", got, err)
	}
}

// The gate's verdicts, end to end over a result: measured-and-clean passes, unexplained and
// stale both refuse, and a sweep that generated nothing refuses LOUDLY — zero survivors from
// a sweep that measured nothing is the same bytes as a clean module.
func TestGateVerdict(t *testing.T) {
	dir := t.TempDir()
	var out strings.Builder

	if code := gateVerdict(dir, &result{}, &out); code != 1 || !strings.Contains(out.String(), "NOT MEASURED") {
		t.Errorf("empty sweep: code %d, out %q — want a loud not-measured refusal", code, out.String())
	}

	out.Reset()
	if code := gateVerdict(dir, &result{killed: 5}, &out); code != 0 {
		t.Errorf("no survivors, no record: code %d (%s), want a pass", code, out.String())
	}

	out.Reset()
	res := &result{killed: 5, survived: []survivor{surv("a.go", "==", "!=", "if a == b {")}}
	if code := gateVerdict(dir, res, &out); code != 1 || !strings.Contains(out.String(), "UNEXPLAINED a.go:1") {
		t.Errorf("unexplained survivor: code %d, out %q", code, out.String())
	}

	if err := os.WriteFile(filepath.Join(dir, RecordName),
		[]byte(`[{"file":"a.go","mutation":"==->!=","line_text":"if a == b {","why":"equivalent"}]`), 0o644); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	if code := gateVerdict(dir, res, &out); code != 0 || !strings.Contains(out.String(), "GATE PASS") {
		t.Errorf("explained survivor: code %d, out %q", code, out.String())
	}

	out.Reset()
	if code := gateVerdict(dir, &result{killed: 5}, &out); code != 1 || !strings.Contains(out.String(), "STALE") {
		t.Errorf("stale entry after the survivor died: code %d, out %q", code, out.String())
	}
}
