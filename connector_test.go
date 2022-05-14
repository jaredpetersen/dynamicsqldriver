package dynamicsqldriver_test

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/jaredpetersen/dynamicsqldriver"
	"github.com/stretchr/testify/assert"
)

func TestConnectorDriverReturnsDriver(t *testing.T) {
	actualDriver := fakeDriver{}
	generator := fakeCredentialsGenerator{}
	dsn := "genusername:genpassword@myhost/mydb"

	expectedWrapperDriver := dynamicsqldriver.Driver{Actual: &actualDriver, CredentialsGenerator: &generator}

	connector := dynamicsqldriver.NewConnector(&actualDriver, &generator, dsn)
	connectorDriver := connector.Driver()
	assert.Equal(t, expectedWrapperDriver, connectorDriver, "Incorrect driver")
}

func TestConnectorUsesDriver(t *testing.T) {
	dsn := "genusername:genpassword@myhost/mydb"
	creds := dynamicsqldriver.Credentials{
		Username:   "user",
		Password:   "thisisatotallysupersecretpassword",
		Expiration: time.Now().Add(10 * time.Minute),
	}

	actualDriver := fakeDriver{}
	actualDriverOpenErr := errors.New("fake driver not implemented")
	actualDriver.openFunc = func(name string) (driver.Conn, error) {
		assert.Equal(t, name, fmt.Sprintf("%s:%s@myhost/mydb", creds.Username, creds.Password), "Incorrect DSN")
		return nil, actualDriverOpenErr
	}

	generator := fakeCredentialsGenerator{}
	generator.generateFunc = func() (dynamicsqldriver.Credentials, error) {
		return creds, nil
	}

	// sql.Open doesn't actually open any connections until the connection starts seeing usage
	db := sql.OpenDB(dynamicsqldriver.NewConnector(&actualDriver, &generator, dsn))
	assert.NotEmpty(t, db, "DB is empty")

	// Start using the database
	err := db.Ping()
	assert.ErrorIs(t, err, actualDriverOpenErr, "Did not call actual driver")
}

func TestSqlOpenDBRegeneratesCredentialsWithMultipleConnectionAttempts(t *testing.T) {
	dsn := "genusername:genpassword@myhost/mydb"

	timesDriverCalled := 0
	timesGeneratorCalled := 0

	actualDriver := fakeDriver{}
	actualDriverOpenErr := errors.New("fake driver not implemented")
	actualDriver.openFunc = func(name string) (driver.Conn, error) {
		timesDriverCalled++
		expectedDSN := fmt.Sprintf("%s:%s@myhost/mydb",
			fmt.Sprintf("user-%d", timesGeneratorCalled),
			fmt.Sprintf("password-%d", timesGeneratorCalled))
		assert.Equal(t, expectedDSN, name, "Incorrect DSN")
		return nil, actualDriverOpenErr
	}

	generator := fakeCredentialsGenerator{}
	generator.generateFunc = func() (dynamicsqldriver.Credentials, error) {
		timesGeneratorCalled++
		creds := dynamicsqldriver.Credentials{
			Username:   fmt.Sprintf("user-%d", timesGeneratorCalled),
			Password:   fmt.Sprintf("password-%d", timesGeneratorCalled),
			Expiration: time.Now().Add(10 * time.Minute),
		}
		return creds, nil
	}

	// sql.Open doesn't actually open any connections until the connection starts seeing usage
	db := sql.OpenDB(dynamicsqldriver.NewConnector(&actualDriver, &generator, dsn))
	assert.NotEmpty(t, db, "DB is empty")

	// Ping the database twice to trigger multiple generations
	attemptCount := 5
	for i := 0; i < attemptCount; i++ {
		db.Ping()
	}

	assert.Equal(t, attemptCount, timesDriverCalled, "Driver not called enough")
	assert.Equal(t, attemptCount, timesGeneratorCalled, "Generator not called enough")
}
