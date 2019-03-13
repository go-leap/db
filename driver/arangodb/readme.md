# usqldrv_arango
--
    import "github.com/go-leap/db/driver/arangodb"

Package `db/driver/arangodb` is a `database/sql`-compatible wrapper over
ArangoDB's Go drivers.

Doc-comments throughout this document refer to the following named imports:

    import ( sqldrv "database/sql/driver" ; arango "github.com/arangodb/go-driver" )

## Usage

```go
const (
	// both returned by our `sqldrv.Rows.Columns()` implementation:
	ColNameDoc  = "_doc"  // map[string]interface{} or other (see `RowsCursor`)
	ColNameMeta = "_meta" // always an arango.DocumentMeta
)
```

#### func  Insert

```go
func Insert(coll string, returnDocKeyOnly bool, returnDocFully bool, doc interface{}) (query string, isTransact bool, err error)
```
Insert constructs a `query` (to insert `doc` into `coll`) that can be passed to
`Conn.ExecContext`.

#### func  Query

```go
func Query(ctx context.Context, wantCountInRowsCursor bool, onReadDocDecodeIntoNewPtr func(RowsCursor) interface{}) context.Context
```
Query returns a `context.Context` that can be passed to `Conn.QueryContext`.

#### func  Transact

```go
func Transact(ctx context.Context, transactionOptions *arango.TransactionOptions, onSucceeded func(result interface{})) context.Context
```
Transact returns a `context.Context` from `ctx` that can be passed to
`Conn.ExecContext` to signify that the query is an ArangoDB "transaction"
JavaScript function to be run on the server using `transactionOptions`. If
`onSucceeded` is given, it is called with either `nil` or the JS func's result.

#### type Conn

```go
type Conn interface {
	sqldrv.Conn
	sqldrv.ExecerContext
	sqldrv.QueryerContext
	arango.Database
}
```

Conn is returned by `Driver.OpenConnector(name).Connect(ctx)`.

Absent `error`s, its `sqldrv.QueryerContext` implementation returns
`RowsCursor`s that implement both `sqldrv.Rows` and `arango.Cursor`.

#### type Driver

```go
type Driver struct {
	// as per arango.ClientConfig
	Authentication arango.Authentication
	// as per arango.ClientConfig
	SynchronizeEndpointsInterval time.Duration

	// Endpoints, TLS, etc..
	Config arangohttp.ConnectionConfig
}
```

Driver implements `sqldrv.Driver` and `sqldrv.DriverContext`. Its fields should
be set-up **once** at initialization time (_before_ opening connections) and not
be subsequently modified. That is: for later connects with a different config,
use a new and different `Driver`.

#### func (*Driver) Open

```go
func (me *Driver) Open(name string) (sqldrv.Conn, error)
```
Open implements `sqldrv.Driver`, but always fails due to being deprecated in
favour of `OpenConnector`.

#### func (*Driver) OpenConnector

```go
func (me *Driver) OpenConnector(name string) (sqldrv.Connector, error)
```
OpenConnector implements `sqldrv.DriverContext`. In the absence of `error`s, the
`Connect` method of the returned `sqldrv.Connector` always returns
`usqldrv_arango.Conn`s.

#### type RowsCursor

```go
type RowsCursor interface {
	sqldrv.Rows
	Conn() Conn
	Context() context.Context
	// only if `wantCountInRowsCursor` in your `Query`
	Count() int64
}
```

RowsCursor is returned by our `Conn`s' `sqldrv.QueryerContext` implementation.
(Although not explicitly stated (to circumvent `Close` method collision) in its
`interface` type-def, it also always implements the `arango.Cursor` interface.)

If `Driver.OnRowCursorReadDocumentIntoPtr` is set, it is called in each `Next()`
iteration to fill the `ColNameDoc`-named row-cell with a well-typed (rather than
generic `map[string]interface{}`) value.
