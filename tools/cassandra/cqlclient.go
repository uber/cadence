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

package cassandra

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/uber/cadence/common/config"
	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin/cassandra/gocql"
	"github.com/uber/cadence/tools/common/schema"
)

type (
	CqlClient struct {
		nReplicas int
		session   gocql.Session
		cfg       *CQLClientConfig
	}

	// CQLClientConfig contains the configuration for cql client
	CQLClientConfig struct {
		Hosts                 string
		Port                  int
		User                  string
		Password              string
		AllowedAuthenticators []string
		Keyspace              string
		Timeout               int
		NumReplicas           int
		ProtoVersion          int
		TLS                   *config.TLS
	}
)

var errNoHosts = errors.New("Cassandra Hosts list is empty or malformed")
var errGetSchemaVersion = errors.New("Failed to get current schema version from cassandra")

const (
	DefaultTimeout       = 30 // Timeout in seconds
	DefaultCassandraPort = 9042
	SystemKeyspace       = "system"
)

const (
	readSchemaVersionCQL        = `SELECT curr_version from schema_version where keyspace_name=?`
	listTablesCQL               = `SELECT table_name from system_schema.tables where keyspace_name=?`
	listTypesCQL                = `SELECT type_name from system_schema.types where keyspace_name=?`
	writeSchemaVersionCQL       = `INSERT into schema_version(keyspace_name, creation_time, curr_version, min_compatible_version) VALUES (?,?,?,?)`
	writeSchemaUpdateHistoryCQL = `INSERT into schema_update_history(year, month, update_time, old_version, new_version, manifest_md5, description) VALUES(?,?,?,?,?,?,?)`

	createSchemaVersionTableCQL = `CREATE TABLE schema_version(keyspace_name text PRIMARY KEY, ` +
		`creation_time timestamp, ` +
		`curr_version text, ` +
		`min_compatible_version text);`

	createSchemaUpdateHistoryTableCQL = `CREATE TABLE schema_update_history(` +
		`year int, ` +
		`month int, ` +
		`update_time timestamp, ` +
		`description text, ` +
		`manifest_md5 text, ` +
		`new_version text, ` +
		`old_version text, ` +
		`PRIMARY KEY ((year, month), update_time));`

	createKeyspaceCQL = `CREATE KEYSPACE IF NOT EXISTS %v ` +
		`WITH replication = { 'class' : 'SimpleStrategy', 'replication_factor' : %v};`
)

var _ schema.SchemaClient = (*CqlClient)(nil)

// NewCQLClient returns a new instance of CQLClient
func NewCQLClient(cfg *CQLClientConfig) (*CqlClient, error) {
	var err error

	cqlClient := new(CqlClient)
	cqlClient.cfg = cfg
	cqlClient.nReplicas = cfg.NumReplicas
	cqlClient.session, err = gocql.GetRegisteredClient().CreateSession(gocql.ClusterConfig{
		Hosts:                 cfg.Hosts,
		Port:                  cfg.Port,
		User:                  cfg.User,
		Password:              cfg.Password,
		AllowedAuthenticators: cfg.AllowedAuthenticators,
		Keyspace:              cfg.Keyspace,
		TLS:                   cfg.TLS,
		Timeout:               time.Duration(cfg.Timeout) * time.Second,
		ProtoVersion:          cfg.ProtoVersion,
		Consistency:           gocql.All,
	})
	if err != nil {
		return nil, err
	}
	return cqlClient, nil
}

func (client *CqlClient) CreateDatabase(name string) error {
	return client.CreateKeyspace(name)
}

func (client *CqlClient) DropDatabase(name string) error {
	return client.DropKeyspace(name)
}

// createKeyspace creates a cassandra Keyspace if it doesn't exist
func (client *CqlClient) CreateKeyspace(name string) error {
	return client.ExecDDLQuery(fmt.Sprintf(createKeyspaceCQL, name, client.nReplicas))
}

// DropKeyspace drops a Keyspace
func (client *CqlClient) DropKeyspace(name string) error {
	return client.ExecDDLQuery(fmt.Sprintf("DROP KEYSPACE %v", name))
}

func (client *CqlClient) DropAllTables() error {
	return client.dropAllTablesTypes()
}

// CreateSchemaVersionTables sets up the schema version tables
func (client *CqlClient) CreateSchemaVersionTables() error {
	if err := client.ExecDDLQuery(createSchemaVersionTableCQL); err != nil {
		return err
	}
	return client.ExecDDLQuery(createSchemaUpdateHistoryTableCQL)
}

// ReadSchemaVersion returns the current schema version for the Keyspace
func (client *CqlClient) ReadSchemaVersion() (string, error) {
	query := client.session.Query(readSchemaVersionCQL, client.cfg.Keyspace)
	iter := query.Iter()
	var version string
	if !iter.Scan(&version) {
		iter.Close()
		return "", errGetSchemaVersion
	}
	if err := iter.Close(); err != nil {
		return "", err
	}
	return version, nil
}

// UpdateShemaVersion updates the schema version for the Keyspace
func (client *CqlClient) UpdateSchemaVersion(newVersion string, minCompatibleVersion string) error {
	query := client.session.Query(writeSchemaVersionCQL, client.cfg.Keyspace, time.Now(), newVersion, minCompatibleVersion)
	return query.Exec()
}

// WriteSchemaUpdateLog adds an entry to the schema update history table
func (client *CqlClient) WriteSchemaUpdateLog(oldVersion string, newVersion string, manifestMD5 string, desc string) error {
	now := time.Now().UTC()
	query := client.session.Query(writeSchemaUpdateHistoryCQL)
	query.Bind(now.Year(), int(now.Month()), now, oldVersion, newVersion, manifestMD5, desc)
	return query.Exec()
}

// ExecDDLQuery executes a cql statement
func (client *CqlClient) ExecDDLQuery(stmt string, args ...interface{}) error {
	return client.session.Query(stmt, args...).Exec()
}

// Close closes the cql client
func (client *CqlClient) Close() {
	if client.session != nil {
		client.session.Close()
	}
}

// ListTables lists the table names in a Keyspace
func (client *CqlClient) ListTables() ([]string, error) {
	query := client.session.Query(listTablesCQL, client.cfg.Keyspace)
	iter := query.Iter()
	var names []string
	var name string
	for iter.Scan(&name) {
		names = append(names, name)
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}
	return names, nil
}

// listTypes lists the User defined types in a Keyspace
func (client *CqlClient) listTypes() ([]string, error) {
	qry := client.session.Query(listTypesCQL, client.cfg.Keyspace)
	iter := qry.Iter()
	var names []string
	var name string
	for iter.Scan(&name) {
		names = append(names, name)
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}
	return names, nil
}

// dropTable drops a given table from the Keyspace
func (client *CqlClient) dropTable(name string) error {
	return client.ExecDDLQuery(fmt.Sprintf("DROP TABLE %v", name))
}

// dropType drops a given type from the Keyspace
func (client *CqlClient) dropType(name string) error {
	return client.ExecDDLQuery(fmt.Sprintf("DROP TYPE %v", name))
}

// dropAllTablesTypes deletes all tables/types in the
// Keyspace without deleting the Keyspace
func (client *CqlClient) dropAllTablesTypes() error {
	tables, err := client.ListTables()
	if err != nil {
		return err
	}
	log.Printf("Dropping following tables: %v\n", tables)
	for _, table := range tables {
		err1 := client.dropTable(table)
		if err1 != nil {
			log.Printf("Error dropping table %v, err=%v\n", table, err1)
		}
	}

	types, err := client.listTypes()
	if err != nil {
		return err
	}
	log.Printf("Dropping following types: %v\n", types)
	numOfTypes := len(types)
	for i := 0; i < numOfTypes && len(types) > 0; i++ {
		var erroredTypes []string
		for _, t := range types {
			err = client.dropType(t)
			if err != nil {
				log.Printf("Error dropping type %v, err=%v\n", t, err)
				erroredTypes = append(erroredTypes, t)
			}
		}
		types = erroredTypes
	}
	if len(types) > 0 {
		return err
	}
	return nil
}
