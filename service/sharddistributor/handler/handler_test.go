package handler

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/uber/cadence/client/history"
	"github.com/uber/cadence/client/matching"
	"github.com/uber/cadence/common/log/testlogger"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/sharddistributor/constants"
)

func TestGetShardOwner(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHistoryPeerResolver := history.NewMockPeerResolver(ctrl)
	mockMatchingPeerResolver := matching.NewMockPeerResolver(ctrl)
	logger := testlogger.New(t)

	handler := &handlerImpl{
		logger:               logger,
		historyPeerResolver:  mockHistoryPeerResolver,
		matchingPeerResolver: mockMatchingPeerResolver,
	}

	tests := []struct {
		name           string
		request        *types.GetShardOwnerRequest
		setupMocks     func()
		expectedOwner  string
		expectedError  bool
		expectedErrMsg string
	}{
		{
			name: "HistoryNamespace_Success",
			request: &types.GetShardOwnerRequest{
				Namespace: constants.HistoryNamespace,
				ShardKey:  "123",
			},
			setupMocks: func() {
				mockHistoryPeerResolver.EXPECT().FromShardID(123).Return("owner1", nil)
			},
			expectedOwner: "owner1",
			expectedError: false,
		},
		{
			name: "MatchingNamespace_Success",
			request: &types.GetShardOwnerRequest{
				Namespace: constants.MatchingNamespace,
				ShardKey:  "taskList1",
			},
			setupMocks: func() {
				mockMatchingPeerResolver.EXPECT().FromTaskList("taskList1").Return("owner2", nil)
			},
			expectedOwner: "owner2",
			expectedError: false,
		},
		{
			name: "InvalidNamespace",
			request: &types.GetShardOwnerRequest{
				Namespace: "namespace not found invalidNamespace",
				ShardKey:  "1",
			},
			setupMocks:     func() {},
			expectedError:  true,
			expectedErrMsg: "namespace not found",
		},
		{
			name: "HistoryNamespace_InvalidShardKey",
			request: &types.GetShardOwnerRequest{
				Namespace: constants.HistoryNamespace,
				ShardKey:  "invalidShardKey",
			},
			setupMocks:     func() {},
			expectedError:  true,
			expectedErrMsg: "invalid syntax",
		},
		{
			name: "HistoryNamespace_LookupError",
			request: &types.GetShardOwnerRequest{
				Namespace: constants.HistoryNamespace,
				ShardKey:  "1",
			},
			setupMocks: func() {
				mockHistoryPeerResolver.EXPECT().FromShardID(1).Return("", errors.New("lookup error"))
			},
			expectedError:  true,
			expectedErrMsg: "lookup error",
		},
		{
			name: "MatchingNamespace_LookupError",
			request: &types.GetShardOwnerRequest{
				Namespace: constants.MatchingNamespace,
				ShardKey:  "taskList1",
			},
			setupMocks: func() {
				mockMatchingPeerResolver.EXPECT().FromTaskList("taskList1").Return("", errors.New("lookup error"))
			},
			expectedError:  true,
			expectedErrMsg: "lookup error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			resp, err := handler.GetShardOwner(context.Background(), tt.request)
			if tt.expectedError {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedErrMsg)
				require.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedOwner, resp.Owner)
				require.Equal(t, tt.request.Namespace, resp.Namespace)
			}
		})
	}
}
