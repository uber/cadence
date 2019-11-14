// Copyright (c) 2017 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package sql

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/uber/cadence/tools/common/schema"
)

type (
	sqlConnectParams struct {
		host       string
		port       int
		user       string
		password   string
		database   string
		driverName string
	}
	sqlConn struct {
		driver   Driver
		database string
		db       *sqlx.DB
	}

	// Driver is the driver interface that each SQL database needs to implement
	Driver interface {
		GetDriverName() string
		CreateDBConnection(driverName, host string, port int, user string, passwd string, database string) (*sqlx.DB, error)
		GetReadSchemaVersionSQL() string
		GetWriteSchemaVersionSQL() string
		GetWriteSchemaUpdateHistorySQL() string
		GetCreateSchemaVersionTableSQL() string
		GetCreateSchemaUpdateHistoryTableSQL() string
		GetCreateDatabaseSQL() string
		GetDropDatabaseSQL() string
		GetListTablesSQL() string
		GetDropTableSQL() string
	}
)

var supportedSQLDrivers = map[string]Driver{}

var _ schema.DB = (*sqlConn)(nil)

// RegisterDriver will register a SQL driver for SQL client CLI
func RegisterDriver(driverName string, driver Driver) {
	if _, ok := supportedSQLDrivers[driverName]; ok {
		panic("driver " + driverName + " already registered")
	}
	supportedSQLDrivers[driverName] = driver
}

func newConnection(params *sqlConnectParams) (*sqlConn, error) {
	driver, ok := supportedSQLDrivers[params.driverName]

	if !ok {
		return nil, fmt.Errorf("not supported driver %v, only supported: %v", params.driverName, supportedSQLDrivers)
	}

	db, err := driver.CreateDBConnection(params.driverName, params.host, params.port, params.user, params.password, params.database)
	if err != nil {
		return nil, err
	}
	return &sqlConn{
		db:       db,
		database: params.database,
		driver:   driver,
	}, nil
}

// CreateSchemaVersionTables sets up the schema version tables
func (c *sqlConn) CreateSchemaVersionTables() error {
	if err := c.Exec(c.driver.GetCreateSchemaVersionTableSQL()); err != nil {
		return err
	}
	return c.Exec(c.driver.GetCreateSchemaUpdateHistoryTableSQL())
}

// ReadSchemaVersion returns the current schema version for the keyspace
func (c *sqlConn) ReadSchemaVersion() (string, error) {
	var version string
	err := c.db.Get(&version, c.driver.GetReadSchemaVersionSQL(), c.database)
	return version, err
}

// UpdateShemaVersion updates the schema version for the keyspace
func (c *sqlConn) UpdateSchemaVersion(newVersion string, minCompatibleVersion string) error {
	return c.Exec(c.driver.GetWriteSchemaVersionSQL(), c.database, time.Now(), newVersion, minCompatibleVersion)
}

// WriteSchemaUpdateLog adds an entry to the schema update history table
func (c *sqlConn) WriteSchemaUpdateLog(oldVersion string, newVersion string, manifestMD5 string, desc string) error {
	now := time.Now().UTC()
	return c.Exec(c.driver.GetWriteSchemaUpdateHistorySQL(), now.Year(), int(now.Month()), now, oldVersion, newVersion, manifestMD5, desc)
}

//// Exec executes a sql statement
func (c *sqlConn) Exec(stmt string, args ...interface{}) error {
	_, err := c.db.Exec(stmt, args...)
	return err
}

// ListTables returns a list of tables in this database
func (c *sqlConn) ListTables() ([]string, error) {
	var tables []string
	err := c.db.Select(&tables, c.driver.GetListTablesSQL(), c.database)
	return tables, err
}

// DropTable drops a given table from the database
func (c *sqlConn) DropTable(name string) error {
	return c.Exec(c.driver.GetDropTableSQL(), name)
}

// DropAllTables drops all tables from this database
func (c *sqlConn) DropAllTables() error {
	tables, err := c.ListTables()
	if err != nil {
		return err
	}
	for _, tab := range tables {
		if err := c.DropTable(tab); err != nil {
			return err
		}
	}
	return nil
}

// CreateDatabase creates a database if it doesn't exist
func (c *sqlConn) CreateDatabase(name string) error {
	return c.Exec(c.driver.GetCreateDatabaseSQL(), name)
}

// DropDatabase drops a database
func (c *sqlConn) DropDatabase(name string) error {
	return c.Exec(c.driver.GetDropDatabaseSQL(), name)
}

// Close closes the sql client
func (c *sqlConn) Close() {
	if c.db != nil {
		c.db.Close()
	}
}
