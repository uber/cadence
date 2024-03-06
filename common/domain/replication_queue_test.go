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
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
)

const (
	preambleVersion0 byte = 0x59
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
		})
	}
}

func TestGetReplicationMessages(t *testing.T) {

	tests := []struct {
		name      string
		lastID    int64
		maxCount  int
		wantErr   bool
		setupMock func(q *persistence.MockQueueManager)
	}{
		{
			name:     "successful message retrieval",
			lastID:   100,
			maxCount: 10,
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
		})
	}
}

func TestReplicationQueueImpl_GetAckLevels(t *testing.T) {
	tests := []struct {
		name      string
		want      map[string]int64
		wantErr   bool
		setupMock func(q *persistence.MockQueueManager)
	}{
		{
			name:    "successful ack levels retrieval",
			want:    map[string]int64{"testCluster": 100},
			wantErr: false,
			setupMock: func(q *persistence.MockQueueManager) {
				q.EXPECT().GetAckLevels(gomock.Any()).Return(map[string]int64{"testCluster": 100}, nil)
			},
		},
		{
			name:    "ack levels retrieval fails",
			wantErr: true,
			setupMock: func(q *persistence.MockQueueManager) {
				q.EXPECT().GetAckLevels(gomock.Any()).Return(nil, errors.New("retrieval error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockQueue := persistence.NewMockQueueManager(ctrl)
			rq := NewReplicationQueue(mockQueue, "testCluster", nil, nil)
			tt.setupMock(mockQueue)
			got, err := rq.GetAckLevels(context.Background())
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func mockEncodeReplicationTask(sourceTaskID int64) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte(preambleVersion0)
	binary.Write(&buf, binary.BigEndian, sourceTaskID)
	return buf.Bytes(), nil
}

func TestGetMessagesFromDLQ(t *testing.T) {
	tests := []struct {
		name      string
		firstID   int64
		lastID    int64
		pageSize  int
		pageToken []byte
		taskID    int64
		wantErr   bool
	}{
		{
			name:      "successful message retrieval",
			firstID:   100,
			lastID:    200,
			pageSize:  10,
			pageToken: []byte("token"),
			taskID:    12345,
			wantErr:   false,
		},
		{
			name:      "read messages fails",
			firstID:   100,
			lastID:    200,
			pageSize:  10,
			pageToken: []byte("token"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockQueue := persistence.NewMockQueueManager(ctrl)
			rq := NewReplicationQueue(mockQueue, "testCluster", nil, nil)

			if !tt.wantErr {
				encodedData, _ := mockEncodeReplicationTask(tt.taskID)
				messages := []*persistence.QueueMessage{
					{ID: 1, Payload: encodedData},
				}
				mockQueue.EXPECT().ReadMessagesFromDLQ(gomock.Any(), tt.firstID, tt.lastID, tt.pageSize, tt.pageToken).Return(messages, []byte("nextToken"), nil)
			} else {
				mockQueue.EXPECT().ReadMessagesFromDLQ(gomock.Any(), tt.firstID, tt.lastID, tt.pageSize, tt.pageToken).Return(nil, nil, errors.New("read error"))
			}

			replicationTasks, token, err := rq.GetMessagesFromDLQ(context.Background(), tt.firstID, tt.lastID, tt.pageSize, tt.pageToken)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, replicationTasks, 1, "Expected one replication task to be returned")
				assert.Equal(t, []byte("nextToken"), token, "Expected token to match 'nextToken'")
			}
		})
	}
}

func TestUpdateDLQAckLevel(t *testing.T) {
	tests := []struct {
		name      string
		lastID    int64
		wantErr   bool
		setupMock func(q *persistence.MockQueueManager)
	}{
		{
			name:    "successful DLQ ack level update",
			lastID:  100,
			wantErr: false,
			setupMock: func(q *persistence.MockQueueManager) {
				q.EXPECT().UpdateDLQAckLevel(gomock.Any(), gomock.Eq(int64(100)), gomock.Eq("domainReplication")).Return(nil)
			},
		},
		{
			name:    "DLQ ack level update fails",
			lastID:  100,
			wantErr: true,
			setupMock: func(q *persistence.MockQueueManager) {
				q.EXPECT().UpdateDLQAckLevel(gomock.Any(), gomock.Eq(int64(100)), gomock.Eq("domainReplication")).Return(errors.New("update DLQ ack level error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockQueue := persistence.NewMockQueueManager(ctrl)
			rq := NewReplicationQueue(mockQueue, "testCluster", nil, nil)
			tt.setupMock(mockQueue)
			err := rq.UpdateDLQAckLevel(context.Background(), tt.lastID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetDLQAckLevel(t *testing.T) {
	tests := []struct {
		name      string
		want      int64
		wantErr   bool
		setupMock func(q *persistence.MockQueueManager)
	}{
		{
			name:    "successful DLQ ack level retrieval",
			want:    100,
			wantErr: false,
			setupMock: func(q *persistence.MockQueueManager) {
				q.EXPECT().GetDLQAckLevels(gomock.Any()).Return(map[string]int64{"domainReplication": 100}, nil)
			},
		},
		{
			name:    "DLQ ack level retrieval fails",
			wantErr: true,
			setupMock: func(q *persistence.MockQueueManager) {
				q.EXPECT().GetDLQAckLevels(gomock.Any()).Return(nil, errors.New("get DLQ ack level error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockQueue := persistence.NewMockQueueManager(ctrl)
			rq := NewReplicationQueue(mockQueue, "testCluster", nil, nil)
			tt.setupMock(mockQueue)
			got, err := rq.GetDLQAckLevel(context.Background())
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestRangeDeleteMessagesFromDLQ(t *testing.T) {
	tests := []struct {
		name      string
		firstID   int64
		lastID    int64
		wantErr   bool
		setupMock func(q *persistence.MockQueueManager)
	}{
		{
			name:    "successful range delete from DLQ",
			firstID: 10,
			lastID:  20,
			wantErr: false,
			setupMock: func(q *persistence.MockQueueManager) {
				q.EXPECT().RangeDeleteMessagesFromDLQ(gomock.Any(), gomock.Eq(int64(10)), gomock.Eq(int64(20))).Return(nil)
			},
		},
		{
			name:    "range delete from DLQ fails",
			firstID: 10,
			lastID:  20,
			wantErr: true,
			setupMock: func(q *persistence.MockQueueManager) {
				q.EXPECT().RangeDeleteMessagesFromDLQ(gomock.Any(), gomock.Eq(int64(10)), gomock.Eq(int64(20))).Return(errors.New("range delete error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockQueue := persistence.NewMockQueueManager(ctrl)
			rq := NewReplicationQueue(mockQueue, "testCluster", nil, nil)
			tt.setupMock(mockQueue)
			err := rq.RangeDeleteMessagesFromDLQ(context.Background(), tt.firstID, tt.lastID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDeleteMessageFromDLQ(t *testing.T) {
	tests := []struct {
		name      string
		messageID int64
		wantErr   bool
		setupMock func(q *persistence.MockQueueManager)
	}{
		{
			name:      "successful delete from DLQ",
			messageID: 15,
			wantErr:   false,
			setupMock: func(q *persistence.MockQueueManager) {
				q.EXPECT().DeleteMessageFromDLQ(gomock.Any(), gomock.Eq(int64(15))).Return(nil)
			},
		},
		{
			name:      "delete from DLQ fails",
			messageID: 15,
			wantErr:   true,
			setupMock: func(q *persistence.MockQueueManager) {
				q.EXPECT().DeleteMessageFromDLQ(gomock.Any(), gomock.Eq(int64(15))).Return(errors.New("delete error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockQueue := persistence.NewMockQueueManager(ctrl)
			rq := NewReplicationQueue(mockQueue, "testCluster", nil, nil)
			tt.setupMock(mockQueue)
			err := rq.DeleteMessageFromDLQ(context.Background(), tt.messageID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetDLQSize(t *testing.T) {
	tests := []struct {
		name      string
		wantSize  int64
		wantErr   bool
		setupMock func(m *persistence.MockQueueManager)
	}{
		{
			name:     "returns correct size for non-empty DLQ",
			wantSize: 10,
			wantErr:  false,
			setupMock: func(m *persistence.MockQueueManager) {
				m.EXPECT().GetDLQSize(gomock.Any()).Return(int64(10), nil)
			},
		},
		{
			name:     "returns zero for empty DLQ",
			wantSize: 0,
			wantErr:  false,
			setupMock: func(m *persistence.MockQueueManager) {
				m.EXPECT().GetDLQSize(gomock.Any()).Return(int64(0), nil)
			},
		},
		{
			name:    "propagates error from underlying queue",
			wantErr: true,
			setupMock: func(m *persistence.MockQueueManager) {
				m.EXPECT().GetDLQSize(gomock.Any()).Return(int64(0), errors.New("database error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockQueueManager := persistence.NewMockQueueManager(ctrl)
			tt.setupMock(mockQueueManager)
			q := &replicationQueueImpl{queue: mockQueueManager}
			size, err := q.GetDLQSize(context.Background())
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantSize, size)
			}
		})
	}
}