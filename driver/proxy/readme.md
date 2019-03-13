# usqldrv_proxy
--
    import "github.com/go-leap/db/driver/proxy"

Package `db/driver/proxy` provides event-handling hooks (for logging, metrics,
tracing, diagnostics or whatever purposes) into any (wrapped via `WrapDriver` or
`WrapConnector`) `database/sql/driver`'s noteworthy operations / situations like
so:

    // import proxy "github.com/go-leap/db/driver/proxy"
    // // import sqldrv "database/sql/driver"

    proxy.On.Connector.Connect.Before =

      func(this proxy.This, args ...interface{}) {
          connectionString := args[0].(string)
          // connector := this.(sqldrv.Connector)
          log.Print("DB connection attempt to: " + connectionString)
      }

## Usage

```go
var On struct {
	Driver struct {
		Open          Hook
		OpenConnector Hook
	}
	Connector struct {
		Connect Hook
	}
	Conn struct {
		Begin           Hook
		BeginTx         Hook
		CheckNamedValue Hook
		Close           Hook
		Exec            Hook
		ExecContext     Hook
		Prepare         Hook
		PrepareContext  Hook
		Query           Hook
		QueryContext    Hook
	}
	Rows struct {
		Close         Hook
		Next          Hook
		NextResultSet Hook
	}
	Stmt struct {
		CheckNamedValue Hook
		Close           Hook
		Exec            Hook
		ExecContext     Hook
		Query           Hook
		QueryContext    Hook
	}
	Tx struct {
		Commit   Hook
		Rollback Hook
	}
}
```
On lets you set custom `Hook.Before`, `Hook.Failed` and/or `Hook.Success`
handlers for noteworthy `database/sql/driver` operations and situations.

It is not explicitly protected against concurrent accesses, so best to use it as
a read-only with all writes to it at `init`-time prior to concurrent use.

#### func  WrapConnector

```go
func WrapConnector(connector sqldrv.Connector) sqldrv.Connector
```
WrapConnector returns a new proxy `database/sql/driver.Connector` wrapping
`connector` and relaying subscribed events to your `Hook` handlers set in `On`
(except those set in `On.Driver`, which only trigger for `WrapDriver`s).

#### func  WrapDriver

```go
func WrapDriver(drv sqldrv.Driver) sqldrv.Driver
```
WrapDriver returns a new proxy `database/sql/driver.Driver` wrapping `drv` and
relaying subscribed events to your `Hook` handlers set in `On`.

#### type Bag

```go
type Bag struct {
	This
	Tag
}
```

Bag is passed to your `Hook.Failed` and `Hook.Success` handlers. Its `Tag` is
your corresponding `Hook.Before` handler's return value, its `This` is the same
as was passed to that `Hook.Before` handler.

#### type Hook

```go
type Hook struct {
	Before  func(This, ...interface{}) Tag
	Failed  func(*Bag, error)
	Success func(*Bag, ...interface{})
}
```

Hook lets you set handlers for some `database/sql/driver` scenario via the
global `On` struct's sub-structs.

`Before`'s var-args following `This` are all the current-operation's args
(compare corresponding `database/sql/driver` interface methods), eg: for
`On.Connector.Connect.Before`, following `This` would merely be the single
`string` that was passed as `name` to the `Connector.Connect(string)` call.

`Failed` is called if the current-operation will return an `error`. Otherwise,
`Success` is called with one "result" arg or none at all.

#### type Tag

```go
type Tag interface{}
```

Tag is returned by your `Hook.Before` handlers for reuse in your corresponding
`Hook.Failed` / `Hook.Success` handlers.

#### type This

```go
type This interface{}
```

This is whatever method receiver is current for the `Hook`: for those in
`On.Conn` it would be a `database/sql/driver.Conn`, for those in `On.Tx` it
would be `database/sql/driver.Tx`, etc.
