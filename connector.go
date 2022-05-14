package dynamicsqldriver

import (
	"context"
	"database/sql/driver"
)

// Connector implements driver.Connector so that the driver can be used with sql.OpenDB.
type Connector struct {
	dsn    string
	driver Driver
}

func (c Connector) Connect(_ context.Context) (driver.Conn, error) {
	return c.driver.Open(c.dsn)
}

func (c Connector) Driver() driver.Driver {
	return c.driver
}

// NewConnector creates a new Connector.
func NewConnector(driver driver.Driver, generator CredentialsGenerator, dsn string) Connector {
	wrapDriver := Driver{
		Actual:               driver,
		CredentialsGenerator: generator,
	}
	return Connector{dsn: dsn, driver: wrapDriver}
}
