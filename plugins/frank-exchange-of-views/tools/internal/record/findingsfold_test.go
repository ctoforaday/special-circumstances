package record

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// A FINDING HAS NO CARRIER LEFT IN THE RECORD (plans/findings-fold.md Part A, T-A2): no table, no
// list table, no anchor kind, and each number it held is reserved, so a stored event cannot hand its
// bytes to a later field.
func TestTheRecordHoldsNoFinding(t *testing.T) {
	ddl, err := recordsql.Schema()
	if err != nil {
		t.Fatal(err)
	}
	for _, table := range []string{`"finding"`, `"mint_found_by"`} {
		if strings.Contains(ddl, "CREATE TABLE "+table) || strings.Contains(ddl, "ON "+table) {
			t.Errorf("the schema still declares %s", table)
		}
	}
	if strings.Contains(ddl, "found_by_count") {
		t.Error("the schema still carries found_by_count")
	}
	if k := anchor.Kind("F-0123abcd"); k != "" {
		t.Errorf("an F- id is an anchor of kind %q, want none", k)
	}
	reserved := func(what string, names protoreflect.Names, numbers interface {
		Has(protoreflect.FieldNumber) bool
	}, name string, n protoreflect.FieldNumber) {
		t.Helper()
		if !names.Has(protoreflect.Name(name)) || !numbers.Has(n) {
			t.Errorf("%s %s (%d) is not reserved", what, name, n)
		}
	}
	et := recordpb.EventType(0).Descriptor()
	if et.Values().ByNumber(11) != nil || !et.ReservedNames().Has("EVENT_TYPE_FINDING") || !et.ReservedRanges().Has(11) {
		t.Error("EVENT_TYPE_FINDING (11) is not reserved")
	}
	mint := (&recordpb.Mint{}).ProtoReflect().Descriptor()
	reserved("Mint field", mint.ReservedNames(), mint.ReservedRanges(), "found_by", 21)
	ev := (&recordpb.Event{}).ProtoReflect().Descriptor()
	reserved("Event body", ev.ReservedNames(), ev.ReservedRanges(), "finding", 37)
	if _, ok := recordpb.BodyField("finding"); ok {
		t.Error("the event still has a finding body arm")
	}
}

// NO ITEM KIND IS A FINDING'S (T-A4): another seat's point on a gap is a proposed act on it, which
// reaches the gap's assignee as its own item, so no work list carries a finding about a gap.
func TestNoItemKindIsAFindings(t *testing.T) {
	for _, k := range ItemKinds() {
		if strings.Contains(string(k), "finding") {
			t.Errorf("item kind %q names a finding", k)
		}
	}
	for _, r := range blockerKinds {
		if strings.Contains(string(r.kind), "finding") {
			t.Errorf("PASS blocker kind %q names a finding", r.kind)
		}
	}
}
