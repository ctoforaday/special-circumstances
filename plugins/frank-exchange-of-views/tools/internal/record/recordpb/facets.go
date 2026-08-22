package recordpb

import (
	"fmt"
	"sort"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// A VALUE FACET IS A COLUMN OF THE VOCABULARY, AND THIS IS ITS ONE INTERPRETER.
//
// `means` says what a word MEANS; a facet says what the record DOES about it. `closes` is the
// first: whether a disposition ends the gap. The distinction matters because the meaning is for a
// seat and the facet is for the machinery — replay folds on it, the schema builds a CHECK from it,
// and the gap view joins on it.
//
// # Why a name, and what a wrong name does
//
// `subset: "closes"` names the facet in a string, and a string naming a thing is the shape
// facts-are-fields is about. It is bounded here: the name resolves through THIS map, and a name
// with no entry is an ERROR at schema build — not a subset that silently admits everything, which
// is what a `nil` lookup would have produced. The miss is loud, which is the property that matters;
// the alternative (a field per facet on the Sql message) makes the schema generator know every
// facet by name in Go instead, which moves the coupling without removing it.
var facets = map[string]protoreflect.ExtensionType{
	"closes": E_Closes,
}

// FacetNames lists the declared facets, for refusals that have to say what WOULD have worked.
func FacetNames() []string {
	out := make([]string, 0, len(facets))
	for n := range facets {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}

// Facet reads a boolean facet off one value. It returns whether the value DECLARED it, because
// undeclared and false are different states: false is an answer, absent is nobody asked — and
// "nobody asked" is precisely how `grade_adjusted` came to close a gap the bench had deferred.
func Facet(vd protoreflect.EnumValueDescriptor, name string) (value bool, declared bool, err error) {
	xt, ok := facets[name]
	if !ok {
		return false, false, fmt.Errorf("recordpb: %q is not a declared value facet (have %v) — "+
			"an unresolved facet name would silently admit every value, so it refuses instead", name, FacetNames())
	}
	opts := vd.Options()
	if opts == nil || !proto.HasExtension(opts, xt) {
		return false, false, nil
	}
	return proto.GetExtension(opts, xt).(bool), true, nil
}

// FacetColumn maps a facet to the vocabulary-table column that carries it. Same name, stated once
// so the DDL and the queries that join on it cannot disagree about spelling.
func FacetColumn(name string) string { return name }

// Closes answers whether a disposition ENDS the gap.
//
// It REPLACES `benchClosesGap`, whose rule was "everything except carried closes". That shape has
// no gap to notice: a value added to the set was classified as closing by DEFAULT, and no test of
// the predicate's stated behaviour could fail. Measured — `grade_adjusted` was added and read as
// closing a gap the bench had explicitly deferred.
//
// Here the answer comes off the value, so a value that never answered is not silently closing: it
// is a vocabulary the schema refuses to build (see the partly-annotated refusal in recordsql).
func Closes(d Disposition) bool {
	vd := d.Descriptor().Values().ByNumber(d.Number())
	if vd == nil {
		return false
	}
	v, _, err := Facet(vd, "closes")
	return err == nil && v
}
