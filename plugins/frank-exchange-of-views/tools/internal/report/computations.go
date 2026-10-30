package report

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/compute"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// THE INVISIBLE COMPUTATION LAYER, MADE VISIBLE — the way weaveCitations does it for sources, and it
// exists because without it the axis was a half-state that read as done.
//
// MEASURED ON ITS OWN FIRST BUILD: a computation ran, cached, anchored and was auditable by red,
// and the assembled report carried the raw anchor token while mentioning the
// computation ZERO times. The evidence existed everywhere except the document a human reads.

// computationNumbers assigns each recorded computation its RUN-WIDE number, in record order.
//
// The number is not per-document, and it is not first-appearance. Across a multi-file set,
// first-appearance numbering makes "K3" mean a different computation in the report than in the
// debate, which is worse than no number at all. Record order is the one ordering every document
// agrees on because none of them produced it.
func computationNumbers(computations []record.Computation) (map[string]int, map[string]record.Computation) {
	num := map[string]int{}
	byLabel := map[string]record.Computation{}
	for i, p := range computations {
		num[p.Label] = i + 1
		byLabel[p.Label] = p
	}
	return num, byLabel
}

// weaveComputationRefs resolves the invisible computation layer in ONE document: every computation
// anchor becomes a visible [^KN] reference, and the definitions for exactly the computations that
// document anchors are appended. It returns the labels it hit so the caller can compose the
// evidence document from the union.
//
// THE DEFINITION IS THE HALF THAT WAS MISSING (#590). The first cut rewrote the anchors to
// numbered references and then wrote the appendix headings as numbered references too, which markdown reads as a SECOND
// dangling reference rather than a heading — so every computation footnote in the shipped report
// pointed at nothing, and a human repaired it by hand after the fact. A reference with no
// definition is not a citation; it is a broken link that renders as one.
func weaveComputationRefs(md string, computations []record.Computation) (string, []string) {
	num, byLabel := computationNumbers(computations)
	var order []string
	seen := map[string]bool{}
	extra := len(computations)
	body := anchor.Replace(md, func(tok, label string) string {
		if anchor.AssemblyOf(label) != anchor.WeaveComputation {
			return tok
		}
		if !seen[label] {
			seen[label] = true
			order = append(order, label)
			if _, ok := num[label]; !ok {
				// Anchored, and not on the record. It gets a number and a definition that says
				// so — a silent drop here is indistinguishable from a document with no computations.
				extra++
				num[label] = extra
			}
		}
		return computationNote(num[label])
	})
	if len(order) == 0 {
		return body, nil
	}
	sort.SliceStable(order, func(i, j int) bool { return num[order[i]] < num[order[j]] })

	var b strings.Builder
	b.WriteString(strings.TrimRight(body, "\n"))
	b.WriteString("\n\n")
	for _, label := range order {
		p, ok := byLabel[label]
		if !ok {
			fmt.Fprintf(&b, "%s: _(unresolved computation %s — no computation event on the record)_\n", computationNote(num[label]), label)
			continue
		}
		fmt.Fprintf(&b, "%s: %s — a computation, not a claim: program, output and sha256 in [evidence.md](%s#%s).\n",
			computationNote(num[label]), escapeNotesInLine(firstNonEmpty(plain(p.Reason), plain(commandOf(p)))), FileEvidence, label)
	}
	return b.String(), order
}

// evidenceDoc is the run's computations in full — the document the computation footnotes point at.
//
// A COMPUTATION SHOWS ITS WORK. A citation can be a one-line reference because the source is
// elsewhere and permanent; a computation's program IS the evidence, so this carries the bundle and
// the output verbatim. A reader who does not trust it can run it, and that is the whole point of
// preferring it to prose.
//
// The bundle and output are read from the STORE under <run>/computations/<bundle_sha>/, not from the
// record: the report must show the exact bytes that ran, not a copy the record made of them. Each
// act shows its OWN recorded output (acts/<computation_id>/output). A missing artifact is stated,
// never silently skipped — a section quietly short of a computation is worse than one that says the
// artifact is gone.
func evidenceDoc(run record.Run, computations []record.Computation, anchored map[string]bool) string {
	if len(computations) == 0 {
		return ""
	}
	num, _ := computationNumbers(computations)

	var b strings.Builder
	b.WriteString("## " + evidenceHeading + "\n\n")
	// SUBJECT VOICE (gblock, 2026-10-19): the preamble describes the computations, never the run, its
	// seats or its process — m30's voice lens and blue both logged "a proof this run ran".
	b.WriteString("Computations that check claims in this report, with their code and output.\n")
	b.WriteString("`reproducible` means the program wrote the same output and exit when run twice;\n")
	b.WriteString("`variant` means the two runs differed, and the computation states what holds across them\n")
	b.WriteString("and what moved. Each program is stored under `computations/<bundle_sha>/`, so any reader\n")
	b.WriteString("can run it again.\n\n")

	for _, p := range computations {
		label := p.Label
		n := num[label]
		b.WriteString(computationAnchorMark(label) + "\n\n")
		fmt.Fprintf(&b, "### K%d — %s\n\n", n, firstNonEmpty(plain(p.Reason), commandOf(p)))
		if !anchored[label] {
			b.WriteString("- **anchored to nothing**: this computation is on the record and no document in this set references it\n")
		}
		fmt.Fprintf(&b, "- **basis**: %s\n", p.Basis)
		if p.Invariant != "" {
			fmt.Fprintf(&b, "- **holds across both runs**: %s\n", plain(p.Invariant))
		}
		fmt.Fprintf(&b, "- **command**: %s (bundle sha256 %s)\n", codeSpanOf(commandOf(p)), codeSpanOf(shortSHA(p.SHA)))
		fmt.Fprintf(&b, "- **exit**: %d\n", p.Exit)
		if p.Drift != "" {
			b.WriteString("- **moved between the two runs**:\n\n")
			b.WriteString(itemBlock(fenced("", strings.TrimRight(p.Drift, "\n"))) + "\n")
		}
		// RED'S INDEPENDENT AUDIT (#343), IN ITS TWO HALVES. The strongest audit the engine has is
		// worth the reader knowing about, and its absence is worth knowing about too: a computation
		// nobody re-ran and a computation that reproduced used to render identically.
		//
		// The RE-RUN belongs to the bundle, and every computation on the same bundle shares it; the
		// JUDGEMENT belongs to this computation's sentence. The dangerous cell is
		// reproduces-but-unsound: it re-runs clean and establishes nothing, which is what a
		// computation looks like at its most credible and least useful.
		switch r := p.Rerun; {
		case r == nil:
			b.WriteString("- **not re-run** — no checker executed this bundle; the computation stands on blue's execution alone\n")
		case r.Rerun == recordpb.Rerun_RERUN_REPRODUCED:
			b.WriteString("- **re-run by red**: it REPRODUCES — the stored program, run again, wrote the recorded output byte for byte\n")
		case r.Rerun == recordpb.Rerun_RERUN_CALLS_THE_TOOL:
			b.WriteString("- **re-run by red — IT CANNOT BE CHECKED.** The program calls the debate's own tool, which refuses every call made inside a re-run, so no reader can run it to its recorded output\n")
		default:
			b.WriteString("- **re-run by red**: it DID NOT REPRODUCE — the stored program no longer writes the recorded output\n")
			if r.Drift != "" {
				b.WriteString("\n" + itemBlock(fenced("", strings.TrimRight(r.Drift, "\n"))) + "\n")
			}
		}
		var notes []string
		for _, j := range p.Judgements {
			if j.Note != "" && (p.Soundness != record.SoundnessUnsound || !j.Sound) {
				notes = append(notes, j.Note)
			}
		}
		note := ""
		if len(notes) > 0 {
			note = " — " + judgementNotes(notes)
		}
		switch p.Soundness {
		case record.SoundnessSound:
			b.WriteString("- **judged by red**: red read the program and accepts that it establishes this sentence" + note + "\n")
		case record.SoundnessUnsound:
			b.WriteString("- **judged by red — IT DOES NOT PROVE THIS SENTENCE.** Red read the program and found it does not establish what this sentence says. Re-running measures determinism, not validity" + note + "\n")
		case record.SoundnessReopened:
			b.WriteString("- **not judged since the sentence changed** — red's verdict was of an earlier wording of this sentence\n")
		default:
			b.WriteString("- **not judged** — nobody has read this program against this sentence\n")
		}
		if p.Cites != "" {
			// The METHOD is cited and the INSTANCE is computed: naming the citation here is
			// what makes the pair legible as one argument rather than two artifacts.
			fmt.Fprintf(&b, "- **applies the method cited at**: %s\n", p.Cites)
		}
		b.WriteString("\n")

		files, output := readArtifact(run, p)
		for _, f := range files {
			b.WriteString(f + "\n")
		}
		b.WriteString("Output:\n\n")
		b.WriteString(fenced("", strings.TrimRight(output, "\n")) + "\n")
	}
	return strings.TrimRight(b.String(), "\n") + "\n"
}

// itemBlock is a block that belongs to the list item above it: every line of it indented under the
// item, its first included, so a fence opens inside the item rather than at the margin.
func itemBlock(s string) string {
	return "  " + itemLines(strings.TrimRight(s, "\n"), "  ") + "\n"
}

// commandOf is the command a computation ran, as one line.
func commandOf(p record.Computation) string {
	return strings.Join(p.Argv, " ")
}

// fenced is content in a fenced code block no line of it can close or lengthen into prose: a fence of
// backticks one longer than the longest run of them in the content, and never fewer than three. A
// fence of three closed at an output's own "```" line, and the rest of the output shipped as prose.
func fenced(lang, content string) string {
	f := strings.Repeat("`", max(3, longestRun(content, '`')+1))
	return f + lang + "\n" + content + "\n" + f + "\n"
}

// codeSpanOf is s as a code span that shows it exactly: delimited by backticks one longer than the
// longest run of them in s, and padded with a space where s would otherwise touch a delimiter with a
// backtick or lose an edge space to the span's stripping. An empty s has no code to show and is
// said to be empty, as the documents mark a value the record does not hold: two backticks with
// nothing between them are no code span, and print as typed.
func codeSpanOf(s string) string {
	if s == "" {
		return "_(empty)_"
	}
	d := strings.Repeat("`", longestRun(s, '`')+1)
	if strings.HasPrefix(s, "`") || strings.HasSuffix(s, "`") || len(s) > 1 && s[0] == ' ' && s[len(s)-1] == ' ' && strings.TrimSpace(s) != "" {
		s = " " + s + " "
	}
	return d + s + d
}

// longestRun is the length of the longest run of c in s.
func longestRun(s string, c byte) int {
	most, n := 0, 0
	for i := 0; i < len(s); i++ {
		if s[i] != c {
			n = 0
			continue
		}
		n++
		most = max(most, n)
	}
	return most
}

// judgementNotes is red's notes on a computation, written after the judgement on its list item's line, each
// without its tokens as it reads where it ships. Notes of one line each are joined on that line, so
// each is read behind the judgement (stripTokensAfter). Where one runs to several lines each note is
// the item's own paragraph, its fences closed where it ends (ownBlock, reading the first as text after
// the tool's, and each after it as a block of its own, as its tokens are read: stripTokensAfter, then
// StripTokens), and the whole the rest of the item (itemLines).
func judgementNotes(notes []string) string {
	multi := false
	inline := make([]string, len(notes))
	for i, n := range notes {
		inline[i] = strings.TrimSpace(lineEnds.Replace(stripTokensAfter(n)))
		multi = multi || strings.Contains(inline[i], "\n")
	}
	if !multi {
		return strings.Join(inline, " / ")
	}
	notes[0] = inline[0]
	for i := 1; i < len(notes); i++ {
		notes[i] = strings.TrimSpace(lineEnds.Replace(StripTokens(notes[i])))
	}
	notes[0] = ownBlock("x" + notes[0])[1:]
	for i := 1; i < len(notes); i++ {
		notes[i] = ownBlock(notes[i])
	}
	return itemLines(strings.Join(notes, "\n\n"), "  ")
}

// oneLine is s as one line of markdown, its runs of whitespace one space: the form of every value the
// tool splices into a line it composes — a heading, a list item, a note, a table cell, a labelled
// line. A value of several lines spliced raw leaves the line: a fence on its second line opens code
// that runs to the end of the document and takes every heading, note and anchor after it, and a `#`
// line is a heading the tool did not write. On one line, behind the tool's own text, a value opens and
// closes no block. strings.Fields splits at every line ending CommonMark has, a lone `\r` included.
func oneLine(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// readArtifact returns the bundle's files, each as the block the document prints, and this act's
// output — or an explicit note in place of what is missing. Silence here would let a report claim a
// computation it cannot show. A file that is not text — not UTF-8, or holding a NUL — is one line
// naming its path, size and sha256 rather than bytes a fence would mangle.
func readArtifact(run record.Run, p record.Computation) (files []string, output string) {
	at := compute.StoreDir + "/" + p.SHA + "/"
	stored, err := compute.Bundle(run.Dir(), p.SHA)
	if err != nil {
		files = []string{"(the bundle is missing from " + at + "dir/)\n"}
	}
	for _, f := range stored {
		if !utf8.Valid(f.Body) || bytes.IndexByte(f.Body, 0) >= 0 {
			sum := sha256.Sum256(f.Body)
			files = append(files, fmt.Sprintf("- %s: %d bytes, sha256 %s\n", codeSpanOf(f.Path), len(f.Body), hex.EncodeToString(sum[:])))
			continue
		}
		files = append(files, codeSpanOf(f.Path)+":\n\n"+fenced(fenceLang(f.Path), strings.TrimRight(string(f.Body), "\n")))
	}
	if b, rerr := compute.ActOutput(run.Dir(), p.SHA, p.Label); rerr == nil {
		output = string(b)
	} else {
		output = "(the output is missing from " + at + "acts/" + p.Label + "/)"
	}
	return files, output
}

// fenceLang picks the code-fence language from a file's extension so the report highlights it the
// way the seat wrote it.
func fenceLang(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".py":
		return "python"
	case ".js", ".mjs":
		return "javascript"
	case ".sh":
		return "bash"
	case ".go":
		return "go"
	default:
		return ""
	}
}

func shortSHA(s string) string {
	if len(s) > 12 {
		return s[:12]
	}
	return s
}

func firstNonEmpty(xs ...string) string {
	for _, x := range xs {
		if strings.TrimSpace(x) != "" {
			return x
		}
	}
	return "(untitled computation)"
}
