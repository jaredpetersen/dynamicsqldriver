package dynamicsqldriver_test

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/jaredpetersen/dynamicsqldriver"
)

type fakeCredentialsGenerator struct {
	generateFunc func() (dynamicsqldriver.Credentials, error)
}

func (g *fakeCredentialsGenerator) Generate() (dynamicsqldriver.Credentials, error) {
	return g.generateFunc()
}

type fakeDriver struct {
	openFunc func(name string) (driver.Conn, error)
}

func (d *fakeDriver) Open(name string) (driver.Conn, error) {
	return d.openFunc(name)
}

func TestDriverOpenGeneratesAndReplacesCredentialsInDSN(t *testing.T) {
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

	dynamicDriver := dynamicsqldriver.Driver{
		Actual:               &actualDriver,
		CredentialsGenerator: &generator,
	}

	conn, err := dynamicDriver.Open(dsn)
	assert.ErrorIs(t, err, actualDriverOpenErr, "Did not call actual driver")
	assert.Empty(t, conn, "Conn is not empty")
}

func TestDriverOpenDoesNotGenerateCredentialsIfNoPattern(t *testing.T) {
	dsn := "username:password@myhost/mydb"

	actualDriver := fakeDriver{}
	actualDriverOpenErr := errors.New("fake driver not implemented")
	actualDriver.openFunc = func(name string) (driver.Conn, error) {
		assert.Equal(t, dsn, name, "Incorrect DSN")
		return nil, actualDriverOpenErr
	}

	generator := fakeCredentialsGenerator{}
	generator.generateFunc = func() (dynamicsqldriver.Credentials, error) {
		assert.Fail(t, "Called credentials generator")
		return dynamicsqldriver.Credentials{}, nil
	}

	dynamicDriver := dynamicsqldriver.Driver{
		Actual:               &actualDriver,
		CredentialsGenerator: &generator,
	}

	dynamicDriver.Open(dsn)
}

func TestDriverOpenReturnsErrorOnGenerateCredentialsFailure(t *testing.T) {
	dsn := "genusername:genpassword@myhost/mydb"

	actualDriver := fakeDriver{}

	genCredsErr := errors.New("uh-oh")

	generator := fakeCredentialsGenerator{}
	generator.generateFunc = func() (dynamicsqldriver.Credentials, error) {
		return dynamicsqldriver.Credentials{}, genCredsErr
	}

	dynamicDriver := dynamicsqldriver.Driver{
		Actual:               &actualDriver,
		CredentialsGenerator: &generator,
	}

	conn, err := dynamicDriver.Open(dsn)
	assert.ErrorIs(t, err, genCredsErr, "Did not call actual driver")
	assert.Empty(t, conn, "Conn is not empty")
}
