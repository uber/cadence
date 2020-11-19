// Copyright (c) 2017-2020 Uber Technologies, Inc.
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

package gocql

import (
	"context"

	"github.com/gocql/gocql"

	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin/cassandra"
)

var _ Client = client{}

type (
	client struct{}
)

var (
	defaultClient = client{}
)

// NewClient creates a default gocql client based on the open source gocql library.
func NewClient() Client {
	return defaultClient
}

func (c client) CreateSession(
	config ClusterConfig,
) (Session, error) {
	// TODO:
	//   1. move NewCassandraCluster function into this package
	//   2. remove the dependency on common/service/config package
	cluster := cassandra.NewCassandraCluster(config.Cassandra)
	cluster.ProtoVersion = config.ProtoVersion
	cluster.Consistency = mustConvertConsistency(config.Consistency)
	cluster.SerialConsistency = mustConvertSerialConsistency(config.SerialConsistency)
	cluster.Timeout = config.Timeout
	gocqlSession, err := cluster.CreateSession()
	if err != nil {
		return nil, err
	}
	return &session{
		Session: gocqlSession,
	}, nil
}

func (c client) IsTimeoutError(err error) bool {
	if err == context.DeadlineExceeded {
		return true
	}
	if err == gocql.ErrTimeoutNoResponse {
		return true
	}
	if err == gocql.ErrConnectionClosed {
		return true
	}
	_, ok := err.(*gocql.RequestErrWriteTimeout)
	return ok
}

func (c client) IsNotFoundError(err error) bool {
	return err == gocql.ErrNotFound
}

func (c client) IsThrottlingError(err error) bool {
	if req, ok := err.(gocql.RequestError); ok {
		// gocql does not expose the constant errOverloaded = 0x1001
		return req.Code() == 0x1001
	}
	return false
}
