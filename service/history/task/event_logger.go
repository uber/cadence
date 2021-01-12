// Copyright (c) 2017-2021 Uber Technologies, Inc.
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

package task

import (
	"fmt"
	"strings"
	"time"

	"github.com/uber/cadence/common/clock"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
)

type (
	eventLogger interface {
		AddEvent(eventMsg string, details ...interface{})
		FlushEvents(msg string) int
	}

	event struct {
		timestamp time.Time
		message   string
		details   []interface{}
	}

	eventLoggerImpl struct {
		logger     log.Logger
		timeSource clock.TimeSource

		events       []event
		nextEventIdx int
		maxSize      int
	}
)

func newEventLogger(
	logger log.Logger,
	timeSource clock.TimeSource,
	maxSize int,
) eventLogger {
	return &eventLoggerImpl{
		logger:     logger,
		timeSource: timeSource,

		events:       make([]event, maxSize),
		nextEventIdx: 0,
		maxSize:      maxSize,
	}
}

func (e *eventLoggerImpl) AddEvent(
	eventMsg string,
	details ...interface{},
) {
	e.events[e.nextEventIdx] = event{
		timestamp: e.timeSource.Now(),
		message:   eventMsg,
		details:   details,
	}

	e.nextEventIdx = (e.nextEventIdx + 1) % e.maxSize
}

func (e *eventLoggerImpl) FlushEvents(
	msg string,
) int {
	var builder strings.Builder
	eventIdx := e.nextEventIdx
	eventsFlushed := 0

	for i := 0; i != len(e.events); i++ {
		// dump all events in reverse order
		eventIdx--
		if eventIdx < 0 {
			eventIdx = len(e.events) - 1
		}

		event := e.events[eventIdx]
		if event.timestamp.IsZero() {
			break
		}

		eventsFlushed++
		builder.WriteString(fmt.Sprintf("%v, %s, %v\n", event.timestamp, event.message, event.details))
	}

	e.logger.Info(msg, tag.Key("events"), tag.Value(builder.String()))

	e.events = make([]event, e.maxSize)

	return eventsFlushed
}
