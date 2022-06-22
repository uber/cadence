// Copyright (c) 2017-2020 Uber Technologies, Inc.
// Portions of the Software are attributed to Copyright (c) 2020 Temporal Technologies Inc.
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
	"sync"
	"sync/atomic"
	"time"

	"github.com/gocql/gocql"

	"github.com/uber/cadence/common"
)

var _ Session = (*session)(nil)

const (
	sessionRefreshMinInternal = 5 * time.Second
)

type (
	session struct {
		atomic.Value // *gocql.Session
		sync.Mutex

		status          int32
		config          ClusterConfig
		sessionInitTime time.Time
	}
)

func NewSession(
	config ClusterConfig,
) (Session, error) {
	gocqlSession, err := initSession(config)
	if err != nil {
		return nil, err
	}
	session := &session{
		status:          common.DaemonStatusStarted,
		config:          config,
		sessionInitTime: time.Now().UTC(),
	}
	session.Value.Store(gocqlSession)
	return session, nil
}

func initSession(
	config ClusterConfig,
) (*gocql.Session, error) {
	cluster := newCassandraCluster(config)
	cluster.Consistency = mustConvertConsistency(config.Consistency)
	cluster.SerialConsistency = mustConvertSerialConsistency(config.SerialConsistency)
	cluster.Timeout = config.Timeout
	cluster.ConnectTimeout = config.ConnectTimeout

	return cluster.CreateSession()
}

func (s *session) refresh() error {
	if atomic.LoadInt32(&s.status) != common.DaemonStatusStarted {
		return nil
	}

	s.Lock()
	defer s.Unlock()

	if time.Now().UTC().Sub(s.sessionInitTime) < sessionRefreshMinInternal {
		return nil
	}

	newSession, err := initSession(s.config)
	if err != nil {
		return err
	}

	s.sessionInitTime = time.Now().UTC()
	oldSession := s.Value.Load().(*gocql.Session)
	s.Value.Store(newSession)
	oldSession.Close()
	return nil
}

func (s *session) Query(
	stmt string,
	values ...interface{},
) Query {
	q := s.Value.Load().(*gocql.Session).Query(stmt, values...)
	if q == nil {
		return nil
	}
	return newQuery(s, q)
}

func (s *session) NewBatch(
	batchType BatchType,
) Batch {
	b := s.Value.Load().(*gocql.Session).NewBatch(mustConvertBatchType(batchType))
	if b == nil {
		return nil
	}
	return newBatch(b)
}

func (s *session) ExecuteBatch(
	b Batch,
) error {
	err := s.Value.Load().(*gocql.Session).ExecuteBatch(b.(*batch).Batch)
	return s.handleError(err)
}

func (s *session) MapExecuteBatchCAS(
	b Batch,
	previous map[string]interface{},
) (bool, Iter, error) {
	applied, iter, err := s.Value.Load().(*gocql.Session).MapExecuteBatchCAS(b.(*batch).Batch, previous)
	if iter == nil {
		return applied, nil, s.handleError(err)
	}
	return applied, iter, s.handleError(err)
}

func (s *session) Close() {
	if !atomic.CompareAndSwapInt32(&s.status, common.DaemonStatusStarted, common.DaemonStatusStopped) {
		return
	}

	s.Value.Load().(*gocql.Session).Close()
}

func (s *session) handleError(err error) error {
	if err == gocql.ErrNoConnections {
		_ = s.refresh()
	}
	return err
}
