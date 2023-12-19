package serialization

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestShard_empty_struct(t *testing.T) {
	var s *ShardInfo
	assert.Equal(t, int32(0), s.GetStolenSinceRenew())
	assert.Equal(t, time.Unix(0, 0), s.GetUpdatedAt())
	assert.Equal(t, int64(0), s.GetReplicationAckLevel())
	assert.Equal(t, int64(0), s.GetTransferAckLevel())
	assert.Equal(t, time.Unix(0, 0), s.GetTimerAckLevel())
	assert.Equal(t, int64(0), s.GetDomainNotificationVersion())
	assert.Equal(t, emptyMap[string, int64](), s.GetClusterTransferAckLevel())
	assert.Equal(t, emptyMap[string, time.Time](), s.GetClusterTimerAckLevel())
	assert.Equal(t, "", s.GetOwner())
	assert.Equal(t, emptyMap[string, int64](), s.GetClusterReplicationLevel())
	assert.Equal(t, emptySlice[uint8](), s.GetPendingFailoverMarkers())
	assert.Equal(t, "", s.GetPendingFailoverMarkersEncoding())
	assert.Equal(t, emptyMap[string, int64](), s.GetReplicationDlqAckLevel())
	assert.Equal(t, emptySlice[uint8](), s.GetTransferProcessingQueueStates())
	assert.Equal(t, "", s.GetTransferProcessingQueueStatesEncoding())
	assert.Equal(t, emptySlice[uint8](), s.GetCrossClusterProcessingQueueStates())
	assert.Equal(t, "", s.GetCrossClusterProcessingQueueStatesEncoding())
	assert.Equal(t, emptySlice[uint8](), s.GetTimerProcessingQueueStates())
	assert.Equal(t, "", s.GetTimerProcessingQueueStatesEncoding())
}

func TestShard_non_empty(t *testing.T) {
	now := time.Now()
	s := ShardInfo{
		StolenSinceRenew:                          1,
		UpdatedAt:                                 now,
		ReplicationAckLevel:                       3,
		TransferAckLevel:                          4,
		TimerAckLevel:                             now.Add(time.Second),
		DomainNotificationVersion:                 6,
		ClusterTransferAckLevel:                   map[string]int64{"key1": 1, "key2": 2},
		ClusterTimerAckLevel:                      map[string]time.Time{"key1": now, "key2": now},
		Owner:                                     "test_owner",
		ClusterReplicationLevel:                   map[string]int64{"key1": 1, "key2": 2},
		PendingFailoverMarkers:                    []byte{1, 2, 3},
		PendingFailoverMarkersEncoding:            "test_encoding",
		ReplicationDlqAckLevel:                    map[string]int64{"key1": 1, "key2": 2},
		TransferProcessingQueueStates:             []byte{1, 2, 3},
		TransferProcessingQueueStatesEncoding:     "test_encoding",
		CrossClusterProcessingQueueStates:         []byte{1, 2, 3},
		CrossClusterProcessingQueueStatesEncoding: "test_encoding",
		TimerProcessingQueueStates:                []byte{1, 2, 3},
		TimerProcessingQueueStatesEncoding:        "test_encoding",
	}
	assert.Equal(t, int32(1), s.GetStolenSinceRenew())
	assert.Equal(t, now, s.GetUpdatedAt())
	assert.Equal(t, int64(3), s.GetReplicationAckLevel())
	assert.Equal(t, int64(4), s.GetTransferAckLevel())
	assert.Equal(t, now.Add(time.Second), s.GetTimerAckLevel())
	assert.Equal(t, int64(6), s.GetDomainNotificationVersion())
	assert.Equal(t, map[string]int64{"key1": 1, "key2": 2}, s.GetClusterTransferAckLevel())
	assert.Equal(t, map[string]time.Time{"key1": now, "key2": now}, s.GetClusterTimerAckLevel())
	assert.Equal(t, "test_owner", s.GetOwner())
	assert.Equal(t, map[string]int64{"key1": 1, "key2": 2}, s.GetClusterReplicationLevel())
	assert.Equal(t, []byte{1, 2, 3}, s.GetPendingFailoverMarkers())
	assert.Equal(t, "test_encoding", s.GetPendingFailoverMarkersEncoding())
}
