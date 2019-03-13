package usqldrv_proxy

// On lets you set custom `Hook.Before`, `Hook.Failed` and/or `Hook.Success`
// handlers for noteworthy `database/sql/driver` operations and situations.
//
// It is not explicitly protected against concurrent accesses, so best to use it
// as a read-only with all writes to it at `init`-time prior to concurrent use.
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

// Hook lets you set handlers for some `database/sql/driver` scenario
// via the global `On` struct's sub-structs.
//
// `Before`'s var-args following `This` are all the current-operation's args
// (compare corresponding `database/sql/driver` interface methods), eg: for
// `On.Connector.Connect.Before`, following `This` would merely be the single
// `string` that was passed as `name` to the `Connector.Connect(string)` call.
//
// `Failed` is called if the current-operation will return an `error`.
// Otherwise, `Success` is called with one "result" arg or none at all.
type Hook struct {
	Before  func(This, ...interface{}) Tag
	Failed  func(*Bag, error)
	Success func(*Bag, ...interface{})
}

// Bag is passed to your `Hook.Failed` and `Hook.Success` handlers.
// Its `Tag` is your corresponding `Hook.Before` handler's return value,
// its `This` is the same as was passed to that `Hook.Before` handler.
type Bag struct {
	done func(...interface{})
	This
	Tag
}

// This is whatever method receiver is current for the `Hook`:
// for those in `On.Conn` it would be a `database/sql/driver.Conn`,
// for those in `On.Tx` it would be  `database/sql/driver.Tx`, etc.
type This interface{}

// Tag is returned by your `Hook.Before` handlers for reuse in
// your corresponding `Hook.Failed` / `Hook.Success` handlers.
type Tag interface{}

var onDoneNoOp = func(...interface{}) {}

func (me Hook) begin(self This, err *error, args ...interface{}) (bag *Bag) {
	bag = &Bag{done: onDoneNoOp}
	if me.Before != nil {
		bag.Tag = me.Before(self, args...)
	}
	if me.Failed != nil || me.Success != nil {
		bag.This, bag.done = self, func(results ...interface{}) {
			if e := *err; e == nil && me.Success != nil {
				me.Success(bag, results...)
			} else if e != nil && me.Failed != nil {
				me.Failed(bag, e)
			}
		}
	}
	return
}
