// Package `db/driver/proxy` provides event-handling hooks (for logging, metrics,
// tracing, diagnostics or whatever purposes) into any (wrapped via `WrapDriver` or
// `WrapConnector`) `database/sql/driver`'s noteworthy operations / situations like so:
//
//  // import proxy "github.com/go-leap/db/driver/proxy"
//  // // import sqldrv "database/sql/driver"
//
//  proxy.On.Connector.Connect.Before =
//
//    func(this proxy.This, args ...interface{}) {
//        connectionString := args[0].(string)
//        // connector := this.(sqldrv.Connector)
//        log.Print("DB connection attempt to: " + connectionString)
//    }
//
package usqldrv_proxy

import (
	"context"
	sqldrv "database/sql/driver"
)

// WrapConnector returns a new proxy `database/sql/driver.Connector` wrapping
// `connector` and relaying subscribed events to your `Hook` handlers set in
// `On` (except those set in `On.Driver`, which only trigger for `WrapDriver`s).
func WrapConnector(connector sqldrv.Connector) sqldrv.Connector {
	if connector != nil {
		connector = proxyConnector{driver: connector.Driver(), inner: connector}
	}
	return connector
}

// WrapDriver returns a new proxy `database/sql/driver.Driver` wrapping
// `drv` and relaying subscribed events to your `Hook` handlers set in `On`.
func WrapDriver(drv sqldrv.Driver) sqldrv.Driver {
	if orig := drv; orig != nil {
		drv = proxyDriver{inner: orig}
		if drvctx, _ := orig.(sqldrv.DriverContext); drvctx != nil {
			drv = proxyDriverContext{inner: drvctx, Driver: drv}
		}
	}
	return drv
}

type proxyDriverContext struct {
	inner sqldrv.DriverContext
	sqldrv.Driver
}

func (me proxyDriverContext) OpenConnector(name string) (connector sqldrv.Connector, err error) {
	on := On.Driver.OpenConnector.begin(me, &err, name)
	if connector, err = me.inner.OpenConnector(name); err != nil {
		connector = nil
	} else if connector != nil {
		connector = proxyConnector{inner: connector, driver: me}
	}
	on.done(connector)
	return
}

type proxyDriver struct {
	inner sqldrv.Driver
}

func (me proxyDriver) Open(name string) (conn sqldrv.Conn, err error) {
	on := On.Driver.Open.begin(me, &err, name)
	conn, err = me.inner.Open(name)
	conn, err = newConn(me.inner, conn, err)
	on.done(conn)
	return
}

type proxyConnector struct {
	inner  sqldrv.Connector
	driver sqldrv.Driver
}

func (me proxyConnector) Connect(ctx context.Context) (conn sqldrv.Conn, err error) {
	on := On.Connector.Connect.begin(me, &err, ctx)
	conn, err = me.inner.Connect(ctx)
	conn, err = newConn(me.driver, conn, err)
	on.done(conn)
	return
}

func (me proxyConnector) Driver() sqldrv.Driver {
	return me.driver
}

func newConn(driver sqldrv.Driver, conn sqldrv.Conn, err error) (sqldrv.Conn, error) {
	if err != nil {
		conn = nil
	} else if conn != nil {
		connbegintx, _ := conn.(sqldrv.ConnBeginTx)
		connprepctx, _ := conn.(sqldrv.ConnPrepareContext)
		connping, _ := conn.(sqldrv.Pinger)
		connresess, _ := conn.(sqldrv.SessionResetter)
		conn = newComboConn(driver, conn, connbegintx, connprepctx, connping, connresess)
	}
	return conn, err
}

type proxyConn struct {
	inner sqldrv.Conn
}

func (me proxyConn) Begin() (tx sqldrv.Tx, err error) {
	on := On.Conn.Begin.begin(me, &err)
	if tx, err = me.inner.Begin(); err != nil {
		tx = nil
	} else if tx != nil {
		tx = proxyTx{inner: tx}
	}
	on.done(tx)
	return
}

func (me proxyConn) Close() (err error) {
	on := On.Conn.Close.begin(me, &err)
	err = me.inner.Close()
	on.done()
	return
}

func (me proxyConn) Prepare(query string) (stmt sqldrv.Stmt, err error) {
	on := On.Conn.Prepare.begin(me, &err, query)
	stmt, err = newStmt(me.inner.Prepare(query))
	on.done(stmt)
	return
}

func (me proxyConn) CheckNamedValue(nv *sqldrv.NamedValue) (err error) {
	if nvcheck, _ := me.inner.(sqldrv.NamedValueChecker); nvcheck == nil {
		err = sqldrv.ErrSkip
	} else {
		on := On.Conn.CheckNamedValue.begin(me, &err, nv)
		err = nvcheck.CheckNamedValue(nv)
		on.done()
	}
	return
}

func (me proxyConn) Exec(query string, args []sqldrv.Value) (result sqldrv.Result, err error) {
	if execer, _ := me.inner.(sqldrv.Execer); execer == nil {
		err = sqldrv.ErrSkip
	} else {
		on := On.Conn.Exec.begin(me, &err, query, args)
		result, err = execer.Exec(query, args)
		on.done(result)
	}
	return
}

func (me proxyConn) ExecContext(ctx context.Context, query string, args []sqldrv.NamedValue) (result sqldrv.Result, err error) {
	if exctx, _ := me.inner.(sqldrv.ExecerContext); exctx == nil {
		err = sqldrv.ErrSkip
	} else {
		on := On.Conn.ExecContext.begin(me, &err, ctx, query, args)
		result, err = exctx.ExecContext(ctx, query, args)
		on.done(result)
	}
	return
}

func (me proxyConn) QueryContext(ctx context.Context, query string, args []sqldrv.NamedValue) (rows sqldrv.Rows, err error) {
	if quctx, _ := me.inner.(sqldrv.QueryerContext); quctx == nil {
		err = sqldrv.ErrSkip
	} else {
		on := On.Conn.QueryContext.begin(me, &err, ctx, query, args)
		if rows, err = quctx.QueryContext(ctx, query, args); err != nil {
			rows = nil
		} else if rows != nil {
			rows = proxyRows{inner: rows}
		}
		on.done(rows)
	}
	return
}

func (me proxyConn) Query(query string, args []sqldrv.Value) (rows sqldrv.Rows, err error) {
	if querier, _ := me.inner.(sqldrv.Queryer); querier == nil {
		err = sqldrv.ErrSkip
	} else {
		on := On.Conn.Query.begin(me, &err, query, args)
		if rows, err = querier.Query(query, args); err != nil {
			rows = nil
		} else if rows != nil {
			rows = proxyRows{inner: rows}
		}
		on.done(rows)
	}
	return
}

func newStmt(stmt sqldrv.Stmt, err error) (sqldrv.Stmt, error) {
	if err != nil {
		stmt = nil
	} else if stmt != nil {
		stmtexec, _ := stmt.(sqldrv.StmtExecContext)
		stmtquery, _ := stmt.(sqldrv.StmtQueryContext)
		stmt = newComboStmt(stmt, stmtexec, stmtquery)
	}
	return stmt, err
}

type proxyStmt struct {
	inner sqldrv.Stmt
}

func (me proxyStmt) Close() (err error) {
	on := On.Stmt.Close.begin(me, &err)
	err = me.inner.Close()
	on.done()
	return
}

func (me proxyStmt) Exec(args []sqldrv.Value) (result sqldrv.Result, err error) {
	on := On.Stmt.Exec.begin(me, &err, args)
	result, err = me.inner.Exec(args)
	on.done(result)
	return
}

func (me proxyStmt) NumInput() (numInput int) {
	numInput = me.inner.NumInput()
	return
}

func (me proxyStmt) Query(args []sqldrv.Value) (rows sqldrv.Rows, err error) {
	on := On.Stmt.Query.begin(me, &err, args)
	rows, err = newRows(me.inner.Query(args))
	on.done(rows)
	return
}

func (me proxyStmt) CheckNamedValue(nv *sqldrv.NamedValue) (err error) {
	if nvcheck, _ := me.inner.(sqldrv.NamedValueChecker); nvcheck == nil {
		err = sqldrv.ErrSkip
	} else {
		on := On.Stmt.CheckNamedValue.begin(me, &err, nv)
		err = nvcheck.CheckNamedValue(nv)
		on.done()
	}
	return
}

func newRows(rows sqldrv.Rows, err error) (sqldrv.Rows, error) {
	if err != nil {
		rows = nil
	} else if orig := rows; orig != nil {
		rows = proxyRows{inner: orig}
		rowsx, _ := orig.(sqldrv.RowsNextResultSet)
		if rowsx != nil {
			rows = proxyRowsNextResultSet{inner: rowsx, Rows: rows}
		}
	}
	return rows, err
}

type proxyRows struct {
	inner sqldrv.Rows
}

func (me proxyRows) Columns() (cols []string) {
	cols = me.inner.Columns()
	return
}

func (me proxyRows) Close() (err error) {
	on := On.Rows.Close.begin(me, &err)
	err = me.inner.Close()
	on.done()
	return
}

func (me proxyRows) Next(dest []sqldrv.Value) (err error) {
	on := On.Rows.Next.begin(me, &err, dest)
	err = me.inner.Next(dest)
	on.done()
	return
}

type proxyTx struct {
	inner sqldrv.Tx
}

func (me proxyTx) Commit() (err error) {
	on := On.Tx.Commit.begin(me, &err)
	err = me.inner.Commit()
	on.done()
	return
}

func (me proxyTx) Rollback() (err error) {
	on := On.Tx.Rollback.begin(me, &err)
	err = me.inner.Rollback()
	on.done()
	return
}
