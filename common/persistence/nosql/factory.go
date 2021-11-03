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

package nosql

import (
	"sync"

	"github.com/uber/cadence/common/config"
	"github.com/uber/cadence/common/log"
	p "github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin"
)

type (
	// Factory vends datastore implementations backed by cassandra
	Factory struct {
		sync.RWMutex
		cfg              config.Cassandra
		clusterName      string
		logger           log.Logger
		execStoreFactory *executionStoreFactory
	}

	executionStoreFactory struct {
		db     nosqlplugin.DB
		logger log.Logger
	}
)

// NewFactory returns an instance of a factory object which can be used to create
// datastores that are backed by cassandra
func NewFactory(cfg config.Cassandra, clusterName string, logger log.Logger) *Factory {
	return &Factory{
		cfg:         cfg,
		clusterName: clusterName,
		logger:      logger,
	}
}

// NewTaskStore returns a new task store
func (f *Factory) NewTaskStore() (p.TaskStore, error) {
	return newNoSQLTaskStore(f.cfg, f.logger)
}

// NewShardStore returns a new shard store
func (f *Factory) NewShardStore() (p.ShardStore, error) {
	return newNoSQLShardStore(f.cfg, f.clusterName, f.logger)
}

// NewHistoryStore returns a new history store
func (f *Factory) NewHistoryStore() (p.HistoryStore, error) {
	return newNoSQLHistoryStore(f.cfg, f.logger)
}

// NewDomainStore returns a metadata store that understands only v2
func (f *Factory) NewDomainStore() (p.DomainStore, error) {
	return newNoSQLDomainStore(f.cfg, f.clusterName, f.logger)
}

// NewExecutionStore returns an ExecutionStore for a given shardID
func (f *Factory) NewExecutionStore(shardID int) (p.ExecutionStore, error) {
	factory, err := f.executionStoreFactory()
	if err != nil {
		return nil, err
	}
	return factory.new(shardID)
}

// NewVisibilityStore returns a visibility store
func (f *Factory) NewVisibilityStore(sortByCloseTime bool) (p.VisibilityStore, error) {
	return newNoSQLVisibilityStore(sortByCloseTime, f.cfg, f.logger)
}

// NewQueue returns a new queue backed by cassandra
func (f *Factory) NewQueue(queueType p.QueueType) (p.Queue, error) {
	return newNoSQLQueueStore(f.cfg, f.logger, queueType)
}

// NewConfigStore returns a new config store
func (f *Factory) NewConfigStore() (p.ConfigStore, error) {
	return NewNoSQLConfigStore(f.cfg, f.logger)
}

// Close closes the factory
func (f *Factory) Close() {
	f.Lock()
	defer f.Unlock()
	if f.execStoreFactory != nil {
		f.execStoreFactory.close()
	}
}

func (f *Factory) executionStoreFactory() (*executionStoreFactory, error) {
	f.RLock()
	if f.execStoreFactory != nil {
		f.RUnlock()
		return f.execStoreFactory, nil
	}
	f.RUnlock()
	f.Lock()
	defer f.Unlock()
	if f.execStoreFactory != nil {
		return f.execStoreFactory, nil
	}

	factory, err := newExecutionStoreFactory(f.cfg, f.logger)
	if err != nil {
		return nil, err
	}
	f.execStoreFactory = factory
	return f.execStoreFactory, nil
}

// newExecutionStoreFactory is used to create an instance of ExecutionStoreFactory implementation
func newExecutionStoreFactory(
	cfg config.Cassandra,
	logger log.Logger,
) (*executionStoreFactory, error) {

	db, err := NewNoSQLDB(&cfg, logger)
	if err != nil {
		return nil, err
	}

	return &executionStoreFactory{
		db:     db,
		logger: logger,
	}, nil
}

func (f *executionStoreFactory) close() {
	f.db.Close()
}

// new implements ExecutionStoreFactory interface
func (f *executionStoreFactory) new(shardID int) (p.ExecutionStore, error) {
	pmgr, err := NewExecutionStore(shardID, f.db, f.logger)
	if err != nil {
		return nil, err
	}
	return pmgr, nil
}
