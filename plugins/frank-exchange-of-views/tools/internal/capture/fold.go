package capture

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
)

// Folding capture's artifacts into the assembled run document.
//
// # The defect this replaces
//
// run.md's own blurb states what it holds: "how the machinery behaved: the friction the seats
// hit, the record's own invariant check, and what the run cost". Assembly supplies the first two.
// The third arrived through a function that sliced ONE table out of cost.md — and cost.md had
// five sections. The tier check, the board telemetry, the cost notes and the per-seat
// measurements were written, archived, and never read by anyone reading the run, because the only
// door into the document was cut to the width of one table.
//
// That is not a missing feature; it is an assembly that reports success having carried a fifth of
// what it was given. The artifact still existed on disk, so nothing looked broken.
//
// # Routed, and the route set is checked against what is actually emitted
//
// Each of cost.md's sections is routed to a section of run.md by NAME, and an UNROUTED SECTION IS
// AN ERROR RATHER THAN A SKIP. That is the whole point: the previous behaviour skipped four
// sections silently and read as working. TestEveryCostSectionIsRouted renders a real cost.md and
// fails if it contains a heading this table does not know, so adding a section to cost.md without
// giving it a home fails there rather than going quietly missing from every future run document.

// costRoutes maps a cost.md section heading to the run.md section it belongs under.
//
// THEY ARE NOT ALL "COST". A tier mismatch is a finding about whether the run was provisioned as
// configured, and board telemetry is the round-by-round trend the stopping judgment reads;
// filing either under a spending heading would put them where nobody looking for them would go.
// run.md is "how the machinery behaved", and these are three different things the machinery did.
var costRoutes = map[string]string{
	"Per seat-round":              "Cost",
	"Per seat (measured)":         "Cost",
	"Notes":                       "Cost",
	"Tier check":                  "Tier check",
	"Board telemetry (per round)": "Board telemetry",
	"Board telemetry":             "Board telemetry",
}

// foldOrder is the reading order of the sections this fold adds to the run document.
var foldOrder = []string{"Cost", "Tier check", "Board telemetry", "Integrity audits"}

// section is one `## heading` and the body beneath it.
type section struct {
	Heading string
	Body    string
}

// splitSections breaks a markdown document at its level-2 headings. Anything before the first one
// — the title and the provenance line — is preamble and is not returned: it describes the file it
// was written into, not the run, and carrying it would put "Measured from N transcripts in <dir>"
// into the middle of the assembled document.
func splitSections(md string) []section {
	var out []section
	lines := strings.Split(md, "\n")
	cur := section{Heading: ""}
	flush := func() {
		if cur.Heading != "" {
			cur.Body = strings.Trim(cur.Body, "\n")
			out = append(out, cur)
		}
	}
	for _, ln := range lines {
		if h, ok := strings.CutPrefix(ln, "## "); ok {
			flush()
			cur = section{Heading: strings.TrimSpace(h)}
			continue
		}
		if cur.Heading != "" {
			cur.Body += ln + "\n"
		}
	}
	flush()
	return out
}

// foldCaptureArtifacts folds cost.md's sections and the integrity audits into the run document.
//
// Returns the line capture prints, which always states WHAT WAS CARRIED. A fold that silently did
// less than it was asked is the failure being repaired here, so the count is in the message rather
// than inferable from the file.
//
// It is idempotent by the same test the previous version used — a document already carrying the
// heading is left alone — because capture is re-runnable and a second pass must not append the
// same sections again.
func foldCaptureArtifacts(reportPath, costPath, auditPath string) string {
	report, err := os.ReadFile(reportPath)
	if err != nil {
		return "" // no run document (a run that never reached assembly): nothing to fold into
	}
	body := strings.TrimRight(string(report), "\n")
	base := filepath.Base(reportPath)

	grouped, unrouted := routeCostSections(costPath)
	if len(unrouted) > 0 {
		// LOUD, because this is the exact failure being repaired: a section with no home used to
		// vanish and the fold reported success.
		return fmt.Sprintf("%s: fold REFUSED — cost.md carries section(s) with no route: %s. "+
			"Add them to costRoutes so they reach the run document.", base, strings.Join(unrouted, ", "))
	}
	if audit := readAuditBody(auditPath); audit != "" {
		grouped["Integrity audits"] = append(grouped["Integrity audits"], section{Body: audit})
	}
	if len(grouped) == 0 {
		return ""
	}

	// A READING ORDER, not a sort. What the run cost, then whether it was provisioned as
	// configured, then how the board moved, then whether the record itself verifies — narrowing
	// from the run to its evidence, the same way report.Files orders the document set. A map walk
	// would also reshuffle the document on every capture.
	var names []string
	for _, name := range foldOrder {
		if len(grouped[name]) > 0 {
			names = append(names, name)
		}
	}
	// ANYTHING NOT IN THE ORDER STILL SHIPS, alphabetically, rather than being dropped: a
	// destination added to costRoutes without being ordered is a formatting oversight, and losing
	// the section over it would be the very failure this function exists to end.
	var extra []string
	for name := range grouped {
		if !slices.Contains(names, name) {
			extra = append(extra, name)
		}
	}
	sort.Strings(extra)
	names = append(names, extra...)

	added := 0
	for _, name := range names {
		if strings.Contains(body, "\n## "+name+"\n") {
			continue // already folded by an earlier capture
		}
		var b strings.Builder
		fmt.Fprintf(&b, "\n\n## %s\n", name)
		for _, s := range grouped[name] {
			if s.Heading != "" {
				// DEMOTED, NOT DUPLICATED. A section pasted under its own name would give the
				// document two headings saying the same thing, one of them empty.
				fmt.Fprintf(&b, "\n### %s\n", s.Heading)
			}
			fmt.Fprintf(&b, "\n%s\n", s.Body)
		}
		body += strings.TrimRight(b.String(), "\n")
		added++
	}
	if added == 0 {
		return ""
	}
	if err := os.WriteFile(reportPath, []byte(body+"\n"), 0o644); err != nil {
		return base + ": fold FAILED — " + jsSlice(err.Error(), 200)
	}
	return fmt.Sprintf("%s: %d section(s) folded in (%s)", base, added, strings.Join(names, ", "))
}

// routeCostSections reads cost.md and groups its sections by their destination, reporting any it
// has no route for.
func routeCostSections(costPath string) (map[string][]section, []string) {
	grouped := map[string][]section{}
	costMd, err := os.ReadFile(costPath)
	if err != nil {
		return grouped, nil // no cost.md: nothing to route, and not a routing failure
	}
	var unrouted []string
	for _, s := range splitSections(string(costMd)) {
		dest, ok := costRoutes[s.Heading]
		if !ok {
			unrouted = append(unrouted, s.Heading)
			continue
		}
		grouped[dest] = append(grouped[dest], s)
	}
	return grouped, unrouted
}

// readAuditBody is run-record-audit.md's content, minus its title line.
//
// The nine integrity audits are the document's own "invariant check" promise, and they were
// written to a sibling file that the assembled run never referenced. A reader of the run had no
// way to learn whether its record verified.
func readAuditBody(auditPath string) string {
	b, err := os.ReadFile(auditPath)
	if err != nil {
		return ""
	}
	body := strings.TrimSpace(string(b))
	if rest, ok := strings.CutPrefix(body, "# "); ok {
		if i := strings.Index(rest, "\n"); i >= 0 {
			body = strings.TrimSpace(rest[i+1:])
		}
	}
	return body
}
