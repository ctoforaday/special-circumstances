// Command massgen writes the engine's grade-weight table from the schema that owns it.
//
// # What this replaces, and why the old answer expired
//
// The mass mapping lived in TWO hand-written tables — `record.MASS` in Go and `const MASS` in
// debate.js — held level by a regex parity test. That test carried a required statement of why it
// was a guard and not generation: "debate.js is loaded by node and by goja at runtime, with no
// build step in between, so there is nowhere to generate INTO."
//
// The premise was true and is no longer. When it was written NEITHER table was authoritative, so
// generating one from the other only moved the guess. A grade's weight is now a FACET ON THE
// SCHEMA (`Grade`'s `(mass)` annotation), which makes `enum_grade.mass` a column and Go's `MASS`
// a derivation — and gives the engine's copy something to be generated FROM. The guard becomes
// generation plus a staleness gate, which is the order facts-are-fields asks for.
//
// # Why it writes INTO debate.js rather than beside it
//
// debate.js has no imports. It is executed by the harness as a workflow script and by goja in the
// probe, and both load the single file — so a generated module beside it could not be reached. The
// block is therefore delimited in place, and everything outside the markers is left byte-for-byte
// alone.
//
// `-check` regenerates in memory and fails if the committed file has drifted, which is the same
// gate classgen and schemagen run in CI.
package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/bufbuild/protocompile"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"

	"github.com/ctoforaday/special-circumstances/scripts/internal/gitx"
)

const (
	schemaRel = "plugins/frank-exchange-of-views/tools/internal/record/recordpb/record.proto"
	engineRel = "plugins/frank-exchange-of-views/skills/research-protocol/scripts/debate.js"

	beginMark = "// BEGIN GENERATED MASS — `cd scripts && go run ./massgen`. Source: record.proto Grade (mass)."
	endMark   = "// END GENERATED MASS"
)

// weights reads each Grade value's `mass` facet off the compiled schema.
//
// THE ZERO IS A REAL WEIGHT AND MUST NOT BE INFERRED. GRADE_REALIZED weighs 0 by design — mass
// forecasts what is still to come — so a value that carries no annotation is an ERROR here rather
// than a zero, exactly as the schema's NOT NULL column refuses one.
func weights(root string) ([]string, []float64, error) {
	dir := filepath.Join(root, filepath.Dir(schemaRel))
	comp := protocompile.Compiler{
		Resolver: protocompile.WithStandardImports(&protocompile.SourceResolver{ImportPaths: []string{dir}}),
	}
	linked, err := comp.Compile(context.Background(), filepath.Base(schemaRel))
	if err != nil {
		return nil, nil, fmt.Errorf("compiling %s: %w", schemaRel, err)
	}
	fd := linked[0]

	var massExt protoreflect.ExtensionDescriptor
	for i := 0; i < fd.Extensions().Len(); i++ {
		if x := fd.Extensions().Get(i); x.Name() == "mass" {
			massExt = x
		}
	}
	if massExt == nil {
		return nil, nil, fmt.Errorf("%s declares no `mass` extension — the facet this generator exists to read", schemaRel)
	}
	ed := fd.Enums().ByName("Grade")
	if ed == nil {
		return nil, nil, fmt.Errorf("%s declares no Grade enum", schemaRel)
	}

	xt := dynamicpb.NewExtensionType(massExt)
	var names []string
	var vals []float64
	for i := 0; i < ed.Values().Len(); i++ {
		vd := ed.Values().Get(i)
		word := strings.ToLower(strings.TrimPrefix(string(vd.Name()), "GRADE_"))
		if word == "unspecified" {
			continue // the sentinel is absence; the engine keys on real grades only
		}
		opts := vd.Options()
		if opts == nil || !proto.HasExtension(opts, xt) {
			return nil, nil, fmt.Errorf("%s carries no (mass) — a grade without a weight cannot be "+
				"generated as one, and 0 is a real weight rather than a default", vd.Name())
		}
		v, ok := proto.GetExtension(opts, xt).(float64)
		if !ok {
			return nil, nil, fmt.Errorf("%s's (mass) is not a double", vd.Name())
		}
		names = append(names, word)
		vals = append(vals, v)
	}
	if len(names) == 0 {
		return nil, nil, fmt.Errorf("Grade yielded no weighted values — refusing to generate an empty table")
	}
	sort.SliceStable(names, func(i, j int) bool { return false }) // declaration order, stated rather than incidental
	return names, vals, nil
}

func render(names []string, vals []float64) string {
	parts := make([]string, len(names))
	for i, n := range names {
		key := n
		if strings.Contains(n, "_") {
			key = "'" + n + "'"
		}
		parts[i] = key + ": " + strconv.FormatFloat(vals[i], 'g', -1, 64)
	}
	return beginMark + "\n" +
		"// A grade's weight is a property of the WORD, annotated on Grade and carried by\n" +
		"// enum_grade.mass; this table is that annotation, rendered for the engine. Editing it here\n" +
		"// is editing a copy — change record.proto and regenerate.\n" +
		"const MASS = { " + strings.Join(parts, ", ") + " }\n" +
		endMark
}

func main() {
	check := flag.Bool("check", false, "verify debate.js matches the schema instead of writing it")
	flag.Parse()

	root, err := gitx.Root()
	if err != nil {
		fmt.Fprintln(os.Stderr, "massgen:", err)
		os.Exit(1)
	}
	names, vals, err := weights(root)
	if err != nil {
		fmt.Fprintln(os.Stderr, "massgen:", err)
		os.Exit(1)
	}
	target := filepath.Join(root, engineRel)
	src, err := os.ReadFile(target)
	if err != nil {
		fmt.Fprintln(os.Stderr, "massgen:", err)
		os.Exit(1)
	}
	b, e := bytes.Index(src, []byte(beginMark)), bytes.Index(src, []byte(endMark))
	if b < 0 || e < 0 || e < b {
		fmt.Fprintf(os.Stderr, "massgen: %s carries no generated MASS block. Put these two lines around "+
			"the const and rerun:\n  %s\n  %s\n", engineRel, beginMark, endMark)
		os.Exit(1)
	}
	want := append(append(append([]byte{}, src[:b]...), []byte(render(names, vals))...), src[e+len(endMark):]...)

	if *check {
		if !bytes.Equal(src, want) {
			fmt.Fprintf(os.Stderr, "massgen: %s has drifted from the schema. Run `cd scripts && go run ./massgen`.\n", engineRel)
			os.Exit(1)
		}
		fmt.Printf("massgen: %s matches the schema (%d grades)\n", engineRel, len(names))
		return
	}
	if bytes.Equal(src, want) {
		fmt.Printf("massgen: %s already matches the schema (%d grades)\n", engineRel, len(names))
		return
	}
	if err := os.WriteFile(target, want, 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "massgen:", err)
		os.Exit(1)
	}
	fmt.Printf("massgen: wrote the MASS table into %s (%d grades)\n", engineRel, len(names))
}
