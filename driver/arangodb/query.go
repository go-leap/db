package usqldrv_arango

import (
	"context"
	sqldrv "database/sql/driver"
	"io"

	arango "github.com/arangodb/go-driver"
)

const (
	// both returned by our `sqldrv.Rows.Columns()` implementation:
	ColNameDoc  = "_doc"  // map[string]interface{} or other (see `RowsCursor`)
	ColNameMeta = "_meta" // always an arango.DocumentMeta
)

type queryCtx struct {
	context.Context
	OnReadDocDecodeIntoNewPtr func(RowsCursor) interface{}
}

// Query returns a `context.Context` that can be passed to `Conn.QueryContext`.
func Query(ctx context.Context, wantCountInRowsCursor bool, onReadDocDecodeIntoNewPtr func(RowsCursor) interface{}) context.Context {
	if wantCountInRowsCursor {
		ctx = arango.WithQueryCount(ctx)
	}
	return &queryCtx{Context: ctx, OnReadDocDecodeIntoNewPtr: onReadDocDecodeIntoNewPtr}
}

// QueryContext implements `sqldrv.QueryerContext` and if no `err`, `rows` will always implement this package's `RowsCursor` interface.
func (me *arangoConn) QueryContext(ctx context.Context, query string, args []sqldrv.NamedValue) (rows sqldrv.Rows, err error) {
	var rowcur arangoRowsCursor
	if qctx, _ := ctx.(*queryCtx); qctx != nil {
		rowcur.onReadDocIntoNewPtr = qctx.OnReadDocDecodeIntoNewPtr
	}
	if rowcur.Cursor, err = me.Database.Query(ctx, query, me.bindVarsFrom(args)); err == nil {
		rows, rowcur.conn, rowcur.ctx = &rowcur, me, ctx
	}
	return
}

// RowsCursor is returned by our `Conn`s' `sqldrv.QueryerContext` implementation.
// (Although not explicitly stated (to circumvent `Close` method collision) in its
// `interface` type-def, it also always implements the `arango.Cursor` interface.)
//
// If `Driver.OnRowCursorReadDocumentIntoPtr` is set, it is called in
// each `Next()` iteration to fill the `ColNameDoc`-named row-cell with
// a well-typed (rather than generic `map[string]interface{}`) value.
type RowsCursor interface {
	sqldrv.Rows
	Conn() Conn
	Context() context.Context
	// only if `wantCountInRowsCursor` in your `Query`
	Count() int64
}

type arangoRowsCursor struct {
	arango.Cursor
	ctx                 context.Context
	conn                *arangoConn
	onReadDocIntoNewPtr func(RowsCursor) interface{}
	eof                 bool
}

var rowsCursorColumns = []string{ColNameDoc, ColNameMeta}

// Columns implements `sqldrv.Rows`
func (me *arangoRowsCursor) Columns() (cols []string) { return rowsCursorColumns }
func (me *arangoRowsCursor) Conn() Conn               { return me.conn }
func (me *arangoRowsCursor) Context() context.Context { return me.ctx }

// Next implements `sqldrv.Rows`
func (me *arangoRowsCursor) Next(cells []sqldrv.Value) (err error) {
	if !me.eof {
		me.eof = !me.HasMore()
	}
	if !me.eof {
		var obj interface{}
		if me.onReadDocIntoNewPtr != nil {
			obj = me.onReadDocIntoNewPtr(me)
		} else {
			foo := map[string]interface{}{}
			obj = &foo
		}
		var meta arango.DocumentMeta
		if meta, err = me.ReadDocument(me.ctx, obj); err == nil {
			cells[0], cells[1] = obj, meta
		} else {
			me.eof = arango.IsNoMoreDocuments(err)
		}
	}
	if me.eof {
		err = io.EOF // to conform with sqldrv.Rows
	}
	return
}
