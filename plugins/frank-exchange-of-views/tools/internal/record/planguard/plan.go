package planguard

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"
)

// PlanRow is one row of EXPLAIN QUERY PLAN as SQLite returns it: the node's id, the id of the
// node it sits under (0 for the top level), and the node's own words.
type PlanRow struct {
	ID, Parent int64
	Detail     string
}

// PlanNode is one node of a statement's plan, with the nodes SQLite placed beneath it.
type PlanNode struct {
	PlanRow
	parent   *PlanNode
	Children []*PlanNode
}

// Plan is a statement's whole plan as a tree. The tree, not the line, is the unit a shape is read
// from: "this scan runs once per outer row" is a fact about the scan's ANCESTORS, which a
// line-by-line reading cannot see.
type Plan struct {
	Roots []*PlanNode
}

// BuildPlan assembles the rows into a tree.
//
// A row naming a parent that is not in the plan, or two rows sharing an id, is an error rather
// than a node dropped or re-rooted: a re-rooted scan loses its correlated ancestor and every
// per-row finding beneath it would read as a clean plan.
func BuildPlan(rows []PlanRow) (Plan, error) {
	if len(rows) == 0 {
		return Plan{}, fmt.Errorf("the plan has no rows — SQLite plans every statement it accepts, so an empty plan means nothing was read")
	}
	byID := map[int64]*PlanNode{}
	nodes := make([]*PlanNode, 0, len(rows))
	for _, r := range rows {
		if _, dup := byID[r.ID]; dup {
			return Plan{}, fmt.Errorf("two plan rows share id %d", r.ID)
		}
		n := &PlanNode{PlanRow: r}
		byID[r.ID] = n
		nodes = append(nodes, n)
	}
	var p Plan
	for _, n := range nodes {
		if n.Parent == 0 {
			p.Roots = append(p.Roots, n)
			continue
		}
		up, ok := byID[n.Parent]
		if !ok {
			return Plan{}, fmt.Errorf("plan row %d (%q) names parent %d, which is not in the plan", n.ID, n.Detail, n.Parent)
		}
		n.parent = up
		up.Children = append(up.Children, n)
	}
	return p, nil
}

// Walk visits every node depth-first, in the order SQLite emitted them.
func (p Plan) Walk(visit func(n *PlanNode, depth int)) {
	var rec func(ns []*PlanNode, depth int)
	rec = func(ns []*PlanNode, depth int) {
		for _, n := range ns {
			visit(n, depth)
			rec(n.Children, depth+1)
		}
	}
	rec(p.Roots, 0)
}

// Render is the plan as indented text with SQLite's subquery numbering removed.
//
// The numbers ("CORRELATED SCALAR SUBQUERY 7", "(subquery-52)") count subqueries in parse order
// across the statement and every view it expands, so an unrelated edit anywhere upstream renumbers
// every line below it. The tree's shape carries what the number identified.
func (p Plan) Render() string {
	var b strings.Builder
	p.Walk(func(n *PlanNode, depth int) {
		b.WriteString(strings.Repeat("  ", depth+1))
		b.WriteString(NormalizeDetail(n.Detail))
		b.WriteByte('\n')
	})
	return b.String()
}

// NormalizeDetail removes the subquery and join numbering from one plan node's words.
func NormalizeDetail(detail string) string {
	fields := strings.Fields(detail)
	out := fields[:0]
	for i := 0; i < len(fields); i++ {
		f := fields[i]
		if f == "SUBQUERY" && i+1 < len(fields) && isDigits(fields[i+1]) {
			out = append(out, f)
			i++
			continue
		}
		for _, kind := range []string{"subquery", "join"} {
			if rest, ok := strings.CutPrefix(f, "("+kind+"-"); ok {
				if num, tail, ok := strings.Cut(rest, ")"); ok && isDigits(num) {
					f = "(" + kind + ")" + tail
				}
			}
		}
		out = append(out, f)
	}
	return strings.Join(out, " ")
}

func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

// PlanQuery plans one statement on db without running it.
func PlanQuery(ctx context.Context, db *sql.DB, query string) (Plan, error) {
	rows, err := db.QueryContext(ctx, "EXPLAIN QUERY PLAN "+query)
	if err != nil {
		return Plan{}, fmt.Errorf("planning %q: %w", query, err)
	}
	defer rows.Close()
	var got []PlanRow
	for rows.Next() {
		var r PlanRow
		var notused int64
		if err := rows.Scan(&r.ID, &r.Parent, &notused, &r.Detail); err != nil {
			return Plan{}, fmt.Errorf("reading the plan of %q: %w", query, err)
		}
		got = append(got, r)
	}
	if err := rows.Err(); err != nil {
		return Plan{}, fmt.Errorf("reading the plan of %q: %w", query, err)
	}
	p, err := BuildPlan(got)
	if err != nil {
		return Plan{}, fmt.Errorf("the plan of %q: %w", query, err)
	}
	return p, nil
}

// View is one view the database declares, with the SQL that defines it.
type View struct {
	Name, SQL string
}

// Views reads every view out of sqlite_master, sorted by name. The set is the database's own
// answer, so a view added to the schema is planned on the day it is added.
func Views(ctx context.Context, db *sql.DB) ([]View, error) {
	rows, err := db.QueryContext(ctx, `SELECT "name", "sql" FROM "sqlite_master" WHERE "type" = 'view'`)
	if err != nil {
		return nil, fmt.Errorf("listing the views: %w", err)
	}
	defer rows.Close()
	var out []View
	for rows.Next() {
		var v View
		var def sql.NullString
		if err := rows.Scan(&v.Name, &def); err != nil {
			return nil, fmt.Errorf("listing the views: %w", err)
		}
		if !def.Valid || strings.TrimSpace(def.String) == "" {
			return nil, fmt.Errorf("view %q has no SQL in sqlite_master, so its aliases cannot be read", v.Name)
		}
		v.SQL = def.String
		out = append(out, v)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("listing the views: %w", err)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("the database declares no views — this is not a record, and planning it measures nothing")
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

// ViewQuery is the statement that plans a view on its own.
func ViewQuery(name string) string {
	return `SELECT * FROM "` + strings.ReplaceAll(name, `"`, `""`) + `"`
}
