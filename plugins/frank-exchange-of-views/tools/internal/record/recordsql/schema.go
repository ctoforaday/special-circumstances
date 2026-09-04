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
	"strconv"
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
  "ts"      TEXT    NOT NULL,
  "type"    TEXT    NOT NULL REFERENCES "enum_event_type"("value"),
  -- The key is the fact that has to be unique, and the partial index below enforces it globally.
  -- This was UNIQUE (seat_id, nonce, seq): a counter nothing read, scoped by a sitting that no
  -- longer exists. Both are gone.
  "key"     TEXT
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
	// THE VOCABULARIES COME FIRST, because every enum column points at one. Emitting them in a
	// stable order keeps the DDL byte-identical across runs; a map walk here would make every
	// derivation a different string for the same schema.
	vocab, err := enumTables(bodies)
	if err != nil {
		return "", err
	}
	b.WriteString(vocab)
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

// scalarField is EVERY constraint one non-repeated field contributes, in one place.
//
// # Why this is extracted rather than written twice
//
// It was written twice. `tableFor` emitted the column, its CHECK, its `references` foreign key, its
// enum foreign key and its subset CHECK; `armTable` — the same job for a oneof's message arm —
// emitted the column and the CHECK and silently stopped there. The golden schema is what made that
// visible: `motion_grade.dimension`, `motion_grade.proposed` and `motion_petition.class` are enum
// columns sitting in the record with no vocabulary behind them, next to `mint.severity` and
// `motion_rule.binds` which have one. Nothing failed. No test could have failed, because a test
// asserts a constraint somebody thought of, and nobody thinks of the constraint that was never
// emitted.
//
// That is the argument for the golden and the argument for this function at the same time: two
// copies of "what a field owes the schema" will diverge, and the divergence is invisible from the
// behaviour side.
func scalarField(fd protoreflect.FieldDescriptor) (col string, checks []string, fks []string, err error) {
	col, check, err := column(fd)
	if err != nil {
		return "", nil, nil, err
	}
	if check != "" {
		checks = append(checks, check)
	}
	fk, err := references(fd)
	if err != nil {
		return "", nil, nil, err
	}
	if fk != "" {
		fks = append(fks, fk)
	}
	if fd.Kind() == protoreflect.EnumKind {
		fks = append(fks, fmt.Sprintf("FOREIGN KEY (%q) REFERENCES %q(\"value\")", fd.Name(), EnumTableName(fd.Enum())))
		// A column may reach only PART of its vocabulary — `merge close` may close and may not
		// carry. The admitted words are expanded from the values' own facet annotation, so the
		// constraint and the Go refusal beside it read the same declaration.
		if facet := Opts(fd).GetSubset(); facet != "" {
			sc, err := subsetCheck(fd, facet)
			if err != nil {
				return "", nil, nil, err
			}
			checks = append(checks, sc)
		}
	}
	return col, checks, fks, nil
}

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
		col, fchecks, ffks, err := scalarField(fd)
		if err != nil {
			return "", err
		}
		cols = append(cols, "  "+col)
		checks = append(checks, fchecks...)
		fks = append(fks, ffks...)
	}

	// REAL oneofs, which proto3 `optional` also uses under the hood — IsSynthetic tells them apart,
	// and conflating the two would turn every optional field into a mutually-exclusive group.
	for i := 0; i < md.Oneofs().Len(); i++ {
		od := md.Oneofs().Get(i)
		if od.IsSynthetic() {
			continue
		}
		oc, ochecks, ofks, ochildren, err := oneofColumns(md, od)
		if err != nil {
			return "", err
		}
		cols = append(cols, oc...)
		checks = append(checks, ochecks...)
		fks = append(fks, ofks...)
		children = append(children, ochildren...)
	}

	// THE RULES THAT SPAN FIELDS, from the message's own annotation. `required` is a property of one
	// field; "repaired_with_regression requires a successor" is a rule about two, and no annotation on
	// either can say it.
	for _, c := range MessageChecks(md) {
		if c.GetExpr() == "" {
			return "", fmt.Errorf("recordsql: %s declares a check with no expression", md.FullName())
		}
		if c.GetWhy() == "" {
			return "", fmt.Errorf("recordsql: %s declares the check %q and does not say what it protects — "+
				"a constraint that fires with only its expression tells a reader what was refused and nothing "+
				"about which invariant they walked into", md.FullName(), c.GetExpr())
		}
		checks = append(checks, c.GetExpr())
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
	// A BOOLEAN IS A TWO-VALUE SET, and STRICT does not make it one: the column type is INTEGER,
	// so 7 and -1 store happily and every reader downstream treats non-zero as true. The
	// vocabulary tables got this CHECK on their `closes` column and the body tables did not,
	// which is the inconsistency the golden made visible.
	if fd.Kind() == protoreflect.BoolKind {
		check = fmt.Sprintf("%q IS NULL OR %q IN (0, 1)", name, name)
	}
	// NO `IN (…)` CHECK FOR AN ENUM. It points at its VOCABULARY TABLE instead — same enforcement,
	// and the meanings come with it. See enumTable.
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

// enumTable is one vocabulary, as rows.
//
// # Why a table and not a CHECK
//
// `CHECK (x IN ('repaired', 'defect_accepted', …))` enforces the same set and throws away everything
// about it. descriptions.go exists to stop exactly that: "A VALUE CARRIES ITS OWN MEANING, OR THE
// SET IS A LIST OF NOUNS" — the meanings were already lost once, sitting in source comments where
// no seat could read them, and putting only the words into the schema would lose them again at the
// layer that is supposed to be the record.
//
// As a table the vocabulary is IN the database: a human reading the record can join a closure class
// to what it means without leaving SQL, and the foreign key gives the same refusal the CHECK did.
//
// # There are no longer any enums nobody documents
//
// EventType and SchemaVersion used to be exempt, on the reason that no seat types them so no
// --help renders them. True, and scoped to the only consumer `means` had at the time. This table
// is a second consumer with a different audience — a human reading the record in SQL — and
// `events.type` is the column every join keys on. So they are documented, the exemption is gone,
// and an undocumented value is now a hard error rather than a placeholder sentence.
func enumTable(ed protoreflect.EnumDescriptor) (string, error) {
	type row struct {
		word, means string
		facets      map[string]bool
		numbers     map[string]float64
	}
	var rows []row

	// declared is the facets ANY value of this enum carries. A facet is a column only if the
	// vocabulary declares it, so `enum_grade` does not grow a `closes` column it has no opinion on.
	declared := map[string]int{}
	for i := 0; i < ed.Values().Len(); i++ {
		v := ed.Values().Get(i)
		w := recordpb.Word(dummyEnum{ed, v.Number()})
		if w == "" {
			continue // the UNSPECIFIED zero is absence, and absence is NULL here
		}
		// EVERY VALUE CARRIES ITS MEANING NOW, so a miss is an error rather than a placeholder.
		// The fallback string stood in for the EventType/SchemaVersion exemption, which is gone —
		// and a default sentence is the shape this package refuses everywhere else: it puts a
		// plausible value in the record where an unanswered question was.
		means, err := recordpb.EnumValueDoc(v)
		if err != nil {
			return "", fmt.Errorf("recordsql: %s: %w", ed.FullName(), err)
		}
		fs := map[string]bool{}
		ns := map[string]float64{}
		for _, name := range recordpb.FacetNames() {
			// A NUMERIC FACET IS READ AS A NUMBER, and the split is on the facet's declared kind
			// rather than on a guess about its value: `mass` 0 is a real weight (GRADE_REALIZED),
			// so reading it as a flag would turn "weighs nothing" into false and lose it.
			if recordpb.IsNumeric(name) {
				val, ok, err := recordpb.Number(v, name)
				if err != nil {
					return "", err
				}
				if ok {
					ns[name] = val
					declared[name]++
				}
				continue
			}
			val, ok, err := recordpb.Facet(v, name)
			if err != nil {
				return "", err
			}
			if ok {
				fs[name] = val
				declared[name]++
			}
		}
		rows = append(rows, row{w, means, fs, ns})
	}

	// A PARTLY-ANNOTATED FACET IS REFUSED, for the reason the facet exists. `closes` was added
	// because "everything except carried" answered the question for values nobody had asked about;
	// letting one value skip the annotation and defaulting its column would rebuild that exact
	// behaviour inside the schema, where it would be even harder to see.
	var cols []string
	for _, name := range recordpb.FacetNames() {
		n := declared[name]
		if n == 0 {
			continue
		}
		if n != len(rows) {
			return "", fmt.Errorf("recordsql: %s declares `%s` on %d of its %d values — a facet the "+
				"machinery switches on cannot have a default, because the default is an answer given "+
				"on behalf of whoever added the value without one",
				ed.FullName(), name, n, len(rows))
		}
		cols = append(cols, recordpb.FacetColumn(name))
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].word < rows[j].word })

	name := EnumTableName(ed)
	var b strings.Builder
	fmt.Fprintf(&b, "\nCREATE TABLE %q (\n  \"value\" TEXT PRIMARY KEY,\n  \"means\" TEXT NOT NULL", name)
	for _, c := range cols {
		// REAL, NOT INTEGER-WITH-A-CHECK, for a numeric facet: the CHECK that keeps a flag honest
		// (0 or 1) is exactly wrong for a weight, and a mass of 0.5 stored into an INTEGER column
		// under STRICT is a refusal rather than a rounding — which is the good failure, but only
		// if the column was never declared that way in the first place.
		if recordpb.IsNumeric(c) {
			fmt.Fprintf(&b, ",\n  %q REAL NOT NULL", c)
			continue
		}
		fmt.Fprintf(&b, ",\n  %q INTEGER NOT NULL CHECK (%q IN (0, 1))", c, c)
	}
	b.WriteString("\n) STRICT;\n")
	for _, r := range rows {
		lhs, vals := `"value", "means"`, fmt.Sprintf("'%s', '%s'", r.word, escape(r.means))
		for _, c := range cols {
			lhs += fmt.Sprintf(", %q", c)
			if recordpb.IsNumeric(c) {
				vals += ", " + strconv.FormatFloat(r.numbers[c], 'g', -1, 64)
				continue
			}
			vals += fmt.Sprintf(", %d", b2i(r.facets[c]))
		}
		fmt.Fprintf(&b, "INSERT INTO %q (%s) VALUES (%s);\n", name, lhs, vals)
	}
	return b.String(), nil
}

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}

// subsetCheck turns `subset: "closes"` into a CHECK over the words that actually carry the facet.
//
// The list is EXPANDED FROM THE DESCRIPTOR at build time, never typed. The alternative that was
// nearly written — `CHECK ("closure_class" <> 'carried')` — encodes the answer to "which words may
// a merge write" as a single literal in an expression, and nothing anywhere could notice when a
// second non-closing word joined the vocabulary. That is the same defect as the negative predicate
// this whole change removed, moved into SQL.
func subsetCheck(fd protoreflect.FieldDescriptor, facet string) (string, error) {
	ed := fd.Enum()
	if ed == nil {
		return "", fmt.Errorf("recordsql: %s declares subset %q but is not an enum — a subset of what?", fd.FullName(), facet)
	}
	var words []string
	for i := 0; i < ed.Values().Len(); i++ {
		v := ed.Values().Get(i)
		w := recordpb.Word(dummyEnum{ed, v.Number()})
		if w == "" {
			continue
		}
		val, ok, err := recordpb.Facet(v, facet)
		if err != nil {
			return "", fmt.Errorf("recordsql: %s: %w", fd.FullName(), err)
		}
		if !ok {
			return "", fmt.Errorf("recordsql: %s restricts to `%s` but %s does not declare it — the "+
				"subset would silently exclude a value nobody ruled on", fd.FullName(), facet, v.Name())
		}
		if val {
			words = append(words, "'"+w+"'")
		}
	}
	if len(words) == 0 {
		return "", fmt.Errorf("recordsql: %s restricts to `%s` and no value in %s carries it — the "+
			"column would admit nothing, which is a constraint that reads as strict and is simply broken",
			fd.FullName(), facet, ed.FullName())
	}
	sort.Strings(words)
	// The bare expression: the caller wraps it, and wrapping it here produced `CHECK (CHECK (...))`.
	return fmt.Sprintf("%q IS NULL OR %q IN (%s)", fd.Name(), fd.Name(), strings.Join(words, ", ")), nil
}

// EnumTableName is the vocabulary table for an enum: `enum_` plus its own name in snake_case.
func EnumTableName(ed protoreflect.EnumDescriptor) string {
	return "enum_" + snake(string(ed.Name()))
}

func escape(s string) string { return strings.ReplaceAll(s, "'", "''") }

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
// A SCALAR ONEOF ARM IS STILL A FIELD, AND OWES THE SCHEMA WHAT EVERY OTHER FIELD OWES.
//
// This was the third copy of "what a field contributes", and the one my own reading of the golden
// missed: `motion_rule`'s grade, petition and direction arms ARE the bench's ruling on a motion,
// and all three sat unconstrained beside `binds`, which was constrained, in the same table. The
// structural test found them after the golden showed me the message-arm hole — which is the honest
// order of events and the reason both exist.
func oneofColumns(md protoreflect.MessageDescriptor, od protoreflect.OneofDescriptor) (cols, checks, fks, children []string, err error) {
	var scalars []string
	messageArms := false
	for i := 0; i < od.Fields().Len(); i++ {
		fd := od.Fields().Get(i)
		if fd.Message() != nil {
			messageArms = true
			arm, aerr := armTable(md, fd)
			if aerr != nil {
				return nil, nil, nil, nil, aerr
			}
			children = append(children, arm)
			continue
		}
		col, fchecks, ffks, cerr := scalarField(fd)
		if cerr != nil {
			return nil, nil, nil, nil, cerr
		}
		cols = append(cols, "  "+col)
		checks = append(checks, fchecks...)
		fks = append(fks, ffks...)
		scalars = append(scalars, fmt.Sprintf("(%q IS NOT NULL)", fd.Name()))
	}
	if len(scalars) > 1 {
		checks = append(checks, strings.Join(scalars, " + ")+" <= 1")
	}
	if messageArms {
		cols = append(cols, fmt.Sprintf("  %q TEXT", string(od.Name())+"_case"))
	}
	return cols, checks, fks, children, nil
}

// armTable is a oneof arm's own table, and it owes the schema everything a body's table does.
//
// IT DID NOT. It emitted columns and CHECKs and dropped foreign keys on the floor — the golden
// schema is where that became visible, as three enum columns (`motion_grade.dimension`,
// `motion_grade.proposed`, `motion_petition.class`) sitting with no vocabulary behind them beside
// `motion_rule.binds`, which has one. The arms are where a motion's SUBSTANCE lives, so the record
// enforced its vocabulary everywhere except on the part a seat actually argues about.
//
// It also swallowed a column error with `continue`, which is worse than dropping the constraint: a
// field the generator could not map simply vanished from the table, and the schema built clean with
// a column missing. That reads, from every other test in this package, exactly like a schema that
// has no such field.
func armTable(parent protoreflect.MessageDescriptor, fd protoreflect.FieldDescriptor) (string, error) {
	md := fd.Message()
	var cols []string
	var checks []string
	var fks []string
	cols = append(cols, fmt.Sprintf(`  "event_id" INTEGER PRIMARY KEY REFERENCES %q("event_id")`, TableName(parent)))
	for i := 0; i < md.Fields().Len(); i++ {
		f := md.Fields().Get(i)
		col, fchecks, ffks, err := scalarField(f)
		if err != nil {
			return "", fmt.Errorf("recordsql: %s (arm %s of %s): %w", f.FullName(), fd.Name(), TableName(parent), err)
		}
		cols = append(cols, "  "+col)
		checks = append(checks, fchecks...)
		fks = append(fks, ffks...)
	}
	var b strings.Builder
	fmt.Fprintf(&b, "\nCREATE TABLE %q (\n%s", TableName(parent)+"_"+string(fd.Name()), strings.Join(cols, ",\n"))
	for _, c := range checks {
		fmt.Fprintf(&b, ",\n  CHECK (%s)", c)
	}
	for _, f := range fks {
		fmt.Fprintf(&b, ",\n  %s", f)
	}
	b.WriteString("\n) STRICT;\n")
	return b.String(), nil
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

// enumTables emits one vocabulary table per enum any column uses, once each and in a stable order.
func enumTables(bodies []protoreflect.MessageDescriptor) (string, error) {
	var order []protoreflect.EnumDescriptor
	seen := map[protoreflect.FullName]bool{}
	var walk func(md protoreflect.MessageDescriptor)
	walk = func(md protoreflect.MessageDescriptor) {
		for i := 0; i < md.Fields().Len(); i++ {
			fd := md.Fields().Get(i)
			if m := fd.Message(); m != nil {
				walk(m) // the oneof's message arms get their own tables, and their enums too
				continue
			}
			if fd.Kind() != protoreflect.EnumKind {
				continue
			}
			ed := fd.Enum()
			if seen[ed.FullName()] {
				continue
			}
			seen[ed.FullName()] = true
			order = append(order, ed)
		}
	}
	// THE ENVELOPE'S OWN VOCABULARY, which no body column reaches.
	//
	// `events.type` is written by the tool, not by a seat, so the walk above never sees it — and it
	// was therefore the only enum-valued column in the record with no wall behind it, in the column
	// every join keys on. It is seeded explicitly rather than inferred, because "the walk did not
	// find it" and "no such vocabulary is wanted" are the same silence.
	for _, ed := range []protoreflect.EnumDescriptor{
		recordpb.EventType(0).Descriptor(),
		recordpb.SchemaVersion(0).Descriptor(),
	} {
		if !seen[ed.FullName()] {
			seen[ed.FullName()] = true
			order = append(order, ed)
		}
	}
	for _, md := range bodies {
		walk(md)
	}
	var b strings.Builder
	for _, ed := range order {
		t, err := enumTable(ed)
		if err != nil {
			return "", err
		}
		b.WriteString(t)
	}
	return b.String(), nil
}

// MessageChecks reads the table-level rules a message declares.
func MessageChecks(md protoreflect.MessageDescriptor) []*recordpb.SqlCheck {
	out, _ := proto.GetExtension(md.Options(), recordpb.E_Check).([]*recordpb.SqlCheck)
	return out
}
