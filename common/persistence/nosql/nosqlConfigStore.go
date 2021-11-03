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
	"fmt"

	"github.com/uber/cadence/common/config"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/persistence"
	p "github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin"
)

type (
	nosqlConfigStore struct {
		nosqlStore
	}
)

func NewNoSQLConfigStore(
	cfg config.NoSQL,
	logger log.Logger,
) (p.ConfigStore, error) {
	db, err := NewNoSQLDB(&cfg, logger)
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

func (m *nosqlConfigStore) FetchConfig(ctx context.Context, configType p.ConfigType) (*p.InternalConfigStoreEntry, error) {
	entry, err := m.db.SelectLatestConfig(ctx, int(configType))
	if err != nil {
		return nil, convertCommonErrors(m.db, "FetchConfig", err)
	}
	return entry, nil
}

func (m *nosqlConfigStore) UpdateConfig(ctx context.Context, value *p.InternalConfigStoreEntry) error {
	err := m.db.InsertConfig(ctx, value)
	if err != nil {
		if _, ok := err.(*nosqlplugin.ConditionFailure); ok {
			return &persistence.ConditionFailedError{Msg: fmt.Sprintf("Version %v already exists. Condition Failed", value.Version)}
		}
		return convertCommonErrors(m.db, "UpdateConfig", err)
	}
	return nil
}
