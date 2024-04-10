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
	ctx "context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin"
	"github.com/uber/cadence/common/types"
)

const (
	testTransactionID = 123
	testShardID       = 456
	testNodeID        = 8
)

func validInternalAppendHistoryNodesRequest() *persistence.InternalAppendHistoryNodesRequest {
	return &persistence.InternalAppendHistoryNodesRequest{
		IsNewBranch: false,
		Info:        "TestInfo",
		BranchInfo: types.HistoryBranch{
			TreeID:   "TestTreeID",
			BranchID: "TestBranchID",
			Ancestors: []*types.HistoryBranchRange{
				{
					BranchID:    "TestAncestorBranchID",
					BeginNodeID: 0,
					EndNodeID:   5,
				},
			},
		},
		NodeID: testNodeID,
		Events: &persistence.DataBlob{
			Encoding: common.EncodingTypeThriftRW,
			Data:     []byte("TestEvents"),
		},
		TransactionID: testTransactionID,
		ShardID:       testShardID,
	}
}

func validHistoryNodeRow() *nosqlplugin.HistoryNodeRow {
	expectedNodeRow := &nosqlplugin.HistoryNodeRow{
		TreeID:       "TestTreeID",
		BranchID:     "TestBranchID",
		NodeID:       testNodeID,
		TxnID:        common.Ptr[int64](123),
		Data:         []byte("TestEvents"),
		DataEncoding: string(common.EncodingTypeThriftRW),
		ShardID:      testShardID,
	}
	return expectedNodeRow
}

func registerCassandraMock(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockDB := nosqlplugin.NewMockDB(ctrl)

	mockPlugin := nosqlplugin.NewMockPlugin(ctrl)
	mockPlugin.EXPECT().CreateDB(gomock.Any(), gomock.Any(), gomock.Any()).Return(mockDB, nil).AnyTimes()
	RegisterPlugin("cassandra", mockPlugin)
}

func TestNewNoSQLHistoryStore(t *testing.T) {
	registerCassandraMock(t)
	cfg := getValidShardedNoSQLConfig()

	store, err := newNoSQLHistoryStore(cfg, log.NewNoop(), nil)
	assert.NoError(t, err)
	assert.NotNil(t, store)
}

func setUpMocks(t *testing.T) (*nosqlHistoryStore, *nosqlplugin.MockDB) {
	ctrl := gomock.NewController(t)
	dbMock := nosqlplugin.NewMockDB(ctrl)

	nosqlSt := nosqlStore{
		logger: log.NewNoop(),
		db:     dbMock,
	}

	shardedNosqlStoreMock := NewMockshardedNosqlStore(ctrl)
	shardedNosqlStoreMock.EXPECT().GetStoreShardByHistoryShard(testShardID).Return(&nosqlSt, nil).AnyTimes()

	store := &nosqlHistoryStore{
		shardedNosqlStore: shardedNosqlStoreMock,
	}

	return store, dbMock
}

func TestAppendHistoryNodes_ErrorIfAppendAbove(t *testing.T) {
	store, _ := setUpMocks(t)

	request := validInternalAppendHistoryNodesRequest()

	// If the nodeID to append is smaller than the last ancestor's end node ID, return an error
	request.NodeID = 3
	ans := request.BranchInfo.Ancestors
	ans[len(ans)-1].EndNodeID = 5

	err := store.AppendHistoryNodes(ctx.Background(), request)

	var invalidErr *persistence.InvalidPersistenceRequestError
	assert.ErrorAs(t, err, &invalidErr)
	assert.ErrorContains(t, err, "cannot append to ancestors' nodes")
}

func TestAppendHistoryNodes_NotNewBranch(t *testing.T) {
	store, dbMock := setUpMocks(t)

	// Expect to insert the node into the history tree and node, as this is not a new branch, expect treeRow to be nil
	dbMock.EXPECT().InsertIntoHistoryTreeAndNode(gomock.Any(), nil, validHistoryNodeRow()).Return(nil).Times(1)

	request := validInternalAppendHistoryNodesRequest()
	err := store.AppendHistoryNodes(ctx.Background(), request)

	assert.NoError(t, err)
}

func TestAppendHistoryNodes_NewBranch(t *testing.T) {
	request := validInternalAppendHistoryNodesRequest()
	request.IsNewBranch = true

	store, dbMock := setUpMocks(t)

	// Expect to insert the node into the history tree and node, as this is a new branch expect treeRow to be set
	dbMock.EXPECT().InsertIntoHistoryTreeAndNode(gomock.Any(), gomock.Any(), validHistoryNodeRow()).
		DoAndReturn(func(ctx ctx.Context, treeRow *nosqlplugin.HistoryTreeRow, nodeRow *nosqlplugin.HistoryNodeRow) error {
			// Assert that the treeRow is as expected, we have to check this here because the treeRow has time.Now() in it
			assert.Equal(t, testShardID, treeRow.ShardID)
			assert.Equal(t, "TestTreeID", treeRow.TreeID)
			assert.Equal(t, "TestBranchID", treeRow.BranchID)
			assert.Equal(t, request.BranchInfo.Ancestors, treeRow.Ancestors)
			assert.Equal(t, request.Info, treeRow.Info)

			assert.WithinDuration(t, time.Now(), treeRow.CreateTimestamp, time.Second)
			return nil
		})

	err := store.AppendHistoryNodes(ctx.Background(), request)

	assert.NoError(t, err)
}
