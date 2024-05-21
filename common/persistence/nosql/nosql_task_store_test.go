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

package nosql

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/types"
)

func TestNewNoSQLStore(t *testing.T) {
	registerCassandraMock(t)
	cfg := getValidShardedNoSQLConfig()

	store, err := newNoSQLTaskStore(cfg, log.NewNoop(), nil)
	assert.NoError(t, err)
	assert.NotNil(t, store)
}

func setupNoSQLStoreMocks(t *testing.T) *nosqlTaskStore {
	ctrl := gomock.NewController(t)
	shardedNosqlStoreMock := NewMockshardedNosqlStore(ctrl)

	store := &nosqlTaskStore{
		shardedNosqlStore: shardedNosqlStoreMock,
	}

	return store
}

func TestGetOrphanTasks(t *testing.T) {
	store := setupNoSQLStoreMocks(t)

	// We just expect the function to return an error so we don't need to check the result
	_, err := store.GetOrphanTasks(ctx.Background(), nil)

	var expectedErr *types.InternalServiceError
	assert.ErrorAs(t, err, &expectedErr)
	assert.ErrorContains(t, err, "Unimplemented call to GetOrphanTasks for NoSQL")
}
