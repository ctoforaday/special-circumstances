package planguard

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
)

// Finding is one statement that scanned one growing table.
type Finding struct {
	Statement string // the SQL as the code wrote it
	Table     string // the growth table the plan scanned
	Detail    string // EXPLAIN QUERY PLAN's own words, kept verbatim
}

// Recorder collects findings across every connection a guarded database opens.
//
// It records rather than fails. What counts as a DEFECT is a judgement about the statement — a
// full replay is supposed to scan `events` — and that judgement belongs to the caller with the
// statement in front of it, not to the interceptor. Mixing the two would bury a policy decision
// inside a driver.
type Recorder struct {
	mu        sync.Mutex
	findings  []Finding
	seen      map[string]bool // one report per (statement, table), however often it runs
	growth    map[string]bool
	explained map[string]bool // statement text -> already planned; see shouldExplain
	plans     []StatementPlan // each distinct statement's plan, in first-seen order
	planErrs  []string        // statements whose plan could not be read; see PlanErrors
	stmts     int64           // atomic: how many statements passed through, so zero findings can be told from zero work
}

// StatementPlan is one intercepted statement and the plan SQLite chose for it.
type StatementPlan struct {
	SQL  string
	Plan Plan
}

// Plans is every distinct statement's plan, in first-seen order.
func (r *Recorder) Plans() []StatementPlan {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]StatementPlan, len(r.plans))
	copy(out, r.plans)
	return out
}

// PlanErrors is every statement the guard tried to plan and could not, with the reason. A
// statement in this list contributes no plan and no finding, so a caller treating an empty
// Findings or Plans as clean must check this is empty first.
func (r *Recorder) PlanErrors() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]string(nil), r.planErrs...)
}

func (r *Recorder) keepPlan(stmt string, rows []PlanRow) {
	p, err := BuildPlan(rows)
	r.mu.Lock()
	defer r.mu.Unlock()
	if err != nil {
		r.planErrs = append(r.planErrs, fmt.Sprintf("%v\n  sql: %s", err, stmt))
		return
	}
	r.plans = append(r.plans, StatementPlan{SQL: stmt, Plan: p})
}

func (r *Recorder) planFailed(stmt string, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.planErrs = append(r.planErrs, fmt.Sprintf("%v\n  sql: %s", err, stmt))
}

// NewRecorder derives the growth set and returns a recorder over it.
func NewRecorder() (*Recorder, error) {
	growth, err := GrowthTables()
	if err != nil {
		return nil, err
	}
	return &Recorder{seen: map[string]bool{}, explained: map[string]bool{}, growth: growth}, nil
}

// Findings is what was seen, in first-seen order.
func (r *Recorder) Findings() []Finding {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]Finding, len(r.findings))
	copy(out, r.findings)
	return out
}

// scanDetail matches EXPLAIN QUERY PLAN's scan lines across the wordings SQLite has used:
// "SCAN TABLE events" (older) and "SCAN events" (3.36+), with or without an alias. Anything
// SEARCHing is using an index and is not this guard's business.
//
// A NO-MATCH HERE IS A SILENT PASS, which is the failure this repo names first, so the pattern
// is deliberately loose (it matches the table name in either wording) and the CALLER is
// responsible for noticing that a run produced no statements at all — see Recorder.Statements.
var scanDetail = regexp.MustCompile(`\bSCAN\s+(?:TABLE\s+)?"?([A-Za-z_][A-Za-z0-9_]*)"?`)

// note records a plan line if it scans a growing table.
func (r *Recorder) note(stmt, detail string) {
	m := scanDetail.FindStringSubmatch(detail)
	if m == nil {
		return
	}
	table := m[1]
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.growth[table] {
		return // a bounded enum table is correctly scanned at any run length
	}
	key := stmt + "\x00" + table
	if r.seen[key] {
		return
	}
	r.seen[key] = true
	r.findings = append(r.findings, Finding{Statement: stmt, Table: table, Detail: detail})
}

// statements counts every statement the guard actually saw, so a caller can tell "nothing
// scanned" from "nothing ran". Those are the same empty findings list and must not read alike.
func (r *Recorder) countStatement() { atomic.AddInt64(&r.stmts, 1) }

// Statements is how many statements passed through the guard. ZERO IS NOT A CLEAN BOARD: it
// means the driver was never used, and every assertion over Findings would pass vacuously.
func (r *Recorder) Statements() int64 { return atomic.LoadInt64(&r.stmts) }

// shouldExplain reports whether this statement still needs planning, and marks it planned.
//
// ONCE PER DISTINCT STATEMENT, NOT ONCE PER EXECUTION. The question this guard asks is about the
// STATEMENT — "can SQLite answer this without walking a growing table" — and the answer does not
// depend on how many times it runs. Asking again on every execution was measured at 2.74x on a
// point lookup (+13.8us) and 1.47x on a full replay (+35us): worst, proportionally, on the
// cheapest statements, because EXPLAIN pays the parse and the planning and skips only the work.
//
// IT IS SAFE HERE FOR A REASON THAT IS NOT GENERALLY TRUE, so it is written down. A plan can
// change under you when the planner's cost estimates change — which for SQLite means ANALYZE has
// populated sqlite_stat1. This schema never runs ANALYZE (checked, nothing in the tree issues
// it), so no stats table exists and plans are chosen from the schema and the indexes alone. They
// are the same on the first row as on the millionth. If ANALYZE is ever introduced, this
// memoisation is the thing that has to go, because a plan measured on an empty fixture would
// stop describing the plan the run gets.
func (r *Recorder) shouldExplain(stmt string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.explained[stmt] {
		return false
	}
	r.explained[stmt] = true
	return true
}

var driverSeq atomic.Int64

// Install registers a driver that runs EXPLAIN QUERY PLAN over every statement before executing
// it, and returns the name to open databases with.
//
// A fresh name per call: database/sql panics on a duplicate registration, and two tests sharing
// one recorder would report each other's statements.
func Install(r *Recorder) (string, error) {
	inner, err := innerDriver()
	if err != nil {
		return "", err
	}
	name := fmt.Sprintf("sqlite-planguard-%d", driverSeq.Add(1))
	sql.Register(name, &guardDriver{inner: inner, rec: r})
	return name, nil
}

// innerDriver is the real sqlite driver, taken from a throwaway handle rather than constructed —
// the driver type is the sqlite package's business and this package does not import it.
func innerDriver() (driver.Driver, error) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return nil, fmt.Errorf("reaching the sqlite driver: %w", err)
	}
	defer db.Close()
	return db.Driver(), nil
}

type guardDriver struct {
	inner driver.Driver
	rec   *Recorder
}

func (d *guardDriver) Open(name string) (driver.Conn, error) {
	c, err := d.inner.Open(name)
	if err != nil {
		return nil, err
	}
	return &guardConn{Conn: c, rec: d.rec}, nil
}

// guardConn embeds the real connection, so everything this file does not override — Begin,
// Prepare, Close — reaches the driver unchanged.
type guardConn struct {
	driver.Conn
	rec *Recorder
}

// QueryContext explains the statement, then runs it.
//
// database/sql prefers this over Prepare when the connection offers it, and the sqlite driver
// does, so every db.Query and db.QueryRow in the record arrives here. A statement issued through
// an explicitly prepared handle would not, which is a real limit and is why the caller checks
// Statements() rather than trusting an empty findings list.
func (c *guardConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	q, ok := c.Conn.(driver.QueryerContext)
	if !ok {
		return nil, driver.ErrSkip
	}
	c.rec.countStatement()
	c.explain(ctx, q, query, args)
	return q.QueryContext(ctx, query, args)
}

func (c *guardConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	e, ok := c.Conn.(driver.ExecerContext)
	if !ok {
		return nil, driver.ErrSkip
	}
	return e.ExecContext(ctx, query, args)
}

func (c *guardConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	b, ok := c.Conn.(driver.ConnBeginTx)
	if !ok {
		return c.Conn.Begin() //nolint:staticcheck // the fallback the interface documents
	}
	return b.BeginTx(ctx, opts)
}

func (c *guardConn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	p, ok := c.Conn.(driver.ConnPrepareContext)
	if !ok {
		return c.Conn.Prepare(query)
	}
	return p.PrepareContext(ctx, query)
}

// explain asks SQLite how it intends to answer, on the SAME connection, so temp tables and
// pragmas in effect are the ones the real statement will meet.
//
// It never fails the query it inspects: a guard that did would turn a diagnostic into an outage.
// A statement it cannot plan contributes no finding and is recorded in PlanErrors instead, and the
// caller is told how many statements ran, so neither reads as a clean plan.
func (c *guardConn) explain(ctx context.Context, q driver.QueryerContext, query string, args []driver.NamedValue) {
	if !isExplainable(query) || !c.rec.shouldExplain(query) {
		return
	}
	rows, err := q.QueryContext(ctx, "EXPLAIN QUERY PLAN "+query, args)
	if err != nil {
		c.rec.planFailed(query, err)
		return
	}
	defer rows.Close()
	cols := rows.Columns()
	at := map[string]int{"id": -1, "parent": -1, "detail": -1}
	for i, name := range cols {
		if _, ok := at[strings.ToLower(name)]; ok {
			at[strings.ToLower(name)] = i
		}
	}
	for name, i := range at {
		if i < 0 {
			c.rec.planFailed(query, fmt.Errorf("EXPLAIN QUERY PLAN returned no %q column (got %v)", name, cols))
			return
		}
	}
	dest := make([]driver.Value, len(cols))
	var plan []PlanRow
	for {
		if err := rows.Next(dest); err != nil {
			if err != io.EOF {
				c.rec.planFailed(query, err)
				return
			}
			break
		}
		row := PlanRow{ID: asInt(dest[at["id"]]), Parent: asInt(dest[at["parent"]]), Detail: asString(dest[at["detail"]])}
		c.rec.note(query, row.Detail)
		plan = append(plan, row)
	}
	c.rec.keepPlan(query, plan)
}

func asInt(v driver.Value) int64 {
	if n, ok := v.(int64); ok {
		return n
	}
	return -1 // not a node id: BuildPlan refuses the parent it cannot find
}

func asString(v driver.Value) string {
	switch s := v.(type) {
	case string:
		return s
	case []byte:
		return string(s)
	}
	return ""
}

// isExplainable keeps the guard away from the statements SQLite will not plan — the transaction
// verbs and PRAGMAs the driver issues around real work.
func isExplainable(query string) bool {
	head := strings.ToUpper(strings.TrimLeft(query, " \t\r\n("))
	for _, verb := range []string{"SELECT", "WITH", "INSERT", "UPDATE", "DELETE", "REPLACE"} {
		if strings.HasPrefix(head, verb) {
			return true
		}
	}
	return false
}
