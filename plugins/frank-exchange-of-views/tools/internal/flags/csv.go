package flags

import "strings"

// CSV is a comma-separated list flag: --supersedes R1-2,R1-7 and --found-by
// L5-F3,L6-F2 and --ids R1-4,R2-7.
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

func (c *CSV) Set(s string) error {
	// DEFECT FIXED: this was `c.items = c.items[:0]`, which truncates but REUSES
	// the backing array. Value() hands that array out without copying, and
	// seat.SetList stores the result straight into the record payload — so a
	// second Set (pflag calls Set once per OCCURRENCE, and a repeated
	// --supersedes is a seat retrying its own command) rewrote a slice the
	// payload was already holding, in place, after the fact. The record would
	// then serialize a lineage the seat never passed, with no event marking the
	// change. Allocating fresh makes each Set's result immutable to the next.
	c.items = nil
	for _, part := range strings.Split(s, ",") {
		if t := strings.TrimSpace(part); t != "" {
			c.items = append(c.items, t)
		}
	}
	c.set = true
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
