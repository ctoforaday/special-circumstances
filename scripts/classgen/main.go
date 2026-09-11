// Command classgen derives the gap-class tables Go reads, out of the registry that actually
// ships them.
//
// Dev tooling for this repository only. Nothing here ships to an installing project.
//
// # Why generate rather than guard
//
// feov-memory/class-registry.json is the shipped vocabulary: `record.StageForRun` writes it
// into a run, and mint validation refuses a class that is not in it. Three test files each
// restated all 38 slugs by hand, under a comment saying "mirrors feov-memory/
// class-registry.json" — and nothing compared them. They were in sync when this was written,
// verified; the failure is what happens LATER. Rename a slug in the registry and the fixtures
// go on validating a vocabulary production no longer accepts, green the whole time, because
// the tests stage the class they also assert.
//
// The repository's own rule says to prefer GENERATING a derived carrier and gating staleness
// over guarding two hand-written copies, and this is exactly that shape: one authored record,
// several derived readers. So the Go copies are generated from the JSON and this command is also
// the gate — `-check` regenerates both in memory and fails if either committed file has drifted.
//
// # The two outputs
//
//   - recordtest/classes_gen.go: the slugs, which fixtures stage.
//   - record/shippedclasses_gen.go: each slug's `material_default`, which `feov-record migrate`
//     reads to translate a run whose staged registry predates the field. It is compiled in so the
//     translation is a property of the binary — the same archive migrates the same way on every
//     machine — rather than of whatever memory directory the operator's working tree holds.
//
// # Why every row must carry `material_default`
//
// The field is where a gap's materiality starts, and the setup, the loader and the migration
// all refuse a row without it. A hand-added row missing it is therefore a CI failure here, at
// generation, rather than a run that refuses to start on the operator's machine.
//
// # Why a committed file rather than reading the JSON at test time
//
// The registry lives at the repository root, OUTSIDE the module that needs it, so `go:embed`
// cannot reach it. Reading it at runtime would make every binary depend on locating the
// repository root — the dependency `internal/record`'s own fixture comment refuses, because a
// reader that hunts for the repo root breaks when files move. A committed generated file has
// neither problem, and `-check` is what keeps it honest.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"sort"

	"github.com/ctoforaday/special-circumstances/scripts/internal/gitx"
)

const (
	registryPath = "feov-memory/class-registry.json"
	classesPath  = "plugins/frank-exchange-of-views/tools/internal/record/recordtest/classes_gen.go"
	shippedPath  = "plugins/frank-exchange-of-views/tools/internal/record/shippedclasses_gen.go"
)

// materialConst maps a registry word to the Go constant the shipped table is written with. The
// set is closed: a word outside it fails generation, naming the row.
var materialConst = map[string]string{
	"always":   "recordpb.ClassMaterial_CLASS_MATERIAL_ALWAYS",
	"never":    "recordpb.ClassMaterial_CLASS_MATERIAL_NEVER",
	"by_grade": "recordpb.ClassMaterial_CLASS_MATERIAL_BY_GRADE",
}

type registry struct {
	Classes []struct {
		Slug            string  `json:"slug"`
		MaterialDefault *string `json:"material_default"`
	} `json:"classes"`
}

func main() {
	check := flag.Bool("check", false, "verify the committed files match the registry instead of writing them")
	flag.Parse()

	root, err := gitx.Root()
	if err != nil {
		fmt.Fprintln(os.Stderr, "classgen:", err)
		os.Exit(1)
	}
	msgs, err := run(root, *check)
	for _, m := range msgs {
		fmt.Println(m)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "classgen:", err)
		os.Exit(1)
	}
}

// run generates both outputs under root and writes them, or with check compares them to what is
// committed and fails on the first that has drifted.
func run(root string, check bool) ([]string, error) {
	raw, err := os.ReadFile(filepath.Join(root, registryPath))
	if err != nil {
		return nil, err
	}
	outs, err := generate(raw)
	if err != nil {
		return nil, err
	}
	var msgs []string
	for _, rel := range []string{classesPath, shippedPath} {
		want := outs[rel]
		target := filepath.Join(root, rel)
		if !check {
			if err := os.WriteFile(target, want, 0o644); err != nil {
				return msgs, err
			}
			msgs = append(msgs, "classgen: wrote "+rel)
			continue
		}
		got, err := os.ReadFile(target)
		if err != nil {
			return msgs, fmt.Errorf("%s is missing (%v) — run `go run ./classgen` in scripts/", rel, err)
		}
		if !bytes.Equal(got, want) {
			return msgs, fmt.Errorf("%s is STALE against %s.\n"+
				"The gap-class registry moved and the generated copy did not, so a reader is using a\n"+
				"vocabulary or a materiality table that may no longer be the shipped one.\n"+
				"Regenerate: (cd scripts && go run ./classgen)", rel, registryPath)
		}
		msgs = append(msgs, fmt.Sprintf("classgen: %s matches %s", filepath.Base(rel), registryPath))
	}
	return msgs, nil
}

// generate parses the registry and renders both files. It refuses a row with no
// `material_default` or an unknown one, naming the slug.
func generate(raw []byte) (map[string][]byte, error) {
	var reg registry
	if err := json.Unmarshal(raw, &reg); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", registryPath, err)
	}
	defaults := map[string]string{}
	slugs := make([]string, 0, len(reg.Classes))
	for _, c := range reg.Classes {
		if c.Slug == "" {
			continue
		}
		if c.MaterialDefault == nil {
			return nil, fmt.Errorf("%s: class %q has no material_default — add `material_default` (always | never | by_grade) to that row", registryPath, c.Slug)
		}
		if _, ok := materialConst[*c.MaterialDefault]; !ok {
			return nil, fmt.Errorf("%s: class %q has material_default %q, which is not always | never | by_grade", registryPath, c.Slug, *c.MaterialDefault)
		}
		defaults[c.Slug] = *c.MaterialDefault
		slugs = append(slugs, c.Slug)
	}
	// A registry that parses to NO classes must not silently generate an empty vocabulary:
	// every fixture would then stage nothing, every class would be refused, and the failure
	// would look like a hundred broken tests rather than one broken file.
	if len(slugs) == 0 {
		return nil, fmt.Errorf("%s parsed to ZERO classes — refusing to generate an empty vocabulary", registryPath)
	}
	sort.Strings(slugs)

	var b bytes.Buffer
	fmt.Fprintf(&b, `// Code generated by scripts/classgen from %s. DO NOT EDIT.

package recordtest

// ShippedClasses is the gap-class vocabulary the plugin actually ships, in the order the
// registry holds it after sorting.
//
// Fixtures stage this rather than restating it. Three test files used to carry their own
// hand-written copy of all %d slugs, under a comment pointing at the registry and nothing
// comparing them to it — so a rename would have left every fixture validating a vocabulary
// production had stopped accepting, with the suite green throughout.
var ShippedClasses = []string{
`, registryPath, len(slugs))
	for _, s := range slugs {
		fmt.Fprintf(&b, "\t%q,\n", s)
	}
	fmt.Fprintln(&b, "}")
	classes, err := format.Source(b.Bytes())
	if err != nil {
		return nil, err
	}

	b.Reset()
	fmt.Fprintf(&b, `// Code generated by scripts/classgen from %s. DO NOT EDIT.

package record

import "github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"

// ShippedMaterialDefaults is each shipped class's material_default, as this binary was built.
//
// The migrate verb reads it to translate a run whose staged registry or class coinings predate the
// field: a slug held here takes its value, and one absent takes by_grade as a stated fill. A
// class adopted into the registry after the build is absent here until the next build.
var ShippedMaterialDefaults = map[string]recordpb.ClassMaterial{
`, registryPath)
	for _, s := range slugs {
		fmt.Fprintf(&b, "\t%q: %s,\n", s, materialConst[defaults[s]])
	}
	fmt.Fprintln(&b, "}")
	shipped, err := format.Source(b.Bytes())
	if err != nil {
		return nil, err
	}
	return map[string][]byte{classesPath: classes, shippedPath: shipped}, nil
}
