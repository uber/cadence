// Copyright (c) 2017-2021 Uber Technologies, Inc.
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

package public

import (
	"context"
	"errors"
	"strings"

	gogocql "github.com/gocql/gocql"

	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin/cassandra/gocql"
)

var _ gocql.Client = client{}

type (
	client struct {
	}
)

func init() {
	gocql.RegisterClient(client{})
}

func (c client) CreateSession(
	config gocql.ClusterConfig,
) (gocql.Session, error) {
	return gocql.NewSession(config)
}

func (c client) IsTimeoutError(err error) bool {
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	if errors.Is(err, gogocql.ErrTimeoutNoResponse) {
		return true
	}
	if errors.Is(err, gogocql.ErrConnectionClosed) {
		return true
	}
	return errors.As(err, new(*gogocql.RequestErrWriteTimeout))
}

func (c client) IsNotFoundError(err error) bool {
	return err == gogocql.ErrNotFound
}

func (c client) IsThrottlingError(err error) bool {
	var req gogocql.RequestError
	if errors.As(err, &req) {
		// gocql does not expose the constant errOverloaded = 0x1001
		return req.Code() == 0x1001
	}
	return false
}

func (c client) IsDBUnavailableError(err error) bool {
	var req gogocql.RequestError
	if errors.As(err, &req) {
		// 0x1000 == UNAVAILABLE
		if req.Code() != 0x1000 {
			return false
		}
		if strings.Contains(req.Message(), "Cannot perform LWT operation") {
			return true
		}
	}
	return false
}

func (c client) IsCassandraConsistencyError(err error) bool {
	var req gogocql.RequestError
	if errors.As(err, &req) {
		// 0x1000 == UNAVAILABLE
		return req.Code() == 0x1000
	}
	return false
}
