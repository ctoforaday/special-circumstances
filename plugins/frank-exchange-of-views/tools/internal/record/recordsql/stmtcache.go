package recordsql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"reflect"
	"strings"
	"sync"
)

// THE RECORD HANDLE KEEPS ITS READS PREPARED. Preparing a statement is most of what a point read
// on a view costs: the `gap` view's plan is built from a dozen CTEs, and its prepare measured
// 0.99 ms against 0.03 ms to execute one that returns nothing (#1700). A process that asks the same
// read twice — a seat's write validates against the gap view, then its standing reads the same
// view; the dashboard re-reads every fifteen seconds; a migration replays a run's acts in one
// process — prepared it each time.
//
// THE CACHE IS ON THE CONNECTION, NOT ON *sql.DB, because the pool holds one connection
// (SetMaxOpenConns(1) in openUncached) and the record's reads arrive on it two ways: through the
// handle, and through a transaction that holds the same connection (readSnapshot, the write path's
// writeView, Replay.Querier). A *sql.Stmt prepared on the handle cannot be asked through a *sql.Tx
// without tx.Stmt, and a Querier does not say which handle a transaction came from; a statement
// kept by the connection serves both, because both reach the one connection.
//
// KEYED BY THE SQL TEXT, AND BOUNDED BY THE CODE. Every read the record issues is a constant, or a
// constant with a table or column name the code chose spliced in (listValuesByEvent, eventsWhere's
// type list), so the distinct texts a process can issue are fixed by the binary — about a hundred.
// A text carrying a value would make the set grow with the data, so the cache also stops taking
// statements at maxCachedStatements and runs any further text as an uncached query; a cache that
// fills reads as slower, never as wrong.
//
// ONLY READS ARE KEPT: a statement beginning SELECT or WITH. Writes and PRAGMAs go to the driver
// as they always have.
//
// A STATEMENT IN USE IS NOT REUSED. The driver's statement is one cursor: asking it again while
// the rows of its last answer are open would rewind those rows under their reader. A text whose
// statement is in use runs uncached, as it would without this cache.
//
// CLOSING THE CONNECTION FINALIZES EVERY KEPT STATEMENT FIRST. SQLite's close_v2 leaves a
// connection with live statements open as a zombie until the last is finalized, holding the file;
// on Windows a held file cannot be removed (CloseUnder's reason for existing). Close, CloseUnder
// and CloseAll close the handle, the handle closes its connection, and the connection finalizes
// its statements before it closes.
const maxCachedStatements = 512

// stmtCacheConnector opens the record's connections through the driver driverName names, each
// wrapped in a statement cache.
type stmtCacheConnector struct {
	drv driver.Driver
	dsn string
}

func (c stmtCacheConnector) Connect(context.Context) (driver.Conn, error) {
	conn, err := c.drv.Open(c.dsn)
	if err != nil {
		return nil, err
	}
	return &stmtCacheConn{Conn: conn, stmts: map[string]*cachedStmt{}}, nil
}

func (c stmtCacheConnector) Driver() driver.Driver { return c.drv }

// openDB opens a handle on dsn through the driver driverName names, with the statement cache.
// sql.Open would reach the same driver with the same dsn (the driver offers no connector of its
// own); OpenDB is what lets the connection it returns be wrapped.
func openDB(dsn string) (*sql.DB, error) {
	probe, err := sql.Open(driverName(), "")
	if err != nil {
		return nil, err
	}
	drv := probe.Driver()
	if err := probe.Close(); err != nil {
		return nil, err
	}
	return sql.OpenDB(stmtCacheConnector{drv: drv, dsn: dsn}), nil
}

type cachedStmt struct {
	stmt  driver.Stmt
	inUse bool
}

// stmtCacheConn is the driver's connection with its reads kept prepared. Everything it does not
// override reaches the driver's connection through the methods below or the embedded Conn.
type stmtCacheConn struct {
	driver.Conn
	mu    sync.Mutex
	stmts map[string]*cachedStmt
}

// isCachedRead is whether query is a read the cache keeps.
func isCachedRead(query string) bool {
	head := strings.ToUpper(strings.TrimLeft(query, " \t\r\n("))
	return strings.HasPrefix(head, "SELECT") || strings.HasPrefix(head, "WITH")
}

// take is the kept statement for query, prepared on first use and marked in use; nil where the
// text is not kept or its statement is in use.
func (c *stmtCacheConn) take(ctx context.Context, query string) (*cachedStmt, error) {
	if !isCachedRead(query) {
		return nil, nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if s, ok := c.stmts[query]; ok {
		if s.inUse {
			return nil, nil
		}
		s.inUse = true
		return s, nil
	}
	if len(c.stmts) >= maxCachedStatements {
		return nil, nil
	}
	p, ok := c.Conn.(driver.ConnPrepareContext)
	if !ok {
		return nil, nil
	}
	st, err := p.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	if _, ok := st.(driver.StmtQueryContext); !ok {
		st.Close()
		return nil, nil
	}
	s := &cachedStmt{stmt: st, inUse: true}
	c.stmts[query] = s
	return s, nil
}

func (c *stmtCacheConn) release(s *cachedStmt) {
	c.mu.Lock()
	s.inUse = false
	c.mu.Unlock()
}

// QueryContext answers a kept read from its prepared statement, and any other query as the driver
// does.
func (c *stmtCacheConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	s, err := c.take(ctx, query)
	if err != nil {
		return nil, err
	}
	if s == nil {
		q, ok := c.Conn.(driver.QueryerContext)
		if !ok {
			return nil, driver.ErrSkip
		}
		return q.QueryContext(ctx, query, args)
	}
	rows, err := s.stmt.(driver.StmtQueryContext).QueryContext(ctx, args)
	if err != nil {
		c.release(s)
		return nil, err
	}
	return &cachedRows{Rows: rows, conn: c, stmt: s}, nil
}

func (c *stmtCacheConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	e, ok := c.Conn.(driver.ExecerContext)
	if !ok {
		return nil, driver.ErrSkip
	}
	return e.ExecContext(ctx, query, args)
}

func (c *stmtCacheConn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	if p, ok := c.Conn.(driver.ConnPrepareContext); ok {
		return p.PrepareContext(ctx, query)
	}
	return c.Conn.Prepare(query)
}

func (c *stmtCacheConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	if b, ok := c.Conn.(driver.ConnBeginTx); ok {
		return b.BeginTx(ctx, opts)
	}
	return c.Conn.Begin() //nolint:staticcheck // the fallback the interface documents
}

func (c *stmtCacheConn) Ping(ctx context.Context) error {
	if p, ok := c.Conn.(driver.Pinger); ok {
		return p.Ping(ctx)
	}
	return nil
}

func (c *stmtCacheConn) ResetSession(ctx context.Context) error {
	if r, ok := c.Conn.(driver.SessionResetter); ok {
		return r.ResetSession(ctx)
	}
	return nil
}

func (c *stmtCacheConn) IsValid() bool {
	if v, ok := c.Conn.(driver.Validator); ok {
		return v.IsValid()
	}
	return true
}

// Close finalizes every kept statement, then closes the connection.
func (c *stmtCacheConn) Close() error {
	c.mu.Lock()
	var errs []error
	for q, s := range c.stmts {
		errs = append(errs, s.stmt.Close())
		delete(c.stmts, q)
	}
	c.mu.Unlock()
	errs = append(errs, c.Conn.Close())
	return errors.Join(errs...)
}

// cachedRows is a kept statement's answer: closing it frees the statement for the next ask.
type cachedRows struct {
	driver.Rows
	conn     *stmtCacheConn
	stmt     *cachedStmt
	released bool
}

func (r *cachedRows) Close() error {
	err := r.Rows.Close()
	if !r.released {
		r.released = true
		r.conn.release(r.stmt)
	}
	return err
}

func (r *cachedRows) ColumnTypeDatabaseTypeName(i int) string {
	if t, ok := r.Rows.(driver.RowsColumnTypeDatabaseTypeName); ok {
		return t.ColumnTypeDatabaseTypeName(i)
	}
	return ""
}

func (r *cachedRows) ColumnTypeScanType(i int) reflect.Type {
	if t, ok := r.Rows.(driver.RowsColumnTypeScanType); ok {
		return t.ColumnTypeScanType(i)
	}
	return reflect.TypeFor[any]()
}

func (r *cachedRows) ColumnTypeNullable(i int) (nullable, ok bool) {
	if t, ok := r.Rows.(driver.RowsColumnTypeNullable); ok {
		return t.ColumnTypeNullable(i)
	}
	return false, false
}

func (r *cachedRows) ColumnTypeLength(i int) (int64, bool) {
	if t, ok := r.Rows.(driver.RowsColumnTypeLength); ok {
		return t.ColumnTypeLength(i)
	}
	return 0, false
}

func (r *cachedRows) ColumnTypePrecisionScale(i int) (precision, scale int64, ok bool) {
	if t, ok := r.Rows.(driver.RowsColumnTypePrecisionScale); ok {
		return t.ColumnTypePrecisionScale(i)
	}
	return 0, 0, false
}
