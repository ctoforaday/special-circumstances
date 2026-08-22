// Command protogen compiles the record schema and regenerates its Go bindings.
//
// Dev tooling for this repository only. Nothing here ships to an installing project.
//
// WHY THIS EXISTS RATHER THAN `protoc` OR `buf`.
//
// The generated .pb.go is COMMITTED, so `go build` needs no compiler — but regeneration has to
// be reachable by anyone, on any machine, without a system package. Both obvious answers fail
// that:
//
//   - `protoc` is a C++ binary installed with apt/brew. That is a privileged, per-platform
//     install for a repository whose own guardrails forbid one, and it is the classic "works on
//     the machine that generated it" dependency.
//   - `buf` is a Go binary and DOES install here — measured, after a first attempt failed and
//     was nearly written up as a blocker. `go install github.com/bufbuild/buf/cmd/buf@latest`
//     needs go >= 1.25.10 while this repository pins 1.25.0, but that constraint binds the
//     INSTALL only: `GOTOOLCHAIN=go1.25.10 go install ...` succeeds and yields a standalone
//     v1.72.0 binary that has nothing further to do with the build's Go version. The reason to
//     prefer this generator is therefore NOT that buf is unavailable. It is that buf is a
//     global binary somebody has to have, and regeneration here should need nothing that is not
//     already in the repository — `go run ./protogen` works on a clean checkout with no install
//     step, which is what makes the CI staleness gate runnable at all.
//
// BUF IS STILL WORTH HAVING, for the job this generator does not do: `buf breaking` is the
// natural runner for the field-number and enum-value superset gates (plans/record-protobuf.md
// §V.2), which need to diff this schema against its previous committed state. That is a
// separate, optional, developer-side check — not a build dependency.
//
// So: `github.com/bufbuild/protocompile` (a pure-Go .proto compiler — the same engine buf uses)
// produces the descriptor, and the descriptor is handed to protoc-gen-go, which is run
// MODULE-PINNED via `go run google.golang.org/protobuf/cmd/protoc-gen-go@<version>`. Nothing is
// installed globally, the plugin version is a string in this file rather than whatever happens
// to be on PATH, and `go run ./protogen` is the whole operator instruction.
//
// THE VERSION PIN IS LOAD-BEARING, not hygiene. protojson's whitespace is seeded per-binary
// (internal/detrand: "unstable across different builds") and its map ordering is an internal
// implementation detail, so the record's byte-level shape is a property of a specific protobuf
// release. The goldens are byte-compared. A floating @latest would re-record 24 goldens on
// somebody else's laptop for no stated reason.
package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/bufbuild/protocompile"
	"github.com/bufbuild/protocompile/protoutil"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"

	"github.com/ctoforaday/special-circumstances/scripts/internal/gitx"
)

// protocGenGoVersion must match the google.golang.org/protobuf version in the TOOLS module's
// go.mod. A generated file produced by a different runtime than the one that will unmarshal it
// is the skew this pin exists to prevent; checkVersionAgreement enforces it.
const protocGenGoVersion = "v1.36.12"

// schemaRel is the one schema. A second .proto would need a decision about import paths and
// generation order, so it is a constant rather than a glob: adding one is a change somebody
// makes deliberately.
const schemaRel = "plugins/frank-exchange-of-views/tools/internal/record/recordpb/record.proto"

// stampRel records the schema's content hash at the moment the bindings were generated.
//
// GIT DOES NOT PRESERVE MTIMES. A staleness gate comparing file times passes vacuously on every
// fresh CI checkout, because checkout stamps both files with the same instant — the plausible
// zero, in the check meant to catch one. The hash is the fact; the file is where it is written.
const stampRel = "plugins/frank-exchange-of-views/tools/internal/record/recordpb/record.proto.sha256"

func main() {
	check := flag.Bool("check", false, "verify the committed bindings match the schema; write nothing")
	flag.Parse()

	root, err := gitx.Root()
	if err != nil {
		fatal("cannot locate the repository root: %v", err)
	}
	if err := run(root, *check); err != nil {
		fatal("%v", err)
	}
}

func run(root string, checkOnly bool) error {
	schema := filepath.Join(root, schemaRel)
	src, err := os.ReadFile(schema)
	if err != nil {
		return fmt.Errorf("reading %s: %w", schemaRel, err)
	}
	sum := sha256.Sum256(src)
	want := hex.EncodeToString(sum[:])

	if checkOnly {
		return verify(root, want)
	}
	if err := checkVersionAgreement(root); err != nil {
		return err
	}

	files, err := generate(root, schema)
	if err != nil {
		return err
	}
	for _, f := range files {
		dest := filepath.Join(filepath.Dir(schema), filepath.Base(f.name))
		if err := os.WriteFile(dest, f.content, 0o644); err != nil {
			return fmt.Errorf("writing %s: %w", dest, err)
		}
		fmt.Printf("wrote %s (%d bytes)\n", mustRel(root, dest), len(f.content))
	}
	if err := os.WriteFile(filepath.Join(root, stampRel), []byte(want+"\n"), 0o644); err != nil {
		return fmt.Errorf("writing the schema stamp: %w", err)
	}
	fmt.Printf("stamped %s = %s\n", stampRel, want[:12])
	return nil
}

// verify is the gate. It fails when the schema has moved without the bindings being
// regenerated — and it fails LOUDLY when the stamp is missing, rather than treating an absent
// stamp as agreement.
func verify(root, want string) error {
	b, err := os.ReadFile(filepath.Join(root, stampRel))
	if err != nil {
		return fmt.Errorf("no schema stamp at %s: %w\n\n"+
			"The stamp records which schema the committed bindings were generated from. Its "+
			"ABSENCE is not agreement — run `go run ./protogen` and commit both", stampRel, err)
	}
	got := strings.TrimSpace(string(b))
	if got != want {
		return fmt.Errorf("the record schema has changed since its bindings were generated.\n\n"+
			"  %s\n    committed stamp: %s\n    schema now:      %s\n\n"+
			"Run `go run ./protogen` and commit the regenerated bindings alongside the .proto. "+
			"A schema whose generated code is stale compiles fine and writes the OLD shape, "+
			"which is indistinguishable from the schema change never having been made.",
			schemaRel, got, want)
	}
	fmt.Printf("bindings are current (%s = %s)\n", schemaRel, want[:12])
	return nil
}

// checkVersionAgreement holds this generator's plugin pin to the runtime the tools module will
// actually unmarshal with. Two carriers of one fact, with a gate between them.
func checkVersionAgreement(root string) error {
	gomod := filepath.Join(root, "plugins/frank-exchange-of-views/tools/go.mod")
	b, err := os.ReadFile(gomod)
	if err != nil {
		return fmt.Errorf("reading the tools go.mod: %w", err)
	}
	for _, line := range strings.Split(string(b), "\n") {
		f := strings.Fields(strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(line), "// indirect")))
		if len(f) >= 2 && f[0] == "google.golang.org/protobuf" {
			if f[1] != protocGenGoVersion {
				return fmt.Errorf("protoc-gen-go is pinned at %s here, but the tools module runs "+
					"google.golang.org/protobuf %s.\n\nThey must agree: the generator and the "+
					"runtime that unmarshals its output are one fact in two files, and a skew "+
					"between them shows up as byte differences in goldens rather than as an error",
					protocGenGoVersion, f[1])
			}
			return nil
		}
	}
	return fmt.Errorf("google.golang.org/protobuf is not required by the tools module — the "+
		"generated bindings would not compile there. Add it at %s", protocGenGoVersion)
}

type genFile struct {
	name    string
	content []byte
}

// generate compiles the schema to a descriptor and drives protoc-gen-go over it.
func generate(root, schema string) ([]genFile, error) {
	dir := filepath.Dir(schema)
	base := filepath.Base(schema)

	comp := protocompile.Compiler{
		Resolver:       protocompile.WithStandardImports(&protocompile.SourceResolver{ImportPaths: []string{dir}}),
		SourceInfoMode: protocompile.SourceInfoStandard, // keep comments: they become Go doc comments
	}
	linked, err := comp.Compile(context.Background(), base)
	if err != nil {
		return nil, fmt.Errorf("compiling %s: %w", schemaRel, err)
	}

	// DEPENDENCIES TRAVEL WITH THE FILE. protoc-gen-go resolves imports against the request alone,
	// so a schema that imports descriptor.proto — which ours does, to declare its own field options
	// — fails with "could not resolve import" unless the imported descriptors are in ProtoFile too.
	// The compiler resolved them; the request has to carry them.
	seen := map[string]bool{}
	var protos []*descriptorpb.FileDescriptorProto
	var add func(fd protoreflect.FileDescriptor)
	add = func(fd protoreflect.FileDescriptor) {
		if seen[fd.Path()] {
			return
		}
		seen[fd.Path()] = true
		imports := fd.Imports()
		for i := 0; i < imports.Len(); i++ {
			add(imports.Get(i).FileDescriptor) // depth first: a dependency must precede its dependent
		}
		protos = append(protos, protoutil.ProtoFromFileDescriptor(fd))
	}
	for _, f := range linked {
		add(f)
	}

	req := &pluginpb.CodeGeneratorRequest{
		FileToGenerate: []string{base},
		ProtoFile:      protos,
		CompilerVersion: &pluginpb.Version{
			Major: proto.Int32(5), Minor: proto.Int32(29), Patch: proto.Int32(0),
		},
	}
	in, err := proto.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshalling the generator request: %w", err)
	}

	cmd := exec.Command("go", "run", "google.golang.org/protobuf/cmd/protoc-gen-go@"+protocGenGoVersion)
	cmd.Dir = root
	cmd.Stdin = bytes.NewReader(in)
	cmd.Env = append(os.Environ(), "GOTOOLCHAIN=go1.25.0")
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errb
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("running protoc-gen-go@%s: %w\n%s", protocGenGoVersion, err, errb.String())
	}

	var resp pluginpb.CodeGeneratorResponse
	if err := proto.Unmarshal(out.Bytes(), &resp); err != nil {
		return nil, fmt.Errorf("decoding the generator response: %w", err)
	}
	if resp.Error != nil && *resp.Error != "" {
		return nil, fmt.Errorf("protoc-gen-go refused the schema: %s", *resp.Error)
	}
	if len(resp.File) == 0 {
		// A generator that produced nothing and reported no error is the silent-zero shape:
		// the run "succeeds" and the bindings never change.
		return nil, fmt.Errorf("protoc-gen-go returned no files and no error for %s", schemaRel)
	}

	files := make([]genFile, 0, len(resp.File))
	for _, f := range resp.File {
		files = append(files, genFile{name: f.GetName(), content: []byte(f.GetContent())})
	}
	return files, nil
}

func mustRel(root, p string) string {
	r, err := filepath.Rel(root, p)
	if err != nil {
		return p
	}
	return r
}

func fatal(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "protogen: "+format+"\n", a...)
	os.Exit(1)
}
