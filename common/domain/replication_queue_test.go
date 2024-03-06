// Copyright (c) 2023 Uber Technologies, Inc.
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

package domain

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
)

func TestReplicationQueueImpl_Publish(t *testing.T) {
	tests := []struct {
		name      string
		task      *types.ReplicationTask
		wantErr   bool
		setupMock func(q *persistence.MockQueueManager)
	}{
		{
			name:    "successful publish",
			task:    &types.ReplicationTask{},
			wantErr: false,
			setupMock: func(q *persistence.MockQueueManager) {
				q.EXPECT().EnqueueMessage(gomock.Any(), gomock.Any()).Return(nil)
			},
		},
		{
			name:    "publish fails",
			task:    &types.ReplicationTask{},
			wantErr: true,
			setupMock: func(q *persistence.MockQueueManager) {
				q.EXPECT().EnqueueMessage(gomock.Any(), gomock.Any()).Return(errors.New("enqueue error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockQueue := persistence.NewMockQueueManager(ctrl)
			rq := NewReplicationQueue(mockQueue, "testCluster", nil, nil)
			tt.setupMock(mockQueue)
			err := rq.Publish(context.Background(), tt.task)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			ctrl.Finish()
		})
	}
}

func TestReplicationQueueImpl_PublishToDLQ(t *testing.T) {
	tests := []struct {
		name      string
		task      *types.ReplicationTask
		wantErr   bool
		setupMock func(q *persistence.MockQueueManager)
	}{
		{
			name:    "successful publish to DLQ",
			task:    &types.ReplicationTask{},
			wantErr: false,
			setupMock: func(q *persistence.MockQueueManager) {
				q.EXPECT().EnqueueMessageToDLQ(gomock.Any(), gomock.Any()).Return(nil)
			},
		},
		{
			name:    "publish to DLQ fails",
			task:    &types.ReplicationTask{},
			wantErr: true,
			setupMock: func(q *persistence.MockQueueManager) {
				q.EXPECT().EnqueueMessageToDLQ(gomock.Any(), gomock.Any()).Return(errors.New("enqueue to DLQ error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockQueue := persistence.NewMockQueueManager(ctrl)
			rq := NewReplicationQueue(mockQueue, "testCluster", nil, nil)
			tt.setupMock(mockQueue)
			err := rq.PublishToDLQ(context.Background(), tt.task)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			ctrl.Finish()
		})
	}
}

func TestGetReplicationMessages(t *testing.T) {

	tests := []struct {
		name      string
		lastID    int64
		maxCount  int
		task      *types.ReplicationTask
		wantErr   bool
		setupMock func(q *persistence.MockQueueManager)
	}{
		{
			name:     "successful message retrieval",
			lastID:   100,
			maxCount: 10,
			task:     &types.ReplicationTask{},
			wantErr:  false,
			setupMock: func(q *persistence.MockQueueManager) {
				q.EXPECT().ReadMessages(gomock.Any(), gomock.Eq(int64(100)), gomock.Eq(10)).Return(persistence.QueueMessageList{}, nil)
			},
		},
		{
			name:     "read messages fails",
			lastID:   100,
			maxCount: 10,
			wantErr:  true,
			setupMock: func(q *persistence.MockQueueManager) {
				q.EXPECT().ReadMessages(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("read error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockQueue := persistence.NewMockQueueManager(ctrl)
			rq := NewReplicationQueue(mockQueue, "testCluster", nil, nil)

			tt.setupMock(mockQueue)
			_, _, err := rq.GetReplicationMessages(context.Background(), tt.lastID, tt.maxCount)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			ctrl.Finish()
		})
	}
}

func TestUpdateAckLevel(t *testing.T) {
	tests := []struct {
		name      string
		lastID    int64
		cluster   string
		wantErr   bool
		setupMock func(q *persistence.MockQueueManager)
	}{
		{
			name:    "successful ack level update",
			lastID:  100,
			cluster: "testCluster",
			wantErr: false,
			setupMock: func(q *persistence.MockQueueManager) {
				q.EXPECT().UpdateAckLevel(gomock.Any(), gomock.Eq(int64(100)), gomock.Eq("testCluster")).Return(nil)
			},
		},
		{
			name:    "ack level update fails",
			lastID:  100,
			cluster: "testCluster",
			wantErr: true,
			setupMock: func(q *persistence.MockQueueManager) {
				q.EXPECT().UpdateAckLevel(gomock.Any(), gomock.Eq(int64(100)), gomock.Eq("testCluster")).Return(errors.New("update error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockQueue := persistence.NewMockQueueManager(ctrl)

			rq := NewReplicationQueue(mockQueue, "testCluster", nil, nil)
			tt.setupMock(mockQueue)
			err := rq.UpdateAckLevel(context.Background(), tt.lastID, tt.cluster)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			ctrl.Finish()
		})
	}
}
