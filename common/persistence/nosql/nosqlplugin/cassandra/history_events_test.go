package cassandra

import (
	"context"
	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin/cassandra/gocql"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin"
	"github.com/uber/cadence/common/types"
)

func TestInsertIntoHistoryTreeAndNode(t *testing.T) {
	tests := []struct {
		name        string
		treeRow     *nosqlplugin.HistoryTreeRow
		nodeRow     *nosqlplugin.HistoryNodeRow
		setupMocks  func(session *fakeSession)
		expectError bool
	}{
		{
			name: "Successfully insert tree and node row",
			treeRow: &nosqlplugin.HistoryTreeRow{
				TreeID:          "treeID",
				BranchID:        "branchID",
				Ancestors:       []*types.HistoryBranchRange{}, // Adjusted to slice of pointers
				CreateTimestamp: time.Now(),
			},
			nodeRow: &nosqlplugin.HistoryNodeRow{
				TreeID:       "treeID",
				BranchID:     "branchID",
				NodeID:       1,
				Data:         []byte("data"),
				DataEncoding: "encoding",
			},
			setupMocks: func(session *fakeSession) {
				session.mapExecuteBatchCASApplied = true
				session.mapExecuteBatchCASErr = nil
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			session := &fakeSession{
				query: gocql.NewMockQuery(ctrl),
			}
			if tt.setupMocks != nil {
				tt.setupMocks(session)
			}

			db := &cdb{session: session}

			err := db.InsertIntoHistoryTreeAndNode(context.Background(), tt.treeRow, tt.nodeRow)
			if tt.expectError {
				assert.Error(t, err, "Expected an error but got none")
			} else {
				assert.NoError(t, err, "Did not expect an error but got one")
			}
		})
	}
}

func TestSelectFromHistoryNode(t *testing.T) {
	txnID1 := int64(1)
	txnID2 := int64(2)
	tests := []struct {
		name          string
		filter        *nosqlplugin.HistoryNodeFilter
		setupMocks    func(*gomock.Controller, *fakeSession)
		expectedRows  []*nosqlplugin.HistoryNodeRow
		expectedToken []byte
		expectError   bool
	}{
		{
			name: "Successfully retrieve history nodes",
			filter: &nosqlplugin.HistoryNodeFilter{
				TreeID:        "treeID",
				BranchID:      "branchID",
				MinNodeID:     1,
				MaxNodeID:     10,
				PageSize:      5,
				NextPageToken: nil,
			},
			setupMocks: func(ctrl *gomock.Controller, session *fakeSession) {
				mockQuery := gocql.NewMockQuery(ctrl)
				mockQuery.EXPECT().WithContext(gomock.Any()).Return(mockQuery).AnyTimes()
				mockQuery.EXPECT().PageSize(gomock.Any()).Return(mockQuery).AnyTimes()
				mockQuery.EXPECT().PageState(gomock.Any()).Return(mockQuery).AnyTimes()
				mockQuery.EXPECT().Iter().Return(&fakeIter{
					scanInputs: [][]interface{}{
						{int64(1), &txnID1, []byte("data1"), "encoding"},
						{int64(2), &txnID2, []byte("data2"), "encoding"},
					},
					pageState: []byte("nextPageToken"),
				}).AnyTimes()

				session.query = mockQuery
			},
			expectedRows: []*nosqlplugin.HistoryNodeRow{
				{NodeID: int64(1), TxnID: &txnID1, Data: []byte("data1"), DataEncoding: "encoding"},
				{NodeID: int64(2), TxnID: &txnID2, Data: []byte("data2"), DataEncoding: "encoding"},
			},
			expectedToken: []byte("nextPageToken"),
			expectError:   false,
		},
		{
			name: "Failure to create query iterator",
			filter: &nosqlplugin.HistoryNodeFilter{
				TreeID:        "treeID",
				BranchID:      "branchID",
				MinNodeID:     1,
				MaxNodeID:     10,
				PageSize:      5,
				NextPageToken: nil,
			},
			setupMocks: func(ctrl *gomock.Controller, session *fakeSession) {
				mockQuery := gocql.NewMockQuery(ctrl)
				mockQuery.EXPECT().WithContext(gomock.Any()).Return(mockQuery).AnyTimes()
				mockQuery.EXPECT().PageSize(gomock.Any()).Return(mockQuery).AnyTimes()
				mockQuery.EXPECT().PageState(gomock.Any()).Return(mockQuery).AnyTimes()
				mockQuery.EXPECT().Iter().Return(nil).AnyTimes() // Simulating failure

				session.query = mockQuery
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			session := &fakeSession{}
			if tt.setupMocks != nil {
				tt.setupMocks(ctrl, session)
			}

			db := &cdb{session: session}
			rows, token, err := db.SelectFromHistoryNode(context.Background(), tt.filter)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRows, rows)
				assert.Equal(t, tt.expectedToken, token)
			}
		})
	}
}

func TestDeleteFromHistoryTreeAndNode(t *testing.T) {
	tests := []struct {
		name        string
		treeFilter  *nosqlplugin.HistoryTreeFilter
		nodeFilters []*nosqlplugin.HistoryNodeFilter
		setupMocks  func(*fakeSession)
		expectError bool
	}{
		{
			name: "Successfully delete tree and nodes",
			treeFilter: &nosqlplugin.HistoryTreeFilter{
				ShardID:  1,
				TreeID:   "treeID",
				BranchID: stringPtr("branchID"),
			},
			nodeFilters: []*nosqlplugin.HistoryNodeFilter{
				{TreeID: "treeID", BranchID: "branchID", MinNodeID: 1},
				{TreeID: "treeID", BranchID: "branchID", MinNodeID: 2},
			},
			setupMocks: func(session *fakeSession) {
				// Simulate successful batch execution
				session.mapExecuteBatchCASApplied = true
			},
			expectError: false,
		},
		{
			name: "Failure in batch execution",
			treeFilter: &nosqlplugin.HistoryTreeFilter{
				ShardID:  1,
				TreeID:   "treeID",
				BranchID: stringPtr("branchID"),
			},
			nodeFilters: []*nosqlplugin.HistoryNodeFilter{
				{TreeID: "treeID", BranchID: "branchID", MinNodeID: 1},
			},
			setupMocks: func(session *fakeSession) {
				// Simulate failure in batch execution
				session.mapExecuteBatchCASErr = types.InternalServiceError{Message: "DB operation failed"}
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			session := &fakeSession{}
			if tt.setupMocks != nil {
				tt.setupMocks(session)
			}

			db := &cdb{session: session}
			err := db.DeleteFromHistoryTreeAndNode(context.Background(), tt.treeFilter, tt.nodeFilters)

			if tt.expectError {
				assert.Error(t, err, "Expected an error but got none")
			} else {
				assert.NoError(t, err, "Did not expect an error but got one")
			}
		})
	}
}

func stringPtr(s string) *string {
	return &s
}
