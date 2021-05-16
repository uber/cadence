// Copyright (c) 2020 Uber Technologies, Inc.
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

	"github.com/uber/cadence/common/config"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin"
	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin/cassandra/gocql"
)

var (
	errConditionFailed = errors.New("internal condition fail error")
)

// cdb represents a logical connection to Cassandra database
type cdb struct {
	logger  log.Logger
	client  gocql.Client
	session gocql.Session
}

var _ nosqlplugin.DB = (*cdb)(nil)

// NewCassandraDBFromSession returns a DB from a session
func NewCassandraDBFromSession(client gocql.Client, session gocql.Session, logger log.Logger) nosqlplugin.DB {
	return &cdb{
		client:  client,
		session: session,
		logger:  logger,
	}
}

// NewCassandraDB return a new DB
func NewCassandraDB(cfg config.Cassandra, logger log.Logger) (nosqlplugin.DB, error) {
	session, err := CreateSession(cfg)
	if err != nil {
		return nil, err
	}
	return &cdb{
		client:  gocql.NewClient(),
		session: session,
		logger:  logger,
	}, nil
}

func (db *cdb) Close() {
	if db.session != nil {
		db.session.Close()
	}
}

func (db *cdb) PluginName() string {
	return PluginName
}

func (db *cdb) IsNotFoundError(err error) bool {
	return db.client.IsNotFoundError(err)
}

func (db *cdb) IsTimeoutError(err error) bool {
	return db.client.IsTimeoutError(err)
}

func (db *cdb) IsThrottlingError(err error) bool {
	return db.client.IsThrottlingError(err)
}

func (db *cdb) IsConditionFailedError(err error) bool {
	if err == errConditionFailed {
		return true
	}
	return false
}
