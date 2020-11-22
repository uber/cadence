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

package messaging

import (
	"sync"

	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
)

// Used to convert out of order acks into ackLevel movement,
// assuming reading messages is in order and continuous(no skipping)
type partitionAckManager struct {
	sync.RWMutex
	ackMgrs map[int32]*ackManager //map from partition to its ackManager
	logger  log.Logger
}

func newPartitionAckManager(logger log.Logger) *partitionAckManager {
	return &partitionAckManager{
		logger:  logger,
		ackMgrs: make(map[int32]*ackManager),
	}
}

// AddMessage mark a messageID as read and waiting for completion
func (pam *partitionAckManager) AddMessage(partitionID int32, messageID int64) {
	pam.RLock()
	if am, ok := pam.ackMgrs[partitionID]; ok {
		am.addMessage(messageID)
		pam.RUnlock()
	} else {
		pam.RUnlock()
		pam.Lock()
		am := newAckManager(partitionID, pam.logger)
		pam.ackMgrs[partitionID] = am
		am.addMessage(messageID)
		pam.Unlock()
	}
}

// CompleteMessage complete the message
func (pam *partitionAckManager) CompleteMessage(partitionID int32, messageID int64) (ackLevel int64) {
	pam.RLock()
	defer pam.RUnlock()
	if am, ok := pam.ackMgrs[partitionID]; ok {
		ackLevel = am.completeMessage(messageID)
	} else {
		pam.logger.Fatal("complete an message that hasn't been added",
			tag.KafkaPartition(partitionID),
			tag.TaskID(messageID))
		ackLevel = -1
	}
	return ackLevel
}

type ackManager struct {
	sync.RWMutex
	partitionID         int32
	outstandingMessages map[int64]bool // key->MessageID, value->(true for acked/completed, false->for non acked)
	readLevel           int64          // Maximum MessageID inserted into outstandingMessages
	ackLevel            int64          // Maximum MessageID below which all messages are acked
	logger              log.Logger
}

func newAckManager(partitionID int32, logger log.Logger) *ackManager {
	return &ackManager{
		partitionID:         partitionID,
		logger:              logger,
		outstandingMessages: make(map[int64]bool),
		readLevel:           -1,
		ackLevel:            -1,
	}
}

// Registers message as in-flight and moves read level to it. Messages can be added in increasing order of messageID only.
// NOTE that ackManager assumes adding messages is in order
func (m *ackManager) addMessage(messageID int64) {
	m.Lock()
	defer m.Unlock()
	if m.readLevel >= messageID {
		m.logger.Error("Next message ID is less than or equal to current read level. This should not happen",
			tag.TaskID(messageID),
			tag.ReadLevel(m.readLevel),
			tag.KafkaPartition(m.partitionID))
		return
	}
	if _, ok := m.outstandingMessages[messageID]; ok {
		m.logger.Error("Already present in outstanding messages but hasn't added. This should not happen",
			tag.TaskID(messageID),
			tag.KafkaPartition(m.partitionID))
		return
	}
	m.readLevel = messageID
	if m.ackLevel == -1 {
		// because of ordering, the first messageID is the minimum to ack
		m.ackLevel = messageID - 1
		m.logger.Info("add first messageID in a session:",
			tag.TaskID(messageID),
			tag.KafkaPartition(m.partitionID),
		)
	}
	m.outstandingMessages[messageID] = false // true is for acked
}

func (m *ackManager) completeMessage(messageID int64) (ackLevel int64) {
	m.Lock()
	defer m.Unlock()
	if completed, ok := m.outstandingMessages[messageID]; ok && !completed {
		m.outstandingMessages[messageID] = true
	} else {
		m.logger.Warn("Duplicated completion for message",
			tag.KafkaPartition(m.partitionID),
			tag.TaskID(messageID))
	}

	// Update ackLevel
	for current := m.ackLevel + 1; current <= m.readLevel; current++ {
		if acked, ok := m.outstandingMessages[current]; ok {
			if acked {
				m.ackLevel = current
				delete(m.outstandingMessages, current)
			} else {
				return m.ackLevel
			}
		} else {
			m.logger.Error("A message is probably skipped when adding message. This should not happen",
				tag.KafkaPartition(m.partitionID),
				tag.TaskID(current))
		}
	}
	return m.ackLevel
}
