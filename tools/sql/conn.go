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
	"github.com/uber/cadence/common/persistence/sql/storage"
	"github.com/uber/cadence/common/service/config"

	"time"

	"github.com/jmoiron/sqlx"

	"github.com/uber/cadence/common/persistence/sql/storage/sqlshared"
	"github.com/uber/cadence/tools/common/schema"
)

type (
	// ConnectParams is the connection param
	ConnectParams struct {
		Host       string
		Port       int
		User       string
		Password   string
		Database   string
		DriverName string
	}

	// Connection is the connection to database
	Connection struct {
		driver   sqlshared.Driver
		database string
		db       *sqlx.DB
	}
)

var _ schema.DB = (*Connection)(nil)

// NewConnection creates a new connection to database
func NewConnection(params *ConnectParams) (*Connection, error) {
	driver := storage.GetDriver(params.DriverName)

	db, err := driver.CreateDBConnection(&config.SQL{
		User:params.User,
		Password:params.Password,
		DatabaseName:params.Database,
		ConnectAddr:fmt.Sprintf("%v:%v", params.Host, params.Port),
	})
	if err != nil {
		return nil, err
	}

	return &Connection{
		db:       db.GetConnection(),
		database: params.Database,
		driver:   driver,
	}, nil
}

// CreateSchemaVersionTables sets up the schema version tables
func (c *Connection) CreateSchemaVersionTables() error {
	if err := c.Exec(c.driver.CreateSchemaVersionTableQuery()); err != nil {
		return err
	}
	return c.Exec(c.driver.CreateSchemaUpdateHistoryTableQuery())
}

// ReadSchemaVersion returns the current schema version for the keyspace
func (c *Connection) ReadSchemaVersion() (string, error) {
	var version string
	err := c.db.Get(&version, c.driver.ReadSchemaVersionQuery(), c.database)
	return version, err
}

// UpdateSchemaVersion updates the schema version for the keyspace
func (c *Connection) UpdateSchemaVersion(newVersion string, minCompatibleVersion string) error {
	return c.Exec(c.driver.WriteSchemaVersionQuery(), c.database, time.Now(), newVersion, minCompatibleVersion)
}

// WriteSchemaUpdateLog adds an entry to the schema update history table
func (c *Connection) WriteSchemaUpdateLog(oldVersion string, newVersion string, manifestMD5 string, desc string) error {
	now := time.Now().UTC()
	return c.Exec(c.driver.WriteSchemaUpdateHistoryQuery(), now.Year(), int(now.Month()), now, oldVersion, newVersion, manifestMD5, desc)
}

// Exec executes a sql statement
func (c *Connection) Exec(stmt string, args ...interface{}) error {
	_, err := c.db.Exec(stmt, args...)
	return err
}

// ListTables returns a list of tables in this database
func (c *Connection) ListTables() ([]string, error) {
	var tables []string
	err := c.db.Select(&tables, fmt.Sprintf(c.driver.ListTablesQuery(), c.database))
	return tables, err
}

// DropTable drops a given table from the database
func (c *Connection) DropTable(name string) error {
	return c.Exec(fmt.Sprintf(c.driver.DropTableQuery(), name))
}

// DropAllTables drops all tables from this database
func (c *Connection) DropAllTables() error {
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
func (c *Connection) CreateDatabase(name string) error {
	return c.Exec(fmt.Sprintf(c.driver.CreateDatabaseQuery(), name))
}

// DropDatabase drops a database
func (c *Connection) DropDatabase(name string) error {
	return c.Exec(fmt.Sprintf(c.driver.DropDatabaseQuery(), name))
}

// Close closes the sql client
func (c *Connection) Close() {
	if c.db != nil {
		c.db.Close()
	}
}
