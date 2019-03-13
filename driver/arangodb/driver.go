// Package `db/driver/arangodb` is a `database/sql`-compatible wrapper over ArangoDB's Go drivers.
//
// Doc-comments throughout this document refer to the following named imports:
//    import ( sqldrv "database/sql/driver" ; arango "github.com/arangodb/go-driver" )
//
package usqldrv_arango

import (
	"context"
	sqldrv "database/sql/driver"
	"errors"
	"sync"
	"time"

	arango "github.com/arangodb/go-driver"
	arangohttp "github.com/arangodb/go-driver/http"
)

type none struct{}

func (me none) UnmarshalJSON([]byte) error   { return nil }
func (me none) MarshalJSON() ([]byte, error) { return []byte("{}"), nil }

// Driver implements `sqldrv.Driver` and `sqldrv.DriverContext`.
// Its fields should be set-up **once** at initialization time (_before_
// opening connections) and not be subsequently modified. That is: for
// later connects with a different config, use a new and different `Driver`.
type Driver struct {
	// as per arango.ClientConfig
	Authentication arango.Authentication
	// as per arango.ClientConfig
	SynchronizeEndpointsInterval time.Duration

	// Endpoints, TLS, etc..
	Config arangohttp.ConnectionConfig

	shared struct {
		arango.Client
		sync.Mutex
	}
}

type connector struct {
	drv    *Driver
	dbName string
}

// Open implements `sqldrv.Driver`, but always fails due to being deprecated in favour of `OpenConnector`.
func (me *Driver) Open(name string) (sqldrv.Conn, error) {
	return nil, errors.New("bug: `Open` should never be called since `usqldrv_arango.Driver` also implements `database/sql/driver.DriverContext`")
}

// OpenConnector implements `sqldrv.DriverContext`.
// In the absence of `error`s, the `Connect` method of the returned
// `sqldrv.Connector` always returns `usqldrv_arango.Conn`s.
func (me *Driver) OpenConnector(name string) (sqldrv.Connector, error) {
	return connector{drv: me, dbName: name}, nil
}

func (me *Driver) ensureClientConn() (err error) {
	me.shared.Lock()
	if me.shared.Client == nil {
		var httpconn arango.Connection
		if httpconn, err = arangohttp.NewConnection(me.Config); err == nil {
			clientConfig := arango.ClientConfig{
				Connection:                   httpconn,
				Authentication:               me.Authentication,
				SynchronizeEndpointsInterval: me.SynchronizeEndpointsInterval,
			}
			me.shared.Client, err = arango.NewClient(clientConfig)
		}
		if err != nil {
			me.shared.Client = nil
		}
	}
	me.shared.Unlock()
	return
}

// Driver implements `sqldrv.Connector`
func (me connector) Driver() sqldrv.Driver {
	return me.drv
}

// Connect implements `sqldrv.Connector`
func (me connector) Connect(ctx context.Context) (conn sqldrv.Conn, err error) {
	if err = me.drv.ensureClientConn(); err == nil {
		var self arangoConn
		if self.Database, err = me.drv.shared.Client.Database(ctx, me.dbName); err == nil {
			self.drv, conn = me.drv, &self
		}
	}
	return
}
