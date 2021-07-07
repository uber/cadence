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
	"context"

	"github.com/uber/cadence/common/config"
	"github.com/uber/cadence/common/log"
	p "github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin/cassandra"
)

type (
	nosqlConfigStore struct {
		nosqlStore
	}
)

func newNoSQLConfigStore(
	cfg config.NoSQL,
	logger log.Logger,
) (p.ConfigStore, error) {
	// TODO hardcoding to Cassandra for now, will switch to dynamically loading later
	db, err := cassandra.NewCassandraDB(cfg, logger)
	if err != nil {
		return nil, err
	}

	return &nosqlConfigStore{
		nosqlStore: nosqlStore{
			db:     db,
			logger: logger,
		},
	}, nil
}

func (m *nosqlConfigStore) FetchConfig(ctx context.Context, config_type string) (*p.InternalConfigStoreEntry, error) {
	entry, err := m.db.SelectLatestConfig(ctx, config_type)
	if err != nil {
		return nil, convertCommonErrors(m.db, "FetchConfig", err)
	}
	return entry, nil
}

func (m *nosqlConfigStore) UpdateConfig(ctx context.Context, value *p.InternalConfigStoreEntry) error {
	err := m.db.InsertConfig(ctx, value)
	return convertCommonErrors(m.db, "UpdateConfig", err)
}
