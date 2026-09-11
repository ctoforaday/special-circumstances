package surface

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/repotree"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/terms"
)

// ONE WORD PER CONCEPT, ON EVERY SURFACE A SEAT READS.
//
// The tool's own names are gated elsewhere: a verb, flag, view or enum word a prompt names must
// exist. Nothing gated the NOUNS — what the document, the record, the time units and the parties
// are called — and a census found fourteen names for the report and "remanded" meaning a gap closed
// on one page and kept open on another. A seat reads a constitution, a prompt and a help page in
// one sitting and cannot tell a second name from a second thing.
//
// This runs every GATED ban in the terms registry over the text a seat or a shipped reader meets:
// constitutions, the engine's prompt text, help pages, the seat-facing Go string literals, the
// proto's annotation text as the compiled descriptor carries it, the protocol skill with its
// templates and command, and both READMEs. A match that no mask blanks and no allow exempts fails,
// naming the file, the line, the variant and the word to use.
//
// The plugin's docs/*.md are REGISTRY-ONLY and NOT scanned: they are design and measurement notes
// that quote history as data, and gating them would mean allow-listing whole tables or falsifying
// recorded measurements. docs/vocabulary.md is generated from the registry instead.

// vocabSource is one text the gate reads, with the repository-relative path allows match against.
type vocabSource struct {
	path string // repository-relative, slash-separated
	loc  func(line int) string
	text string
}

func TestVocabularyProse(t *testing.T) {
	reg, err := terms.Load()
	if err != nil {
		t.Fatal(err)
	}
	sources := vocabularySources(t)
	u := terms.NewUsage()
	var bad []string
	for _, s := range sources {
		for _, h := range reg.Scan(s.path, s.text, u) {
			if h.AllowedBy != "" {
				continue
			}
			bad = append(bad, s.loc(h.Line)+": "+strconv.Quote(h.Match)+" is the banned variant "+
				strconv.Quote(h.Variant)+" — the word is "+strconv.Quote(h.Term))
		}
	}
	sort.Strings(bad)
	if len(bad) > 0 {
		t.Errorf("%d seat-facing use(s) of a banned variant:\n  %s\n\n"+
			"The terms registry (tools/internal/terms/terms.json) holds one word per concept. Use the word\n"+
			"it names. A legitimate neighbour phrase is a MASK on the ban; a file that must quote the variant\n"+
			"is an ALLOW with a reason. Both live in the registry, never in this test.",
			len(bad), strings.Join(bad, "\n  "))
	}
	// A MASK OR ALLOW THAT DID NO WORK is a hole in the ban nobody can see. It is reported here,
	// over the whole scan, because only the whole scan can say it matched nothing.
	if stale := reg.Stale(u); len(stale) > 0 {
		t.Errorf("%d mask(s) or allow(s) in the terms registry did no work in this scan — delete them:\n  %s",
			len(stale), strings.Join(stale, "\n  "))
	}
}

// TestVocabularyProseMutations runs the variants the gate exists to catch through the same
// matcher, at the path they would appear in. The gate is a negative assertion over a large scan,
// so a matcher that silently matched nothing would read exactly like a clean tree.
func TestVocabularyProseMutations(t *testing.T) {
	reg, err := terms.Load()
	if err != nil {
		t.Fatal(err)
	}
	const f = "plugins/frank-exchange-of-views/"
	for _, m := range []struct{ path, text string }{
		{f + "skills/research-protocol/scripts/debate.js", "Write your full candidate draft to the lane file."},
		{f + "agents/red-chair.md", "CLOSE THE OPERATOR CHANNEL EVERY SITTING."},
		{f + "tools/internal/cli/seat/help/edit.md", "an edit changes the living report."},
		{f + "tools/internal/cli/seat/verbs.go", "read how YOUR CHAIR is doing"},
		{f + "skills/research-protocol/SKILL.md", "## The run directory (the blackboard)"},
		{f + "agents/lead-judge.md", "The judge rules on the closings."},
		{"README.md", "records/*.jsonl is the event log."},
	} {
		hit := false
		for _, h := range reg.Scan(m.path, m.text, nil) {
			if h.AllowedBy == "" {
				hit = true
			}
		}
		if !hit {
			t.Errorf("%s: %q passed the vocabulary gate; it must fail", m.path, m.text)
		}
	}
}

// vocabularySources gathers the gate's scope. Every glob and walk refuses an empty result, so a
// moved directory is an error rather than a silently smaller scan.
func vocabularySources(t *testing.T) []vocabSource {
	t.Helper()
	root, err := repotree.Root()
	if err != nil {
		t.Fatal(err)
	}
	rel := func(p string) string {
		r, err := filepath.Rel(root, p)
		if err != nil {
			t.Fatal(err)
		}
		return filepath.ToSlash(r)
	}
	var out []vocabSource
	addFile := func(p string, transform func(string) string) {
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		text := string(b)
		if transform != nil {
			text = transform(text)
		}
		r := rel(p)
		out = append(out, vocabSource{path: r, text: text, loc: func(l int) string { return r + ":" + strconv.Itoa(l) }})
	}
	glob := func(parts ...string) []string {
		m, err := repotree.Glob(parts...)
		if err != nil {
			t.Fatal(err)
		}
		return m
	}
	const plugin = "plugins/frank-exchange-of-views"

	// C, T, the help pages and the READMEs: prose files, read whole.
	for _, g := range [][]string{
		{plugin, "agents", "*.md"},
		{plugin, "skills", "adversarial-audit", "SKILL.md"},
		{plugin, "skills", "research-protocol", "SKILL.md"},
		{plugin, "skills", "research-protocol", "references", "*.md"},
		{plugin, "commands", "research.md"},
		{plugin, "tools", "internal", "cli", "seat", "help", "*.md"},
		{"README.md"},
		{plugin, "README.md"},
	} {
		for _, p := range glob(g...) {
			addFile(p, nil)
		}
	}

	// P: the engine's prompt text, with line comments removed exactly as the prompt-verb gates
	// remove them — a comment is history for a maintainer, and no seat ever reads one.
	js, err := repotree.DebateJS()
	if err != nil {
		t.Fatal(err)
	}
	addFile(js, func(s string) string { return jsComment.ReplaceAllString(s, "") })

	// H-go: the seat-facing Go set, as string literals.
	for _, p := range seatFacingGo(t) {
		out = append(out, goLiterals(t, p, rel(p))...)
	}

	// The proto's annotation text, from the compiled descriptor — what EnumValueDoc, Usage and the
	// required-field refusals hand a seat. record.pb.go is never scanned: it is generated bytes.
	out = append(out, protoAnnotations(rel(filepath.Join(root, plugin, "tools", "internal", "record", "recordpb", "record.proto")))...)
	return out
}

// seatFacingGo is the Go set a seat can be handed text from: the seat-role command packages, the
// root-level cli files that are not operator-only, the flag declarations, and the record package
// minus its migrator, its SQL layer and generated code.
func seatFacingGo(t *testing.T) []string {
	t.Helper()
	const cli = "plugins/frank-exchange-of-views/tools/internal/cli"
	var out []string
	walk := func(dir ...string) []string {
		files, err := repotree.GoSources(dir...)
		if err != nil {
			t.Fatal(err)
		}
		return files
	}
	for _, pkg := range []string{"seat", "blue", "lens", "chair", "bench", "motion", "enumhelp"} {
		out = append(out, walk(cli, pkg)...)
	}
	// Operator-only root files: their text reaches the person running the research, never a seat.
	operatorOnly := map[string]bool{
		"capture.go": true, "setup.go": true, "dashboard.go": true, "dashboard_serve.go": true,
		"graph.go": true, "migrate.go": true, "ocr.go": true, "diagnostics.go": true, "tiers.go": true,
		"scorecard.go": true, "verify.go": true, "countclaims.go": true, "log.go": true,
	}
	root, err := repotree.Root()
	if err != nil {
		t.Fatal(err)
	}
	top, err := filepath.Glob(filepath.Join(root, cli, "*.go"))
	if err != nil || len(top) == 0 {
		t.Fatalf("no Go files directly under %s (%v)", cli, err)
	}
	for _, p := range top {
		if b := filepath.Base(p); !strings.HasSuffix(b, "_test.go") && !operatorOnly[b] {
			out = append(out, p)
		}
	}
	out = append(out, walk("plugins/frank-exchange-of-views/tools/internal/flags")...)
	// The projections a seat reads with `show`: their headers and notes are text a seat is handed.
	out = append(out, walk("plugins/frank-exchange-of-views/tools/internal/view")...)
	for _, p := range walk("plugins/frank-exchange-of-views/tools/internal/record") {
		s := filepath.ToSlash(p)
		// recordtest is fixture support for tests — its generated class list holds slugs such as
		// causal-narrative-fails-reproduction — and nothing in it reaches a seat.
		if strings.Contains(s, "/record/migrate/") || strings.Contains(s, "/recordsql/") ||
			strings.Contains(s, "/record/recordtest/") ||
			strings.Contains(s, "/testdata/") || strings.HasSuffix(s, ".pb.go") {
			continue
		}
		out = append(out, p)
	}
	return out
}

// goLiterals returns each string literal in a file as its own source — or, for a chain of literals
// joined with `+`, each run of adjacent literals as one, so a phrase split across a line break in
// the source is still one phrase. Comments are never read: they are for maintainers.
func goLiterals(t *testing.T, path, relPath string) []vocabSource {
	t.Helper()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var out []vocabSource
	emit := func(lits []*ast.BasicLit) {
		if len(lits) == 0 {
			return
		}
		var b strings.Builder
		for _, l := range lits {
			s, err := strconv.Unquote(l.Value)
			if err != nil {
				s = l.Value
			}
			b.WriteString(s)
		}
		// A LITERAL WITH NO WHITESPACE IS A NAME, NOT PROSE: a map key, a flag, an enum word, a
		// renderer's id (`view.MarkdownFrom(in, "ledger", "")`). The name gates own those. The bans
		// here are on the words a sentence uses, and a sentence has a space in it.
		if !strings.ContainsAny(b.String(), " \t\n") {
			return
		}
		line := fset.Position(lits[0].Pos()).Line
		raw := strings.HasPrefix(lits[0].Value, "`") && len(lits) == 1
		out = append(out, vocabSource{path: relPath, text: b.String(), loc: func(l int) string {
			if raw {
				return relPath + ":" + strconv.Itoa(line+l-1)
			}
			return relPath + ":" + strconv.Itoa(line)
		}})
	}
	var visit func(n ast.Node) bool
	visit = func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.BinaryExpr:
			if x.Op != token.ADD {
				return true
			}
			var leaves []ast.Expr
			var flatten func(e ast.Expr)
			flatten = func(e ast.Expr) {
				if be, ok := e.(*ast.BinaryExpr); ok && be.Op == token.ADD {
					flatten(be.X)
					flatten(be.Y)
					return
				}
				leaves = append(leaves, e)
			}
			flatten(x)
			var run []*ast.BasicLit
			for _, l := range leaves {
				if bl, ok := l.(*ast.BasicLit); ok && bl.Kind == token.STRING {
					run = append(run, bl)
					continue
				}
				emit(run)
				run = nil
				ast.Inspect(l, visit)
			}
			emit(run)
			return false
		case *ast.BasicLit:
			if x.Kind == token.STRING {
				emit([]*ast.BasicLit{x})
			}
		}
		return true
	}
	ast.Inspect(f, visit)
	return out
}

// protoAnnotations is every (means), (sql).why and (check).why string in the compiled record
// descriptor, each located by the full name it annotates.
func protoAnnotations(protoPath string) []vocabSource {
	var out []vocabSource
	add := func(where, text string) {
		if text == "" {
			return
		}
		out = append(out, vocabSource{path: protoPath, text: text, loc: func(int) string { return protoPath + " (" + where + ")" }})
	}
	var enums func(es protoreflect.EnumDescriptors)
	enums = func(es protoreflect.EnumDescriptors) {
		for i := 0; i < es.Len(); i++ {
			vs := es.Get(i).Values()
			for j := 0; j < vs.Len(); j++ {
				v := vs.Get(j)
				m, _ := proto.GetExtension(v.Options(), recordpb.E_Means).(string)
				add(string(v.FullName())+" means", m)
			}
		}
	}
	var msgs func(ms protoreflect.MessageDescriptors)
	msgs = func(ms protoreflect.MessageDescriptors) {
		for i := 0; i < ms.Len(); i++ {
			md := ms.Get(i)
			if cs, _ := proto.GetExtension(md.Options(), recordpb.E_Check).([]*recordpb.SqlCheck); len(cs) > 0 {
				for k, c := range cs {
					add(string(md.FullName())+" check["+strconv.Itoa(k)+"].why", c.GetWhy())
				}
			}
			fs := md.Fields()
			for j := 0; j < fs.Len(); j++ {
				fd := fs.Get(j)
				if s, _ := proto.GetExtension(fd.Options(), recordpb.E_Sql).(*recordpb.Sql); s != nil {
					add(string(fd.FullName())+" sql.why", s.GetWhy())
				}
			}
			enums(md.Enums())
			msgs(md.Messages())
		}
	}
	fd := recordpb.File_record_proto
	enums(fd.Enums())
	msgs(fd.Messages())
	return out
}
