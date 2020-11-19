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
	"time"

	"github.com/uber/cadence/common/service/config"
)

type (
	Client interface {
		CreateSession(ClusterConfig) (Session, error)

		IsTimeoutError(error) bool
		IsNotFoundError(error) bool
		IsThrottlingError(error) bool
	}

	Session interface {
		Query(string, ...interface{}) Query
		NewBatch(BatchType) Batch
		MapExecuteBatchCAS(Batch, map[string]interface{}) (bool, Iter, error)
		Close()
	}

	// no implemtation needed if we don't need Consistency
	Query interface {
		Exec() error
		Scan(...interface{}) error
		MapScan(map[string]interface{}) error
		MapScanCAS(map[string]interface{}) (bool, error)
		Iter() Iter
		PageSize(int) Query
		PageState([]byte) Query
		WithContext(context.Context) Query
		Consistency(Consistency) Query
	}

	Batch interface {
		Query(string, ...interface{})
		WithContext(context.Context) Batch
	}

	Iter interface {
		Scan(...interface{}) bool
		MapScan(map[string]interface{}) bool
		PageState() []byte
		Close() error
	}

	// no implemtation needed
	// but note that we do value.([]gocql.UUID) in our code,
	// this need to be changed to use reflection and type assert each element
	UUID interface {
		String() string
	}

	BatchType byte

	Consistency uint16

	SerialConsistency uint16

	ClusterConfig struct {
		// TODO: explicitly define all the fields here so remove the dependency on common/service/config package
		config.Cassandra

		ProtoVersion      int
		Consistency       Consistency
		SerialConsistency SerialConsistency
		Timeout           time.Duration
	}
)
