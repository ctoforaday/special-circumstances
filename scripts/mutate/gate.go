// The release gate: reconcile a sweep's survivors against the module's explained-survivor
// record (ruled 2026-09-03: the sweep is a RELEASE gate).
//
// WHAT THE GATE ENFORCES IS THE EXPLAINING, NOT A NUMBER. The tool's own header argues that
// driving survivors to zero buys contorted tests, and turning it into a gate does not repeal
// that: there is still no threshold. What a release must prove is that every survivor in the
// released module has been JUDGED — killed by a test, or carrying a written reason in the
// record — and that no reason outlives the survivor it explained. A stale entry is the
// allowlist rot facts-are-fields warns about, one level up, which is why it fails the gate
// rather than warning.
//
// Records are born at first gated release, not grandfathered (decided over bootstrapping
// them with UNEXPLAINED placeholders: a record whose entries nobody judged is an allowlist
// wearing a schema). The first release of each plugin pays its backlog.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// RecordName is the record's filename, at the swept module's root — it travels with the
// module it describes.
const RecordName = "mutation-survivors.json"

// recordEntry is one accepted survivor and the explanation that accepted it.
//
// THE KEY IS (file, mutation, line text), AND DELIBERATELY NOT THE LINE NUMBER. A line
// number rots with every edit above it, so keying on it would fail releases over unrelated
// diffs; text that changed is a different mutant nobody has judged yet, and re-judging it
// is exactly what should happen.
type recordEntry struct {
	File     string `json:"file"`
	Mutation string `json:"mutation"`
	LineText string `json:"line_text"`
	Why      string `json:"why"`
}

func (e recordEntry) key() string { return e.File + "\x00" + e.Mutation + "\x00" + e.LineText }

func survivorKey(s survivor) string {
	return s.rel + "\x00" + s.from + "->" + s.to + "\x00" + s.text
}

// loadRecord reads the module's record. AN ABSENT FILE IS AN EMPTY RECORD — a module whose
// sweep produces no survivors owes no explanations and no file. MALFORMED IS AN ERROR, never
// an absence, for the reason every record reader in this repository gives: "unreadable" and
// "empty" reported in the same breath would let a corrupt record pass a release. An entry
// with an empty why is refused at the read: without the explanation the record is an
// allowlist, and an allowlist is the thing this gate exists not to be.
func loadRecord(moduleDir string) ([]recordEntry, error) {
	path := filepath.Join(moduleDir, RecordName)
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var entries []recordEntry
	if err := json.Unmarshal(b, &entries); err != nil {
		return nil, fmt.Errorf("%s is unreadable: %w", path, err)
	}
	for i, e := range entries {
		if e.File == "" || e.Mutation == "" || e.LineText == "" {
			return nil, fmt.Errorf("%s entry %d is missing file, mutation or line_text — the gate cannot match it to anything", path, i)
		}
		if e.Why == "" {
			return nil, fmt.Errorf("%s entry %d (%s %s) has no why — an entry without its explanation is an allowlist line, not a judgement", path, i, e.File, e.Mutation)
		}
	}
	return entries, nil
}

// reconcile matches survivors against record entries as MULTISETS: two identical lines in
// one file are two mutants and need two judgements. It returns the survivors no entry
// explains and the entries no survivor matches.
func reconcile(survived []survivor, entries []recordEntry) (unexplained []survivor, stale []recordEntry) {
	counts := map[string]int{}
	for _, e := range entries {
		counts[e.key()]++
	}
	for _, s := range survived {
		if k := survivorKey(s); counts[k] > 0 {
			counts[k]--
			continue
		}
		unexplained = append(unexplained, s)
	}
	for _, e := range entries {
		if k := e.key(); counts[k] > 0 {
			counts[k]--
			stale = append(stale, e)
		}
	}
	return unexplained, stale
}

// gateVerdict is the release gate's judgement over a finished sweep: 0 to release, 1 to
// refuse. A sweep that generated NO behavioural mutants refuses: "0 survivors" from a sweep
// that measured nothing is the same bytes as a clean module, and the gate's first duty is to
// keep those apart.
func gateVerdict(moduleDir string, res *result, out io.Writer) int {
	if len(res.survived)+res.killed == 0 {
		fmt.Fprintln(out, "GATE FAIL: the sweep generated no behavioural mutants — this module is NOT MEASURED, which is not the same fact as clean")
		return 1
	}
	entries, err := loadRecord(moduleDir)
	if err != nil {
		fmt.Fprintln(out, "GATE FAIL:", err)
		return 1
	}
	unexplained, stale := reconcile(res.survived, entries)
	for _, s := range unexplained {
		fmt.Fprintf(out, "UNEXPLAINED %s:%d  %s->%s  |  %s\n", s.rel, s.line, s.from, s.to, s.text)
	}
	for _, e := range stale {
		fmt.Fprintf(out, "STALE %s  %s  |  %s  (was explained: %s)\n", e.File, e.Mutation, e.LineText, e.Why)
	}
	if len(unexplained)+len(stale) > 0 {
		fmt.Fprintf(out, "\nGATE FAIL: %d unexplained survivor(s) and %d stale record entr%s in %s.\n",
			len(unexplained), len(stale), plural(len(stale), "y", "ies"), RecordName)
		fmt.Fprintln(out, "An UNEXPLAINED survivor needs a test that kills it, or an entry in the record with a")
		fmt.Fprintln(out, "written why (equivalent mutants and platform-conditional branches are legitimate reasons).")
		fmt.Fprintln(out, "A STALE entry explains a survivor that no longer exists — delete it; if the line merely")
		fmt.Fprintln(out, "changed, the new text is a mutant nobody has judged yet, and it re-judges as itself.")
		return 1
	}
	fmt.Fprintf(out, "\nGATE PASS: %d survivor(s), every one explained in %s; no stale entries.\n", len(res.survived), RecordName)
	return 0
}

func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}
