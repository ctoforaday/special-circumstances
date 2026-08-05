package report

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// proofAnchor matches an invisible proof anchor "<!--proof:p-<id>-->" — the tool-inserted
// token marking a sentence a COMPUTATION backs.
var proofAnchor = regexp.MustCompile(`<!--proof:(p-[0-9a-f]+)-->`)

// weaveProofs turns the invisible proof layer into a visible one, the way weaveCitations does
// for sources — and it exists because without it the axis was a half-state that read as done.
//
// MEASURED ON ITS OWN FIRST BUILD: a proof ran, cached, anchored and was auditable by red,
// and the assembled report carried the raw "<!--proof:p-…-->" token while mentioning the
// computation ZERO times. The evidence existed everywhere except the document a human reads.
//
// A PROOF SHOWS ITS WORK. A citation can be a one-line reference because the source is
// elsewhere and permanent; a computation's source IS the evidence, so the section carries the
// script and the output verbatim. A reader who does not trust it can run it, and that is the
// whole point of preferring it to prose.
//
// The script and output are read from the ARTIFACT under <run>/proofs/<sha256>/, not from the
// record: the report must show the exact bytes that ran, not a copy the record made of them.
// A missing artifact is stated, never silently skipped — a proof section quietly short of a
// proof is worse than one that says the artifact is gone.
func weaveProofs(runDir, md string, proofs []record.Proof) string {
	byLabel := map[string]record.Proof{}
	for _, p := range proofs {
		byLabel[p.Label] = p
	}
	var order []string
	num := map[string]int{}
	body := proofAnchor.ReplaceAllStringFunc(md, func(tok string) string {
		label := proofAnchor.FindStringSubmatch(tok)[1]
		n, seen := num[label]
		if !seen {
			n = len(order) + 1
			num[label] = n
			order = append(order, label)
		}
		return fmt.Sprintf("[^P%d]", n)
	})
	if len(order) == 0 {
		return body
	}

	var b strings.Builder
	b.WriteString(strings.TrimRight(body, "\n"))
	b.WriteString("\n\n## Proofs\n\n")
	b.WriteString("Each of these is a computation this report ran, with the exact script and its output.\n")
	b.WriteString("`reproducible` means the same script produced identical output on two consecutive runs;\n")
	b.WriteString("`observed` means the output moved between them, which makes it a measurement of a system\n")
	b.WriteString("in motion rather than a proof. Every script is cached under `proofs/<sha256>/` in the run\n")
	b.WriteString("directory, so any reader can run it again.\n\n")

	for _, label := range order {
		n := num[label]
		p, ok := byLabel[label]
		if !ok {
			fmt.Fprintf(&b, "### [^P%d] _(unresolved proof %s — no proof event on the record)_\n\n", n, label)
			continue
		}
		fmt.Fprintf(&b, "### [^P%d] %s\n\n", n, strings.TrimSpace(firstNonEmpty(p.Reason, p.Script)))
		fmt.Fprintf(&b, "- **basis**: %s", p.Basis)
		if p.Drift != "" {
			fmt.Fprintf(&b, " — %s", p.Drift)
		}
		b.WriteString("\n")
		fmt.Fprintf(&b, "- **script**: `%s` (sha256 `%s`)\n", p.Script, shortSHA(p.SHA))
		fmt.Fprintf(&b, "- **exit**: %d\n", p.Exit)
		if p.Cites != "" {
			// The METHOD is cited and the INSTANCE is computed: naming the citation here is
			// what makes the pair legible as one argument rather than two artifacts.
			fmt.Fprintf(&b, "- **applies the method cited at**: %s\n", p.Cites)
		}
		b.WriteString("\n")

		script, output := readArtifact(runDir, p.SHA)
		b.WriteString("```" + fenceLang(p.Script) + "\n" + strings.TrimRight(script, "\n") + "\n```\n\n")
		b.WriteString("Output:\n\n")
		b.WriteString("```\n" + strings.TrimRight(output, "\n") + "\n```\n\n")
	}
	return strings.TrimRight(b.String(), "\n") + "\n"
}

// readArtifact returns the script and output as they were recorded, or an explicit note in
// their place. Silence here would let a report claim a proof it cannot show.
func readArtifact(runDir, sha string) (script, output string) {
	dir := filepath.Join(runDir, "proofs", sha)
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
