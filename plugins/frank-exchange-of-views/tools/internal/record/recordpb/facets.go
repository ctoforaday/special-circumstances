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
// A FACET IS TYPED BY ITS EXTENSION, and the interpreter now carries two kinds because the second
// one arrived: `closes` answers a yes/no about a word, `mass` says what a word WEIGHS. Both are
// "what the record DOES about it" — the distinction from `means` is unchanged — so they share the
// registry, the partly-annotated refusal, and the column-name rule, and differ only in the Go type
// a reader gets back.
var facets = map[string]protoreflect.ExtensionType{
	"closes": E_Closes,
	"mass":   E_Mass,
}

// numericFacets are the facets whose values are numbers rather than flags. The schema generator
// asks this to pick a column type, so a facet added to `facets` and forgotten here becomes a
// column of the wrong type rather than a silent success.
var numericFacets = map[string]bool{"mass": true}

// IsNumeric reports whether a declared facet carries a number.
func IsNumeric(name string) bool { return numericFacets[name] }

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

// Number reads a NUMERIC facet off one value, with the same declared/undeclared split Facet
// makes and for the same reason: 0 is a real weight here (GRADE_REALIZED contributes zero mass by
// design), so "weighs nothing" and "nobody said" must not arrive as the same answer.
func Number(vd protoreflect.EnumValueDescriptor, name string) (value float64, declared bool, err error) {
	xt, ok := facets[name]
	if !ok {
		return 0, false, fmt.Errorf("recordpb: %q is not a declared value facet (have %v)", name, FacetNames())
	}
	if !numericFacets[name] {
		return 0, false, fmt.Errorf("recordpb: facet %q is not numeric — read it with Facet", name)
	}
	opts := vd.Options()
	if opts == nil || !proto.HasExtension(opts, xt) {
		return 0, false, nil
	}
	return proto.GetExtension(opts, xt).(float64), true, nil
}

// GradeMass is what one grade weighs, off the vocabulary rather than out of a hand-kept map.
//
// THE MAP THIS REPLACES HAD A TWIN IN JAVASCRIPT and a regex parity test to keep them equal —
// a guard built because they had already drifted. The weight is now a property of the word, so
// there is one source and the database can join on it.
func GradeMass(g Grade) float64 {
	vd := g.Descriptor().Values().ByNumber(g.Number())
	if vd == nil {
		return 0
	}
	v, _, err := Number(vd, "mass")
	if err != nil {
		return 0
	}
	return v
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
