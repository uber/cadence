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

package queue

import (
	"fmt"

	"github.com/uber/cadence/service/history/task"
)

type processingQueueStateImpl struct {
	level        int
	ackLevel     task.Key
	readLevel    task.Key
	maxLevel     task.Key
	domainFilter DomainFilter
}

// NewProcessingQueueState creates a new state instance for processing queue
// readLevel will be set to the same value as ackLevel
func NewProcessingQueueState(
	level int,
	ackLevel task.Key,
	maxLevel task.Key,
	domainFilter DomainFilter,
) ProcessingQueueState {
	return newProcessingQueueState(
		level,
		ackLevel,
		ackLevel,
		maxLevel,
		domainFilter,
	)
}

func newProcessingQueueState(
	level int,
	ackLevel task.Key,
	readLevel task.Key,
	maxLevel task.Key,
	domainFilter DomainFilter,
) *processingQueueStateImpl {
	return &processingQueueStateImpl{
		level:        level,
		ackLevel:     ackLevel,
		readLevel:    readLevel,
		maxLevel:     maxLevel,
		domainFilter: domainFilter,
	}
}

func (s *processingQueueStateImpl) Level() int {
	return s.level
}

func (s *processingQueueStateImpl) MaxLevel() task.Key {
	return s.maxLevel
}

func (s *processingQueueStateImpl) AckLevel() task.Key {
	return s.ackLevel
}

func (s *processingQueueStateImpl) ReadLevel() task.Key {
	return s.readLevel
}

func (s *processingQueueStateImpl) DomainFilter() DomainFilter {
	return s.domainFilter
}

func (s *processingQueueStateImpl) String() string {
	return fmt.Sprintf("&{level: %+v, ackLevel: %+v, readLevel: %+v, maxLevel: %+v, domainFilter: %+v}",
		s.level, s.ackLevel, s.readLevel, s.maxLevel, s.domainFilter,
	)
}

func copyQueueState(state ProcessingQueueState) *processingQueueStateImpl {
	return newProcessingQueueState(
		state.Level(),
		state.AckLevel(),
		state.ReadLevel(),
		state.MaxLevel(),
		state.DomainFilter(),
	)
}
