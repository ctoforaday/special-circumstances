package report

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// proofAnchor matches an invisible proof anchor "<!--proof:p-<id>-->" — the tool-inserted
// token marking a sentence a COMPUTATION backs.
var proofAnchor = regexp.MustCompile(`<!--proof:(p-[0-9a-f]+)-->`)

// THE INVISIBLE PROOF LAYER, MADE VISIBLE — the way weaveCitations does it for sources, and it
// exists because without it the axis was a half-state that read as done.
//
// MEASURED ON ITS OWN FIRST BUILD: a proof ran, cached, anchored and was auditable by red,
// and the assembled report carried the raw "<!--proof:p-…-->" token while mentioning the
// computation ZERO times. The evidence existed everywhere except the document a human reads.

// proofNumbers assigns each recorded proof its RUN-WIDE number, in record order.
//
// The number is not per-document, and it is not first-appearance. Across a seven-file set,
// first-appearance numbering makes "P3" mean a different computation in the report than in the
// debate, which is worse than no number at all. Record order is the one ordering every document
// agrees on because none of them produced it.
func proofNumbers(proofs []record.Proof) (map[string]int, map[string]record.Proof) {
	num := map[string]int{}
	byLabel := map[string]record.Proof{}
	for i, p := range proofs {
		num[p.Label] = i + 1
		byLabel[p.Label] = p
	}
	return num, byLabel
}

// weaveProofRefs resolves the invisible proof layer in ONE document: every "<!--proof:p-…-->"
// anchor becomes a visible [^PN] reference, and the definitions for exactly the proofs that
// document anchors are appended. It returns the labels it hit so the caller can compose the
// evidence document from the union.
//
// THE DEFINITION IS THE HALF THAT WAS MISSING (#590). The first cut rewrote the anchors to
// [^PN] and then wrote the appendix headings as "### [^PN] …", which markdown reads as a SECOND
// dangling reference rather than a heading — so every proof footnote in the shipped report
// pointed at nothing, and a human repaired it by hand after the fact. A reference with no
// definition is not a citation; it is a broken link that renders as one.
func weaveProofRefs(md string, proofs []record.Proof) (string, []string) {
	num, byLabel := proofNumbers(proofs)
	var order []string
	seen := map[string]bool{}
	extra := len(proofs)
	body := proofAnchor.ReplaceAllStringFunc(md, func(tok string) string {
		label := proofAnchor.FindStringSubmatch(tok)[1]
		if !seen[label] {
			seen[label] = true
			order = append(order, label)
			if _, ok := num[label]; !ok {
				// Anchored, and not on the record. It gets a number and a definition that says
				// so — a silent drop here is indistinguishable from a document with no proofs.
				extra++
				num[label] = extra
			}
		}
		return fmt.Sprintf("[^P%d]", num[label])
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
			fmt.Fprintf(&b, "[^P%d]: _(unresolved proof %s — no proof event on the record)_\n", num[label], label)
			continue
		}
		fmt.Fprintf(&b, "[^P%d]: %s — a computation, not a claim: script, output and sha256 in [evidence.md](%s#%s).\n",
			num[label], strings.TrimSpace(oneLine(firstNonEmpty(p.Reason, p.Script))), FileEvidence, label)
	}
	return b.String(), order
}

// evidenceDoc is the run's computations in full — the document the proof footnotes point at.
//
// A PROOF SHOWS ITS WORK. A citation can be a one-line reference because the source is
// elsewhere and permanent; a computation's source IS the evidence, so this carries the script
// and the output verbatim. A reader who does not trust it can run it, and that is the whole
// point of preferring it to prose.
//
// The script and output are read from the ARTIFACT under <run>/proofs/<sha256>/, not from the
// record: the report must show the exact bytes that ran, not a copy the record made of them.
// A missing artifact is stated, never silently skipped — a proof section quietly short of a
// proof is worse than one that says the artifact is gone.
func evidenceDoc(run record.Run, proofs []record.Proof, anchored map[string]bool) string {
	if len(proofs) == 0 {
		return ""
	}
	num, _ := proofNumbers(proofs)

	var b strings.Builder
	b.WriteString("## Proofs\n\n")
	b.WriteString("Each of these is a computation this run ran, with the exact script and its output.\n")
	b.WriteString("`reproducible` means the same script produced identical output on two consecutive runs;\n")
	b.WriteString("`observed` means the output moved between them, which makes it a measurement of a system\n")
	b.WriteString("in motion rather than a proof. Every script is cached under `proofs/<sha256>/` in the run\n")
	b.WriteString("directory, so any reader can run it again.\n\n")

	for _, p := range proofs {
		label := p.Label
		n := num[label]
		// An explicit anchor, because a heading's generated id is a function of its whole text
		// and the footnotes in six other files link here by LABEL, which is stable.
		fmt.Fprintf(&b, "<a id=\"%s\"></a>\n\n", label)
		fmt.Fprintf(&b, "### P%d — %s\n\n", n, strings.TrimSpace(firstNonEmpty(p.Reason, p.Script)))
		if !anchored[label] {
			b.WriteString("- **anchored to nothing**: this computation is on the record and no document in this set references it\n")
		}
		fmt.Fprintf(&b, "- **basis**: %s", p.Basis)
		if p.Drift != "" {
			fmt.Fprintf(&b, " — %s", p.Drift)
		}
		b.WriteString("\n")
		fmt.Fprintf(&b, "- **script**: `%s` (sha256 `%s`)\n", p.Script, shortSHA(p.SHA))
		fmt.Fprintf(&b, "- **exit**: %d\n", p.Exit)
		// RED'S INDEPENDENT RE-RUN (#343). The strongest audit the engine has is worth the
		// reader knowing about, and its absence is worth knowing about too: a proof nobody
		// re-ran and a proof that reproduced used to render identically.
		if v := p.Verified; v != nil {
			// TWO AXES. Reproducing measures determinism; soundness is red's judgement from
			// reading the script. The dangerous cell is reproduces-but-unsound: it re-runs
			// clean and establishes nothing, which is what a proof looks like at its most
			// credible and least useful.
			switch {
			case v.Reproduced && v.Sound:
				fmt.Fprintf(&b, "- **audited by %s at r%d**: it REPRODUCES, and red read the script and accepts that it establishes the claim", v.SeatID, v.Round)
			case v.Reproduced && !v.Sound:
				fmt.Fprintf(&b, "- **audited by %s at r%d — REPRODUCES BUT DOES NOT PROVE THE CLAIM.** The script re-runs to the same output, and red read it and found it does not establish what it is anchored to. Re-running measures determinism, not validity", v.SeatID, v.Round)
			case !v.Reproduced && v.Sound:
				fmt.Fprintf(&b, "- **audited by %s at r%d**: the method is sound but it DID NOT REPRODUCE — the script no longer produces the recorded output", v.SeatID, v.Round)
			default:
				fmt.Fprintf(&b, "- **audited by %s at r%d — it neither reproduces NOR establishes the claim**", v.SeatID, v.Round)
			}
			if v.Note != "" {
				fmt.Fprintf(&b, " — %s", v.Note)
			}
			b.WriteString("\n")
			if !v.Reproduced && (v.Recorded != "" || v.Observed != "") {
				fmt.Fprintf(&b, "  - recorded: `%s`\n  - on re-run: `%s`\n", oneLine(v.Recorded), oneLine(v.Observed))
			}
		} else {
			b.WriteString("- **not independently audited** — nobody re-ran this script or read it; the computation stands on blue's execution alone\n")
		}
		if p.Cites != "" {
			// The METHOD is cited and the INSTANCE is computed: naming the citation here is
			// what makes the pair legible as one argument rather than two artifacts.
			fmt.Fprintf(&b, "- **applies the method cited at**: %s\n", p.Cites)
		}
		b.WriteString("\n")

		script, output := readArtifact(run, p.SHA)
		b.WriteString("```" + fenceLang(p.Script) + "\n" + strings.TrimRight(script, "\n") + "\n```\n\n")
		b.WriteString("Output:\n\n")
		b.WriteString("```\n" + strings.TrimRight(output, "\n") + "\n```\n\n")
	}
	return strings.TrimRight(b.String(), "\n") + "\n"
}

// oneLine keeps a captured output on a single markdown line.
func oneLine(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// readArtifact returns the script and output as they were recorded, or an explicit note in
// their place. Silence here would let a report claim a proof it cannot show.
func readArtifact(run record.Run, sha string) (script, output string) {
	dir := filepath.Join(run.Dir(), "proofs", sha)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "(the script artifact is missing from this run directory)", "(the output artifact is missing from this run directory)"
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "script") {
			if b, rerr := os.ReadFile(filepath.Join(dir, e.Name())); rerr == nil {
				script = string(b)
			}
		}
	}
	if script == "" {
		script = "(the script artifact is missing from this run directory)"
	}
	if b, rerr := os.ReadFile(filepath.Join(dir, "output")); rerr == nil {
		output = string(b)
	} else {
		output = "(the output artifact is missing from this run directory)"
	}
	return script, output
}

// fenceLang picks the code-fence language from the script's extension so the report
// highlights it the way the seat wrote it.
func fenceLang(script string) string {
	switch strings.ToLower(filepath.Ext(script)) {
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
	return "(untitled proof)"
}
