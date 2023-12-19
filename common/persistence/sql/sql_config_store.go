// The MIT License (MIT)

// Copyright (c) 2017-2020 Uber Technologies Inc.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package sql

import (
	"context"
	"fmt"

	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/persistence/serialization"
	"github.com/uber/cadence/common/persistence/sql/sqlplugin"
)

type (
	sqlConfigStore struct {
		sqlStore
	}
)

// NewSQLConfigStore creates a config store for SQL
func NewSQLConfigStore(
	db sqlplugin.DB,
	logger log.Logger,
	parser serialization.Parser,
) (persistence.ConfigStore, error) {
	return &sqlConfigStore{
		sqlStore: sqlStore{
			db:     db,
			logger: logger,
			parser: parser,
		},
	}, nil
}

func (m *sqlConfigStore) FetchConfig(ctx context.Context, configType persistence.ConfigType) (*persistence.InternalConfigStoreEntry, error) {
	entry, err := m.db.SelectLatestConfig(ctx, int(configType))
	if m.db.IsNotFoundError(err) {
		return nil, nil
	}
	if err != nil {
		return nil, convertCommonErrors(m.db, "FetchConfig", "", err)
	}
	return entry, nil
}

func (m *sqlConfigStore) UpdateConfig(ctx context.Context, value *persistence.InternalConfigStoreEntry) error {
	err := m.db.InsertConfig(ctx, value)
	if err != nil {
		if m.db.IsDupEntryError(err) {
			return &persistence.ConditionFailedError{Msg: fmt.Sprintf("Version %v already exists. Condition Failed", value.Version)}
		}
		return convertCommonErrors(m.db, "UpdateConfig", "", err)
	}
	return nil
}
