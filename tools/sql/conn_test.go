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
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/environment"
	"github.com/uber/cadence/tools/common/schema/test"
	"github.com/uber/cadence/tools/sql/mysql"
	"github.com/uber/cadence/tools/sql/postgres"
)

type (
	MySQLConnTestSuite struct {
		test.DBTestBase
	}

	PostgresConnTestSuite struct {
		test.DBTestBase
	}
)

var _ test.DB = (*sqlConn)(nil)

const (
	testUser     = "uber"
	testPassword = "uber"
)

func TestMySQLConnTestSuite(t *testing.T) {
	suite.Run(t, new(MySQLConnTestSuite))
}

func (s *MySQLConnTestSuite) SetupTest() {
	s.Assertions = require.New(s.T()) // Have to define our overridden assertions in the test setup. If we did it earlier, s.T() will return nil
}

func (s *MySQLConnTestSuite) SetupSuite() {
	conn, err := newTestConn("")
	if err != nil {
		s.Log.Fatal("error creating sql conn, ", tag.Error(err))
	}
	s.SetupSuiteBase(conn)
}

func (s *MySQLConnTestSuite) TearDownSuite() {
	s.TearDownSuiteBase()
}

func (s *MySQLConnTestSuite) TestParseCQLFile() {
	s.RunParseFileTest(createTestSQLFileContent())
}

func (s *MySQLConnTestSuite) TestSQLConn() {
	conn, err := newConn(&sqlConnectParams{
		host:       environment.GetSQLAddress(),
		port:       environment.GetSQLPort(),
		user:       testUser,
		password:   testPassword,
		driverName: mysql.DriverName,
		database:   s.DBName,
	})
	s.Nil(err)
	s.RunCreateTest(conn)
	s.RunUpdateTest(conn)
	s.RunDropTest(conn)
	conn.Close()
}

// --

func TestPostgresConnTestSuite(t *testing.T) {
	suite.Run(t, new(PostgresConnTestSuite))
}

func (s *PostgresConnTestSuite) SetupTest() {
	s.Assertions = require.New(s.T()) // Have to define our overridden assertions in the test setup. If we did it earlier, s.T() will return nil
}

func (s *PostgresConnTestSuite) SetupSuite() {
	conn, err := newTestConn("")
	if err != nil {
		s.Log.Fatal("error creating sql conn, ", tag.Error(err))
	}
	s.SetupSuiteBase(conn)
}

func (s *PostgresConnTestSuite) TearDownSuite() {
	s.TearDownSuiteBase()
}

func (s *PostgresConnTestSuite) TestParseCQLFile() {
	s.RunParseFileTest(createTestSQLFileContent())
}

func (s *PostgresConnTestSuite) TestSQLConn() {
	conn, err := newConn(&sqlConnectParams{
		host:       environment.GetSQLAddress(),
		port:       environment.GetSQLPort(),
		user:       testUser,
		password:   testPassword,
		driverName: postgres.DriverName,
		database:   s.DBName,
	})
	s.Nil(err)
	s.RunCreateTest(conn)
	s.RunUpdateTest(conn)
	s.RunDropTest(conn)
	conn.Close()
}

// -

func newTestConn(database string) (*sqlConn, error) {
	return newConn(&sqlConnectParams{
		host:       environment.GetSQLAddress(),
		port:       environment.GetSQLPort(),
		user:       testUser,
		password:   testPassword,
		driverName: mysql.DriverName,
		database:   database,
	})
}

func createTestSQLFileContent() string {
	return `
-- test sql file content

CREATE TABLE task_maps (
  shard_id INT NOT NULL,
  domain_id BINARY(16) NOT NULL,
  workflow_id VARCHAR(255) NOT NULL,
  run_id BINARY(16) NOT NULL,
  first_event_id BIGINT NOT NULL,
--
  version BIGINT NOT NULL,
  next_event_id BIGINT NOT NULL,
  history MEDIUMBLOB,
  history_encoding VARCHAR(16) NOT NULL,
  new_run_history BLOB,
  new_run_history_encoding VARCHAR(16) NOT NULL DEFAULT 'json',
  event_store_version          INT NOT NULL, -- indiciates which version of event store to query
  new_run_event_store_version  INT NOT NULL, -- indiciates which version of event store to query for new run(continueAsNew)
  PRIMARY KEY (shard_id, domain_id, workflow_id, run_id, first_event_id)
);


CREATE TABLE tasks (
  domain_id BINARY(16) NOT NULL,
  task_list_name VARCHAR(255) NOT NULL,
  task_type TINYINT NOT NULL, -- {Activity, Decision}
  task_id BIGINT NOT NULL,
  --
  data BLOB NOT NULL,
  data_encoding VARCHAR(16) NOT NULL,
  PRIMARY KEY (domain_id, task_list_name, task_type, task_id)
);
`
}
