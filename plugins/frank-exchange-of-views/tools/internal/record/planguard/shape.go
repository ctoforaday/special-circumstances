package planguard

import (
	"fmt"
	"sort"
	"strings"
)

// ShapeKind is a plan shape whose cost grows faster than the run.
type ShapeKind string

const (
	// AutomaticIndex is SQLite building a throwaway index for this statement: the index the
	// statement needs is missing from the schema, and every execution pays to build it.
	AutomaticIndex ShapeKind = "automatic-index"
	// CorrelatedScan is a heap scan of a growing table, or of a CTE or view built from one,
	// beneath a correlated subquery: the whole table is walked once per outer row.
	CorrelatedScan ShapeKind = "correlated-scan"
	// RepeatedSubplan is one CTE or view materialized or co-routined more than once in a plan:
	// the same rows are computed again for each appearance.
	RepeatedSubplan ShapeKind = "repeated-subplan"
	// ConstrainedScan is a statement that narrows what it reads (WHERE or JOIN) whose plan still
	// walks the whole of a growing table: the signature of a missing index. rule.go says why the
	// constraint is what separates it from a bulk read that must scan.
	ConstrainedScan ShapeKind = "constrained-scan"
	// CorrelatedSort is a temporary b-tree beneath a correlated subquery: a sort per outer row.
	CorrelatedSort ShapeKind = "correlated-sort"
)

// Shape is one finding in one plan. Count is how many plan nodes show it, so a fix that removes
// one of three appearances changes the finding rather than leaving it as it was.
type Shape struct {
	Kind  ShapeKind
	Name  string
	Count int
}

func (s Shape) String() string { return fmt.Sprintf("%s | %s | x%d", s.Kind, s.Name, s.Count) }

// Analyzer reads shapes out of plans of statements against one schema.
type Analyzer struct {
	views  []View
	prefix string          // the enum vocabularies' table-name prefix, read from the generator
	growth map[string]bool // the schema's growing tables (GrowthTables)
}

// NewAnalyzer returns an analyzer over the database's views.
func NewAnalyzer(views []View) (*Analyzer, error) {
	prefix, err := enumPrefix()
	if err != nil {
		return nil, err
	}
	growth, err := GrowthTables()
	if err != nil {
		return nil, err
	}
	return &Analyzer{views: views, prefix: prefix, growth: growth}, nil
}

// Shapes is every finding in p, sorted by kind and name. stmt is the SQL p is the plan of; its
// text, and the views it names, are where the plan's aliases are bound (see Scope).
func (a *Analyzer) Shapes(stmt string, p Plan) []Shape {
	scope := NewScope(stmt, a.views)
	counts := map[[2]string]int{}
	add := func(k ShapeKind, name string) { counts[[2]string{string(k), name}]++ }
	subplans := map[string]int{}
	constrained := constrains(stmt)
	built := subplanNodes(p)

	p.Walk(func(n *PlanNode, _ int) {
		f := strings.Fields(n.Detail)
		if len(f) < 2 {
			return
		}
		if constrained && heapScan(f) && (a.grows(scope.Targets(f[1])) || minMaxOverSubplan(f, built, a, scope)) {
			add(ConstrainedScan, scanName(f, scope))
		}
		switch {
		case (f[0] == "MATERIALIZE" || f[0] == "CO-ROUTINE") && !strings.HasPrefix(f[1], "("):
			// SQLite names a materialized or co-routined FROM item by its CTE or view name
			// where it has one, not by its alias, so the word needs no resolving.
			subplans[f[1]]++
		case automaticIndex(f):
			add(AutomaticIndex, scope.Name(f[1])+" "+constraintOf(f))
		case heapScan(f) && perRow(n) && !a.bounded(scope.Targets(f[1])):
			add(CorrelatedScan, scanName(f, scope))
		case f[0] == "USE" && len(f) >= 3 && f[1] == "TEMP" && f[2] == "B-TREE" && perRow(n):
			add(CorrelatedSort, sortName(f, n, scope))
		}
	})
	for name, c := range subplans {
		if c > 1 {
			counts[[2]string{string(RepeatedSubplan), name}] = c
		}
	}

	out := make([]Shape, 0, len(counts))
	for k, c := range counts {
		out = append(out, Shape{Kind: ShapeKind(k[0]), Name: k[1], Count: c})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Kind != out[j].Kind {
			return out[i].Kind < out[j].Kind
		}
		return out[i].Name < out[j].Name
	})
	return out
}

// bounded reports whether every table the word may stand for is an enum vocabulary. A CTE, a
// view or a subquery is built from the record and grows with it.
func (a *Analyzer) bounded(targets []string) bool {
	for _, t := range targets {
		if !strings.HasPrefix(t, a.prefix) {
			return false
		}
	}
	return len(targets) > 0
}

// grows reports whether any table the word may stand for is one of the schema's growing tables.
// An ambiguous alias counts if any candidate grows: the finding is then read, not dropped.
func (a *Analyzer) grows(targets []string) bool {
	for _, t := range targets {
		if a.growth[t] {
			return true
		}
	}
	return false
}

// subplanNodes is every named CTE or view the plan materializes or co-routines, by the name SQLite
// gives it.
func subplanNodes(p Plan) map[string][]*PlanNode {
	out := map[string][]*PlanNode{}
	p.Walk(func(n *PlanNode, _ int) {
		f := strings.Fields(n.Detail)
		if len(f) >= 2 && (f[0] == "MATERIALIZE" || f[0] == "CO-ROUTINE") && !strings.HasPrefix(f[1], "(") {
			out[f[1]] = append(out[f[1]], n)
		}
	})
	return out
}

// minMaxOverSubplan reads a min()/max() loop over a CTE or view SQLite computes as a subplan:
// "SEARCH v" beside "CO-ROUTINE v" or "MATERIALIZE v". A subplan's output has no index, so the loop
// cannot stop early — every row the subplan yields is produced and compared, and when the subplan
// reads a growing table that is a cost that rises with the run. The subplan's own loops do not
// always show it: built from index walks ("SCAN events USING COVERING INDEX events_type"), which
// heapScan rightly does not count, the whole walk sits under a min/max that reads as one row. A
// SCAN of a subplan is not judged here: a filter pushed into the subplan can make that loop cheap
// (a count over a view whose arms are rowid lookups), and its cost is read where the subplan runs.
func minMaxOverSubplan(f []string, built map[string][]*PlanNode, a *Analyzer, scope *Scope) bool {
	if f[0] != "SEARCH" {
		return false
	}
	return a.subplanGrows(f[1], built, scope, map[string]bool{})
}

// subplanGrows reports whether any loop inside the named subplan reads a growing table, directly
// or through another subplan.
func (a *Analyzer) subplanGrows(name string, built map[string][]*PlanNode, scope *Scope, seen map[string]bool) bool {
	if seen[name] {
		return false
	}
	seen[name] = true
	for _, root := range built[name] {
		found := false
		Plan{Roots: root.Children}.Walk(func(n *PlanNode, _ int) {
			f := strings.Fields(n.Detail)
			if found || len(f) < 2 || (f[0] != "SCAN" && f[0] != "SEARCH") || strings.HasPrefix(f[1], "(") {
				return
			}
			if strings.Contains(n.Detail, "(rowid=?)") {
				return // one row by its key, whatever the table's size
			}
			found = a.grows(scope.Targets(f[1])) || a.subplanGrows(f[1], built, scope, seen)
		})
		if found {
			return true
		}
	}
	return false
}

// automaticIndex reads "SEARCH ss USING AUTOMATIC [PARTIAL] [COVERING] INDEX (id=?)".
func automaticIndex(f []string) bool {
	for i := 0; i+1 < len(f); i++ {
		if f[i] == "USING" && f[i+1] == "AUTOMATIC" {
			return true
		}
	}
	return false
}

// constraintOf is the "(col=?)" an index lookup is keyed on.
func constraintOf(f []string) string {
	for _, w := range f[2:] {
		if strings.HasPrefix(w, "(") {
			return w
		}
	}
	return "()"
}

// heapScan reads a loop that walks a table's rows rather than an index: "SCAN t", and also
// "SEARCH t" with no USING — SQLite says SEARCH for a min()/max() loop it may stop early, and the
// plan does not say whether it can. Measured on this driver's SQLite 3.54.1 over 1,000,000 rows:
// max(id) on the rowid answers in 26us, max(w) on an unindexed column walks every row (113ms), and
// max(id) with a filter on an unindexed column walks back from the end until a row matches — one
// row when matches are dense, the whole table when there are none. The cost is the data's, so the
// loop is reported, and scanName marks it so the reader can tell which form it is.
func heapScan(f []string) bool {
	if f[0] != "SCAN" && f[0] != "SEARCH" {
		return false
	}
	if strings.HasPrefix(f[1], "(") || f[1] == "CONSTANT" {
		return false // a co-routine's output, whose own scans are judged where they run, or a single row
	}
	for _, w := range f[2:] {
		if w == "USING" || w == "VIRTUAL" {
			return false
		}
	}
	return true
}

// scanName names a heap scan by its table, marking the min()/max() SEARCH form (see heapScan).
func scanName(f []string, scope *Scope) string {
	name := scope.Name(f[1])
	if f[0] == "SEARCH" {
		return name + " [min/max]"
	}
	return name
}

// perRow reports whether the node runs once per row of an outer query: its nearest ancestor that
// decides how often it runs is a CORRELATED subquery. A MATERIALIZE between them is computed once
// and reused, so its contents do not repeat.
func perRow(n *PlanNode) bool {
	for up := n.parent; up != nil; up = up.parent {
		switch first(up.Detail) {
		case "CORRELATED":
			return true
		case "MATERIALIZE":
			return false
		}
	}
	return false
}

// sortName names a temporary b-tree by what it is for and what the loop beside it reads, so two
// sorts in one plan are two findings and fixing one shows.
func sortName(f []string, n *PlanNode, scope *Scope) string {
	purpose := strings.Join(f[3:], " ")
	seen := map[string]bool{}
	var over []string
	if n.parent != nil {
		for _, sib := range n.parent.Children {
			sf := strings.Fields(sib.Detail)
			if len(sf) < 2 || (sf[0] != "SCAN" && sf[0] != "SEARCH") || strings.HasPrefix(sf[1], "(") {
				continue
			}
			name := scope.Name(sf[1])
			if !seen[name] {
				seen[name] = true
				over = append(over, name)
			}
		}
	}
	sort.Strings(over)
	if len(over) == 0 {
		return purpose
	}
	return purpose + " over " + strings.Join(over, ",")
}

func first(detail string) string {
	f := strings.Fields(detail)
	if len(f) == 0 {
		return ""
	}
	return f[0]
}
