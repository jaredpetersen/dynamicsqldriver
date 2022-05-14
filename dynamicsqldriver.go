package dynamicsqldriver

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

const usernameDSN = "genusername"
const passwordDSN = "genpassword"

// Credentials stores database credentials.
type Credentials struct {
	Username   string
	Password   string
	Expiration time.Time
}

// CredentialsGenerator generates new database credentials.
type CredentialsGenerator interface {
	Generate() (Credentials, error)
}

// Driver implements driver.Driver, retrieving database credentials and supports retrieving connection information from Vault.
type Driver struct {
	Actual               driver.Driver
	CredentialsGenerator CredentialsGenerator
}

// Open generates new database credentials using Vault and updates the provided DSN before opening a connection using
// actual driver.
func (d Driver) Open(dsn string) (driver.Conn, error) {
	// Only generate credentials if the DSN supports it
	if strings.Contains(dsn, usernameDSN) || strings.Contains(dsn, passwordDSN) {
		creds, err := d.CredentialsGenerator.Generate()
		if err != nil {
			return nil, fmt.Errorf("failed to generate database credentials: %w", err)
		}

		dsn = strings.Replace(dsn, usernameDSN, creds.Username, -1)
		dsn = strings.Replace(dsn, passwordDSN, creds.Password, -1)
	}

	return d.Actual.Open(dsn)
}

// Register registers a new Driver with database/sql so that it may be used with sql.Open.
func Register(actualDriver driver.Driver, gen CredentialsGenerator) {
	wrapDriver := Driver{
		Actual:               actualDriver,
		CredentialsGenerator: gen,
	}

	sql.Register("dynamicsql", &wrapDriver)
}
