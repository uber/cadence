// Copyright (c) 2017-2020 Uber Technologies, Inc.
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
		AddEvent(eventMsg string, detail interface{})
		DumpEvents(msg string)
	}

	event struct {
		timestamp time.Time
		message   string
		detail    interface{}
	}

	eventLoggerImpl struct {
		logger     log.Logger
		timeSource clock.TimeSource

		events       []*event
		nextEventIdx int
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

		events:       make([]*event, maxSize),
		nextEventIdx: 0,
	}
}

func (e *eventLoggerImpl) AddEvent(
	eventMsg string,
	detail interface{},
) {
	e.events[e.nextEventIdx] = &event{
		timestamp: e.timeSource.Now(),
		message:   eventMsg,
		detail:    detail,
	}

	e.nextEventIdx = (e.nextEventIdx + 1) % len(e.events)
}

func (e *eventLoggerImpl) DumpEvents(
	msg string,
) {
	var builder strings.Builder
	eventIdx := e.nextEventIdx
	for i := 0; i != len(e.events); i++ {
		// dump all events in reverse order
		eventIdx--
		if eventIdx < 0 {
			eventIdx = len(e.events) - 1
		}

		event := e.events[eventIdx]
		if event == nil {
			break
		}

		builder.WriteString(fmt.Sprintf("%v, %s, %v\n", event.timestamp, event.message, event.detail))
	}

	e.logger.Info(msg, tag.Key("events"), tag.Value(builder.String()))
}
