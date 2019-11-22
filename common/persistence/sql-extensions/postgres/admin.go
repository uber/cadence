// Copyright (c) 2019 Uber Technologies, Inc.
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

package postgres

const (
	readSchemaVersionPostgres = `SELECT curr_version from schema_version where db_name=$1`

	writeSchemaVersionPostgres = `INSERT into schema_version(db_name, creation_time, curr_version, min_compatible_version) VALUES ($1,$2,$3,$4)
										ON CONFLICT (db_name) DO UPDATE 
										  SET creation_time = excluded.creation_time,
										   	  curr_version = excluded.curr_version,
										      min_compatible_version = excluded.min_compatible_version;`

	writeSchemaUpdateHistoryPostgres = `INSERT into schema_update_history(year, month, update_time, old_version, new_version, manifest_md5, description) VALUES($1,$2,$3,$4,$5,$6,$7)`

	createSchemaVersionTablePostgres = `CREATE TABLE schema_version(db_name VARCHAR(255) not null PRIMARY KEY, ` +
		`creation_time TIMESTAMP, ` +
		`curr_version VARCHAR(64), ` +
		`min_compatible_version VARCHAR(64));`

	createSchemaUpdateHistoryTablePostgres = `CREATE TABLE schema_update_history(` +
		`year int not null, ` +
		`month int not null, ` +
		`update_time TIMESTAMP not null, ` +
		`description VARCHAR(255), ` +
		`manifest_md5 VARCHAR(64), ` +
		`new_version VARCHAR(64), ` +
		`old_version VARCHAR(64), ` +
		`PRIMARY KEY (year, month, update_time));`

	//NOTE we have to use %v because somehow mysql doesn't work with ? here
	createDatabasePostgres = "CREATE database %v"

	dropDatabasePostgres = "Drop database %v"

	listTablesPostgres = "select table_name from information_schema.tables where table_schema='public'"

	dropTablePostgres = "DROP TABLE %v"
)

func (d *driver) ReadSchemaVersionQuery() string {
	return readSchemaVersionPostgres
}

func (d *driver) WriteSchemaVersionQuery() string {
	return writeSchemaVersionPostgres
}

func (d *driver) WriteSchemaUpdateHistoryQuery() string {
	return writeSchemaUpdateHistoryPostgres
}

func (d *driver) CreateSchemaVersionTableQuery() string {
	return createSchemaVersionTablePostgres
}

func (d *driver) CreateSchemaUpdateHistoryTableQuery() string {
	return createSchemaUpdateHistoryTablePostgres
}

func (d *driver) CreateDatabaseQuery() string {
	return createDatabasePostgres
}

func (d *driver) DropDatabaseQuery() string {
	return dropDatabasePostgres
}

func (d *driver) ListTablesQuery() string {
	return listTablesPostgres
}

func (d *driver) DropTableQuery() string {
	return dropTablePostgres
}
