package report

import (
	"runtime"
	"strings"
	"testing"

	"github.com/yuin/goldmark/ast"
	east "github.com/yuin/goldmark/extension/ast"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/compute"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// shapes counts the nodes of a parsed document that a string a program chose could open: headings
// (by text), links, images, raw HTML, footnote references, emphasis, code spans and table cells.
func shapes(md string) map[string]int {
	src := []byte(md)
	out := map[string]int{}
	ast.Walk(parsePage(src), func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch n := n.(type) {
		case *ast.Heading:
			out["heading "+string(n.Lines().Value(src))]++
		case *ast.Link:
			out["link"]++
		case *ast.Image:
			out["image"]++
		case *ast.RawHTML, *ast.HTMLBlock:
			out["html"]++
		case *ast.Emphasis:
			out["emphasis"]++
		case *ast.CodeSpan:
			out["code span"]++
		case *east.FootnoteLink:
			out["footnote"]++
		case *east.TableCell:
			out["cell"]++
		case *ast.FencedCodeBlock:
			out["fence"]++
		}
		return ast.WalkContinue, nil
	})
	return out
}

// LITERAL RENDERS A STRING AS ITS OWN CHARACTERS (T1). Each input would open something spliced raw;
// through Literal it opens nothing, and its text is the input with its line breaks folded.
func TestLiteralOpensNothingAndReadsAsTyped(t *testing.T) {
	for _, s := range []string{
		"![x](computations/abc/acts/K-1/out/f.png)", "[unbalanced", "[^K1]", "a | b", "a\n## X", "a\r\nb\rc",
		"*a* _b_", "a`b``c", "<b>bold</b>", "&amp; &#60;", "x](y.png)", "\\[^K1]", "1. item", "- item", "> quote",
	} {
		lit := Literal(s)
		if strings.ContainsAny(lit, "\r\n") {
			t.Errorf("Literal(%q) = %q holds a line break", s, lit)
		}
		for k, n := range shapes("Text " + lit + " end.\n") {
			if !strings.HasPrefix(k, "heading") || n > 0 {
				t.Errorf("Literal(%q) = %q opens %s", s, lit, k)
			}
		}
		var text strings.Builder
		src := []byte("Text " + lit + " end.\n")
		ast.Walk(parsePage(src), func(n ast.Node, entering bool) (ast.WalkStatus, error) {
			if tx, ok := n.(*ast.Text); ok && entering {
				text.Write(tx.Segment.Value(src))
				if tx.SoftLineBreak() {
					text.WriteByte(' ')
				}
			}
			if st, ok := n.(*ast.String); ok && entering {
				text.Write(st.Value)
			}
			return ast.WalkContinue, nil
		})
		folded := strings.NewReplacer("\r\n", " ", "\r", " ", "\n", " ").Replace(s)
		if got := strings.ReplaceAll(text.String(), "\\", ""); !strings.Contains(got, strings.ReplaceAll(folded, "\\", "")) {
			t.Errorf("Literal(%q) reads as %q", s, text.String())
		}
	}
}

// hostileFile and hostileOut are a bundle file and an out/ file a program named. Both are written to
// disk, so on Windows, which refuses `<`, `>`, `|`, `:`, `"`, `?`, `*` and control characters in a
// file name, they are built from the Markdown-active bytes it allows — a backtick, brackets,
// parentheses, `#`, `!`, `_`, `~`, `&`, `'` and spaces — and still pass through Literal there.
var hostileFile, hostileOut = func() (string, string) {
	if runtime.GOOS == "windows" {
		return "p](x) `b` ## Y ![i](j) _e_ ~s~ &amp; 'q'.py", "o`ut ## X](y.png) [^K1] !"
	}
	return "p](x) `<b>`\n## Y.py", "o`ut\n## X](y.png)"
}()

// A NAME A PROGRAM CHOSE OPENS NOTHING IN evidence.md (T1). One computation whose command, toolchain
// path, bundle file, out/ file, drifts and outputs each hold a backtick, a line break before `## X`,
// a link, raw HTML and a fence, beside the same computation with plain names: the two documents
// parse to the same headings, links, images, raw HTML, footnotes, emphasis and fences.
func TestAProgramsNamesOpenNothingInTheEvidenceDocument(t *testing.T) {
	doc := func(t *testing.T, hostile bool) string {
		runDir := newRun(t)
		name := func(h, plain string) string {
			if hostile {
				return h
			}
			return plain
		}
		file := name(hostileFile, "p.py")
		outName := name(hostileOut, "out.png")
		argv := []string{"python3", file, name("a\n## Z <i>[^K1]", "arg")}
		files := []compute.File{{Path: file, Body: []byte("print(1)\n")}}
		bundle := compute.BundleSha(argv, compute.Sums(files))
		if err := compute.WriteBundle(runDir, bundle, argv, files); err != nil {
			t.Fatal(err)
		}
		output := name("```\n## W\n====\n  Traceback (most recent call last):\n", "plain output\n")
		if err := compute.WriteAct(runDir, bundle, "K-00000001", []byte(output), []compute.File{{Path: outName, Body: []byte("x")}}); err != nil {
			t.Fatal(err)
		}
		rerun := compute.Execution{Output: name("````\n## V\n", "other output\n")}
		staged, err := compute.StageRerun(runDir, bundle, rerun)
		if err != nil {
			t.Fatal(err)
		}
		if err := staged.Promote(); err != nil {
			t.Fatal(err)
		}
		k := record.Computation{
			Label: "K-00000001", SHA: bundle, Argv: argv, Basis: "variant", Reason: "counts",
			ToolchainPath: name("/usr/bin/py](t)<i>`\n## T", "/usr/bin/python3"), ToolchainSha256: strings.Repeat("a", 64),
			Drift: "out/" + outName + " differs", Invariant: "a count",
			Executed: true,
			Rerun: &record.ComputationRerun{Rerun: recordpb.Rerun_RERUN_DIFFERS, RerunSha: staged.Sha,
				Drift: name("out/"+outName+" written by this act's run only; on re-run the program wrote: ````\n## V\n====\n    File \"x\"", "out/out.png written by this act's run only; on re-run the program wrote: other output")},
		}
		return evidenceDoc(runtest.Open(t, runDir), []record.Computation{k}, map[string]bool{"K-00000001": true})
	}
	plain, hostile := shapes(doc(t, false)), shapes(doc(t, true))
	for k, n := range plain {
		if hostile[k] != n {
			t.Errorf("the hostile names change %s: %d, the plain document's %d", k, hostile[k], n)
		}
	}
	for k, n := range hostile {
		if _, ok := plain[k]; !ok {
			t.Errorf("the hostile names open %d %s", n, k)
		}
	}
}
