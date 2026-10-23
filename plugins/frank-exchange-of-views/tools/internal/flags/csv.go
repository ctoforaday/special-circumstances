package flags

import (
	"fmt"
	"strings"
)

// CSV is a comma-separated list flag: --supersedes G2,G7 and --found-by
// L5-F3,L6-F2 and --ids G4,G7.
//
// It is a type rather than a strings.Split at each call site because the record
// format has an opinion the split does not: these fields are ALWAYS present in
// the event, as an array, even when empty. A gap with no ancestors records
// "supersedes": [] rather than omitting the key — absence would read as "lineage
// unknown" where the truth is "lineage none", and lineage is the field the whole
// supersedes-chain analysis is computed from.
//
// So Value() always returns a non-nil slice, and Given() answers the separate
// question of whether the seat passed the flag at all.
type CSV struct {
	items []string
	set   bool
}

func (c *CSV) String() string {
	if c == nil {
		return ""
	}
	return strings.Join(c.items, ",")
}

// Set takes the flag's one value. A SECOND OCCURRENCE IS REFUSED, never merged and never
// replaced: pflag calls Set once per occurrence, and keeping either list would record a value the
// seat stated and drop another it also stated. AN EMPTY ITEM IS REFUSED for the same reason —
// "G1,,G2" and "G1," name an item that is not there, and dropping it silently records a list other
// than the one typed. The empty value is the list with no items, stated. Space around an item is
// not part of it.
func (c *CSV) Set(s string) error {
	if c.set {
		return fmt.Errorf("given more than once (%q after %q): pass one comma-separated list", s, strings.Join(c.items, ","))
	}
	// A FRESH SLICE: Value() hands it out without copying, and seat.SetList stores it straight into
	// the record payload.
	var items []string
	if strings.TrimSpace(s) == "" {
		c.items, c.set = items, true
		return nil
	}
	for _, part := range strings.Split(s, ",") {
		t := strings.TrimSpace(part)
		if t == "" {
			return fmt.Errorf("%q holds an empty item: name each item, separated by single commas", s)
		}
		items = append(items, t)
	}
	c.items, c.set = items, true
	return nil
}

func (c *CSV) Type() string { return "list" }

// Value is always a real slice, never nil — see the type comment.
func (c *CSV) Value() []string {
	if c == nil || c.items == nil {
		return []string{}
	}
	return c.items
}

func (c *CSV) Given() bool { return c != nil && c.set }
