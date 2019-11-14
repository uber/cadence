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

	"github.com/iancoleman/strcase"
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
		driverName string
		database string
		db       *sqlx.DB
	}
)

// DriverName refers to the name of the mysql driver
const MySQLDriverName = "mysql"

var SupportedSQLDrivers = map[string]bool{
	MySQLDriverName : true,
}

var NewConnectionFuncs = map[string] func (driverName, host string, port int, user string, passwd string, database string) (*sqlx.DB, error){
	MySQLDriverName : newMySQLConn,
}

var ReadSchemaVersionSQL = map[string]string{
	MySQLDriverName:readSchemaVersionMySQL,
}

var WriteSchemaVersionSQL = map[string]string{
	MySQLDriverName:writeSchemaVersionMySQL,
}

var WriteSchemaUpdateHistorySQL = map[string]string{
	MySQLDriverName:writeSchemaUpdateHistoryMySQL,
}

var CreateSchemaVersionTableSQL = map[string]string{
	MySQLDriverName:createSchemaVersionTableMySQL,
}

var CreateSchemaUpdateHistoryTableSQL = map[string]string{
	MySQLDriverName:createSchemaUpdateHistoryTableMySQL,
}

var CreateDatabaseSQL = map[string]string{
	MySQLDriverName:createDatabaseMySQL,
}

var DropDatabaseSQL = map[string]string{
	MySQLDriverName:dropDatabaseMySQL,
}

var ListTablesSQL = map[string]string{
	MySQLDriverName:listTablesMySQL,
}

var DropTableSQL = map[string]string{
	MySQLDriverName:dropTableMySQL,
}

const (
	dataSourceNameMySQL = "%s:%s@%v(%v:%v)/%s?multiStatements=true&parseTime=true&clientFoundRows=true"

	readSchemaVersionMySQL        = `SELECT curr_version from schema_version where db_name=?`

	writeSchemaVersionMySQL       = `REPLACE into schema_version(db_name, creation_time, curr_version, min_compatible_version) VALUES (?,?,?,?)`

	writeSchemaUpdateHistoryMySQL = `INSERT into schema_update_history(year, month, update_time, old_version, new_version, manifest_md5, description) VALUES(?,?,?,?,?,?,?)`

	createSchemaVersionTableMySQL = `CREATE TABLE schema_version(db_name VARCHAR(255) not null PRIMARY KEY, ` +
		`creation_time DATETIME(6), ` +
		`curr_version VARCHAR(64), ` +
		`min_compatible_version VARCHAR(64));`

	createSchemaUpdateHistoryTableMySQL = `CREATE TABLE schema_update_history(` +
		`year int not null, ` +
		`month int not null, ` +
		`update_time DATETIME(6) not null, ` +
		`description VARCHAR(255), ` +
		`manifest_md5 VARCHAR(64), ` +
		`new_version VARCHAR(64), ` +
		`old_version VARCHAR(64), ` +
		`PRIMARY KEY (year, month, update_time));`


	createDatabaseMySQL = "CREATE database %v CHARACTER SET UTF8"

	dropDatabaseMySQL = "Drop database %v"

	listTablesMySQL = "SHOW TABLES FROM %v"
	
	dropTableMySQL = "DROP TABLE %v"
)

var _ schema.DB = (*sqlConn)(nil)

func switcher(templates map[string]string, driver string) string{
	s, ok := templates[driver]
	if !ok{
		panic(fmt.Sprintf("Template for %v is not defined", driver))
	}
	return s
}

// newSQLConn returns a new connection to mysql database
func newMySQLConn(driverName, host string, port int, user string, passwd string, database string) (*sqlx.DB, error) {
	db, err := sqlx.Connect(driverName, fmt.Sprintf(dataSourceNameMySQL, user, passwd, "tcp", host, port, database))

	if err != nil {
		return nil, err
	}
	// Maps struct names in CamelCase to snake without need for db struct tags.
	db.MapperFunc(strcase.ToSnake)
	return db, nil
}

func NewConnection(params *sqlConnectParams) (*sqlConn, error) {
	if !SupportedSQLDrivers[params.driverName]{
		return nil, fmt.Errorf("not supported driver %v, only supported: %v", params.driverName, SupportedSQLDrivers)
	}

	db, err := NewConnectionFuncs[params.driverName](params.driverName, params.host, params.port, params.user, params.password, params.database)
	if err != nil {
		return nil, err
	}
	return &sqlConn{
		db: db,
		database: params.database,
	 	driverName:params.driverName,
	}, nil
}

// CreateSchemaVersionTables sets up the schema version tables
func (c *sqlConn) CreateSchemaVersionTables() error {
	if err := c.Exec(switcher(CreateSchemaVersionTableSQL,c.driverName)); err != nil {
		return err
	}
	return c.Exec(switcher(CreateSchemaUpdateHistoryTableSQL,c.driverName))
}

// ReadSchemaVersion returns the current schema version for the keyspace
func (c *sqlConn) ReadSchemaVersion() (string, error) {
	var version string
	sql := switcher(ReadSchemaVersionSQL,c.driverName)
	err := c.db.Get(&version, sql, c.database)
	return version, err
}

// UpdateShemaVersion updates the schema version for the keyspace
func (c *sqlConn) UpdateSchemaVersion(newVersion string, minCompatibleVersion string) error {
	_, err := c.db.Exec(switcher(WriteSchemaVersionSQL,c.driverName), c.database, time.Now(), newVersion, minCompatibleVersion)
	return err
}

// WriteSchemaUpdateLog adds an entry to the schema update history table
func (c *sqlConn) WriteSchemaUpdateLog(oldVersion string, newVersion string, manifestMD5 string, desc string) error {
	now := time.Now().UTC()
	_, err := c.db.Exec(switcher(WriteSchemaUpdateHistorySQL,c.driverName), now.Year(), int(now.Month()), now, oldVersion, newVersion, manifestMD5, desc)
	return err
}

// Exec executes a sql statement
func (c *sqlConn) Exec(stmt string) error {
	_, err := c.db.Exec(stmt)
	return err
}

// ListTables returns a list of tables in this database
func (c *sqlConn) ListTables() ([]string, error) {
	var tables []string
	err := c.db.Select(&tables, fmt.Sprintf(switcher(ListTablesSQL,c.driverName), c.database))
	return tables, err
}

// DropTable drops a given table from the database
func (c *sqlConn) DropTable(name string) error {
	return c.Exec(fmt.Sprintf(switcher(DropTableSQL,c.driverName), name))
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
	return c.Exec(fmt.Sprintf(switcher(CreateDatabaseSQL,c.driverName), name))
}

// DropDatabase drops a database
func (c *sqlConn) DropDatabase(name string) error {
	return c.Exec(fmt.Sprintf(switcher(DropDatabaseSQL,c.driverName), name))
}

// Close closes the sql client
func (c *sqlConn) Close() {
	if c.db != nil {
		c.db.Close()
	}
}
