package usqldrv_proxy

import (
	"context"
	sqldrv "database/sql/driver"
)

type proxyConnBeginTx struct {
	inner sqldrv.ConnBeginTx
}

func (me proxyConnBeginTx) BeginTx(ctx context.Context, opts sqldrv.TxOptions) (tx sqldrv.Tx, err error) {
	on := On.Conn.BeginTx.begin(me, &err, ctx, opts)
	if tx, err = me.inner.BeginTx(ctx, opts); err != nil {
		tx = nil
	} else if tx != nil {
		tx = proxyTx{inner: tx}
	}
	on.done(tx)
	return
}

type proxyConnPrepareContext struct {
	inner sqldrv.ConnPrepareContext
}

func (me proxyConnPrepareContext) PrepareContext(ctx context.Context, query string) (stmt sqldrv.Stmt, err error) {
	on := On.Conn.PrepareContext.begin(me, &err, ctx, query)
	stmt, err = newStmt(me.inner.PrepareContext(ctx, query))
	on.done(stmt)
	return
}

type proxyRowsNextResultSet struct {
	sqldrv.Rows
	inner sqldrv.RowsNextResultSet
}

func (me proxyRowsNextResultSet) HasNextResultSet() (hasNextResultSet bool) {
	hasNextResultSet = me.inner.HasNextResultSet()
	return
}

func (me proxyRowsNextResultSet) NextResultSet() (err error) {
	on := On.Rows.NextResultSet.begin(me, &err)
	err = me.inner.NextResultSet()
	on.done()
	return
}

type proxyStmtExecContext struct {
	sqldrv.StmtExecContext
}

func (me proxyStmtExecContext) ExecContext(ctx context.Context, args []sqldrv.NamedValue) (result sqldrv.Result, err error) {
	on := On.Stmt.ExecContext.begin(me, &err, ctx, args)
	result, err = me.StmtExecContext.ExecContext(ctx, args)
	on.done(result)
	return
}

type proxyStmtQueryContext struct {
	sqldrv.StmtQueryContext
}

func (me proxyStmtQueryContext) QueryContext(ctx context.Context, args []sqldrv.NamedValue) (rows sqldrv.Rows, err error) {
	on := On.Stmt.QueryContext.begin(me, &err, ctx, args)
	rows, err = newRows(me.StmtQueryContext.QueryContext(ctx, args))
	on.done(rows)
	return
}

func newComboStmt(stmt sqldrv.Stmt, stmtExec sqldrv.StmtExecContext, stmtQuery sqldrv.StmtQueryContext) sqldrv.Stmt {
	stmt = proxyStmt{inner: stmt}
	if stmtExec != nil {
		stmtExec = proxyStmtExecContext{stmtExec}
	}
	if stmtQuery != nil {
		stmtQuery = proxyStmtQueryContext{stmtQuery}
	}
	if ex, qu := stmtExec != nil, stmtQuery != nil; ex || qu {
		switch {
		case ex && qu:
			return &struct {
				sqldrv.Stmt
				sqldrv.StmtExecContext
				sqldrv.StmtQueryContext
			}{stmt, stmtExec, stmtQuery}
		case ex:
			return &struct {
				sqldrv.Stmt
				sqldrv.StmtExecContext
			}{stmt, stmtExec}
		case qu:
			return &struct {
				sqldrv.Stmt
				sqldrv.StmtQueryContext
			}{stmt, stmtQuery}
		}
	}
	return stmt
}

func newComboConn(driver sqldrv.Driver, conn sqldrv.Conn, cbt sqldrv.ConnBeginTx, cpc sqldrv.ConnPrepareContext, cp sqldrv.Pinger, csr sqldrv.SessionResetter) sqldrv.Conn {
	conn = proxyConn{inner: conn}
	if cbt != nil {
		cbt = proxyConnBeginTx{cbt}
	}
	if cpc != nil {
		cpc = proxyConnPrepareContext{cpc}
	}

	if bt, pc, p, sr := cbt != nil, cpc != nil, cp != nil, csr != nil; bt || pc || p || sr {
		switch {
		case bt && pc && p && sr:
			return &struct {
				sqldrv.Conn
				sqldrv.ConnBeginTx
				sqldrv.ConnPrepareContext
				sqldrv.Pinger
				sqldrv.SessionResetter
			}{conn, cbt, cpc, cp, csr}
		case pc && p && sr:
			return &struct {
				sqldrv.Conn
				sqldrv.ConnPrepareContext
				sqldrv.Pinger
				sqldrv.SessionResetter
			}{conn, cpc, cp, csr}
		case bt && p && sr:
			return &struct {
				sqldrv.Conn
				sqldrv.ConnBeginTx
				sqldrv.Pinger
				sqldrv.SessionResetter
			}{conn, cbt, cp, csr}
		case bt && pc && sr:
			return &struct {
				sqldrv.Conn
				sqldrv.ConnBeginTx
				sqldrv.ConnPrepareContext
				sqldrv.SessionResetter
			}{conn, cbt, cpc, csr}
		case bt && pc && p:
			return &struct {
				sqldrv.Conn
				sqldrv.ConnBeginTx
				sqldrv.ConnPrepareContext
				sqldrv.Pinger
			}{conn, cbt, cpc, cp}
		case pc && p:
			return &struct {
				sqldrv.Conn
				sqldrv.ConnPrepareContext
				sqldrv.Pinger
			}{conn, cpc, cp}
		case bt && sr:
			return &struct {
				sqldrv.Conn
				sqldrv.ConnBeginTx
				sqldrv.SessionResetter
			}{conn, cbt, csr}
		case bt && p:
			return &struct {
				sqldrv.Conn
				sqldrv.ConnBeginTx
				sqldrv.Pinger
			}{conn, cbt, cp}
		case pc && sr:
			return &struct {
				sqldrv.Conn
				sqldrv.ConnPrepareContext
				sqldrv.SessionResetter
			}{conn, cpc, csr}
		case p && sr:
			return &struct {
				sqldrv.Conn
				sqldrv.Pinger
				sqldrv.SessionResetter
			}{conn, cp, csr}
		case bt && pc:
			return &struct {
				sqldrv.Conn
				sqldrv.ConnBeginTx
				sqldrv.ConnPrepareContext
			}{conn, cbt, cpc}
		case sr:
			return &struct {
				sqldrv.Conn
				sqldrv.SessionResetter
			}{conn, csr}
		case p:
			return &struct {
				sqldrv.Conn
				sqldrv.Pinger
			}{conn, cp}
		case pc:
			return &struct {
				sqldrv.Conn
				sqldrv.ConnPrepareContext
			}{conn, cpc}
		case bt:
			return &struct {
				sqldrv.Conn
				sqldrv.ConnBeginTx
			}{conn, cbt}
		}
	}
	return conn
}
