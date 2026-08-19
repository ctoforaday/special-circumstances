package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/gray-area/tools/internal/claims"
	"github.com/ctoforaday/special-circumstances/plugins/gray-area/tools/internal/prbody"
	"github.com/ctoforaday/special-circumstances/plugins/gray-area/tools/internal/trajectory"
)

// resolveTrace turns an optional transcript argument into a path, resolving this
// session's own from gray-area's manifest when none is given.
//
// Factored out of `checkpoint` so every inspection resolves the SAME way. Three
// verbs each rolling their own default is how one of them ends up globbing
// ~/.claude/projects/ — the thing plans/gray-area.md §3 rules out on principle,
// and the thing a reader would never notice because all three would still work.
func resolveTrace(arg string, stdout, stderr io.Writer, projectDir string) (string, bool) {
	if arg != "" {
		return arg, true
	}
	wd, _ := os.Getwd()
	root := defaultProjectDir(projectDir, os.Getenv("CLAUDE_PROJECT_DIR"), wd)
	resolved, err := claims.ResolveSession(claims.ManifestDir(root), claims.GlobFS, os.ReadFile, statReadable)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return "", false
	}
	if !resolved.Resolved {
		fmt.Fprintf(stderr, "gray-area: %s:%d names a transcript that is not readable (%s) — inspecting it would report an empty session, which is a lie about the session rather than about the tool\n",
			resolved.Manifest, resolved.Line, resolved.CaptureError)
		return "", false
	}
	fmt.Fprintf(stdout, "resolved this session's trajectory from %s:%d (session %s, captured %s)\n  -> %s\n\n",
		resolved.Manifest, resolved.Line, short(resolved.SessionID), resolved.CapturedAt, resolved.TranscriptPath)
	return resolved.TranscriptPath, true
}

// readTrace opens and parses a transcript, reporting coverage on stderr.
func readTrace(path string, stderr io.Writer, open func(string) (io.ReadCloser, error)) (trajectory.Result, bool) {
	f, err := open(path)
	if err != nil {
		fmt.Fprintf(stderr, "gray-area: cannot open %s: %v\n", path, err)
		return trajectory.Result{}, false
	}
	res, err := trajectory.Parse(f, path)
	_ = f.Close()
	if err != nil {
		fmt.Fprintf(stderr, "gray-area: reading %s: %v\n", path, err)
		return trajectory.Result{}, false
	}
	if res.BadLines > 0 {
		fmt.Fprintf(stderr, "gray-area: WARNING %d of %d lines did not parse (first at %v) — this inspection is over a SUBSET of %s\n",
			res.BadLines, res.Lines, res.BadSample, path)
	}
	return res, true
}

// coverage is the line every inspection prints, including — especially — the ones
// that found nothing.
//
// THE PLAUSIBLE-ZERO GATE, and it is a function rather than three format strings
// because it is the thing most likely to be dropped from one of them. "no rework
// found" reads identically whether the session was clean or whether the grouping
// keyed almost nothing; only this line separates the two, and a reader must not
// have to know that it is sometimes absent.
func coverage(g trajectory.GroupResult, what string) string {
	return fmt.Sprintf(
		"searched %d citable tool uses for %s: %d could not be keyed (a Read, or a tool this grouping does not model), %d acts happened exactly once",
		g.Considered, what, g.Unkeyed, g.Singles)
}

// reworkRow is one repeated act, as emitted.
type reworkRow struct {
	Kind        string             `json:"kind"`
	Target      string             `json:"target"`
	Count       int                `json:"count"`
	Consecutive int                `json:"consecutive"`
	Occurrences []trajectory.Event `json:"occurrences"`
}

// rework reports acts done more than once: the same file written again, the same
// command run again.
//
// IT DOES NOT SAY THIS WAS WASTE. A file written five times may be five careful
// increments; a suite run five times may be five different fixes verified. The
// tool reports the repetition and cites every occurrence, and the reader decides
// — the same division of labour NO-EVIDENCE rests on, applied where there is no
// claim to adjudicate at all.
func rework(tracePath string, stdout, stderr io.Writer, open func(string) (io.ReadCloser, error), asJSON bool, projectDir string, min int) int {
	path, ok := resolveTrace(tracePath, stdout, stderr, projectDir)
	if !ok {
		return 1
	}
	res, ok := readTrace(path, stderr, open)
	if !ok {
		return 1
	}
	g := trajectory.Group(res.Events)

	var rows []reworkRow
	for _, r := range g.Repeats {
		if r.Count() < min {
			continue
		}
		rows = append(rows, reworkRow{
			Kind: r.Kind, Target: r.Target, Count: r.Count(),
			Consecutive: r.Consecutive(), Occurrences: r.Occurrences,
		})
	}

	if asJSON {
		enc := json.NewEncoder(stdout)
		for _, r := range rows {
			if err := enc.Encode(r); err != nil {
				fmt.Fprintln(stderr, "gray-area: encoding row:", err)
				return 1
			}
		}
	} else {
		for _, r := range rows {
			fmt.Fprintf(stdout, "%-7s x%-3d (%d in a row)  %s\n", r.Kind, r.Count, r.Consecutive, truncate(r.Target, 80))
			for _, e := range r.Occurrences {
				fmt.Fprintf(stdout, "    %s:%d %s %s\n", e.File, e.Line, short(e.UUID), e.Timestamp)
			}
		}
	}
	fmt.Fprintf(stdout, "\n%d act(s) repeated %d+ times. %s\n", len(rows), min, coverage(g, "repetition"))
	return 0
}

// strikeLimit is [[anti-spinning]]'s rule, and the number is quoted from the rule
// rather than chosen here: "YOU MUST NOT attempt the same fix more than 3 times".
//
// It is a CONSECUTIVE count. The same check run three times across a session is
// ordinary iteration; three times in a row with nothing between them is the loop
// the rule is about. Counting it over the whole session would fire on every
// well-run day, and a rule that fires on everything is a rule nobody reads.
const strikeLimit = 3

// stallRow is one act that repeated back-to-back.
type stallRow struct {
	Kind        string             `json:"kind"`
	Target      string             `json:"target"`
	Consecutive int                `json:"consecutive"`
	Count       int                `json:"count"`
	Occurrences []trajectory.Event `json:"occurrences"`
	Note        string             `json:"note"`
}

// stalls adjudicates the 3-strike rule against what actually happened.
//
// WHY IT EXISTS ALONGSIDE THE HOOK COUNTER. prosthetic-conscience's
// PostToolUseFailure hook counts this deterministically and is documented as
// blind to most of it: it counts TOOL failures only, so a command that exits 0
// while the work still failed, an edit that applies cleanly and does not fix the
// bug, and a test runner reporting failures in its own output all pass it
// silently. The rule's own text says so — "the counter's silence is not evidence
// the loop was broken". This reads the trajectory instead, where a repetition is
// a repetition whatever the exit code was.
//
// IT CANNOT SEE WHETHER THE REPEAT WAS A RETRY. The trajectory records what was
// run, never what it said (see plans/gray-area-phase-2-build.md §2.1), so three
// identical invocations in a row might be three failures or three deliberate
// re-runs. That limit is printed on every row rather than buried, because a
// verdict that reads as "you were spinning" when the record cannot support it is
// exactly the false positive this phase was told to solve.
func stalls(tracePath string, stdout, stderr io.Writer, open func(string) (io.ReadCloser, error), asJSON bool, projectDir string) int {
	path, ok := resolveTrace(tracePath, stdout, stderr, projectDir)
	if !ok {
		return 1
	}
	res, ok := readTrace(path, stderr, open)
	if !ok {
		return 1
	}
	g := trajectory.Group(res.Events)

	const limit = "the record shows repetition, not failure: the trajectory holds invocations and not their output, so whether these were retries or deliberate re-runs is NOT MEASURED here"
	var rows []stallRow
	for _, r := range g.Repeats {
		c := r.Consecutive()
		if c < strikeLimit {
			continue
		}
		rows = append(rows, stallRow{
			Kind: r.Kind, Target: r.Target, Consecutive: c, Count: r.Count(),
			Occurrences: r.Occurrences, Note: limit,
		})
	}

	if asJSON {
		enc := json.NewEncoder(stdout)
		for _, r := range rows {
			if err := enc.Encode(r); err != nil {
				fmt.Fprintln(stderr, "gray-area: encoding row:", err)
				return 1
			}
		}
	} else {
		// STATED ONCE, ABOVE THE ROWS, and only when there are rows. Repeating it
		// per row made a real listing unreadable — twenty identical paragraphs
		// between the findings — and a limit nobody reads is a limit that is not
		// stated. It stays on EVERY JSON row, because a consumer reads one row at a
		// time and has no header to have seen.
		if len(rows) > 0 {
			fmt.Fprintf(stdout, "not measured: %s\n\n", limit)
		}
		for _, r := range rows {
			fmt.Fprintf(stdout, "%d IN A ROW  %-7s %s\n", r.Consecutive, r.Kind, truncate(r.Target, 80))
			for _, e := range r.Occurrences {
				fmt.Fprintf(stdout, "    %s:%d %s %s\n", e.File, e.Line, short(e.UUID), e.Timestamp)
			}
		}
	}
	fmt.Fprintf(stdout, "\n%d act(s) repeated %d+ times CONSECUTIVELY (the [[anti-spinning]] limit). %s\n",
		len(rows), strikeLimit, coverage(g, "consecutive repetition"))
	return 0
}

// prBody adjudicates a pull request body's claims against the trajectory.
//
// THE RULING (plans/gray-area-phase-2.md §0): full post-hoc adjudication. A body
// stops being testimony and becomes a record under inspection.
//
// THE BOUNDARY (spec §2, and it is the whole design): the trajectory records what
// was RUN and not what the run SAID, so a claim's ACT is adjudicable and its
// RESULT is not. Every row whose claim asserted an outcome carries that outcome
// in Unmeasured, in BOTH output shapes, so a CITED verdict cannot be read as
// endorsing a number nothing checked.
//
// EXIT IS 0 EVEN WITH FINDINGS, unlike `checkpoint`. A checkpoint's claims are
// this session's own gate and a stale one is a fault to fix now; a pull request
// body is adjudicated AFTER the fact, for a human to read. Failing the process on
// a NO-EVIDENCE row would turn "the tokens did not match" into "the pull request
// is wrong", which is the conviction this verdict was named not to make.
func prBody(bodyPath, tracePath string, stdout, stderr io.Writer, open func(string) (io.ReadCloser, error), asJSON bool, projectDir string) int {
	bf, err := open(bodyPath)
	if err != nil {
		fmt.Fprintf(stderr, "gray-area: cannot open %s: %v\n", bodyPath, err)
		return 1
	}
	declared, err := prbody.Parse(bf, bodyPath)
	_ = bf.Close()
	if err != nil {
		fmt.Fprintf(stderr, "gray-area: reading %s: %v\n", bodyPath, err)
		return 1
	}
	if len(declared) == 0 {
		// NOT a silent zero. A body with no backticked commands has made no
		// adjudicable claim, and that is a different fact from a body whose claims
		// all checked out — a reader must not have to guess which happened.
		fmt.Fprintf(stdout, "%s carries no adjudicable claim: no code span in it reads as a command. A body may be entirely accurate and still say nothing this record can check.\n", bodyPath)
		return 0
	}

	path, ok := resolveTrace(tracePath, stdout, stderr, projectDir)
	if !ok {
		return 1
	}
	res, ok := readTrace(path, stderr, open)
	if !ok {
		return 1
	}

	plain := make([]claims.Claim, 0, len(declared))
	for _, c := range declared {
		plain = append(plain, c.Claim)
	}
	findings := claims.Adjudicate(plain, res)

	// Carry the unmeasurable half onto the row it belongs to, keyed by LINE AND
	// COMMAND. Line alone is wrong and was measured wrong: a body line holding
	// three claims gave all three the last one's outcome, so a caveat named a
	// number that belonged to a different check.
	byClaim := map[string]string{}
	for _, c := range declared {
		if c.Result != "" {
			byClaim[fmt.Sprintf("%d\x00%s", c.Claim.Line, c.Claim.Cmd)] = c.Result
		}
	}
	for i := range findings {
		if r, ok := byClaim[fmt.Sprintf("%d\x00%s", findings[i].NoteLine, findings[i].Index)]; ok {
			findings[i].Unmeasured = append(findings[i].Unmeasured,
				"the asserted outcome ("+strings.TrimSpace(r)+") — the trajectory records invocations, not their output")
		}
	}

	if asJSON {
		enc := json.NewEncoder(stdout)
		for _, f := range findings {
			if err := enc.Encode(f); err != nil {
				fmt.Fprintln(stderr, "gray-area: encoding finding:", err)
				return 1
			}
		}
	} else {
		for _, f := range findings {
			fmt.Fprintf(stdout, "%-11s %s:%d  %s\n", f.Verdict, f.NoteFile, f.NoteLine, truncate(f.Claim, 90))
			if f.UUID != "" {
				fmt.Fprintf(stdout, "    evidence: %s:%d %s %s  %s\n", f.File, f.Line, short(f.UUID), f.At, truncate(f.Command, 70))
			}
			if len(f.Searched) > 0 {
				fmt.Fprintf(stdout, "    searched: %v across %d citable events\n", f.Searched, f.EventsSeen)
			}
			// THE ACT/RESULT GATE. In the table too, not only in the JSON — a human
			// reading this is exactly who the false positive would mislead.
			for _, u := range f.Unmeasured {
				fmt.Fprintf(stdout, "    NOT MEASURED: %s\n", u)
			}
			if f.Note != "" {
				fmt.Fprintf(stdout, "    note: %s\n", f.Note)
			}
		}
	}
	fmt.Fprintf(stdout, "\n%d claim(s) adjudicated from %s. A NO-EVIDENCE row is an ABSENCE, not a finding that the body is wrong; exit is 0 either way.\n",
		len(findings), bodyPath)
	return 0
}

// coverageVerb answers the question every seat-scoped inspection must ask before
// printing a number: does the manifest name every seat transcript that exists?
//
// It exits 1 when it could NOT measure, and 0 when it measured — whatever it
// found. That split is deliberate. Unnamed transcripts are a finding for a human;
// an unmeasurable board is a broken instrument, and an instrument that reports a
// clean board when it cannot see is the failure this whole plugin is about.
func coverageVerb(stdout, stderr io.Writer, projectDir string) int {
	wd, _ := os.Getwd()
	root := defaultProjectDir(projectDir, os.Getenv("CLAUDE_PROJECT_DIR"), wd)
	resolved, err := claims.ResolveSession(claims.ManifestDir(root), claims.GlobFS, os.ReadFile, statReadable)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	raw, err := os.ReadFile(resolved.Manifest)
	if err != nil {
		fmt.Fprintf(stderr, "gray-area: cannot read %s: %v\n", resolved.Manifest, err)
		return 1
	}
	seats, unclassifiable := claims.SeatIDs(raw)
	c := claims.Reconcile(resolved.TranscriptPath, seats, os.ReadDir)

	fmt.Fprintf(stdout, "manifest: %s\n", resolved.Manifest)
	fmt.Fprintf(stdout, "seat directory: %s\n\n", c.Dir)
	if unclassifiable > 0 {
		fmt.Fprintf(stdout, "%d seat row(s) carry no agent_id and could be reconciled with nothing — counted here rather than dropped\n", unclassifiable)
	}
	if !c.Measured {
		// NOT a zero. The two states this separates produce the same empty lists.
		fmt.Fprintf(stdout, "NOT MEASURED — %s\n", c.Why)
		return 1
	}
	fmt.Fprintf(stdout, "%d seat row(s) named, %d transcript(s) on disk\n", len(c.Named), len(c.OnDisk))
	if c.Why != "" {
		fmt.Fprintf(stdout, "%s\n", c.Why)
	}
	// CAPPED, AND THE REMAINDER IS STATED. A silent truncation reads as "that was
	// all of them"; the first real run printed 146 lines and buried its own summary.
	const listCap = 20
	listed := func(label, why string, ids []string) {
		for i, id := range ids {
			if i == listCap {
				fmt.Fprintf(stdout, "  ... and %d more %s (not listed here)\n", len(ids)-listCap, label)
				break
			}
			fmt.Fprintf(stdout, "  %-9s %s — %s\n", label, id, why)
		}
	}
	listed("UNNAMED", "on disk, and no manifest row accounts for it", c.Unnamed)
	listed("MISSING", "a seat row names it and the file is not there", c.Missing)
	if len(c.Unnamed) == 0 && len(c.Missing) == 0 && len(c.Named) > 0 {
		fmt.Fprintf(stdout, "  reconciled: every seat row names a transcript that exists, and every transcript is named\n")
	}
	if len(c.Unnamed) > 0 {
		fmt.Fprintf(stdout, "\nAn unnamed transcript is a seat NOTHING in the manifest can lead a reader to. "+
			"Two benign explanations exist and neither is measured here: the subagent may have been killed before "+
			"SubagentStop fired, or it may not have been spawned through the Agent tool at all. Until one of them is "+
			"established, seat coverage cannot be stated as a number (#469).\n")
	}
	return 0
}
