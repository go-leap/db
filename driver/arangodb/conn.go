package usqldrv_arango

import (
	sqldrv "database/sql/driver"
	"errors"

	arango "github.com/arangodb/go-driver"
)

// Conn is returned by `Driver.OpenConnector(name).Connect(ctx)`.
//
// Absent `error`s, its `sqldrv.QueryerContext` implementation returns
// `RowsCursor`s that implement both `sqldrv.Rows` and `arango.Cursor`.
type Conn interface {
	sqldrv.Conn
	sqldrv.ExecerContext
	sqldrv.QueryerContext
	arango.Database
}

type arangoConn struct {
	drv *Driver
	arango.Database
}

func (me *arangoConn) Client() arango.Client {
	return me.drv.shared.Client
}

// Begin implements `sqldrv.Conn`, but always fails due to lack of support
// by ArangoDB for the proper rollback/commit transactions paradigm.
//
// The "transactions" that ArangoDB does support (that is, the server-side
// execution of a client-provided JavaScript `function`), can be accessed
// via this `Conn`'s `sqldrv.ExecerContext` implementation.
func (*arangoConn) Begin() (sqldrv.Tx, error) {
	return nil, errors.New("bug: Begin should never be called since ArangoDB (and/or its currently-used underlying Go driver) does not support the proper rollback/commit transactions paradigm (use this `Conn`'s `ExecerContext` implementation for ArangoDB's own (so-called) 'transactions')")
}

// Close implements `sqldrv.Conn` and is a no-op, given that ArangoDB talks via HTTP-APIs (no comment =)
func (*arangoConn) Close() (err error) { return }

// Prepare implements `sqldrv.Conn`, but always fails due to being deprecated in favour of `ExecerContext` / `QueryerContext`.
func (*arangoConn) Prepare(query string) (sqldrv.Stmt, error) {
	return nil, errors.New("bug: `Prepare` should never be called since this `database/sql/driver` also implements `ExecerContext` and `QueryerContext`")
}

func (*arangoConn) bindVarsFrom(args []sqldrv.NamedValue) (bindVars map[string]interface{}) {
	if len(args) > 0 {
		bindVars = make(map[string]interface{}, len(args))
		for i := range args {
			bindVars[args[i].Name] = args[i].Value
		}
	}
	return
}
