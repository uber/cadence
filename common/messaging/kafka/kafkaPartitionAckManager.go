// Copyright (c) 2020 Uber Technologies, Inc.
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

package kafka

import (
	"sync"

	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/messaging"
)

// Used to convert out of order acks into ackLevel movement,
// assuming reading messages is in order and continuous(no skipping)
type kafkaPartitionAckManager struct {
	sync.RWMutex
	ackMgrs map[int32]messaging.AckManager //map from partition to its ackManager
	logger  log.Logger
}

func newPartitionAckManager(logger log.Logger) *kafkaPartitionAckManager {
	return &kafkaPartitionAckManager{
		logger:  logger,
		ackMgrs: make(map[int32]messaging.AckManager),
	}
}

// AddMessage mark a messageID as read and waiting for completion
func (pam *kafkaPartitionAckManager) AddMessage(partitionID int32, messageID int64) {
	var err error
	pam.RLock()
	if am, ok := pam.ackMgrs[partitionID]; ok {
		err = am.ReadItem(messageID)
		pam.RUnlock()
	} else {
		pam.RUnlock()
		pam.Lock()
		partitionLogger := pam.logger.WithTags(tag.KafkaPartition(partitionID))
		am := messaging.NewAckManager(partitionLogger)
		pam.ackMgrs[partitionID] = am
		err = am.ReadItem(messageID)
		pam.Unlock()
	}
	if err != nil {
		pam.logger.Warn("potential bug when adding message to ackManager", tag.Error(err), tag.KafkaPartition(partitionID), tag.TaskID(messageID))
	}
}

// CompleteMessage complete the message from ack/nack kafka message
func (pam *kafkaPartitionAckManager) CompleteMessage(partitionID int32, messageID int64) (ackLevel int64) {
	pam.RLock()
	defer pam.RUnlock()
	if am, ok := pam.ackMgrs[partitionID]; ok {
		ackLevel = am.AckItem(messageID)
	} else {
		pam.logger.Fatal("complete an message that hasn't been added",
			tag.KafkaPartition(partitionID),
			tag.KafkaOffset(messageID))
		ackLevel = -1
	}
	return ackLevel
}
