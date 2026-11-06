package recordpb

import (
	"fmt"
	"slices"
	"strings"
	"text/template"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// WHAT RECORD.PROTO DECLARES ON EACH ENUM VALUE, AS GO LITERALS.
//
// The annotations are authored on the value in record.proto and read here off a generated table,
// never off the descriptor's options at run time. Decoding a value's options means unmarshalling
// its option bytes against the extension registry, and the record package's own initialisation
// reads every Disposition, every LogType and every closed set a flag offers — paid by every
// feov-record, feov-sitting-write and feov-migrate process, each of which makes one call. A
// sync.Once saves nothing in a process that asks once; a literal costs nothing at all.
//
// The table is written by `go -C plugins/frank-exchange-of-views/tools run ./devcmd/vocabgen`
// from AnnotationsFromOptions, and two gates hold it to the schema: vocabgen -check, and
// TestTheAnnotationTableIsWhatTheSchemaDeclares, which derives every row afresh and compares.

// ValueAnnotations is every annotation one enum value carries.
type ValueAnnotations struct {
	Name protoreflect.Name

	Means, DefinedTerm string
	RuledBy            string
	Neighbor           string
	Distinguisher      string

	Closes, SeatMayFile DeclaredBool
	Mass                DeclaredNumber
	MaterialDefault     DeclaredEnum

	// GapClasses and FoundSourceGapClasses are a SourceOutcome's `gap_classes` and
	// `found_source_gap_classes` options, in declaration order: the classes a reading with that
	// outcome may mint its gap in. Nil where the value declares none.
	GapClasses, FoundSourceGapClasses []GapClass
}

// DeclaredBool is a boolean facet with its declared/undeclared split: false is an answer, absent
// is nobody asked.
type DeclaredBool struct{ Declared, Value bool }

// DeclaredNumber is a numeric facet with the same split: 0 is a real weight.
type DeclaredNumber struct {
	Declared bool
	Value    float64
}

// DeclaredEnum is an enum-valued facet with the same split: absent is never the zero word.
type DeclaredEnum struct {
	Declared bool
	Value    protoreflect.EnumNumber
}

// AnnotationsFromOptions reads one value's annotations off its descriptor options. It is the
// derivation the generator writes the table from and the staleness test compares it against; no
// shipped path calls it.
func AnnotationsFromOptions(vd protoreflect.EnumValueDescriptor) ValueAnnotations {
	a := ValueAnnotations{Name: vd.Name()}
	opts := vd.Options()
	if opts == nil {
		return a
	}
	str := func(xt protoreflect.ExtensionType) string { s, _ := proto.GetExtension(opts, xt).(string); return s }
	flag := func(xt protoreflect.ExtensionType) DeclaredBool {
		if !proto.HasExtension(opts, xt) {
			return DeclaredBool{}
		}
		return DeclaredBool{true, proto.GetExtension(opts, xt).(bool)}
	}
	a.Means, a.DefinedTerm, a.RuledBy = renderMeans(vd, str(E_Means)), str(E_DefinedTerm), str(E_RuledBy)
	a.Neighbor, a.Distinguisher = str(E_Neighbor), str(E_Distinguisher)
	a.Closes, a.SeatMayFile = flag(E_Closes), flag(E_SeatMayFile)
	if proto.HasExtension(opts, E_Mass) {
		a.Mass = DeclaredNumber{true, proto.GetExtension(opts, E_Mass).(float64)}
	}
	if proto.HasExtension(opts, E_MaterialDefault) {
		a.MaterialDefault = DeclaredEnum{true, proto.GetExtension(opts, E_MaterialDefault).(protoreflect.Enum).Number()}
	}
	classes := func(xt protoreflect.ExtensionType) []GapClass {
		if !proto.HasExtension(opts, xt) {
			return nil
		}
		return slices.Clone(proto.GetExtension(opts, xt).([]GapClass))
	}
	a.GapClasses, a.FoundSourceGapClasses = classes(E_GapClasses), classes(E_FoundSourceGapClasses)
	return a
}

// meansValues is what a `(means)` may render: the texts stated once in Go that a meaning carries
// word for word.
var meansValues = map[string]string{"DemandClause": DemandClause}

// renderMeans executes a `(means)` that holds `{{` over meansValues, so the generated table, and
// every reader of a value's meaning through it (--help, a refusal, the record's vocabulary tables),
// holds the rendered sentence; any other `(means)` returns as written. A template that does not
// parse or names no value is a record.proto authoring error, and the generator stops on it:
// AnnotationsFromOptions runs in vocabgen and its staleness test, never in a shipped path.
func renderMeans(vd protoreflect.EnumValueDescriptor, means string) string {
	if !strings.Contains(means, "{{") {
		return means
	}
	t, err := template.New(string(vd.FullName())).Option("missingkey=error").Parse(means)
	if err != nil {
		panic(fmt.Sprintf("the (means) of %s does not parse as a template: %v", vd.FullName(), err))
	}
	var b strings.Builder
	if err := t.Execute(&b, meansValues); err != nil {
		panic(fmt.Sprintf("the (means) of %s names a value it cannot render: %v", vd.FullName(), err))
	}
	return b.String()
}

// annotationsOf is one value's row of the generated table.
//
// The miss is LOUD. A value of another schema has no row, and a row whose name is not the value's
// is a table older than record.proto; either one answered as "undeclared" would read exactly like
// a value nobody annotated.
func annotationsOf(vd protoreflect.EnumValueDescriptor) (*ValueAnnotations, error) {
	if vd.ParentFile() != File_record_proto {
		return nil, fmt.Errorf("enum value %s is not declared in record.proto, so it carries no record annotations", vd.FullName())
	}
	rows := generatedAnnotations(vd.Parent().FullName())
	i := vd.Index()
	if i >= len(rows) || rows[i].Name != vd.Name() {
		return nil, fmt.Errorf("the generated annotation table has no row for %s — it is older than record.proto; "+
			"run `go -C plugins/frank-exchange-of-views/tools run ./devcmd/vocabgen`", vd.FullName())
	}
	return &rows[i], nil
}
