package planguard

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
)

// Recorder collects plans across every connection a guarded database opens.
//
// It records rather than judges. What counts as a defect is a judgement about the statement — a
// full replay is supposed to scan `events` — and that judgement belongs to Analyzer.Shapes with
// the statement in front of it, not to the interceptor. Mixing the two would bury a policy
// decision inside a driver.
type Recorder struct {
	mu        sync.Mutex
	explained map[string]bool // statement text -> already planned; see shouldExplain
	plans     []StatementPlan // each distinct statement's plan, in first-seen order
	planErrs  []string        // statements whose plan could not be read; see PlanErrors
	stmts     int64           // atomic: how many statements passed through, so zero findings can be told from zero work

	// issuedBy, when set, decides whether a source file is one whose statements this recorder
	// answers for; issued holds each statement text some such file has issued. See noteIssuer.
	issuedBy func(file string) bool
	issued   map[string]bool
}

// StatementPlan is one intercepted statement and the plan SQLite chose for it.
type StatementPlan struct {
	SQL  string
	Plan Plan
	// Issued is whether a file the recorder's issuedBy accepts issued this text (always false
	// on a recorder without one).
	Issued bool
}

// Plans is every distinct statement's plan, in first-seen order.
func (r *Recorder) Plans() []StatementPlan {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]StatementPlan, len(r.plans))
	copy(out, r.plans)
	for i := range out {
		out[i].Issued = r.issued[out[i].SQL]
	}
	return out
}

// AttributeTo makes the recorder note, for each statement text, whether a source file issuedBy
// accepts has issued it. A suite's own assertions issue statements too; this is what tells the
// record's statements from the test's.
func (r *Recorder) AttributeTo(issuedBy func(file string) bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.issuedBy, r.issued = issuedBy, map[string]bool{}
}

// noteIssuer records whether the code that issued stmt is a file the recorder answers for.
//
// THE ISSUER IS THE FRAME THAT CALLED database/sql. Below it are the drivers — this guard and
// the record's statement cache — and above it whatever asked; the first frame out of
// database/sql is the code that wrote the statement or was handed it. It is asked on every
// execution until one issuer qualifies, because a test may issue a text before the record does.
func (r *Recorder) noteIssuer(stmt string) {
	r.mu.Lock()
	ask := r.issuedBy
	done := ask == nil || r.issued[stmt]
	r.mu.Unlock()
	if done {
		return
	}
	var pcs [64]uintptr
	frames := runtime.CallersFrames(pcs[:runtime.Callers(2, pcs[:])])
	inSQL := false
	for {
		f, more := frames.Next()
		if strings.HasPrefix(f.Function, "database/sql.") {
			inSQL = true
		} else if inSQL {
			if ask(f.File) {
				r.mu.Lock()
				r.issued[stmt] = true
				r.mu.Unlock()
			}
			return
		}
		if !more {
			return
		}
	}
}

// PlanErrors is every statement the guard tried to plan and could not, with the reason. A
// statement in this list contributes no plan and no finding, so a caller treating an empty
// Plans as clean must check this is empty first.
func (r *Recorder) PlanErrors() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]string(nil), r.planErrs...)
}

func (r *Recorder) keepPlan(stmt string, rows []PlanRow) {
	var p Plan
	var err error
	if len(rows) > 0 || !readsNothing(stmt) {
		p, err = BuildPlan(rows)
	}
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

// NewRecorder returns an empty recorder.
func NewRecorder() *Recorder {
	return &Recorder{explained: map[string]bool{}}
}

// statements counts every statement the guard actually saw, so a caller can tell "nothing
// scanned" from "nothing ran". Those are the same empty findings list and must not read alike.
func (r *Recorder) countStatement() { atomic.AddInt64(&r.stmts, 1) }

// Statements is how many statements passed through the guard. ZERO IS NOT A CLEAN BOARD: it
// means the driver was never used, and every assertion over Plans would pass vacuously.
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
// does, so a db.Query or db.QueryRow the record's statement cache does not keep arrives here; one
// it keeps arrives at PrepareContext and is watched there (guardStmt). The caller still checks
// Statements() rather than trusting an empty findings list.
func (c *guardConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	q, ok := c.Conn.(driver.QueryerContext)
	if !ok {
		return nil, driver.ErrSkip
	}
	c.rec.countStatement()
	c.rec.noteIssuer(query)
	c.explain(ctx, q, query, args)
	return q.QueryContext(ctx, query, args)
}

// ExecContext explains the statement, then runs it: the record's WRITES are planned as its reads
// are. A write carries its own reads — an INSERT … SELECT, an UPDATE's WHERE — and the
// transaction that holds a write lock is where a slow plan costs every other writer.
func (c *guardConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	e, ok := c.Conn.(driver.ExecerContext)
	if !ok {
		return nil, driver.ErrSkip
	}
	c.rec.countStatement()
	c.rec.noteIssuer(query)
	if q, ok := c.Conn.(driver.QueryerContext); ok {
		c.explain(ctx, q, query, args)
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

// PrepareContext prepares the statement, and hands back one that explains and counts each time
// it is queried or executed.
//
// THE RECORD'S READS ARRIVE HERE AS WELL AS AT QueryContext: recordsql keeps its reads prepared on
// the connection (recordsql/stmtcache.go), so a read is prepared once and then queried through
// the statement, never through the connection. A statement that ran unseen would be the hole
// Statements() exists to expose, so the guard watches the statement as it watches the connection.
func (c *guardConn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	p, ok := c.Conn.(driver.ConnPrepareContext)
	if !ok {
		return c.Conn.Prepare(query)
	}
	st, err := p.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	return &guardStmt{Stmt: st, conn: c, query: query}, nil
}

// guardStmt is a prepared statement the guard explains and counts on each query and each exec, as
// it does a statement on the connection.
type guardStmt struct {
	driver.Stmt
	conn  *guardConn
	query string
}

func (s *guardStmt) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	q, ok := s.Stmt.(driver.StmtQueryContext)
	if !ok {
		return nil, driver.ErrSkip
	}
	s.conn.rec.countStatement()
	s.conn.rec.noteIssuer(s.query)
	if cq, ok := s.conn.Conn.(driver.QueryerContext); ok {
		s.conn.explain(ctx, cq, s.query, args)
	}
	return q.QueryContext(ctx, args)
}

func (s *guardStmt) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	e, ok := s.Stmt.(driver.StmtExecContext)
	if !ok {
		return nil, driver.ErrSkip
	}
	s.conn.rec.countStatement()
	s.conn.rec.noteIssuer(s.query)
	if cq, ok := s.conn.Conn.(driver.QueryerContext); ok {
		s.conn.explain(ctx, cq, s.query, args)
	}
	return e.ExecContext(ctx, args)
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

// readsNothing is whether stmt is an insert of literal rows, whose plan SQLite reports as no rows
// at all: there is no table to search, so an empty plan is the whole answer and not a failed read.
// An INSERT … SELECT reads, and is planned like any other read.
func readsNothing(stmt string) bool {
	head := strings.ToUpper(strings.TrimLeft(stmt, " \t\r\n"))
	if !strings.HasPrefix(head, "INSERT") && !strings.HasPrefix(head, "REPLACE") {
		return false
	}
	toks := strings.Fields(strings.NewReplacer("(", " ( ", ")", " ) ", ",", " , ").Replace(head))
	for _, t := range toks {
		if t == "SELECT" {
			return false
		}
	}
	return true
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
