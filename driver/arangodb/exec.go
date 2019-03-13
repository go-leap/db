package usqldrv_arango

import (
	"context"
	sqldrv "database/sql/driver"
	"encoding/json"
	"io"
	"strconv"

	arango "github.com/arangodb/go-driver"
)

type execResult struct {
	numRows int64
	last    string
}

func (me execResult) RowsAffected() (int64, error) { return me.numRows, nil }
func (me execResult) LastInsertId() (int64, error) { return strconv.ParseInt(me.last, 10, 64) }

type transactCtx struct {
	context.Context
	Options   *arango.TransactionOptions
	OnSuccess func(interface{})
}

// Insert constructs a `query` (to insert `doc` into `coll`) that can be passed to `Conn.ExecContext`.
func Insert(coll string, returnDocKeyOnly bool, returnDocFully bool, doc interface{}) (query string, isTransact bool, err error) {
	var data []byte
	if data, err = json.Marshal(doc); err == nil {
		if query = "INSERT " + string(data) + " IN " + coll; returnDocKeyOnly || returnDocFully {
			if query += " LET d = NEW RETURN"; returnDocKeyOnly {
				query += " { _key: d._key }"
			} else if returnDocFully {
				query += " d"
			}
		}
	}
	return
}

// Transact returns a `context.Context` from `ctx` that can be passed to
// `Conn.ExecContext` to signify that the query is an ArangoDB "transaction"
// JavaScript function to be run on the server using `transactionOptions`. If
// `onSucceeded` is given, it is called with either `nil` or the JS func's result.
func Transact(ctx context.Context, transactionOptions *arango.TransactionOptions, onSucceeded func(result interface{})) context.Context {
	return &transactCtx{Context: ctx, Options: transactionOptions, OnSuccess: onSucceeded}
}

// ExecContext implements `sqldrv.ExecerContext`
func (me *arangoConn) ExecContext(ctx context.Context, query string, args []sqldrv.NamedValue) (r sqldrv.Result, err error) {
	var wastransact bool
	if wastransact, err = me.transactMaybe(ctx, query); err == nil {
		res := execResult{numRows: -1}
		if !wastransact {
			var rows sqldrv.Rows
			var nope none
			if rows, err = me.QueryContext(ctx, query, args); err == nil {
				rowcur := rows.(*arangoRowsCursor)
				res.numRows, rowcur.onReadDocIntoNewPtr = 0, func(RowsCursor) interface{} { return &nope }
				cells := make([]sqldrv.Value, len(rowcur.Columns()))
				for err == nil && !rowcur.eof {
					if err = rowcur.Next(cells); err == nil {
						res.numRows, res.last = res.numRows+1, cells[1].(arango.DocumentMeta).Key
					} else if err == io.EOF {
						err = nil
						break
					}
				}
				if res.numRows == 0 {
					if querystats := rowcur.Cursor.Statistics(); querystats != nil {
						res.numRows = querystats.WritesExecuted()
					}
				}
			}
		}
		r = res
	}
	return
}

func (me *arangoConn) transactMaybe(ctx context.Context, query string) (didAttempt bool, err error) {
	tctx, _ := ctx.(*transactCtx)
	if didAttempt = tctx != nil; didAttempt {
		var result interface{}
		if result, err = me.Transaction(tctx.Context, query, tctx.Options); err == nil && tctx.OnSuccess != nil {
			tctx.OnSuccess(result)
		}
	}
	return
}
