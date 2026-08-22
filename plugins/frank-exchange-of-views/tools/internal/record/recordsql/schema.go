// Package recordsql derives the record's SQLite schema from the protobuf descriptors.
//
// # Why the schema is derived and not written down
//
// A committed .sql file is a second copy of the proto, and this repository has spent a migration
// removing second copies. It would need a staleness gate, the gate would need an allowlist, and the
// allowlist would be hand-kept — the defect reproduced one level up, which [[facts-are-fields]]
// names directly. Struct tags have the same shape with an ORM attached.
//
// Deriving it in process means the schema cannot disagree with the schema. There is nothing to keep
// in step and no gate to write, because there is only one artifact.
//
// # Why there are no migrations
//
// A run directory IS the database: every run creates its own, so a schema change never meets an old
// one. That is the ground `a12362c` already stands on — a project in building mode whose every
// record is a test run — and it is what makes derivation affordable rather than reckless.
package recordsql

import (
	"fmt"
	"sort"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// EnvelopeDDL is the one table written by hand, because it is the one thing the body messages do
// not describe: WHO recorded an act, WHEN, and in what order.
//
// STRICT is not decoration. SQLite's default column affinity accepts a string into an INTEGER
// column and stores it as a string, which would undo the typing this migration exists to create.
//
// The append-only guarantee is a TRIGGER rather than a convention. The record is evidence in an
// adversarial process: red audits what blue wrote and the bench rules on both, so a row that can be
// edited after the fact is not evidence. Nothing in the application layer is trusted to remember.
const EnvelopeDDL = `
CREATE TABLE "events" (
  "id"      INTEGER PRIMARY KEY,
  "seat_id" TEXT    NOT NULL,
  "round"   INTEGER NOT NULL,
  "seq"     INTEGER NOT NULL,
  "nonce"   TEXT    NOT NULL,
  "ts"      TEXT    NOT NULL,
  "type"    TEXT    NOT NULL,
  "key"     TEXT,
  UNIQUE ("seat_id", "nonce", "seq")
) STRICT;

CREATE UNIQUE INDEX "events_key" ON "events" ("key") WHERE "key" IS NOT NULL;
CREATE INDEX "events_type" ON "events" ("type");
CREATE INDEX "events_round" ON "events" ("round");

CREATE TRIGGER "events_are_append_only_update" BEFORE UPDATE ON "events" BEGIN
  SELECT RAISE(ABORT, 'the record is append-only: an event cannot be edited after it is written');
END;
CREATE TRIGGER "events_are_append_only_delete" BEFORE DELETE ON "events" BEGIN
  SELECT RAISE(ABORT, 'the record is append-only: an event cannot be removed after it is written');
END;
`

// Schema returns the full DDL: the envelope, then one table per body type in a stable order.
func Schema() (string, error) {
	var b strings.Builder
	b.WriteString(EnvelopeDDL)
	bodies, err := Bodies()
	if err != nil {
		return "", err
	}
	for _, body := range bodies {
		ddl, err := tableFor(body)
		if err != nil {
			return "", err
		}
		b.WriteString("\n")
		b.WriteString(ddl)
	}
	return b.String(), nil
}

// Bodies lists the body messages, in the order the `body` oneof declares them, so the DDL is
// byte-stable across runs. A map iteration here would make every regeneration a diff.
func Bodies() ([]protoreflect.MessageDescriptor, error) {
	md := (&recordpb.Event{}).ProtoReflect().Descriptor()
	od := md.Oneofs().ByName("body")
	if od == nil {
		return nil, fmt.Errorf("recordsql: Event has no `body` oneof — the schema cannot be derived")
	}
	out := make([]protoreflect.MessageDescriptor, 0, od.Fields().Len())
	for i := 0; i < od.Fields().Len(); i++ {
		f := od.Fields().Get(i)
		if f.Message() == nil {
			return nil, fmt.Errorf("recordsql: body arm %q is not a message", f.Name())
		}
		out = append(out, f.Message())
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("recordsql: the body oneof declares no arms — an empty schema would apply cleanly and hold nothing")
	}
	return out, nil
}

// TableName is the SQL name for a body message: its own name in snake_case, which is what the
// proto already calls the oneof field.
func TableName(md protoreflect.MessageDescriptor) string { return snake(string(md.Name())) }

func tableFor(md protoreflect.MessageDescriptor) (string, error) {
	var cols []string
	var checks []string
	var fks []string
	var children []string

	cols = append(cols, `  "event_id" INTEGER PRIMARY KEY REFERENCES "events"("id")`)

	for i := 0; i < md.Fields().Len(); i++ {
		fd := md.Fields().Get(i)
		if fd.ContainingOneof() != nil && !fd.ContainingOneof().IsSynthetic() {
			continue // handled below, as a group
		}
		if fd.IsList() {
			children = append(children, listTable(md, fd))
			continue
		}
		col, check, err := column(fd)
		if err != nil {
			return "", err
		}
		cols = append(cols, "  "+col)
		if check != "" {
			checks = append(checks, check)
		}
		fk, err := references(fd)
		if err != nil {
			return "", err
		}
		if fk != "" {
			fks = append(fks, fk)
		}
	}

	// REAL oneofs, which proto3 `optional` also uses under the hood — IsSynthetic tells them apart,
	// and conflating the two would turn every optional field into a mutually-exclusive group.
	for i := 0; i < md.Oneofs().Len(); i++ {
		od := md.Oneofs().Get(i)
		if od.IsSynthetic() {
			continue
		}
		oc, ochecks, ochildren, err := oneofColumns(md, od)
		if err != nil {
			return "", err
		}
		cols = append(cols, oc...)
		checks = append(checks, ochecks...)
		children = append(children, ochildren...)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "CREATE TABLE %q (\n%s", TableName(md), strings.Join(cols, ",\n"))
	for _, c := range checks {
		fmt.Fprintf(&b, ",\n  CHECK (%s)", c)
	}
	for _, f := range fks {
		fmt.Fprintf(&b, ",\n  %s", f)
	}
	b.WriteString("\n) STRICT;\n")
	for _, c := range children {
		b.WriteString(c)
	}
	return b.String(), nil
}

// Opts reads the field's own SQL annotation. Nil when it has none, which is the ordinary case: a
// plain optional column with no constraint beyond its type.
func Opts(fd protoreflect.FieldDescriptor) *recordpb.Sql {
	o, _ := proto.GetExtension(fd.Options(), recordpb.E_Sql).(*recordpb.Sql)
	return o
}

// column maps one scalar field. NOT NULL comes from the field's OWN annotation, so the constraint
// and the refusal a seat reads are one statement rather than two artifacts a file apart.
func column(fd protoreflect.FieldDescriptor) (col, check string, err error) {
	name := string(fd.Name())
	typ, err := sqlType(fd)
	if err != nil {
		return "", "", err
	}
	col = fmt.Sprintf("%q %s", name, typ)
	o := Opts(fd)
	if o.GetRequired() {
		col += " NOT NULL"
	}
	if o.GetUnique() {
		col += " UNIQUE"
	}
	if fd.Kind() == protoreflect.EnumKind {
		check = fmt.Sprintf("%q IS NULL OR %q IN (%s)", name, name, enumLiterals(fd.Enum()))
	}
	return col, check, nil
}

// references renders the field's foreign key, if it declared one. "table.column" is split rather
// than passed through so a malformed annotation fails HERE, at derivation, instead of producing a
// schema SQLite refuses at open with a syntax error nobody can trace back to a field.
func references(fd protoreflect.FieldDescriptor) (string, error) {
	ref := Opts(fd).GetReferences()
	if ref == "" {
		return "", nil
	}
	table, col, ok := strings.Cut(ref, ".")
	if !ok || table == "" || col == "" {
		return "", fmt.Errorf("recordsql: %s declares references %q — it must be \"table.column\"", fd.FullName(), ref)
	}
	return fmt.Sprintf("FOREIGN KEY (%q) REFERENCES %q(%q)", fd.Name(), table, col), nil
}

func sqlType(fd protoreflect.FieldDescriptor) (string, error) {
	switch fd.Kind() {
	case protoreflect.StringKind:
		return "TEXT", nil
	case protoreflect.BoolKind, protoreflect.Int32Kind, protoreflect.Int64Kind:
		return "INTEGER", nil
	case protoreflect.BytesKind:
		return "BLOB", nil
	case protoreflect.EnumKind:
		// TEXT AND NOT AN INTEGER, deliberately. The number is an implementation detail of the
		// wire format; the WORD is what every reader and every human already uses, and a database
		// nobody can read with a plain SELECT is one people will build a second view over.
		return "TEXT", nil
	}
	return "", fmt.Errorf("recordsql: %s has no SQL type for kind %s", fd.FullName(), fd.Kind())
}

func enumLiterals(ed protoreflect.EnumDescriptor) string {
	var out []string
	for i := 0; i < ed.Values().Len(); i++ {
		w := recordpb.Word(dummyEnum{ed, ed.Values().Get(i).Number()})
		if w == "" {
			continue // the UNSPECIFIED zero is absence, and absence is NULL here
		}
		out = append(out, "'"+w+"'")
	}
	sort.Strings(out)
	return strings.Join(out, ", ")
}

// listTable keeps a repeated field JOINABLE. Flattening it to a comma string would put a list back
// inside a value, which is where `supersedes` lived before the schema and is the shape that made a
// gap's lineage unqueryable.
func listTable(parent protoreflect.MessageDescriptor, fd protoreflect.FieldDescriptor) string {
	name := TableName(parent) + "_" + string(fd.Name())
	return fmt.Sprintf(`
CREATE TABLE %q (
  "event_id" INTEGER NOT NULL REFERENCES %q("event_id"),
  "ord"      INTEGER NOT NULL,
  "value"    TEXT    NOT NULL,
  PRIMARY KEY ("event_id", "ord")
) STRICT;
`, name, TableName(parent))
}

// oneofColumns models a real oneof.
//
// Scalar and enum arms become nullable columns with a CHECK that at most one is present — the
// schema saying, structurally, that a grade motion cannot carry a petition's verdict.
//
// Message arms cannot be columns, so each becomes its own table and the parent carries a `_case`
// discriminator. The exactly-one rule then spans tables and is enforced on the write instead; that
// is stated here rather than left for a reader to discover it missing.
func oneofColumns(md protoreflect.MessageDescriptor, od protoreflect.OneofDescriptor) (cols, checks, children []string, err error) {
	var scalars []string
	messageArms := false
	for i := 0; i < od.Fields().Len(); i++ {
		fd := od.Fields().Get(i)
		if fd.Message() != nil {
			messageArms = true
			children = append(children, armTable(md, fd))
			continue
		}
		col, check, cerr := column(fd)
		if cerr != nil {
			return nil, nil, nil, cerr
		}
		cols = append(cols, "  "+col)
		if check != "" {
			checks = append(checks, check)
		}
		scalars = append(scalars, fmt.Sprintf("(%q IS NOT NULL)", fd.Name()))
	}
	if len(scalars) > 1 {
		checks = append(checks, strings.Join(scalars, " + ")+" <= 1")
	}
	if messageArms {
		cols = append(cols, fmt.Sprintf("  %q TEXT", string(od.Name())+"_case"))
	}
	return cols, checks, children, nil
}

func armTable(parent protoreflect.MessageDescriptor, fd protoreflect.FieldDescriptor) string {
	md := fd.Message()
	var cols []string
	cols = append(cols, fmt.Sprintf(`  "event_id" INTEGER PRIMARY KEY REFERENCES %q("event_id")`, TableName(parent)))
	var checks []string
	for i := 0; i < md.Fields().Len(); i++ {
		f := md.Fields().Get(i)
		col, check, err := column(f)
		if err != nil {
			continue
		}
		cols = append(cols, "  "+col)
		if check != "" {
			checks = append(checks, check)
		}
	}
	var b strings.Builder
	fmt.Fprintf(&b, "\nCREATE TABLE %q (\n%s", TableName(parent)+"_"+string(fd.Name()), strings.Join(cols, ",\n"))
	for _, c := range checks {
		fmt.Fprintf(&b, ",\n  CHECK (%s)", c)
	}
	b.WriteString("\n) STRICT;\n")
	return b.String()
}

// dummyEnum lets recordpb.Word spell a value we hold only as a descriptor and a number, so the
// CHECK constraint carries the SAME words the readers and the seats use. Spelling them here from
// the constant names would be a second vocabulary.
type dummyEnum struct {
	ed protoreflect.EnumDescriptor
	n  protoreflect.EnumNumber
}

func (d dummyEnum) Descriptor() protoreflect.EnumDescriptor { return d.ed }
func (d dummyEnum) Type() protoreflect.EnumType             { return nil }
func (d dummyEnum) Number() protoreflect.EnumNumber         { return d.n }

func snake(s string) string {
	var b strings.Builder
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				b.WriteByte('_')
			}
			b.WriteRune(r + 32)
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}
