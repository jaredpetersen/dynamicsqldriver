# dynamicsqldriver
[![CI](https://github.com/jaredpetersen/dynamicsqldriver/actions/workflows/ci.yaml/badge.svg)](https://github.com/jaredpetersen/dynamicsqldriver/actions/workflows/ci.yaml)
[![Go Reference](https://pkg.go.dev/badge/github.com/jaredpetersen/dynamicsqldriver.svg)](https://pkg.go.dev/github.com/jaredpetersen/dynamicsqldriver)

dynamicsqldriver is a SQL driver implementation for Go that wraps your favorite SQL driver and adds the ability to
generate credentials any time the SQL package opens up a new connection. This is particularly useful with secrets
management systems like HashiCorp Vault that generate and manage database users on your behalf.

Create your own implementation of `dynamicsqldriver.CredentialsGenerator` and set the username and password portion of
your database connection string to `genusername` and `genpassword` respectively. The generator will replace those
values in the connection string with the generated values on your behalf.

## Usage

```go
package main

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/jaredpetersen/dynamicsqldriver"
)

// Generator is an implementation of dynamicsqldriver.CredentialsGenerator that generates credentials and caches them.
type Generator struct {
	Cache dynamicsqldriver.Credentials
}

func (g *Generator) Generate() (dynamicsqldriver.Credentials, error) {
	now := time.Now()
	if now.After(g.Cache.Expiration) || now.Equal(g.Cache.Expiration) {
		g.Cache = dynamicsqldriver.Credentials{
			Username:   uuid.NewString(),
			Password:   uuid.NewString(),
			Expiration: now.Add(30 * time.Minute),
		}
	}

	return g.Cache, nil
}

func main() {
	generator := Generator{}

	dbHost := "localhost:3306"
	dbName := "mydb"

	// Specify "genusername" and "genpassword" to have the values replaced by the generator function
	dsn := fmt.Sprintf("genusername:genpassword@tcp(%s)/%s?parseTime=true", dbHost, dbName)
	db := sql.OpenDB(dynamicsqldriver.NewConnector(mysql.MySQLDriver{}, &generator, dsn))
}
```

## Install
```shell
go get github.com/jaredpetersen/dynamicsqldriver
```

