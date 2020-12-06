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
	"fmt"
	"sync"

	"go.uber.org/atomic"

	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
)

type ackManager struct {
	sync.RWMutex
	outstandingMessages map[int64]bool // key->itemID, value->(true for acked/completed, false->for non acked)
	readLevel           int64          // Maximum itemID inserted into outstandingMessages
	ackLevel            int64          // Maximum itemID below which all messages are acked
	backlogCounter      atomic.Int64
	logIncontinuousErr  bool // emit error for itemID being incontinuous when consuming for potential bugs
	logger              log.Logger
}

// NewAckManager returns a AckManager without monitoring the itemIDs continousness.
// For example, our internal matching task queue doesn't guarantee it.
func NewAckManager(logger log.Logger) AckManager {
	return newAckManager(false, logger)
}

// NewContinuousAckManager returns a ContinuousAckManager
// it will emit error logs for itemIDs being incontinuous
// This is useful for some message queue system that guarantees continuousness
// that we want to monitor it's behaving correctly
func NewContinuousAckManager(logger log.Logger) AckManager {
	return newAckManager(true, logger)
}

func newAckManager(logIncontinuousErr bool, logger log.Logger) AckManager {
	return &ackManager{
		logger:              logger,
		outstandingMessages: make(map[int64]bool),
		readLevel:           -1,
		ackLevel:            -1,
		logIncontinuousErr:  logIncontinuousErr,
	}
}

// Registers message as in-flight and moves read level to it. Messages can be added in increasing order of messageID only.
// NOTE that ackManager assumes adding messages is in order
func (m *ackManager) ReadItem(itemID int64) error {
	m.Lock()
	defer m.Unlock()
	m.backlogCounter.Inc()
	if m.readLevel >= itemID {
		return fmt.Errorf("next item ID is less than or equal to current read level. itemID %d, readLevel %d", itemID, m.readLevel)
	}
	if _, ok := m.outstandingMessages[itemID]; ok {
		return fmt.Errorf("already present in outstanding items but hasn't added itemID:%d", itemID)
	}
	m.readLevel = itemID
	if m.ackLevel == -1 {
		// because of ordering, the first itemID is the minimum to ack
		m.ackLevel = itemID - 1
		m.logger.Info("this is the very first itemID being read in this ackManager",
			tag.TaskID(itemID),
		)
	}
	m.outstandingMessages[itemID] = false // true is for acked
	return nil
}

func (m *ackManager) AckItem(itemID int64) (ackLevel int64) {
	m.Lock()
	defer m.Unlock()
	if completed, ok := m.outstandingMessages[itemID]; ok && !completed {
		m.outstandingMessages[itemID] = true
		m.backlogCounter.Dec()
	} else {
		m.logger.Warn("Duplicated completion for item",
			tag.TaskID(itemID))
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
			if m.logIncontinuousErr {
				m.logger.Error("potential bug, an item is probably skipped when adding", tag.TaskID(current))
			}
		}
	}
	return m.ackLevel
}

func (m *ackManager) GetReadLevel() int64 {
	m.RLock()
	defer m.RUnlock()
	return m.readLevel
}

func (m *ackManager) SetReadLevel(readLevel int64) {
	m.Lock()
	defer m.Unlock()
	m.readLevel = readLevel
}

func (m *ackManager) GetAckLevel() int64 {
	m.RLock()
	defer m.RUnlock()
	return m.ackLevel
}

func (m *ackManager) SetAckLevel(ackLevel int64) {
	m.Lock()
	defer m.Unlock()
	if ackLevel > m.ackLevel {
		m.ackLevel = ackLevel
	}
	if ackLevel > m.readLevel {
		m.readLevel = ackLevel
	}
}

func (m *ackManager) GetBacklogCount() int64 {
	return m.backlogCounter.Load()
}
