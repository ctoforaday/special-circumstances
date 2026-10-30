package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/compute"
)

// The outcomes a bundle's act is graded with: compute.Grade's, and refused.
const (
	Reproduced = string(compute.Reproduced)
	Moved      = string(compute.Moved)
	Differs    = string(compute.Differs)
	Refused    = "refused"
)

// Row is one archived bundle's measured outcome: the outcome its acts agree on, or each distinct
// one, and each act's own.
type Row struct {
	Archive, OldSha, Outcome string
	// Acts is each act's outcome and basis, as measured; the outcomes file holds only Outcome.
	Acts []Act
}

// Act is one act's measured outcome, and whether it was recorded variant.
type Act struct {
	Outcome string
	Variant bool
}

func (r Row) key() string { return r.Archive + "\t" + r.OldSha }

// joinOutcomes is the outcome a bundle's acts agree on, or each distinct one, sorted and
// comma-separated.
func joinOutcomes(set map[string]bool) string {
	var out []string
	for o := range set {
		out = append(out, o)
	}
	sort.Strings(out)
	return strings.Join(out, ",")
}

// The file's shape: comment lines, an interpreter line per command, the column header, the rows.
const (
	interpreterPrefix = "# interpreter\t"
	columns           = "archive\told_sha\toutcome"
)

// Format is the outcomes file: a header naming each interpreter's version, then one row per bundle,
// sorted.
func Format(versions map[string]string, rows []Row) string {
	var b strings.Builder
	b.WriteString("# Measured by `go run ./devcmd/archiverun -write` in the archive-rerun job of\n")
	b.WriteString("# .github/workflows/hooks.yml, under the interpreters named below. Never edited by hand.\n")
	var names []string
	for n := range versions {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		fmt.Fprintf(&b, "%s%s\t%s\n", interpreterPrefix, n, versions[n])
	}
	b.WriteString(columns + "\n")
	sorted := append([]Row{}, rows...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].key() < sorted[j].key() })
	for _, r := range sorted {
		fmt.Fprintf(&b, "%s\t%s\t%s\n", r.Archive, r.OldSha, r.Outcome)
	}
	return b.String()
}

// Compare holds measured rows against a held outcomes file and returns every disagreement: an
// interpreter whose version is not the header's, a file with no rows, a row the file lacks, a row
// the measurement lacks, and an outcome that moved. A file that measured nothing is a failure,
// never a pass over zero rows.
func Compare(held string, versions map[string]string, rows []Row) []string {
	var problems []string
	heldVersions := map[string]string{}
	heldRows := map[string]string{}
	sawColumns := false
	for i, line := range strings.Split(held, "\n") {
		switch {
		case line == "":
		case strings.HasPrefix(line, interpreterPrefix):
			parts := strings.SplitN(strings.TrimPrefix(line, interpreterPrefix), "\t", 2)
			if len(parts) != 2 {
				problems = append(problems, fmt.Sprintf("line %d: an interpreter line names no version: %q", i+1, line))
				continue
			}
			heldVersions[parts[0]] = parts[1]
		case strings.HasPrefix(line, "#"):
		case line == columns:
			sawColumns = true
		default:
			parts := strings.Split(line, "\t")
			if len(parts) != 3 || !sawColumns {
				problems = append(problems, fmt.Sprintf("line %d: not a row under %q: %q", i+1, columns, line))
				continue
			}
			heldRows[parts[0]+"\t"+parts[1]] = parts[2]
		}
	}
	if len(heldRows) == 0 {
		return append(problems, "NOT MEASURED: the outcomes file holds no row. The outcomes are measured on the pinned image: dispatch the archive-rerun job, download its outcomes artifact and commit it")
	}
	for n, v := range versions {
		if heldVersions[n] != v {
			problems = append(problems, fmt.Sprintf("interpreter %s is %q here and %q in the file: the outcomes were measured under another interpreter, and no row is compared", n, v, heldVersions[n]))
		}
	}
	for n := range heldVersions {
		if _, ok := versions[n]; !ok {
			problems = append(problems, fmt.Sprintf("the file names interpreter %s, which this check does not ask for", n))
		}
	}
	if len(problems) > 0 {
		return problems
	}
	seen := map[string]bool{}
	for _, r := range rows {
		seen[r.key()] = true
		want, ok := heldRows[r.key()]
		switch {
		case !ok:
			problems = append(problems, fmt.Sprintf("%s %s: re-ran %s, and the file holds no row for it", r.Archive, r.OldSha, r.Outcome))
		case !satisfies(want, r):
			problems = append(problems, fmt.Sprintf("%s %s: re-ran %s, the file holds %s", r.Archive, r.OldSha, r.Outcome, want))
		}
	}
	var gone []string
	for k := range heldRows {
		if !seen[k] {
			gone = append(gone, strings.ReplaceAll(k, "\t", " "))
		}
	}
	sort.Strings(gone)
	for _, k := range gone {
		problems = append(problems, fmt.Sprintf("%s: the file holds a row no archive re-ran", k))
	}
	return problems
}

// satisfies is whether a measured row stands for the held outcome. A REPRODUCIBLE act's outcome is
// held exactly. A VARIANT act's re-run may write its recorded output or not — the act was recorded
// variant because its runs already differed — so either of its non-failing outcomes, reproduced or
// moved, stands for the other: the held set is satisfied when the fixed outcomes are all in it, and
// the variant acts can land on the rest of it.
func satisfies(want string, r Row) bool {
	if want == r.Outcome {
		return true
	}
	held := map[string]bool{}
	for _, o := range strings.Split(want, ",") {
		held[o] = true
	}
	flexible := func(o string) bool { return o == Reproduced || o == Moved }
	fixed, free := map[string]bool{}, 0
	for _, a := range r.Acts {
		if a.Variant && flexible(a.Outcome) {
			free++
		} else {
			fixed[a.Outcome] = true
		}
	}
	for o := range fixed {
		if !held[o] {
			return false
		}
	}
	rest := 0
	for o := range held {
		if fixed[o] {
			continue
		}
		if !flexible(o) {
			return false
		}
		rest++
	}
	if rest > free {
		return false
	}
	return free == 0 || rest > 0 || held[Reproduced] || held[Moved]
}

// summary counts the bundles by outcome.
func summary(rows []Row) string {
	n := map[string]int{}
	for _, r := range rows {
		n[r.Outcome]++
	}
	var parts []string
	for o, c := range n {
		parts = append(parts, fmt.Sprintf("%d %s", c, o))
	}
	sort.Strings(parts)
	return fmt.Sprintf("re-ran %d bundle(s): %s", len(rows), strings.Join(parts, ", "))
}
